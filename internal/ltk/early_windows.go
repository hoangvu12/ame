package ltk

import (
	"fmt"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/hoangvu12/ame/internal/display"
	"golang.org/x/sys/windows"
)

// installProcessHooks uses the same signed hook procedure as the original host.
// This experiment requires initialized target GUI state, which is NOT guaranteed
// at process discovery. Production handoff must not call it on a held process.
func installProcessHooks(dir string, pid uint32) (func(), error) {
	if pid == 0 {
		return nil, fmt.Errorf("missing held process")
	}
	module, err := windows.LoadLibraryEx(filepath.Join(dir, "ltk_patcher_dll.dll"), 0, 1) // DONT_RESOLVE_DLL_REFERENCES
	if err != nil {
		return nil, err
	}
	u := windows.NewLazySystemDLL("user32.dll")
	set, unset := u.NewProc("SetWindowsHookExW"), u.NewProc("UnhookWindowsHookEx")
	var hooks []uintptr
	var wakeStop, wakeDone chan struct{}
	release := func() {
		if wakeStop != nil {
			close(wakeStop)
			<-wakeDone
			wakeStop = nil
		}
		for _, hook := range hooks {
			unset.Call(hook)
		}
		hooks = nil
		if module != 0 {
			windows.FreeLibrary(module)
			module = 0
		}
	}
	proc, err := windows.GetProcAddress(module, "ltk_patcher_hookproc")
	if err != nil {
		release()
		return nil, err
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(4, 0)
	if err != nil {
		release()
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	type threadEntry struct {
		Size, Usage, ID, Owner uint32
		Base, Delta            int32
		Flags                  uint32
	}
	var entry threadEntry
	entry.Size = uint32(unsafe.Sizeof(entry))
	k := windows.NewLazySystemDLL("kernel32.dll")
	first, next := k.NewProc("Thread32First"), k.NewProc("Thread32Next")
	times := k.NewProc("GetThreadTimes")
	var primary uint32
	oldest := ^uint64(0)
	ok, _, _ := first.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		if entry.Owner == pid {
			thread, queryErr := windows.OpenThread(0x0800, false, entry.ID)
			if queryErr != nil {
				release()
				return nil, fmt.Errorf("query held thread: %w", queryErr)
			}
			var created, exited, kernel, user windows.Filetime
			got, _, queryErr := times.Call(uintptr(thread), uintptr(unsafe.Pointer(&created)), uintptr(unsafe.Pointer(&exited)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user)))
			windows.CloseHandle(thread)
			if got == 0 {
				release()
				return nil, fmt.Errorf("query held thread age: %w", queryErr)
			}
			age := uint64(created.HighDateTime)<<32 | uint64(created.LowDateTime)
			if age < oldest {
				oldest, primary = age, entry.ID
			}
		}
		entry.Size = uint32(unsafe.Sizeof(entry))
		ok, _, _ = next.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&entry)))
	}
	if primary == 0 {
		release()
		return nil, fmt.Errorf("no threads for held PID %d", pid)
	}
	// The process's original thread will initialize the game window. Worker
	// threads can reject GUI hooks and must not determine readiness.
	hook, _, hookErr := set.Call(3, proc, uintptr(module), uintptr(primary))
	if hook == 0 {
		release()
		return nil, fmt.Errorf("early hook for PID %d primary thread %d: %w", pid, primary, hookErr)
	}
	hooks = append(hooks, hook)
	// The message queue may be created only after resume. Wake it once when
	// available so the hook executes on the first message retrieval.
	wakeStop, wakeDone = make(chan struct{}), make(chan struct{})
	stop, done := wakeStop, wakeDone
	go func() {
		defer close(done)
		post := u.NewProc("PostThreadMessageW")
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for {
			if posted, _, _ := post.Call(uintptr(primary), 0, 0, 0); posted != 0 {
				return
			}
			select {
			case <-stop:
				return
			case <-deadline.C:
				return
			case <-tick.C:
			}
		}
	}()
	display.Log(fmt.Sprintf("LTK: primary thread hook installed pid=%d tid=%d before resume", pid, primary))
	return release, nil
}
