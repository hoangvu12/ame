package ltk

// Preparing the map copies ahead of time.
//
// A skin set only keeps map copies alive that already exist (maps.go), so
// ame makes them outside champion select: at startup, when the League
// client connects (often right after an update) and after each game. It
// installs a package holding only the keep-alive entries and runs the
// manager's own rebuild_overlay, which builds without starting the patcher,
// so no game is hooked while the copies are written.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
)

// Map preparation states reported to the plugin.
const (
	MapPrepIdle      = "idle"
	MapPrepPreparing = "preparing"
	MapPrepReady     = "ready"
	MapPrepFailed    = "failed"
	MapPrepSkipped   = "skipped"
	MapPrepOff       = "off"
)

const (
	// mapPrepMinFree: the three map copies take about 4.4 GB, plus room for
	// champion copies and League's own updates.
	mapPrepMinFree uint64 = 8 << 30
	// mapPrepRetryDelay is the back-off after a failed preparation.
	mapPrepRetryDelay = 10 * time.Minute
	// mapPrepArmWait bounds how long arming a skin set waits for a running
	// preparation before giving up on this attempt.
	mapPrepArmWait = 3 * time.Minute
)

// mapPrepTimeout bounds one preparation. A variable for tests.
var mapPrepTimeout = 20 * time.Minute

// mapPrepPoll is the interval between reads of the running rebuild.
var mapPrepPoll = time.Second

// ErrMapPrepBusy: a map preparation is still writing the copies, and a
// library edit now would replace the archive the rebuild reads.
var ErrMapPrepBusy = errors.New("the map copies are still being prepared")

// MapPrepStatus is what the plugin is told about the map copies.
type MapPrepStatus struct {
	State   string `json:"state"`
	Reason  string `json:"reason,omitempty"`
	Stage   string `json:"stage,omitempty"`
	Current int    `json:"current,omitempty"`
	Total   int    `json:"total,omitempty"`
	TookMs  int64  `json:"tookMs,omitempty"`
}

// OnMapPrep is told every change of the map preparation status.
var OnMapPrep func(MapPrepStatus)

var (
	mapPrepMu         sync.Mutex
	mapPrepStatus     = MapPrepStatus{State: MapPrepIdle}
	mapPrepRunning    bool
	mapPrepDone       chan struct{} // closed when the running preparation ends
	mapPrepRetryAfter time.Time
)

// CurrentMapPrep is the latest map preparation status.
func CurrentMapPrep() MapPrepStatus {
	mapPrepMu.Lock()
	defer mapPrepMu.Unlock()
	return mapPrepStatus
}

// MapPrepRunning reports whether a preparation is writing the map copies.
func MapPrepRunning() bool {
	mapPrepMu.Lock()
	defer mapPrepMu.Unlock()
	return mapPrepRunning
}

func publishMapPrep(status MapPrepStatus) {
	mapPrepMu.Lock()
	changed := status != mapPrepStatus
	mapPrepStatus = status
	mapPrepMu.Unlock()
	if notify := OnMapPrep; changed && notify != nil {
		notify(status)
	}
}

// RefreshMapPrepSetting reports the feature as off right after the setting
// is turned off; turning it on is answered by the next preparation.
func RefreshMapPrepSetting() {
	if !mapPrepEnabled() && !MapPrepRunning() {
		publishMapPrep(MapPrepStatus{State: MapPrepOff})
	}
}

// mapsPackagePath is the keep-alive-only package a preparation installs.
func mapsPackagePath() string {
	return filepath.Join(config.AmeDir, "overlay-maps.fantome")
}

// existingParent is path, or its nearest ancestor that exists.
func existingParent(path string) string {
	for {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return path
		}
		path = parent
	}
}

// PrepareMaps makes the manager's map copies unless they are ready. Only
// one preparation runs at a time; a call while one runs returns at once.
// The caller makes sure no skin set is being prepared for a match: the
// patcher must stay stopped and the library untouched while the copies are
// written.
func PrepareMaps(reason string) error {
	mapPrepMu.Lock()
	if mapPrepRunning {
		mapPrepMu.Unlock()
		return nil
	}
	if wait := time.Until(mapPrepRetryAfter); wait > 0 {
		mapPrepMu.Unlock()
		display.Log(fmt.Sprintf("LTK: map preparation skipped (reason=%s): retrying after a failure in %s", reason, wait.Round(time.Second)))
		return nil
	}
	mapPrepRunning = true
	done := make(chan struct{})
	mapPrepDone = done
	mapPrepMu.Unlock()
	defer func() {
		mapPrepMu.Lock()
		mapPrepRunning = false
		close(done)
		mapPrepMu.Unlock()
		scheduleIdleStop()
	}()

	skip := func(why string) error {
		display.Log(fmt.Sprintf("LTK: map preparation skipped (reason=%s): %s", reason, why))
		publishMapPrep(MapPrepStatus{State: MapPrepSkipped, Reason: why})
		return nil
	}
	if !mapPrepEnabled() {
		publishMapPrep(MapPrepStatus{State: MapPrepOff})
		return nil
	}
	game := findGameDir()
	if game == "" {
		return skip("League not found")
	}
	ok, why := MapsPrepared(game)
	if ok {
		display.Log(fmt.Sprintf("LTK: maps ready, skipping (reason=%s)", reason))
		publishMapPrep(MapPrepStatus{State: MapPrepReady})
		return nil
	}
	display.Log(fmt.Sprintf("LTK: preparing maps (reason=%s): %s", reason, why))
	if gameProcessRunning() {
		return skip("League is running")
	}
	if root := managerOverlayRoot(); root != "" {
		if free, err := diskFree(existingParent(root)); err == nil && free < mapPrepMinFree {
			return skip(fmt.Sprintf("only %.1f GB free on the manager's drive", float64(free)/(1<<30)))
		}
	}

	started := time.Now()
	publishMapPrep(MapPrepStatus{State: MapPrepPreparing, Reason: reason})
	err := runMapRebuild(game, started)
	if errors.Is(err, errMapPrepArmed) || errors.Is(err, ErrGameRunning) {
		return skip(err.Error())
	}
	if err == nil {
		if ok, why := MapsPrepared(game); !ok {
			err = fmt.Errorf("the rebuild ended but the copies are not ready: %s", why)
		}
	}
	took := time.Since(started)
	if err != nil {
		mapPrepMu.Lock()
		mapPrepRetryAfter = time.Now().Add(mapPrepRetryDelay)
		mapPrepMu.Unlock()
		display.Log(fmt.Sprintf("LTK: map preparation failed after %s (reason=%s): %v", took.Round(100*time.Millisecond), reason, err))
		publishMapPrep(MapPrepStatus{State: MapPrepFailed, Reason: err.Error()})
		return err
	}
	display.Log(fmt.Sprintf("LTK: maps prepared in %s (reason=%s)", took.Round(100*time.Millisecond), reason))
	publishMapPrep(MapPrepStatus{State: MapPrepReady, TookMs: took.Milliseconds()})
	return nil
}

