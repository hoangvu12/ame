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
	"sync"
	"time"
)

// DefaultCDNURL and DefaultCDNToken are injected at release time via
// -ldflags="-X ...DefaultCDNURL=... -X ...DefaultCDNToken=...". Source builds
// leave them empty, which disables remote packages entirely: those builds
// behave exactly like local-generation-only ame.
var DefaultCDNURL string
var DefaultCDNToken string

const (
	cdnURLEnv   = "AME_SKIN_CDN_URL"
	cdnTokenEnv = "AME_SKIN_CDN_TOKEN"

	cdnCatalogTTL     = 5 * time.Minute
	cdnCatalogTimeout = 15 * time.Second
	cdnErrorBackoff   = time.Minute
	cdnFetchTimeout   = 90 * time.Second

	// Hard cap regardless of the catalog's declared size.
	cdnPackageLimit = 256 << 20
)

var cdnClient = &http.Client{Timeout: cdnFetchTimeout}

// remoteMetadata is the cache record for a package fetched from the CDN.
type remoteMetadata struct {
	Build    string `json:"gameBuild"`
	Digest   string `json:"sha256"`
	Key      string `json:"key"`
	Artifact string `json:"artifact"`
}

// cdnPackage is one entry of the CDN catalog.
type cdnPackage struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// cdnCatalog is the published package index.
type cdnCatalog struct {
	Version     int                   `json:"version"`
	Generator   string                `json:"generator"`
	Build       string                `json:"build"`
	GeneratedAt string                `json:"generatedAt"`
	Packages    map[string]cdnPackage `json:"packages"`
}

type cdnConfig struct {
	BaseURL string
	Token   string
}

func activeCDN() (*cdnConfig, error) {
	base := strings.TrimSpace(os.Getenv(cdnURLEnv))
	token := os.Getenv(cdnTokenEnv)
	if base == "" {
		base = strings.TrimSpace(DefaultCDNURL)
		token = coalesceEnv(token, DefaultCDNToken)
	}
	if base == "" || token == "" {
		return nil, fmt.Errorf("cdn not configured")
	}
	return &cdnConfig{BaseURL: strings.TrimSuffix(base, "/"), Token: token}, nil
}

func coalesceEnv(env, fallback string) string {
	if strings.TrimSpace(env) != "" {
		return env
	}
	return fallback
}

// packageKey matches the worker route: skinID for plain skins and
// skinID-baseSkinId for chromas and forms. normalizeRequest has already run.
// The separator must stay Windows-filename-safe (no colons).
func packageKey(r PackageRequest) string {
	base := strings.TrimSpace(r.BaseSkinID)
	if base == "" || base == "0" || base == r.SkinID {
		return r.SkinID
	}
	return r.SkinID + "-" + base
}

var cdnState = struct {
	sync.Mutex
	catalog    *cdnCatalog
	fetchedAt  time.Time
	errorUntil time.Time
	lastError  error
}{}

