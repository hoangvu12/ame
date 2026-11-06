package ltk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// isolateManagerState gives a test empty manager state and empties it again
// afterwards: the snapshot, the rebuild tracker and the armed times are
// package globals.
func isolateManagerState(t *testing.T) {
	t.Helper()
	reset := func() {
		snapshotMu.Lock()
		snapshot, rebuild = managerSnapshot{}, nil
		armedAt, gameEndedAt = time.Time{}, time.Time{}
		snapshotMu.Unlock()
	}
	previous := manager
	reset()
	t.Cleanup(func() { reset(); manager = previous })
}

// isolateRebuildHistory starts a test without measured rebuilds.
func isolateRebuildHistory(t *testing.T) {
	t.Helper()
	reset := func() {
		rebuildMu.Lock()
		rebuilds, rebuildLoaded = rebuildHistory{}, true
		rebuildMu.Unlock()
		os.Remove(rebuildHistoryFile())
	}
	reset()
	t.Cleanup(reset)
}

// stubGame answers gameProcessRunning from running.
func stubGame(t *testing.T, running func() bool) {
	t.Helper()
	previous := gameProcessRunning
	gameProcessRunning = running
	t.Cleanup(func() { gameProcessRunning = previous })
}

func dialFake(t *testing.T, fake *fakeCDP) *cdpClient {
	t.Helper()
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.close)
	return c
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// clock reads a log timestamp of 2026-11-06 (UTC).
func clock(t *testing.T, hms string) time.Time {
	t.Helper()
	parsed, err := time.Parse("15:04:05.000", hms)
	if err != nil {
		t.Fatal(err)
	}
	return time.Date(2026, 11, 6, parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond(), time.UTC)
}

func TestRebuildTrackerObservedBuild(t *testing.T) {
	edit := clock(t, "16:35:08.094")
	r := newRebuildTracker(edit, "patching")
	if r.observe("patching", false, edit.Add(500*time.Millisecond)) {
		t.Fatal("done before the edit's build appeared")
	}
	if r.observe("building", false, edit.Add(time.Second)) {
		t.Fatal("done while building")
	}
	// The friend's 16:35 rebuild: the scan restarted 98.7s after the edit.
	restart := clock(t, "16:36:46.818")
	if !r.observe("patching", false, restart) {
		t.Fatal("scan restart after the build not recognized")
	}
	if got := r.duration(); got != restart.Sub(edit) {
		t.Fatalf("duration=%s", got)
	}
}

func TestRebuildTrackerGraceWithoutVisibleBuild(t *testing.T) {
	edit := clock(t, "16:35:01.086")
	r := newRebuildTracker(edit, "patching")
	if r.observe("patching", false, edit.Add(rebuildGrace-time.Millisecond)) {
		t.Fatal("done inside the grace without a build")
	}
	if !r.observe("patching", false, edit.Add(rebuildGrace)) {
		t.Fatal("a rebuild hidden between polls never completed")
	}
	if r.duration() != 0 {
		t.Fatal("an unobserved rebuild was measured")
	}
}

// An edit made while a build runs reaches the scan only after that build and
// its own; the scan restarting in between must not look final.
func TestRebuildTrackerWaitsOutStaleBuild(t *testing.T) {
	edit := clock(t, "16:31:07.328")
	r := newRebuildTracker(edit, "building")
	if r.observe("building", false, edit.Add(time.Second)) || r.observe("patching", false, edit.Add(2*time.Second)) {
		t.Fatal("the pre-edit build counted as the edit's rebuild")
	}
	if r.observe("building", false, edit.Add(2100*time.Millisecond)) {
		t.Fatal("done while the edit's build runs")
	}
	if !r.observe("patching", false, edit.Add(9*time.Second)) {
		t.Fatal("the edit's own build was not recognized")
	}
	if r.duration() != 9*time.Second {
		t.Fatalf("duration=%s", r.duration())
	}

	// The friend's 16:31:23.608: the restarted scan was stopped again within
	// microseconds, so polls see one long build.
	merged := newRebuildTracker(edit, "building")
	merged.observe("building", false, clock(t, "16:31:20.000"))
	end := clock(t, "16:31:37.324")
	if merged.observe("patching", false, end) || merged.observe("patching", false, end.Add(rebuildGrace-time.Millisecond)) {
		t.Fatal("done before the grace after a merged build")
	}
	if !merged.observe("patching", false, end.Add(rebuildGrace)) || merged.duration() != end.Sub(edit) {
		t.Fatalf("merged build not measured to its end: %s", merged.duration())
	}
}

