package skin

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/hoangvu12/ame/internal/config"
)

const SKIN_BASE_URL = "https://raw.githubusercontent.com/Alban1911/LeagueSkins/main/skins"
const SKIN_IDS_URL = "https://raw.githubusercontent.com/Alban1911/LeagueSkins/refs/heads/main/resources/en/skin_ids.json"

var (
	skinIDsCache   map[string]string
	skinIDsCacheMu sync.Mutex
)

// fetchSkinIDs fetches the skin ID-to-name mapping and caches it
func fetchSkinIDs() (map[string]string, error) {
	skinIDsCacheMu.Lock()
	if skinIDsCache != nil {
		skinIDsCacheMu.Unlock()
		return skinIDsCache, nil
	}
	skinIDsCacheMu.Unlock()

	resp, err := http.Get(SKIN_IDS_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skin IDs returned status %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	skinIDsCacheMu.Lock()
	skinIDsCache = data
	skinIDsCacheMu.Unlock()

	return data, nil
}

// resolveEnglishNames looks up English champion/skin/chroma names from the skin IDs mapping
func resolveEnglishNames(championID, skinID, baseSkinID string) (champName, skinName, chromaName string) {
	data, err := fetchSkinIDs()
	if err != nil {
		return "", "", ""
	}

	// Champion name is the base skin entry (championID * 1000)
	championIDNum, _ := strconv.Atoi(championID)
	champName = data[strconv.Itoa(championIDNum*1000)]

	baseSkinIDNum, _ := strconv.Atoi(baseSkinID)

	if baseSkinID != "" && baseSkinIDNum != 0 {
		// Chroma: baseSkinID is the parent skin, skinID is the chroma
		skinName = data[baseSkinID]
		chromaName = data[skinID]
	} else {
		// Non-chroma: skinID is the skin itself
		skinName = data[skinID]
	}

	return
}

// Download downloads a skin file (.fantome or .zip)
func Download(championID, skinID, baseSkinID, championName, skinName, chromaName string) (string, error) {
	// Try RSE source first (encrypted skins)
	if rseAvailable() {
		if path, err := downloadRSE(championID, skinID, baseSkinID); err == nil {
			return path, nil
		}
	}

	// Fallback to LeagueSkins repo
	// Resolve English names from skin IDs mapping (overrides localized names from client)
	enChamp, enSkin, enChroma := resolveEnglishNames(championID, skinID, baseSkinID)
	if enChamp != "" {
		championName = enChamp
	}
	if enSkin != "" {
		skinName = enSkin
	}
	if enChroma != "" {
		chromaName = enChroma
	}

	skinDir := filepath.Join(config.SkinsDir, championID, skinID)
	os.MkdirAll(skinDir, os.ModePerm)

	extensions := []string{"zip", "fantome"}

	for _, ext := range extensions {
		var downloadURL string
		if championName != "" && skinName != "" {
			if chromaName != "" {
				// Chroma: skins/{ChampionName}/{SkinName}/{ChromaName}/{ChromaName}.ext
				downloadURL = fmt.Sprintf("%s/%s/%s/%s/%s.%s",
					SKIN_BASE_URL,
					url.PathEscape(championName),
					url.PathEscape(skinName),
					url.PathEscape(chromaName),
					url.PathEscape(chromaName),
					ext,
				)
			} else {
				// Base skin: skins/{ChampionName}/{SkinName}/{SkinName}.ext
				downloadURL = fmt.Sprintf("%s/%s/%s/%s.%s",
					SKIN_BASE_URL,
					url.PathEscape(championName),
					url.PathEscape(skinName),
					url.PathEscape(skinName),
					ext,
				)
			}
		} else {
			// Fallback to old numeric URL pattern
			if baseSkinID != "" {
				downloadURL = fmt.Sprintf("%s/%s/%s/%s/%s.%s", SKIN_BASE_URL, championID, baseSkinID, skinID, skinID, ext)
			} else {
				downloadURL = fmt.Sprintf("%s/%s/%s/%s.%s", SKIN_BASE_URL, championID, skinID, skinID, ext)
			}
		}

		filePath := filepath.Join(skinDir, fmt.Sprintf("%s.%s", skinID, ext))

		if err := downloadFile(downloadURL, filePath); err == nil {
			return filePath, nil
		}
	}

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
	skinDir := filepath.Join(config.SkinsDir, championID, skinID)

	// Check for zip first
	zipPath := filepath.Join(skinDir, fmt.Sprintf("%s.zip", skinID))
	if _, err := os.Stat(zipPath); err == nil {
		return zipPath
	}

	// Check for fantome
	fantomePath := filepath.Join(skinDir, fmt.Sprintf("%s.fantome", skinID))
	if _, err := os.Stat(fantomePath); err == nil {
		return fantomePath
	}

	return ""
}

// downloadFile downloads a file from URL to destination
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
