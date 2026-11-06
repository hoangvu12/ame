package ltk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/winproc"
)

var (
	modadvapi32             = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procOpenProcessToken    = modadvapi32.NewProc("OpenProcessToken")
	procDuplicateTokenEx    = modadvapi32.NewProc("DuplicateTokenEx")
	procSetTokenInformation = modadvapi32.NewProc("SetTokenInformation")
	procGetCurrentProcess   = modkernel32.NewProc("GetCurrentProcess")
)

const (
	TOKEN_DUPLICATE       = 0x0002
	TOKEN_QUERY           = 0x0008
	TOKEN_ADJUST_DEFAULT  = 0x0080
	TOKEN_ASSIGN_PRIMARY  = 0x0001
	TokenPrimary          = 1
	SecurityImpersonation = 2
	TokenIntegrityLevel   = 25
)

// applyManagerLaunch keeps the LTK Manager hidden and unelevated, and
// parents it to the Windows shell. The patcher refuses elevated sessions
// (ADR-0067/0074), so the manager runs with a medium-integrity token
// whenever one is available. The injected DLL also verifies its host's
// ancestor chain (VRF00000008: every ancestor must be the Windows shell or
// a LeagueToolkit binary; VRF00000009: signatures) before it overlays a
// game, and ame is neither — so the manager is launched with the real
// explorer.exe as its recorded parent, the exact chain a user gets by
// starting the manager themselves. The returned release function closes
// the shell handle once the process has started; it is a no-op when the
// shell could not be used and the manager launches as a plain child.
func applyManagerLaunch(cmd *exec.Cmd) func() {
	attr := &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	release := func() {}
	if parent, err := shellParentHandle(); err == nil {
		attr.ParentProcess = syscall.Handle(parent)
		release = func() { windows.CloseHandle(parent) }
	} else {
		// The chain check will refuse the overlay for this session; that
		// surfaces as a visible failure instead of a silent one.
		fmt.Fprintf(os.Stderr, "ame: manager parent shell unavailable: %v\n", err)
	}
	if token, err := mediumToken(); err == nil {
		attr.Token = token
	}
	cmd.SysProcAttr = attr
	return release
}

// shellParentHandle opens the Windows shell's explorer.exe for use as the
// manager's recorded parent. Only the real shell under the Windows
// directory qualifies: the DLL's allowlist checks the path, and a
// same-named process elsewhere would still be refused. The handle carries
// both PROCESS_CREATE_PROCESS (to attribute the child to it) and
// PROCESS_DUP_HANDLE (Go's StartProcess duplicates the child's standard
// handles into the recorded parent, and the launch fails with access
// denied without it).
func shellParentHandle() (windows.Handle, error) {
	windir := os.Getenv("windir")
	if windir == "" {
		windir = `C:\Windows`
	}
	for _, pid := range winproc.FindPIDs("explorer.exe") {
		path := winproc.ImagePath(pid)
		if path == "" || !strings.EqualFold(filepath.Dir(path), windir) {
			continue
		}
		handle, err := windows.OpenProcess(windows.PROCESS_CREATE_PROCESS|windows.PROCESS_DUP_HANDLE, false, pid)
		if err != nil {
			continue
		}
		return handle, nil
	}
	return 0, fmt.Errorf("the Windows shell is not running")
}

// mediumToken duplicates the current process token and lowers it to medium
// integrity. Returns a primary token that Go's exec.Cmd can use via
// SysProcAttr.Token to launch a process unelevated.
func mediumToken() (syscall.Token, error) {
	proc, _, _ := procGetCurrentProcess.Call()

	var hToken windows.Token
	r, _, errno := procOpenProcessToken.Call(proc, TOKEN_DUPLICATE|TOKEN_QUERY|TOKEN_ADJUST_DEFAULT, uintptr(unsafe.Pointer(&hToken)))
	if r == 0 {
		return 0, fmt.Errorf("OpenProcessToken: %v", errno)
	}

	var dupToken windows.Token
	r, _, errno = procDuplicateTokenEx.Call(
		uintptr(hToken),
		TOKEN_QUERY|TOKEN_ASSIGN_PRIMARY|TOKEN_DUPLICATE|TOKEN_ADJUST_DEFAULT,
		0,
		SecurityImpersonation,
		TokenPrimary,
		uintptr(unsafe.Pointer(&dupToken)),
	)
	windows.CloseHandle(windows.Handle(hToken))
	if r == 0 {
		return 0, fmt.Errorf("DuplicateTokenEx: %v", errno)
	}

	// TOKEN_MANDATORY_LABEL { SID_AND_ATTRIBUTES { PSID; DWORD Attributes } }
	// followed by the actual SID bytes. The SID pointer must point into
	// this same contiguous buffer.
	const seGroupIntegrity = 0x20
	sidData := []byte{
		0x01,                               // Revision
		0x01,                               // SubAuthorityCount
		0x00, 0x00, 0x00, 0x00, 0x00, 0x10, // Authority
		0x00, 0x20, 0x00, 0x00, // 8192 (SECURITY_MANDATORY_MEDIUM_RID)
	}

	// TOKEN_MANDATORY_LABEL: { SID_AND_ATTRIBUTES { PSID; DWORD } }
	// SID_AND_ATTRIBUTES is 16 bytes on x64 (pointer + uint32 + padding).
	const sidAndAttrSize = 16
	buf := make([]byte, sidAndAttrSize+len(sidData))

	// Sid pointer → offset within buf
	sidPtr := uintptr(unsafe.Pointer(&buf[sidAndAttrSize]))
	// Write SID_AND_ATTRIBUTES: Sid pointer (8 bytes) + Attributes (4 bytes) + padding (4 bytes)
	*(*uintptr)(unsafe.Pointer(&buf[0])) = sidPtr
	*(*uint32)(unsafe.Pointer(&buf[8])) = seGroupIntegrity
	// SID after the struct
	copy(buf[sidAndAttrSize:], sidData)

	r, _, errno = procSetTokenInformation.Call(
		uintptr(dupToken),
		TokenIntegrityLevel,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if r == 0 {
		windows.CloseHandle(windows.Handle(dupToken))
		return 0, fmt.Errorf("SetTokenInformation(TokenIntegrityLevel): %v", errno)
	}

	return syscall.Token(dupToken), nil
}
