package ltk

import (
	"testing"
	"time"
)

// Exercise the actual CDP failure path, not just the tracker's flag.
func TestCDPFailureInvalidatesPendingBuildMeasurement(t *testing.T) {
	isolateManagerState(t)
	fake := startFakeCDP(t, func(string) (any, error) { return nil, nil })
	c := dialFake(t, fake)
	edit := time.Now()
	snapshotMu.Lock()
	rebuild = newRebuildTracker(edit, "patching")
	rebuild.observe("building", false, edit.Add(time.Second))
	snapshotMu.Unlock()
	c.lock()
	c.fail()
	c.unlock()
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	if !rebuild.observe("patching", false, edit.Add(172*time.Second)) {
		t.Fatal("recovered scan not ready")
	}
	if rebuild.duration() != 0 {
		t.Fatal("CDP outage poisoned the rebuild estimate")
	}
}
