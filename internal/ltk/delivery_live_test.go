package ltk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLiveDeliveryFetch downloads the current patcher DLL and host from the
// real delivery service. Enabled only with AME_LIVE_DELIVERY=1.
func TestLiveDeliveryFetch(t *testing.T) {
	if os.Getenv("AME_LIVE_DELIVERY") != "1" {
		t.Skip("live delivery disabled; set AME_LIVE_DELIVERY=1")
	}
	dir := t.TempDir()
	dll, version, err := DeliveryFetch(dir)
	if err != nil {
		t.Fatalf("DeliveryFetch: %v", err)
	}
	st, err := os.Stat(dll)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() == 0 {
		t.Fatal("delivered DLL is empty")
	}
	t.Logf("dll=%s version=%s size=%d", dll, version, st.Size())

	host, err := os.Stat(filepath.Join(dir, "ltk-manager.exe"))
	if err != nil {
		t.Fatalf("delivered host missing: %v", err)
	}
	t.Logf("host size=%d", host.Size())
	marker, err := os.ReadFile(filepath.Join(dir, hostReleaseFile))
	if err != nil || strings.TrimSpace(string(marker)) == "" {
		t.Fatalf("host release marker missing: %v", err)
	}
	t.Logf("host release=%s", strings.TrimSpace(string(marker)))

	// A second fetch must not re-download the host installer.
	start := time.Now()
	if _, _, err := DeliveryFetch(dir); err != nil {
		t.Fatalf("second DeliveryFetch: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("second fetch took %s; host is being re-downloaded every start", elapsed)
	}
	t.Logf("second fetch took %s", time.Since(start))
}