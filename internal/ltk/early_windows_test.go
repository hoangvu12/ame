package ltk

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/suspend"
	"golang.org/x/sys/windows"
)

func TestEarlyHookChild(t *testing.T) {
	if os.Getenv("AME_EARLY_HOOK_CHILD") != "1" {
		return
	}
	// League statically imports user32; mirror that without creating a window
	// or pumping messages. Console-only Go children do not load it by default.
	if _, err := windows.LoadLibrary("user32.dll"); err != nil {
		os.Exit(31)
	}
	fmt.Println("ready-without-window")
	for {
		time.Sleep(time.Second)
	}
}

// No game is touched. Hooks are removed before this owned child is resumed, so
// the signed DLL never executes in the test process. We verify installation at
// precisely the no-window, suspended point where scanning cannot find a game.
func TestSignedEarlyHookWithoutWindow(t *testing.T) {
	if os.Getenv("AME_TEST_LTK") != "1" {
		t.Skip("set AME_TEST_LTK=1 for owned-process early hook test")
	}
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	cmd := exec.Command(os.Args[0], "-test.run=^TestEarlyHookChild$")
	cmd.Env = append(os.Environ(), "AME_EARLY_HOOK_CHILD=1")
	hide(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		ready <- scanner.Scan() && scanner.Text() == "ready-without-window"
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("child not ready")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("child startup timeout")
	}
	pid := uint32(cmd.Process.Pid)
	s, err := suspend.NewSuspender(pid)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err = s.SuspendProcess(); err != nil {
		t.Fatal(err)
	}
	defer s.Resume()
	if err = StartSuspended(t.TempDir(), pid); err != nil {
		t.Fatal(err)
	}
	defer Stop() // runs before Resume
	if err = WaitForHook(pid, nil, time.Second); err != nil {
		t.Fatal(err)
	}
	if !Running() || Hooked() || WaitingForGame() {
		t.Fatal("wrong pre-resume host state")
	}
	Stop()
	if Running() {
		t.Fatal("early hook host survived cleanup")
	}
}
