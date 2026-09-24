package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayExists(t *testing.T) {
	root := t.TempDir()
	if overlayExists(root) {
		t.Fatal("empty overlay accepted")
	}
	path := filepath.Join(root, "DATA", "FINAL", "Champions", "Ashe.wad.client")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte("wad"), 0600)
	if !overlayExists(root) {
		t.Fatal("prebuilt overlay rejected")
	}
	// Root-level .wad files alone are not a runtime overlay.
	root2 := t.TempDir()
	os.WriteFile(filepath.Join(root2, "Ashe.wad"), []byte("wad"), 0600)
	if overlayExists(root2) {
		t.Fatal("root-level .wad accepted as an overlay")
	}
}
