package server

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeSkinZip builds a minimal skin package holding one WAD file.
func writeSkinZip(t *testing.T, path, wad string) {
	t.Helper()
	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	entry, err := archive.Create("WAD/Annie.wad.client")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(wad)); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}

func readExtractedWAD(t *testing.T, modsDir, skinID string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(modsDir, "skin_"+skinID, "WAD", "Annie.wad.client"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// The mods directory must only be reworked for mods that changed: a repeat
// apply of the same selection may not re-extract, and mods the selection no
// longer names are pruned.
func TestRefreshModsDirExtractsOnceAndPrunes(t *testing.T) {
	modsDir := t.TempDir()
	zipPath := filepath.Join(t.TempDir(), "skin.fantome")

	// A stale mod from a previous selection must go.
	stale := filepath.Join(modsDir, "skin_99")
	os.MkdirAll(filepath.Join(stale, "WAD"), 0755)
	os.WriteFile(filepath.Join(stale, "WAD", "Old.wad.client"), []byte("old"), 0600)

	writeSkinZip(t, zipPath, "v1")
	if err := refreshModsDir(modsDir, "1060", zipPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("stale mod directory survived the refresh")
	}
	if got := readExtractedWAD(t, modsDir, "1060"); got != "v1" {
		t.Fatalf("own skin not extracted: %q", got)
	}

	// A sentinel proves the second refresh did not re-extract the same bytes.
	sentinel := filepath.Join(modsDir, "skin_1060", "WAD", "Annie.wad.client")
	os.WriteFile(sentinel, []byte("sentinel"), 0600)
	if err := refreshModsDir(modsDir, "1060", zipPath); err != nil {
		t.Fatal(err)
	}
	if got := readExtractedWAD(t, modsDir, "1060"); got != "sentinel" {
		t.Fatalf("unchanged package was re-extracted: %q", got)
	}

	// New package bytes must invalidate the extracted mod.
	writeSkinZip(t, zipPath, "v2")
	if err := refreshModsDir(modsDir, "1060", zipPath); err != nil {
		t.Fatal(err)
	}
	if got := readExtractedWAD(t, modsDir, "1060"); got != "v2" {
		t.Fatalf("changed package was not re-extracted: %q", got)
	}

	// A different skin swaps mods: the previous one is pruned.
	if err := refreshModsDir(modsDir, "1070", zipPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(modsDir, "skin_1060")); !os.IsNotExist(err) {
		t.Fatal("previous skin's mod directory survived the swap")
	}
	if got := readExtractedWAD(t, modsDir, "1070"); got != "v2" {
		t.Fatalf("new skin not extracted: %q", got)
	}
}

// Default skin selections prune everything and extract nothing.
func TestRefreshModsDirDefaultSkin(t *testing.T) {
	modsDir := t.TempDir()
	keep := filepath.Join(modsDir, "skin_1060")
	os.MkdirAll(keep, 0755)
	if err := refreshModsDir(modsDir, "0", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keep); !os.IsNotExist(err) {
		t.Fatal("mod directory survived a default-skin refresh")
	}
}
