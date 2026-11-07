package ltk

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/hoangvu12/ame/internal/display"
)

func hide(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

// stopAfterGame ends the manager's process tree once no League game process
// is left, from a hidden PowerShell that outlives ame. A variable so tests
// never start it.
var stopAfterGame = func(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	script := afterGameScript(cmd.Process.Pid)
	watcher := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
	hide(watcher)
	if err := watcher.Start(); err != nil {
		display.Log("LTK: could not schedule the manager stop after the game: " + err.Error())
		return
	}
	go func() { _ = watcher.Process.Release() }()
}

// afterGameScript waits out the game, then kills the manager by pid. A new
// ame may have replaced that manager meanwhile and its pid gone to another
// program, so the kill is limited to a process still named like the manager.
func afterGameScript(pid int) string {
	return fmt.Sprintf("while (Get-Process -Name 'League of Legends' -ErrorAction SilentlyContinue) { Start-Sleep -Seconds 3 }; Start-Sleep -Seconds 5; taskkill /PID %d /FI 'IMAGENAME eq %s' /T /F", pid, managerExeName)
}
