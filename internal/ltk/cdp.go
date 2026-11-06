// LTK Manager lifecycle and DevTools protocol client.
//
// ame launches the signed ltk-manager.exe as a hidden background helper with
// WebView2 remote debugging enabled, connects to its page over the Chrome
// DevTools protocol, and drives the same Tauri IPC commands the manager's own
// UI uses. The manager — not ame — owns the injection host, so the patcher
// DLL's parent-process verification accepts it.

package ltk

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/display"
)

// cdpPort is the fixed DevTools port the manager's WebView listens on. A
// process already bound to it aborts the launch with an error.
const cdpPort = 9222

// Timeouts for the manager handshake and the Tauri commands ame issues.
const (
	cdpHandshakeTimeout = 15 * time.Second // manager boot + WebView2 init
	cdpReadyTimeout     = 10 * time.Second // page internals + shim injection
	cdpInvokeTimeout    = 30 * time.Second
	cdpInstallTimeout   = 180 * time.Second // large skin packages copy slowly
	cdpStatusTimeout    = 10 * time.Second
)

// --- CDP client ---

// cdpClient is a WebSocket connection to the manager page's DevTools
// endpoint. Commands are serialized; each call owns the connection until its
// matching response arrives.
type cdpClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
	next int64
}

func cdpDial(wsURL string) (*cdpClient, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("CDP dial: %w", err)
	}
	return &cdpClient{conn: conn}, nil
}

func (c *cdpClient) lock()   { c.mu.Lock() }
func (c *cdpClient) unlock() { c.mu.Unlock() }

func (c *cdpClient) close() {
	c.lock()
	defer c.unlock()
	c.conn.Close()
}

type cdpCall struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// evaluate runs expr in the manager page and returns the value the
// expression resolved to.
func (c *cdpClient) evaluate(expr string, timeout time.Duration) (json.RawMessage, error) {
	c.lock()
	defer c.unlock()
	id := c.next
	c.next++
	c.conn.SetWriteDeadline(time.Now().Add(timeout))
	if err := c.conn.WriteJSON(cdpCall{
		ID:     id,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    expr,
			"awaitPromise":  true,
			"returnByValue": true,
		},
	}); err != nil {
		return nil, fmt.Errorf("CDP write: %w", err)
	}
	deadline := time.Now().Add(timeout)
	for {
		c.conn.SetReadDeadline(deadline)
		var msg struct {
			// Events carry no id; the pointer distinguishes an event from a
			// response whose id happens to be zero.
			ID     *int64          `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		if err := c.conn.ReadJSON(&msg); err != nil {
			return nil, fmt.Errorf("CDP read: %w", err)
		}
		if msg.ID == nil || *msg.ID != id {
			continue // an event, or a late response from an aborted call
		}
		var eval struct {
			Result struct {
				Type  string          `json:"type"`
				Value json.RawMessage `json:"value"`
			} `json:"result"`
			ExceptionDetails *struct {
				Text      string          `json:"text"`
				Exception json.RawMessage `json:"exception"`
			} `json:"exceptionDetails"`
		}
		if err := json.Unmarshal(msg.Result, &eval); err != nil {
			return nil, fmt.Errorf("CDP response: %w", err)
		}
		if eval.ExceptionDetails != nil {
			detail := eval.ExceptionDetails.Text
			if len(eval.ExceptionDetails.Exception) > 0 {
				detail = strings.TrimSpace(detail + " " + string(eval.ExceptionDetails.Exception))
			}
			return nil, fmt.Errorf("manager page error: %s", detail)
		}
		if eval.Result.Type == "undefined" {
			return json.RawMessage("null"), nil
		}
		return eval.Result.Value, nil
	}
}

// jsLiteral renders v as a JavaScript value literal. JSON literals are valid
// JavaScript except for two line separators, which are escaped.
func jsLiteral(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	literal := string(data)
	literal = strings.ReplaceAll(literal, "\u2028", "\\u2028")
	literal = strings.ReplaceAll(literal, "\u2029", "\\u2029")
	return literal
}

// tauriInvokeExpr builds the page expression that calls a Tauri command and
// reports the manager's IpcResult envelope without letting a rejection escape.
func tauriInvokeExpr(cmd string, args map[string]any) string {
	return `(async()=>{const I=window.__TAURI_INTERNALS__;` +
		`if(!I||typeof I.invoke!=="function")return{__ipc:false};` +
		`try{const v=await I.invoke(` + jsLiteral(cmd) + `,` + jsLiteral(args) + `);` +
		`return{__ipc:true,ok:!!(v&&v.ok===true),value:v?v.value:null,error:v?v.error:null}}` +
		`catch(e){return{__ipc:true,rejected:true,error:String(e&&e.message||e)}}})()`
}

// invokeTauri calls window.__TAURI_INTERNALS__.invoke(cmd, args) in the page
// and unwraps the manager's IpcResult envelope.
func (c *cdpClient) invokeTauri(cmd string, args map[string]any, timeout time.Duration) (json.RawMessage, error) {
	if args == nil {
		args = map[string]any{}
	}
	raw, err := c.evaluate(tauriInvokeExpr(cmd, args), timeout)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", shortCommand(cmd), err)
	}
	var env struct {
		IPC      bool            `json:"__ipc"`
		OK       bool            `json:"ok"`
		Value    json.RawMessage `json:"value"`
		Error    json.RawMessage `json:"error"`
		Rejected bool            `json:"rejected"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%s: %w", shortCommand(cmd), err)
	}
	if !env.IPC {
		return nil, fmt.Errorf("%s: Tauri IPC unavailable in manager page", shortCommand(cmd))
	}
	if !env.OK {
		return nil, fmt.Errorf("LTK manager %s failed: %s", shortCommand(cmd), errorText(env.Error))
	}
	return env.Value, nil
}

