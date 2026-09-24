//go:build !windows

package ltk

import "fmt"

func installProcessHooks(dir string, pid uint32) (func(), error) {
	return nil, fmt.Errorf("early hooks require Windows")
}
