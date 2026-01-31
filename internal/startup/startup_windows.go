//go:build windows

package startup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const taskName = "ame"

// Enable creates a Windows Task Scheduler task to run ame.exe at user logon
// with highest privileges (avoids UAC prompt).
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Remove existing task first (ignore errors)
	Disable()

	cmd := exec.Command("schtasks", "/create",
		"/tn", taskName,
		"/tr", fmt.Sprintf(`"%s" --minimized`, exePath),
		"/sc", "onlogon",
		"/rl", "highest",
		"/f",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks create failed: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// Disable removes the scheduled task.
func Disable() error {
	cmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run() // Ignore errors (task may not exist)
	return nil
}

// IsEnabled checks if the scheduled task exists.
func IsEnabled() bool {
	cmd := exec.Command("schtasks", "/query", "/tn", taskName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