func shortCommand(cmd string) string {
	if i := strings.LastIndexByte(cmd, '|'); i >= 0 {
		return cmd[i+1:]
	}
	return cmd
}

// errorText flattens an AppErrorResponse (or a plain rejection message) into
// one readable line.
func errorText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "unknown error"
	}
	var resp struct {
		Code   string          `json:"code"`
		Detail string          `json:"detail"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &resp); err == nil {
		var detail strings.Builder
		if resp.Code != "" {
			detail.WriteString(resp.Code)
		}
		if len(resp.Error) > 0 {
			var kind struct {
				Kind    string `json:"kind"`
				Reason  string `json:"reason"`
				Message string `json:"message"`
				Stage   string `json:"stage"`
			}
			if json.Unmarshal(resp.Error, &kind) == nil {
				for _, part := range []string{kind.Kind, kind.Stage, kind.Reason, kind.Message} {
					if part != "" {
						if detail.Len() > 0 {
							detail.WriteString(": ")
						}
						detail.WriteString(part)
					}
				}
			}
		}
		if resp.Detail != "" {
			if detail.Len() > 0 {
				detail.WriteString(": ")
			}
			detail.WriteString(resp.Detail)
		}
		if detail.Len() > 0 {
			return detail.String()
		}
	}
	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil && plain != "" {
		return plain
	}
	return string(raw)
}

// --- Tauri IPC helpers ---

// managerMod is the library view of one installed mod.
type managerMod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Enabled     bool   `json:"enabled"`
}

// isAmeMod reports whether a library mod is one ame installed. The manager
// normalizes the archive name (spaces become dashes), so both spellings are
// accepted.
func isAmeMod(mod managerMod, modID string) bool {
	if mod.ID != "" && mod.ID == modID {
		return true
	}
	normalize := func(name string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")
	}
	want := normalize(ameModName)
	return normalize(mod.Name) == want || normalize(mod.DisplayName) == want
}

// patcherStatus is the manager's patcher session state.
type patcherStatus struct {
	Running bool   `json:"running"`
	Phase   string `json:"phase"` // idle | building | patching
	Session *struct {
		OverlayPrefix string `json:"overlayPrefix"`
	} `json:"session"`
	// ChangesAfterGame: a library edit waits for the running game to end
	// before the overlay is rebuilt; the game keeps the set it loaded.
	ChangesAfterGame bool `json:"changesAfterGame"`
}

// prefix is the overlay root the session patches against, if known.
func (s patcherStatus) prefix() string {
	if s.Session == nil {
		return ""
	}
	return s.Session.OverlayPrefix
}

// installMod installs a .fantome archive into the manager's library and
// returns the new mod's library id.
func (c *cdpClient) installMod(path string) (string, error) {
	raw, err := c.invokeTauri("plugin:library|install_mod", map[string]any{"filePath": path}, cdpInstallTimeout)
	if err != nil {
		return "", err
	}
	var outcome struct {
		Kind string `json:"kind"`
		Mod  struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"mod"`
	}
	if err := json.Unmarshal(raw, &outcome); err != nil {
		return "", fmt.Errorf("install_mod: %w", err)
	}
	if outcome.Mod.ID == "" {
		return "", fmt.Errorf("install_mod: manager returned no mod id (%s)", outcome.Kind)
	}
	return outcome.Mod.ID, nil
}

