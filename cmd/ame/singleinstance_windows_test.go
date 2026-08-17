//go:build windows

package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/sys/windows"

	"github.com/hoangvu12/ame/internal/server"
)

func uniqueMutexName(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf(`Local\ame-test-%s-%d`, t.Name(), os.Getpid())
}

func TestClaimNamedMutexIsExclusive(t *testing.T) {
	name := uniqueMutexName(t)

	handle, ok := claimNamedMutex(name)
	if !ok {
		t.Fatal("first claim failed — nothing else should hold this name")
	}
	if handle == 0 {
		t.Fatal("first claim returned no handle to hold the lock open")
	}
	defer windows.CloseHandle(handle)

	if _, ok := claimNamedMutex(name); ok {
		t.Error("second claim succeeded — two instances would run concurrently")
	}
}

// The kernel drops the mutex when its last handle closes, which is what makes a
// crashed instance unable to lock out the next one.
func TestClaimNamedMutexReleasesOnClose(t *testing.T) {
	name := uniqueMutexName(t)

	handle, ok := claimNamedMutex(name)
	if !ok || handle == 0 {
		t.Fatal("first claim failed")
	}
	if _, ok := claimNamedMutex(name); ok {
		t.Fatal("second claim should have been refused while the first was held")
	}

	windows.CloseHandle(handle)

	reclaimed, ok := claimNamedMutex(name)
	if !ok {
		t.Error("could not reclaim after release — a crashed instance would lock ame out permanently")
	}
	if reclaimed != 0 {
		windows.CloseHandle(reclaimed)
	}
}

// requestShow must recognise a real ame and, just as importantly, must not
// mistake an unrelated program holding the port for one.
func TestRequestShowDistinguishesAmeFromOtherProcesses(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"running ame", server.ShowAck, true},
		{"unrelated process on the port", "hello from something else", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewUnstartedServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(tc.body))
				}))

			ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", PORT))
			if err != nil {
				t.Skipf("port %d is in use (ame may be running): %v", PORT, err)
			}
			srv.Listener.Close()
			srv.Listener = ln
			srv.Start()
			defer srv.Close()

			if got := requestShow(false); got != tc.want {
				t.Errorf("requestShow() = %v, want %v", got, tc.want)
			}
		})
	}
}

// Nothing listening must not be reported as a running instance, or a genuine
// first launch would exit immediately instead of starting.
func TestRequestShowFalseWhenNothingListening(t *testing.T) {
	if ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", PORT)); err != nil {
		t.Skipf("port %d is in use (ame may be running)", PORT)
	} else {
		ln.Close()
	}

	if requestShow(false) {
		t.Error("requestShow() = true with nothing listening")
	}
}