func TestRebuildTrackerNeverCompletesDeferredEdit(t *testing.T) {
	edit := clock(t, "16:00:00.000")
	r := newRebuildTracker(edit, "patching")
	for i := 0; i < 10; i++ {
		if r.observe("patching", true, edit.Add(time.Duration(i)*time.Second)) {
			t.Fatal("an edit deferred until the game ends counted as applied")
		}
	}
}

func TestRebuildBudget(t *testing.T) {
	sample := func(ms int64, heavy bool) rebuildSample { return rebuildSample{Ms: ms, Heavy: heavy} }
	cases := []struct {
		name    string
		history rebuildHistory
		want    time.Duration
	}{
		{"unmeasured PC", rebuildHistory{}, unprovenRebuildBudget},
		// Champion-only rebuilds on the friend's disk took 1.3-3.6s; they
		// must not lower the budget before a map rebuild was seen.
		{"only champion rebuilds", rebuildHistory{Recent: []rebuildSample{sample(2486, false), sample(3423, false)}}, unprovenRebuildBudget},
		// The friend's 16:35 map rebuild: 98.7s from edit to scan.
		{"slow disk", rebuildHistory{Recent: []rebuildSample{sample(2212, false)}, Heavy: []rebuildSample{sample(98724, true)}}, 98724*time.Millisecond*3/2 + 5*time.Second},
		// The owner's NVMe at 16:35: 4.8s with Map12/Map30 rewritten.
		{"fast disk", rebuildHistory{Heavy: []rebuildSample{sample(4777, true)}}, minRebuildBudget},
		// The owner at 18:34: 33s, Map11 copied while a health sweep ran.
		{"contended fast disk", rebuildHistory{Heavy: []rebuildSample{sample(4777, true), sample(33056, true)}}, 33056*time.Millisecond*3/2 + 5*time.Second},
	}
	for _, c := range cases {
		if got := rebuildBudgetFrom(c.history, false); got != c.want {
			t.Fatalf("%s: budget=%s, want %s", c.name, got, c.want)
		}
	}
}

// With the map copies prepared, no change copies a map: the budget comes
// from the rebuilds measured with the maps ready, not from old full copies.
func TestRebuildBudgetWithMapsReady(t *testing.T) {
	sample := func(ms int64, heavy, ready bool) rebuildSample {
		return rebuildSample{Ms: ms, Heavy: heavy, MapsReady: ready}
	}
	friend := rebuildHistory{
		Recent: []rebuildSample{sample(2212, false, false), sample(98724, true, false)},
		Heavy:  []rebuildSample{sample(98724, true, false)},
	}
	cases := []struct {
		name    string
		history rebuildHistory
		want    time.Duration
	}{
		{"unmeasured PC", rebuildHistory{}, readyRebuildBudget},
		{"only old full copies", rebuildHistory{Recent: []rebuildSample{sample(98724, true, false)}, Heavy: []rebuildSample{sample(98724, true, false)}}, readyRebuildBudget},
		// No ready sample yet: the champion-only rebuilds stand in.
		{"champion rebuilds", friend, minRebuildBudget},
		{"slow ready rebuild", rebuildHistory{Recent: []rebuildSample{sample(2212, false, false), sample(12000, false, true)}}, 12*time.Second*3/2 + 5*time.Second},
		{"fast ready rebuilds", rebuildHistory{Recent: append(append([]rebuildSample{}, friend.Recent...), sample(1400, false, true)), Heavy: friend.Heavy}, minRebuildBudget},
	}
	for _, c := range cases {
		if got := rebuildBudgetFrom(c.history, true); got != c.want {
			t.Fatalf("%s: budget=%s, want %s", c.name, got, c.want)
		}
	}
	// Maps not ready: the formula is unchanged and keeps the full copies.
	if got, want := rebuildBudgetFrom(friend, false), 98724*time.Millisecond*3/2+5*time.Second; got != want {
		t.Fatalf("maps not ready: budget=%s, want %s", got, want)
	}
}

