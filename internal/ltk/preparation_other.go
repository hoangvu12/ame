//go:build !windows

package ltk

var gameWindowExists = func() bool { return false }

var gameProcessRunning = func() bool { return false }
