package display

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const diagnosticLimit = 512 * 1024

var privateURL = regexp.MustCompile(`(?i)(?:https?|wss?)://\S+`)

// Keep launch diagnostics only; selection, cache and room chatter obscures hangs.
func launchDiagnostic(msg string) bool {
	if strings.HasPrefix(msg, "Preparation ") || strings.HasPrefix(msg, "Process suspend unavailable") {
		return true
	}
	for _, prefix := range []string{"Session:", "Apply:", "Prefetch:", "[Suspend]", "[Resume]", "Unstuck:", "Mod-tools:", "LTK:", "LTK build:", "LTK host", "LTK output:", "Game detected", "Game suspended", "Game released", "Apply finished", "Suspend ", "Suspender ", "Safety timeout", "Failed to create suspender", "Failed to suspend game", "Skin ready (prebuilt", "Client connected", "Client disconnected"} {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return strings.HasPrefix(msg, "! ") || (strings.HasPrefix(msg, "Skin:") && strings.Contains(msg, "failed"))
}

func cleanDiagnostic(msg string) string {
	msg = privateURL.ReplaceAllString(msg, "<url>")
	if home := os.Getenv("USERPROFILE"); home != "" {
		msg = strings.ReplaceAll(msg, home, "<user>")
	}
	if len(msg) > 4096 {
		msg = msg[:4096] + " [truncated]"
	}
	return msg
}

// Called under display.mu. Append rather than truncate so ordinary restarts
// preserve evidence; two rotated files bound disk usage to roughly 1.5 MiB.
func appendDiagnostic(path string, entry LogExportEntry, limit int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if info, err := os.Stat(path); err == nil && info.Size() >= limit {
		if err := os.Remove(path + ".2"); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(path+".1", path+".2"); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(path, path+".1"); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	// A separator also isolates a partial final record left by abrupt shutdown.
	_, err = f.WriteString("\n")
	if err == nil {
		err = json.NewEncoder(f).Encode(entry)
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func readDiagnostics(path string) []LogExportEntry {
	var entries []LogExportEntry
	for _, suffix := range []string{".2", ".1", ""} {
		f, err := os.Open(path + suffix)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var entry LogExportEntry
			if json.Unmarshal(scanner.Bytes(), &entry) == nil {
				entries = append(entries, entry)
			}
		}
		f.Close()
	}
	return entries
}

func diagnosticWriteError(err error) string {
	return cleanDiagnostic(fmt.Sprintf("Session: diagnostic file write failed: %v", err))
}
