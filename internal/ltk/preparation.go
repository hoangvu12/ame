package ltk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hoangvu12/ame/internal/display"
)

// BeginPreparation establishes the scan session before expensive generation.
// Pause only our owned host while its stable overlay prefix is being updated.
// Restarting the host would reset the session start, and the session must
// span generation so a game created meanwhile is still covered by it.
func BeginPreparation(overlay string) (func() error, error) {
	if err := os.MkdirAll(overlay, 0755); err != nil {
		return nil, err
	}
	if err := Start(overlay); err != nil {
		return nil, err
	}
	return pausePreparedScanner()
}

func pausePreparedScanner() (func() error, error) {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	h := active
	if h == nil {
		return nil, fmt.Errorf("LTK scanner is not running")
	}
	h.mu.Lock()
	ready := h.running && h.scanning && h.hookPID == 0 && h.failure == ""
	h.mu.Unlock()
	if !ready || h.resumeScanner != nil {
		return nil, fmt.Errorf("LTK attachment has started or preparation is already active")
	}
	resume, err := pauseOwnedScanner(h.cmd.Process)
	if err != nil {
		return nil, err
	}
	h.resumeScanner = resume
	display.Log(fmt.Sprintf("LTK: paused owned scanner pid=%d for preparation; preserving session start", h.cmd.Process.Pid))
	var once sync.Once
	var releaseErr error
	return func() error {
		once.Do(func() {
			lifecycle.Lock()
			defer lifecycle.Unlock()
			if active != h || h.resumeScanner == nil {
				releaseErr = fmt.Errorf("LTK preparation session was stopped")
				return
			}
			releaseErr = h.resumeScanner()
			h.resumeScanner = nil
			if releaseErr != nil {
				stopLocked()
				return
			}
			display.Log(fmt.Sprintf("LTK: resumed original scanner pid=%d after preparation", h.cmd.Process.Pid))
		})
		return releaseErr
	}, nil
}

func publishOverlay(staged, overlay string, owner *host) error {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	if active != owner {
		return fmt.Errorf("LTK preparation session changed or was cancelled; keeping previous overlay")
	}
	if active != nil {
		prefix, err := filepath.Abs(overlay)
		if err != nil {
			return err
		}
		if !strings.EqualFold(prefix, active.overlay) {
			return fmt.Errorf("overlay does not belong to the preparation session")
		}
		active.mu.Lock()
		ready := active.running && active.scanning && active.hookPID == 0 && active.failure == ""
		active.mu.Unlock()
		if !ready || active.resumeScanner == nil || gameWindowExists() {
			return fmt.Errorf("game attachment may have started; keeping the previous overlay for this match")
		}
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
	if hadPrevious {
		if err := os.RemoveAll(backup); err != nil {
			display.Log("LTK: old overlay cleanup: " + err.Error())
		}
	}
	return nil
}
