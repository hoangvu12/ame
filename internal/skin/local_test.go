package skin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHelperInputsMatch(t *testing.T) {
	inputs := []inputFingerprint{
		{Path: `C:\Game\League of Legends.exe`},
		{Path: `C:\Game\DATA\FINAL\Champions\Karma.wad.client`},
	}
	if !helperInputsMatch([]string{`c:\game\data\final\champions\karma.wad.client`}, inputs) {
		t.Fatal("the helper's inputs did not match")
	}
	if helperInputsMatch([]string{`C:\Game\DATA\FINAL\Champions\Orianna.wad.client`}, inputs) {
		t.Fatal("a different input was accepted")
	}
	if helperInputsMatch([]string{}, inputs) {
		t.Fatal("missing inputs were accepted")
	}
}

// A published package for another patch is replaced by a package built from
// the installed files: the patcher DLL refuses the published one and leaves
// the match without skins.
func TestStalePublishedPackageIsReplaced(t *testing.T) {
	dir := t.TempDir()
	published := writePackage(t, dir, "Patch 16.22")

	oldInstalled := installedGamePatch
	oldPublished := leagueSkinsPackageFor
	oldLocal := localPackage
	installedGamePatch = func(string) string { return "16.23" }
	leagueSkinsPackageFor = func(context.Context, PackageRequest) (string, error) {
		return published, nil
	}
	localPackage = func(context.Context, PackageRequest, string) (string, error) {
		return filepath.Join(dir, "local.fantome"), nil
	}
	defer func() {
		installedGamePatch = oldInstalled
		leagueSkinsPackageFor = oldPublished
		localPackage = oldLocal
	}()

	path, err := ensurePackage(context.Background(), PackageRequest{ChampionID: "43", SkinID: "43061"}, dir)
	if err != nil {
		t.Fatalf("ensurePackage: %v", err)
	}
	if filepath.Base(path) != "local.fantome" {
		t.Fatalf("ensurePackage=%q, want the package built from the installed files", path)
	}
}

// A published package for this patch is used as it is.
func TestFreshPublishedPackageIsUsed(t *testing.T) {
	dir := t.TempDir()
	published := writePackage(t, dir, "Patch 16.23")

	oldInstalled := installedGamePatch
	oldPublished := leagueSkinsPackageFor
	oldLocal := localPackage
	installedGamePatch = func(string) string { return "16.23" }
	leagueSkinsPackageFor = func(context.Context, PackageRequest) (string, error) {
		return published, nil
	}
	localPackage = func(context.Context, PackageRequest, string) (string, error) {
		t.Fatal("a package was built although the published one fits")
		return "", nil
	}
	defer func() {
		installedGamePatch = oldInstalled
		leagueSkinsPackageFor = oldPublished
		localPackage = oldLocal
	}()

	path, err := ensurePackage(context.Background(), PackageRequest{ChampionID: "43", SkinID: "43061"}, dir)
	if err != nil {
		t.Fatalf("ensurePackage: %v", err)
	}
	if path != published {
		t.Fatalf("ensurePackage=%q, want the published package", path)
	}
}

// When the package cannot be built from the installed files, the published
// one is used and the mismatch is remembered for the user.
func TestStalePublishedPackageIsReportedWhenItCannotBeReplaced(t *testing.T) {
	dir := t.TempDir()
	published := writePackage(t, dir, "Patch 16.22")

	oldInstalled := installedGamePatch
	oldPublished := leagueSkinsPackageFor
	oldLocal := localPackage
	installedGamePatch = func(string) string { return "16.23" }
	leagueSkinsPackageFor = func(context.Context, PackageRequest) (string, error) {
		return published, nil
	}
	localPackage = func(context.Context, PackageRequest, string) (string, error) {
		return "", errors.New("no bundle")
	}
	defer func() {
		installedGamePatch = oldInstalled
		leagueSkinsPackageFor = oldPublished
		localPackage = oldLocal
		setStalePackageNotice("")
	}()

	path, err := ensurePackage(context.Background(), PackageRequest{ChampionID: "43", SkinID: "43061"}, dir)
	if err != nil {
		t.Fatalf("ensurePackage: %v", err)
	}
	if path != published {
		t.Fatalf("ensurePackage=%q, want the published package", path)
	}
	if StalePackageNotice() == "" {
		t.Fatal("the patch mismatch was not reported")
	}
}

// A package generated earlier is reused only while the installed files it
// was built from are unchanged.
func TestLocalCachedFollowsTheInstallation(t *testing.T) {
	testCache(t)
	cache := skinDir("43", "43061")
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatalf("cache dir: %v", err)
	}
	artifact := "local-abc.fantome"
	if err := os.WriteFile(filepath.Join(cache, artifact), []byte("package"), 0600); err != nil {
		t.Fatalf("write package: %v", err)
	}
	inputs := []inputFingerprint{{Path: "Karma.wad.client", Size: 10}}
	content, _ := os.ReadFile(filepath.Join(cache, artifact))
	meta := &localMetadata{
		Inputs:   inputs,
		Artifact: artifact,
		Digest:   fmt.Sprintf("%x", sha256.Sum256(content)),
	}
	encoded, _ := json.MarshalIndent(cacheMetadata{Source: "local", Local: meta}, "", "  ")
	if err := os.WriteFile(skinCacheMetadataPath("43", "43061"), encoded, 0600); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	same := &localMetadata{Inputs: inputs, Artifact: artifact, Digest: meta.Digest}

	if localCached(PackageRequest{ChampionID: "43", SkinID: "43061"}, same) == "" {
		t.Fatal("a package built from the same files was not reused")
	}
	changed := &localMetadata{Inputs: []inputFingerprint{{Path: "Karma.wad.client", Size: 11}}, Artifact: artifact, Digest: meta.Digest}
	if localCached(PackageRequest{ChampionID: "43", SkinID: "43061"}, changed) != "" {
		t.Fatal("a package built from other files was reused")
	}
}
