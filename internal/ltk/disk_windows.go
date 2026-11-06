package ltk

import "golang.org/x/sys/windows"

// diskFree is the space free to this user on the drive holding path, which
// must exist. A variable for tests.
var diskFree = func(path string) (uint64, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var free, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(name, &free, &total, &totalFree); err != nil {
		return 0, err
	}
	return free, nil
}
