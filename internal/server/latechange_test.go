package server

import (
	"fmt"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/lcu"
	"github.com/hoangvu12/ame/internal/ltk"
)

// TestMain keeps every server test away from the real manager, client, game
// and ame directory: the suite may run while ame, the manager and a match
// are live on this machine. A test that reaches the manager fails loudly.
func TestMain(m *testing.M) {
	gameRunning = func() bool { return false }
	preparedState = func() ltk.Prepared { return ltk.Prepared{} }
	rebuildBudget = func() time.Duration { return 15 * time.Second }
	patcherWaiting = func() bool { return false }
	probeMatchTiming = func() matchTiming { return matchTiming{stage: stageUnknown, detail: "no client in tests"} }
	startRuntime = func(string) error { panic("test reached the manager hand-off") }
	stopRuntime = func() { panic("test reached the runtime stop") }
	dir, err := os.MkdirTemp("", "ame-server-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	config.AmeDir = dir
	config.ToolsDir = filepath.Join(dir, "tools")
	config.SkinsDir = filepath.Join(dir, "skins")
	config.ModsDir = filepath.Join(dir, "mods")
	config.OverlayDir = filepath.Join(dir, "overlay")
	config.PenguDir = filepath.Join(dir, "pengu")
	config.CustomModsDir = filepath.Join(dir, "custom-mods")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

type lateStubs struct {
	probes int
	stops  int
}

func stubLateChange(t *testing.T, running bool, state ltk.Prepared, timing matchTiming, budget time.Duration) *lateStubs {
	t.Helper()
	stubs := &lateStubs{}
	oldRunning, oldState, oldBudget, oldProbe, oldStop := gameRunning, preparedState, rebuildBudget, probeMatchTiming, stopRuntime
	gameRunning = func() bool { return running }
	preparedState = func() ltk.Prepared { return state }
	rebuildBudget = func() time.Duration { return budget }
	probeMatchTiming = func() matchTiming { stubs.probes++; return timing }
	stopRuntime = func() { stubs.stops++ }
	setPreparedChampion("")
	t.Cleanup(func() {
		gameRunning, preparedState, rebuildBudget, probeMatchTiming, stopRuntime = oldRunning, oldState, oldBudget, oldProbe, oldStop
		setPreparedChampion("")
	})
	return stubs
}

// loopbackClient registers a server-side connection whose peer reads what
// the server sends, without the application connection handler.
func loopbackClient(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		accepted <- conn
		<-release
	}))
	peer, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := <-accepted
	clientsMu.Lock()
	clients[conn] = &sync.Mutex{}
	clientsMu.Unlock()
	t.Cleanup(func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		close(release)
		peer.Close()
		conn.Close()
		srv.Close()
	})
	return conn, peer
}

func readStatus(t *testing.T, peer *websocket.Conn) StatusMessage {
	t.Helper()
	peer.SetReadDeadline(time.Now().Add(3 * time.Second))
	var msg StatusMessage
	if err := peer.ReadJSON(&msg); err != nil {
		t.Fatalf("no status from the server: %v", err)
	}
	return msg
}

func finalization(left time.Duration) matchTiming {
	return matchTiming{stage: stageChampSelect, remaining: left, timed: true, detail: "FINALIZATION"}
}

