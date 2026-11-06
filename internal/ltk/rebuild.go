package ltk

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// rebuildGrace is how long an edit may go without a visible rebuild. The
// manager picks an edit up after its 750ms settle window (refresh.rs), and a
// rebuild that reuses every WAD can finish between two status reads, so the
// phase never shows "building".
const rebuildGrace = 3 * time.Second

// rebuildTracker follows one library edit until the manager's scan runs with
// it. Every edit stops the scan, rebuilds the overlay and starts a new scan
// session (injector.rs, thread.rs); the injected DLL refuses any game created
// before that new session ("joined too late"). A build already running when
// the edit was made read the old library, so it must end and the edit's own
// build follow. The phase flips to "patching" a few milliseconds before the
// host's "start scan", so armedAt leads the real scan start by that much.
type rebuildTracker struct {
	editAt     time.Time
	staleBuild bool      // the build running at edit time predates the edit
	building   bool      // a build that includes the edit was observed
	merged     bool      // the edit's build was hidden inside the stale one
	quietSince time.Time // start of the current stretch without a build
	armedAt    time.Time // the scan ran with the edit from here on
	measured   bool      // armedAt follows an observed build
	mapsReady  bool      // the map copies were prepared before the edit
}

func newRebuildTracker(editAt time.Time, phaseAtEdit string) *rebuildTracker {
	r := &rebuildTracker{editAt: editAt}
	if phaseAtEdit == "building" {
		r.staleBuild = true
	} else {
		r.quietSince = editAt
	}
	return r
}

// observe folds one status read in and reports whether the scan now runs
// with the edit. A deferred edit (the manager holds it until a game ends)
// never counts as applied.
func (r *rebuildTracker) observe(phase string, deferred bool, now time.Time) bool {
	if deferred {
		r.quietSince = time.Time{}
		return false
	}
	switch phase {
	case "building":
		r.quietSince = time.Time{}
		if !r.staleBuild {
			r.building = true
		}
	case "patching":
		if r.building {
			r.armedAt, r.measured = now, true
			return true
		}
		if r.staleBuild {
			// The old build ended. A scan restarted only to be stopped again
			// for this edit lasts microseconds, so the edit's build may have
			// run inside the stretch already seen as building.
			r.staleBuild, r.merged = false, true
			r.quietSince = now
			return false
		}
		if r.quietSince.IsZero() {
			r.quietSince = now
		}
		if now.Sub(r.quietSince) >= rebuildGrace {
			if r.merged {
				r.armedAt, r.measured = r.quietSince, true
			} else {
				r.armedAt = now
			}
			return true
		}
	}
	return false
}

// duration is the edit-to-scan time a completed, measured tracker saw.
func (r *rebuildTracker) duration() time.Duration {
	if !r.measured || r.armedAt.Before(r.editAt) {
		return 0
	}
	return r.armedAt.Sub(r.editAt)
}

// --- rebuild budget ---

// The budget is how long this PC needs from a library edit until the scan
// runs again. Map WADs dominate it: an override routed into Map11/12/30
// copies a multi-gigabyte file, measured at 4-7s on an NVMe drive and
// 45-98s on a slow disk. A PC that has never been seen copying a map WAD
// keeps the unproven floor, so quick champion-only rebuilds cannot talk the
// budget down before a heavy one was measured.
//
// With the map copies prepared (mapprep.go), a change that touches a map
// only rewrites the end of its copy, so the copying cannot happen and the
// budget comes from the rebuilds measured with the maps ready.
const (
	minRebuildBudget      = 15 * time.Second
	unprovenRebuildBudget = 60 * time.Second
	// readyRebuildBudget: maps ready, but no rebuild measured yet that
	// did not copy one.
	readyRebuildBudget = 20 * time.Second
	rebuildSampleLimit = 8
)

type rebuildSample struct {
	At    time.Time `json:"at"`
	Ms    int64     `json:"ms"`
	Heavy bool      `json:"heavy"` // a map WAD was copied in full
	// MapsReady: the map copies were prepared before the edit, so a map
	// WAD written by it was rewritten in place.
	MapsReady bool `json:"mapsReady,omitempty"`
}

type rebuildHistory struct {
	Recent []rebuildSample `json:"recent"`
	Heavy  []rebuildSample `json:"heavy"`
}

var (
	rebuildMu     sync.Mutex
	rebuildLoaded bool
	rebuilds      rebuildHistory
)

func rebuildHistoryFile() string {
	return filepath.Join(config.AmeDir, "ltk-rebuild-times.json")
}

