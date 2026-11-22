package server

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveDiagnosticZIPCreatesUniqueReadableFiles(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("APPDATA", filepath.Join(root, "roaming"))
	first, err := saveDiagnosticZIP()
	if err != nil {
		t.Fatal(err)
	}
	second, err := saveDiagnosticZIP()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("export overwrote previous ZIP")
	}
	for _, name := range []string{first, second} {
		if filepath.Dir(name) != filepath.Join(root, "ame", "diagnostics") {
			t.Fatalf("unexpected destination: %s", name)
		}
		z, err := zip.OpenReader(name)
		if err != nil {
			t.Fatal(err)
		}
		if len(z.File) == 0 {
			t.Fatal("empty ZIP")
		}
		z.Close()
	}
}

func TestSaveDiagnosticZIPReportsUnwritableDestination(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", root)
	if err := os.WriteFile(filepath.Join(root, "ame"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := saveDiagnosticZIP(); err == nil {
		t.Fatal("save failure was hidden")
	}
}