func TestDecideChange(t *testing.T) {
	friendBudget := 98724*time.Millisecond*3/2 + 5*time.Second // 153s
	cases := []struct {
		name string
		c    changeContext
		want changeVerdict
	}{
		{"nothing armed, no game: always worth trying", changeContext{timing: matchTiming{stage: stageUnknown}}, changeAllowed},
		{"nothing armed, League running: hard gate", changeContext{gameRunning: true}, changeRefused},
		{"armed, League running: keep it", changeContext{gameRunning: true, armed: true}, changeKept},
		{"rebuild in flight when League started: honest refusal", changeContext{gameRunning: true, pending: true, pendingFor: 17 * time.Second}, changeRefused},
		{"armed before champion select", changeContext{armed: true, budget: friendBudget, timing: matchTiming{stage: stageBefore}}, changeAllowed},
		// The friend's 16:35:07 apply: champion select had just ended.
		{"armed, champion select ended", changeContext{armed: true, budget: 15 * time.Second, timing: matchTiming{stage: stageLaunching}}, changeKept},
		{"armed, client silent: assume no time", changeContext{armed: true, budget: 15 * time.Second, timing: matchTiming{stage: stageUnknown}}, changeKept},
		{"armed, untimed champion select", changeContext{armed: true, budget: 15 * time.Second, timing: matchTiming{stage: stageChampSelect}}, changeKept},
		{"slow disk, 18s left", changeContext{armed: true, budget: friendBudget, timing: finalization(18 * time.Second)}, changeKept},
		{"fast disk, 18s left", changeContext{armed: true, budget: 15 * time.Second, timing: finalization(18 * time.Second)}, changeAllowed},
		{"fast disk, 12s left", changeContext{armed: true, budget: 15 * time.Second, timing: finalization(12 * time.Second)}, changeKept},
		// A change queued behind a running rebuild pays for both.
		{"behind a rebuild, too little time", changeContext{pending: true, pendingFor: 5 * time.Second, budget: 60 * time.Second, timing: finalization(100 * time.Second)}, changeKept},
		{"behind a rebuild, enough time", changeContext{pending: true, pendingFor: 5 * time.Second, budget: 60 * time.Second, timing: finalization(120 * time.Second)}, changeAllowed},
	}
	for _, c := range cases {
		got, reason := decideChange(c.c)
		if got != c.want {
			t.Fatalf("%s: verdict=%d (%s), want %d", c.name, got, reason, c.want)
		}
		if got != changeAllowed && reason == "" {
			t.Fatalf("%s: blocked without a reason", c.name)
		}
	}
}

func TestChampSelectTiming(t *testing.T) {
	now := time.Date(2026, 11, 6, 16, 35, 0, 0, time.UTC)
	timer := func(phase string, leftMs int64, updated time.Duration) lcu.ChampSelectTimer {
		return lcu.ChampSelectTimer{Phase: phase, AdjustedTimeLeftInPhase: leftMs, InternalNowInEpochMs: now.Add(-updated).UnixMilli()}
	}
	cases := []struct {
		name  string
		timer lcu.ChampSelectTimer
		want  time.Duration
		timed bool
	}{
		{"finalization, updated 5s ago", timer("FINALIZATION", 20000, 5*time.Second), 15 * time.Second, true},
		{"pick turn ending: finalization still follows", timer("BAN_PICK", 3000, 0), 13 * time.Second, true},
		{"game starting", timer("GAME_STARTING", 4000, 0), 0, true},
		{"stale timer", timer("FINALIZATION", 3000, 10*time.Second), 0, true},
		{"clock ahead of the client", timer("FINALIZATION", 20000, -3*time.Second), 20 * time.Second, true},
		{"infinite timer", lcu.ChampSelectTimer{Phase: "FINALIZATION", IsInfinite: true}, 0, false},
	}
	for _, c := range cases {
		got := champSelectTiming(c.timer, now)
		if got.stage != stageChampSelect || got.timed != c.timed || got.remaining != c.want {
			t.Fatalf("%s: %+v, want remaining=%s timed=%v", c.name, got, c.want, c.timed)
		}
	}
}

func TestLateChangeAsksTheClientOnlyWhenAnArmedSetIsAtStake(t *testing.T) {
	stubs := stubLateChange(t, false, ltk.Prepared{}, matchTiming{stage: stageLaunching}, time.Minute)
	if verdict, _ := lateChange("104"); verdict != changeAllowed || stubs.probes != 0 {
		t.Fatalf("nothing armed: verdict=%d probes=%d", verdict, stubs.probes)
	}
	preparedState = func() ltk.Prepared { return ltk.Prepared{Armed: true} }
	setPreparedChampion("104")
	if verdict, _ := lateChange("104"); verdict != changeKept || stubs.probes != 1 {
		t.Fatalf("armed for this champion: verdict=%d probes=%d", verdict, stubs.probes)
	}
	// A set built for the champion the player swapped away from protects
	// nothing of theirs: trying the new one cannot cost the own skin.
	if verdict, _ := lateChange("238"); verdict != changeAllowed {
		t.Fatalf("set for another champion protected: %d", verdict)
	}
}

