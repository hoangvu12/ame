package server

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticZIPPrivacyAndBounds(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "ame", "logs")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "launch.jsonl"), []byte(strings.Repeat("x", diagnosticFileCap)+"\npassword=secret token=private C:/Users/Alice/game https://user:pass@example.org\n"), 0600)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte("never bundle"), 0600)
	manager := filepath.Join(root, "dev.leaguetoolkit.manager", "logs")
	os.MkdirAll(manager, 0700)
	for _, name := range []string{"a.log", "b.log", "c.log", "d.log", "settings.json"} {
		os.WriteFile(filepath.Join(manager, name), []byte("Bearer abc123 /home/alice/game"), 0600)
	}
	data, err := buildDiagnosticZIP(root, "", "test-version")
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 5 {
		t.Fatalf("entries: %d", len(zr.File))
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, "settings") || strings.Contains(f.Name, "..") {
			t.Fatal(f.Name)
		}
		r, _ := f.Open()
		b, _ := io.ReadAll(r)
		r.Close()
		if len(b) > diagnosticFileCap {
			t.Fatal("unbounded entry")
		}
		for _, secret := range []string{"Alice", "alice", "abc123", "=secret", "=private", "user:pass"} {
			if bytes.Contains(b, []byte(secret)) {
				t.Fatalf("leaked %s: %s", secret, b)
			}
		}
	}
}

func TestDiagnosticZIPMissing(t *testing.T) {
	data, err := buildDiagnosticZIP("", "", "v1")
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(zr.File) != 1 {
		t.Fatalf("missing logs: %v", err)
	}
}

func TestDiagnosticDownloadSecurity(t *testing.T) {
	for _, tc := range []struct {
		method, origin, site string
		want                 int
	}{
		{"GET", "https://evil.example", "", 403}, {"GET", "", "cross-site", 403}, {"POST", "", "", 405}, {"GET", "https://riot", "", 200}, {"GET", "", "", 200},
	} {
		t.Setenv("LOCALAPPDATA", t.TempDir())
		t.Setenv("APPDATA", t.TempDir())
		req := httptest.NewRequest(tc.method, "http://localhost:18765/diagnostics.zip", nil)
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("Sec-Fetch-Site", tc.site)
		w := httptest.NewRecorder()
		handleDiagnosticsDownload(w, req)
		if w.Code != tc.want {
			t.Fatalf("%+v: %d", tc, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("CORS exposed")
		}
		if w.Code == 200 && w.Header().Get("Content-Type") != "application/zip" {
			t.Fatal("not ZIP")
		}
	}
}
