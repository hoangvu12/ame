package ltk

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// The manager's scanner hooks this window. Once it exists, conservatively
// refuse publication: the skin set for the running match must not change.
func gameWindowExists() bool {
	title, _ := windows.UTF16PtrFromString("League of Legends (TM) Client")
	window, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("FindWindowW").Call(0, uintptr(unsafe.Pointer(title)))
	return window != 0
}
