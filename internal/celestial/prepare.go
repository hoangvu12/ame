package celestial

import (
	"fmt"
	"os"
	"path/filepath"
)

// PreparedOverlay owns only Directory. Referenced source WADs remain owned by
// the caller and must stay immutable until the runtime stops using the overlay.
// Preparation does not authorize or start the Celestial DLL.
type PreparedOverlay struct {
	Directory string
	SealRoot  [32]byte
	Sources   int
	Targets   int
	Chunks    int
}

// PrepareOverlay builds the verified single-index format into a fresh directory
// under parent. Explicit routing and later-input-wins priority follow BuildIndex.
// No existing overlay is replaced, including on failure.
func PrepareOverlay(parent string, inputs []WADInput) (*PreparedOverlay, error) {
	if !filepath.IsAbs(parent) {
		return nil, fmt.Errorf("overlay parent must be absolute")
	}
	index, err := BuildIndex(inputs)
	if err != nil {
		return nil, err
	}
	chunks := 0
	for _, target := range index.Targets {
		chunks += len(target.Chunks)
	}
	if chunks == 0 {
		return nil, fmt.Errorf("overlay contains no replacement chunks")
	}
	encoded, err := EncodeIndex(index)
	if err != nil {
		return nil, err
	}
	seal, err := IndexSeal(encoded)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(parent, "ame-celestial-")
	if err != nil {
		return nil, err
	}
	// Only remove individual files inside the directory created by this call.
	// Never recursively remove a caller-supplied path.
	complete := false
	defer func() {
		if !complete {
			os.Remove(filepath.Join(dir, "wad.idx"))
			os.Remove(filepath.Join(dir, "celestial-overlay.seal"))
			os.Remove(dir)
		}
	}()
	if err := os.WriteFile(filepath.Join(dir, "wad.idx"), encoded, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "celestial-overlay.seal"), seal, 0600); err != nil {
		return nil, err
	}
	complete = true
	return &PreparedOverlay{dir, SealRoot(seal), len(index.Sources), len(index.Targets), chunks}, nil
}
