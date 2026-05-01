package updater

import (
	"debug/pe"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

const (
	GITHUB_REPO    = "hoangvu12/ame"
	GITHUB_API_URL = "https://api.github.com/repos/" + GITHUB_REPO + "/releases/latest"
	minExeSize     = 1024 * 1024
)

var (
	VERSION_FILE     = filepath.Join(config.AmeDir, "version.txt")
	UPDATE_FILE      = filepath.Join(config.AmeDir, "ame_update.exe")
	UPDATE_TEMP_FILE = filepath.Join(config.AmeDir, "ame_update.exe.tmp")
	CORE_FILE        = filepath.Join(config.AmeDir, "ame_core.exe")
)

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateResult contains the result of an update check
type UpdateResult struct {
	UpdateAvailable bool
	CurrentVersion  string
	LatestVersion   string
	DownloadURL     string
	Downloaded      bool
}

// GetSavedVersion reads the version from the version file
func GetSavedVersion() string {
	data, err := os.ReadFile(VERSION_FILE)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SaveVersion writes the version to the version file
func SaveVersion(version string) error {
	os.MkdirAll(config.AmeDir, os.ModePerm)
	return os.WriteFile(VERSION_FILE, []byte(version), 0644)
}

// fetchLatestRelease gets the latest release info from GitHub
func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", GITHUB_API_URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ame-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// getExeDownloadURL finds the ame.exe download URL from release assets
func getExeDownloadURL(release *GitHubRelease) string {
	for _, asset := range release.Assets {
		if asset.Name == "ame.exe" {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

// downloadUpdate downloads the new version to the update path
func downloadUpdate(url string) error {
	os.MkdirAll(config.AmeDir, os.ModePerm)
	os.Remove(UPDATE_TEMP_FILE)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(UPDATE_TEMP_FILE)
	if err != nil {
		return err
	}

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		os.Remove(UPDATE_TEMP_FILE)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(UPDATE_TEMP_FILE)
		return err
	}

	if written < minExeSize {
		os.Remove(UPDATE_TEMP_FILE)
		return fmt.Errorf("download too small (%d bytes), likely failed", written)
	}
	if resp.ContentLength > 0 && written != resp.ContentLength {
		os.Remove(UPDATE_TEMP_FILE)
		return fmt.Errorf("download incomplete (%d/%d bytes)", written, resp.ContentLength)
	}
	if !verifyExeFile(UPDATE_TEMP_FILE) {
		os.Remove(UPDATE_TEMP_FILE)
		return fmt.Errorf("downloaded update is not a valid executable")
	}

	os.Remove(UPDATE_FILE)
	return os.Rename(UPDATE_TEMP_FILE, UPDATE_FILE)
}

// compareVersions returns true if latest is newer than current
// Versions are expected in format "v1.2.3"
func compareVersions(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// If current is empty, any version is newer
	if current == "" {
		return true
	}

	// Simple string comparison works for semantic versioning
	// For more robust comparison, we could parse major.minor.patch
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Pad shorter version with zeros
	for len(currentParts) < 3 {
		currentParts = append(currentParts, "0")
	}
	for len(latestParts) < 3 {
		latestParts = append(latestParts, "0")
	}

	for i := 0; i < 3; i++ {
		var c, l int
		fmt.Sscanf(currentParts[i], "%d", &c)
		fmt.Sscanf(latestParts[i], "%d", &l)

		if l > c {
			return true
		} else if c > l {
			return false
		}
	}

	return false
}

// CheckForUpdates checks if a new version is available and downloads it
func CheckForUpdates(currentVersion string) (*UpdateResult, error) {
	result := &UpdateResult{
		CurrentVersion: currentVersion,
	}

	// Fetch latest release from GitHub
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	result.LatestVersion = release.TagName
	result.DownloadURL = getExeDownloadURL(release)

	// Compare versions
	if !compareVersions(currentVersion, release.TagName) {
		result.UpdateAvailable = false
		return result, nil
	}

	result.UpdateAvailable = true

	// Download the update
	if result.DownloadURL != "" {
		if err := downloadUpdate(result.DownloadURL); err != nil {
			return result, fmt.Errorf("failed to download update: %w", err)
		}
		result.Downloaded = true
	}

	return result, nil
}

// NeedsPluginReinstall checks if the saved version differs from current
// This indicates an update was applied and plugins should be reinstalled
func NeedsPluginReinstall(currentVersion string) bool {
	savedVersion := GetSavedVersion()
	// If no saved version, this is first run - don't force reinstall
	if savedVersion == "" {
		return false
	}
	// If versions differ, an update was applied
	return savedVersion != currentVersion
}

// CleanupUpdateFile removes the downloaded update file if it exists
func CleanupUpdateFile() {
	os.Remove(UPDATE_FILE)
	os.Remove(UPDATE_TEMP_FILE)
}

// VerifyUpdateFile checks if the update file exists and is valid
func VerifyUpdateFile() bool {
	return verifyExeFile(UPDATE_FILE)
}

func verifyExeFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.Size() < minExeSize {
		return false
	}
	f, err := pe.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// ApplyPendingUpdate replaces ame_core.exe with ame_update.exe if a pending update exists.
// Called by the launcher before starting core.
func ApplyPendingUpdate() bool {
	if !VerifyUpdateFile() {
		return false
	}

	// Remove old core — retry because the previous core process may not have
	// fully released its file lock yet (Windows keeps exe files locked briefly
	// after process exit).
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		err := os.Remove(CORE_FILE)
		if err == nil || os.IsNotExist(err) {
			break
		}
		if i < maxRetries-1 {
			fmt.Printf("  Waiting for old process to release... (%d/%d)\n", i+1, maxRetries)
			time.Sleep(500 * time.Millisecond)
		} else {
			fmt.Printf("  ! Could not remove old core after %d attempts: %v\n", maxRetries, err)
			return false
		}
	}

	// Move update → core
	if err := os.Rename(UPDATE_FILE, CORE_FILE); err != nil {
		// Rename failed (cross-device?), try copy+delete
		if copyErr := copyFile(UPDATE_FILE, CORE_FILE); copyErr != nil {
			fmt.Printf("  ! Failed to apply update: %v\n", copyErr)
			return false
		}
		os.Remove(UPDATE_FILE)
	}

	return true
}

// BootstrapCore copies the current executable to ame_core.exe if it doesn't exist.
func BootstrapCore(srcExe string) error {
	if verifyExeFile(CORE_FILE) {
		return nil // already exists
	}
	os.Remove(CORE_FILE)
	os.MkdirAll(config.AmeDir, os.ModePerm)
	return copyFile(srcExe, CORE_FILE)
}

// CorePath returns the path to ame_core.exe
func CorePath() string {
	return CORE_FILE
}

// copyFile copies src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
