//go:build !windows

package ltk

import (
	"fmt"
	"os"
)

func pauseOwnedScanner(*os.Process) (func() error, error) {
	return nil, fmt.Errorf("LTK scanner preparation requires Windows")
}
func gameWindowExists() bool { return false }
