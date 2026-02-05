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
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/i18n"
	"github.com/hoangvu12/ame/internal/lcu"
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

// checkForUpdates checks for new versions. Downloads the update and either
// auto-restarts (when minimized or autoUpdate is on) or prompts the user.
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

	if !result.Downloaded || !updater.VerifyUpdateFile() {
		fmt.Println("  Download: https://github.com/hoangvu12/ame/releases/latest")
		fmt.Println()
		return
	}

	// Auto-update when running minimized or when the setting is enabled
	if minimized || config.AutoUpdate() {
		fmt.Println("  Applying update...")
		restartViaLauncher()
		return
	}

	fmt.Println()
	fmt.Print("  Update now? [Y/n]: ")

	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" || input == "y" || input == "yes" {
		fmt.Println("  Restarting...")
		restartViaLauncher()
	}
	fmt.Println()
}

// restartViaLauncher launches ame.exe (the launcher) which will apply the
// pending update and start the new core. Then exits the current process.
func restartViaLauncher() {
	// Find the launcher (ame.exe) — it's the exe that originally launched us,
	// or we can find it via the startup task / original path.
	// The simplest approach: look for ame.exe next to the user's original location.
	// Since we're running as ame_core.exe in AppData, we need the launcher path.
	// The launcher passes its own path to core via --launcher flag.
	launcherPath := getLauncherPath()
	if launcherPath == "" {
		fmt.Println("  ! Could not find launcher to restart")
		return
	}

	// Forward flags to launcher
	var args []string
	if minimized {
		args = append(args, "--minimized")
	}

	cmd := exec.Command(launcherPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("  ! Failed to restart: %v\n", err)
		return
	}
	os.Exit(0)
}

// getLauncherPath returns the path to ame.exe (the launcher).
// Core receives it via --launcher flag, falls back to scanning args.
func getLauncherPath() string {
	for i, arg := range os.Args {
		if arg == "--launcher" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
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

// clearConsole clears the console window using the Windows console API directly
func clearConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")
	fillChar := kernel32.NewProc("FillConsoleOutputCharacterW")
	fillAttr := kernel32.NewProc("FillConsoleOutputAttribute")
	setCursor := kernel32.NewProc("SetConsoleCursorPosition")

	handle, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)

	type coord struct{ X, Y int16 }
	type smallRect struct{ Left, Top, Right, Bottom int16 }
	type bufferInfo struct {
		Size       coord
		Cursor     coord
		Attrs      uint16
		Window     smallRect
		MaxSize    coord
	}

	var info bufferInfo
	getInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&info)))

	size := uintptr(info.Size.X) * uintptr(info.Size.Y)
	var written uint32
	origin := uintptr(0) // coord{0,0} packed as uint32

	fillChar.Call(uintptr(handle), uintptr(' '), size, origin, uintptr(unsafe.Pointer(&written)))
	fillAttr.Call(uintptr(handle), uintptr(info.Attrs), size, origin, uintptr(unsafe.Pointer(&written)))
	setCursor.Call(uintptr(handle), origin)
}

// promptSettings shows an interactive settings menu in the console
func promptSettings() {
	display.Pause()
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

	display.Resume()
}

func cleanup() {
	display.Pause()
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

// runLauncher runs in launcher mode: applies pending updates, bootstraps core, launches it.
func runLauncher() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("  ! Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Apply pending update (replaces ame_core.exe before it starts)
	if updater.ApplyPendingUpdate() {
		fmt.Println("  Update applied!")
	}

	// Bootstrap: copy self to ame_core.exe if it doesn't exist yet
	if err := updater.BootstrapCore(exePath); err != nil {
		fmt.Printf("  ! Failed to bootstrap core: %v\n", err)
		os.Exit(1)
	}

	// Build args for core: add --core flag and pass launcher path
	var args []string
	args = append(args, "--core")
	args = append(args, "--launcher", exePath)
	for _, arg := range os.Args[1:] {
		args = append(args, arg)
	}

	// Clear console before starting core so stale output from a previous
	// update cycle is removed.
	clearConsole()

	corePath := updater.CorePath()
	cmd := exec.Command(corePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("  ! Failed to start core: %v\n", err)
		os.Exit(1)
	}

	// Launcher exits — core runs independently
	os.Exit(0)
}

func main() {
	// Check for admin privileges first (both launcher and core need it)
	if !isAdmin() {
		runAsAdmin()
		os.Exit(0)
	}

	// Dev mode or --core flag: run as core (the actual app)
	// Otherwise: run as launcher
	if Version != "dev" && !hasFlag("--core") {
		runLauncher()
		return
	}

	// === CORE MODE ===
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

	// Load settings (migrates gamedir.txt → settings.json on first run)
	if err := config.Init(); err != nil {
		fmt.Printf("  ! Failed to load settings: %v\n", err)
	}

	// Initialize console locale (best-effort)
	i18n.Init()

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

	// Check for updates (auto-applies when minimized or autoUpdate is on)
	checkForUpdates()

	// Track activation state before setup
	wasActivated := setup.IsPenguActivated()

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

	// If Pengu was freshly activated (installed) and client is running, restart it
	if !wasActivated && setup.IsPenguActivated() && lcu.IsClientRunning() {
		fmt.Println("  Restarting League Client...")
		lcu.RestartClient()
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

	// Wire up uninstall callback to trigger graceful exit
	server.OnUninstall = func() {
		quitTray()
	}

	// Start WebSocket server in background
	go server.StartServer(PORT)

	display.Init(Version)
	display.Log("Started")

	// Run system tray (blocks until quit)
	runTray()

	// Cleanup when tray exits
	cleanup()
}
