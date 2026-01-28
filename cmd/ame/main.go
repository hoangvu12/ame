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
	"github.com/hoangvu12/ame/internal/updater"
)

// Version is set at build time via -ldflags
var Version = "dev"

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

// printBanner prints the application banner
func printBanner() {
	fmt.Println()
	fmt.Println("  ame " + Version)
	fmt.Println("  https://github.com/hoangvu12/ame")
	fmt.Println()
}

// checkForUpdates checks for new versions and prompts user to update
func checkForUpdates() {
	if Version == "dev" {
		return
	}

	fmt.Print("  Checking for updates... ")

	result, err := updater.CheckForUpdates(Version)
	if err != nil {
		fmt.Println("failed")
		return
	}

	if !result.UpdateAvailable {
		fmt.Println("up to date")
		return
	}

	fmt.Println("update available!")
	fmt.Println()
	fmt.Printf("  New version: %s (current: %s)\n", result.LatestVersion, Version)

	if result.Downloaded {
		fmt.Println()
		fmt.Print("  Update now? [Y/n]: ")

		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "" || input == "y" || input == "yes" {
			exePath, err := os.Executable()
			if err != nil {
				fmt.Printf("  ! Failed to update: %v\n", err)
				return
			}

			scriptPath, err := updater.ApplyUpdate(exePath)
			if err != nil {
				fmt.Printf("  ! Failed to update: %v\n", err)
				return
			}

			fmt.Println("  Restarting...")
			cmd := exec.Command("cmd", "/C", scriptPath)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd.Start()
			os.Exit(0)
		}
	} else {
		fmt.Println("  Download: https://github.com/hoangvu12/ame/releases/latest")
	}
	fmt.Println()
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
		runAsAdmin()
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Handle process exit signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n  Shutting down...")
		server.HandleCleanup()
		killPenguLoader()
		os.Exit(0)
	}()

	// Check if plugin reinstall is needed (after an update)
	if updater.NeedsPluginReinstall(Version) {
		fmt.Printf("  > Updating plugins (%s -> %s)...\n", updater.GetSavedVersion(), Version)
		setup.SetupPlugin(setupConfig.PluginURL)
		updater.SaveVersion(Version)
		updater.CleanupUpdateFile()
	}

	// Check for updates
	checkForUpdates()

	// Run setup (downloads dependencies if needed)
	if !setup.RunSetup(setupConfig) {
		fmt.Println()
		fmt.Println("  ! Setup failed. Check your internet connection.")
		fmt.Println("  Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

	// Save current version after successful setup
	updater.SaveVersion(Version)

	// Start WebSocket server (blocks)
	server.StartServer(PORT)
}
