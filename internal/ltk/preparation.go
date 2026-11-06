package ltk

import (
	"fmt"
	"os"
)

// BeginPreparation readies the overlay directory for a skin preparation. It
// no longer starts the patcher early: a scan started on the library the
// last match left behind ends at ame's first edit anyway (every edit
// restarts the scan), and the manager's build of that stale library raced
// the edit — reading an archive while it was being replaced failed the
// patcher start ("InvalidHeader") and could cost a full map WAD rebuild for
// a set nobody chose. Start installs the package first, then starts the
// patcher, so the first build is the one that matters. The returned release
// function is kept for the preparation flow.
func BeginPreparation(overlay string) (func() error, error) {
	if err := os.MkdirAll(overlay, 0755); err != nil {
		return nil, err
	}
	return func() error { return nil }, nil
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
