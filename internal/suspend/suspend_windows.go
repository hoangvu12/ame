//go:build windows

package suspend

import (
	"fmt"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	thSnapThread         = 0x00000004
	threadSuspendResume  = 0x0002
	processCreateThread  = 0x0002
	processVMOperation   = 0x0008
	processVMRead        = 0x0010
	processVMWrite       = 0x0020
	processQueryInfo     = 0x0400
	processDupHandle     = 0x0040
	processSuspendResume = 0x0800
)

var (
	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procCreateRemoteThread = kernel32.NewProc("CreateRemoteThread")
	procOpenThread         = kernel32.NewProc("OpenThread")
	procThread32First      = kernel32.NewProc("Thread32First")
	procThread32Next       = kernel32.NewProc("Thread32Next")
	procGetExitCodeThread  = kernel32.NewProc("GetExitCodeThread")
	procSuspendThread      = kernel32.NewProc("SuspendThread")
	procResumeThread       = kernel32.NewProc("ResumeThread")

	procDebugActiveProcess     = kernel32.NewProc("DebugActiveProcess")
	procDebugActiveProcessStop = kernel32.NewProc("DebugActiveProcessStop")
	procDebugSetKillOnExit     = kernel32.NewProc("DebugSetProcessKillOnExit")

	ntdll                = windows.NewLazySystemDLL("ntdll.dll")
	procNtSuspendProcess = ntdll.NewProc("NtSuspendProcess")
	procNtResumeProcess  = ntdll.NewProc("NtResumeProcess")
)

type threadEntry32 struct {
	Size           uint32
	Usage          uint32
	ThreadID       uint32
	OwnerProcessID uint32
	BasePri        int32
	DeltaPri       int32
	Flags          uint32
}

type frozenThread struct {
	localHandle  windows.Handle
	remoteHandle uintptr
	tid          uint32
}

type suspendMethod int

const (
	methodRemoteThread  suspendMethod = iota // CreateRemoteThread + DuplicateHandle
	methodDirectSuspend                      // SuspendThread called from our process
	methodNtSuspend                          // NtSuspendProcess
	methodDebugger                           // DebugActiveProcess
)

// Suspender manages suspending and resuming a process.
// Tries multiple methods in order: remote thread injection, direct SuspendThread,
// NtSuspendProcess, and DebugActiveProcess.
type Suspender struct {
	pid     uint32
	hProc   windows.Handle
	threads []frozenThread
	method  suspendMethod

	addrSuspendThread uintptr
	addrResumeThread  uintptr
	addrCloseHandle   uintptr
}

// FindProcess returns the PID of the first process matching name, or 0 if not found.
func FindProcess(name string) uint32 {
	snap, err := windows.CreateToolhelp32Snapshot(0x00000002, 0) // TH32CS_SNAPPROCESS
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snap)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	err = windows.Process32First(snap, &pe)
	for err == nil {
		if strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), name) {
			return pe.ProcessID
		}
		err = windows.Process32Next(snap, &pe)
	}
	return 0
}

// WaitForProcess polls for a process by name until found or timeout.
func WaitForProcess(name string, timeout time.Duration) (uint32, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pid := FindProcess(name); pid != 0 {
			return pid, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("timeout waiting for %s", name)
}

func resolveProc(name string) uintptr {
	k32, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
		return 0
	}
	addr, err := windows.GetProcAddress(k32, name)
	if err != nil {
		return 0
	}
	return addr
}

func getThreadIDs(pid uint32) ([]uint32, error) {
	snap, err := windows.CreateToolhelp32Snapshot(thSnapThread, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snap)

	var te threadEntry32
	te.Size = uint32(unsafe.Sizeof(te))
	var tids []uint32
	r1, _, _ := procThread32First.Call(uintptr(snap), uintptr(unsafe.Pointer(&te)))
	for r1 != 0 {
		if te.OwnerProcessID == pid {
			tids = append(tids, te.ThreadID)
		}
		te.Size = uint32(unsafe.Sizeof(te))
		r1, _, _ = procThread32Next.Call(uintptr(snap), uintptr(unsafe.Pointer(&te)))
	}
	return tids, nil
}

// callRemote executes funcAddr inside the target process with a single parameter.
// Returns the thread exit code.
func callRemote(hProc windows.Handle, funcAddr uintptr, param uintptr) (uint32, error) {
	hThread, _, err := procCreateRemoteThread.Call(
		uintptr(hProc), 0, 0, funcAddr, param, 0, 0,
	)
	if hThread == 0 {
		return 0, fmt.Errorf("CreateRemoteThread: %v", err)
	}
	defer windows.CloseHandle(windows.Handle(hThread))
	windows.WaitForSingleObject(windows.Handle(hThread), 10000)

	var exitCode uint32
	procGetExitCodeThread.Call(hThread, uintptr(unsafe.Pointer(&exitCode)))
	return exitCode, nil
}

