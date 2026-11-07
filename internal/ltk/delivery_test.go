package ltk

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A manager launch reuses the release check setup made at startup, so no
// GitHub round-trip sits between a skin pick and the patcher scan.
func TestManagerFetchReusesARecentReleaseCheck(t *testing.T) {
	oldDir, oldSetup := managerDir, latestSetup
	managerDir = t.TempDir()
	checks := 0
	latestSetup = func() (string, string, error) {
		checks++
		return "v1.33.0", "", nil
	}
	defer func() {
		managerDir, latestSetup = oldDir, oldSetup
		releaseCheckMu.Lock()
		releaseChecked = time.Time{}
		releaseCheckMu.Unlock()
	}()
	payload := bytes.Repeat([]byte{1}, 2048)
	for _, name := range []string{managerExeName, managerHostName} {
		if err := os.WriteFile(filepath.Join(managerDir, name), payload, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(managerDir, managerReleaseFile), []byte("v1.33.0"), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		exe, err := ManagerFetch()
		if err != nil || exe != filepath.Join(managerDir, managerExeName) {
			t.Fatalf("exe=%q err=%v", exe, err)
		}
	}
	if checks != 1 {
		t.Fatalf("release checks=%d, want 1", checks)
	}

	// A failed check falls back to the delivered pair and is not repeated
	// at the next launch either.
	releaseCheckMu.Lock()
	releaseChecked = time.Time{}
	releaseCheckMu.Unlock()
	checks = 0
	latestSetup = func() (string, string, error) {
		checks++
		return "", "", fmt.Errorf("offline")
	}
	for i := 0; i < 2; i++ {
		if _, err := ManagerFetch(); err != nil {
			t.Fatal(err)
		}
	}
	if checks != 1 {
		t.Fatalf("failed release checks=%d, want 1", checks)
	}
}
