package ltk

import (
	"fmt"
	"os"

	"github.com/hoangvu12/ame/internal/display"
)

// BeginPreparation arms the patcher during champion select: the LTK Manager
// session is brought up and its patcher starts scanning with whatever the
// library holds. The session's start time is what the injected DLL compares
// against the game's creation — a patcher session that begins only when the
// skin is applied starts after the game and the DLL stays inert ("joined too
// late") — so the session must be live before the game launches, spanning
// the whole preparation. The returned release function is kept for the
// preparation flow; there is no scanner to pause anymore: the manager owns
// the injection host and never reads ame's overlay directory, so ame's
// overlay is swapped freely while the patcher runs.
func BeginPreparation(overlay string) (func() error, error) {
	if err := os.MkdirAll(overlay, 0755); err != nil {
		return nil, err
	}
	if err := armPatcher(); err != nil {
		return nil, err
	}
	return func() error { return nil }, nil
}

// armPatcher ensures a manager session exists and its patcher is scanning.
// The library keeps the last session's skin package installed and enabled,
// so the patcher arms with it until Start swaps in the newly built one.
func armPatcher() error {
	managerMu.Lock()
	defer managerMu.Unlock()
	if m := manager; m != nil && m.alive() {
		if status, err := m.cdp.patcherStatus(); err == nil && status.Running {
			return nil
		}
	}
	m, err := ensureManagerLocked()
	if err != nil {
		return err
	}
	if status, err := m.cdp.patcherStatus(); err == nil && status.Running {
		return nil
	}
	m.cdp.clearBridgeFailure()
	setSnapshot(func(s *managerSnapshot) { s.failure = "" })
	if err := m.cdp.startPatcher(); err != nil {
		return fmt.Errorf("LTK patcher start: %w", err)
	}
	return awaitPatcherStart(m)
}

// publishOverlay swaps the staged overlay into place, together with the
// staged skin package, and never leaves a mix of old and new behind. A game
// window refuses publication: the armed overlay for that match is kept.
func publishOverlay(staged, overlay, pkgFrom, pkgTo string) error {
	if gameWindowExists() {
		return fmt.Errorf("a game window already exists; keeping the previous overlay for this match")
	}
	backup := staged + "-previous"
	hadPrevious := false
	if _, err := os.Stat(overlay); err == nil {
		if err = os.Rename(overlay, backup); err != nil {
			return err
		}
		hadPrevious = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(staged, overlay); err != nil {
		if hadPrevious {
			if restoreErr := os.Rename(backup, overlay); restoreErr != nil {
				return fmt.Errorf("publish failed: %v; restore failed: %w (previous overlay at %s)", err, restoreErr, backup)
			}
		}
		return err
	}
	// The skin package follows the overlay it was built from. Without a new
	// package the previous one is removed: Start must never install a
	// package that does not match the published overlay.
	var pkgErr error
	if pkgFrom != "" {
		pkgErr = os.Rename(pkgFrom, pkgTo)
	} else {
		pkgErr = os.Remove(pkgTo)
	}
	if pkgErr != nil && !os.IsNotExist(pkgErr) {
		display.Log("LTK: skin package publish: " + pkgErr.Error())
	}
	if hadPrevious {
		if err := os.RemoveAll(backup); err != nil {
			display.Log("LTK: old overlay cleanup: " + err.Error())
		}
	}
	return nil
}
