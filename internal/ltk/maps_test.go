package ltk

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// fakeMaps lays out a game folder and a manager overlay folder in temp
// directories and points the package at them.
type fakeMaps struct {
	game, overlay string
}

func newFakeMaps(t *testing.T, enabled bool) *fakeMaps {
	t.Helper()
	f := &fakeMaps{game: t.TempDir(), overlay: t.TempDir()}
	oldRoot, oldEnabled := managerOverlayRoot, mapPrepEnabled
	managerOverlayRoot = func() string { return f.overlay }
	mapPrepEnabled = func() bool { return enabled }
	t.Cleanup(func() { managerOverlayRoot, mapPrepEnabled = oldRoot, oldEnabled })
	return f
}

func writeSized(t *testing.T, path string, size int, modified time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
}

// gameMap writes League's copy of a map WAD.
func (f *fakeMaps) gameMap(t *testing.T, wad string, size int, modified time.Time) {
	writeSized(t, filepath.Join(gameMapsDir(f.game), wad), size, modified)
}

// copyMap writes the manager's overlay copy of a map WAD.
func (f *fakeMaps) copyMap(t *testing.T, wad string, size int, modified time.Time) {
	writeSized(t, filepath.Join(gameMapsDir(f.overlay), wad), size, modified)
}

func zipEntries(t *testing.T, path string) map[string]string {
	t.Helper()
	z, err := zip.OpenReader(path)
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
	return entries
}

func TestKeepAliveEntryNames(t *testing.T) {
	seen := map[string]bool{}
	for _, wad := range keepAliveMaps {
		entry := keepAliveEntry(wad)
		if !strings.HasPrefix(entry.name, "WAD/"+wad+"/") {
			t.Fatalf("%s: entry %q is not in the directory layout of its map", wad, entry.name)
		}
		if string(entry.data) != keepAliveContent {
			t.Fatalf("%s: keep-alive content is not constant: %q", wad, entry.data)
		}
		if seen[entry.name] {
			t.Fatalf("duplicate keep-alive path %q", entry.name)
		}
		seen[entry.name] = true
	}
	if got := keepAliveEntry("Map11.wad.client").name; got != "WAD/Map11.wad.client/ame/keepalive/map11.txt" {
		t.Fatalf("Map11 keep-alive path changed: %q", got)
	}
}

func TestMapsPrepared(t *testing.T) {
	f := newFakeMaps(t, true)
	if ok, reason := MapsPrepared(f.game); ok || reason != "League has none of the maps" {
		t.Fatalf("no maps in League: ok=%v reason=%q", ok, reason)
	}
	patched := time.Now().Add(-time.Hour)
	f.gameMap(t, "Map11.wad.client", 100, patched)
	f.gameMap(t, "Map12.wad.client", 100, patched)
	// Map30 is absent from this install: it is not required.
	if ok, reason := MapsPrepared(f.game); ok || !strings.Contains(reason, "Map11 missing") || !strings.Contains(reason, "Map12 missing") {
		t.Fatalf("missing copies: ok=%v reason=%q", ok, reason)
	}
	f.copyMap(t, "Map11.wad.client", 120, patched.Add(time.Minute))
	f.copyMap(t, "Map12.wad.client", 90, patched.Add(time.Minute))
	if ok, reason := MapsPrepared(f.game); ok || reason != "Map12 smaller than League's file" {
		t.Fatalf("smaller copy: ok=%v reason=%q", ok, reason)
	}
	f.copyMap(t, "Map12.wad.client", 120, patched.Add(-time.Minute))
	if ok, reason := MapsPrepared(f.game); ok || reason != "Map12 older than League's file" {
		t.Fatalf("older copy: ok=%v reason=%q", ok, reason)
	}
	f.copyMap(t, "Map12.wad.client", 120, patched.Add(time.Minute))
	if ok, reason := MapsPrepared(f.game); !ok {
		t.Fatalf("fresh copies not prepared: %q", reason)
	}
	// A League update rewrites its file: the copy is stale again.
	f.gameMap(t, "Map11.wad.client", 100, time.Now())
	if ok, reason := MapsPrepared(f.game); ok || reason != "Map11 older than League's file" {
		t.Fatalf("after a League update: ok=%v reason=%q", ok, reason)
	}
	if ok, _ := MapsPrepared(""); ok {
		t.Fatal("no game folder counted as prepared")
	}
}

