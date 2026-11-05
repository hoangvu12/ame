package ltk

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestLiveNsisExtraction extracts the manager runtime pair from a real LTK
// Manager installer with the pure-Go NSIS reader. With 7z on PATH it also
// cross-checks the extracted bytes against 7-Zip's own output. Enabled with
// AME_LIVE_NSIS=1 and AME_LIVE_NSIS_INSTALLER=<path to the setup.exe>.
func TestLiveNsisExtraction(t *testing.T) {
	installer := os.Getenv("AME_LIVE_NSIS_INSTALLER")
	if installer == "" {
		if os.Getenv("AME_LIVE_NSIS") != "1" {
			t.Skip("set AME_LIVE_NSIS=1 (and optionally AME_LIVE_NSIS_INSTALLER) to run the live NSIS extraction test")
		}
		installer = filepath.Join(os.TempDir(), "ltk-setup.exe")
		if _, err := os.Stat(installer); err != nil {
			t.Skipf("no installer at %s: %v", installer, err)
		}
	}
	data, err := os.ReadFile(installer)
	if err != nil {
		t.Fatal(err)
	}
	files, err := nsisExtractFiles(data, managerExeName, "ltk_patcher_host.exe")
	if err != nil {
		t.Fatalf("NSIS extraction: %v", err)
	}

	dir := t.TempDir()
	for _, name := range []string{managerExeName, "ltk_patcher_host.exe"} {
		content := files[name]
		if len(content) == 0 {
			t.Fatalf("%s extracted as empty", name)
		}
		if name == managerExeName && len(content) < 1024*1024 {
			t.Fatalf("%s looks truncated: %d bytes", name, len(content))
		}
		t.Logf("%s: %d bytes sha256=%s", name, len(content), shortHash(content))
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Cross-check with 7-Zip when it is available.
	sevenZip, err := exec.LookPath("7z")
	if err != nil {
		t.Log("7z not on PATH; skipping the cross-check")
		return
	}
	outDir := t.TempDir()
	cmd := exec.Command(sevenZip, "x", "-o"+outDir, "-y", installer, managerExeName, "ltk_patcher_host.exe")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		t.Logf("7z failed (%v); skipping the cross-check", err)
		return
	}
	for _, name := range []string{managerExeName, "ltk_patcher_host.exe"} {
		reference, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Logf("7z produced no %s; skipping it", name)
			continue
		}
		gotSum := sha256.Sum256(files[name])
		wantSum := sha256.Sum256(reference)
		if !bytes.Equal(gotSum[:], wantSum[:]) {
			t.Fatalf("%s differs from 7-Zip's extraction (got %x, want %x)", name, gotSum, wantSum)
		}
		t.Logf("%s matches 7-Zip's extraction", name)
	}
}

func shortHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}
