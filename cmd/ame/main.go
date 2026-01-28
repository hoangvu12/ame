//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/server"
	"github.com/hoangvu12/ame/internal/setup"
)

const PORT = 18765

// Setup URLs
var setupConfig = setup.Config{
	ToolsURL:  "https://raw.githubusercontent.com/Alban1911/Rose/main/injection/tools",
	PenguURL:  "https://github.com/PenguLoader/PenguLoader/releases/download/v1.1.6/pengu-loader-v1.1.6.zip",
	PluginURL: "https://raw.githubusercontent.com/hoangvu12/ame/main/src",
}

// isAdmin checks if running with admin privileges
func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

// runAsAdmin restarts the program with admin privileges
func runAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, windows.SW_SHOWNORMAL)
	return err
}

// killPenguLoader kills Pengu Loader process on exit
func killPenguLoader() {
	cmd := exec.Command("taskkill", "/F", "/IM", "Pengu Loader.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run() // Ignore errors
}

// disableQuickEdit disables QuickEdit mode to prevent terminal from pausing on click
func disableQuickEdit() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle, _ := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)

	var mode uint32
	getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))

	// Disable ENABLE_QUICK_EDIT_MODE (0x0040) and ENABLE_EXTENDED_FLAGS (0x0080)
	mode &^= 0x0040 // Remove QUICK_EDIT
	mode |= 0x0080  // Add EXTENDED_FLAGS (required)

	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}

func main() {
	// Disable QuickEdit mode so terminal doesn't pause on click
	disableQuickEdit()

	// Check for admin privileges
	if !isAdmin() {
		fmt.Println("[ame] Requesting admin privileges...")
		err := runAsAdmin()
		if err != nil {
			fmt.Printf("[ame] Failed to elevate: %v\n", err)
			fmt.Println("[ame] Please run as administrator manually.")
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
		os.Exit(0)
	}

	// Handle process exit signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[ame] Shutting down...")
		server.HandleCleanup()
		killPenguLoader()
		os.Exit(0)
	}()

	// Run setup (downloads dependencies if needed)
	if !setup.RunSetup(setupConfig) {
		fmt.Println("[ame] Setup failed. Please check your internet connection and try again.")
		os.Exit(1)
	}

	fmt.Printf("[ame] Starting server on port %d...\n", PORT)

	// Start WebSocket server (blocks)
	server.StartServer(PORT)
}
