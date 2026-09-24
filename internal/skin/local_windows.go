//go:build windows

package skin

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

func hideHelper(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

func installedBuild(path string) string {
	dll := syscall.NewLazyDLL("version.dll")
	p, _ := syscall.UTF16PtrFromString(path)
	size, _, _ := dll.NewProc("GetFileVersionInfoSizeW").Call(uintptr(unsafe.Pointer(p)), 0)
	if size == 0 {
		return "unknown"
	}
	data := make([]byte, size)
	ok, _, _ := dll.NewProc("GetFileVersionInfoW").Call(uintptr(unsafe.Pointer(p)), 0, size, uintptr(unsafe.Pointer(&data[0])))
	if ok == 0 {
		return "unknown"
	}
	root, _ := syscall.UTF16PtrFromString(`\`)
	var info *[13]uint32
	var length uint32
	ok, _, _ = dll.NewProc("VerQueryValueW").Call(uintptr(unsafe.Pointer(&data[0])), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&length)))
	if ok == 0 || length < 52 {
		return "unknown"
	}
	return fmt.Sprintf("%d.%d.%d.%d", info[2]>>16, info[2]&65535, info[3]>>16, info[3]&65535)
}
