package celestial

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareInstalledOverlay(t *testing.T) {
	name := os.Getenv("AME_TEST_CELESTIAL_INDEX")
	if name == "" {
		t.Skip("set AME_TEST_CELESTIAL_INDEX for installed Ahri fixture")
	}
	encoded, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := DecodeIndex(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixture.Sources) != 1 || len(fixture.Targets) != 1 {
		t.Fatal("requires one-source, one-target fixture")
	}
	prepared, err := PrepareOverlay(t.TempDir(), []WADInput{{fixture.Sources[0], []string{fixture.Targets[0].Name}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"wad.idx", "celestial-overlay.seal"} {
		want, err := os.ReadFile(filepath.Join(filepath.Dir(name), file))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(prepared.Directory, file))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s differs from installed fixture", file)
		}
		t.Logf("%s: %d bytes match installed fixture", file, len(got))
	}
}

func TestPrepareInvalidInputLeavesParentUntouched(t *testing.T) {
	parent := t.TempDir()
	sentinel := filepath.Join(parent, "wad.idx")
	if err := os.WriteFile(sentinel, []byte("existing overlay"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareOverlay(parent, nil); err == nil {
		t.Fatal("accepted empty overlay")
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || string(got) != "existing overlay" {
		t.Fatal("modified existing files")
	}
}
