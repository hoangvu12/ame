package skin

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

func testCache(t *testing.T) {
	t.Helper()
	old, oldAme := config.SkinsDir, config.AmeDir
	config.AmeDir = t.TempDir()
	config.SkinsDir = filepath.Join(config.AmeDir, "skins")
	t.Cleanup(func() { config.SkinsDir, config.AmeDir = old, oldAme })
}

func TestRequestIDs(t *testing.T) {
	if normalizeRequest(PackageRequest{"103", "103001", "103001"}).BaseSkinID != "" {
		t.Fatal("reconnected ordinary skin treated as chroma")
	}
	for _, r := range []PackageRequest{{"103", "103001", ""}, {"103", "103002", "103001"}} {
		if err := validRequest(r); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range []PackageRequest{{"../1", "103001", ""}, {"103", "103000", ""}, {"103", "99001", ""}, {"103", "103001", "99001"}, {"103", "0103001", ""}} {
		if validRequest(r) == nil {
			t.Fatalf("accepted invalid request %+v", r)
		}
	}
}

func TestLocalCacheProvenanceAndZIPPrecedence(t *testing.T) {
	testCache(t)
	r := PackageRequest{"103", "103001", ""}
	dir := skinDir(r.ChampionID, r.SkinID)
	os.MkdirAll(dir, 0755)
	content := []byte("immutable local archive")
	m := &localMetadata{Request: r, Generator: generatorVersion, Schema: "revision", Build: "16.19", Artifact: "local.fantome", Digest: fmt.Sprintf("%x", sha256.Sum256(content)), Inputs: []inputFingerprint{{Path: "Ahri.wad.client", Size: 100, Header: "old"}}}
	os.WriteFile(filepath.Join(dir, m.Artifact), content, 0600)
	os.WriteFile(filepath.Join(dir, r.SkinID+".zip"), []byte("stale"), 0600)
	encoded, _ := json.Marshal(cacheMetadata{Source: "local", Local: m})
	os.WriteFile(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded, 0600)
	want := filepath.Join(dir, m.Artifact)
	if got := localCached(r, m); got != want {
		t.Fatal("local cache not reused")
	}
	for _, mutate := range []func(*localMetadata){func(n *localMetadata) { n.Build = "16.20" }, func(n *localMetadata) { n.Schema = "next" }, func(n *localMetadata) { n.Generator = "next" }, func(n *localMetadata) { n.Request.BaseSkinID = "103002" }, func(n *localMetadata) { n.MeshCompatibility = true }} {
		n := *m
		mutate(&n)
		if localCached(r, &n) != "" {
			t.Fatal("stale local package accepted")
		}
	}
	os.WriteFile(filepath.Join(dir, m.Artifact), []byte("corrupt"), 0600)
	if localCached(r, m) != "" {
		t.Fatal("corrupt artifact accepted")
	}
}

func TestConcurrentAcquisitionCancellationAndRepeatedSelections(t *testing.T) {
	old := acquirePackage
	defer func() { acquirePackage = old }()
	var calls, active, maxActive atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	acquirePackage = func(ctx context.Context, r PackageRequest, gameDir string) (string, error) {
		calls.Add(1)
		n := active.Add(1)
		if n > maxActive.Load() {
			maxActive.Store(n)
		}
		defer active.Add(-1)
		once.Do(func() { close(started) })
		select {
		case <-release:
		case <-ctx.Done():
			return "", ctx.Err()
		}
		return r.SkinID, nil
	}
	r := PackageRequest{"103", "103001", ""}
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { _, err := EnsurePackage(ctx, r, "game"); first <- err }()
	<-started
	second := make(chan string, 1)
	go func() { path, _ := EnsurePackage(context.Background(), r, "game"); second <- path }()
	// Join the same flight before releasing it.
	time.Sleep(20 * time.Millisecond)
	cancel()
	if err := <-first; err != context.Canceled {
		t.Fatalf("cancel = %v", err)
	}
	close(release)
	if path := <-second; path != r.SkinID {
		t.Fatal("caller cancellation killed shared work")
	}
	if calls.Load() != 1 {
		t.Fatalf("duplicate generation: %d", calls.Load())
	}
	var wg sync.WaitGroup
	for _, id := range []string{"103002", "103003", "22001", "99001", "86001"} {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			champ := id[:len(id)-3]
			path, err := EnsurePackage(context.Background(), PackageRequest{champ, id, ""}, "game")
			if err != nil || path != id {
				t.Errorf("obsolete completion: %s %v", path, err)
			}
		}()
	}
	wg.Wait()
	if maxActive.Load() != 1 {
		t.Fatal("unbounded generation")
	}
}

func writeTestPackage(t *testing.T, path, entry string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	w, _ := z.Create("META/info.json")
	w.Write([]byte(`{"Name":"test"}`))
	w, _ = z.Create("WAD/" + entry)
	w.Write([]byte("RW package fixture longer than header"))
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func TestExtractionReplacesStaleModAndRejectsPartial(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "skin.fantome")
	dest := filepath.Join(root, "mod")
	writeTestPackage(t, archive, "old.wad.client")
	if err := validatePackage(archive); err != nil {
		t.Fatal(err)
	}
	if err := ExtractPackage(archive, dest); err != nil {
		t.Fatal(err)
	}
	writeTestPackage(t, archive, "new.wad.client")
	if err := ExtractPackage(archive, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "WAD", "old.wad.client")); !os.IsNotExist(err) {
		t.Fatal("stale mod retained")
	}
	os.WriteFile(archive, []byte("partial zip"), 0600)
	if validatePackage(archive) == nil {
		t.Fatal("partial package accepted")
	}
	if ExtractPackage(archive, dest) == nil {
		t.Fatal("partial extraction accepted")
	}
	if _, err := os.Stat(filepath.Join(dest, "WAD", "new.wad.client")); err != nil {
		t.Fatal("failed extraction destroyed working mod")
	}
}

