//go:build windows

package main

import (
	_ "embed"
	"syscall"

	"github.com/energye/systray"
)

//go:embed icon.ico
var iconData []byte

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	user32              = syscall.NewLazyDLL("user32.dll")
	getConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
	showWindow          = user32.NewProc("ShowWindow")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)

const (
	SW_HIDE = 0
	SW_SHOW = 5
)

var consoleVisible = true
var consoleHwnd uintptr

// initConsoleHandle gets and stores the console window handle
func initConsoleHandle() {
	consoleHwnd, _, _ = getConsoleWindow.Call()
}

// disableCloseButton is a no-op, X button closes the app normally
func disableCloseButton() {
	// X button works normally - closes the app
}

// hideConsole hides the console window
func hideConsole() {
	if consoleHwnd != 0 {
		showWindow.Call(consoleHwnd, SW_HIDE)
		consoleVisible = false
	}
}

// showConsole shows the console window
func showConsole() {
	if consoleHwnd != 0 {
		showWindow.Call(consoleHwnd, SW_SHOW)
		setForegroundWindow.Call(consoleHwnd)
		consoleVisible = true
	}
}

// toggleConsole toggles console visibility
func toggleConsole() {
	if consoleVisible {
		hideConsole()
	} else {
		showConsole()
	}
}

// onTrayReady is called when the system tray is ready
func onTrayReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("ame")
	systray.SetTooltip("ame - League Skin Changer")

	// Double-click tray icon to toggle console visibility
	systray.SetOnDClick(func(menu systray.IMenu) {
		toggleConsole()
	})

	// Add menu items
	mToggle := systray.AddMenuItem("Show/Hide Console", "Toggle console window visibility")
	mToggle.Click(func() {
		toggleConsole()
	})

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit the application")
	mQuit.Click(func() {
		systray.Quit()
	})
}

// onTrayExit is called when the system tray is exiting
func onTrayExit() {
	// Cleanup will be handled by main
}

// runTray starts the system tray (blocks until quit)
func runTray() {
	systray.Run(onTrayReady, onTrayExit)
}

// quitTray signals the tray to quit
func quitTray() {
	systray.Quit()
}
