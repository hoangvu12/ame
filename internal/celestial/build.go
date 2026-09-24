package celestial

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// WADInput explicitly specifies routing. Discovering cross-WAD destinations
// from the installed game is a separate, as-yet unimplemented backend step.
type WADInput struct {
	Path    string
	Targets []string
}

// BuildIndex reads source WAD directories without modifying them. Inputs are
// in AME priority order: later inputs win on equal target/hash pairs. Only the
// ordinary zstd records verified against the installed fixture are supported.
// Sources must remain immutable while the resulting index is in use.
func BuildIndex(inputs []WADInput) (*Index, error) {
	index := &Index{}
	targets := make(map[string]map[uint64]Chunk)
	for _, input := range inputs {
		if !filepath.IsAbs(input.Path) {
			return nil, fmt.Errorf("source WAD must have an absolute path")
		}
		if len(input.Targets) == 0 {
			return nil, fmt.Errorf("source WAD has no explicit targets")
		}
		for _, name := range input.Targets {
			if name != strings.ToLower(name) || strings.ContainsAny(name, "\\:\x00") ||
				!strings.HasPrefix(name, "data/final/") || !strings.HasSuffix(name, ".wad.client") || path.Clean(name) != name {
				return nil, fmt.Errorf("invalid target WAD path: %q", name)
			}
		}
		file, err := os.Open(input.Path)
		if err != nil {
			return nil, err
		}
		info, err := file.Stat()
		if err != nil {
			file.Close()
			return nil, err
		}
		if !info.Mode().IsRegular() {
			file.Close()
			return nil, fmt.Errorf("source WAD is not a regular file")
		}
		chunks, err := readWADDirectory(file, info.Size())
		endInfo, statErr := file.Stat()
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("read WAD directory: %w", err)
		}
		if statErr != nil {
			return nil, statErr
		}
		if info.Size() != endInfo.Size() || !info.ModTime().Equal(endInfo.ModTime()) {
			return nil, fmt.Errorf("source WAD changed while reading")
		}
		sourceID := uint32(len(index.Sources))
		index.Sources = append(index.Sources, input.Path)
		for _, name := range input.Targets {
			if targets[name] == nil {
				targets[name] = make(map[uint64]Chunk)
			}
			for _, chunk := range chunks {
				chunk.Source = sourceID
				targets[name][chunk.Hash] = chunk
			}
		}
	}
	for name, chunks := range targets {
		target := Target{Name: name}
		for _, chunk := range chunks {
			target.Chunks = append(target.Chunks, chunk)
		}
		sort.Slice(target.Chunks, func(i, j int) bool { return target.Chunks[i].Hash < target.Chunks[j].Hash })
		index.Targets = append(index.Targets, target)
	}
	sort.Slice(index.Targets, func(i, j int) bool { return index.Targets[i].Name < index.Targets[j].Name })
	return index, nil
}

func readWADDirectory(reader io.ReaderAt, size int64) ([]Chunk, error) {
	var header [272]byte
	if _, err := reader.ReadAt(header[:], 0); err != nil {
		return nil, err
	}
	if string(header[:2]) != "RW" || header[2] != 3 || header[3] > 4 {
		return nil, fmt.Errorf("unsupported WAD version")
	}
	count := binary.LittleEndian.Uint32(header[268:])
	if size < 272 || uint64(count)*32+272 > uint64(size) {
		return nil, fmt.Errorf("truncated WAD directory")
	}
	if uint64(count)*40 > MaxIndexBytes {
		return nil, fmt.Errorf("WAD directory exceeds index limit")
	}
	chunks := make([]Chunk, 0, count)
	seen := make(map[uint64]bool)
	var entry [32]byte
	for n := uint32(0); n < count; n++ {
		if _, err := reader.ReadAt(entry[:], 272+int64(n)*32); err != nil {
			return nil, err
		}
		// Nonzero subchunk metadata requires verification of the remaining
		// CHK3 trailer fields. Never emit a guessed record for those formats.
		if entry[20] != 3 || binary.LittleEndian.Uint16(entry[22:]) != 0 || entry[21] > 1 {
			return nil, fmt.Errorf("unsupported WAD compression/subchunk metadata")
		}
		chunk := Chunk{
			Hash:           binary.LittleEndian.Uint64(entry[:]),
			Offset:         uint64(binary.LittleEndian.Uint32(entry[8:])),
			CompressedSize: binary.LittleEndian.Uint32(entry[12:]),
			Size:           binary.LittleEndian.Uint32(entry[16:]),
			Checksum:       binary.LittleEndian.Uint64(entry[24:]),
			Trailer:        [4]byte{3, 0, 0, 0},
		}
		if chunk.Offset < uint64(count)*32+272 || chunk.Offset+uint64(chunk.CompressedSize) > uint64(size) {
			return nil, fmt.Errorf("WAD chunk outside data region")
		}
		if seen[chunk.Hash] {
			return nil, fmt.Errorf("duplicate WAD path hash")
		}
		seen[chunk.Hash] = true
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}
