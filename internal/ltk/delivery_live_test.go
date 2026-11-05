package ltk

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLiveManagerFetch delivers the newest LTK Manager runtime into the
// shared manager cache: ltk-manager.exe and its bundled host, extracted from
// the release installer with the pure-Go NSIS reader. The second call must
// hit the release marker and download nothing. Enabled with
// AME_LIVE_DELIVERY=1.
func TestLiveManagerFetch(t *testing.T) {
	if os.Getenv("AME_LIVE_DELIVERY") != "1" {
		t.Skip("live delivery disabled; set AME_LIVE_DELIVERY=1")
	}
	fetchStub := managerFetch
	managerFetch = ManagerFetch
	defer func() { managerFetch = fetchStub }()

	exe, err := ManagerFetch()
	if err != nil {
		t.Fatalf("ManagerFetch: %v", err)
	}
	st, err := os.Stat(exe)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() < 1024*1024 {
		t.Fatalf("delivered manager is only %d bytes", st.Size())
	}
	host, err := os.Stat(filepath.Join(managerDir, managerHostName))
	if err != nil {
		t.Fatalf("delivered host missing: %v", err)
	}
	t.Logf("manager=%s (%d bytes) host=%d bytes", exe, st.Size(), host.Size())

	start := time.Now()
	if _, err := ManagerFetch(); err != nil {
		t.Fatalf("second ManagerFetch: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("second fetch took %s; the release is being re-downloaded every start", elapsed)
	}
	t.Logf("second fetch took %s (marker cache hit)", time.Since(start))
}
