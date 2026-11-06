package server

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/display"
)

const diagnosticFileCap = 512 * 1024

var diagnosticSecrets = regexp.MustCompile(`(?i)((?:authorization|password|passwd|access[_-]?token|refresh[_-]?token|token|api[_-]?key|secret)["']?\s*[:=]\s*["']?)[^\s,"'}]+`)
var diagnosticURLs = regexp.MustCompile(`(?i)(?:https?|wss?)://[^\s"<>]+`)
var diagnosticHomes = regexp.MustCompile(`(?i)(?:[a-z]:[\\/]+Users[\\/]+|/home/|/Users/)[^\\/\s"<>]+`)
var diagnosticBearer = regexp.MustCompile(`(?i)\b(?:Bearer|Basic)\s+[A-Za-z0-9+/_.=:-]+`)

func redactDiagnostic(data []byte) []byte {
	s := string(data)
	for _, key := range []string{"USERPROFILE", "HOME"} {
		if home := os.Getenv(key); home != "" {
			re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(home))
			s = re.ReplaceAllString(s, "<home>")
			escapedHome := strings.ReplaceAll(home, `\`, `\\`)
			escapedRE := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(escapedHome))
			s = escapedRE.ReplaceAllString(s, "<home>")
		}
	}
	s = diagnosticHomes.ReplaceAllString(s, "<home>")
	s = diagnosticURLs.ReplaceAllString(s, "<url>")
	s = diagnosticBearer.ReplaceAllString(s, "<credential>")
	return []byte(diagnosticSecrets.ReplaceAllString(s, "${1}<redacted>"))
}

type diagnosticSource struct {
	path, name string
	modified   time.Time
}

// Only known log locations are inspected; never recurse through settings or caches.
func diagnosticSources(local, roaming string) []diagnosticSource {
	var sources []diagnosticSource
	if local != "" {
		for i, suffix := range []string{"", ".1", ".2"} {
			sources = append(sources, diagnosticSource{path: filepath.Join(local, "ame", "logs", "launch.jsonl"+suffix), name: fmt.Sprintf("ame/launch-%d.jsonl", i)})
		}
	}
	var manager []diagnosticSource
	for _, root := range []string{local, roaming} {
		if root == "" {
			continue
		}
		dir := filepath.Join(root, "dev.leaguetoolkit.manager", "logs")
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".log") {
				continue
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			manager = append(manager, diagnosticSource{path: filepath.Join(dir, entry.Name()), modified: info.ModTime()})
		}
	}
	sort.SliceStable(manager, func(i, j int) bool { return manager[i].modified.After(manager[j].modified) })
	for i := 0; i < len(manager) && i < 3; i++ {
		manager[i].name = fmt.Sprintf("ltk/manager-%d.log", i)
		sources = append(sources, manager[i])
	}
	return sources
}

func buildDiagnosticZIP(local, roaming, version string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, data []byte) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	var included []string
	for _, source := range diagnosticSources(local, roaming) {
		info, err := os.Lstat(source.path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		f, err := os.Open(source.path)
		if err != nil {
			continue
		}
		start := info.Size() - diagnosticFileCap
		if start > 0 {
			_, err = f.Seek(start, io.SeekStart)
		}
		var data []byte
		if err == nil {
			data, err = io.ReadAll(io.LimitReader(f, diagnosticFileCap))
		}
		f.Close()
		if err != nil {
			continue
		}
		// Drop a partial leading record when taking the tail.
		if start > 0 {
			if n := bytes.IndexByte(data, '\n'); n >= 0 {
				data = data[n+1:]
			}
		}
		data = redactDiagnostic(data)
		if len(data) > diagnosticFileCap {
			data = data[:diagnosticFileCap]
		}
		if err := add(source.name, data); err != nil {
			return nil, err
		}
		included = append(included, source.name)
	}
	metadata := fmt.Sprintf("ame diagnostics\nVersion: %s\nExported UTC: %s\nIncluded: %s\nOnly available log tails (up to 512 KiB each, 6 files) are included. Missing/unreadable logs are skipped.\nNo settings, caches or packages. Redaction is best effort: review before sharing. No automatic upload.\n", version, time.Now().UTC().Format(time.RFC3339), strings.Join(included, ", "))
	if err := add("README.txt", redactDiagnostic([]byte(metadata))); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// handleDiagnosticsDownload exports only bounded log tails, never settings.
func handleDiagnosticsDownload(w http.ResponseWriter, r *http.Request) {
	// No CORS wrapper: websites must not read local diagnostics. Anchor downloads
	// carry no Origin; cross-site fetches are rejected by Fetch Metadata as well.
	origin := r.Header.Get("Origin")
	if (origin != "" && origin != "https://riot") || (r.Header.Get("Sec-Fetch-Site") == "cross-site" && !(r.Header.Get("Sec-Fetch-Mode") == "navigate" && r.Header.Get("Sec-Fetch-User") == "?1")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := buildDiagnosticZIP(os.Getenv("LOCALAPPDATA"), os.Getenv("APPDATA"), display.GetVersion())
	if err != nil {
		http.Error(w, "Unable to create diagnostics ZIP", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="ame-diagnostics.zip"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(data)
}
