package server

// What waits for the League game process to end.
//
// The plugin asks for cleanup at the post-game phases, but League's process
// can still be closing then (22:11:54 on the owner's PC), and a cleanup is
// declined while it exists. Nothing asked again, so the patcher stayed armed
// with the finished match's set and the after-game map preparation never
// ran. An ame started mid-match got no cleanup at all: its plugin had no
// skins on record to clean up. Both now wait for the game to end here.

import (
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/display"
)

// gameEndPoll is how often the waiter looks for the game process; a
// variable for tests.
var gameEndPoll = 2 * time.Second

// gameEndWaitLimit bounds one wait (a long match plus a slow exit).
const gameEndWaitLimit = 3 * time.Hour

var (
	gameEndMu       sync.Mutex
	gameEndWaiting  bool
	gameEndCleanup  bool          // a cleanup was declined during the match
	gameEndNotified chan struct{} // closed after each finished wait, for tests
)

// afterGame runs the post-game work once League's process is gone: the
// declined cleanup when cleanup is set, and the map preparation. One waiter
// serves every request.
func afterGame(cleanup bool) {
	gameEndMu.Lock()
	if cleanup {
		gameEndCleanup = true
	}
	if gameEndWaiting {
		gameEndMu.Unlock()
		return
	}
	gameEndWaiting = true
	gameEndMu.Unlock()
	go waitForGameEnd()
}

func waitForGameEnd() {
	deadline := time.Now().Add(gameEndWaitLimit)
	for gameRunning() && time.Now().Before(deadline) {
		time.Sleep(gameEndPoll)
	}
	gameEndMu.Lock()
	cleanup := gameEndCleanup
	gameEndWaiting, gameEndCleanup = false, false
	notified := gameEndNotified
	gameEndMu.Unlock()
	defer func() {
		if notified != nil {
			close(notified)
		}
	}()
	if gameRunning() {
		return
	}
	if cleanup {
		// A set armed after the game ended belongs to the next match.
		if state := preparedState(); state.Armed || state.Pending {
			display.Log("LTK: League closed; a skin set is already prepared for the next match, so the cleanup declined during the match is dropped")
		} else {
			display.Log("LTK: League closed; disarming the patcher (cleanup declined during the match)")
			resetSelection(disarmRuntime)
		}
	}
	scheduleMapPrep("after game", mapPrepPostGame)
}
