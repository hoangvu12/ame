package server

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/ltk"
)

// expectGameEndWaiter makes the after-game waiter fast and observable. The
// test must let it finish (waitGameEnd) before its stubs are restored.
func expectGameEndWaiter(t *testing.T) chan struct{} {
	t.Helper()
	ended := make(chan struct{})
	oldPoll := gameEndPoll
	gameEndPoll = 5 * time.Millisecond
	gameEndMu.Lock()
	gameEndNotified = ended
	gameEndMu.Unlock()
	t.Cleanup(func() {
		gameEndPoll = oldPoll
		gameEndMu.Lock()
		gameEndNotified = nil
		gameEndMu.Unlock()
		mapPrepTimerMu.Lock()
		if mapPrepTimer != nil {
			mapPrepTimer.Stop()
		}
		mapPrepTimerMu.Unlock()
	})
	return ended
}

func waitGameEnd(t *testing.T, ended chan struct{}) {
	t.Helper()
	select {
	case <-ended:
	case <-time.After(3 * time.Second):
		t.Fatal("the after-game waiter never finished")
	}
}

func mapPrepScheduled() bool {
	mapPrepTimerMu.Lock()
	defer mapPrepTimerMu.Unlock()
	return mapPrepTimer != nil && mapPrepTimer.Stop()
}

// The post-game cleanup arrived while League's process was still closing
// (22:11:54): it is declined, then runs once the process is gone, and the
// map preparation follows.
func TestDeclinedCleanupRunsWhenTheGameEnds(t *testing.T) {
	stubs := stubLateChange(t, true, ltk.Prepared{}, matchTiming{}, 15*time.Second)
	var game atomic.Bool
	game.Store(true)
	gameRunning = game.Load
	ended := expectGameEndWaiter(t)
	conn, peer := loopbackClient(t)
	handleCleanupRequest(conn)
	if msg := readStatus(t, peer); msg.Status != "cleanupDeclined" {
		t.Fatalf("status=%+v", msg)
	}
	time.Sleep(30 * time.Millisecond)
	if stubs.disarms != 0 {
		t.Fatal("disarmed while League ran")
	}
	game.Store(false)
	waitGameEnd(t, ended)
	if stubs.disarms != 1 || stubs.stops != 0 {
		t.Fatalf("disarms=%d stops=%d, want one disarm after the game", stubs.disarms, stubs.stops)
	}
	if !mapPrepScheduled() {
		t.Fatal("no map preparation after the game")
	}
}

// A set armed after the game ended is for the next match; the late cleanup
// must not take it away.
func TestDeclinedCleanupSparesTheNextMatchSet(t *testing.T) {
	stubs := stubLateChange(t, true, ltk.Prepared{Armed: true}, matchTiming{}, 15*time.Second)
	var game atomic.Bool
	game.Store(true)
	gameRunning = game.Load
	ended := expectGameEndWaiter(t)
	conn, peer := loopbackClient(t)
	handleCleanupRequest(conn)
	readStatus(t, peer)
	game.Store(false)
	waitGameEnd(t, ended)
	if stubs.disarms != 0 {
		t.Fatal("the late cleanup disarmed the set for the next match")
	}
}