func TestOverlayKeyChangesOnInstallationUpdate(t *testing.T) {
	testCache(t)
	r := PackageRequest{"103", "103001", ""}
	dir := skinDir(r.ChampionID, r.SkinID)
	os.MkdirAll(dir, 0755)
	input := filepath.Join(t.TempDir(), "Ahri.wad.client")
	os.WriteFile(input, []byte("old"), 0600)
	fp, _ := fingerprint(input)
	encoded, _ := json.Marshal(cacheMetadata{Source: "local", Local: &localMetadata{Digest: "package", Inputs: []inputFingerprint{fp}}})
	os.WriteFile(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded, 0600)
	before := OverlayKey("103001;custom")
	os.WriteFile(input, []byte("new patch"), 0600)
	if before == OverlayKey("103001;custom") {
		t.Fatal("stale overlay reusable across patch")
	}
	if OverlayKey(";custom") != ";custom" {
		t.Fatal("custom-only flow changed")
	}
}

func TestGenerationFailureAndInstallationChangeDoNotPublish(t *testing.T) {
	testCache(t)
	old := invokeHelper
	defer func() { invokeHelper = old }()
	for _, mode := range []string{"failure", "partial", "patch", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			r := PackageRequest{"103", "103001", ""}
			inputDir := t.TempDir()
			m := &localMetadata{Request: r, Generator: generatorVersion}
			for _, name := range []string{"League.exe", "Ahri.wad.client"} {
				path := filepath.Join(inputDir, name)
				os.WriteFile(path, []byte("input"), 0600)
				fp, _ := fingerprint(path)
				m.Inputs = append(m.Inputs, fp)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			invokeHelper = func(_ context.Context, _ string, req helperRequest) (helperResult, error) {
				os.WriteFile(req.Output, []byte("partial"), 0600)
				if mode == "failure" {
					return helperResult{}, fmt.Errorf("unsupported conversion")
				}
				if mode != "partial" {
					writeTestPackage(t, req.Output, "Ahri.wad.client")
				}
				if mode == "patch" {
					os.WriteFile(m.Inputs[1].Path, []byte("updated input"), 0600)
				}
				if mode == "cancel" {
					cancel()
				}
				return helperResult{OK: true, Generator: generatorVersion, Inputs: []string{m.Inputs[1].Path}}, nil
			}
			if path, err := generatePackage(ctx, r, "game", "helper", m); err == nil || path != "" {
				t.Fatalf("published failed result %s %v", path, err)
			}
			entries, _ := os.ReadDir(skinDir(r.ChampionID, r.SkinID))
			if len(entries) != 0 {
				t.Fatalf("failed generation left cache entries: %v", entries)
			}
		})
	}
}

func TestExtractWindowsZIPDirectories(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bundle.zip")
	f, _ := os.Create(path)
	z := zip.NewWriter(f)
	z.Create(`licenses\`)
	w, _ := z.Create(`licenses\dependency\LICENSE`)
	w.Write([]byte("notice"))
	z.Close()
	f.Close()
	if err := Extract(path, filepath.Join(root, "out")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "out", "licenses", "dependency", "LICENSE")); err != nil {
		t.Fatal(err)
	}
}

func TestHelperDirInstallsFromLocalBundle(t *testing.T) {
	testCache(t)
	bundle := filepath.Join(t.TempDir(), "bundle.zip")
	f, _ := os.Create(bundle)
	z := zip.NewWriter(f)
	w, _ := z.Create("skin-generator.exe")
	w.Write([]byte("helper"))
	z.Close()
	f.Close()
	t.Setenv(generatorAssetEnv, bundle)
	dir, err := helperDir()
	if err != nil {
		t.Fatal(err)
	}
	if readyMarker(dir) == "" {
		t.Fatal("installed bundle has no ready marker")
	}
	if _, err := os.Stat(filepath.Join(dir, "skin-generator.exe")); err != nil {
		t.Fatal("bundle contents not extracted")
	}
	install, err := readGeneratorInstall(generatorManifestPath())
	if err != nil || install.Version != generatorVersion || install.Dir != dir {
		t.Fatalf("manifest incorrect %+v %v", install, err)
	}
	// A recorded install is reused without touching the bundle source.
	os.Remove(bundle)
	dir2, err := helperDir()
	if err != nil || dir2 != dir {
		t.Fatalf("installed bundle not reused: %s %v", dir2, err)
	}
	// A stale ready marker forces a reinstall after a restart.
	os.Remove(bundle)
	os.WriteFile(filepath.Join(dir, "ready"), []byte("stale"), 0600)
	preparedHelper, preparedRoot = "", ""
	if _, err := helperDir(); err == nil {
		t.Fatal("stale bundle accepted without reinstall")
	}
}
