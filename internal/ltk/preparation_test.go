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

func TestOverlayPublicationAndRollback(t *testing.T) {
	stubGameWindow(t, false)
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			root := t.TempDir()
			overlay, staged := filepath.Join(root, "overlay"), filepath.Join(root, "staged")
			pkg, pkgTo := filepath.Join(root, "staged.pkg"), filepath.Join(root, "overlay.fantome")
			os.Mkdir(overlay, 0755)
			os.WriteFile(filepath.Join(overlay, "old.wad"), []byte("old"), 0600)
			os.WriteFile(pkgTo, []byte("old package"), 0600)
			if !fail {
				os.Mkdir(staged, 0755)
				os.WriteFile(filepath.Join(staged, "new.wad"), []byte("new"), 0600)
				os.WriteFile(pkg, []byte("new package"), 0600)
			}
			err := publishOverlay(staged, overlay, pkg, pkgTo)
			if (err != nil) != fail {
				t.Fatalf("publication error=%v, expected failure=%v", err, fail)
			}
			name, content := "new.wad", "new"
			pkgContent := "new package"
			if fail {
				name, content, pkgContent = "old.wad", "old", "old package"
			}
			got, err := os.ReadFile(filepath.Join(overlay, name))
			if err != nil || string(got) != content {
				t.Fatalf("complete overlay not preserved: %q %v", got, err)
			}
			entries, err := os.ReadDir(overlay)
			if err != nil || len(entries) != 1 {
				t.Fatalf("mixed overlay after publication: %v %v", entries, err)
			}
			pkgGot, err := os.ReadFile(pkgTo)
			if err != nil || string(pkgGot) != pkgContent {
				t.Fatalf("skin package not swapped with the overlay: %q %v", pkgGot, err)
			}
		})
	}
}

func TestOverlayPublicationWithoutPackage(t *testing.T) {
	stubGameWindow(t, false)
	root := t.TempDir()
	overlay, staged := filepath.Join(root, "overlay"), filepath.Join(root, "staged")
	os.Mkdir(overlay, 0755)
	os.WriteFile(filepath.Join(overlay, "old.wad"), []byte("old"), 0600)
	pkgTo := filepath.Join(root, "overlay.fantome")
	os.WriteFile(pkgTo, []byte("stale package"), 0600)
	os.Mkdir(staged, 0755)
	os.WriteFile(filepath.Join(staged, "new.wad"), []byte("new"), 0600)
	if err := publishOverlay(staged, overlay, "", pkgTo); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(pkgTo); !os.IsNotExist(err) {
		t.Fatalf("package from the previous overlay survived: %v", err)
	}
}

func TestOverlayPublicationRefusedDuringGame(t *testing.T) {
	stubGameWindow(t, true)
	root := t.TempDir()
	overlay, staged := filepath.Join(root, "overlay"), filepath.Join(root, "staged")
	os.Mkdir(overlay, 0755)
	os.WriteFile(filepath.Join(overlay, "old.wad"), []byte("old"), 0600)
	os.Mkdir(staged, 0755)
	os.WriteFile(filepath.Join(staged, "new.wad"), []byte("new"), 0600)
	if err := publishOverlay(staged, overlay, "", overlay+".fantome"); err == nil {
		t.Fatal("publication accepted while a game window exists")
	}
	got, err := os.ReadFile(filepath.Join(overlay, "old.wad"))
	if err != nil || string(got) != "old" {
		t.Fatalf("overlay for the running match was replaced: %q %v", got, err)
	}
}
