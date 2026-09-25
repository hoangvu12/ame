package skin

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"github.com/hoangvu12/ame/internal/zipname"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func validatePackage(path string) error {
	z, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer z.Close()
	meta, wad := false, false
	for _, f := range z.File {
		name := strings.ReplaceAll(f.Name, `\`, "/")
		if zipname.Unsafe(f.Name) {
			return fmt.Errorf("unsafe package path")
		}
		if name == "META/info.json" {
			meta = true
		}
		if strings.HasPrefix(name, "WAD/") && f.UncompressedSize64 > 16 {
			wad = true
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, r)
		r.Close()
		if err != nil {
			return err
		}
	}
	if !meta || !wad {
		return fmt.Errorf("package missing metadata or WAD")
	}
	return nil
}

// ArtifactKey includes source provenance and live input fingerprints. A patch
// invalidates an overlay even before the corresponding package is regenerated.
func ArtifactKey(skinID string) string {
	id, err := strconv.Atoi(skinID)
	if err != nil || id < 1000 {
		return skinID
	}
	champion := strconv.Itoa(id / 1000)
	m, err := readCacheMetadata(champion, skinID)
	if err != nil {
		return skinID
	}
	h := sha256.New()
	// CheckedAt is cache bookkeeping, not artifact identity.
	if m.Source == "local" && m.Local != nil {
		fmt.Fprint(h, m.Local.Digest, m.Local.Generator, m.Local.Schema, generatorVersion)
		if dir, err := helperDir(); err == nil {
			fmt.Fprint(h, filepath.Base(dir))
		}
		for _, input := range m.Local.Inputs {
			current, _ := fingerprint(input.Path)
			fmt.Fprint(h, current)
		}
		return fmt.Sprintf("%s@%x", skinID, h.Sum(nil)[:12])
	}
	if m.Source == "remote" && m.Remote != nil && len(m.Remote.Digest) >= 12 {
		// The remote artifact identity is its content digest; a catalog
		// revision with new bytes must invalidate cached overlays.
		return skinID + "@" + m.Remote.Digest[:12]
	}
	return skinID
}

// OverlayKey keeps the existing selection key format and only adds provenance.
func OverlayKey(selection string) string {
	parts := strings.SplitN(selection, ";", 2)
	ids := strings.Split(parts[0], ",")
	for i, id := range ids {
		ids[i] = ArtifactKey(id)
	}
	parts[0] = strings.Join(ids, ",")
	return strings.Join(parts, ";")
}

var extractMu sync.Mutex

// ExtractPackage refreshes a teammate mod only when its package bytes changed.
// Publishing a complete directory avoids retaining files from the previous mod.
func ExtractPackage(archive, dest string) error {
	extractMu.Lock()
	defer extractMu.Unlock()
	data, err := os.ReadFile(archive)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%x", sha256.Sum256(data))
	marker := filepath.Join(dest, ".ame-package")
	if old, err := os.ReadFile(marker); err == nil && string(old) == id {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := Extract(archive, tmp); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, ".ame-package"), []byte(id), 0600); err != nil {
		return err
	}
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}
