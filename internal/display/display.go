//go:build windows

package display

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
	"unicode/utf8"

	"github.com/hoangvu12/ame/internal/i18n"
)

const maxLogs = 8

type logEntry struct {
	Time    time.Time
	Message string
}

var (
	mu      sync.Mutex
	version string
	status  string
	skin    string
	overlay string
	party   string
	logs    []logEntry
	paused  bool
	started bool
)

// enableVT enables Virtual Terminal Processing on the Windows console,
// allowing ANSI escape codes to work.
func enableVT() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)

	var mode uint32
	getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}

// Init initializes the display with the given version and renders the initial view.
func Init(ver string) {
	mu.Lock()
	defer mu.Unlock()

	enableVT()

	version = ver
	status = i18n.T("display.value.waiting_client", nil)
	skin = i18n.T("display.value.none", nil)
	overlay = i18n.T("display.value.overlay_inactive", nil)
	party = i18n.T("display.value.party_off", nil)
	logs = nil
	started = true
	paused = false

	render()
}

// SetStatusKey updates the connection status and re-renders.
func SetStatusKey(key string, vars map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	status = i18n.T(key, vars)
	render()
}

// SetStatus updates the connection status and re-renders using a raw string.
func SetStatus(s string) {
	mu.Lock()
	defer mu.Unlock()

	status = s
	render()
}

// SetSkin updates the displayed skin name and re-renders.
// If name is empty, displays "None". If chroma is non-empty, appends it in parentheses.
func SetSkin(name, chroma string) {
	mu.Lock()
	defer mu.Unlock()

	if name == "" {
		skin = i18n.T("display.value.none", nil)
	} else if chroma != "" {
		skin = fmt.Sprintf("%s (%s)", name, chroma)
	} else {
		skin = name
	}
	render()
}

// SetPartyKey updates the room party status and re-renders.
func SetPartyKey(key string, vars map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	party = i18n.T(key, vars)
	render()
}

// SetParty updates the room party status and re-renders using a raw string.
func SetParty(s string) {
	mu.Lock()
	defer mu.Unlock()

	party = s
	render()
}

// SetOverlayKey updates the overlay status and re-renders.
func SetOverlayKey(key string, vars map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	overlay = i18n.T(key, vars)
	render()
}

// SetOverlay updates the overlay status and re-renders using a raw string.
func SetOverlay(s string) {
	mu.Lock()
	defer mu.Unlock()

	overlay = s
	render()
}

// Log adds a message to the activity log and re-renders.
func Log(msg string) {
	mu.Lock()
	defer mu.Unlock()

	logs = append(logs, logEntry{Time: time.Now(), Message: msg})
	if len(logs) > maxLogs {
		logs = logs[len(logs)-maxLogs:]
	}
	render()
}

// Pause stops rendering. State updates are still recorded
// but the display is not redrawn until Resume is called.
func Pause() {
	mu.Lock()
	defer mu.Unlock()

	paused = true
}

// Resume re-enables rendering and immediately redraws the display.
func Resume() {
	mu.Lock()
	defer mu.Unlock()

	paused = false
	render()
}

// render redraws the entire display using ANSI escape codes.
// Must be called while holding mu.
func render() {
	if !started || paused {
		return
	}

	var b strings.Builder

	labels := []string{
		i18n.T("display.label.status", nil),
		i18n.T("display.label.skin", nil),
		i18n.T("display.label.overlay", nil),
		i18n.T("display.label.party", nil),
	}
	labelWidth := 12
	for _, lbl := range labels {
		if w := utf8.RuneCountInString(lbl); w > labelWidth {
			labelWidth = w
		}
	}
	if labelWidth < 12 {
		labelWidth = 12
	} else if labelWidth > 24 {
		labelWidth = 24
	}

	// Move cursor to home position and clear screen
	b.WriteString("\033[H\033[J")

	b.WriteString("\n")
	b.WriteString("  ame " + version + "\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %-*s %s\n", labelWidth, labels[0], status))
	b.WriteString(fmt.Sprintf("  %-*s %s\n", labelWidth, labels[1], skin))
	b.WriteString(fmt.Sprintf("  %-*s %s\n", labelWidth, labels[2], overlay))
	b.WriteString(fmt.Sprintf("  %-*s %s\n", labelWidth, labels[3], party))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n", i18n.T("display.label.log", nil)))
	b.WriteString("  ---\n")

	for _, entry := range logs {
		b.WriteString(fmt.Sprintf("  %s  %s\n", entry.Time.Format("15:04:05"), entry.Message))
	}

	os.Stdout.WriteString(b.String())
}
