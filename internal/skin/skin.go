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
const CDRAGON_BASE_URL = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/champions"

// CommunityDragon data structures for English name resolution
type cdragonChroma struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type cdragonSkin struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Chromas []cdragonChroma `json:"chromas"`
}

type cdragonChampion struct {
	Name  string        `json:"name"`
	Skins []cdragonSkin `json:"skins"`
}

var (
	champCache   = make(map[string]*cdragonChampion)
	champCacheMu sync.Mutex
)

// fetchChampionEnglish fetches English champion data from CommunityDragon and caches it
func fetchChampionEnglish(championID string) (*cdragonChampion, error) {
	champCacheMu.Lock()
	if cached, ok := champCache[championID]; ok {
		champCacheMu.Unlock()
		return cached, nil
	}
	champCacheMu.Unlock()

	fetchURL := fmt.Sprintf("%s/%s.json", CDRAGON_BASE_URL, championID)
	resp, err := http.Get(fetchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CommunityDragon returned status %d", resp.StatusCode)
	}

	var data cdragonChampion
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	champCacheMu.Lock()
	champCache[championID] = &data
	champCacheMu.Unlock()

	return &data, nil
}

// resolveEnglishNames looks up English champion/skin/chroma names from CommunityDragon
func resolveEnglishNames(championID, skinID, baseSkinID string) (champName, skinName, chromaName string) {
	data, err := fetchChampionEnglish(championID)
	if err != nil {
		return "", "", ""
	}

	champName = data.Name

	skinIDNum, _ := strconv.Atoi(skinID)
	baseSkinIDNum, _ := strconv.Atoi(baseSkinID)

	if baseSkinID != "" && baseSkinIDNum != 0 {
		// Chroma: baseSkinID is the parent skin, skinID is the chroma
		for _, s := range data.Skins {
			if s.ID == baseSkinIDNum {
				skinName = s.Name
				for _, c := range s.Chromas {
					if c.ID == skinIDNum {
						chromaName = c.Name
						break
					}
				}
				break
			}
		}
	} else {
		// Non-chroma: skinID is the skin itself
		for _, s := range data.Skins {
			if s.ID == skinIDNum {
				skinName = s.Name
				break
			}
		}
	}

	return
}

// Download downloads a skin file (.fantome or .zip)
func Download(championID, skinID, baseSkinID, championName, skinName, chromaName string) (string, error) {
	// Resolve English names from CommunityDragon (overrides localized names from client)
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
