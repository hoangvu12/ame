package ltk

import (
	"fmt"
	"os"
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

// publishPackage swaps the staged skin package into place atomically: the
// previous package stays intact until the new one is complete (an aborted
// build's staged file is simply overwritten next time), and a game window
// refuses publication: the armed package for that match is kept. The
// extracted overlay ame used to build beside the package is gone — the
// manager builds its own overlay from the installed package.
func publishPackage(staged, pkg string) error {
	if gameWindowExists() {
		return fmt.Errorf("a game window already exists; keeping the previous skin package for this match")
	}
	return os.Rename(staged, pkg)
}