func (c *cdpClient) toggleMod(id string, enabled bool) error {
	_, err := c.invokeTauri("plugin:library|toggle_mod", map[string]any{"modId": id, "enabled": enabled}, cdpInvokeTimeout)
	return err
}

func (c *cdpClient) uninstallMod(id string) error {
	_, err := c.invokeTauri("plugin:library|uninstall_mod", map[string]any{"modId": id}, cdpInvokeTimeout)
	return err
}

func (c *cdpClient) installedMods() ([]managerMod, error) {
	raw, err := c.invokeTauri("plugin:library|get_installed_mods", nil, cdpInvokeTimeout)
	if err != nil {
		return nil, err
	}
	var mods []managerMod
	if err := json.Unmarshal(raw, &mods); err != nil {
		return nil, fmt.Errorf("get_installed_mods: %w", err)
	}
	return mods, nil
}

func (c *cdpClient) startPatcher() error {
	// Every PatcherConfig field is optional; an empty config keeps the
	// manager's stored preferences.
	_, err := c.invokeTauri("plugin:patcher|start_patcher", map[string]any{"config": map[string]any{}}, cdpInvokeTimeout)
	return err
}

// stopPatcher stops the patcher session. Stopping an idle patcher succeeds.
func (c *cdpClient) stopPatcher() error {
	_, err := c.invokeTauri("plugin:patcher|stop_patcher", nil, cdpInvokeTimeout)
	if err != nil && strings.Contains(err.Error(), "NOT_RUNNING") {
		return nil
	}
	return err
}

func (c *cdpClient) patcherStatus() (patcherStatus, error) {
	raw, err := c.invokeTauri("plugin:patcher|get_patcher_status", nil, cdpStatusTimeout)
	if err != nil {
		return patcherStatus{}, err
	}
	var status patcherStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return patcherStatus{}, fmt.Errorf("get_patcher_status: %w", err)
	}
	return status, nil
}

// --- health interop shim and event bridge ---

