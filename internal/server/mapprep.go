package server

// When ame asks the manager to prepare its map copies (see ltk/mapprep.go):
// after setup, after the League client connects and after each game, each
// time only if no skin set is armed or being prepared and no match is near.

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/ltk"
	"github.com/hoangvu12/ame/internal/modtools"
)

// Delays before a trigger fires: the client is busy for a while after it
// connects (often right after an update), and a finished game is followed
// by the end-of-game screens.
const (
	mapPrepStartupDelay = 5 * time.Second
	mapPrepClientDelay  = 30 * time.Second
	mapPrepPostGame     = 15 * time.Second
	mapPrepBusyRetry    = 30 * time.Second
)

// Variables so tests never reach a manager.
var (
	prepareMaps    = ltk.PrepareMaps
	runtimeRunning = modtools.IsRunning
)

var (
	mapPrepTimerMu sync.Mutex
	mapPrepTimer   *time.Timer
)

// scheduleMapPrep asks for a preparation after delay. A later request
// replaces a pending one, so bursts of triggers prepare once.
func scheduleMapPrep(reason string, delay time.Duration) {
	mapPrepTimerMu.Lock()
	defer mapPrepTimerMu.Unlock()
	if mapPrepTimer != nil {
		mapPrepTimer.Stop()
	}
	mapPrepTimer = time.AfterFunc(delay, func() { runMapPrep(reason) })
}

// mapPrepBlocker says why a preparation must not start now, or "".
func mapPrepBlocker() string {
	if !ready.Load() {
		return "ame is still starting"
	}
	if gameRunning() {
		return "League is running"
	}
	if runtimeRunning() {
		return "a skin set is armed"
	}
	switch probeMatchTiming().stage {
	case stageChampSelect:
		return "champion select is in progress"
	case stageLaunching:
		return "a match is starting"
	}
	return ""
}

func runMapPrep(reason string) {
	if why := mapPrepBlocker(); why != "" {
		display.Log(fmt.Sprintf("LTK: map preparation not started (reason=%s): %s", reason, why))
		return
	}
	// A selection being prepared right now goes first; it arms the patcher,
	// and the post-game trigger comes back for the maps.
	if !skinPreparationMu.TryLock() {
		scheduleMapPrep(reason, mapPrepBusyRetry)
		return
	}
	skinPreparationMu.Unlock()
	if err := prepareMaps(reason); err != nil {
		display.Log("LTK: map preparation: " + err.Error())
	}
}

// MapPrepMessage tells the plugin how the map copies are doing.
type MapPrepMessage struct {
	Type string `json:"type"`
	ltk.MapPrepStatus
}

func mapPrepMessage(status ltk.MapPrepStatus) []byte {
	data, _ := json.Marshal(MapPrepMessage{Type: "mapPrep", MapPrepStatus: status})
	return data
}

func broadcastMapPrep(status ltk.MapPrepStatus) {
	data := mapPrepMessage(status)
	for _, conn := range connectedClients() {
		writeClientMessage(conn, websocket.TextMessage, data)
	}
}

func sendMapPrep(conn *websocket.Conn) {
	writeClientMessage(conn, websocket.TextMessage, mapPrepMessage(ltk.CurrentMapPrep()))
}
