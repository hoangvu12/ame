package ltk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// TestLiveStartSession runs the full production path: Build packs the skin,
// Start installs it into the LTK Manager and starts its patcher, and the
// status accessors report the session. No game is launched, so the patcher
// stays in its scanning phase. Enabled only with AME_LIVE_START=1.
func TestLiveStartSession(t *testing.T) {
	if os.Getenv("AME_LIVE_START") != "1" {
		t.Skip("set AME_LIVE_START=1 to run the live start session test")
	}

	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	fetchStub := managerFetch
	managerFetch = ManagerFetch
	defer func() { managerFetch = fetchStub }()

	modsDir := os.Getenv("AME_LIVE_MODS")
	if modsDir == "" {
		modsDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame", "mods")
	}
	modName := os.Getenv("AME_LIVE_MOD_NAME")
	if modName == "" {
		modName = "skin_103088"
	}
	if _, err := os.Stat(filepath.Join(modsDir, modName)); err != nil {
		t.Skipf("no mod directory %s in %s: %v", modName, modsDir, err)
	}
	game := os.Getenv("AME_LIVE_GAME_DIR")
	if game == "" {
		game = `C:\Riot Games\League of Legends\Game`
	}
	if _, err := os.Stat(filepath.Join(game, "DATA")); err != nil {
		t.Skipf("no League game directory at %s: %v", game, err)
	}

	overlay := filepath.Join(config.AmeDir, "overlay")
	if err := os.MkdirAll(overlay, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Build(modsDir, overlay, game, modName); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(overlay + ".fantome"); err != nil {
		t.Fatalf("skin package missing after Build: %v", err)
	}

	if err := Start(overlay); err != nil {
		Stop()
		t.Fatalf("Start: %v", err)
	}
	defer Stop()

	// Exactly one ame mod survives the swap: leftovers from earlier runs are
	// uninstalled before the new package is installed.
	managerMu.Lock()
	session := manager
	managerMu.Unlock()
	if session == nil {
		t.Fatal("manager session missing after Start")
	}
	ameCount, total := countAmeMods(session)
	if ameCount != 1 {
		t.Fatalf("library holds %d ame mods after Start (total %d)", ameCount, total)
	}

	// The first run delivers the patcher DLL and builds the game index
	// before the patcher enters its scanning phase.
	deadline := time.Now().Add(4 * time.Minute)
	phase := ""
	for time.Now().Before(deadline) {
		s := currentSnapshot()
		phase = s.phase
		if s.failure != "" {
			Stop()
			t.Fatalf("patcher session failed: %s", s.failure)
		}
		if s.phase == "patching" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if phase != "patching" {
		t.Fatalf("patcher never left phase %q (running=%v)", phase, Running())
	}
	if !Running() || Hooked() || !WaitingForGame() {
		t.Fatalf("session state after start: running=%v hooked=%v waiting=%v failure=%q", Running(), Hooked(), WaitingForGame(), Failure())
	}
	fmt.Printf("phase=%s waiting=%v\n", phase, WaitingForGame())

	// A second Start for the same overlay preserves the session.
	if err := Start(overlay); err != nil {
		t.Fatalf("second Start: %v", err)
	}

	// A different skin package must stop the live patcher, replace the mod,
	// and start a new session.
	overlay2 := filepath.Join(config.AmeDir, "overlay2")
	if err := os.MkdirAll(overlay2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Build(modsDir, overlay2, game, modName); err != nil {
		t.Fatalf("Build second overlay: %v", err)
	}
	if err := Start(overlay2); err != nil {
		Stop()
		t.Fatalf("Start second package: %v", err)
	}
	deadline = time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) && currentSnapshot().phase != "patching" {
		if s := currentSnapshot(); s.failure != "" {
			Stop()
			t.Fatalf("second patcher session failed: %s", s.failure)
		}
		time.Sleep(500 * time.Millisecond)
	}
	managerMu.Lock()
	session = manager
	managerMu.Unlock()
	if session == nil {
		Stop()
		t.Fatal("manager session missing after the skin switch")
	}
	ameCount, _ = countAmeMods(session)
	if ameCount != 1 {
		t.Fatalf("library holds %d ame mods after the skin switch", ameCount)
	}
	if !Running() || Hooked() || !WaitingForGame() {
		t.Fatalf("session state after switch: running=%v hooked=%v waiting=%v failure=%q", Running(), Hooked(), WaitingForGame(), Failure())
	}

	Stop()
	if Running() || WaitingForGame() {
		t.Fatal("session survived Stop")
	}
}

// countAmeMods counts the library mods ame owns. The caller holds managerMu.
func countAmeMods(m *managerSession) (int, int) {
	mods, err := m.cdp.installedMods()
	if err != nil {
		return -1, -1
	}
	ame := 0
	for _, mod := range mods {
		if isAmeMod(mod, m.modID) {
			ame++
		}
	}
	return ame, len(mods)
}
