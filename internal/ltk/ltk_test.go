package ltk

import (
	"archive/zip"
	"bufio"
	"fmt"
	"github.com/hoangvu12/ame/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHostChild(t *testing.T) {
	mode := os.Getenv("AME_LTK_TEST_CHILD")
	if mode == "" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	step := 0
	for scanner.Scan() {
		line := scanner.Text()
		step++
		switch step {
		case 1:
			if line != "config loglevel 16" {
				os.Exit(20)
			}
		case 2:
			if line != "config flags 0" {
				os.Exit(21)
			}
		case 3:
			if !strings.HasPrefix(line, "config prefix ") || !strings.HasSuffix(line, `\`) {
				os.Exit(22)
			}
		case 4:
			if line != "start scan" {
				os.Exit(23)
			}
		}
		if line == "start scan" {
			switch mode {
			case "fail":
				fmt.Println("error 0 test failure")
			case "exit":
				os.Exit(24)
			case "scan":
				fmt.Println("status 0 injecting scanning for game")
			case "late":
				fmt.Println("status 0 injecting scanning for game")
				fmt.Println("status 0 injected dll attached")
				fmt.Println("dll 0 1 2 INFO ltk_patcher_dll::entry: joined too late, not overlaying")
			default:
				fmt.Println("status 0 injecting scanning for game")
				fmt.Println("status 0 injected dll attached")
			}
		}
		if line == "stop" {
			os.Exit(0)
		}
		if line == "probe" {
			fmt.Fprintln(os.Stderr, "game found; hook installed tid=1 pid=424242")
		}
	}
	os.Exit(0)
}

func TestHeldGameKeepsPreparedScanner(t *testing.T) {
	factory := func(_ string, _ ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
		cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD=scan")
		return cmd
	}
	if err := startHost(t.TempDir(), t.TempDir(), factory, true); err != nil {
		t.Fatal(err)
	}
	defer Stop()
	original := active
	if err := ArmSuspended(123); err != nil {
		t.Fatal(err)
	}
	if active != original || !WaitingForGame() {
		t.Fatal("handoff replaced the prepared scanner before the game could initialize its GUI thread")
	}
}

func TestRepeatedStartPreservesScanSession(t *testing.T) {
	factory := func(_ string, _ ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
		cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD=scan")
		return cmd
	}
	dir, overlay := t.TempDir(), t.TempDir()
	if err := startHost(dir, overlay, factory, true); err != nil {
		t.Fatal(err)
	}
	defer Stop()
	original := active
	// A changed skin still uses the same overlay prefix. Starting a new host
	// here changes its scan timestamp, potentially to AFTER game creation.
	if err := startHost(dir, overlay, factory, true); err != nil {
		t.Fatal(err)
	}
	if active != original {
		t.Fatal("skin update reset the scan session")
	}
}
func TestHostLifecycle(t *testing.T) {
	for _, mode := range []string{"ready", "fail", "exit"} {
		t.Run(mode, func(t *testing.T) {
			factory := func(_ string, _ ...string) *exec.Cmd {
				cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
				cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD="+mode)
				return cmd
			}
			err := startHost(t.TempDir(), t.TempDir(), factory, true)
			if mode != "ready" {
				if err == nil {
					Stop()
					t.Fatal("startup failure ignored")
				}
				if Running() {
					t.Fatal("failed host retained")
				}
				return
			}
			defer Stop()
			if err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(time.Second)
			for !Hooked() && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if !Running() || !Hooked() {
				t.Fatal("missing attached state")
			}
			Stop()
			Stop()
			if Running() || Hooked() {
				t.Fatal("stale stopped state")
			}
		})
	}
}
func TestPackMod(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "WAD"), 0755)
	os.WriteFile(filepath.Join(root, "WAD", "A.wad.client"), []byte("payload"), 0600)
	os.WriteFile(filepath.Join(root, ".ame-package"), []byte("marker"), 0600)
	path := filepath.Join(t.TempDir(), "mod.fantome")
	if err := packMod(root, path); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	if len(z.File) != 1 || z.File[0].Name != "WAD/A.wad.client" {
		t.Fatal("wrong extracted mod contents")
	}
}
func TestSignedHostConfiguration(t *testing.T) {
	if os.Getenv("AME_TEST_LTK") != "1" {
		t.Skip("opt-in signed host config-only test")
	}
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	dir, err := Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if err = startHost(dir, t.TempDir(), exec.Command, false); err != nil {
		t.Fatal(err)
	}
	defer Stop()
	if !Running() || Hooked() {
		t.Fatal("config-only host must not attach")
	}
}

func TestScanningAndLateAttachment(t *testing.T) {
	for _, mode := range []string{"scan", "late"} {
		t.Run(mode, func(t *testing.T) {
			factory := func(_ string, _ ...string) *exec.Cmd {
				cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
				cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD="+mode)
				return cmd
			}
			if err := startHost(t.TempDir(), t.TempDir(), factory, true); err != nil {
				t.Fatal(err)
			}
			defer Stop()
			if mode == "scan" {
				if !WaitingForGame() || Hooked() {
					t.Fatal("scanner not available for champion-select replacement")
				}
			} else {
				deadline := time.Now().Add(time.Second)
				for Failure() == "" && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if Failure() == "" || Hooked() || WaitingForGame() {
					t.Fatal("late attachment treated as successful")
				}
			}
		})
	}
}

func TestWaitForHookDoesNotAcceptScanning(t *testing.T) {
	factory := func(_ string, _ ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^TestHostChild$")
		cmd.Env = append(os.Environ(), "AME_LTK_TEST_CHILD=scan")
		return cmd
	}
	if err := startHost(t.TempDir(), t.TempDir(), factory, true); err != nil {
		t.Fatal(err)
	}
	defer Stop()
	if err := WaitForHook(123, nil, 30*time.Millisecond); err == nil {
		t.Fatal("scanning alone must not release a suspended game")
	}
	active.recordHook("1.287102100s INFO ltk_patcher_host::worker: game found; hook installed tid=45908 pid=456")
	if err := WaitForHook(123, nil, 30*time.Millisecond); err == nil {
		t.Fatal("wrong game's hook accepted")
	}
	active.recordHook("1.287102100s INFO ltk_patcher_host::worker: game found; hook installed tid=45908 pid=123")
	if err := WaitForHook(123, nil, time.Second); err != nil {
		t.Fatal(err)
	}
	if WaitingForGame() {
		t.Fatal("installed hook must prevent selection replacement")
	}
	cancel := make(chan struct{})
	close(cancel)
	if err := WaitForHook(456, cancel, time.Second); err == nil {
		t.Fatal("cancel ignored")
	}
	active.mu.Lock()
	active.failure = "overlay verification failed"
	active.mu.Unlock()
	if err := WaitForHook(123, nil, time.Second); err == nil {
		t.Fatal("failed overlay treated as ready")
	}
}
