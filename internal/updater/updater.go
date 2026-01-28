package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	GITHUB_REPO    = "hoangvu12/ame"
	GITHUB_API_URL = "https://api.github.com/repos/" + GITHUB_REPO + "/releases/latest"
)

var (
	AME_DIR      = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame")
	VERSION_FILE = filepath.Join(AME_DIR, "version.txt")
	UPDATE_FILE  = filepath.Join(AME_DIR, "ame_update.exe")
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
	UpdatePath      string
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
	os.MkdirAll(AME_DIR, os.ModePerm)
	return os.WriteFile(VERSION_FILE, []byte(version), 0644)
}

// fetchLatestRelease gets the latest release info from GitHub
func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{}
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
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(UPDATE_FILE)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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
		result.UpdatePath = UPDATE_FILE
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
}

// ApplyUpdate creates a batch script to replace the exe and relaunch
// Returns the path to the batch script
func ApplyUpdate(currentExePath string) (string, error) {
	scriptPath := filepath.Join(AME_DIR, "update.bat")

	// Batch script that:
	// 1. Waits for the current process to exit
	// 2. Copies the new exe over the old one
	// 3. Deletes the update file and itself
	// 4. Relaunches the app
	script := fmt.Sprintf(`@echo off
echo Updating ame...
timeout /t 2 /nobreak >nul
copy /y "%s" "%s"
del "%s"
start "" "%s"
del "%%~f0"
`, UPDATE_FILE, currentExePath, UPDATE_FILE, currentExePath)

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return "", err
	}

	return scriptPath, nil
}
