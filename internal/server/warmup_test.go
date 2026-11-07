package server

import (
	"testing"
	"time"
)

func stubWarmup(t *testing.T, game, client bool) chan string {
	t.Helper()
	reasons := make(chan string, 4)
	oldWarm, oldClient, oldGame, oldReady := warmManager, clientRunning, gameRunning, ready.Load()
	warmManager = func(reason string) error { reasons <- reason; return nil }
	clientRunning = func() bool { return client }
	gameRunning = func() bool { return game }
	t.Cleanup(func() {
		warmManager, clientRunning, gameRunning = oldWarm, oldClient, oldGame
		ready.Store(oldReady)
		mapPrepTimerMu.Lock()
		if mapPrepTimer != nil {
			mapPrepTimer.Stop()
		}
		mapPrepTimerMu.Unlock()
	})
	return reasons
}

func warmedWith(reasons chan string) string {
	select {
	case reason := <-reasons:
		return reason
	case <-time.After(200 * time.Millisecond):
		return ""
	}
}

// ame started with the League client open (champion select may already be
// running) starts the manager right away instead of at the first pick.
func TestSetupEndWarmsTheManagerWhenTheClientRuns(t *testing.T) {
	reasons := stubWarmup(t, false, true)
	ready.Store(false)
	SetReady()
	if got := warmedWith(reasons); got != "startup" {
		t.Fatalf("warm-up reason=%q, want startup", got)
	}
}

func TestNoWarmUpWithoutAClientOrDuringAGame(t *testing.T) {
	for _, c := range []struct {
		name         string
		game, client bool
	}{
		{"no client", false, false},
		{"League running", true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			reasons := stubWarmup(t, c.game, c.client)
			ready.Store(false)
			SetReady()
			if got := warmedWith(reasons); got != "" {
				t.Fatalf("warmed (%s)", got)
			}
		})
	}
}

func TestWarmUpWaitsForSetup(t *testing.T) {
	reasons := stubWarmup(t, false, true)
	ready.Store(false)
	warmManagerSoon("client connected")
	if got := warmedWith(reasons); got != "" {
		t.Fatalf("warmed before setup ended (%s)", got)
	}
}
