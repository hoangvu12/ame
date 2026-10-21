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
		t.Fatalf("duplicate acquisition: %d", calls.Load())
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
		t.Fatal("unbounded acquisition")
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

func TestOverlayKeyChangesWithPackageDigest(t *testing.T) {
	testCache(t)
	r := PackageRequest{"103", "103001", ""}
	os.MkdirAll(skinDir(r.ChampionID, r.SkinID), 0755)
	record := func(digest string) {
		encoded, _ := json.Marshal(cacheMetadata{Source: "remote", Remote: &remoteMetadata{Digest: digest}})
		os.WriteFile(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded, 0600)
	}
	record(fmt.Sprintf("%x", sha256.Sum256([]byte("old package"))))
	before := OverlayKey("103001;custom")
	record(fmt.Sprintf("%x", sha256.Sum256([]byte("new package"))))
	if before == OverlayKey("103001;custom") {
		t.Fatal("stale overlay reusable across catalog revisions")
	}
	if OverlayKey(";custom") != ";custom" {
		t.Fatal("custom-only flow changed")
	}
}

func TestRemoveLocalGenerator(t *testing.T) {
	testCache(t)
	bundle := filepath.Join(config.AmeDir, "generators", "0123456789abcdef")
	os.MkdirAll(bundle, 0755)
	os.WriteFile(filepath.Join(bundle, "skin-generator.exe"), []byte("helper"), 0600)
	RemoveLocalGenerator()
	if _, err := os.Stat(filepath.Join(config.AmeDir, "generators")); !os.IsNotExist(err) {
		t.Fatal("old generator bundles left installed")
	}
	RemoveLocalGenerator()
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
