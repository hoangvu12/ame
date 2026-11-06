package ltk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// prepHarness stubs everything PrepareMaps reaches outside the package.
type prepHarness struct {
	*fakeMaps
	patched  time.Time
	rebuilds int
	statuses []MapPrepStatus
}

func newPrepHarness(t *testing.T) *prepHarness {
	t.Helper()
	h := &prepHarness{fakeMaps: newFakeMaps(t, true), patched: time.Now().Add(-time.Hour)}
	for _, wad := range keepAliveMaps {
		h.gameMap(t, wad, 10, h.patched)
	}
	oldFind, oldGame, oldDisk, oldRebuild, oldNotify := findGameDir, gameProcessRunning, diskFree, runMapRebuild, OnMapPrep
	findGameDir = func() string { return h.game }
	gameProcessRunning = func() bool { return false }
	diskFree = func(string) (uint64, error) { return 100 << 30, nil }
	runMapRebuild = func(string, time.Time) error {
		h.rebuilds++
		h.writeCopies(t)
		return nil
	}
	OnMapPrep = func(s MapPrepStatus) { h.statuses = append(h.statuses, s) }
	t.Cleanup(func() {
		findGameDir, gameProcessRunning, diskFree, runMapRebuild, OnMapPrep = oldFind, oldGame, oldDisk, oldRebuild, oldNotify
		mapPrepMu.Lock()
		mapPrepRetryAfter = time.Time{}
		mapPrepStatus = MapPrepStatus{State: MapPrepIdle}
		mapPrepMu.Unlock()
	})
	return h
}

func (h *prepHarness) writeCopies(t *testing.T) {
	for _, wad := range keepAliveMaps {
		h.copyMap(t, wad, 20, time.Now())
	}
}

func (h *prepHarness) last() MapPrepStatus {
	if len(h.statuses) == 0 {
		return MapPrepStatus{}
	}
	return h.statuses[len(h.statuses)-1]
}

func TestPrepareMapsSuccess(t *testing.T) {
	h := newPrepHarness(t)
	if err := PrepareMaps("startup"); err != nil {
		t.Fatal(err)
	}
	if h.rebuilds != 1 {
		t.Fatalf("rebuilds=%d, want 1", h.rebuilds)
	}
	if h.statuses[0].State != MapPrepPreparing || h.last().State != MapPrepReady {
		t.Fatalf("statuses=%v", h.statuses)
	}
	if MapPrepRunning() {
		t.Fatal("preparation still marked running")
	}
	// Fresh copies: the next trigger reaches no manager.
	if err := PrepareMaps("client connected"); err != nil {
		t.Fatal(err)
	}
	if h.rebuilds != 1 {
		t.Fatalf("fresh maps were prepared again: rebuilds=%d", h.rebuilds)
	}
}

func TestPrepareMapsSkips(t *testing.T) {
	cases := map[string]func(h *prepHarness){
		"game running": func(h *prepHarness) { gameProcessRunning = func() bool { return true } },
		"low disk":     func(h *prepHarness) { diskFree = func(string) (uint64, error) { return 2 << 30, nil } },
		"armed": func(h *prepHarness) {
			runMapRebuild = func(string, time.Time) error { h.rebuilds++; return errMapPrepArmed }
		},
		"no League": func(h *prepHarness) { findGameDir = func() string { return "" } },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			h := newPrepHarness(t)
			setup(h)
			if err := PrepareMaps("test"); err != nil {
				t.Fatal(err)
			}
			if got := h.last().State; got != MapPrepSkipped {
				t.Fatalf("state=%q, want skipped (%v)", got, h.statuses)
			}
			if ok, _ := MapsPrepared(h.game); ok {
				t.Fatal("copies appeared on a skipped preparation")
			}
		})
	}
}

func TestPrepareMapsOff(t *testing.T) {
	h := newPrepHarness(t)
	mapPrepEnabled = func() bool { return false }
	if err := PrepareMaps("startup"); err != nil {
		t.Fatal(err)
	}
	if h.rebuilds != 0 || h.last().State != MapPrepOff {
		t.Fatalf("feature off: rebuilds=%d statuses=%v", h.rebuilds, h.statuses)
	}
}

