package server

// Skin set changes close to a game launch.
//
// LTK Manager rebuilds its overlay after every library edit and starts a new
// patcher scan once the rebuild ends; the injected DLL refuses a game created
// before that scan ("joined too late"). An edit therefore only helps if its
// rebuild ends before League starts, and when it does not, the match loses
// every skin, including the set that was already armed for it. Pausing the
// game cannot help: the DLL compares the process creation time, which a pause
// does not move. So once a set is armed for the coming match, a change is
// made only while champion select has clearly more time left than this PC
// needs to rebuild; otherwise the armed set stays and the change waits for
// the next match. Nothing is ever edited while League's process exists.

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/lcu"
	"github.com/hoangvu12/ame/internal/ltk"
	"github.com/hoangvu12/ame/internal/modtools"
)

// Variables so tests never depend on, or act on, a client, manager or game on
// the test machine.
var (
	gameRunning    = ltk.GameRunning
	preparedState  = ltk.PreparedState
	rebuildBudget  = ltk.RebuildBudget
	patcherWaiting = ltk.WaitingForGame
	startRuntime   = modtools.RunOverlay
	stopRuntime    = modtools.KillModTools
)

type launchStage int

const (
	stageUnknown     launchStage = iota // the client did not answer
	stageBefore                         // no champion select: no launch in sight
	stageChampSelect                    // the timer says how much time is left
	stageLaunching                      // champion select ended; League is starting or running
)

// matchTiming is what the League client says about the coming launch.
type matchTiming struct {
	stage launchStage
	// remaining is a lower bound on the time until League starts, valid
	// during champion select when timed is set.
	remaining time.Duration
	timed     bool
	detail    string
}

// banPickFloor is the least time that still follows the planning and pick
// phases: champion select always closes with its finalization phase, and
// League's process appears seconds after that.
const banPickFloor = 10 * time.Second

// champSelectTiming converts the champion select timer into a lower bound
// on the time until League starts. The timer's values are as of the
// client's last update, so the time since then is subtracted.
func champSelectTiming(timer lcu.ChampSelectTimer, now time.Time) matchTiming {
	t := matchTiming{stage: stageChampSelect, detail: timer.Phase}
	if timer.IsInfinite {
		return t
	}
	left := time.Duration(timer.AdjustedTimeLeftInPhase) * time.Millisecond
	if timer.InternalNowInEpochMs > 0 {
		if elapsed := now.Sub(time.UnixMilli(timer.InternalNowInEpochMs)); elapsed > 0 {
			left -= elapsed
		}
	}
	if left < 0 {
		left = 0
	}
	switch strings.ToUpper(timer.Phase) {
	case "PLANNING", "BAN_PICK":
		t.remaining = left + banPickFloor
	case "FINALIZATION":
		t.remaining = left
	default: // GAME_STARTING, or a phase this code does not know
		t.remaining = 0
	}
	t.timed = true
	return t
}

var probeMatchTiming = func() matchTiming {
	phase, err := lcu.GameflowPhase()
	if err != nil {
		return matchTiming{stage: stageUnknown, detail: err.Error()}
	}
	switch phase {
	case "ChampSelect":
		timer, err := lcu.GetChampSelectTimer()
		if err != nil {
			return matchTiming{stage: stageChampSelect, detail: "timer: " + err.Error()}
		}
		return champSelectTiming(timer, time.Now())
	case "GameStart", "InProgress", "Reconnect":
		return matchTiming{stage: stageLaunching, detail: phase}
	}
	return matchTiming{stage: stageBefore, detail: phase}
}

type changeVerdict int

const (
	changeAllowed changeVerdict = iota
	changeKept                  // the set armed for this match stays; the change waits
	changeRefused               // League runs and no usable set is armed for it
)

// changeContext is everything decideChange weighs.
type changeContext struct {
	gameRunning bool
	armed       bool // a set for the selected champion scans for this match
	pending     bool // a set for the selected champion is still rebuilding
	pendingFor  time.Duration
	budget      time.Duration // this PC's measured edit-to-scan time
	timing      matchTiming
}

func roundSeconds(d time.Duration) time.Duration { return d.Round(time.Second) }

