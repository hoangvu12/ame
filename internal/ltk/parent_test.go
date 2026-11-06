package ltk

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/winproc"
)

// procBasicInformation mirrors PROCESS_BASIC_INFORMATION for the
// InheritedFromUniqueProcessId field: the pid the patcher DLL's ancestor
// chain verification walks.
type procBasicInformation struct {
	Reserved1                    uintptr
	PebBaseAddress               uintptr
	Reserved2                    [2]uintptr
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessID uintptr
}

var ntQueryInformationProcess = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtQueryInformationProcess")

// recordedParentPID returns the pid a process regards as its parent.
func recordedParentPID(t *testing.T, pid uint32) uint32 {
	t.Helper()
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		t.Fatalf("open pid %d: %v", pid, err)
	}
	defer windows.CloseHandle(handle)
	var info procBasicInformation
	status, _, _ := ntQueryInformationProcess.Call(
		uintptr(handle), 0, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info), 0)
	if status != 0 {
		t.Fatalf("NtQueryInformationProcess: 0x%x", status)
	}
	return uint32(info.InheritedFromUniqueProcessID)
}

// TestManagerLaunchParentsToShell verifies the manager launch attributes
// give the child the Windows shell as its recorded parent: the injected
// patcher DLL refuses to overlay a game unless every ancestor of its host
// is the real shell or a LeagueToolkit binary (VRF00000008), and ame
// itself is neither. The test skips when no shell is available (headless
// CI) since the launcher then falls back to a plain child.
func TestManagerLaunchParentsToShell(t *testing.T) {
	parent, err := shellParentHandle()
	if err != nil {
		t.Skipf("no Windows shell to parent to: %v", err)
	}
	explorerPID := winproc.FindPIDs("explorer.exe")
	explorerPath := ""
	for _, pid := range explorerPID {
		explorerPath = winproc.ImagePath(pid)
		if explorerPath != "" {
			break
		}
	}
	windows.CloseHandle(parent)
	if explorerPath == "" {
		t.Fatal("shell handle resolved but explorer.exe is not running")
	}

	cmd := exec.Command(`C:\Windows\System32\whoami.exe`)
	release := applyManagerLaunch(cmd)
	startErr := cmd.Start()
	release()
	if startErr != nil {
		t.Skipf("test child launch failed: %v", startErr)
	}
	defer func() { _ = cmd.Process.Kill() }()

	got := recordedParentPID(t, uint32(cmd.Process.Pid))
	gotPath := winproc.ImagePath(got)
	if gotPath == "" {
		t.Fatalf("child's recorded parent pid %d has no readable image", got)
	}
	if !strings.EqualFold(filepath.Base(gotPath), "explorer.exe") {
		t.Fatalf("child's recorded parent is %s (pid %d), want the Windows shell", gotPath, got)
	}
}
