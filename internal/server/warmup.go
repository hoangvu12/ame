package server

// Starting the manager before the first skin pick (see ltk.WarmManager):
// when setup ends with the League client open, and when the plugin
// connects. A pick then finds the manager running instead of paying for a
// cold launch inside champion select.

import (
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/lcu"
	"github.com/hoangvu12/ame/internal/ltk"
)

// Variables so tests never reach a manager or a client.
var (
	warmManager   = ltk.WarmManager
	clientRunning = lcu.IsClientRunning
)

// warmManagerSoon starts the manager in the background unless setup is
// still running or League is.
func warmManagerSoon(reason string) {
	if !ready.Load() || gameRunning() {
		return
	}
	go func() {
		if err := warmManager(reason); err != nil {
			display.Log("LTK: manager warm-up: " + err.Error())
		}
	}()
}