func TestPrepareMapsFailureBacksOff(t *testing.T) {
	h := newPrepHarness(t)
	runMapRebuild = func(string, time.Time) error { h.rebuilds++; return errors.New("disk full") }
	if err := PrepareMaps("startup"); err == nil {
		t.Fatal("failure not reported")
	}
	if h.last().State != MapPrepFailed {
		t.Fatalf("statuses=%v", h.statuses)
	}
	if err := PrepareMaps("after game"); err != nil {
		t.Fatal(err)
	}
	if h.rebuilds != 1 {
		t.Fatalf("retried before the back-off ended: rebuilds=%d", h.rebuilds)
	}
	mapPrepMu.Lock()
	mapPrepRetryAfter = time.Now().Add(-time.Second)
	mapPrepMu.Unlock()
	runMapRebuild = func(string, time.Time) error { h.rebuilds++; h.writeCopies(t); return nil }
	if err := PrepareMaps("after game"); err != nil {
		t.Fatal(err)
	}
	if h.rebuilds != 2 || h.last().State != MapPrepReady {
		t.Fatalf("retry after back-off: rebuilds=%d statuses=%v", h.rebuilds, h.statuses)
	}
}

func TestPrepareMapsChecksCopiesAfterRebuild(t *testing.T) {
	h := newPrepHarness(t)
	runMapRebuild = func(string, time.Time) error { return nil } // wrote nothing
	if err := PrepareMaps("startup"); err == nil {
		t.Fatal("a rebuild that left no copies counted as prepared")
	}
	if h.last().State != MapPrepFailed {
		t.Fatalf("statuses=%v", h.statuses)
	}
}

func TestArmWaitsForRunningPreparation(t *testing.T) {
	h := newPrepHarness(t)
	release := make(chan struct{})
	entered := make(chan struct{})
	runMapRebuild = func(string, time.Time) error {
		close(entered)
		<-release
		h.writeCopies(t)
		return nil
	}
	prepared := make(chan error, 1)
	go func() { prepared <- PrepareMaps("startup") }()
	<-entered
	waited := make(chan error, 1)
	go func() { waited <- waitForMapPrep() }()
	select {
	case err := <-waited:
		t.Fatalf("arm did not wait for the preparation: %v", err)
	case <-time.After(400 * time.Millisecond):
	}
	close(release)
	if err := <-waited; err != nil {
		t.Fatal(err)
	}
	if err := <-prepared; err != nil {
		t.Fatal(err)
	}
	if err := waitForMapPrep(); err != nil {
		t.Fatalf("no preparation running, yet the arm waited: %v", err)
	}
}

func TestArmWaitEndsWhenLeagueStarts(t *testing.T) {
	newPrepHarness(t)
	release := make(chan struct{})
	entered := make(chan struct{})
	runMapRebuild = func(string, time.Time) error {
		close(entered)
		<-release
		return nil
	}
	prepared := make(chan error, 1)
	go func() { prepared <- PrepareMaps("startup") }()
	defer func() { close(release); <-prepared }()
	<-entered
	gameProcessRunning = func() bool { return true }
	if err := waitForMapPrep(); !errors.Is(err, ErrGameRunning) {
		t.Fatalf("err=%v, want ErrGameRunning", err)
	}
}

func TestRemoveMapCopies(t *testing.T) {
	h := newPrepHarness(t)
	h.writeCopies(t)
	other := filepath.Join(gameMapsDir(h.overlay), "Map22.wad.client")
	writeSized(t, other, 1, time.Now())
	RemoveMapCopies()
	for _, wad := range keepAliveMaps {
		if _, err := os.Stat(filepath.Join(gameMapsDir(h.overlay), wad)); !os.IsNotExist(err) {
			t.Fatalf("%s copy survived: %v", wad, err)
		}
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("a copy ame never kept alive was removed: %v", err)
	}
}
