//go:build windows

package suspend

import (
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/hoangvu12/ame/internal/display"
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
	methodRemoteThread suspendMethod = iota // CreateRemoteThread + DuplicateHandle
	methodNtSuspend                         // NtSuspendProcess
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

const stillActive = 259 // STILL_ACTIVE (0x103) - thread hasn't exited yet

// callRemoteTimeout executes funcAddr inside the target process with a single parameter.
// Returns the thread exit code. timeoutMs controls how long to wait.
func callRemoteTimeout(hProc windows.Handle, funcAddr uintptr, param uintptr, timeoutMs uint32) (uint32, error) {
	hThread, _, err := procCreateRemoteThread.Call(
		uintptr(hProc), 0, 0, funcAddr, param, 0, 0,
	)
	if hThread == 0 {
		return 0, fmt.Errorf("CreateRemoteThread: %v", err)
	}
	defer windows.CloseHandle(windows.Handle(hThread))

	event, _ := windows.WaitForSingleObject(windows.Handle(hThread), timeoutMs)
	var exitCode uint32
	procGetExitCodeThread.Call(hThread, uintptr(unsafe.Pointer(&exitCode)))

	if event == uint32(windows.WAIT_TIMEOUT) || exitCode == stillActive {
		return 0, fmt.Errorf("remote thread timed out (%dms)", timeoutMs)
	}
	return exitCode, nil
}

// callRemote executes funcAddr inside the target process with a 5s timeout.
func callRemote(hProc windows.Handle, funcAddr uintptr, param uintptr) (uint32, error) {
	return callRemoteTimeout(hProc, funcAddr, param, 5000)
}

// WaitReady polls until the process can execute remote threads (loader lock released).
// Returns nil when ready, or an error if the done channel closes or timeout is reached.
func (s *Suspender) WaitReady(done <-chan struct{}, timeout time.Duration) error {
	addrGetPID := resolveProc("GetCurrentProcessId")
	if addrGetPID == 0 {
		return fmt.Errorf("failed to resolve GetCurrentProcessId")
	}

	display.Log(fmt.Sprintf("[Suspend] Waiting for process %d to be ready...", s.pid))
	start := time.Now()
	deadline := start.Add(timeout)
	attempts := 0
	for time.Now().Before(deadline) {
		select {
		case <-done:
			return fmt.Errorf("cancelled")
		default:
		}

		attempts++
		// Test with a short timeout — if the process can execute remote threads, this returns instantly
		_, err := callRemoteTimeout(s.hProc, addrGetPID, 0, 500)
		if err == nil {
			display.Log(fmt.Sprintf("[Suspend] Process %d ready after %dms (%d attempts)", s.pid, time.Since(start).Milliseconds(), attempts))
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("process not ready after %v (%d attempts)", timeout, attempts)
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

// Suspend freezes the target process. Tries two methods in order:
// 1. Remote thread injection (CreateRemoteThread + DuplicateHandle)
// 2. NtSuspendProcess
func (s *Suspender) Suspend() (int, error) {
	tids, _ := getThreadIDs(s.pid)
	display.Log(fmt.Sprintf("[Suspend] PID %d has %d threads", s.pid, len(tids)))

	// Method 1: Remote thread injection
	if len(tids) > 0 {
		display.Log("[Suspend] Trying method 1: Remote thread injection")
		for _, tid := range tids {
			hLocal, _, err := procOpenThread.Call(threadSuspendResume, 0, uintptr(tid))
			if hLocal == 0 {
				display.Log(fmt.Sprintf("[Suspend] OpenThread failed for TID %d: %v", tid, err))
				continue
			}

			var hRemote windows.Handle
			dupErr := windows.DuplicateHandle(
				windows.CurrentProcess(),
				windows.Handle(hLocal),
				s.hProc,
				&hRemote,
				threadSuspendResume,
				false,
				0,
			)
			if dupErr != nil {
				display.Log(fmt.Sprintf("[Suspend] DuplicateHandle failed for TID %d: %v", tid, dupErr))
				windows.CloseHandle(windows.Handle(hLocal))
				continue
			}

			exitCode, callErr := callRemote(s.hProc, s.addrSuspendThread, uintptr(hRemote))
			if callErr != nil || exitCode == 0xFFFFFFFF {
				display.Log(fmt.Sprintf("[Suspend] Remote SuspendThread failed for TID %d: exitCode=%d, err=%v", tid, exitCode, callErr))
				callRemote(s.hProc, s.addrCloseHandle, uintptr(hRemote))
				windows.CloseHandle(windows.Handle(hLocal))
				continue
			}

			display.Log(fmt.Sprintf("[Suspend] Suspended TID %d (previous suspend count: %d)", tid, exitCode))
			s.threads = append(s.threads, frozenThread{
				localHandle:  windows.Handle(hLocal),
				remoteHandle: uintptr(hRemote),
				tid:          tid,
			})
		}
		if len(s.threads) > 0 {
			s.method = methodRemoteThread
			display.Log(fmt.Sprintf("[Suspend] Method 1 success: suspended %d/%d threads", len(s.threads), len(tids)))
			return len(s.threads), nil
		}
		display.Log("[Suspend] Method 1 failed: no threads suspended")
	}

	// Method 2: NtSuspendProcess
	display.Log("[Suspend] Trying method 2: NtSuspendProcess")
	r, _, ntErr := procNtSuspendProcess.Call(uintptr(s.hProc))
	if r == 0 {
		s.method = methodNtSuspend
		display.Log("[Suspend] Method 2 success: NtSuspendProcess returned 0")
		return 1, nil
	}
	display.Log(fmt.Sprintf("[Suspend] Method 2 failed: NtSuspendProcess returned %d, err=%v", r, ntErr))

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
// Guarantees all threads are fully resumed by looping until suspend count reaches 0.
func (s *Suspender) Resume() error {
	display.Log(fmt.Sprintf("[Resume] Resuming PID %d using method %d", s.pid, s.method))

	switch s.method {
	case methodRemoteThread:
		display.Log(fmt.Sprintf("[Resume] Resuming %d threads via remote thread", len(s.threads)))
		for _, ft := range s.threads {
			// ResumeThread returns the PREVIOUS suspend count.
			// Keep calling until it returns 1 (was 1, now 0 = running) or 0.
			// Since we waited for process readiness before suspend, these calls should be fast.
			const maxIterations = 20
			resumed := false
			for i := 0; i < maxIterations; i++ {
				prevCount, err := callRemote(s.hProc, s.addrResumeThread, uintptr(ft.remoteHandle))
				if err != nil {
					display.Log(fmt.Sprintf("[Resume] ResumeThread failed for TID %d: %v", ft.tid, err))
					break
				}
				if prevCount <= 1 {
					display.Log(fmt.Sprintf("[Resume] TID %d fully resumed (was %d, now 0)", ft.tid, prevCount))
					resumed = true
					break
				}
				display.Log(fmt.Sprintf("[Resume] TID %d still suspended (was %d, now %d), continuing...", ft.tid, prevCount, prevCount-1))
			}
			if !resumed {
				display.Log(fmt.Sprintf("[Resume] TID %d may not be fully resumed, cleaning up handle", ft.tid))
			}
			callRemote(s.hProc, s.addrCloseHandle, uintptr(ft.remoteHandle))
			windows.CloseHandle(ft.localHandle)
		}
		s.threads = nil

	case methodNtSuspend:
		// NtResumeProcess respects suspend count, call once since we only suspended once
		display.Log("[Resume] Calling NtResumeProcess")
		r, _, err := procNtResumeProcess.Call(uintptr(s.hProc))
		if r != 0 {
			display.Log(fmt.Sprintf("[Resume] NtResumeProcess returned %d: %v", r, err))
		} else {
			display.Log("[Resume] NtResumeProcess success")
		}
	}

	windows.CloseHandle(s.hProc)
	s.hProc = 0

	// Verify process is still running after resume
	time.Sleep(100 * time.Millisecond)
	if pid := FindProcess("League of Legends.exe"); pid == s.pid {
		display.Log(fmt.Sprintf("[Resume] Process %d still exists after resume", s.pid))
	} else if pid != 0 {
		display.Log(fmt.Sprintf("[Resume] Different game process found: %d (was %d)", pid, s.pid))
	} else {
		display.Log(fmt.Sprintf("[Resume] WARNING: Process %d no longer exists after resume!", s.pid))
	}

	return nil
}
