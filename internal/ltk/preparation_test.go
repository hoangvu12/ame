package ltk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayPublicationAndRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			root := t.TempDir()
			overlay, staged := filepath.Join(root, "overlay"), filepath.Join(root, "staged")
			os.Mkdir(overlay, 0755)
			os.WriteFile(filepath.Join(overlay, "old.wad"), []byte("old"), 0600)
			if !fail {
				os.Mkdir(staged, 0755)
				os.WriteFile(filepath.Join(staged, "new.wad"), []byte("new"), 0600)
			}
			err := publishOverlay(staged, overlay, nil)
			if (err != nil) != fail {
				t.Fatalf("publication error=%v, expected failure=%v", err, fail)
			}
			name, content := "new.wad", "new"
			if fail {
				name, content = "old.wad", "old"
			}
			got, err := os.ReadFile(filepath.Join(overlay, name))
			if err != nil || string(got) != content {
				t.Fatalf("complete overlay not preserved: %q %v", got, err)
			}
			entries, err := os.ReadDir(overlay)
			if err != nil || len(entries) != 1 {
				t.Fatalf("mixed overlay after publication: %v %v", entries, err)
			}
		})
	}
}
