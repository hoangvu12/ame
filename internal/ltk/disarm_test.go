package ltk

import (
	"errors"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakePatcher answers the manager page calls Disarm and the idle stop make.
type fakePatcher struct {
	mu        sync.Mutex
	running   bool
	stubborn  bool // keeps running after stop_patcher
	stopCalls int
}

func (f *fakePatcher) handle(expr string) (any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case strings.Contains(expr, "plugin:patcher|stop_patcher"):
		f.stopCalls++
		if !f.stubborn {
			f.running = false
		}
		return map[string]any{"__ipc": true, "ok": true, "value": nil}, nil
	case strings.Contains(expr, "plugin:patcher|get_patcher_status"):
		phase := "idle"
		if f.running {
			phase = "patching"
		}
		return map[string]any{"__ipc": true, "ok": true, "value": map[string]any{"running": f.running, "phase": phase}}, nil
	}
	return nil, nil
}

func (f *fakePatcher) stops() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stopCalls
}

// installFakeManager makes a live manager session backed by fake.
func installFakeManager(t *testing.T, fake *fakePatcher) *managerSession {
	t.Helper()
	server := startFakeCDP(t, fake.handle)
	c, err := cdpDial(wsURL(server.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	m := &managerSession{cdp: c, modID: "installed", modFile: "set.fantome", stop: make(chan struct{}), exited: make(chan struct{})}
	managerMu.Lock()
	manager = m
	managerMu.Unlock()
	setSnapshot(func(s *managerSnapshot) {
		*s = managerSnapshot{alive: true, running: fake.running, phase: "patching", observed: time.Now()}
	})
	snapshotMu.Lock()
	armedAt = time.Now()
	snapshotMu.Unlock()
	oldTimeout, oldIdle := disarmStopTimeout, managerIdleTimeout
	t.Cleanup(func() {
		managerMu.Lock()
		if manager != nil {
			stopLocked()
		}
		managerMu.Unlock()
		disarmStopTimeout, managerIdleTimeout = oldTimeout, oldIdle
		idleStopMu.Lock()
		if idleStopTimer != nil {
			idleStopTimer.Stop()
		}
		idleStopMu.Unlock()
	})
	return m
}

func TestDisarmKeepsTheManager(t *testing.T) {
	fake := &fakePatcher{running: true}
	m := installFakeManager(t, fake)
	if !PreparedState().Armed {
		t.Fatal("fixture: the set is not armed")
	}
	if err := Disarm(); err != nil {
		t.Fatal(err)
	}
	if fake.stops() != 1 {
		t.Fatalf("stop_patcher calls=%d, want 1", fake.stops())
	}
	managerMu.Lock()
	kept := manager == m && m.alive() && m.modID == "installed"
	managerMu.Unlock()
	if !kept {
		t.Fatal("disarm tore the manager session or its package bookkeeping down")
	}
	if Running() || PreparedState().Armed || PreparedState().Pending {
		t.Fatalf("the snapshot still reports an armed patcher: %+v", currentSnapshot())
	}
}

func TestDisarmNeverStopsAPatcherServingAMatch(t *testing.T) {
	fake := &fakePatcher{running: true}
	installFakeManager(t, fake)
	gameProcessRunning = func() bool { return true }
	defer func() { gameProcessRunning = func() bool { return false } }()
	if err := Disarm(); !errors.Is(err, ErrGameRunning) {
		t.Fatalf("err=%v, want ErrGameRunning", err)
	}
	if fake.stops() != 0 {
		t.Fatal("the patcher was stopped while League ran")
	}
}

func TestDisarmFallsBackToAFullStop(t *testing.T) {
	fake := &fakePatcher{running: true, stubborn: true}
	installFakeManager(t, fake)
	disarmStopTimeout = 200 * time.Millisecond
	if err := Disarm(); err == nil {
		t.Fatal("a patcher that kept running was reported disarmed")
	}
	managerMu.Lock()
	gone := manager == nil
	managerMu.Unlock()
	if !gone {
		t.Fatal("a patcher that would not stop kept its manager")
	}
}

func TestIdleManagerStops(t *testing.T) {
	fake := &fakePatcher{}
	installFakeManager(t, fake)
	stopIfIdle()
	managerMu.Lock()
	gone := manager == nil
	managerMu.Unlock()
	if !gone {
		t.Fatal("an idle manager was kept")
	}
}

func TestIdleStopSparesAnArmedPatcher(t *testing.T) {
	fake := &fakePatcher{running: true}
	m := installFakeManager(t, fake)
	stopIfIdle()
	managerMu.Lock()
	kept := manager == m
	managerMu.Unlock()
	if !kept || fake.stops() != 0 {
		t.Fatal("the idle stop took down an armed patcher")
	}
}

func TestDisarmStartsTheIdleCountdown(t *testing.T) {
	fake := &fakePatcher{running: true}
	installFakeManager(t, fake)
	managerIdleTimeout = 50 * time.Millisecond
	if err := Disarm(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		managerMu.Lock()
		gone := manager == nil
		managerMu.Unlock()
		if gone {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("the idle manager was never stopped after the disarm")
}

// Quitting ame during a match must not kill the injection host under the
// game: the Garen game whose host was killed mid-match hung on exit. The
// manager is handed to the after-game stop instead.
func TestStopLeavesTheManagerToAGameInProgress(t *testing.T) {
	fake := &fakePatcher{running: true}
	m := installFakeManager(t, fake)
	m.cmd = &exec.Cmd{}
	var deferred *exec.Cmd
	oldGame, oldAfter := gameProcessRunning, stopAfterGame
	gameProcessRunning = func() bool { return true }
	stopAfterGame = func(cmd *exec.Cmd) { deferred = cmd }
	defer func() { gameProcessRunning, stopAfterGame = oldGame, oldAfter }()
	Stop()
	if fake.stops() != 0 {
		t.Fatal("the patcher was stopped under a running game")
	}
	if deferred != m.cmd {
		t.Fatal("the manager was not handed to the after-game stop")
	}
	managerMu.Lock()
	gone := manager == nil
	managerMu.Unlock()
	if !gone || Running() {
		t.Fatal("ame still holds the manager it let go")
	}
}