// managerShimJS installs window.__AME_BRIDGE__ once per manager session:
//
//   - IpcResult answers for plugin:library|get_mod_health_verdicts are
//     replaced with "healthy" verdicts for every installed mod, so the
//     manager's frontend never gates or sweeps the skin package ame builds
//     from extracted game files.
//   - The manager's patcher events are recorded on the bridge, which ame
//     polls for attachment and failure state.
const managerShimJS = `(function(){
const I=window.__TAURI_INTERNALS__;
if(!I||typeof I.invoke!=="function")return "internals-missing";
if(window.__AME_BRIDGE__)return "already";
const bridge={attached:false,pid:null,failure:null,overlay:null};
window.__AME_BRIDGE__=bridge;
const real=I.invoke.bind(I);
const healthy=id=>({modId:id,health:"healthy",fixable:0,counts:{fatals:0,errors:0,warnings:0,infos:0},checkedAt:new Date().toISOString(),basis:{build:null,manager:"",hashtables:null}});
I.invoke=function(cmd,args,options){
if(cmd==="plugin:library|get_mod_health_verdicts"){
return real("plugin:library|get_installed_mods",{},options).then(res=>{
const mods=(res&&res.ok&&Array.isArray(res.value))?res.value:[];
const out={};
for(const m of mods)out[m.id]=healthy(m.id);
return{ok:true,value:out};
},()=>({ok:false,error:{code:"IO",detail:"ame health shim failed"}}));
}
return real(cmd,args,options);
};
const listen=(name,fn)=>{
if(typeof I.transformCallback!=="function")return;
try{
const cb=I.transformCallback(fn,false);
real("plugin:event|listen",{event:name,target:{kind:"Any"},handler:cb}).catch(()=>{});
}catch(e){}
};
listen("patcher-game-attached",ev=>{
const p=(ev&&ev.payload!==undefined)?ev.payload:ev;
bridge.attached=true;
bridge.pid=(p&&typeof p.pid==="number")?p.pid:null;
});
listen("patcher-game-exited",()=>{bridge.attached=false;bridge.pid=null;bridge.overlay=null;});
listen("patcher-game-overlay",ev=>{
const p=(ev&&ev.payload!==undefined)?ev.payload:ev;
bridge.overlay=(p&&typeof p.outcome==="string")?p.outcome:null;
});
listen("overlay-progress",ev=>{
const p=(ev&&ev.payload!==undefined)?ev.payload:ev;
bridge.progress=(p&&typeof p==="object")?p:null;
});
listen("patcher-error",ev=>{
const p=(ev&&ev.payload!==undefined)?ev.payload:ev;
try{bridge.failure=(p&&typeof p==="object")?JSON.stringify(p):String(p);}catch(e){bridge.failure=String(p);}
});
return "installed";
})()`

// injectShim installs the bridge. Re-running it after a page reload is safe.
func (c *cdpClient) injectShim() (string, error) {
	raw, err := c.evaluate(managerShimJS, cdpInvokeTimeout)
	if err != nil {
		return "", err
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("shim injection: %w", err)
	}
	return result, nil
}

// clearBridgeFailure forgets the recorded session failure before a new
// patcher session starts.
func (c *cdpClient) clearBridgeFailure() {
	_, _ = c.evaluate(`(()=>{const b=window.__AME_BRIDGE__;if(b)b.failure=null;return true})()`, cdpStatusTimeout)
}

// overlayRebuild is the manager's full overlay rebuild that ame started in
// the page without awaiting it (see startOverlayRebuild).
type overlayRebuild struct {
	State    string          `json:"state"` // running | done | failed | missing
	Error    json.RawMessage `json:"error"`
	Progress *struct {
		Stage       string `json:"stage"`
		CurrentFile string `json:"currentFile"`
		Current     int    `json:"current"`
		Total       int    `json:"total"`
	} `json:"progress"`
}

// startOverlayRebuildExpression starts plugin:patcher|rebuild_overlay and
// returns at once. The command runs for as long as the map copies take
// (minutes on a slow disk), and evaluate holds the connection for a whole
// call, so the outcome is kept on the bridge and read with short polls.
const startOverlayRebuildExpression = `(()=>{
const I=window.__TAURI_INTERNALS__;
const b=window.__AME_BRIDGE__;
if(!I||typeof I.invoke!=="function"||!b)return "bridge-missing";
if(b.rebuild&&b.rebuild.state==="running")return "running";
const r={state:"running",error:null};
b.rebuild=r;b.progress=null;
I.invoke("plugin:patcher|rebuild_overlay",{}).then(v=>{
if(v&&v.ok===true){r.state="done"}else{r.state="failed";r.error=v?v.error:null}
},e=>{r.state="failed";r.error=String(e&&e.message||e)});
return "started"})()`

