package ltk

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func pauseOwnedScanner(process *os.Process) (func() error, error) {
	h, err := windows.OpenProcess(0x0800, false, uint32(process.Pid)) // PROCESS_SUSPEND_RESUME
	if err != nil {
		return nil, err
	}
	nt := windows.NewLazySystemDLL("ntdll.dll")
	status, _, _ := nt.NewProc("NtSuspendProcess").Call(uintptr(h))
	if status != 0 {
		windows.CloseHandle(h)
		return nil, fmt.Errorf("pause owned LTK scanner: NTSTATUS 0x%x", status)
	}
	return func() error {
		defer windows.CloseHandle(h)
		status, _, _ := nt.NewProc("NtResumeProcess").Call(uintptr(h))
		if status != 0 {
			return fmt.Errorf("resume owned LTK scanner: NTSTATUS 0x%x", status)
		}
		return nil
	}, nil
}

// The signed scanner hooks this window. Once it exists, conservatively refuse
// publication: hook output may still be in the pipe when the host was paused.
func gameWindowExists() bool {
	title, _ := windows.UTF16PtrFromString("League of Legends (TM) Client")
	window, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("FindWindowW").Call(0, uintptr(unsafe.Pointer(title)))
	return window != 0
}