func fetchCDNCatalog(ctx context.Context, cfg *cdnConfig) (*cdnCatalog, error) {
	cdnState.Lock()
	defer cdnState.Unlock()
	now := time.Now()
	if cdnState.catalog != nil && now.Sub(cdnState.fetchedAt) < cdnCatalogTTL {
		return cdnState.catalog, nil
	}
	if now.Before(cdnState.errorUntil) {
		return nil, cdnState.lastError
	}

	cctx, cancel := context.WithTimeout(ctx, cdnCatalogTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, "GET", cfg.BaseURL+"/catalog", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("User-Agent", "ame")
	resp, err := cdnClient.Do(req)
	if err != nil {
		return cdnFail(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cdnFail(fmt.Errorf("catalog status %d", resp.StatusCode))
	}
	var catalog cdnCatalog
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&catalog); err != nil {
		return cdnFail(fmt.Errorf("catalog decode: %w", err))
	}
	if catalog.Version != 1 || len(catalog.Packages) == 0 {
		return cdnFail(fmt.Errorf("catalog version %d with %d packages", catalog.Version, len(catalog.Packages)))
	}
	cdnState.catalog = &catalog
	cdnState.fetchedAt = now
	cdnState.errorUntil = time.Time{}
	return &catalog, nil
}

func cdnFail(err error) (*cdnCatalog, error) {
	cdnState.errorUntil = time.Now().Add(cdnErrorBackoff)
	cdnState.lastError = err
	return nil, err
}

// remoteEnsurePackage tries the pre-generated CDN package for a selection.
// It returns "" (never an error) so callers fall through to local generation.
func remoteEnsurePackage(ctx context.Context, r PackageRequest, gameDir string) string {
	cfg, err := activeCDN()
	if err != nil {
		return ""
	}
	catalog, err := fetchCDNCatalog(ctx, cfg)
	if err != nil {
		logSkin("cdn catalog unavailable: " + err.Error())
		return ""
	}
	if catalog.Generator != "" && catalog.Generator != generatorVersion {
		return ""
	}
	build := gameBuild(gameDir)
	if !catalogBuildMatches(catalog.Build, build) {
		return ""
	}
	entry, ok := catalog.Packages[packageKey(r)]
	if !ok {
		return ""
	}
	if path := remoteCached(r, catalog.Build, entry); path != "" {
		return path
	}
	path, err := fetchRemotePackage(ctx, cfg, r, entry, catalog.Build)
	if err != nil {
		logSkin("cdn package fetch failed for " + r.SkinID + ": " + err.Error())
		return ""
	}
	return path
}

// remoteCached validates an existing cached CDN package without network
// access. The recorded game build must match the catalog build.
func remoteCached(r PackageRequest, build string, entry cdnPackage) string {
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "remote" || m.Remote == nil {
		return ""
	}
	if build != "" && m.Remote.Build != build {
		return ""
	}
	if m.Remote.Digest != entry.SHA256 {
		return ""
	}
	if filepath.Base(m.Remote.Artifact) != m.Remote.Artifact || m.Remote.Artifact == "" {
		return ""
	}
	path := filepath.Join(skinDir(r.ChampionID, r.SkinID), m.Remote.Artifact)
	data, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != m.Remote.Digest {
		return ""
	}
	return path
}

func fetchRemotePackage(ctx context.Context, cfg *cdnConfig, r PackageRequest, entry cdnPackage, build string) (string, error) {
	key := packageKey(r)
	if len(entry.SHA256) < 12 {
		return "", fmt.Errorf("catalog digest malformed")
	}
	url := fmt.Sprintf("%s/p/%s?v=%s", cfg.BaseURL, key, entry.SHA256[:12])
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("User-Agent", "ame")
	resp, err := cdnClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	limit := entry.Size + (1 << 20)
	if limit <= 0 || limit > cdnPackageLimit {
		limit = cdnPackageLimit
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return "", err
	}
	if entry.Size > 0 && int64(len(data)) != entry.Size {
		return "", fmt.Errorf("size mismatch: %d != %d", len(data), entry.Size)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != strings.ToLower(entry.SHA256) {
		return "", fmt.Errorf("digest mismatch")
	}

	cache := skinDir(r.ChampionID, r.SkinID)
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(cache, ".remote-*.fantome")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := validatePackage(tmpPath); err != nil {
		return "", err
	}
	artifact := "remote-" + strings.ToLower(entry.SHA256) + ".fantome"
	path := filepath.Join(cache, artifact)
	if err := os.Rename(tmpPath, path); err != nil {
		return "", err
	}

	meta := cacheMetadata{
		Source:    "remote",
		Remote:    &remoteMetadata{Build: build, Digest: strings.ToLower(entry.SHA256), Key: key, Artifact: artifact},
		CheckedAt: time.Now().UTC(),
	}
	encoded, _ := json.MarshalIndent(meta, "", "  ")
	if err := atomicWrite(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded); err != nil {
		return "", err
	}
	return path, nil
}

// remoteValidCachedPath is the offline validation used by GetValidCachedPath:
// the cached CDN package must match the currently installed game build.
func remoteValidCachedPath(r PackageRequest, installedBuildVersion string) string {
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "remote" || m.Remote == nil {
		return ""
	}
	if installedBuildVersion != "unknown" && m.Remote.Build != "" && m.Remote.Build != installedBuildVersion {
		return ""
	}
	if filepath.Base(m.Remote.Artifact) != m.Remote.Artifact || m.Remote.Artifact == "" {
		return ""
	}
	path := filepath.Join(skinDir(r.ChampionID, r.SkinID), m.Remote.Artifact)
	data, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != m.Remote.Digest {
		return ""
	}
	return path
}

// catalogBuildMatches reports whether a catalog built against a game build
// can be used on the installed one. Catalogs without build information and
// clients that cannot determine their build are always accepted.
func catalogBuildMatches(catalogBuild, installed string) bool {
	if catalogBuild == "" || installed == "unknown" {
		return true
	}
	return catalogBuild == installed
}

func gameBuild(gameDir string) string {
	return installedBuild(filepath.Join(gameDir, "League of Legends.exe"))
}
