package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// freePort asks the OS for an unused port and releases it again.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func TestShowHandlerAcksAndFocuses(t *testing.T) {
	called := 0
	OnShowRequested = func() { called++ }
	t.Cleanup(func() { OnShowRequested = nil })

	rec := httptest.NewRecorder()
	showHandler(rec, httptest.NewRequest(http.MethodGet, "/show", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != ShowAck {
		t.Errorf("body = %q, want %q", got, ShowAck)
	}
	if called != 1 {
		t.Errorf("OnShowRequested called %d times, want 1", called)
	}
}

// A second launch that was itself started minimized (the logon task firing
// while ame already runs) must confirm the instance without yanking its
// console into view.
func TestShowHandlerFocusZeroDoesNotRaiseWindow(t *testing.T) {
	called := 0
	OnShowRequested = func() { called++ }
	t.Cleanup(func() { OnShowRequested = nil })

	rec := httptest.NewRecorder()
	showHandler(rec, httptest.NewRequest(http.MethodGet, "/show?focus=0", nil))

	if got := strings.TrimSpace(rec.Body.String()); got != ShowAck {
		t.Errorf("body = %q, want %q — the ack is how a second instance tells ame apart from an unrelated process on the port", got, ShowAck)
	}
	if called != 0 {
		t.Errorf("OnShowRequested called %d times, want 0", called)
	}
}

// The command surface includes uninstall, which deletes directories and
// rewrites HKLM keys. It must not be reachable from the network.
func TestStartServerBindsLoopbackOnly(t *testing.T) {
	port := freePort(t)

	errc := make(chan error, 1)
	go func() { errc <- StartServer(port) }()

	waitForListener(t, fmt.Sprintf("127.0.0.1:%d", port))

	// Anything bound to 0.0.0.0 is also reachable on a routable local address.
	for _, ip := range routableIPv4(t) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			t.Errorf("server is reachable on %s — it should be bound to 127.0.0.1 only", ip)
		}
	}

	select {
	case err := <-errc:
		t.Fatalf("server exited early: %v", err)
	default:
	}
}

// A process that cannot bind the port cannot serve the plugin, so it must
// report that rather than lingering as a silent duplicate.
func TestStartServerReturnsErrorWhenPortIsTaken(t *testing.T) {
	port := freePort(t)

	blocker, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("occupy port: %v", err)
	}
	defer blocker.Close()

	// Shorter than listenTimeout would allow, so the test does not sit for the
	// full retry window; we only need to observe that it reports failure.
	done := make(chan error, 1)
	go func() { done <- StartServer(port) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("StartServer returned nil error while the port was occupied")
		}
	case <-time.After(listenTimeout + 5*time.Second):
		t.Fatal("StartServer neither bound nor reported failure")
	}
}

// The whole point of starting the server before setup is that the plugin can
// connect and read state immediately; only the commands that need mod-tools
// and the game install wait.
func TestReadyGateCoversOnlySetupDependentCommands(t *testing.T) {
	gated := []string{"apply", "prefetch", "cleanup", "unstuck", "roomPartySkin", "importCustomMod"}
	for _, cmd := range gated {
		if !commandsNeedingSetup[cmd] {
			t.Errorf("%q should be held back until setup finishes", cmd)
		}
	}

	// These are what the plugin sends the moment it connects.
	for _, cmd := range []string{"query", "getSettings", "getCustomMods", "getLogs", "getGamePath"} {
		if commandsNeedingSetup[cmd] {
			t.Errorf("%q must not be gated — it is part of the connect handshake", cmd)
		}
	}
}

func waitForListener(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			body := readBody(t, "http://"+addr+"/show?focus=0")
			conn.Close()
			if strings.TrimSpace(body) == ShowAck {
				return
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server never started listening on %s", addr)
}

func readBody(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

// routableIPv4 returns this machine's non-loopback IPv4 addresses.
func routableIPv4(t *testing.T) []string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skipf("cannot enumerate interfaces: %v", err)
	}
	var out []string
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
			continue
		}
		out = append(out, ipnet.IP.String())
	}
	if len(out) == 0 {
		t.Skip("no routable IPv4 address to test against")
	}
	return out
}
