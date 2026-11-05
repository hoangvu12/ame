package ltk

import (
	"os"
	"path/filepath"
	"testing"
)

// stubGameWindow keeps publication tests deterministic while the League
// client may be open on the test machine.
func stubGameWindow(t *testing.T, exists bool) {
	t.Helper()
	previous := gameWindowExists
	gameWindowExists = func() bool { return exists }
	t.Cleanup(func() { gameWindowExists = previous })
}

func TestSkinPackagePublication(t *testing.T) {
	stubGameWindow(t, false)
	root := t.TempDir()
	pkg, staged := filepath.Join(root, "overlay.fantome"), filepath.Join(root, "overlay.fantome.new")
	os.WriteFile(pkg, []byte("old package"), 0600)
	os.WriteFile(staged, []byte("new package"), 0600)
	if err := publishPackage(staged, pkg); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(pkg)
	if err != nil || string(got) != "new package" {
		t.Fatalf("package not swapped: %q %v", got, err)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatalf("staged package survived publication: %v", err)
	}
}

func TestSkinPackagePublicationFirstBuild(t *testing.T) {
	stubGameWindow(t, false)
	root := t.TempDir()
	pkg, staged := filepath.Join(root, "overlay.fantome"), filepath.Join(root, "overlay.fantome.new")
	os.WriteFile(staged, []byte("new package"), 0600)
	if err := publishPackage(staged, pkg); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(pkg)
	if err != nil || string(got) != "new package" {
		t.Fatalf("package not published: %q %v", got, err)
	}
}

func TestSkinPackagePublicationRefusedDuringGame(t *testing.T) {
	stubGameWindow(t, true)
	root := t.TempDir()
	pkg, staged := filepath.Join(root, "overlay.fantome"), filepath.Join(root, "overlay.fantome.new")
	os.WriteFile(pkg, []byte("old package"), 0600)
	os.WriteFile(staged, []byte("new package"), 0600)
	if err := publishPackage(staged, pkg); err == nil {
		t.Fatal("publication accepted while a game window exists")
	}
	got, err := os.ReadFile(pkg)
	if err != nil || string(got) != "old package" {
		t.Fatalf("package for the running match was replaced: %q %v", got, err)
	}
	if _, err := os.Stat(staged); err != nil {
		t.Fatalf("staged package disappeared on refusal: %v", err)
	}
}
