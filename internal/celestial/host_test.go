package celestial

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestHostChild(t *testing.T) {
	mode := os.Getenv("AME_TEST_HOST_CHILD")
	if mode == "" {
		return
	}
	packet, err := io.ReadAll(os.Stdin)
	if err != nil || len(packet) != 140 {
		os.Exit(9)
	}
	switch mode {
	case "error":
		fmt.Println("EVT INIT_ERR test rejection")
		os.Exit(2)
	case "exit":
		os.Exit(3)
	case "ready":
		fmt.Print("EVT INIT_")
		fmt.Print("OK\nEVT WAITING_FOR_GAME\n")
	case "timeout":
	default:
		os.Exit(8)
	}
	for {
		time.Sleep(time.Second)
	}
}

func TestHostLifecycle(t *testing.T) {
	for _, mode := range []string{"ready", "error", "exit", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			var child *exec.Cmd
			factory := func(ctx context.Context, _ string, args ...string) *exec.Cmd {
				if len(args) != 4 || args[2] != "500" || args[3] != "0" {
					t.Fatal("wrong host arguments")
				}
				child = exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHostChild$")
				child.Env = append(os.Environ(), "AME_TEST_HOST_CHILD="+mode)
				return child
			}
			config := HostConfig{Executable: os.Args[0], DLL: os.Args[0], Overlay: t.TempDir(), HookTimeoutMS: 500, StartupTimeout: 2 * time.Second}
			if mode == "timeout" {
				config.StartupTimeout = 100 * time.Millisecond
			}
			h, err := startHost(context.Background(), config, make([]byte, 140), factory)
			if mode == "ready" {
				if err != nil {
					t.Fatal(err)
				}
				if s := h.Status(); !s.Running || !s.Initialized || s.Hooked {
					h.Stop()
					t.Fatalf("wrong initialized state: %+v", s)
				}
				h.Stop()
				h.Stop()
				if s := h.Status(); s.Running || s.Hooked || s.Phase != "stopped" {
					t.Fatalf("wrong stopped state: %+v", s)
				}
			} else {
				if err == nil {
					h.Stop()
					t.Fatal("startup unexpectedly succeeded")
				}
				if mode == "error" && !strings.Contains(err.Error(), "test rejection") {
					t.Fatal(err)
				}
			}
			if child == nil || child.ProcessState == nil {
				t.Fatal("child was not reaped")
			}
		})
	}
}

func TestHostRequiresAuthorizationBeforeSpawn(t *testing.T) {
	called := false
	_, err := startHost(context.Background(), HostConfig{}, nil, func(context.Context, string, ...string) *exec.Cmd { called = true; return nil })
	if err == nil || called {
		t.Fatal("missing authorization reached process launch")
	}
}

func TestHostHookStates(t *testing.T) {
	h := &Host{status: HostStatus{Running: true}}
	h.consume("EVT INIT_OK")
	h.consume("EVT WAITING_FOR_GAME")
	h.consume("EVT GAME_FOUND tid=123")
	if h.Status().Hooked {
		t.Fatal("discovery incorrectly marked as hooked")
	}
	for _, line := range []string{"[DLL] EVT HOOK_APPLIED tid=123", "EVT HOOK_APPLIED tid=abc", "EVT HOOK_APPLIED tid=456"} {
		h.consume(line)
	}
	if h.Status().Hooked {
		t.Fatal("unrelated/invalid hook message changed state")
	}
	h.consume("EVT HOOK_APPLIED tid=123")
	if !h.Status().Hooked {
		t.Fatal("matching hook event not recognized")
	}
	h.consume("EVT GAME_EXIT tid=123 elapsed_ms=1000")
	if s := h.Status(); s.Hooked || !s.Running || s.ThreadID != 0 {
		t.Fatalf("game exit should leave host running: %+v", s)
	}
}