// A ready map copy is rewritten in place, which changes its time like a
// full copy does: only a write to a map that was not ready is heavy.
func TestObservedRebuildClassifiesPreparedMaps(t *testing.T) {
	for _, ready := range []bool{false, true} {
		isolateManagerState(t)
		isolateRebuildHistory(t)
		previous := mapsReadyNow
		mapsReadyNow = func() bool { return ready }
		prefix := t.TempDir()
		edit := time.Now().Add(-10 * time.Second)
		path := filepath.Join(prefix, "DATA", "FINAL", "Maps", "Shipping", "Map11.wad.client")
		writeSized(t, path, 1, edit.Add(time.Second))
		var status patcherStatus
		if err := json.Unmarshal([]byte(`{"running":true,"phase":"building","session":{"overlayPrefix":`+jsLiteral(prefix)+`}}`), &status); err != nil {
			t.Fatal(err)
		}
		beginRebuild(edit, "patching")
		mapsReadyNow = previous
		observeRebuild(status, edit.Add(time.Second))
		status.Phase = "patching"
		if !observeRebuild(status, edit.Add(2*time.Second)) {
			t.Fatal("the rebuild did not complete")
		}
		rebuildMu.Lock()
		got := rebuilds.Recent[len(rebuilds.Recent)-1]
		rebuildMu.Unlock()
		if got.MapsReady != ready || got.Heavy == ready {
			t.Fatalf("maps ready=%v: sample=%+v", ready, got)
		}
	}
}

func TestRecordRebuildPersistsAndBounds(t *testing.T) {
	isolateRebuildHistory(t)
	at := clock(t, "18:35:10.208")
	for i := 0; i < rebuildSampleLimit+3; i++ {
		recordRebuild(time.Duration(i+1)*time.Second, i == 0, false, at)
	}
	// A new run reads the history back from disk.
	rebuildMu.Lock()
	rebuildLoaded, rebuilds = false, rebuildHistory{}
	rebuildMu.Unlock()
	want := time.Duration(rebuildSampleLimit+3)*time.Second*3/2 + 5*time.Second
	if got := RebuildBudget(); got != want {
		t.Fatalf("budget after reload=%s, want %s", got, want)
	}
	rebuildMu.Lock()
	defer rebuildMu.Unlock()
	if len(rebuilds.Recent) != rebuildSampleLimit || len(rebuilds.Heavy) != 1 {
		t.Fatalf("history not bounded: recent=%d heavy=%d", len(rebuilds.Recent), len(rebuilds.Heavy))
	}
}

