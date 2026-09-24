//go:build windows

package suspend

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestSuspendChild(t *testing.T) {
	if os.Getenv("AME_SUSPEND_CHILD") != "1" {
		return
	}
	for {
		fmt.Println("tick")
		time.Sleep(20 * time.Millisecond)
	}
}

// Exercise real suspend APIs against only a child owned by this test.
func TestOwnedProcessSuspendResume(t *testing.T) {
	if os.Getenv("AME_TEST_SUSPEND") != "1" {
		t.Skip("set AME_TEST_SUSPEND=1 for owned-child Windows API tests")
	}
	for _, method := range []string{"process", "threads"} {
		t.Run(method, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSuspendChild$")
			cmd.Env = append(os.Environ(), "AME_SUSPEND_CHILD=1")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { cmd.Process.Kill(); cmd.Wait() }()
			ticks := make(chan struct{}, 100)
			go func() {
				scanner := bufio.NewScanner(out)
				for scanner.Scan() {
					select {
					case ticks <- struct{}{}:
					default:
					}
				}
			}()
			select {
			case <-ticks:
			case <-time.After(3 * time.Second):
				t.Fatal("child did not start")
			}
			s, err := NewSuspender(uint32(cmd.Process.Pid))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if method == "process" {
				err = s.SuspendProcess()
			} else {
				var count int
				count, err = s.Suspend()
				if count == 0 && err == nil {
					t.Fatal("no suspended threads")
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if s.hProc != 0 {
					s.Resume()
				}
			}()
			time.Sleep(80 * time.Millisecond)
			for len(ticks) > 0 {
				<-ticks
			}
			select {
			case <-ticks:
				t.Fatal("heartbeat continued while suspended")
			case <-time.After(150 * time.Millisecond):
			}
			if err = s.Resume(); err != nil {
				t.Fatal(err)
			}
			select {
			case <-ticks:
			case <-time.After(3 * time.Second):
				t.Fatal("heartbeat failed to resume")
			}
		})
	}
}
