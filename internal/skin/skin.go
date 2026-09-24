package skin

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/zipname"
)

const cacheMetadataFile = "cache.json"

// cacheMetadata is the per-skin cache record. Only locally generated
// packages are tracked; fields from older records are ignored on read.
type cacheMetadata struct {
	Local     *localMetadata `json:"local,omitempty"`
	Source    string         `json:"source"`
	CheckedAt time.Time      `json:"checkedAt"`
}

func skinDir(championID, skinID string) string {
	return filepath.Join(config.SkinsDir, championID, skinID)
}

func skinCacheMetadataPath(championID, skinID string) string {
	return filepath.Join(skinDir(championID, skinID), cacheMetadataFile)
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

// Extract extracts a zip/fantome file to destination directory
func Extract(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := strings.ReplaceAll(f.Name, `\`, "/")
		if zipname.Unsafe(f.Name) {
			return fmt.Errorf("unsafe entry name: %s", f.Name)
		}
		fpath := filepath.Join(destDir, filepath.FromSlash(name))

		// Defense in depth: the shared name rule above already rejects
		// escapes; this resolved-path check catches anything else that
		// could still leave destDir.
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() || strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
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
