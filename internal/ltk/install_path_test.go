package ltk

import (
	"errors"
	"testing"
)

func TestInstallChangeNeverRestartsManagerDuringGame(t *testing.T) {
	isolateManagerState(t)
	previous := findGameDir
	findGameDir = func() string { return "D:/League/Game" }
	t.Cleanup(func() { findGameDir = previous })
	stubGame(t, func() bool { return true })
	manager = &managerSession{league: "C:/League/Game", stop: make(chan struct{}), exited: make(chan struct{})}
	fetched := false
	_, err := ensureManagerWith(func() (string, error) { fetched = true; return "", nil })
	if !errors.Is(err, ErrGameRunning) || fetched {
		t.Fatalf("unsafe manager replacement: err=%v fetched=%v", err, fetched)
	}
	select {
	case <-manager.stop:
		t.Fatal("stopped manager during match")
	default:
	}
}
