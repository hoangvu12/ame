package celestial

import (
	"os/exec"
	"syscall"
)

func configureHostProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