// decideChange decides whether a new skin set may replace what the manager
// holds for the coming match.
func decideChange(c changeContext) (changeVerdict, string) {
	if c.gameRunning {
		switch {
		case c.armed:
			return changeKept, "League is already running"
		case c.pending:
			return changeRefused, "League started before the skin set finished rebuilding"
		}
		return changeRefused, "League is already running"
	}
	if !c.armed && !c.pending {
		// Nothing prepared for this champion: the change cannot cost anything.
		return changeAllowed, ""
	}
	need := c.budget
	if c.pending && c.pendingFor < c.budget {
		// The running rebuild ends first; this change rebuilds after it.
		need += c.budget - c.pendingFor
	}
	switch c.timing.stage {
	case stageBefore:
		return changeAllowed, ""
	case stageLaunching:
		return changeKept, "champion select has ended"
	case stageUnknown:
		return changeKept, "the League client did not say how long champion select has left"
	}
	if !c.timing.timed {
		return changeKept, "the champion select timer is unavailable"
	}
	if c.timing.remaining < need {
		return changeKept, fmt.Sprintf("%s left before the game, this PC needs about %s to rebuild skins", roundSeconds(c.timing.remaining), roundSeconds(need))
	}
	return changeAllowed, ""
}

// preparedChampionID is the champion the set handed to the manager was built
// for. A set prepared for another champion is not worth protecting. Guarded
// by stateMu.
var preparedChampionID string

func samePreparedChampion(championID string) bool {
	stateMu.Lock()
	prepared := preparedChampionID
	stateMu.Unlock()
	if prepared == "" || championID == "" || championID == "0" {
		return true
	}
	return prepared == championID
}

func setPreparedChampion(championID string) {
	stateMu.Lock()
	preparedChampionID = championID
	stateMu.Unlock()
}

// lateChange weighs a change for championID against what the manager holds.
// The client is asked about timing only when an armed set is at stake.
func lateChange(championID string) (changeVerdict, string) {
	state := preparedState()
	same := samePreparedChampion(championID)
	c := changeContext{
		gameRunning: gameRunning(),
		armed:       state.Armed && same,
		pending:     state.Pending && same,
		pendingFor:  state.PendingFor,
	}
	if !c.gameRunning && (c.armed || c.pending) {
		c.budget = rebuildBudget()
		c.timing = probeMatchTiming()
	}
	return decideChange(c)
}

// reportBlockedChange tells the plugin why a change did not reach the
// manager. "kept" says the armed set stays for this match; the plugin shows
// it instead of an apply success.
func reportBlockedChange(conn *websocket.Conn, source string, verdict changeVerdict, reason string) {
	switch verdict {
	case changeKept:
		display.Log(fmt.Sprintf("%s: keeping the skins prepared for this match (%s); the change applies next match", source, reason))
		sendStatus(conn, "kept", fmt.Sprintf("Keeping the skins prepared for this match (%s). Changes this late apply next match.", reason))
	case changeRefused:
		display.Log(fmt.Sprintf("%s: skins cannot change now: %s", source, reason))
		sendStatus(conn, "error", fmt.Sprintf("%s, so this match runs without these skins. They apply from the next match.", reason))
	}
}

// verdictMessages explain a DLL verdict that kept skins out of a match.
var verdictMessages = map[string]string{
	"too-late":    "This match started before the skin set was ready, so it runs without the new skins. They apply from the next match.",
	"end-of-life": "The skin patcher does not support this game version yet, so this match runs without skins.",
	"disabled":    "The skin patcher rejected the skin files, so this match runs without skins.",
	"hook-failed": "The skin patcher could not hook the game, so this match runs without skins.",
}

// reportOverlayVerdict surfaces the DLL's verdict on each game: only "live"
// confirms skins reached a match.
func reportOverlayVerdict(pid uint64, verdict string) {
	if verdict == "live" {
		display.SetOverlayKey("display.value.overlay_active", nil)
		return
	}
	message, ok := verdictMessages[verdict]
	if !ok {
		return
	}
	display.SetOverlayKey("display.value.overlay_inactive", nil)
	broadcastStatus("error", message)
}

// handleCleanupRequest serves the plugin's cleanup message. While League
// runs, the patcher serves that match: stopping it now cannot remove skins
// from the game, and the plugin also asks after transient phases (a client
// that closes during the game and comes back, a phase that briefly reads
// "None"), so the request is declined instead. Uninstall, the stuck-game
// button and shutting ame down still clean up unconditionally.
func handleCleanupRequest(conn *websocket.Conn) {
	if gameRunning() {
		display.Log("LTK: cleanup request ignored while League is running; the patcher keeps serving this match")
		// Its own status, so the plugin can undo the optimistic "skins off"
		// it set when sending the request; a "kept" can also answer a
		// selection.
		sendStatus(conn, "cleanupDeclined", "Skins cannot change during a match. Changes apply next match.")
		return
	}
	HandleCleanup()
	scheduleMapPrep("after game", mapPrepPostGame)
}
