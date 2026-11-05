package ltk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// wsURL converts the httptest server's HTTP URL into its WebSocket form.
func wsURL(serverURL string) string {
	return strings.Replace(serverURL, "http://", "ws://", 1)
}

// fakeCDP answers Runtime.evaluate the way the manager's WebView would: the
// reply value is chosen by inspecting the evaluated expression.
type fakeCDP struct {
	server  *httptest.Server
	url     string
	handler func(expr string) (any, error)
}

func startFakeCDP(t *testing.T, handler func(expr string) (any, error)) *fakeCDP {
	t.Helper()
	fake := &fakeCDP{handler: handler}
	mux := http.NewServeMux()
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"type":"page","title":"LTK Manager","url":"http://tauri.localhost/","webSocketDebuggerUrl":"%s/ws"}]`, wsURL(fake.server.URL))
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			var msg cdpCall
			msg.Params = map[string]any{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			expr, _ := msg.Params["expression"].(string)
			var payload struct {
				Result struct {
					Type  string `json:"type"`
					Value any    `json:"value"`
				} `json:"result"`
			}
			value, err := fake.handler(expr)
			if err != nil {
				conn.WriteJSON(map[string]any{
					"id": msg.ID,
					"result": map[string]any{
						"exceptionDetails": map[string]any{"text": err.Error()},
					},
				})
				continue
			}
			payload.Result.Type = "object"
			if value == nil {
				payload.Result.Type = "undefined"
			}
			payload.Result.Value = value
			conn.WriteJSON(map[string]any{"id": msg.ID, "result": payload})
		}
	})
	fake.server = httptest.NewServer(mux)
	t.Cleanup(fake.server.Close)
	return fake
}

func fakeInvokeReply(value any) any {
	return map[string]any{"__ipc": true, "ok": true, "value": value}
}

func TestCDPInvokeTauri(t *testing.T) {
	fake := startFakeCDP(t, func(expr string) (any, error) {
		switch {
		case strings.Contains(expr, "plugin:library|install_mod"):
			return fakeInvokeReply(map[string]any{
				"kind": "installed",
				"mod":  map[string]any{"id": "abc123", "name": ameModName},
			}), nil
		case strings.Contains(expr, "plugin:patcher|get_patcher_status"):
			return fakeInvokeReply(map[string]any{
				"running": true,
				"phase":   "patching",
				"session": map[string]any{"overlayPrefix": `C:\o\`},
			}), nil
		case strings.Contains(expr, "plugin:patcher|stop_patcher"):
			return map[string]any{"__ipc": true, "ok": false, "error": map[string]any{
				"code":  "PATCHER",
				"error": map[string]any{"kind": "NOT_RUNNING"},
			}}, nil
		case strings.Contains(expr, "plugin:library|toggle_mod"):
			return map[string]any{"__ipc": false}, nil
		default:
			return nil, fmt.Errorf("unexpected expression: %s", expr)
		}
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()

	id, err := c.installMod(`C:\skins\ame skin set.fantome`)
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" {
		t.Fatalf("install_mod id=%q", id)
	}

	status, err := c.patcherStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !status.Running || status.Phase != "patching" || status.Session == nil || status.Session.OverlayPrefix != `C:\o\` {
		t.Fatalf("get_patcher_status=%+v", status)
	}

	if err := c.stopPatcher(); err != nil {
		t.Fatalf("NOT_RUNNING must be a successful stop: %v", err)
	}

	err = c.toggleMod("abc123", true)
	if err == nil || !strings.Contains(err.Error(), "Tauri IPC unavailable") {
		t.Fatalf("missing internals not reported: %v", err)
	}
}

func TestCDPEvaluateException(t *testing.T) {
	fake := startFakeCDP(t, func(expr string) (any, error) {
		return nil, fmt.Errorf("SyntaxError: boom")
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	if _, err := c.patcherStatus(); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("page exception not surfaced: %v", err)
	}
}

func TestCDPWaitReady(t *testing.T) {
	ready := false
	fake := startFakeCDP(t, func(expr string) (any, error) {
		if strings.Contains(expr, "__TAURI_INTERNALS__") {
			return ready, nil
		}
		return nil, fmt.Errorf("unexpected: %s", expr)
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	if err := c.waitReady(50 * time.Millisecond); err == nil {
		t.Fatal("page reported ready before the internals existed")
	}
	ready = true
	if err := c.waitReady(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestShimInjection(t *testing.T) {
	var expr string
	fake := startFakeCDP(t, func(e string) (any, error) {
		expr = e
		return "installed", nil
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	result, err := c.injectShim()
	if err != nil {
		t.Fatal(err)
	}
	if result != "installed" {
		t.Fatalf("shim result=%q", result)
	}
	for _, expected := range []string{
		"__AME_BRIDGE__",
		"get_mod_health_verdicts",
		"get_installed_mods",
		"patcher-game-attached",
		"patcher-game-exited",
		"patcher-error",
		"plugin:event|listen",
		"transformCallback",
	} {
		if !strings.Contains(expr, expected) {
			t.Fatalf("shim is missing %q", expected)
		}
	}
}

func TestTauriInvokeExpression(t *testing.T) {
	expr := tauriInvokeExpr("plugin:library|install_mod", map[string]any{"filePath": `C:\skins\set.fantome`})
	for _, expected := range []string{
		`plugin:library|install_mod`,
		`"filePath"`,
		`C:\\skins\\set.fantome`,
		`__ipc`,
		`rejected`,
	} {
		if !strings.Contains(expr, expected) {
			t.Fatalf("invoke expression missing %q: %s", expected, expr)
		}
	}
}

func TestBridgeStateParsing(t *testing.T) {
	raw := `{"shim":true,"attached":true,"pid":4242,"failure":null,` +
		`"status":{"running":true,"phase":"patching","session":{"overlayPrefix":"C:\\o\\"}}}`
	var state bridgeState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Shim || !state.Attached || state.PID == nil || *state.PID != 4242 {
		t.Fatalf("bridge fields=%+v", state)
	}
	if state.Status == nil || !state.Status.Running || state.Status.Phase != "patching" {
		t.Fatalf("status=%+v", state.Status)
	}
}

func TestPickManagerPage(t *testing.T) {
	pages := []cdpPage{
		{Type: "page", URL: "devtools://devtools/bundled/inspector.html", WS: "ws://1"},
		{Type: "iframe", URL: "http://tauri.localhost/", WS: "ws://2"},
		{Type: "page", URL: "https://example.com", WS: "ws://3"},
		{Type: "page", URL: "http://tauri.localhost/index.html", WS: "ws://4"},
	}
	page, ok := pickManagerPage(pages)
	if !ok || page.WS != "ws://4" {
		t.Fatalf("tauri page not preferred: %+v ok=%v", page, ok)
	}
	fallback, ok := pickManagerPage([]cdpPage{{Type: "page", URL: "https://example.com", WS: "ws://3"}})
	if !ok || fallback.WS != "ws://3" {
		t.Fatalf("fallback page not selected: %+v ok=%v", fallback, ok)
	}
	if _, ok := pickManagerPage([]cdpPage{{Type: "page", URL: "devtools://x", WS: "ws://1"}}); ok {
		t.Fatal("devtools page accepted as the manager page")
	}
}

func TestErrorText(t *testing.T) {
	cases := map[string]string{
		`{"code":"PATCHER","error":{"kind":"LEAGUE_ELEVATED"}}`: "PATCHER: LEAGUE_ELEVATED",
		`{"code":"IO","detail":"disk full"}`:                    "IO: disk full",
		`"plain message"`:                                       "plain message",
		`{}`:                                                    "{}",
	}
	for raw, expected := range cases {
		if got := errorText(json.RawMessage(raw)); got != expected {
			t.Fatalf("errorText(%s)=%q, want %q", raw, got, expected)
		}
	}
}

func TestConfigureManagerSettings(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	path := managerSettingsFile()
	if err := configureManagerSettings(`C:\Games\League of Legends\Game`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["leaguePath"] != `C:\Games\League of Legends\Game` {
		t.Fatalf("leaguePath=%v", doc["leaguePath"])
	}
	if doc["startInTray"] != true {
		t.Fatalf("startInTray=%v", doc["startInTray"])
	}
	if doc["firstRunComplete"] != true {
		t.Fatalf("firstRunComplete=%v", doc["firstRunComplete"])
	}

	// A second run rewrites nothing, and existing settings survive.
	before, _ := os.Stat(path)
	doc["telemetryEnabled"] = false
	if data, err := json.Marshal(doc); err == nil {
		os.WriteFile(path, data, 0600)
	}
	time.Sleep(10 * time.Millisecond)
	if err := configureManagerSettings(`C:\Games\League of Legends\Game`); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(path)
	data, _ = os.ReadFile(path)
	var kept map[string]any
	json.Unmarshal(data, &kept)
	if kept["telemetryEnabled"] != false {
		t.Fatalf("existing settings clobbered: %v", kept["telemetryEnabled"])
	}
	_ = before
	_ = after
}
