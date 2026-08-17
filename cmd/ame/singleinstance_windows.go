//go:build windows

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/server"
	"github.com/hoangvu12/ame/internal/winproc"
)

// Running two copies of ame is not merely redundant, it is destructive. Both
// write the whole of settings.json from their own stale snapshot, both delete
// and rebuild the shared overlay directory, and KillModTools kills mod-tools by
// image name — so the second copy tears down the first one's live overlay just
// by quitting. Only one may run.
const (
	// instanceMutexName is per-session. Port 18765 is loopback-only and the
	// League client runs in the user's own session, so a machine-wide lock
	// would add no protection while risking a service session locking out the
	// interactive user.
	instanceMutexName = `Local\ame-single-instance`

	// asfwAny grants foreground-activation rights to any process (ASFW_ANY).
	// Windows will not let the running instance raise its own window on
	// request — only a process that currently holds the foreground right can
	// hand it over, and that is the newly launched process, because the user
	// just clicked it. The grant lapses at the next user input.
	asfwAny = ^uintptr(0)

	// showProbeTimeout bounds a single /show request.
	showProbeTimeout = 700 * time.Millisecond

	// showRetryWindow is how long to keep asking a known-present instance to
	// show itself before treating it as wedged. It mostly covers the gap where
	// the running instance holds the lock but has not bound the port yet.
	showRetryWindow = 3 * time.Second

	// lockWaitTimeout is how long a post-update core waits for the outgoing
	// core to release the lock before giving up.
	lockWaitTimeout = 10 * time.Second

	mbIconWarning = 0x30
)

var (
	procAllowSetForegroundWindow = user32.NewProc("AllowSetForegroundWindow")
	procMessageBoxW              = user32.NewProc("MessageBoxW")
)

// instanceMutex is held for the process lifetime. The kernel releases it when
// the process dies, so unlike a lockfile it cannot go stale after a crash.
var instanceMutex windows.Handle

// claimNamedMutex creates a named mutex, reporting whether this process is the
// first to do so. The returned handle must stay open for as long as the claim
// should hold; the kernel closes it on process death, which is what makes this
// immune to the stale-lock problem a lockfile would have.
func claimNamedMutex(name string) (windows.Handle, bool) {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		// The name is a constant we control; if it cannot be encoded,
		// something is deeply wrong and blocking startup helps nobody.
		return 0, true
	}

	handle, err := windows.CreateMutex(nil, false, namePtr)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if handle != 0 {
			windows.CloseHandle(handle)
		}
		return 0, false
	}
	if handle == 0 {
		return 0, true
	}
	return handle, true
}

// acquireInstanceLock claims the single-instance lock. It reports false when
// another ame process already holds it.
func acquireInstanceLock() bool {
	handle, ok := claimNamedMutex(instanceMutexName)
	if !ok {
		return false
	}
	if handle != 0 {
		instanceMutex = handle
	}
	return true
}

// waitForInstanceLock retries acquireInstanceLock until it succeeds or timeout
// elapses. Used on the update handoff, where the outgoing core is already on
// its way out and the incoming core should wait for it rather than treating it
// as a duplicate and exiting — which would leave nothing running at all.
func waitForInstanceLock(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if acquireInstanceLock() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// requestShow pings a running ame, returning true once one has acknowledged.
// When focus is true the running instance also raises its console. A non-ame
// process holding the port will not produce the ack.
func requestShow(focus bool) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/show", PORT)
	if focus {
		// Hand over our foreground right before asking, not after: the running
		// instance calls SetForegroundWindow while servicing this request.
		procAllowSetForegroundWindow.Call(asfwAny)
	} else {
		url += "?focus=0"
	}

	client := &http.Client{Timeout: showProbeTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
	return resp.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) == server.ShowAck
}

// notifyRunningInstance keeps pinging a known-present instance. Unlike
// requestShow it retries, because the other instance may hold the lock while
// still starting up and not be listening yet.
func notifyRunningInstance(focus bool) bool {
	deadline := time.Now().Add(showRetryWindow)
	for {
		if requestShow(focus) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(150 * time.Millisecond)
	}
}

// showAlreadyRunningMessage reports a running-but-unresponsive instance.
//
// It deliberately does not offer to kill the other process. That instance may
// own a live mod-tools overlay and, worse, threads it suspended in a running
// game — state tracked only in its own memory. Killing it can leave the game
// permanently frozen with nothing left able to resume it.
func showAlreadyRunningMessage() {
	msg := "ame is already running but is not responding."
	if pids := winproc.FindPIDs("ame_core.exe"); len(pids) > 0 {
		msg = fmt.Sprintf("ame is already running (PID %d) but is not responding.", pids[0])
	}
	msg += "\n\nEnd it in Task Manager, then start ame again."

	text, err := windows.UTF16PtrFromString(msg)
	if err != nil {
		return
	}
	caption, err := windows.UTF16PtrFromString("ame")
	if err != nil {
		return
	}
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), mbIconWarning)
}
