// Minimal NSIS 3 installer reader.
//
// ame extracts ltk-manager.exe and ltk_patcher_host.exe from the LTK Manager
// release installer without 7-Zip on PATH. The supported subset is what
// makensis emits for Tauri releases: the 28-byte first header, a header
// block (LZMA, deflate, or stored), the entry/string tables of the
// decompressed header, and either the solid stream or per-file data blocks.
// Every compressed stream is decoded with an unknown output size, read to
// its end-of-stream marker, exactly as 7-Zip decodes NSIS archives.
//
// Format reference: NSIS Source/fileform.h and 7-Zip's NsisIn.cpp
// (CInArchive::Open2), as documented by the nsis-rs parser.

package ltk

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/ulikunitz/xz/lzma"
)

const (
	nsisFirstHeaderSize = 28
	nsisEntrySize       = 28 // opcode + six offsets (NSIS 2.x/3.x)
	nsisBlocksNum       = 8

	// First header flags.
	nsisFlagNoCRC uint32 = 0x04

	// Entry opcodes (NSIS 3, non-logging build).
	nsisOpExtractFile int32 = 20

	// Block indices in the common header.
	nsisBlockEntries = 2
	nsisBlockStrings = 3

	// Bounds the decompressed size of anything read from an installer, so a
	// corrupt archive cannot exhaust memory.
	nsisMaxDecompressed = 512 << 20
)

// nsisFirstHeader is the 28-byte record at the start of the NSIS archive:
// flags, the 0xDEADBEEF magic, "NullsoftInst", the decompressed header size,
// and the total size of everything from the first header onward (including
// the optional trailing CRC).
type nsisFirstHeader struct {
	flags           uint32
	lengthOfHeader  int
	lengthOfAllData int
}

// nsisArchive is a parsed installer: the decompressed header (entries and
// strings) plus the file data, either as one decompressed solid stream or as
// the raw region following the header block.
type nsisArchive struct {
	entries   []nsisEntry
	strings   []byte
	solidData []byte // solid: decompressed file data
	data      []byte // non-solid: raw bytes after the header block
}

// nsisEntry is one bytecode instruction: an opcode and six parameters.
type nsisEntry struct {
	which   int32
	offsets [6]int32
}

// nsisOpen parses an NSIS installer image held in memory.
func nsisOpen(data []byte) (*nsisArchive, error) {
	overlay := data[peOverlayStart(data):]
	offset, first, ok := nsisFindFirstHeader(overlay)
	if !ok {
		return nil, fmt.Errorf("no NSIS archive in installer")
	}
	// lengthOfAllData counts everything from the first header onward, so the
	// archive ends at offset + lengthOfAllData.
	end := offset + first.lengthOfAllData
	if end > len(overlay) {
		return nil, fmt.Errorf("NSIS archive truncated: need %d bytes, have %d", end, len(overlay))
	}
	after := overlay[offset+nsisFirstHeaderSize:]

	// The solid stream ends at the archive bounds, excluding the optional
	// trailing CRC: the LZMA decoder must not see bytes past the marker.
	streamLen := first.lengthOfAllData - nsisFirstHeaderSize
	if first.flags&nsisFlagNoCRC == 0 {
		streamLen -= 4
	}
	if streamLen > len(after) {
		streamLen = len(after)
	}
	stream := after[:streamLen]

	archive := &nsisArchive{}
	var header []byte
	switch {
	case len(stream) >= 12 && binary.LittleEndian.Uint32(stream) == uint32(first.lengthOfHeader):
		// Stored, non-solid header block.
		if 4+first.lengthOfHeader > len(stream) {
			return nil, fmt.Errorf("stored header block truncated")
		}
		header = stream[4 : 4+first.lengthOfHeader]
		archive.data = after[4+first.lengthOfHeader:]
	case nsisIsLZMAProps(stream):
		// Solid LZMA: one stream holds the size-prefixed header and every
		// file.
		decoded, err := nsisDecompressStream("lzma", stream)
		if err != nil {
			return nil, fmt.Errorf("NSIS solid stream: %w", err)
		}
		header, archive.solidData = nsisSplitSolidStream(decoded, first.lengthOfHeader)
	case len(stream) >= 13 && stream[0] <= 1 && nsisIsLZMAProps(stream[1:]):
		// Solid LZMA behind a filter flag byte. ame cannot decode the BCJ
		// filter, but no release it targets uses it.
		if stream[0] == 1 {
			return nil, fmt.Errorf("BCJ-filtered NSIS installer is not supported")
		}
		decoded, err := nsisDecompressStream("lzma", stream[1:])
		if err != nil {
			return nil, fmt.Errorf("NSIS solid stream: %w", err)
		}
		header, archive.solidData = nsisSplitSolidStream(decoded, first.lengthOfHeader)
	case len(stream) >= 12 && stream[3]&0x80 != 0:
		// Non-solid: a 4-byte compressed-size prefix, then the stream.
		size := int(binary.LittleEndian.Uint32(stream) & 0x7fffffff)
		if 4+size > len(stream) {
			return nil, fmt.Errorf("header block truncated")
		}
		payload := stream[4 : 4+size]
		content, err := nsisDecompressStream("", payload)
		if err != nil {
			return nil, fmt.Errorf("NSIS header block: %w", err)
		}
		if len(content) < first.lengthOfHeader {
			return nil, fmt.Errorf("decompressed header is %d bytes, expected %d", len(content), first.lengthOfHeader)
		}
		header = content[:first.lengthOfHeader]
		archive.data = after[4+size:]
	default:
		// Solid deflate.
		decoded, err := nsisDecompressStream("deflate", stream)
		if err != nil {
			return nil, fmt.Errorf("NSIS solid stream: %w", err)
		}
		header, archive.solidData = nsisSplitSolidStream(decoded, first.lengthOfHeader)
	}

	blocks := nsisBlockTable(header)
	if !nsisBlocksPlausible(blocks, len(header)) {
		return nil, fmt.Errorf("decompressed data is not an NSIS header")
	}
	archive.entries = nsisEntries(header, blocks)
	stringsOffset := int(blocks[nsisBlockStrings].offset)
	archive.strings = header[stringsOffset:]
	return archive, nil
}