var errMapPrepArmed = errors.New("a skin set is armed")

// runMapRebuild has the manager write the map copies. A variable so tests
// never reach a manager.
var runMapRebuild = func(game string, started time.Time) error {
	m, err := startMapRebuild(game)
	if err != nil {
		return err
	}
	return awaitMapRebuild(m, started)
}

// startMapRebuild installs the keep-alive-only package and starts the
// manager's full rebuild. It holds managerMu only for these short calls.
func startMapRebuild(game string) (*managerSession, error) {
	managerMu.Lock()
	defer managerMu.Unlock()
	if gameProcessRunning() {
		return nil, ErrGameRunning
	}
	m, err := ensureManagerLocked()
	if err != nil {
		return nil, err
	}
	status, err := m.cdp.patcherStatus()
	if err != nil {
		return nil, fmt.Errorf("patcher status: %w", err)
	}
	if status.Running {
		// The patcher runs only with a skin set ame armed; rebuild_overlay
		// refuses while it runs anyway.
		return nil, errMapPrepArmed
	}
	pkg := mapsPackagePath()
	var entries []packageEntry
	for _, wad := range keepAliveMaps {
		if present, _ := mapCopyState(game, "", wad); present {
			entries = append(entries, keepAliveEntry(wad))
		}
	}
	staged := pkg + ".new"
	if err := mergeFantomes(nil, staged, entries); err != nil {
		os.Remove(staged)
		return nil, fmt.Errorf("keep-alive package: %w", err)
	}
	if err := os.Rename(staged, pkg); err != nil {
		return nil, err
	}
	digest, err := packageDigest(pkg)
	if err != nil {
		return nil, err
	}
	if !m.matchesPackage(pkg, digest) {
		// The patcher is stopped, so this edit rebuilds nothing by itself.
		if err := replaceAmeMod(m, pkg, digest, false, ""); err != nil {
			return nil, err
		}
	}
	if err := m.cdp.startOverlayRebuild(); err != nil {
		return nil, err
	}
	return m, nil
}

// awaitMapRebuild follows the rebuild with short polls, so other manager
// calls are never held behind it.
func awaitMapRebuild(m *managerSession, started time.Time) error {
	deadline := started.Add(mapPrepTimeout)
	failures := 0
	for {
		time.Sleep(mapPrepPoll)
		if !m.alive() {
			return fmt.Errorf("the manager stopped during the rebuild")
		}
		state, err := m.cdp.overlayRebuildState()
		if err != nil {
			if failures++; failures >= 5 {
				return fmt.Errorf("rebuild poll: %w", err)
			}
			continue
		}
		failures = 0
		switch state.State {
		case "done":
			return nil
		case "failed":
			return fmt.Errorf("rebuild_overlay: %s", errorText(state.Error))
		case "missing":
			return fmt.Errorf("the manager page lost the rebuild (reloaded)")
		}
		if p := state.Progress; p != nil {
			publishMapPrep(MapPrepStatus{State: MapPrepPreparing, Stage: p.Stage, Current: p.Current, Total: p.Total})
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("the rebuild did not finish within %s", mapPrepTimeout)
		}
	}
}

// waitForMapPrep holds an arm back while a preparation writes the copies:
// a library edit would replace the archive the rebuild reads, and the
// preparation does the copying the arm needs anyway. League starting ends
// the wait, since no edit may follow then.
func waitForMapPrep() error {
	mapPrepMu.Lock()
	running, done := mapPrepRunning, mapPrepDone
	mapPrepMu.Unlock()
	if !running {
		return nil
	}
	display.Log("LTK: waiting for the map preparation before arming the skin set")
	started := time.Now()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			display.Log(fmt.Sprintf("LTK: map preparation finished after a %s wait", time.Since(started).Round(100*time.Millisecond)))
			return nil
		case <-tick.C:
		}
		if gameProcessRunning() {
			return ErrGameRunning
		}
		if time.Since(started) > mapPrepArmWait {
			return ErrMapPrepBusy
		}
	}
}
