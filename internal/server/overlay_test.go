package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkinPackageReady(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "overlay.fantome")
	if skinPackageReady(pkg) {
		t.Fatal("missing package accepted")
	}
	os.WriteFile(pkg, []byte("package"), 0600)
	if !skinPackageReady(pkg) {
		t.Fatal("published package rejected")
	}
	os.Remove(pkg)
	os.Mkdir(pkg, 0755)
	if skinPackageReady(pkg) {
		t.Fatal("directory accepted as a package")
	}
	os.Remove(pkg)
	os.WriteFile(pkg, nil, 0600)
	if skinPackageReady(pkg) {
		t.Fatal("empty package accepted")
	}
}
