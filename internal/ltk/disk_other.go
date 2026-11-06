//go:build !windows

package ltk

import "math"

var diskFree = func(path string) (uint64, error) { return math.MaxUint64, nil }
