package skin

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLeagueSkinsPath(t *testing.T) {
	cases := []struct {
		r    PackageRequest
		want string
	}{
		{PackageRequest{"103", "103001", ""}, "103/103001/103001.fantome"},
		{PackageRequest{"103", "103052", "103001"}, "103/103001/103052/103052.fantome"},
		{PackageRequest{"99", "99991", "99007"}, "99/99007/99991/99991.fantome"},
	}
	for _, c := range cases {
		if got := leagueSkinsPath(normalizeRequest(c.r)); got != c.want {
			t.Fatalf("leagueSkinsPath(%+v) = %q, want %q", c.r, got, c.want)
		}
	}
}

// serveLeagueSkins serves pkg at the chroma's repository path with an ETag
// and counts full and conditional requests.
func serveLeagueSkins(t *testing.T, pkg *[]byte, full, notModified *int32) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/103/103001/103052/103052.fantome" {
			http.NotFound(w, r)
			return
		}
		etag := fmt.Sprintf(`"%x"`, sha256.Sum256(*pkg))
		if r.Header.Get("If-None-Match") == etag {
			atomic.AddInt32(notModified, 1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		atomic.AddInt32(full, 1)
		w.Header().Set("ETag", etag)
		w.Write(*pkg)
	}))
	t.Cleanup(server.Close)
	old := leagueSkinsBaseURL
	leagueSkinsBaseURL = server.URL
	t.Cleanup(func() { leagueSkinsBaseURL = old })
	t.Setenv(cdnURLEnv, "")
}

func republishedPackageBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("META/info.json")
	w.Write([]byte(`{"name":"republished"}`))
	w2, _ := zw.Create("WAD/Ahri.wad.client")
	w2.Write(bytes.Repeat([]byte{8}, 64))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// expireLeagueSkinsCheck makes the cached record old enough to revalidate.
func expireLeagueSkinsCheck(t *testing.T, r PackageRequest) {
	t.Helper()
	_, m := leagueSkinsCached(r)
	if m == nil {
		t.Fatal("no cached LeagueSkins package")
	}
	m.CheckedAt = time.Now().Add(-2 * leagueSkinsRecheck)
	data, _ := json.Marshal(m)
	if err := os.WriteFile(skinCacheMetadataPath(r.ChampionID, r.SkinID), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLeagueSkinsFetchCachesAndRevalidates(t *testing.T) {
	testCache(t)
	pkg := testPackageBytes(t)
	var full, notModified int32
	serveLeagueSkins(t, &pkg, &full, &notModified)
	r := PackageRequest{"103", "103052", "103001"}

	path, err := EnsurePackage(context.Background(), r, "")
	if err != nil || path == "" {
		t.Fatalf("EnsurePackage failed: %v", err)
	}
	if got := GetValidCachedPath("103", "103052", "103001"); got != path {
		t.Fatalf("fresh cached path = %q, want %q", got, path)
	}
	if again, err := EnsurePackage(context.Background(), r, ""); err != nil || again != path || full != 1 || notModified != 0 {
		t.Fatalf("fresh cache refetched: %q %v full=%d 304=%d", again, err, full, notModified)
	}

	expireLeagueSkinsCheck(t, r)
	if got := GetValidCachedPath("103", "103052", "103001"); got != "" {
		t.Fatal("stale cache served without revalidation")
	}
	if again, err := EnsurePackage(context.Background(), r, ""); err != nil || again != path || full != 1 || notModified != 1 {
		t.Fatalf("unchanged package not revalidated: %q %v full=%d 304=%d", again, err, full, notModified)
	}
	if GetValidCachedPath("103", "103052", "103001") != path {
		t.Fatal("revalidation did not refresh the check time")
	}

	// A republished package replaces the cached copy and its overlay key.
	key := ArtifactKey("103052")
	expireLeagueSkinsCheck(t, r)
	pkg = republishedPackageBytes(t)
	updated, err := EnsurePackage(context.Background(), r, "")
	if err != nil || updated == path || full != 2 {
		t.Fatalf("republished package not fetched: %q %v full=%d", updated, err, full)
	}
	if ArtifactKey("103052") == key || !strings.HasPrefix(ArtifactKey("103052"), "103052@") {
		t.Fatalf("artifact key did not follow package: %s -> %s", key, ArtifactKey("103052"))
	}
}

func TestLeagueSkinsUsesCacheWhenUnreachable(t *testing.T) {
	testCache(t)
	pkg := testPackageBytes(t)
	var full, notModified int32
	serveLeagueSkins(t, &pkg, &full, &notModified)
	r := PackageRequest{"103", "103052", "103001"}
	path, err := EnsurePackage(context.Background(), r, "")
	if err != nil {
		t.Fatal(err)
	}

	expireLeagueSkinsCheck(t, r)
	leagueSkinsBaseURL = "http://127.0.0.1:1"
	if got, err := EnsurePackage(context.Background(), r, ""); err != nil || got != path {
		t.Fatalf("offline fallback = %q, %v", got, err)
	}
}

func TestLeagueSkinsMissingPackage(t *testing.T) {
	testCache(t)
	pkg := testPackageBytes(t)
	var full, notModified int32
	serveLeagueSkins(t, &pkg, &full, &notModified)

	path, err := EnsurePackage(context.Background(), PackageRequest{"103", "103001", ""}, "")
	if err == nil || path != "" || !strings.Contains(err.Error(), "no package for 103/103001/103001.fantome") {
		t.Fatalf("missing package = %q, %v", path, err)
	}
	entries, _ := os.ReadDir(skinDir("103", "103001"))
	if len(entries) != 0 {
		t.Fatalf("failed acquisition left cache entries: %v", entries)
	}
}

func TestLeagueSkinsRejectsInvalidPackage(t *testing.T) {
	testCache(t)
	pkg := []byte("not a zip")
	var full, notModified int32
	serveLeagueSkins(t, &pkg, &full, &notModified)

	if path, err := EnsurePackage(context.Background(), PackageRequest{"103", "103052", "103001"}, ""); err == nil || path != "" {
		t.Fatalf("invalid package accepted: %q", path)
	}
}
