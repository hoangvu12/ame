//go:build windows

// Package winproc enumerates running processes through the Win32 Toolhelp API.
//
// It exists because wmic.exe — which this codebase used to shell out to — was
// removed from Windows 11 24H2/25H2 in August 2026 and is no longer available
// even as a Feature on Demand. On those builds every wmic call fails instantly
// and indistinguishably from "the process isn't running", which silently broke
// League client detection. Toolhelp is documented, needs no privileges, costs
// no process spawn, and cannot be uninstalled.
package winproc

import (
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// FindPIDs returns the process IDs of every running process whose executable
// name matches exeName (case-insensitive, e.g. "LeagueClientUx.exe").
func FindPIDs(exeName string) []uint32 {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	var pids []uint32
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), exeName) {
			pids = append(pids, entry.ProcessID)
		}
	}
	return pids
}

// IsRunning reports whether at least one process with exeName is running.
func IsRunning(exeName string) bool {
	return len(FindPIDs(exeName)) > 0
}

// ImagePath returns the full path to the executable backing pid.
//
// This replaces the wmic "get ExecutablePath" query. QueryFullProcessImageName
// only needs PROCESS_QUERY_LIMITED_INFORMATION, which an elevated process has
// for anything it can see, and unlike the PEB-walking alternative it is fully
// documented and immune to 32/64-bit mismatch.
func ImagePath(pid uint32) string {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(handle)

	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(handle, 0, &buf[0], &size); err != nil {
		return ""
	}
	return windows.UTF16ToString(buf[:size])
}

// FindImagePath returns the executable path of the first running process
// matching exeName, or "" when none is running or the path can't be read.
func FindImagePath(exeName string) string {
	for _, pid := range FindPIDs(exeName) {
		if path := ImagePath(pid); path != "" {
			return filepath.Clean(path)
		}
	}
	return ""
}
