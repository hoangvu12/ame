package skin

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hoangvu12/ame/internal/game"
	"github.com/klauspost/compress/zstd"
)

// Voice banks are per locale. A package is generated on one locale's install,
// so its <Alias>.<locale>.wad.client carries that locale's voice events banks
// with the skin's events renamed. Those banks also hold the Sound objects,
// whose media IDs differ between dubbed locales: routed into another locale's
// WAD they point at media the player's .wpk does not have, and the skin's
// voice goes silent. Every other object, every event's action list and the
// object set are the same in all locales, so each shipped bank is rebuilt
// from the player's own copy, taking only the renamed event IDs from ours.

var localeWAD = regexp.MustCompile(`^(.+)\.([A-Za-z]{2}_[A-Za-z]{2})\.wad\.client$`)

// Overridden by tests.
var localizationGameDir = game.FindGameDir

const maxBankSize = 64 << 20

type wadEntry struct {
	hash             uint64
	offset           uint32
	compressed, size uint32
	kind             uint8
}

func readWADTOC(r io.ReaderAt) ([]wadEntry, error) {
	header := make([]byte, 272)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, err
	}
	if string(header[:2]) != "RW" || header[2] != 3 {
		return nil, fmt.Errorf("unsupported WAD version")
	}
	count := binary.LittleEndian.Uint32(header[268:])
	if count > 1<<20 {
		return nil, fmt.Errorf("WAD entry count %d", count)
	}
	toc := make([]byte, 32*int(count))
	if _, err := r.ReadAt(toc, 272); err != nil {
		return nil, err
	}
	entries := make([]wadEntry, count)
	for i := range entries {
		e := toc[32*i:]
		entries[i] = wadEntry{
			hash:       binary.LittleEndian.Uint64(e),
			offset:     binary.LittleEndian.Uint32(e[8:]),
			compressed: binary.LittleEndian.Uint32(e[12:]),
			size:       binary.LittleEndian.Uint32(e[16:]),
			kind:       e[20] & 0x0f,
		}
	}
	return entries, nil
}

var zstdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(maxBankSize))

// readWADChunk supports the stored and single-frame zstd chunks banks use.
func readWADChunk(r io.ReaderAt, e wadEntry) ([]byte, error) {
	if e.size > maxBankSize || e.compressed > maxBankSize {
		return nil, fmt.Errorf("chunk too large")
	}
	raw := make([]byte, e.compressed)
	if _, err := r.ReadAt(raw, int64(e.offset)); err != nil {
		return nil, err
	}
	var data []byte
	switch e.kind {
	case 0:
		data = raw
	case 3:
		out, err := zstdDecoder.DecodeAll(raw, make([]byte, 0, e.size))
		if err != nil {
			return nil, err
		}
		data = out
	default:
		return nil, fmt.Errorf("unsupported chunk compression %d", e.kind)
	}
	if uint32(len(data)) != e.size {
		return nil, fmt.Errorf("chunk size mismatch")
	}
	return data, nil
}

type hircObject struct {
	kind uint8
	id   uint32
	pos  int // offset of the ID inside the bank
	body []byte
}

// bankObjects reads only section and object boundaries of a Wwise bank.
func bankObjects(raw []byte) (bankID uint32, objects []hircObject, err error) {
	header, hirc := false, false
	for off := 0; off < len(raw); {
		if off+8 > len(raw) {
			return 0, nil, fmt.Errorf("truncated bank section")
		}
		tag := string(raw[off : off+4])
		end := off + 8 + int(binary.LittleEndian.Uint32(raw[off+4:]))
		if end > len(raw) || end < off+8 {
			return 0, nil, fmt.Errorf("bank section overruns bank")
		}
		switch tag {
		case "BKHD":
			if end-off < 16 {
				return 0, nil, fmt.Errorf("short bank header")
			}
			bankID, header = binary.LittleEndian.Uint32(raw[off+12:]), true
		case "HIRC":
			if hirc || end-off < 12 {
				return 0, nil, fmt.Errorf("malformed HIRC")
			}
			hirc = true
			count := binary.LittleEndian.Uint32(raw[off+8:])
			pos := off + 12
			for i := uint32(0); i < count; i++ {
				if pos+9 > end {
					return 0, nil, fmt.Errorf("object overruns HIRC")
				}
				length := int(binary.LittleEndian.Uint32(raw[pos+1:]))
				if length < 4 || pos+5+length > end {
					return 0, nil, fmt.Errorf("object overruns HIRC")
				}
				objects = append(objects, hircObject{raw[pos], binary.LittleEndian.Uint32(raw[pos+5:]), pos + 5, raw[pos+9 : pos+5+length]})
				pos += 5 + length
			}
			if pos != end {
				return 0, nil, fmt.Errorf("HIRC size mismatch")
			}
		}
		off = end
	}
	if !header || !hirc {
		return 0, nil, fmt.Errorf("bank without header or HIRC")
	}
	return bankID, objects, nil
}

const hircEvent = 4

