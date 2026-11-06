//go:build !windows

package ltk

import "os/exec"

func hide(cmd *exec.Cmd) {}

var stopAfterGame = func(cmd *exec.Cmd) {}
