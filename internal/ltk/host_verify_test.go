package ltk

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHostVerificationBypass runs the full host startup with the delivered
// DLL and checks that host verification passes (no VRF00000008).
//
// Requires: the real ltk_patcher_host.exe and ltk_patcher_dll.dll in the
// runtime directory. Set AME_LIVE_HOST_TEST=1 to enable.
func TestHostVerificationBypass(t *testing.T) {
	if os.Getenv("AME_LIVE_HOST_TEST") != "1" {
		t.Skip("set AME_LIVE_HOST_TEST=1 to run live host verification test")
	}

	// Find the runtime directory with a real host + DLL
	runtimeDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "ame", "runtimes", "cdde82fdf97b9b35")
	hostExe := filepath.Join(runtimeDir, "ltk_patcher_host.exe")
	dllPath := filepath.Join(runtimeDir, "ltk_patcher_dll.dll")

	if _, err := os.Stat(hostExe); err != nil {
		t.Skipf("host not found at %s: %v", hostExe, err)
	}
	if _, err := os.Stat(dllPath); err != nil {
		// Try to fetch via delivery
		t.Log("DLL not found, fetching via delivery...")
		deliveryFetch = nil // restore real DeliveryFetch
		dll, _, err := DeliveryFetch(runtimeDir)
		if err != nil {
			t.Skipf("delivery fetch failed: %v", err)
		}
		t.Logf("fetched DLL: %s", dll)
	}

	// Build the stub
	stubExe, err := ensureLauncherStub(runtimeDir)
	if err != nil {
		t.Skipf("Go not available for stub build: %v", err)
	}
	t.Logf("stub: %s", stubExe)

	// Launch the stub (which launches the host)
	cmd := exec.Command(stubExe)
	cmd.Dir = runtimeDir
	cmd.SysProcAttr = nil // don't hide — we want to see output

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		t.Fatalf("stub Start: %v", err)
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Read combined stdout+stderr
	type line struct {
		source string
		text   string
	}
	lines := make(chan line, 200)

	readStream := func(prefix string, r io.Reader) {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 4096), 256*1024)
		for sc.Scan() {
			lines <- line{prefix, sc.Text()}
		}
	}
	go readStream("out", stdout)
	go readStream("err", stderr)

	// Send config and start scan
	time.Sleep(500 * time.Millisecond) // wait for host to load DLL
	io.WriteString(stdin, "config loglevel 32\nconfig flags 4\nconfig prefix .\\\nstart scan\n")

	// Collect output for up to 10 seconds
	deadline := time.After(10 * time.Second)
	var allLines []line

loop:
	for {
		select {
		case l := <-lines:
			allLines = append(allLines, l)
			t.Logf("[%s] %s", l.source, l.text)
			if strings.Contains(l.text, "scanning for game") {
				t.Log("✅ host verification PASSED — scanning for game (no VRF error)")
				break loop
			}
			if strings.Contains(l.text, "host verification failed") || strings.Contains(l.text, "VRF00000008") {
				t.Errorf("❌ host verification FAILED: %s", l.text)
				break loop
			}
			if strings.Contains(l.text, "end of life") {
				t.Log("⚠️  DLL EOL reached (expected for bundled DLL); verification result unknown")
				break loop
			}
		case <-deadline:
			t.Log("timeout — host output:")
			for _, l := range allLines {
				t.Logf("  [%s] %s", l.source, l.text)
			}
			t.Error("did not see 'scanning for game' or VRF error within timeout")
			break loop
		}
	}
}