// transplantBank returns the player's bank with each event carrying the ID of
// the shipped event that has the same action list. Any structural difference
// beyond Sound bodies and object order fails.
func transplantBank(shipped, user []byte) ([]byte, int, error) {
	shippedID, shippedObjects, err := bankObjects(shipped)
	if err != nil {
		return nil, 0, err
	}
	userID, userObjects, err := bankObjects(user)
	if err != nil {
		return nil, 0, err
	}
	if shippedID != userID {
		return nil, 0, fmt.Errorf("different bank")
	}
	if !sameNonEvents(shippedObjects, userObjects) {
		return nil, 0, fmt.Errorf("bank objects differ")
	}
	// Events with the same action list behave identically; pair them in order.
	byActions := map[string][]uint32{}
	events := 0
	for _, o := range shippedObjects {
		if o.kind == hircEvent {
			byActions[string(o.body)] = append(byActions[string(o.body)], o.id)
			events++
		}
	}
	out := append([]byte(nil), user...)
	renamed := 0
	for _, o := range userObjects {
		if o.kind != hircEvent {
			continue
		}
		ids := byActions[string(o.body)]
		if len(ids) == 0 {
			return nil, 0, fmt.Errorf("event %08x has no shipped counterpart", o.id)
		}
		byActions[string(o.body)] = ids[1:]
		events--
		if ids[0] != o.id {
			binary.LittleEndian.PutUint32(out[o.pos:], ids[0])
			renamed++
		}
	}
	if events != 0 {
		return nil, 0, fmt.Errorf("event count differs")
	}
	return out, renamed, nil
}

func sameNonEvents(a, b []hircObject) bool {
	keys := func(objects []hircObject) []uint64 {
		var out []uint64
		for _, o := range objects {
			if o.kind != hircEvent {
				out = append(out, uint64(o.kind)<<32|uint64(o.id))
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
		return out
	}
	x, y := keys(a), keys(b)
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// localizeVoiceWADs rewrites each shipped locale WAD in an extracted mod into
// the locales the game has installed. A bank that cannot be rebuilt keeps the
// shipped copy, which is what the package delivered before.
func localizeVoiceWADs(modDir, gameDir string) {
	if gameDir == "" {
		return
	}
	wadDir := filepath.Join(modDir, "WAD")
	entries, err := os.ReadDir(wadDir)
	if err != nil {
		return
	}
	installed, err := os.ReadDir(filepath.Join(gameDir, "DATA", "FINAL", "Champions"))
	if err != nil {
		return
	}
	for _, entry := range entries {
		m := localeWAD.FindStringSubmatch(entry.Name())
		if m == nil || entry.IsDir() {
			continue
		}
		var targets []string
		keep := false
		for _, game := range installed {
			g := localeWAD.FindStringSubmatch(game.Name())
			if g == nil || game.IsDir() || !strings.EqualFold(g[1], m[1]) {
				continue
			}
			if strings.EqualFold(g[2], m[2]) {
				keep = true
				continue
			}
			targets = append(targets, game.Name())
		}
		if len(targets) == 0 {
			continue
		}
		shipped := filepath.Join(wadDir, entry.Name())
		ok := true
		for _, target := range targets {
			if err := localizeVoiceWAD(shipped, filepath.Join(gameDir, "DATA", "FINAL", "Champions", target), filepath.Join(wadDir, target)); err != nil {
				logSkin(fmt.Sprintf("voice banks for %s kept as shipped: %v", target, err))
				ok = false
			}
		}
		if ok && !keep {
			os.Remove(shipped)
		}
	}
}

// localizeVoiceWAD writes a directory WAD at dest holding every chunk of the
// shipped WAD, rebuilt from the installed locale WAD where possible.
func localizeVoiceWAD(shippedPath, installedPath, dest string) error {
	shippedFile, err := os.Open(shippedPath)
	if err != nil {
		return err
	}
	defer shippedFile.Close()
	shippedTOC, err := readWADTOC(shippedFile)
	if err != nil {
		return err
	}
	installedFile, err := os.Open(installedPath)
	if err != nil {
		return err
	}
	defer installedFile.Close()
	installedTOC, err := readWADTOC(installedFile)
	if err != nil {
		return err
	}
	byHash := make(map[uint64]wadEntry, len(installedTOC))
	for _, e := range installedTOC {
		byHash[e.hash] = e
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	rebuilt, renamed := 0, 0
	for _, e := range shippedTOC {
		data, err := readWADChunk(shippedFile, e)
		if err != nil {
			return err
		}
		if own, found := byHash[e.hash]; found {
			if user, err := readWADChunk(installedFile, own); err == nil {
				if out, n, err := transplantBank(data, user); err == nil {
					data = out
					rebuilt++
					renamed += n
				}
			}
		}
		if err := os.WriteFile(filepath.Join(dest, fmt.Sprintf("%016x.bnk", e.hash)), data, 0644); err != nil {
			return err
		}
	}
	logSkin(fmt.Sprintf("rebuilt %d of %d voice banks for %s (%d events)", rebuilt, len(shippedTOC), filepath.Base(dest), renamed))
	return nil
}