func TestKeepAliveEntriesFollowReadyCopies(t *testing.T) {
	f := newFakeMaps(t, true)
	patched := time.Now().Add(-time.Hour)
	for _, wad := range keepAliveMaps {
		f.gameMap(t, wad, 10, patched)
	}
	if entries := keepAliveEntries(f.game); len(entries) != 0 {
		t.Fatalf("keep-alive entries without ready copies would force full copies: %v", entries)
	}
	f.copyMap(t, "Map12.wad.client", 20, patched.Add(time.Minute))
	entries := keepAliveEntries(f.game)
	if len(entries) != 1 || entries[0].name != keepAliveEntry("Map12.wad.client").name {
		t.Fatalf("only the ready Map12 copy is kept alive, got %v", entries)
	}
	f.copyMap(t, "Map11.wad.client", 20, patched.Add(time.Minute))
	f.copyMap(t, "Map30.wad.client", 20, patched.Add(time.Minute))
	if entries := keepAliveEntries(f.game); len(entries) != 3 {
		t.Fatalf("all ready maps kept alive, got %d", len(entries))
	}
	mapPrepEnabled = func() bool { return false }
	if entries := keepAliveEntries(f.game); len(entries) != 0 {
		t.Fatalf("feature off still adds entries: %v", entries)
	}
}

func TestMergeFantomesKeepAlive(t *testing.T) {
	mod := t.TempDir()
	os.MkdirAll(filepath.Join(mod, "WAD"), 0755)
	os.WriteFile(filepath.Join(mod, "WAD", "Zed.wad.client"), []byte("zed"), 0600)
	// A map mod ships Map30 packed: a keep-alive directory entry would hide it.
	os.WriteFile(filepath.Join(mod, "WAD", "Map30.wad.client"), []byte("arena"), 0600)
	scratch := t.TempDir()
	pack := filepath.Join(scratch, "0.fantome")
	if err := packMod(mod, pack); err != nil {
		t.Fatal(err)
	}
	var extra []packageEntry
	for _, wad := range keepAliveMaps {
		extra = append(extra, keepAliveEntry(wad))
	}
	merged := filepath.Join(scratch, "merged.fantome")
	if err := mergeFantomes([]string{pack}, merged, extra); err != nil {
		t.Fatal(err)
	}
	entries := zipEntries(t, merged)
	if entries["WAD/Zed.wad.client"] != "zed" || entries["WAD/Map30.wad.client"] != "arena" {
		t.Fatalf("mod entries changed: %v", entries)
	}
	for _, wad := range []string{"Map11.wad.client", "Map12.wad.client"} {
		if entries[keepAliveEntry(wad).name] != keepAliveContent {
			t.Fatalf("%s keep-alive entry missing: %v", wad, entries)
		}
	}
	if _, ok := entries[keepAliveEntry("Map30.wad.client").name]; ok {
		t.Fatal("keep-alive entry added for a map the mod already ships")
	}
}

func TestBuildCarriesKeepAliveForReadyMaps(t *testing.T) {
	stubGameWindow(t, false)
	f := newFakeMaps(t, true)
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	patched := time.Now().Add(-time.Hour)
	for _, wad := range keepAliveMaps {
		f.gameMap(t, wad, 10, patched)
		f.copyMap(t, wad, 20, patched.Add(time.Minute))
	}
	modsDir := filepath.Join(config.AmeDir, "mods")
	overlay := filepath.Join(config.AmeDir, "overlay")
	os.MkdirAll(filepath.Join(modsDir, "skin_1", "WAD"), 0755)
	os.WriteFile(filepath.Join(modsDir, "skin_1", "WAD", "Graves.wad.client"), []byte("graves"), 0600)
	if err := Build(modsDir, overlay, f.game, "skin_1"); err != nil {
		t.Fatal(err)
	}
	entries := zipEntries(t, overlay+".fantome")
	if entries["WAD/Graves.wad.client"] != "graves" {
		t.Fatalf("skin entry changed: %v", entries)
	}
	for _, wad := range keepAliveMaps {
		if entries[keepAliveEntry(wad).name] != keepAliveContent {
			t.Fatalf("%s keep-alive entry missing", wad)
		}
	}
}
