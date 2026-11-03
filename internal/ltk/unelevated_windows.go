package ltk

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modadvapi32                        = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32                        = windows.NewLazySystemDLL("kernel32.dll")
	procOpenProcessToken               = modadvapi32.NewProc("OpenProcessToken")
	procDuplicateTokenEx               = modadvapi32.NewProc("DuplicateTokenEx")
	procSetTokenInformation            = modadvapi32.NewProc("SetTokenInformation")
	procGetCurrentProcess              = modkernel32.NewProc("GetCurrentProcess")
)

const (
	TOKEN_DUPLICATE      = 0x0002
	TOKEN_QUERY          = 0x0008
	TOKEN_ADJUST_DEFAULT = 0x0080
	TOKEN_ASSIGN_PRIMARY = 0x0001
	TokenPrimary         = 1
	SecurityImpersonation = 2
	TokenIntegrityLevel  = 25
)

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
		0x01,                   // Revision
		0x01,                   // SubAuthorityCount
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