// NewSuspender opens the target process and prepares for suspension.
func NewSuspender(pid uint32) (*Suspender, error) {
	access := uint32(processCreateThread | processVMOperation | processVMWrite |
		processVMRead | processQueryInfo | processDupHandle | processSuspendResume)
	hProc, err := windows.OpenProcess(access, false, pid)
	if err != nil {
		return nil, fmt.Errorf("OpenProcess: %w", err)
	}

	addrSuspend := resolveProc("SuspendThread")
	addrResume := resolveProc("ResumeThread")
	addrClose := resolveProc("CloseHandle")
	if addrSuspend == 0 || addrResume == 0 || addrClose == 0 {
		windows.CloseHandle(hProc)
		return nil, fmt.Errorf("failed to resolve kernel32 functions")
	}

	return &Suspender{
		pid:               pid,
		hProc:             hProc,
		addrSuspendThread: addrSuspend,
		addrResumeThread:  addrResume,
		addrCloseHandle:   addrClose,
	}, nil
}

// Suspend freezes the target process. Tries four methods in order:
// 1. Remote thread injection (CreateRemoteThread + DuplicateHandle)
// 2. Direct SuspendThread on each thread
// 3. NtSuspendProcess
// 4. DebugActiveProcess
func (s *Suspender) Suspend() (int, error) {
	tids, _ := getThreadIDs(s.pid)

	// Method 1: Remote thread injection
	if len(tids) > 0 {
		for _, tid := range tids {
			hLocal, _, _ := procOpenThread.Call(threadSuspendResume, 0, uintptr(tid))
			if hLocal == 0 {
				continue
			}

			var hRemote windows.Handle
			err := windows.DuplicateHandle(
				windows.CurrentProcess(),
				windows.Handle(hLocal),
				s.hProc,
				&hRemote,
				threadSuspendResume,
				false,
				0,
			)
			if err != nil {
				windows.CloseHandle(windows.Handle(hLocal))
				continue
			}

			exitCode, err := callRemote(s.hProc, s.addrSuspendThread, uintptr(hRemote))
			if err != nil || exitCode == 0xFFFFFFFF {
				callRemote(s.hProc, s.addrCloseHandle, uintptr(hRemote))
				windows.CloseHandle(windows.Handle(hLocal))
				continue
			}

			s.threads = append(s.threads, frozenThread{
				localHandle:  windows.Handle(hLocal),
				remoteHandle: uintptr(hRemote),
				tid:          tid,
			})
		}
		if len(s.threads) > 0 {
			s.method = methodRemoteThread
			return len(s.threads), nil
		}
	}

	// Method 2: Direct SuspendThread from our process
	if len(tids) > 0 {
		for _, tid := range tids {
			hThread, _, _ := procOpenThread.Call(threadSuspendResume, 0, uintptr(tid))
			if hThread == 0 {
				continue
			}
			ret, _, _ := procSuspendThread.Call(hThread)
			if ret == 0xFFFFFFFF {
				windows.CloseHandle(windows.Handle(hThread))
				continue
			}
			s.threads = append(s.threads, frozenThread{
				localHandle: windows.Handle(hThread),
				tid:         tid,
			})
		}
		if len(s.threads) > 0 {
			s.method = methodDirectSuspend
			return len(s.threads), nil
		}
	}

	// Method 3: NtSuspendProcess
	r, _, _ := procNtSuspendProcess.Call(uintptr(s.hProc))
	if r == 0 {
		s.method = methodNtSuspend
		return 1, nil
	}

	// Method 4: DebugActiveProcess
	r, _, _ = procDebugActiveProcess.Call(uintptr(s.pid))
	if r != 0 {
		procDebugSetKillOnExit.Call(0)
		s.method = methodDebugger
		return 1, nil
	}

	return 0, fmt.Errorf("all suspend methods failed")
}

// Close releases the process handle without resuming threads.
// Use this to clean up when Suspend() fails.
func (s *Suspender) Close() {
	if s.hProc != 0 {
		windows.CloseHandle(s.hProc)
		s.hProc = 0
	}
}

// Resume unfreezes the process and releases handles.
func (s *Suspender) Resume() error {
	switch s.method {
	case methodRemoteThread:
		for _, ft := range s.threads {
			callRemote(s.hProc, s.addrResumeThread, uintptr(ft.remoteHandle))
			callRemote(s.hProc, s.addrCloseHandle, uintptr(ft.remoteHandle))
			windows.CloseHandle(ft.localHandle)
		}
		s.threads = nil

	case methodDirectSuspend:
		for _, ft := range s.threads {
			procResumeThread.Call(uintptr(ft.localHandle))
			windows.CloseHandle(ft.localHandle)
		}
		s.threads = nil

	case methodNtSuspend:
		procNtResumeProcess.Call(uintptr(s.hProc))

	case methodDebugger:
		procDebugActiveProcessStop.Call(uintptr(s.pid))
	}

	windows.CloseHandle(s.hProc)
	s.hProc = 0
	return nil
}