func (c *cdpClient) startOverlayRebuild() error {
	raw, err := c.evaluate(startOverlayRebuildExpression, cdpInvokeTimeout)
	if err != nil {
		return err
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("rebuild_overlay: %w", err)
	}
	if result != "started" {
		return fmt.Errorf("rebuild_overlay: %s", result)
	}
	return nil
}

// overlayRebuildState reads the rebuild startOverlayRebuild left on the
// bridge. "missing" means the page lost it (a reload drops the bridge).
func (c *cdpClient) overlayRebuildState() (overlayRebuild, error) {
	raw, err := c.evaluate(`(()=>{const b=window.__AME_BRIDGE__;if(!b||!b.rebuild)return{state:"missing"};return{state:b.rebuild.state,error:b.rebuild.error,progress:b.progress||null}})()`, cdpStatusTimeout)
	if err != nil {
		return overlayRebuild{}, err
	}
	var state overlayRebuild
	if err := json.Unmarshal(raw, &state); err != nil {
		return overlayRebuild{}, fmt.Errorf("rebuild poll: %w", err)
	}
	return state, nil
}

// bridgeState is one poll of the manager page: the event bridge plus the
// patcher status command.
type bridgeState struct {
	Shim        bool            `json:"shim"`
	Attached    bool            `json:"attached"`
	PID         *uint64         `json:"pid"`
	Failure     string          `json:"failure"`
	Overlay     string          `json:"overlay"`
	Status      *patcherStatus  `json:"status"`
	StatusError json.RawMessage `json:"statusError"`
}

// bridgeStateExpression reads the event bridge and the patcher status in one
// page round-trip.
const bridgeStateExpression = `(async()=>{
const I=window.__TAURI_INTERNALS__;
const b=window.__AME_BRIDGE__;
const o={shim:!!b};
if(b){o.attached=!!b.attached;o.pid=b.pid||null;o.failure=b.failure||null;o.overlay=b.overlay||null}
if(!I||typeof I.invoke!=="function")return o;
try{
const v=await I.invoke("plugin:patcher|get_patcher_status",{});
if(v&&v.ok===true){o.status=v.value}else if(v&&v.error){o.statusError=v.error}
}catch(e){o.statusError=String(e&&e.message||e)}
return o})()`

func (c *cdpClient) bridgeState() (bridgeState, error) {
	raw, err := c.evaluate(bridgeStateExpression, cdpStatusTimeout)
	if err != nil {
		return bridgeState{}, err
	}
	var state bridgeState
	if err := json.Unmarshal(raw, &state); err != nil {
		return bridgeState{}, fmt.Errorf("bridge poll: %w", err)
	}
	return state, nil
}

