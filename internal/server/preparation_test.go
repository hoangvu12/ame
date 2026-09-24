package server

import (
	"fmt"
	"testing"
	"time"
)

type fakePreparationSuspender struct {
	suspended chan struct{}
	resumed   int
	fallback  bool
}

func (s *fakePreparationSuspender) SuspendProcess() error {
	if s.fallback {
		return fmt.Errorf("denied")
	}
	close(s.suspended)
	return nil
}
func (s *fakePreparationSuspender) WaitReady(<-chan struct{}, time.Duration) error { return nil }
func (s *fakePreparationSuspender) Suspend() (int, error)                          { close(s.suspended); return 1, nil }
func (s *fakePreparationSuspender) Resume() error                                  { s.resumed++; return nil }
func (s *fakePreparationSuspender) Close()                                         {}
func TestPreparationAlwaysReleasesSuspender(t *testing.T) {
	oldFind, oldNew := findPreparationProcess, newPreparationSuspender
	defer func() { findPreparationProcess = oldFind; newPreparationSuspender = oldNew }()
	for _, mode := range []string{"success", "build-error", "cancel", "thread-fallback"} {
		t.Run(mode, func(t *testing.T) {
			fake := &fakePreparationSuspender{suspended: make(chan struct{}), fallback: mode == "thread-fallback"}
			findPreparationProcess = func(string) uint32 { return 123 }
			newPreparationSuspender = func(uint32) (preparationSuspender, error) { return fake, nil }
			finish := beginPreparationSuspend()
			select {
			case <-fake.suspended:
			case <-time.After(time.Second):
				finish()
				t.Fatal("guard did not suspend")
			}
			if mode == "cancel" {
				activeSuspenderMu.Lock()
				close(activeSuspendState.cancel)
				activeSuspenderMu.Unlock()
			}
			// Normal completion and build errors both defer this same release operation.
			finish()
			if fake.resumed != 1 {
				t.Fatalf("resume count=%d", fake.resumed)
			}
			activeSuspenderMu.Lock()
			remaining := activeSuspendState
			activeSuspenderMu.Unlock()
			if remaining != nil {
				t.Fatal("stale suspend ownership")
			}
		})
	}
}

func TestPreparationWaitsForHookBeforeResume(t *testing.T) {
	oldFind, oldNew, oldWait := findPreparationProcess, newPreparationSuspender, waitPreparationHook
	defer func() {
		findPreparationProcess, newPreparationSuspender, waitPreparationHook = oldFind, oldNew, oldWait
	}()
	fake := &fakePreparationSuspender{suspended: make(chan struct{})}
	findPreparationProcess = func(string) uint32 { return 123 }
	newPreparationSuspender = func(uint32) (preparationSuspender, error) { return fake, nil }
	waited := false
	waitPreparationHook = func(pid uint32, cancel <-chan struct{}) error {
		if pid != 123 || fake.resumed != 0 {
			t.Error("game released before hook installation")
		}
		waited = true
		return nil
	}
	finish := beginPreparationSuspend()
	<-fake.suspended
	finish()
	if !waited || fake.resumed != 1 {
		t.Fatal("missing hook-ready handoff before resume")
	}
}

func TestQueuedSelectionsKeepOneHold(t *testing.T) {
	oldFind, oldNew, oldWait := findPreparationProcess, newPreparationSuspender, waitPreparationHook
	defer func() {
		findPreparationProcess, newPreparationSuspender, waitPreparationHook = oldFind, oldNew, oldWait
	}()
	fake := &fakePreparationSuspender{suspended: make(chan struct{})}
	created, handoffs := 0, 0
	findPreparationProcess = func(string) uint32 { return 123 }
	newPreparationSuspender = func(uint32) (preparationSuspender, error) { created++; return fake, nil }
	waitPreparationHook = func(uint32, <-chan struct{}) error { handoffs++; return nil }
	first, second := reservePreparation(), reservePreparation()
	work1 := beginPreparationSuspend()
	select {
	case <-fake.suspended:
	case <-time.After(time.Second):
		t.Fatal("not suspended")
	}
	work1()
	first()
	if fake.resumed != 0 || handoffs != 0 {
		t.Fatal("released between queued selections")
	}
	work2 := beginPreparationSuspend()
	work2()
	second()
	second() // lease release must be idempotent
	if created != 1 || fake.resumed != 1 || handoffs != 1 {
		t.Fatalf("created=%d resumes=%d handoffs=%d", created, fake.resumed, handoffs)
	}
}

func TestCancelReleasesDuringBlockedHandoff(t *testing.T) {
	testBlockedHandoffRelease(t, true)
}

func TestDeadlineReleasesDuringBlockedHandoff(t *testing.T) {
	oldLimit := preparationHoldLimit
	preparationHoldLimit = 50 * time.Millisecond
	defer func() { preparationHoldLimit = oldLimit }()
	testBlockedHandoffRelease(t, false)
}

func testBlockedHandoffRelease(t *testing.T, cancel bool) {
	t.Helper()
	oldFind, oldNew, oldWait := findPreparationProcess, newPreparationSuspender, waitPreparationHook
	defer func() {
		findPreparationProcess, newPreparationSuspender, waitPreparationHook = oldFind, oldNew, oldWait
	}()
	fake := &fakePreparationSuspender{suspended: make(chan struct{})}
	findPreparationProcess = func(string) uint32 { return 123 }
	newPreparationSuspender = func(uint32) (preparationSuspender, error) { return fake, nil }
	entered, unblock := make(chan struct{}), make(chan struct{})
	waitPreparationHook = func(uint32, <-chan struct{}) error {
		close(entered)
		<-unblock
		return nil
	}
	finish := beginPreparationSuspend()
	<-fake.suspended
	finished := make(chan struct{})
	go func() { finish(); close(finished) }()
	<-entered
	if cancel {
		activeSuspenderMu.Lock()
		close(activeSuspendState.cancel)
		activeSuspenderMu.Unlock()
	}
	select {
	case <-finished:
		close(unblock)
		if fake.resumed != 1 {
			t.Fatal("guard did not resume exactly once")
		}
	case <-time.After(200 * time.Millisecond):
		close(unblock)
		<-finished
		t.Fatal("blocked hook prevented the guard from releasing the game")
	}
}
