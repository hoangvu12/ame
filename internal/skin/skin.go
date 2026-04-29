package skin

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
)

const cacheMetadataFile = "cache.json"
const cacheValidationTTL = 12 * time.Hour

type cacheMetadata struct {
	Source       string    `json:"source"`
	Path         string    `json:"path"`
	ETag         string    `json:"etag"`
	LastModified string    `json:"lastModified"`
	Size         int64     `json:"size"`
	CheckedAt    time.Time `json:"checkedAt"`
}

type remoteSkinVersion struct {
	Path         string
	ETag         string
	LastModified string
	Size         int64
}

func skinDir(championID, skinID string) string {
	return filepath.Join(config.SkinsDir, championID, skinID)
}

func skinCacheMetadataPath(championID, skinID string) string {
	return filepath.Join(skinDir(championID, skinID), cacheMetadataFile)
}

func rseSkinPath(championID, skinID, baseSkinID string) string {
	if baseSkinID != "" && baseSkinID != "0" {
		return fmt.Sprintf("%s/%s/%s/%s.rse", championID, baseSkinID, skinID, skinID)
	}
	return fmt.Sprintf("%s/%s/%s.rse", championID, skinID, skinID)
}

func rseSkinURL(championID, skinID, baseSkinID string) string {
	return strings.TrimRight(rseSkinBaseURL, "/") + "/" + rseSkinPath(championID, skinID, baseSkinID)
}

func logSkin(msg string) {
	display.Log("Skin: " + msg)
}