func TestApplyKeepsTheArmedSetWhenTooLate(t *testing.T) {
	stubLateChange(t, false, ltk.Prepared{Armed: true}, matchTiming{stage: stageLaunching, detail: "GameStart"}, 15*time.Second)
	setPreparedChampion("104")
	conn, peer := loopbackClient(t)
	handleApply(conn, "104", "104042", "", "Graves", "Skin", "")
	msg := readStatus(t, peer)
	if msg.Status != "kept" || !strings.Contains(msg.Message, "next match") {
		t.Fatalf("status=%+v", msg)
	}
}

func TestApplyWhileLeagueRunsNeverReachesTheManager(t *testing.T) {
	stubLateChange(t, true, ltk.Prepared{}, matchTiming{}, 15*time.Second)
	conn, peer := loopbackClient(t)
	handleApply(conn, "104", "104042", "", "Graves", "Skin", "")
	msg := readStatus(t, peer)
	if msg.Status != "error" || !strings.Contains(msg.Message, "League is already running") {
		t.Fatalf("status=%+v", msg)
	}
}

func TestPrefetchKeepsTheArmedSetWhenTooLate(t *testing.T) {
	stubLateChange(t, false, ltk.Prepared{Armed: true}, finalization(18*time.Second), 153*time.Second)
	setPreparedChampion("104")
	conn, peer := loopbackClient(t)
	handlePrefetch(conn, "104", "104042", "", "Graves", "Skin", "", selectionRevision.Load())
	msg := readStatus(t, peer)
	if msg.Status != "kept" || !strings.Contains(msg.Message, "needs about 2m33s") {
		t.Fatalf("status=%+v", msg)
	}
}

// The plugin asks for cleanup after post-game phases, including when a client
// that closed during the match comes back or a phase briefly reads "None";
// while League runs that must not stop the patcher serving the match.
func TestCleanupRequestNeverStopsTheRuntimeDuringAMatch(t *testing.T) {
	stubs := stubLateChange(t, true, ltk.Prepared{Armed: true}, matchTiming{}, 15*time.Second)
	conn, peer := loopbackClient(t)
	handleCleanupRequest(conn)
	if stubs.stops != 0 {
		t.Fatal("cleanup stopped the runtime while League ran")
	}
	if msg := readStatus(t, peer); msg.Status != "cleanupDeclined" {
		t.Fatalf("status=%+v", msg)
	}
	gameRunning = func() bool { return false }
	handleCleanupRequest(conn)
	if stubs.stops != 1 {
		t.Fatalf("cleanup after the match did not stop the runtime: %d", stubs.stops)
	}
	// Uninstall, the stuck-game button and shutdown stay unconditional.
	gameRunning = func() bool { return true }
	HandleCleanup()
	if stubs.stops != 2 {
		t.Fatal("explicit cleanup was blocked")
	}
}

func TestPrefetchDisarmNeverStopsTheRuntimeDuringAMatch(t *testing.T) {
	stubs := stubLateChange(t, true, ltk.Prepared{}, matchTiming{}, 15*time.Second)
	previous := patcherWaiting
	patcherWaiting = func() bool { return true }
	t.Cleanup(func() { patcherWaiting = previous })
	handlePrefetch(nil, "104", "0", "", "", "", "", selectionRevision.Load())
	if stubs.stops != 0 {
		t.Fatal("an empty selection stopped the runtime while League ran")
	}
	gameRunning = func() bool { return false }
	handlePrefetch(nil, "104", "0", "", "", "", "", selectionRevision.Load())
	if stubs.stops != 1 {
		t.Fatal("an empty selection before the match did not disarm the previous one")
	}
}

func TestTooLateVerdictIsReported(t *testing.T) {
	_, peer := loopbackClient(t)
	reportOverlayVerdict(3464, "too-late")
	msg := readStatus(t, peer)
	if msg.Status != "error" || !strings.Contains(msg.Message, "next match") {
		t.Fatalf("status=%+v", msg)
	}
}

// Pausing League cannot move its creation time, which is what the patcher
// DLL compares against its scan; under the manager it only hid the game
// window the manager waits for. The server must never pause the game again.
func TestServerNeverSuspendsTheGame(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			if strings.Contains(spec.Path.Value, "internal/suspend") {
				t.Fatalf("%s imports the game suspender", file)
			}
		}
	}
}
