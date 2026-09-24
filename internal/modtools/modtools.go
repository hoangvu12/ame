// Package modtools dispatches skin overlay work to the bundled runtime.
package modtools

import (
	"time"

	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/ltk"
)

// KillModTools stops the runtime host.
func KillModTools() {
	ltk.Stop()
}

// RunMkOverlay builds the overlay for the given mods.
func RunMkOverlay(modsDir, overlayDir, gameDir, modName string) (bool, int) {
	if err := ltk.Build(modsDir, overlayDir, gameDir, modName); err != nil {
		display.Log(err.Error())
		return false, 1
	}
	return true, 0
}

// RunOverlay starts the runtime host for an overlay prefix.
func RunOverlay(overlayDir string) error {
	return ltk.Start(overlayDir)
}

// Exists reports whether the bundled runtime can be prepared.
func Exists() bool {
	_, err := ltk.Prepare()
	return err == nil
}

// IsRunning reports whether the runtime host is running.
func IsRunning() bool {
	return ltk.Running()
}

// IsHooked reports whether the runtime has attached to a game.
func IsHooked() bool {
	return ltk.Hooked()
}

// WaitForHook blocks until the runtime attaches to a game or the timeout
// expires.
func WaitForHook(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ltk.Hooked() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
