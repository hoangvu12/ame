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
	status = "Waiting for client"
	skin = "None"
	overlay = "Inactive"
	logs = nil
	started = true
	paused = false

	render()
}

// SetStatus updates the connection status and re-renders.
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
		skin = "None"
	} else if chroma != "" {
		skin = fmt.Sprintf("%s (%s)", name, chroma)
	} else {
		skin = name
	}
	render()
}

// SetOverlay updates the overlay status and re-renders.
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

	// Move cursor to home position and clear screen
	b.WriteString("\033[H\033[J")

	b.WriteString("\n")
	b.WriteString("  ame " + version + "\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %-10s%s\n", "Status", status))
	b.WriteString(fmt.Sprintf("  %-10s%s\n", "Skin", skin))
	b.WriteString(fmt.Sprintf("  %-10s%s\n", "Overlay", overlay))
	b.WriteString("\n")
	b.WriteString("  Log\n")
	b.WriteString("  ---\n")

	for _, entry := range logs {
		b.WriteString(fmt.Sprintf("  %s  %s\n", entry.Time.Format("15:04:05"), entry.Message))
	}

	os.Stdout.WriteString(b.String())
}