func TestRewroteMapWAD(t *testing.T) {
	prefix := t.TempDir()
	maps := filepath.Join(prefix, "DATA", "FINAL", "Maps", "Shipping")
	champions := filepath.Join(prefix, "DATA", "FINAL", "Champions")
	for _, dir := range []string{maps, champions} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	edit := time.Now().Add(-time.Minute)
	armed := edit.Add(30 * time.Second)
	write := func(path string, at time.Time) {
		t.Helper()
		if err := os.WriteFile(path, []byte("wad"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, at, at); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(champions, "Zed.wad.client"), edit.Add(5*time.Second))
	write(filepath.Join(maps, "Map12.wad.client"), edit.Add(-time.Hour)) // reused from an earlier build
	if rewroteMapWAD(prefix, edit, armed) {
		t.Fatal("champion WADs or a reused map WAD counted as a map rebuild")
	}
	write(filepath.Join(maps, "Map11.wad.client"), edit.Add(29*time.Second))
	if !rewroteMapWAD(prefix+string(os.PathSeparator), edit, armed) {
		t.Fatal("a map WAD rewritten by the build was missed")
	}
	if rewroteMapWAD("", edit, armed) {
		t.Fatal("an unknown overlay root counted as a map rebuild")
	}
}

func TestPreparedStateBelongsToTheNextMatch(t *testing.T) {
	isolateManagerState(t)
	setSnapshot(func(s *managerSnapshot) {
		*s = managerSnapshot{alive: true, running: true, phase: "patching", observed: time.Now()}
	})
	if PreparedState().Armed {
		t.Fatal("armed before ame's package ever reached the scan")
	}
	snapshotMu.Lock()
	armedAt = time.Now()
	snapshotMu.Unlock()
	if !PreparedState().Armed {
		t.Fatal("a scan running with ame's package is not armed")
	}
	snapshotMu.Lock()
	gameEndedAt = armedAt.Add(time.Second)
	snapshotMu.Unlock()
	if PreparedState().Armed {
		t.Fatal("the set left from a finished match counted as armed for the next one")
	}
	beginRebuild(time.Now().Add(-2*time.Second), "patching")
	if st := PreparedState(); !st.Pending || st.Armed || st.PendingFor < 2*time.Second {
		t.Fatalf("pending edit reported as %+v", st)
	}
	setSnapshot(func(s *managerSnapshot) { s.running = false })
	if st := PreparedState(); st.Pending || st.Armed {
		t.Fatal("a stopped patcher reported a prepared set")
	}
}

// No library edit — and no manager relaunch — may happen while League's
// process exists: the edit restarts the scan, and the DLL refuses any game
// created before the new scan. The check runs again right before the write.
func TestStartRefusesLibraryEditsWhileGameRuns(t *testing.T) {
	for _, when := range []string{"at entry", "right before the write"} {
		t.Run(when, func(t *testing.T) {
			isolateManagerState(t)
			pkg := filepath.Join(t.TempDir(), "overlay.fantome")
			if err := os.WriteFile(pkg, []byte("late set"), 0600); err != nil {
				t.Fatal(err)
			}
			var writes atomic.Int32
			fake := startFakeCDP(t, func(expr string) (any, error) {
				switch {
				case strings.Contains(expr, "get_patcher_status"):
					return fakeInvokeReply(map[string]any{"running": true, "phase": "patching"}), nil
				case strings.Contains(expr, "get_installed_mods"):
					return fakeInvokeReply([]any{map[string]any{"id": "armed", "name": ameModName, "enabled": true}}), nil
				case strings.Contains(expr, "plugin:library|"), strings.Contains(expr, "start_patcher"), strings.Contains(expr, "stop_patcher"):
					writes.Add(1)
					return fakeInvokeReply(nil), nil
				}
				return nil, nil
			})
			manager = &managerSession{cdp: dialFake(t, fake), modID: "armed", modFile: pkg, modHash: [32]byte{1}, stop: make(chan struct{}), exited: make(chan struct{})}
			calls := 0
			stubGame(t, func() bool {
				calls++
				return when == "at entry" || calls > 1
			})
			err := Start(strings.TrimSuffix(pkg, ".fantome"))
			if !errors.Is(err, ErrGameRunning) {
				t.Fatalf("Start while League runs: %v", err)
			}
			if n := writes.Load(); n != 0 {
				t.Fatalf("%d library or patcher writes while League ran", n)
			}
			if st := PreparedState(); st.Pending {
				t.Fatal("a refused change left a rebuild fence")
			}
		})
	}
}

// A new skin set is installed and enabled before the previous ame packages
// are uninstalled, in one burst: a refresh in between never sees an empty
// set. update_mod is never used: the manager answers it by deleting every
// cached overlay WAD, the prepared map copies included.
func TestReplaceInstallsBeforeUninstalling(t *testing.T) {
	for _, already := range []bool{false, true} {
		t.Run(fmt.Sprintf("alreadyInstalled=%v", already), func(t *testing.T) {
			isolateManagerState(t)
			pkg := filepath.Join(t.TempDir(), "overlay.fantome")
			if err := os.WriteFile(pkg, []byte("next set"), 0600); err != nil {
				t.Fatal(err)
			}
			newID := "fresh"
			if already {
				// The manager already holds these bytes as the armed mod.
				newID = "armed"
			}
			var mu sync.Mutex
			var calls []string
			record := func(call string) {
				mu.Lock()
				calls = append(calls, call)
				mu.Unlock()
			}
			fake := startFakeCDP(t, func(expr string) (any, error) {
				switch {
				case strings.Contains(expr, "get_installed_mods"):
					return fakeInvokeReply([]any{
						map[string]any{"id": "leftover", "name": "ame-skin-set", "enabled": true},
						map[string]any{"id": "armed", "name": ameModName, "enabled": true},
						map[string]any{"id": "users-mod", "name": "Someone else's mod", "enabled": true},
					}), nil
				case strings.Contains(expr, "update_mod"):
					return nil, fmt.Errorf("update_mod purges the manager's overlay cache: %s", expr)
				case strings.Contains(expr, "uninstall_mod"):
					for _, id := range []string{"leftover", "armed"} {
						if strings.Contains(expr, `"modId":"`+id+`"`) {
							record("uninstall:" + id)
							return fakeInvokeReply(nil), nil
						}
					}
					return nil, fmt.Errorf("uninstalled the wrong mod: %s", expr)
				case strings.Contains(expr, "toggle_mod"):
					if !strings.Contains(expr, `"modId":"`+newID+`"`) || !strings.Contains(expr, `"enabled":true`) {
						return nil, fmt.Errorf("enabled the wrong mod: %s", expr)
					}
					record("enable:" + newID)
					return fakeInvokeReply(nil), nil
				case strings.Contains(expr, "install_mod"):
					if !strings.Contains(expr, "overlay.fantome") {
						return nil, fmt.Errorf("installed the wrong file: %s", expr)
					}
					record("install")
					return fakeInvokeReply(map[string]any{"kind": "installed", "mod": map[string]any{"id": newID, "name": ameModName}}), nil
				}
				return nil, fmt.Errorf("unexpected IPC: %s", expr)
			})
			m := &managerSession{cdp: dialFake(t, fake), modID: "armed"}
			if err := swapAmeMod(m, pkg); err != nil {
				t.Fatal(err)
			}
			want := []string{"install", "enable:fresh", "uninstall:leftover", "uninstall:armed"}
			if already {
				want = []string{"install", "enable:armed", "uninstall:leftover"}
			}
			if strings.Join(calls, ",") != strings.Join(want, ",") {
				t.Fatalf("library edits=%v, want %v", calls, want)
			}
			digest, _ := packageDigest(pkg)
			if !m.matchesPackage(pkg, digest) || m.modID != newID {
				t.Fatalf("the installed package's identity was not recorded: %q", m.modID)
			}
			if !PreparedState().Pending && !currentSnapshot().pending {
				t.Fatal("the edit raised no rebuild fence")
			}
		})
	}
}

// With the patcher idle, the package goes into the library first and the
// patcher starts after it, so its first build is the selected set; it used
// to start on the previous match's library and race the replacement.
func TestIdlePatcherStartsAfterTheLibraryHoldsThePackage(t *testing.T) {
	isolateManagerState(t)
	isolateRebuildHistory(t)
	pkg := filepath.Join(t.TempDir(), "overlay.fantome")
	if err := os.WriteFile(pkg, []byte("first set"), 0600); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var calls []string
	started := false
	polls := 0
	fake := startFakeCDP(t, func(expr string) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case strings.Contains(expr, "get_patcher_status"):
			if !started {
				return fakeInvokeReply(map[string]any{"running": false, "phase": "idle"}), nil
			}
			polls++
			phase := "building"
			if polls > 3 {
				phase = "patching"
			}
			return fakeInvokeReply(map[string]any{"running": true, "phase": phase}), nil
		case strings.Contains(expr, "get_installed_mods"):
			return fakeInvokeReply([]any{}), nil
		case strings.Contains(expr, "install_mod"):
			calls = append(calls, "install")
			return fakeInvokeReply(map[string]any{"kind": "installed", "mod": map[string]any{"id": "new", "name": ameModName}}), nil
		case strings.Contains(expr, "toggle_mod"):
			calls = append(calls, "enable")
			return fakeInvokeReply(nil), nil
		case strings.Contains(expr, "start_patcher"):
			calls = append(calls, "start")
			started = true
			return fakeInvokeReply(nil), nil
		case strings.Contains(expr, "__AME_BRIDGE__"):
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected IPC: %s", expr)
	})
	manager = &managerSession{cdp: dialFake(t, fake), stop: make(chan struct{}), exited: make(chan struct{})}
	if err := Start(strings.TrimSuffix(pkg, ".fantome")); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "install,enable,start" {
		t.Fatalf("sequence=%v", calls)
	}
	if st := PreparedState(); st.Pending || currentSnapshot().pending {
		t.Fatal("the fence stayed up after the scan started")
	}
}