func loadRebuildsLocked() {
	if rebuildLoaded {
		return
	}
	rebuildLoaded = true
	if data, err := os.ReadFile(rebuildHistoryFile()); err == nil {
		_ = json.Unmarshal(data, &rebuilds)
	}
}

func appendSample(samples []rebuildSample, sample rebuildSample) []rebuildSample {
	samples = append(samples, sample)
	if len(samples) > rebuildSampleLimit {
		samples = samples[len(samples)-rebuildSampleLimit:]
	}
	return samples
}

// recordRebuild remembers one measured edit-to-scan duration.
func recordRebuild(d time.Duration, heavy, mapsReady bool, at time.Time) {
	if d <= 0 {
		return
	}
	rebuildMu.Lock()
	defer rebuildMu.Unlock()
	loadRebuildsLocked()
	sample := rebuildSample{At: at.UTC(), Ms: d.Milliseconds(), Heavy: heavy, MapsReady: mapsReady}
	rebuilds.Recent = appendSample(rebuilds.Recent, sample)
	if heavy {
		rebuilds.Heavy = appendSample(rebuilds.Heavy, sample)
	}
	if data, err := json.MarshalIndent(rebuilds, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(rebuildHistoryFile()), 0755)
		_ = os.WriteFile(rebuildHistoryFile(), data, 0644)
	}
}

func budgetFor(worst time.Duration) time.Duration {
	budget := worst*3/2 + 5*time.Second
	if budget < minRebuildBudget {
		budget = minRebuildBudget
	}
	return budget
}

// rebuildBudgetFrom is the budget the history allows. mapsReady: the map
// copies are prepared now, so the next change rewrites no map in full.
func rebuildBudgetFrom(history rebuildHistory, mapsReady bool) time.Duration {
	if mapsReady {
		var ready, light time.Duration
		var haveReady, haveLight bool
		for _, sample := range history.Recent {
			d := time.Duration(sample.Ms) * time.Millisecond
			switch {
			case sample.MapsReady:
				haveReady = true
				if d > ready {
					ready = d
				}
			case !sample.Heavy:
				haveLight = true
				if d > light {
					light = d
				}
			}
		}
		switch {
		case haveReady:
			return budgetFor(ready)
		case haveLight:
			return budgetFor(light)
		}
		return readyRebuildBudget
	}
	var worst time.Duration
	for _, samples := range [][]rebuildSample{history.Recent, history.Heavy} {
		for _, sample := range samples {
			if d := time.Duration(sample.Ms) * time.Millisecond; d > worst {
				worst = d
			}
		}
	}
	budget := budgetFor(worst)
	if len(history.Heavy) == 0 && budget < unprovenRebuildBudget {
		budget = unprovenRebuildBudget
	}
	return budget
}

// mapsReadyNow reports whether the map copies are prepared, so a change now
// rewrites maps in place. A variable for tests.
var mapsReadyNow = func() bool {
	if !mapPrepEnabled() {
		return false
	}
	ready, _ := MapsPrepared(findGameDir())
	return ready
}

// RebuildBudget estimates how long a skin set change takes on this PC, from
// the edit until the manager's scan runs again with it.
func RebuildBudget() time.Duration {
	ready := mapsReadyNow()
	rebuildMu.Lock()
	defer rebuildMu.Unlock()
	loadRebuildsLocked()
	return rebuildBudgetFrom(rebuilds, ready)
}

// rewroteMapWAD reports whether the build that ended at armedAt wrote one
// of the map WADs in the session's overlay: only those multi-gigabyte files
// make a rebuild heavy, and only when the copy was not ready (a ready copy
// is rewritten in place, which changes its time too).
func rewroteMapWAD(prefix string, editAt, armedAt time.Time) bool {
	if prefix == "" {
		return false
	}
	dir := filepath.Join(prefix, "DATA", "FINAL", "Maps", "Shipping")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".wad.client") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		modified := info.ModTime()
		if !modified.Before(editAt.Add(-time.Second)) && !modified.After(armedAt.Add(time.Second)) {
			return true
		}
	}
	return false
}

// --- DLL verdicts ---

// overlayVerdict normalizes what the DLL said about an attached game.
// Manager v1.33.0 serializes OverlayOutcome in kebab-case ("live",
// "too-late", "end-of-life", "disabled", "hook-failed", "none"); the
// PascalCase spelling earlier builds were matched against is accepted too.
func overlayVerdict(outcome string) string {
	key := strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(outcome))
	switch key {
	case "live":
		return "live"
	case "toolate":
		return "too-late"
	case "endoflife":
		return "end-of-life"
	case "disabled":
		return "disabled"
	case "hookfailed":
		return "hook-failed"
	case "", "none":
		return ""
	}
	return strings.ToLower(strings.TrimSpace(outcome))
}
