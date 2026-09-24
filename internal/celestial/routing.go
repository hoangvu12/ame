package celestial

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverTargets locates existing holders of the requested chunk hashes.
// It reads WAD directories only; it does not infer a destination for new hashes,
// choose TFT policy, or alter the caller's explicit mod routing.
func DiscoverTargets(gameDir string, hashes []uint64) (map[uint64][]string, error) {
	result := make(map[uint64][]string, len(hashes))
	for _, hash := range hashes {
		result[hash] = nil
	}
	if len(result) == 0 {
		return result, nil
	}
	root := filepath.Join(gameDir, "DATA", "FINAL")
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".wad.client") {
			return nil
		}
		file, err := os.Open(name)
		if err != nil {
			return err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		var header [272]byte
		if _, err := file.ReadAt(header[:], 0); err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		if string(header[:2]) != "RW" || header[2] != 3 || header[3] > 4 {
			return fmt.Errorf("unsupported game WAD: %s", entry.Name())
		}
		count := binary.LittleEndian.Uint32(header[268:])
		if uint64(count)*32+272 > uint64(info.Size()) {
			return fmt.Errorf("truncated game WAD: %s", entry.Name())
		}
		relative, err := filepath.Rel(gameDir, name)
		if err != nil {
			return err
		}
		target := strings.ToLower(filepath.ToSlash(relative))
		// Bounded batches avoid a syscall per entry without allocating based
		// on an untrusted count from the file header.
		buffer := make([]byte, 32*4096)
		seen := make(map[uint64]bool)
		for first := uint32(0); first < count; {
			n := count - first
			if n > 4096 {
				n = 4096
			}
			batch := buffer[:int(n)*32]
			if _, err := file.ReadAt(batch, 272+int64(first)*32); err != nil {
				return err
			}
			for off := 0; off < len(batch); off += 32 {
				hash := binary.LittleEndian.Uint64(batch[off:])
				if _, wanted := result[hash]; wanted && !seen[hash] {
					result[hash] = append(result[hash], target)
					seen[hash] = true
				}
			}
			first += n
		}
		end, err := file.Stat()
		if err != nil {
			return err
		}
		if info.Size() != end.Size() || !info.ModTime().Equal(end.ModTime()) {
			return fmt.Errorf("game WAD changed during routing scan")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for hash := range result {
		sort.Strings(result[hash])
	}
	return result, nil
}
