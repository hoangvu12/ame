package ltk

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// TestMain keeps the manager tests offline: the release client would
// otherwise reach GitHub on every fresh manager start.
func TestMain(m *testing.M) {
	managerFetch = func() (string, error) {
		return "", fmt.Errorf("delivery disabled in tests")
	}
	os.Exit(m.Run())
}

func TestPackMod(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "WAD"), 0755)
	os.WriteFile(filepath.Join(root, "WAD", "A.wad.client"), []byte("payload"), 0600)
	os.WriteFile(filepath.Join(root, ".ame-package"), []byte("marker"), 0600)
	path := filepath.Join(t.TempDir(), "mod.fantome")
	if err := packMod(root, path); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	if len(z.File) != 1 || z.File[0].Name != "WAD/A.wad.client" {
		t.Fatal("wrong extracted mod contents")
	}
}

func TestMergeFantomes(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	for _, mod := range []struct {
		dir, wad, content string
	}{
		{first, "First.wad.client", "one"},
		{second, "Second.wad.client", "two"},
		{second, "First.wad.client", "shadowed"},
	} {
		os.MkdirAll(filepath.Join(mod.dir, "WAD"), 0755)
		os.WriteFile(filepath.Join(mod.dir, "WAD", mod.wad), []byte(mod.content), 0600)
		os.MkdirAll(filepath.Join(mod.dir, "META"), 0755)
		os.WriteFile(filepath.Join(mod.dir, "META", "info.json"), []byte(`{"Name":"source"}`), 0600)
	}
	scratch := t.TempDir()
	packs := make([]string, 2)
	for i, dir := range []string{first, second} {
		packs[i] = filepath.Join(scratch, fmt.Sprintf("%d.fantome", i))
		if err := packMod(dir, packs[i]); err != nil {
			t.Fatal(err)
		}
	}
	merged := filepath.Join(scratch, "merged.fantome")
	if err := mergeFantomes(packs, merged); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(merged)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	entries := map[string]string{}
	for _, entry := range z.File {
		rc, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[entry.Name] = string(data)
	}
	// Earlier mods win: the shadowed WAD keeps the first mod's content.
	if entries["WAD/First.wad.client"] != "one" {
		t.Fatalf("first mod did not win the merge: %q", entries["WAD/First.wad.client"])
	}
	if entries["WAD/Second.wad.client"] != "two" {
		t.Fatalf("second mod content missing: %q", entries["WAD/Second.wad.client"])
	}
	info := entries["META/info.json"]
	if !strings.Contains(info, ameModName) {
		t.Fatalf("merged package does not carry ame's library name: %q", info)
	}
	if strings.Contains(info, "source") {
		t.Fatalf("source metadata leaked into the merged package: %q", info)
	}
}

func TestStartRequiresSkinPackage(t *testing.T) {
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	overlay := filepath.Join(config.AmeDir, "overlay")
	if _, err := Prepare(); err != nil {
		t.Fatal(err)
	}
	err := Start(overlay)
	if err == nil {
		Stop()
		t.Fatal("Start accepted an overlay without a built skin package")
	}
	if !strings.Contains(err.Error(), "skin package") {
		t.Fatalf("missing-package error is unclear: %v", err)
	}
}

func TestStatusAccessors(t *testing.T) {
	resetSnapshot := func() {
		setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{} })
	}
	resetSnapshot()
	defer resetSnapshot()

	if Running() || Hooked() || WaitingForGame() || Failure() != "" {
		t.Fatal("empty snapshot reports activity")
	}

	setSnapshot(func(s *managerSnapshot) {
		s.alive = true
		s.running = true
		s.phase = "patching"
	})
	if !Running() || !WaitingForGame() || Hooked() {
		t.Fatal("scanning session not reported as waiting for a game")
	}

	setSnapshot(func(s *managerSnapshot) { s.attached = true; s.pid = 4242 })
	if !Hooked() || WaitingForGame() {
		t.Fatal("attached session not reported as hooked")
	}
	if err := WaitForHook(4242, nil, time.Second); err != nil {
		t.Fatalf("attach wait failed: %v", err)
	}
	if err := WaitForHook(1111, nil, 50*time.Millisecond); err == nil {
		t.Fatal("wrong game's hook accepted")
	}
	// A closed cancel channel only aborts a wait that has not attached yet.
	setSnapshot(func(s *managerSnapshot) { s.attached = false; s.pid = 0 })
	cancel := make(chan struct{})
	close(cancel)
	if err := WaitForHook(0, cancel, time.Second); err == nil {
		t.Fatal("cancel ignored")
	}

	setSnapshot(func(s *managerSnapshot) { s.failure = "overlay failed" })
	if Hooked() || WaitingForGame() {
		t.Fatal("failed session reported as usable")
	}
	if Failure() == "" {
		t.Fatal("failure not reported")
	}
}

