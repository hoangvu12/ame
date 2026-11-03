package skin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LeagueSkins (github.com/Alban1911/LeagueSkins) publishes the Rose
// community's per-patch skin packages. It is the package source unless
// AME_SKIN_CDN_URL points this client at ame's own CDN.

// Overridden by tests.
var leagueSkinsBaseURL = "https://raw.githubusercontent.com/Alban1911/LeagueSkins/main/skins"

// A cached package is reused without asking GitHub for this long; after that
// one conditional request picks up packages republished for a patch or fix.
const leagueSkinsRecheck = 30 * time.Minute

// leagueSkinsMetadata is the cache record for a package fetched from LeagueSkins.
type leagueSkinsMetadata struct {
	Path     string `json:"path"`
	ETag     string `json:"etag"`
	Digest   string `json:"sha256"`
	Artifact string `json:"artifact"`
}

func useCDN() bool {
	return strings.TrimSpace(os.Getenv(cdnURLEnv)) != ""
}

// leagueSkinsPath mirrors the repository layout: chromas and forms live in a
// folder under their base skin. normalizeRequest has already run.
func leagueSkinsPath(r PackageRequest) string {
	dir := r.ChampionID + "/" + r.SkinID
	if r.BaseSkinID != "" {
		dir = r.ChampionID + "/" + r.BaseSkinID + "/" + r.SkinID
	}
	return dir + "/" + r.SkinID + ".fantome"
}

// leagueSkinsCached returns the cached package for r and its record, when the
// file is intact and was fetched from the path r maps to.
func leagueSkinsCached(r PackageRequest) (string, *cacheMetadata) {
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "leagueskins" || m.LeagueSkins == nil {
		return "", nil
	}
	ls := m.LeagueSkins
	if ls.Path != leagueSkinsPath(r) || ls.Artifact == "" || filepath.Base(ls.Artifact) != ls.Artifact {
		return "", nil
	}
	path := filepath.Join(skinDir(r.ChampionID, r.SkinID), ls.Artifact)
	data, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != ls.Digest {
		return "", nil
	}
	return path, m
}

// leagueSkinsFreshPath is the cached package when it was checked against
// LeagueSkins recently enough to use without a request.
func leagueSkinsFreshPath(r PackageRequest) string {
	path, m := leagueSkinsCached(r)
	if path == "" || time.Since(m.CheckedAt) >= leagueSkinsRecheck {
		return ""
	}
	return path
}

// leagueSkinsEnsurePackage returns the LeagueSkins package for a selection,
// revalidating an older cached copy and falling back to it while GitHub is
// unreachable.
func leagueSkinsEnsurePackage(ctx context.Context, r PackageRequest) (string, error) {
	cached, m := leagueSkinsCached(r)
	if cached != "" && time.Since(m.CheckedAt) < leagueSkinsRecheck {
		return cached, nil
	}
	etag := ""
	if cached != "" {
		etag = m.LeagueSkins.ETag
	}
	path, err := fetchLeagueSkinsPackage(ctx, r, etag)
	if err != nil {
		if cached != "" {
			logSkin("LeagueSkins check failed for " + r.SkinID + ", using cached package: " + err.Error())
			return cached, nil
		}
		return "", err
	}
	if path != "" {
		return path, nil
	}
	m.CheckedAt = time.Now().UTC()
	encoded, _ := json.MarshalIndent(m, "", "  ")
	if err := atomicWrite(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded); err != nil {
		logSkin("could not record LeagueSkins check: " + err.Error())
	}
	return cached, nil
}

// fetchLeagueSkinsPackage downloads and caches the package for r. It returns
// an empty path when etag still matches the published package.
func fetchLeagueSkinsPackage(ctx context.Context, r PackageRequest, etag string) (string, error) {
	rel := leagueSkinsPath(r)
	req, err := http.NewRequestWithContext(ctx, "GET", leagueSkinsBaseURL+"/"+rel, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ame")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := cdnClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotModified && etag != "":
		return "", nil
	case resp.StatusCode == http.StatusNotFound:
		return "", fmt.Errorf("LeagueSkins has no package for %s", rel)
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("LeagueSkins status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, cdnPackageLimit+1))
	if err != nil {
		return "", err
	}
	if len(data) > cdnPackageLimit {
		return "", fmt.Errorf("package exceeds %d bytes", cdnPackageLimit)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	path, artifact, err := storePackage(r, data, "leagueskins-"+digest+".fantome")
	if err != nil {
		return "", err
	}

	meta := cacheMetadata{
		Source:      "leagueskins",
		LeagueSkins: &leagueSkinsMetadata{Path: rel, ETag: resp.Header.Get("ETag"), Digest: digest, Artifact: artifact},
		CheckedAt:   time.Now().UTC(),
	}
	encoded, _ := json.MarshalIndent(meta, "", "  ")
	if err := atomicWrite(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded); err != nil {
		return "", err
	}
	return path, nil
}
