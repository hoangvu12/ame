package ltk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func startOwnedScannerFixture(t *testing.T) {
	t.Helper()
	factory := func(_ string, _ ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
		cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD=scan")
		return cmd
	}
	if err := startHost(t.TempDir(), t.TempDir(), factory, true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(Stop)
}

func TestScannerPauseRetainsSessionThroughPreparation(t *testing.T) {
	startOwnedScannerFixture(t)
	original := active
	resume, err := pausePreparedScanner()
	if err != nil {
		t.Fatal(err)
	}
	defer resume()
	// The child must not consume this until resumed. It simulates detection,
	// without looking for or injecting any real game process.
	if _, err := fmt.Fprintln(original.stdin, "probe"); err != nil {
		t.Fatal(err)
	}
	if err := WaitForHook(424242, nil, 75*time.Millisecond); err == nil {
		t.Fatal("owned scanner processed a detection while overlay was being prepared")
	}
	if err := resume(); err != nil {
		t.Fatal(err)
	}
	if err := resume(); err != nil {
		t.Fatal("release must be idempotent:", err)
	}
	if err := WaitForHook(424242, nil, time.Second); err != nil {
		t.Fatal(err)
	}
	if active != original {
		t.Fatal("preparation replaced the scan session")
	}
}

func TestStopDuringScannerPreparation(t *testing.T) {
	startOwnedScannerFixture(t)
	resume, err := pausePreparedScanner()
	if err != nil {
		t.Fatal(err)
	}
	owner := active
	Stop()
	if active != nil {
		t.Fatal("stopped scanner retained")
	}
	if err := resume(); err == nil {
		t.Fatal("cancelled preparation resurrected the scanner")
	}
	staged, overlay := filepath.Join(t.TempDir(), "staged"), t.TempDir()
	os.Mkdir(staged, 0755)
	if err := publishOverlay(staged, overlay, owner); err == nil {
		t.Fatal("cancelled build published after its session was stopped")
	}
}
