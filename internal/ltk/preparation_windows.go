package ltk

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/winproc"
)

// The manager's scanner hooks this window. Once it exists, conservatively
// refuse publication: the skin set for the running match must not change. A
// variable so tests can run with the client open.
var gameWindowExists = func() bool {
	title, _ := windows.UTF16PtrFromString("League of Legends (TM) Client")
	window, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("FindWindowW").Call(0, uintptr(unsafe.Pointer(title)))
	return window != 0
}

// gameProcessRunning reports whether a League game process exists, window or
// not: the DLL's join check compares against the process's creation, so a
// game counts from the moment it is created. A variable so tests never
// depend on a game running on the test machine.
var gameProcessRunning = func() bool {
	return winproc.IsRunning("League of Legends.exe")
}