// nsisSplitSolidStream splits a decompressed solid stream into its header and
// file data. The stream begins with the header's 4-byte size; some writers
// omit it.
func nsisSplitSolidStream(stream []byte, expected int) ([]byte, []byte) {
	if len(stream) >= 4 && int(binary.LittleEndian.Uint32(stream)) == expected {
		return stream[4 : 4+expected], stream[4+expected:]
	}
	if len(stream) >= expected {
		return stream[:expected], stream[expected:]
	}
	return nil, nil
}

// extract returns the content of the file whose destination name ends with
// name. Names are matched on the basename, which sidesteps the $INSTDIR and
// $PLUGINSDIR variables the installer script prefixes.
func (a *nsisArchive) extract(name string) ([]byte, error) {
	for _, entry := range a.entries {
		if entry.which != nsisOpExtractFile {
			continue
		}
		if a.string(entry.offsets[1]) != strings.ToLower(name) {
			continue
		}
		source := a.solidData
		if source == nil {
			source = a.data
		}
		offset := int(uint32(entry.offsets[2]))
		if offset < 0 || offset+4 > len(source) {
			return nil, fmt.Errorf("file %q data offset %d outside the data block", name, offset)
		}
		prefix := binary.LittleEndian.Uint32(source[offset:])
		size := int(prefix & 0x7fffffff)
		payload := source[offset+4:]
		if size > len(payload) {
			return nil, fmt.Errorf("file %q data truncated", name)
		}
		payload = payload[:size]
		if prefix&0x80000000 == 0 {
			return payload, nil
		}
		content, err := nsisDecompressStream("", payload)
		if err != nil {
			return nil, fmt.Errorf("file %q: %w", name, err)
		}
		return content, nil
	}
	return nil, fmt.Errorf("file %q not found in installer", name)
}

// string reads the literal text of the string at the given table offset.
// Variable references are dropped: only the literal characters of a name
// are needed for basename matching. Unicode tables index by character, so
// the byte position is twice the offset (7-Zip NsisIn.cpp: _stringsPos +
// pos * 2).
func (a *nsisArchive) string(offset int32) string {
	if offset < 0 {
		return ""
	}
	table := a.strings
	unicode := nsisStringTableUnicode(a.strings)
	pos := int(offset)
	if unicode {
		pos *= 2
	}
	if pos >= len(table) {
		return ""
	}
	table = table[pos:]
	var b strings.Builder
	if unicode {
		for i := 0; i+2 <= len(table); i += 2 {
			ch := binary.LittleEndian.Uint16(table[i:])
			if ch == 0 {
				break
			}
			switch {
			case ch >= 0x0001 && ch <= 0x0003: // lang, shell, var: skip the argument
				i += 2
			case ch == 0x0004: // skip: the next unit is a literal
				if i+4 <= len(table) {
					b.WriteRune(rune(binary.LittleEndian.Uint16(table[i+2:])))
				}
				i += 2
			default:
				b.WriteRune(rune(ch))
			}
		}
	} else {
		for i := 0; i < len(table); i++ {
			ch := table[i]
			if ch == 0 {
				break
			}
			switch {
			case ch == 0x03: // NSIS 3 var: two argument bytes
				i += 2
			case ch == 0x04: // skip: the next byte is a literal
				if i+1 < len(table) {
					b.WriteByte(table[i+1])
				}
				i++
			case ch < 0x20:
				// Other control codes (NSIS 2 variables): drop.
			default:
				b.WriteByte(ch)
			}
		}
	}
	return baseName(b.String())
}

