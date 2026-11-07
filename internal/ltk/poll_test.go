package ltk

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// At 22:14 the manager page stalled for a few seconds beside a loading game;
// the poll gave up for good, and ame told the user the match had no skins
// while the patcher had them live. A stall now leaves the manager unknown,
// and the poll picks it up again once the page answers.
func TestPollRecoversAfterTheManagerStopsAnswering(t *testing.T) {
	var stalled atomic.Bool
	stalled.Store(true)
	server := startFakeCDP(t, func(expr string) (any, error) {
		if stalled.Load() {
			return nil, errors.New("page busy")
		}
		if strings.Contains(expr, "__AME_BRIDGE__") && strings.Contains(expr, "get_patcher_status") {
			return map[string]any{"shim": true, "status": map[string]any{"running": true, "phase": "patching"}}, nil
		}
		return nil, nil
	})
	c, err := cdpDial(wsURL(server.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	m := &managerSession{cdp: c, stop: make(chan struct{}), exited: make(chan struct{})}
	defer close(m.stop)
	resetSnapshot()
	defer resetSnapshot()
	go pollManager(m)

	waitFor(t, "the stalled manager to be unknown", func() bool { return PreparedState().Unknown })
	if Running() {
		t.Fatal("an unanswering manager was reported running")
	}
	stalled.Store(false)
	waitFor(t, "the poll to recover", func() bool {
		s := currentSnapshot()
		return s.alive && s.running && !s.lost
	})
	if PreparedState().Unknown {
		t.Fatal("still unknown after the page answered")
	}
}

// An ame started mid-match has no session, but a manager an earlier ame
// left behind may be serving the game.
func TestAManagerAmeDoesNotDriveIsUnknown(t *testing.T) {
	resetSnapshot()
	old := managerProcessRunning
	defer func() { managerProcessRunning = old }()
	managerProcessRunning = func() bool { return true }
	if !PreparedState().Unknown {
		t.Fatal("a manager ame cannot see was reported as nothing armed")
	}
	managerProcessRunning = func() bool { return false }
	if PreparedState().Unknown {
		t.Fatal("no manager at all was reported unknown")
	}
}