func TestBeginPreparationNeverStartsThePatcher(t *testing.T) {
	isolateManagerState(t)
	manager = nil
	// killExistingManagers panics in tests; a manager launch would reach it.
	if _, err := BeginPreparation(filepath.Join(t.TempDir(), "overlay")); err != nil {
		t.Fatal(err)
	}
	if Running() {
		t.Fatal("preparation started a patcher")
	}
}

// The friend's 16:35 match (logs of 2026-11-06): a teammate swap at the end of
// champion select made the manager rebuild the map WADs for 98.7s; League
// started 17s into it and the DLL refused the match.
func TestManagerWatchReportsFriendLateSwap(t *testing.T) {
	var w managerWatch
	w.note(managerView{running: true, phase: "patching"}, clock(t, "16:35:04.069"), 0)
	edit := clock(t, "16:35:07.300")
	expect := func(lines []string, parts ...string) {
		t.Helper()
		joined := strings.Join(lines, "\n")
		for _, part := range parts {
			if !strings.Contains(joined, part) {
				t.Fatalf("missing %q in:\n%s", part, joined)
			}
		}
	}
	at := clock(t, "16:35:08.095")
	expect(w.note(managerView{running: true, phase: "building"}, at, at.Sub(edit)), "patching -> building", "scan stopped")
	at = clock(t, "16:35:25.084")
	expect(w.note(managerView{running: true, phase: "building", game: true}, at, at.Sub(edit)),
		"League started while ame's skin set was still rebuilding", "joined too late", "next match")
	at = clock(t, "16:36:46.818")
	expect(w.note(managerView{running: true, phase: "patching", game: true}, at, at.Sub(edit)), "building -> patching after 1m38.7s", "scan started")
	verdict := managerView{running: true, phase: "patching", game: true, attached: true, pid: 3464, verdict: "too-late", failure: overlayFailureMessage("too-late")}
	lines := w.note(verdict, clock(t, "16:36:47.077"), 0)
	expect(lines, "attached pid=3464", "verdict pid=3464: too-late")
	if strings.Contains(strings.Join(lines, "\n"), "patcher failure") {
		t.Fatalf("the verdict was logged twice: %v", lines)
	}
	expect(w.note(managerView{running: true, phase: "patching"}, clock(t, "17:02:54.529"), 0), "League game process ended")
}