// nsisStringTableUnicode reports whether the string table is UTF-16LE: an
// empty leading string then occupies two zero bytes.
func nsisStringTableUnicode(table []byte) bool {
	return len(table) >= 2 && table[0] == 0 && table[1] == 0
}

// baseName lowercases the last path segment of a name.
func baseName(name string) string {
	name = strings.ReplaceAll(name, "/", `\`)
	if i := strings.LastIndexByte(name, '\\'); i >= 0 {
		name = name[i+1:]
	}
	return strings.ToLower(strings.TrimSpace(name))
}

// peOverlayStart returns the offset of the data appended after the PE image:
// the end of the last section, or 0 when the file is not a PE.
func peOverlayStart(data []byte) int {
	const sectionHeaderSize = 40
	if len(data) < 0x40 || binary.LittleEndian.Uint16(data) != 0x5a4d { // "MZ"
		return 0
	}
	peOffset := int(binary.LittleEndian.Uint32(data[0x3c:]))
	if peOffset <= 0 || peOffset+24 > len(data) || !bytes.Equal(data[peOffset:peOffset+4], []byte("PE\x00\x00")) {
		return 0
	}
	numSections := int(binary.LittleEndian.Uint16(data[peOffset+6:]))
	optionalSize := int(binary.LittleEndian.Uint16(data[peOffset+20:]))
	table := peOffset + 24 + optionalSize
	if numSections == 0 || table+numSections*sectionHeaderSize > len(data) {
		return 0
	}
	end := 0
	for i := 0; i < numSections; i++ {
		header := data[table+i*sectionHeaderSize:]
		rawSize := int(binary.LittleEndian.Uint32(header[16:]))
		rawPointer := int(binary.LittleEndian.Uint32(header[20:]))
		if sectionEnd := rawPointer + rawSize; sectionEnd > end && sectionEnd <= len(data) {
			end = sectionEnd
		}
	}
	if end == 0 || end >= len(data) {
		return 0
	}
	return end
}

// nsisFindFirstHeader locates the first header in the overlay. It is written
// at a 512-byte aligned offset, but installers with extra data before the
// archive are scanned byte-wise as well.
func nsisFindFirstHeader(overlay []byte) (int, nsisFirstHeader, bool) {
	for offset := 0; offset+nsisFirstHeaderSize <= len(overlay); offset += 512 {
		if first, ok := nsisParseFirstHeader(overlay[offset:]); ok {
			return offset, first, true
		}
	}
	for offset := 0; offset+nsisFirstHeaderSize <= len(overlay); offset++ {
		if offset%512 == 0 {
			continue
		}
		if first, ok := nsisParseFirstHeader(overlay[offset:]); ok {
			return offset, first, true
		}
	}
	return 0, nsisFirstHeader{}, false
}

func nsisParseFirstHeader(data []byte) (nsisFirstHeader, bool) {
	if len(data) < nsisFirstHeaderSize {
		return nsisFirstHeader{}, false
	}
	first := nsisFirstHeader{
		flags:           binary.LittleEndian.Uint32(data),
		lengthOfHeader:  int(int32(binary.LittleEndian.Uint32(data[20:]))),
		lengthOfAllData: int(int32(binary.LittleEndian.Uint32(data[24:]))),
	}
	if binary.LittleEndian.Uint32(data[4:8]) != 0xdeadbeef {
		return nsisFirstHeader{}, false
	}
	if string(data[8:20]) != "NullsoftInst" {
		return nsisFirstHeader{}, false
	}
	if first.flags&^0x0f != 0 {
		return nsisFirstHeader{}, false
	}
	if first.lengthOfHeader <= 0 || first.lengthOfAllData < nsisFirstHeaderSize {
		return nsisFirstHeader{}, false
	}
	return first, true
}

// nsisIsLZMAProps matches the LZMA signature 7-Zip checks (NsisIn.cpp):
// properties byte 0x5D (lc=3, lp=0, pb=2), a small dictionary size, and the
// stream's first two bytes.
func nsisIsLZMAProps(p []byte) bool {
	return len(p) >= 7 && p[0] == 0x5D && p[1] == 0x00 && p[2] == 0x00 &&
		p[5] == 0x00 && p[6]&0x80 == 0
}

// nsisDecompressStream decodes one NSIS stream of unknown output size, read
// to its natural end (the LZMA end-of-stream marker or the last deflate
// block). An empty method picks LZMA when the payload carries the standard
// properties, deflate otherwise.
func nsisDecompressStream(method string, payload []byte) ([]byte, error) {
	switch method {
	case "lzma":
		return nsisLzmaStream(payload)
	case "deflate":
		return nsisDeflateStream(payload)
	default:
		if nsisIsLZMAProps(payload) {
			return nsisLzmaStream(payload)
		}
		return nsisDeflateStream(payload)
	}
}

// nsisLzmaStream decodes an NSIS LZMA stream: a properties byte and a
// 4-byte dictionary size, then the data. The synthetic 13-byte .lzma header
// carries the unknown-size sentinel, so the decoder runs to the end-of-stream
// marker.
func nsisLzmaStream(payload []byte) ([]byte, error) {
	if len(payload) < 5 {
		return nil, fmt.Errorf("lzma stream too short")
	}
	header := make([]byte, 13)
	copy(header, payload[:5])
	binary.LittleEndian.PutUint64(header[5:], math.MaxUint64)
	reader, err := lzma.NewReader(io.MultiReader(bytes.NewReader(header), bytes.NewReader(payload[5:])))
	if err != nil {
		return nil, fmt.Errorf("lzma: %w", err)
	}
	data, err := io.ReadAll(io.LimitReader(reader, nsisMaxDecompressed))
	if err != nil {
		return nil, fmt.Errorf("lzma: %w", err)
	}
	if len(data) == nsisMaxDecompressed {
		return nil, fmt.Errorf("lzma output exceeds %d bytes", nsisMaxDecompressed)
	}
	return data, nil
}

// nsisDeflateStream decodes a raw deflate stream.
func nsisDeflateStream(payload []byte) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(payload)), nsisMaxDecompressed))
	if err != nil {
		return nil, fmt.Errorf("deflate: %w", err)
	}
	return data, nil
}

// nsisBlockTable reads the eight block descriptors from the common header.
func nsisBlockTable(header []byte) [nsisBlocksNum]struct{ offset, num uint32 } {
	var blocks [nsisBlocksNum]struct{ offset, num uint32 }
	if len(header) < 4+nsisBlocksNum*8 {
		return blocks
	}
	for i := range blocks {
		blocks[i].offset = binary.LittleEndian.Uint32(header[4+i*8:])
		blocks[i].num = binary.LittleEndian.Uint32(header[4+i*8+4:])
	}
	return blocks
}

// nsisBlocksPlausible validates the block table the way NSIS writes it:
// blocks appear in order, so the string table sits after the entries and
// both are inside the decompressed header. This rejects a misdetected
// compression method before any entry is trusted.
func nsisBlocksPlausible(blocks [nsisBlocksNum]struct{ offset, num uint32 }, headerLen int) bool {
	entries := blocks[nsisBlockEntries]
	strings := blocks[nsisBlockStrings]
	if entries.num == 0 || entries.offset == 0 {
		return false
	}
	entriesEnd := int(entries.offset) + int(entries.num)*nsisEntrySize
	if int(entries.offset) >= headerLen || entriesEnd > headerLen {
		return false
	}
	if strings.offset == 0 || int(strings.offset) < entriesEnd || int(strings.offset) >= headerLen {
		return false
	}
	return true
}

// nsisEntries slices the entry block into instructions.
func nsisEntries(header []byte, blocks [nsisBlocksNum]struct{ offset, num uint32 }) []nsisEntry {
	offset := int(blocks[nsisBlockEntries].offset)
	count := int(blocks[nsisBlockEntries].num)
	maxCount := (len(header) - offset) / nsisEntrySize
	if count > maxCount {
		count = maxCount
	}
	entries := make([]nsisEntry, 0, count)
	for i := 0; i < count; i++ {
		entry := header[offset+i*nsisEntrySize:]
		var e nsisEntry
		e.which = int32(binary.LittleEndian.Uint32(entry))
		for j := 0; j < 6; j++ {
			e.offsets[j] = int32(binary.LittleEndian.Uint32(entry[4+j*4:]))
		}
		entries = append(entries, e)
	}
	return entries
}

// nsisExtractFiles pulls the named files out of an installer image.
func nsisExtractFiles(installer []byte, names ...string) (map[string][]byte, error) {
	archive, err := nsisOpen(installer)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte, len(names))
	for _, name := range names {
		content, err := archive.extract(name)
		if err != nil {
			return nil, err
		}
		files[name] = content
	}
	return files, nil
}
