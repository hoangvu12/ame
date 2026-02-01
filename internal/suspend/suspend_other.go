//go:build !windows

package suspend

import (
	"fmt"
	"time"
)

type Suspender struct{}

func FindProcess(name string) uint32 { return 0 }

func WaitForProcess(name string, timeout time.Duration) (uint32, error) {
	return 0, fmt.Errorf("suspend not supported on this platform")
}

func NewSuspender(pid uint32) (*Suspender, error) {
	return nil, fmt.Errorf("suspend not supported on this platform")
}

func (s *Suspender) Suspend() (int, error) { return 0, nil }
func (s *Suspender) Resume() error         { return nil }
func (s *Suspender) Close()                {}