// The owner's 18:34 match: the first arm came 18s before League, and the
// rebuild (Map11 copied in 31.9s) restarted the scan after League existed.
func TestManagerWatchReportsOwnerFirstArmTooLate(t *testing.T) {
	var w managerWatch
	w.note(managerView{running: true, phase: "building"}, clock(t, "18:34:37.204"), time.Second)
	at := clock(t, "18:34:53.884")
	lines := w.note(managerView{running: true, phase: "building", game: true}, at, at.Sub(clock(t, "18:34:35.794")))
	if !strings.Contains(strings.Join(lines, "\n"), "still rebuilding (edit 18.1s ago)") {
		t.Fatalf("in-flight launch not reported: %v", lines)
	}
	lines = w.note(managerView{running: true, phase: "patching", game: true, attached: true, pid: 33748, verdict: "too-late"}, clock(t, "18:35:14.190"), 0)
	if !strings.Contains(strings.Join(lines, "\n"), "verdict pid=33748: too-late") {
		t.Fatalf("verdict not reported: %v", lines)
	}
}

// The owner's successful 16:35 match: the scan restarted 6.9s before League.
func TestManagerWatchReportsMarginAndLiveVerdict(t *testing.T) {
	var w managerWatch
	w.note(managerView{running: true, phase: "building"}, clock(t, "16:35:13.960"), time.Second)
	w.note(managerView{running: true, phase: "patching"}, clock(t, "16:35:18.775"), 0)
	lines := w.note(managerView{running: true, phase: "patching", game: true}, clock(t, "16:35:25.640"), 0)
	if !strings.Contains(strings.Join(lines, "\n"), "League started 6.9s after the patcher scan began") {
		t.Fatalf("margin not reported: %v", lines)
	}
	lines = w.note(managerView{running: true, phase: "patching", game: true, attached: true, pid: 35816, verdict: "live"}, clock(t, "16:35:28.107"), 0)
	if !strings.Contains(strings.Join(lines, "\n"), "verdict pid=35816: live, the patcher DLL confirmed") {
		t.Fatalf("live verdict not reported: %v", lines)
	}
}