// waitReady blocks until the page exposes the Tauri internals.
func (c *cdpClient) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		raw, err := c.evaluate(`!!(window.__TAURI_INTERNALS__&&typeof window.__TAURI_INTERNALS__.invoke==="function"&&typeof window.__TAURI_INTERNALS__.transformCallback==="function")`, 5*time.Second)
		if err == nil {
			var ready bool
			if json.Unmarshal(raw, &ready) == nil && ready {
				return nil
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("manager page not ready: %v", lastErr)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// --- DevTools HTTP discovery ---

type cdpPage struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
	WS    string `json:"webSocketDebuggerUrl"`
}

func cdpHTTPGet(path string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d%s", cdpPort, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

// cdpFindPage waits for the manager's WebView page to appear in the DevTools
// target list. The webview briefly shows about:blank before the Tauri
// document loads, and a bridge built there dies with the navigation, so the
// Tauri page is preferred; a bare page is accepted only after waiting for it.
func cdpFindPage(timeout time.Duration) (cdpPage, error) {
	deadline := time.Now().Add(timeout)
	tauriDeadline := time.Now().Add(5 * time.Second)
	var fallback cdpPage
	haveFallback := false
	var lastErr error
	for {
		data, err := cdpHTTPGet("/json", 2*time.Second)
		if err == nil {
			var pages []cdpPage
			if json.Unmarshal(data, &pages) == nil {
				for _, page := range pages {
					if page.Type != "page" || page.WS == "" || strings.Contains(page.URL, "devtools") {
						continue
					}
					if strings.Contains(strings.ToLower(page.URL), "tauri") {
						return page, nil
					}
					if !haveFallback {
						fallback, haveFallback = page, true
					}
				}
				if haveFallback && time.Now().After(tauriDeadline) {
					return fallback, nil
				}
				lastErr = fmt.Errorf("no page target in %d targets", len(pages))
			} else {
				lastErr = fmt.Errorf("invalid /json response")
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if haveFallback {
				return fallback, nil
			}
			return cdpPage{}, fmt.Errorf("DevTools endpoint: %v", lastErr)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// --- manager process helpers ---

func managerPortInUse() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", cdpPort), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// killExistingManagers terminates every running ltk-manager.exe, including
// the copy the user may have open: the single-instance plugin would otherwise
// focus the existing window instead of answering on the DevTools port. A
// variable so tests can never reach a real manager.
var killExistingManagers = func() {
	cmd := exec.Command("taskkill", "/IM", managerExeName, "/F", "/T")
	hide(cmd)
	out, err := cmd.CombinedOutput()
	if err == nil && strings.Contains(string(out), "SUCCESS") {
		display.Log("LTK: terminated a running LTK Manager instance")
	}
}

// waitForPortFree waits for the DevTools port to be released after the
// previous manager instances were killed.
func waitForPortFree(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for managerPortInUse() {
		if time.Now().After(deadline) {
			return fmt.Errorf("port %d is still in use after terminating LTK Manager", cdpPort)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return nil
}

// launchManager starts the signed manager hidden, unelevated, and with
// WebView2 debugging on the fixed port. The --shell-relaunched flag skips
// the manager's own restart through the Windows shell: that copy would lose
// ame's environment and run detached from the manager session.
func launchManager(exe string) (*exec.Cmd, error) {
	cmd := exec.Command(exe, "--shell-relaunched")
	cmd.Dir = filepath.Dir(exe)
	cmd.Env = append(os.Environ(), "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS=--remote-debugging-port=9222")
	releaseParent := applyManagerLaunch(cmd)
	startErr := cmd.Start()
	// The parent handle is consumed by CreateProcess itself; the manager
	// keeps only the recorded parent pid after the launch returns.
	releaseParent()
	if startErr != nil {
		return nil, startErr
	}
	return cmd, nil
}

// managerSettingsFile is the manager's persisted settings, shared with any
// LTK Manager installation the user runs themselves.
func managerSettingsFile() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	return filepath.Join(appData, "dev.leaguetoolkit.manager", "settings.json")
}

// managerRequiredSettings are the settings.json fields the manager's
// Settings struct (v1.33.0 src-tauri/src/state.rs) has no default for, with
// the values a fresh install writes. A document missing one fails to parse
// ("missing field `theme`"); the manager then falls back to its defaults and
// overwrites the file, discarding the League path and tray setting below.
var managerRequiredSettings = map[string]any{
	"theme":              "system",
	"accentColor":        map[string]any{},
	"patchTft":           false,
	"migrationDismissed": false,
}

// configureManagerSettings points the manager at ame's game directory and
// keeps its window in the tray so nothing is shown to the user. The file is
// only written before a launch: the manager owns it while it runs.
func configureManagerSettings(game string) error {
	path := managerSettingsFile()
	if path == "" {
		return nil
	}
	doc := map[string]any{}
	fresh := false
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			doc = map[string]any{}
		}
	} else {
		fresh = true
	}
	changed := false
	if game != "" {
		if current, _ := doc["leaguePath"].(string); current != game {
			doc["leaguePath"] = game
			changed = true
		}
	}
	if startInTray, _ := doc["startInTray"].(bool); !startInTray {
		doc["startInTray"] = true
		changed = true
	}
	if _, ok := doc["firstRunComplete"]; fresh || !ok {
		doc["firstRunComplete"] = true
		changed = true
	}
	for key, value := range managerRequiredSettings {
		if _, ok := doc[key]; !ok {
			doc[key] = value
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
