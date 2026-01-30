package skin

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hoangvu12/ame/internal/config"
)

const SKIN_BASE_URL = "https://raw.githubusercontent.com/Alban1911/LeagueSkins/main/skins"

// Download downloads a skin file (.fantome or .zip)
func Download(championID, skinID, baseSkinID string) (string, error) {
	skinDir := filepath.Join(config.SkinsDir, championID, skinID)
	os.MkdirAll(skinDir, os.ModePerm)

	extensions := []string{"zip", "fantome"}

	for _, ext := range extensions {
		var url string
		if baseSkinID != "" {
			url = fmt.Sprintf("%s/%s/%s/%s/%s.%s", SKIN_BASE_URL, championID, baseSkinID, skinID, skinID, ext)
		} else {
			url = fmt.Sprintf("%s/%s/%s/%s.%s", SKIN_BASE_URL, championID, skinID, skinID, ext)
		}

		filePath := filepath.Join(skinDir, fmt.Sprintf("%s.%s", skinID, ext))

		if err := downloadFile(url, filePath); err == nil {
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
