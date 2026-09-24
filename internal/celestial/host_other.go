//go:build !windows

package celestial

import "os/exec"

func configureHostProcess(cmd *exec.Cmd) {}
