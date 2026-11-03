package skin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/hoangvu12/ame/internal/game"
)

// TestLiveLeagueSkinsSmoke fetches one real skin and one real chroma from
// LeagueSkins. Enabled only with AME_LIVE_LEAGUESKINS=1.
func TestLiveLeagueSkinsSmoke(t *testing.T) {
	if os.Getenv("AME_LIVE_LEAGUESKINS") != "1" {
		t.Skip("live LeagueSkins smoke disabled; set AME_LIVE_LEAGUESKINS=1")
	}
	testCache(t)
	for _, r := range []PackageRequest{{"103", "103001", ""}, {"103", "103052", "103001"}, {"99", "99991", "99007"}} {
		path, err := leagueSkinsEnsurePackage(context.Background(), r)
		if err != nil {
			t.Fatalf("LeagueSkins did not serve %s: %v", r.SkinID, err)
		}
		if err := validatePackage(path); err != nil {
			t.Fatalf("live package failed validation: %v", err)
		}
		t.Logf("fetched %s (key %s)", path, ArtifactKey(r.SkinID))
	}
}

// TestLiveCDNSmoke fetches one real package through the configured CDN.
// Enabled only with AME_LIVE_CDN=1 plus AME_SKIN_CDN_URL/AME_SKIN_CDN_TOKEN;
// skipped otherwise so normal test runs stay offline.
func TestLiveCDNSmoke(t *testing.T) {
	if os.Getenv("AME_LIVE_CDN") != "1" {
		t.Skip("live CDN smoke disabled; set AME_LIVE_CDN=1")
	}
	if os.Getenv(cdnURLEnv) == "" {
		t.Skip("no CDN configured")
	}
	testCache(t)
	resetCDNState()
	defer resetCDNState()

	r := PackageRequest{"103", "103001", ""}
	gameDir := game.FindGameDir()
	if gameDir == "" {
		t.Skip("no League installation for live smoke")
	}

	path, err := remoteEnsurePackage(context.Background(), r, gameDir)
	if err != nil {
		t.Fatalf("live CDN did not serve the package: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := readCacheMetadata("103", "103001")
	if err != nil || m.Source != "remote" || m.Remote == nil {
		t.Fatalf("cache metadata missing for remote package: %+v %v", m, err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != m.Remote.Digest {
		t.Fatalf("cached digest mismatch: %s != %s", got, m.Remote.Digest)
	}
	if err := validatePackage(path); err != nil {
		t.Fatalf("live package failed validation: %v", err)
	}
	t.Logf("fetched %s (%d bytes, build %s)", path, len(data), m.Remote.Build)
}
