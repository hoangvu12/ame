//go:build !windows

package ltk

var gameWindowExists = func() bool { return false }
