package server

import (
	"sync/atomic"
	"testing"
	"time"
)

type mapPrepStubs struct {
	game    atomic.Bool
	armed   bool
	stage   launchStage
	reasons []string
}

func stubMapPrep(t *testing.T) *mapPrepStubs {
	t.Helper()
	s := &mapPrepStubs{stage: stageBefore}
	oldGame, oldRunning, oldProbe, oldPrepare, oldReady := gameRunning, runtimeRunning, probeMatchTiming, prepareMaps, ready.Load()
	gameRunning = func() bool { return s.game.Load() }
	runtimeRunning = func() bool { return s.armed }
	probeMatchTiming = func() matchTiming { return matchTiming{stage: s.stage} }
	prepareMaps = func(reason string) error { s.reasons = append(s.reasons, reason); return nil }
	ready.Store(true)
	t.Cleanup(func() {
		gameRunning, runtimeRunning, probeMatchTiming, prepareMaps = oldGame, oldRunning, oldProbe, oldPrepare
		ready.Store(oldReady)
		mapPrepTimerMu.Lock()
		if mapPrepTimer != nil {
			mapPrepTimer.Stop()
		}
		mapPrepTimerMu.Unlock()
	})
	return s
}

func TestMapPrepRunsOnlyWhenNothingIsAtStake(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*mapPrepStubs)
		runs  bool
	}{
		{"idle client", func(*mapPrepStubs) {}, true},
		{"client closed", func(s *mapPrepStubs) { s.stage = stageUnknown }, true},
		{"League running", func(s *mapPrepStubs) { s.game.Store(true) }, false},
		{"skin set armed", func(s *mapPrepStubs) { s.armed = true }, false},
		{"champion select", func(s *mapPrepStubs) { s.stage = stageChampSelect }, false},
		{"match starting", func(s *mapPrepStubs) { s.stage = stageLaunching }, false},
		{"still starting", func(*mapPrepStubs) { ready.Store(false) }, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := stubMapPrep(t)
			ended := expectGameEndWaiter(t)
			c.setup(s)
			runMapPrep("test")
			if ran := len(s.reasons) == 1; ran != c.runs {
				t.Fatalf("prepared=%v, want %v", ran, c.runs)
			}
			if s.game.Load() {
				// Blocked by the game: the preparation comes back after it.
				s.game.Store(false)
				waitGameEnd(t, ended)
			}
		})
	}
}

// A selection being prepared right now goes first: the preparation comes
// back later instead of waiting on the lock.
func TestMapPrepYieldsToSkinPreparation(t *testing.T) {
	s := stubMapPrep(t)
	skinPreparationMu.Lock()
	runMapPrep("after game")
	skinPreparationMu.Unlock()
	if len(s.reasons) != 0 {
		t.Fatal("prepared maps while a skin set was being prepared")
	}
	mapPrepTimerMu.Lock()
	pending := mapPrepTimer != nil && mapPrepTimer.Stop()
	mapPrepTimerMu.Unlock()
	if !pending {
		t.Fatal("the preparation was dropped instead of retried")
	}
}

func TestScheduleMapPrepDebounces(t *testing.T) {
	s := stubMapPrep(t)
	done := make(chan struct{})
	prepareMaps = func(reason string) error {
		s.reasons = append(s.reasons, reason)
		close(done)
		return nil
	}
	scheduleMapPrep("startup", time.Hour)
	scheduleMapPrep("client connected", 10*time.Millisecond)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduled preparation never ran")
	}
	time.Sleep(50 * time.Millisecond)
	if len(s.reasons) != 1 || s.reasons[0] != "client connected" {
		t.Fatalf("reasons=%v, want only the latest trigger", s.reasons)
	}
}
