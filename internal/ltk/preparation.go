package ltk

import (
	"fmt"
	"os"

	"github.com/hoangvu12/ame/internal/display"
)

// BeginPreparation prewarms the LTK Manager during champion select so the
// first skin application does not pay the manager's WebView startup cost.
// The returned release function is kept for the preparation flow; there is
// no scanner to pause anymore — the manager owns the injection host and
// never reads ame's overlay directory, so the overlay can be swapped freely.
func BeginPreparation(overlay string) (func() error, error) {
	if err := os.MkdirAll(overlay, 0755); err != nil {
		return nil, err
	}
	if err := prewarmManager(); err != nil {
		return nil, err
	}
	return func() error { return nil }, nil
}

// prewarmManager ensures a manager session exists without starting a patcher.
func prewarmManager() error {
	managerMu.Lock()
	defer managerMu.Unlock()
	if m := manager; m != nil && m.alive() {
		if _, err := m.cdp.patcherStatus(); err == nil {
			return nil
		}
	}
	dir, err := Prepare()
	if err != nil {
		return err
	}
	_, err = ensureManagerLocked(dir)
	return err
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
