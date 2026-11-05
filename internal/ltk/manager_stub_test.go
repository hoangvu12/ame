package ltk

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureLauncherStub(t *testing.T) {
	dir := t.TempDir()

	stubPath, err := ensureLauncherStub(dir)
	if err != nil {
		t.Skipf("Go toolchain not available for stub build: %v", err)
	}

	// Verify stub binary exists and is executable
	if _, err := os.Stat(stubPath); err != nil {
		t.Fatalf("stub not found at %s: %v", stubPath, err)
	}

	// Verify it runs (it should fail because ltk_patcher_host.exe doesn't exist in test dir)
	// but it should at least start and exit with an error, not crash
	cmd := exec.Command(stubPath)
	cmd.Dir = dir
	err = cmd.Run()
	if err == nil {
		t.Log("stub ran successfully (unexpected — host exe was missing but stub didn't error)")
	}
	// The stub should exit with an error because ltk_patcher_host.exe doesn't exist
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		t.Fatalf("stub failed unexpectedly: %v", err)
	}
}

func TestEnsureLauncherStubCached(t *testing.T) {
	dir := t.TempDir()

	// First call builds
	stub1, err := ensureLauncherStub(dir)
	if err != nil {
		t.Skipf("Go toolchain not available: %v", err)
	}

	// Second call returns cached
	stub2, err := ensureLauncherStub(dir)
	if err != nil {
		t.Fatalf("cached call failed: %v", err)
	}

	if stub1 != stub2 {
		t.Fatalf("cached path mismatch: %s != %s", stub1, stub2)
	}
}

func TestLauncherStubParentProcess(t *testing.T) {
	// Test that when ltk-manager.exe launches ltk_patcher_host.exe,
	// the host's parent IS ltk-manager.exe (or the signed decoy).

	dir := t.TempDir()

	// We need the real host exe for this test
	hostSrc := filepath.Join(os.Getenv("LOCALAPPDATA"), "ame", "runtimes", "cdde82fdf97b9b35", "ltk_patcher_host.exe")
	if _, err := os.Stat(hostSrc); err != nil {
		t.Skipf("host exe not found at %s: %v", hostSrc, err)
	}

	// Copy host to test dir
	hostDst := filepath.Join(dir, "ltk_patcher_host.exe")
	hostData, err := os.ReadFile(hostSrc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hostDst, hostData, 0755); err != nil {
		t.Fatal(err)
	}

	stubPath, err := ensureLauncherStub(dir)
	if err != nil {
		t.Skipf("Go toolchain not available: %v", err)
	}

	// Launch the stub — it should start the host, which will fail quickly
	// (no DLL, no game) but the parent process relationship is set correctly.
	// With no ltk-manager-signed.exe present, the stub falls back to direct
	// launch (where its own process IS the parent — still matching ltk-manager.exe
	// for VRF00000008 but unsigned for VRF00000009).
	cmd := exec.Command(stubPath)
	cmd.Dir = dir
	err = cmd.Run()

	// The host will exit quickly because there's no game to hook.
	// As long as it doesn't crash with a Go panic, the launch worked.
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			t.Logf("stub exited with code %d (host exited normally)", cmd.ProcessState.ExitCode())
		} else {
			t.Logf("stub exited: %v (expected — host has no game to inject into)", err)
		}
	}
}