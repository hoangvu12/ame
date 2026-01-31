//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/server"
	"github.com/hoangvu12/ame/internal/setup"
	"github.com/hoangvu12/ame/internal/startup"
	"github.com/hoangvu12/ame/internal/updater"
)

// Version is set at build time via -ldflags
var Version = "dev"

const PORT = 18765

var minimized bool

// Setup URLs
var setupConfig = setup.Config{
	ToolsURL:  "https://raw.githubusercontent.com/Alban1911/Rose/main/injection/tools",
	PenguURL:  "https://github.com/PenguLoader/PenguLoader/releases/download/v1.1.6/pengu-loader-v1.1.6.zip",
	PluginURL: "https://github.com/hoangvu12/ame/releases/latest/download/plugin.zip",
}

// findDevSrcDir locates the local plugin source directory relative to the executable.
func findDevSrcDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	srcDir := filepath.Join(filepath.Dir(exe), "src")
	if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
		return srcDir
	}
	return ""
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

	if result.Downloaded && updater.VerifyUpdateFile() {
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

			fmt.Println("  Restarting...")

			// Use PowerShell with rename-based atomic swap (more reliable than copy-overwrite)
			// Pattern: rename old exe -> move new exe into place -> delete old -> launch
			// This avoids file locking issues since rename works even if file is "in use"
			psScript := fmt.Sprintf(
				`$host.UI.RawUI.WindowTitle = 'ame updater'; `+
					`Write-Host 'Updating ame...' -ForegroundColor Cyan; `+
					`$ErrorActionPreference = 'Stop'; `+
					`$updatePath = '%s'; `+
					`$exePath = '%s'; `+
					`$backupPath = $exePath + '.old'; `+
					`$maxRetries = 15; `+
					`$success = $false; `+
					`Write-Host "Update file: $updatePath"; `+
					`Write-Host "Target: $exePath"; `+
					`Remove-Item -Path $backupPath -Force -ErrorAction SilentlyContinue; `+
					`for ($i = 1; $i -le $maxRetries; $i++) { `+
					`Write-Host "Attempt $i/$maxRetries..." -ForegroundColor Yellow; `+
					`Start-Sleep -Seconds 2; `+
					`try { `+
					`Rename-Item -Path $exePath -NewName ($exePath + '.old') -Force; `+
					`Move-Item -Path $updatePath -Destination $exePath -Force; `+
					`$success = $true; `+
					`Write-Host 'Update successful!' -ForegroundColor Green; `+
					`break; `+
					`} catch { `+
					`Write-Host "  Error: $_" -ForegroundColor Red; `+
					`} `+
					`} `+
					`if ($success) { `+
					`Remove-Item -Path $backupPath -Force -ErrorAction SilentlyContinue; `+
					`Write-Host 'Launching new version...' -ForegroundColor Cyan; `+
					`Start-Process -FilePath $exePath -Verb RunAs; `+
					`} else { `+
					`if (Test-Path $backupPath) { Rename-Item -Path $backupPath -NewName $exePath -Force -ErrorAction SilentlyContinue }; `+
					`Write-Host 'Update failed after 30s. Please close the app fully and try again.' -ForegroundColor Red; `+
					`Read-Host 'Press Enter to exit'; `+
					`}`,
				updater.GetUpdateFilePath(), exePath)

			cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-NoExit", "-Command", psScript)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
			}
			if err := cmd.Start(); err != nil {
				fmt.Printf("  ! Failed to start updater: %v\n", err)
				return
			}
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

// clearConsole clears the console window
func clearConsole() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// promptSettings shows an interactive settings menu in the console
func promptSettings() {
	clearConsole()
	reader := bufio.NewReader(os.Stdin)

	gamePath := config.GamePath()
	if gamePath == "" {
		gamePath = "(not set)"
	}

	fmt.Println()
	fmt.Println("  Settings")
	fmt.Println("  ────────")
	fmt.Println()
	fmt.Printf("  1. Game Path: %s\n", gamePath)
	fmt.Println()
	fmt.Print("  Choice (or press Enter to go back): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		fmt.Println()
		game.PromptGameDir()
	}

	clearConsole()
	printBanner()
	fmt.Println("  Ready! Open League client to use skins.")
	fmt.Println()
}

func cleanup() {
	fmt.Println("\n  Shutting down...")
	server.HandleCleanup()
	killPenguLoader()
}

// hasFlag checks if a command-line flag is present.
func hasFlag(flag string) bool {
	for _, arg := range os.Args[1:] {
		if arg == flag {
			return true
		}
	}
	return false
}

// syncStartup ensures the Task Scheduler entry matches the saved setting.
func syncStartup() {
	enabled := config.StartWithWindows()
	registered := startup.IsEnabled()

	if enabled && !registered {
		startup.Enable()
	} else if !enabled && registered {
		startup.Disable()
	}
}

func main() {
	minimized = hasFlag("--minimized")

	// Initialize console handle for tray functionality
	initConsoleHandle()

	// Disable QuickEdit mode so terminal doesn't pause on click
	disableQuickEdit()

	// Disable close button - users must use tray menu to quit
	disableCloseButton()

	// Hide console immediately when launched minimized
	if minimized {
		hideConsole()
	}

	// Check for admin privileges
	if !isAdmin() {
		runAsAdmin()
		os.Exit(0)
	}

	// Load settings (migrates gamedir.txt → settings.json on first run)
	if err := config.Init(); err != nil {
		fmt.Printf("  ! Failed to load settings: %v\n", err)
	}

	// Sync startup registration with saved setting
	syncStartup()

	// Print banner
	printBanner()

	// Handle process exit signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cleanup()
		quitTray()
	}()

	// Dev mode: use local plugin source instead of downloading from cloud
	if Version == "dev" {
		if srcDir := findDevSrcDir(); srcDir != "" {
			setupConfig.DevSrcDir = srcDir
		}
	}

	// Check if plugin reinstall is needed (after an update)
	if updater.NeedsPluginReinstall(Version) {
		fmt.Println("  Updating plugin...")
		setup.DetectAndSetPenguPaths()
		setup.SetupPlugin(setupConfig.PluginURL)
		updater.SaveVersion(Version)
		updater.CleanupUpdateFile()
	}

	// Skip interactive update prompt when running minimized
	if !minimized {
		checkForUpdates()
	}

	// Run setup (downloads dependencies if needed)
	if !setup.RunSetup(setupConfig) {
		fmt.Println()
		fmt.Println("  ! Setup failed. Check your internet connection.")
		if !minimized {
			fmt.Println("  Press Enter to exit...")
			fmt.Scanln()
		}
		os.Exit(1)
	}

	// Save current version after successful setup
	updater.SaveVersion(Version)

	// Detect game directory (skip interactive prompt when minimized)
	fmt.Print("  Detecting League of Legends... ")
	if dir := game.FindGameDir(); dir != "" {
		fmt.Println("found")
	} else {
		fmt.Println("not found")
		if !minimized {
			fmt.Println()
			game.PromptGameDir()
		}
	}

	// Start WebSocket server in background
	go server.StartServer(PORT)

	fmt.Println()
	fmt.Println("  +-----------------------------------------+")
	fmt.Println("  |  Ready! Open League client to use skins |")
	fmt.Println("  |                                         |")
	fmt.Println("  |  Keep this window running.              |")
	fmt.Println("  |  To quit: right-click tray icon > Quit  |")
	fmt.Println("  +-----------------------------------------+")
	fmt.Println()

	// Run system tray (blocks until quit)
	runTray()

	// Cleanup when tray exits
	cleanup()
}
