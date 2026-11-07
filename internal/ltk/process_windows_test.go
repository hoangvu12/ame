package ltk

import (
	"strings"
	"testing"
)

func TestAfterGameStopOnlyKillsTheManager(t *testing.T) {
	script := afterGameScript(4242)
	if !strings.Contains(script, "taskkill /PID 4242 /FI 'IMAGENAME eq ltk-manager.exe' /T /F") {
		t.Fatalf("script=%q", script)
	}
}
