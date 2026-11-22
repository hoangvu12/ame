package ltk

import (
	"testing"
	"time"
)

// A stalled DevTools page hides when the build really finished. Recovery
// must clear readiness, but must not train the budget on the connection gap.
func TestRebuildMeasurementExcludesObservationOutage(t *testing.T) {
	edit := time.Now()
	r := newRebuildTracker(edit, "patching")
	r.observe("building", false, edit.Add(time.Second))
	r.measurementLost = true // the status call timed out
	if !r.observe("patching", false, edit.Add(172*time.Second)) {
		t.Fatal("recovered scan must become ready")
	}
	if got := r.duration(); got != 0 {
		t.Fatalf("connection outage counted as rebuild time: %s", got)
	}
}

func TestRebuildMeasurementSurvivesNormalPolling(t *testing.T) {
	edit := time.Now()
	r := newRebuildTracker(edit, "patching")
	for i := 1; i <= 100; i++ {
		r.observe("building", false, edit.Add(time.Duration(i)*time.Second))
	}
	r.observe("patching", false, edit.Add(101*time.Second))
	if got := r.duration(); got != 101*time.Second {
		t.Fatalf("genuinely slow observed build lost: %s", got)
	}
}
