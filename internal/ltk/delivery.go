// Delivery of the LTK Manager runtime ame drives.
//
// ame downloads the newest LTK Manager release from GitHub and extracts the
// signed ltk-manager.exe plus its bundled ltk_patcher_host.exe into one
// shared, release-addressed directory that every ame build reuses. The
// patcher DLL is not delivered here: the manager fetches it itself from
// LeagueToolkit's delivery service when its patcher starts.
//
// Nothing is re-downloaded while the release tag and the extracted pair stay
// in place, and an unreachable GitHub falls back to the already-delivered
// copy, so an offline machine keeps patching on its current version.

package ltk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
)

// ltkReleasesURL is the GitHub API entry for the newest LTK Manager release.
const ltkReleasesURL = "https://api.github.com/repos/LeagueToolkit/ltk-manager/releases/latest"

// managerExeName and managerHostName are the pair ame extracts from the
// release installer: the manager ame launches, and the host it uses as its
// bundled injector, which must sit next to it.
const (
	managerExeName  = "ltk-manager.exe"
	managerHostName = "ltk_patcher_host.exe"
)

// managerReleaseFile records the LTK Manager release the delivered pair came
// from.
const managerReleaseFile = "ltk-manager.release"

// managerDir is the shared cache for the delivered manager runtime. It is
// deliberately addressed by release tag, not by ame's content-addressed
// runtime hash, so every ame build reuses the same delivered pair instead of
// stacking copies per runtime revision. It is resolved once at startup, so
// tests that redirect config.AmeDir do not disturb the cache.
var managerDir = filepath.Join(config.AmeDir, "manager")

// managerFetch is ManagerFetch, overridden by tests to stay offline.
var managerFetch = ManagerFetch

// latestSetup is ltkLatestSetup, overridden by tests to stay offline.
var latestSetup = ltkLatestSetup

// managerRecheckInterval: a release check this recent is reused. Setup
// checks at startup, and every manager launch used to check again, which
// put a GitHub round-trip (up to its 30s timeout on a bad network) between
// a skin pick and the patcher scan.
const managerRecheckInterval = 6 * time.Hour

var (
	releaseCheckMu sync.Mutex
	releaseChecked time.Time // the last check that left a usable pair
)

func markReleaseChecked() {
	releaseCheckMu.Lock()
	releaseChecked = time.Now()
	releaseCheckMu.Unlock()
}

func releaseCheckedRecently() bool {
	releaseCheckMu.Lock()
	defer releaseCheckMu.Unlock()
	return !releaseChecked.IsZero() && time.Since(releaseChecked) < managerRecheckInterval
}

// ManagerFetch ensures the shared manager directory holds the signed
// ltk-manager.exe and its bundled ltk_patcher_host.exe from the newest LTK
// Manager release, and returns the path of ltk-manager.exe.
//
// Nothing is downloaded when the release marker matches the newest tag, and
// GitHub is not asked again within managerRecheckInterval of a check. When
// GitHub cannot be reached, or the download or extraction fails, the
// previously delivered pair keeps being used; only a machine that has never
// delivered the pair fails.
func ManagerFetch() (string, error) {
	exePath := filepath.Join(managerDir, managerExeName)
	hostPath := filepath.Join(managerDir, managerHostName)
	if releaseCheckedRecently() && managerPairUsable(exePath, hostPath) {
		return exePath, nil
	}

	tag, setupURL, err := latestSetup()
	if err == nil {
		if marked, readErr := os.ReadFile(filepath.Join(managerDir, managerReleaseFile)); readErr == nil &&
			strings.TrimSpace(string(marked)) == tag && managerPairUsable(exePath, hostPath) {
			markReleaseChecked()
			return exePath, nil
		}
		if setup, downloadErr := downloadFile(setupURL); downloadErr == nil {
			if pair, extractErr := nsisExtractFiles(setup, managerExeName, managerHostName); extractErr == nil {
				if err := installManagerPair(exePath, hostPath, pair); err != nil {
					display.Log("LTK: manager delivery: " + err.Error())
				} else {
					if err := os.WriteFile(filepath.Join(managerDir, managerReleaseFile), []byte(tag), 0644); err != nil {
						display.Log("LTK: manager release marker: " + err.Error())
					}
					display.Log(fmt.Sprintf("LTK: delivered LTK Manager %s to %s", tag, managerDir))
					markReleaseChecked()
					return exePath, nil
				}
			} else {
				display.Log("LTK: manager extraction: " + extractErr.Error())
			}
		} else {
			display.Log("LTK: manager download: " + downloadErr.Error())
		}
	} else {
		display.Log("LTK: manager release check: " + err.Error())
	}

	if managerPairUsable(exePath, hostPath) {
		display.Log("LTK: using the previously delivered LTK Manager runtime")
		// A failed check is not repeated at every launch either.
		markReleaseChecked()
		return exePath, nil
	}
	return "", fmt.Errorf("LTK Manager runtime is unavailable: %v (ame must reach GitHub once to deliver it)", err)
}

// managerPairUsable verifies the delivered pair exists and is not truncated.
func managerPairUsable(exePath, hostPath string) bool {
	for _, path := range []string{exePath, hostPath} {
		info, err := os.Stat(path)
		if err != nil || info.Size() < 1024 {
			return false
		}
	}
	return true
}

// installManagerPair writes the extracted pair atomically: temporary files
// renamed into place, so a killed download never leaves a half-written
// executable behind. Any running manager must already be terminated.
func installManagerPair(exePath, hostPath string, pair map[string][]byte) error {
	if err := os.MkdirAll(managerDir, 0755); err != nil {
		return err
	}
	writes := []struct {
		final string
		temp  string
		data  []byte
	}{
		{exePath, exePath + ".tmp", pair[managerExeName]},
		{hostPath, hostPath + ".tmp", pair[managerHostName]},
	}
	for _, write := range writes {
		if len(write.data) < 1024 {
			return fmt.Errorf("extracted %s looks truncated (%d bytes)", filepath.Base(write.final), len(write.data))
		}
		if err := os.WriteFile(write.temp, write.data, 0644); err != nil {
			return err
		}
	}
	for _, write := range writes {
		if err := os.Rename(write.temp, write.final); err != nil {
			return err
		}
	}
	return nil
}

// ltkLatestSetup returns the newest LTK Manager release tag and its NSIS
// installer URL.
func ltkLatestSetup() (string, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", ltkReleasesURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "ame")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api: HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("github json: %w", err)
	}
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, "-setup.exe") && !strings.HasSuffix(asset.Name, ".sig") {
			return release.TagName, asset.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("no setup.exe in %s", release.TagName)
}

// downloadFile fetches an installer-sized file. The timeout is generous but
// finite so a stalled host fetch cannot block startup forever.
func downloadFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
