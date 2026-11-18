//go:build !windows

package skin

import "os/exec"

func hideHelper(cmd *exec.Cmd) {}