// pollManager turns the bridge's kebab-case verdict into a failure and a
// report, and a finished game makes the armed set belong to the past match.
func TestPollManagerReportsVerdictAndGameEnd(t *testing.T) {
	isolateManagerState(t)
	var mu sync.Mutex
	state := map[string]any{"shim": true, "attached": false, "status": map[string]any{"running": true, "phase": "patching"}}
	fake := startFakeCDP(t, func(expr string) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		if strings.Contains(expr, "statusError") {
			copied := map[string]any{}
			for k, v := range state {
				copied[k] = v
			}
			return copied, nil
		}
		return nil, fmt.Errorf("unexpected expression: %s", expr)
	})
	var game atomic.Bool
	stubGame(t, game.Load)
	verdicts := make(chan string, 4)
	previous := OnOverlayVerdict
	OnOverlayVerdict = func(pid uint64, verdict string) { verdicts <- fmt.Sprintf("%d:%s", pid, verdict) }
	t.Cleanup(func() { OnOverlayVerdict = previous })
	snapshotMu.Lock()
	armedAt = time.Now()
	snapshotMu.Unlock()
	m := &managerSession{cdp: dialFake(t, fake), stop: make(chan struct{}), exited: make(chan struct{})}
	done := make(chan struct{})
	go func() { pollManager(m); close(done) }()
	t.Cleanup(func() { close(m.stop); <-done })

	waitFor(t, "a scanning snapshot", func() bool { return currentSnapshot().phase == "patching" })
	if !PreparedState().Armed {
		t.Fatal("the armed set is not prepared before the game")
	}
	game.Store(true)
	mu.Lock()
	state["attached"], state["pid"], state["overlay"] = true, 3464, "too-late"
	mu.Unlock()
	select {
	case got := <-verdicts:
		if got != "3464:too-late" {
			t.Fatalf("verdict=%s", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("too-late verdict never reported")
	}
	if !strings.Contains(Failure(), "next match") || Hooked() {
		t.Fatalf("too-late game reported as hooked or without a failure: %q", Failure())
	}
	game.Store(false)
	mu.Lock()
	state["attached"], state["pid"], state["overlay"] = false, nil, nil
	mu.Unlock()
	waitFor(t, "the game end", func() bool {
		snapshotMu.Lock()
		defer snapshotMu.Unlock()
		return !gameEndedAt.IsZero()
	})
	if PreparedState().Armed {
		t.Fatal("the set armed for the finished match still counts as armed")
	}
}