func readCacheMetadata(championID, skinID string) (*cacheMetadata, error) {
	data, err := os.ReadFile(skinCacheMetadataPath(championID, skinID))
	if err != nil {
		return nil, err
	}

	var meta cacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func writeCacheMetadata(championID, skinID string, remote *remoteSkinVersion) {
	if !hasRemoteVersion(remote) {
		return
	}

	meta := cacheMetadata{
		Source:       "RSE",
		Path:         remote.Path,
		ETag:         remote.ETag,
		LastModified: remote.LastModified,
		Size:         remote.Size,
		CheckedAt:    time.Now().UTC(),
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(skinDir(championID, skinID), os.ModePerm); err != nil {
		return
	}
	_ = os.WriteFile(skinCacheMetadataPath(championID, skinID), data, 0644)
}

func fetchRSESkinVersion(championID, skinID, baseSkinID string) (*remoteSkinVersion, error) {
	if !rseAvailable() {
		return nil, fmt.Errorf("RSE source not configured")
	}

	path := rseSkinPath(championID, skinID, baseSkinID)
	logSkin(fmt.Sprintf("checking cache version for %s", path))
	req, err := http.NewRequest("HEAD", rseSkinURL(championID, skinID, baseSkinID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ame")

	resp, err := rseClient.Do(req)
	if err != nil {
		logSkin(fmt.Sprintf("version check failed for %s: %v", path, err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		logSkin(fmt.Sprintf("version check missing for %s (404)", path))
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		logSkin(fmt.Sprintf("version check failed for %s (status %d)", path, resp.StatusCode))
		return nil, fmt.Errorf("skin source returned status %d", resp.StatusCode)
	}
	logSkin(fmt.Sprintf("version check ok for %s", path))

	return &remoteSkinVersion{
		Path:         path,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Size:         resp.ContentLength,
	}, nil
}

func hasRemoteVersion(remote *remoteSkinVersion) bool {
	return remote != nil && (remote.ETag != "" || remote.LastModified != "" || remote.Size > 0)
}

func sameRemoteVersion(a, b *remoteSkinVersion) bool {
	if a == nil || b == nil || a.Path != b.Path {
		return false
	}
	if a.ETag != "" && b.ETag != "" {
		return a.ETag == b.ETag
	}
	if a.LastModified != "" && b.LastModified != "" && a.Size > 0 && b.Size > 0 {
		return a.LastModified == b.LastModified && a.Size == b.Size
	}
	return a.Size > 0 && b.Size > 0 && a.Size == b.Size
}

func removeCacheMetadata(championID, skinID string) {
	_ = os.Remove(skinCacheMetadataPath(championID, skinID))
}

// Download downloads a skin file.
func Download(championID, skinID, baseSkinID, championName, skinName, chromaName string) (string, error) {
	if !rseAvailable() {
		logSkin(fmt.Sprintf("source not configured (keyURL=%v baseURL=%v)", rseKeyURL != "", rseSkinBaseURL != ""))
		return "", fmt.Errorf("skin source not configured")
	}
	path, err := downloadRSE(championID, skinID, baseSkinID)
	if err == nil {
		return path, nil
	}
	logSkin(fmt.Sprintf("download failed for %s: %v", rseSkinPath(championID, skinID, baseSkinID), err))
	removeCacheMetadata(championID, skinID)
	return "", fmt.Errorf("skin not available for download")
}

// Extract extracts a zip/fantome file to destination directory
func Extract(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Prevent zip slip vulnerability
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// GetCachedPath returns the path to a cached skin file if it exists
func GetCachedPath(championID, skinID string) string {
	dir := skinDir(championID, skinID)

	// Check for zip first
	zipPath := filepath.Join(dir, fmt.Sprintf("%s.zip", skinID))
	if _, err := os.Stat(zipPath); err == nil {
		return zipPath
	}

	// Check for fantome
	fantomePath := filepath.Join(dir, fmt.Sprintf("%s.fantome", skinID))
	if _, err := os.Stat(fantomePath); err == nil {
		return fantomePath
	}

	return ""
}

// GetValidCachedPath returns a cached skin only when it is not known to be stale.
// Legacy caches without metadata stay usable when the configured skin source cannot be checked.
func GetValidCachedPath(championID, skinID, baseSkinID string) string {
	cachedPath := GetCachedPath(championID, skinID)
	if cachedPath == "" {
		return ""
	}
	if !rseAvailable() {
		logSkin(fmt.Sprintf("using cached %s without validation (source not configured)", rseSkinPath(championID, skinID, baseSkinID)))
		return cachedPath
	}

	meta, err := readCacheMetadata(championID, skinID)
	if err == nil && meta.Source == "RSE" && meta.Path == rseSkinPath(championID, skinID, baseSkinID) {
		if time.Since(meta.CheckedAt) < cacheValidationTTL {
			logSkin(fmt.Sprintf("using cached %s (recent metadata)", meta.Path))
			return cachedPath
		}

		remote, err := fetchRSESkinVersion(championID, skinID, baseSkinID)
		if err != nil || !hasRemoteVersion(remote) {
			logSkin(fmt.Sprintf("using cached %s (validation unavailable)", meta.Path))
			return cachedPath
		}
		cached := &remoteSkinVersion{Path: meta.Path, ETag: meta.ETag, LastModified: meta.LastModified, Size: meta.Size}
		if sameRemoteVersion(cached, remote) {
			writeCacheMetadata(championID, skinID, remote)
			logSkin(fmt.Sprintf("using cached %s (up to date)", meta.Path))
			return cachedPath
		}

		logSkin(fmt.Sprintf("cache stale for %s", meta.Path))
		removeCacheMetadata(championID, skinID)
		return ""
	}

	remote, err := fetchRSESkinVersion(championID, skinID, baseSkinID)
	if err != nil || !hasRemoteVersion(remote) {
		logSkin(fmt.Sprintf("using legacy cached %s (validation unavailable)", rseSkinPath(championID, skinID, baseSkinID)))
		return cachedPath
	}

	// A legacy cache may predate the current source file; redownload once to bootstrap metadata.
	logSkin(fmt.Sprintf("refreshing legacy cached %s to add metadata", remote.Path))
	removeCacheMetadata(championID, skinID)
	return ""
}