func TestArmSuspendedWaitsForPatching(t *testing.T) {
	old := armSuspendTimeout
	armSuspendTimeout = 150 * time.Millisecond
	defer func() { armSuspendTimeout = old }()
	reset := func() {
		setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{} })
	}
	reset()
	defer reset()

	setSnapshot(func(s *managerSnapshot) {
		s.alive = true
		s.running = true
		s.phase = "building"
	})
	if err := ArmSuspended(123); err == nil {
		t.Fatal("overlay build released the game early")
	}

	setSnapshot(func(s *managerSnapshot) {
		s.alive = true
		s.running = false
		s.phase = "idle"
	})
	if err := ArmSuspended(123); err == nil {
		t.Fatal("stopped patcher released the game")
	}

	setSnapshot(func(s *managerSnapshot) {
		s.alive = true
		s.running = true
		s.phase = "patching"
	})
	if err := ArmSuspended(123); err != nil {
		t.Fatalf("armed patcher did not release the game: %v", err)
	}
}

func TestOverlayFailureMessage(t *testing.T) {
	if overlayFailureMessage("Live") != "" || overlayFailureMessage("None") != "" || overlayFailureMessage("") != "" {
		t.Fatal("non-failing overlay outcomes reported as failures")
	}
	for _, outcome := range []string{"TooLate", "EndOfLife", "Disabled", "HookFailed"} {
		if overlayFailureMessage(outcome) == "" {
			t.Fatalf("%s is not reported as a failure", outcome)
		}
	}
}

// Build produces only the skin package the manager installs: no extracted
// overlay is written (the manager builds its own), and a second Build
// replaces the package in place.
func TestBuildProducesSkinPackageOnly(t *testing.T) {
	stubGameWindow(t, false)
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	root := config.AmeDir
	modsDir := filepath.Join(root, "mods")
	overlay := filepath.Join(root, "overlay")
	modDir := filepath.Join(modsDir, "skin_1060", "WAD")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "Annie.wad.client"), []byte("wad"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Build(modsDir, overlay, `C:\Riot Games\League of Legends\Game`, "skin_1060"); err != nil {
		t.Fatal(err)
	}
	pkg := overlay + ".fantome"
	data, err := os.ReadFile(pkg)
	if err != nil {
		t.Fatalf("skin package not published: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("published package is not a zip: %v", err)
	}
	var sawWAD, sawMeta bool
	for _, entry := range reader.File {
		switch entry.Name {
		case "WAD/Annie.wad.client":
			sawWAD = true
		case "META/info.json":
			sawMeta = true
		}
	}
	if !sawWAD || !sawMeta {
		t.Fatalf("package contents incomplete: wad=%v meta=%v", sawWAD, sawMeta)
	}
	if _, err := os.Stat(overlay); !os.IsNotExist(err) {
		t.Fatalf("extracted overlay published: %v", err)
	}
	if _, err := os.Stat(overlay + ".fantome.new"); !os.IsNotExist(err) {
		t.Fatalf("staged package survived publication: %v", err)
	}

	// A rebuild replaces the package and still writes no overlay.
	if err := os.WriteFile(filepath.Join(modDir, "Annie.wad.client"), []byte("wad2"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Build(modsDir, overlay, `C:\Riot Games\League of Legends\Game`, "skin_1060"); err != nil {
		t.Fatal(err)
	}
	entries, err := zip.OpenReader(pkg)
	if err != nil {
		t.Fatal(err)
	}
	defer entries.Close()
	var replaced bool
	for _, entry := range entries.File {
		if entry.Name != "WAD/Annie.wad.client" {
			continue
		}
		body, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			t.Fatal(err)
		}
		replaced = string(content) == "wad2"
	}
	if !replaced {
		t.Fatal("second build did not replace the package contents")
	}
	if _, err := os.Stat(overlay); !os.IsNotExist(err) {
		t.Fatalf("extracted overlay published on rebuild: %v", err)
	}
}
