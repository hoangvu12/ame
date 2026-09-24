package display

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnosticsSurviveReopenAndRotate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "launch.jsonl")
	for i := int64(1); i <= 4; i++ {
		if err := appendDiagnostic(path, LogExportEntry{Timestamp: i, Source: "server", Message: "Apply: begin"}, 1); err != nil {
			t.Fatal(err)
		}
	}
	// Read from disk with no memory history, as happens after restarting AME.
	got := readDiagnostics(path)
	if len(got) != 3 || got[0].Timestamp != 2 || got[2].Timestamp != 4 {
		t.Fatalf("unexpected retained history: %+v", got)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("{interrupted write")
	f.Close()
	if len(readDiagnostics(path)) != 3 {
		t.Fatal("damaged final record hid valid history")
	}
	if err := appendDiagnostic(path, LogExportEntry{Timestamp: 5, Message: "Session: restarted"}, diagnosticLimit); err != nil {
		t.Fatal(err)
	}
	got = readDiagnostics(path)
	if len(got) != 4 || got[3].Timestamp != 5 {
		t.Fatal("partial record swallowed the next session")
	}
}

func TestDiagnosticScopeAndRedaction(t *testing.T) {
	for _, msg := range []string{"Apply: begin", "[Resume] remote thread timed out", "Unstuck: attempting game termination", "Mod-tools: stderr: failure", "Preparation hook handoff incomplete: timeout", "Preparation complete, runtime handoff ready"} {
		if !launchDiagnostic(msg) {
			t.Errorf("lost diagnostic %q", msg)
		}
	}
	for _, msg := range []string{"Room key: private", "Room Party: own skin updated", "Skin: using cached skin", "Prefetched teammate skin: example"} {
		if launchDiagnostic(msg) {
			t.Errorf("retained noise %q", msg)
		}
	}
	t.Setenv("USERPROFILE", `C:\Users\private`)
	got := cleanDiagnostic(`Skin: failed https://example.test/?token=secret C:\Users\private\file`)
	if got != `Skin: failed <url> <user>\file` {
		t.Fatalf("unexpected redaction %q", got)
	}
}
