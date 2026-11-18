//go:build windows

package skin

import (
	"os/exec"
	"syscall"
)

// hideHelper keeps the generator's console window from flashing.
func hideHelper(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
