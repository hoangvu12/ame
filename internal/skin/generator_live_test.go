package skin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hoangvu12/ame/internal/game"
)

// TestLiveStalePackageFallback fetches a real published package, checks the
// patch it was built for against the installed game, and makes sure ame
// builds the package from the installed files when the published one is for
// another patch. Enabled only with AME_LIVE_GENERATOR=1.
func TestLiveStalePackageFallback(t *testing.T) {
	if os.Getenv("AME_LIVE_GENERATOR") != "1" {
		t.Skip("live generator test disabled; set AME_LIVE_GENERATOR=1")
	}
	testCache(t)
	gameDir := game.FindGameDir()
	if gameDir == "" {
		t.Skip("no League installation for the live test")
	}
	r := PackageRequest{"43", "43061", ""}
	published, err := leagueSkinsEnsurePackage(context.Background(), r)
	if err != nil {
		t.Fatalf("LeagueSkins did not serve %s: %v", r.SkinID, err)
	}
	pkg, installed, mismatch := patchMismatch(published, gameDir)
	t.Logf("published package: patch %s, installed patch %s", pkg, installed)
	if !mismatch {
		t.Skipf("the published package is for the installed patch %s", installed)
	}
	path, err := ensurePackage(context.Background(), r, gameDir)
	if err != nil {
		t.Fatalf("ensurePackage: %v", err)
	}
	if filepath.Base(path) == filepath.Base(published) {
		t.Fatalf("the package for another patch was used: %s", path)
	}
	if err := validatePackage(path); err != nil {
		t.Fatalf("generated package failed validation: %v", err)
	}
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "local" || m.Local == nil {
		t.Fatalf("cache metadata missing for the generated package: %+v %v", m, err)
	}
	t.Logf("built %s from the installed files (build %s)", path, m.Local.Build)
}
