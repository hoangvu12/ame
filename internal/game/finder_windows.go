//go:build windows

package game

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// getSysProcAttr returns Windows-specific process attributes to hide console window
func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

// logicalFixedDrives returns the roots of every local fixed disk ("C:\", ...).
//
// Only DRIVE_FIXED letters are returned. Network drives are excluded because a
// disconnected mapping blocks on access for the full SMB timeout, and removable
// drives because probing an empty optical drive can spin it up.
func logicalFixedDrives() []string {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil
	}

	var drives []string
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		root := string(rune('A'+i)) + `:\`
		rootPtr, err := windows.UTF16PtrFromString(root)
		if err != nil {
			continue
		}
		if windows.GetDriveType(rootPtr) == windows.DRIVE_FIXED {
			drives = append(drives, root)
		}
	}
	return drives
}
