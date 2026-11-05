package ltk

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

// --- fixture builders ---

// nsisFixture builds a synthetic NSIS installer. solid selects the solid
// framing, method picks the compression ("lzma", "deflate", or "" for
// stored), and files maps destination names to contents.
func nsisFixture(t *testing.T, solid bool, method string, files map[string]string, names []string) []byte {
	t.Helper()

	// Common header: flags, then eight block descriptors.
	entriesOffset := 4 + nsisBlocksNum*8
	entries := make([]byte, 0, len(names)*nsisEntrySize)
	// Unicode string table: an empty leading string, then each name.
	stringsTable := []byte{0, 0}
	nameOffsets := make(map[string]int)
	for _, name := range names {
		nameOffsets[name] = len(stringsTable) / 2
		for _, r := range name {
			stringsTable = append(stringsTable, byte(r), 0)
		}
		stringsTable = append(stringsTable, 0, 0)
	}

	// Solid: one stream holds [4-byte header size][header][file data...];
	// file offsets index the decompressed region after the header.
	// Non-solid: each file is its own [4-byte prefix][payload] block, and the
	// offsets index the raw region after the header block.
	var dataRegion []byte
	dataOffsets := make(map[string]int)
	for _, name := range names {
		content := []byte(files[name])
		dataOffsets[name] = len(dataRegion)
		if solid {
			// Files inside the solid stream are stored uncompressed.
			var prefix [4]byte
			binary.LittleEndian.PutUint32(prefix[:], uint32(len(content)))
			dataRegion = append(dataRegion, prefix[:]...)
			dataRegion = append(dataRegion, content...)
		} else {
			compressed := nsisCompress(method, content)
			var prefix [4]byte
			if len(compressed) < len(content) {
				binary.LittleEndian.PutUint32(prefix[:], uint32(len(compressed))|0x80000000)
				dataRegion = append(dataRegion, prefix[:]...)
				dataRegion = append(dataRegion, compressed...)
			} else {
				binary.LittleEndian.PutUint32(prefix[:], uint32(len(content)))
				dataRegion = append(dataRegion, prefix[:]...)
				dataRegion = append(dataRegion, content...)
			}
		}
	}

	for _, name := range names {
		entry := make([]byte, nsisEntrySize)
		binary.LittleEndian.PutUint32(entry[0:], uint32(nsisOpExtractFile))
		binary.LittleEndian.PutUint32(entry[4:], 8)                         // overwrite flags
		binary.LittleEndian.PutUint32(entry[8:], uint32(nameOffsets[name])) // name (characters)
		binary.LittleEndian.PutUint32(entry[12:], uint32(dataOffsets[name]))
		binary.LittleEndian.PutUint32(entry[16:], 0) // date low
		binary.LittleEndian.PutUint32(entry[20:], 0) // date high
		binary.LittleEndian.PutUint32(entry[24:], 0) // error text
		entries = append(entries, entry...)
	}

	stringsOffset := entriesOffset + len(entries)
	header := make([]byte, entriesOffset)
	binary.LittleEndian.PutUint32(header[nsisBlockEntries*8+4:], uint32(entriesOffset))
	binary.LittleEndian.PutUint32(header[nsisBlockEntries*8+8:], uint32(len(names)))
	binary.LittleEndian.PutUint32(header[nsisBlockStrings*8+4:], uint32(stringsOffset))
	binary.LittleEndian.PutUint32(header[nsisBlockStrings*8+8:], uint32(len(stringsTable)))
	header = append(header, entries...)
	header = append(header, stringsTable...)

	// Assemble the payload that follows the 28-byte first header.
	var after []byte
	if solid {
		stream := make([]byte, 4, 4+len(header)+len(dataRegion))
		binary.LittleEndian.PutUint32(stream, uint32(len(header)))
		stream = append(stream, header...)
		stream = append(stream, dataRegion...)
		after = nsisCompress(method, stream)
	} else {
		switch method {
		case "":
			// Stored header block: the prefix equals the header size.
			var prefix [4]byte
			binary.LittleEndian.PutUint32(prefix[:], uint32(len(header)))
			after = append(after, prefix[:]...)
			after = append(after, header...)
		default:
			compressed := nsisCompress(method, header)
			var prefix [4]byte
			binary.LittleEndian.PutUint32(prefix[:], uint32(len(compressed))|0x80000000)
			after = append(after, prefix[:]...)
			after = append(after, compressed...)
		}
		after = append(after, dataRegion...)
	}

	// A 512-byte stub before the archive, the first header, the payload. No
	// CRC, so the stream ends exactly at the archive bounds.
	firstHeader := make([]byte, nsisFirstHeaderSize)
	binary.LittleEndian.PutUint32(firstHeader[0:], nsisFlagNoCRC)
	binary.LittleEndian.PutUint32(firstHeader[4:], 0xdeadbeef)
	copy(firstHeader[8:20], "NullsoftInst")
	binary.LittleEndian.PutUint32(firstHeader[20:], uint32(len(header)))
	binary.LittleEndian.PutUint32(firstHeader[24:], uint32(nsisFirstHeaderSize+len(after)))

	installer := make([]byte, 512)
	copy(installer[500:], "stub!")
	installer = append(installer, firstHeader...)
	installer = append(installer, after...)
	return installer
}

// nsisCompress compresses data the way NSIS frames it: LZMA writes the
// properties and dictionary size followed by the stream (the .lzma header's
// explicit size field is dropped); deflate writes the raw stream.
func nsisCompress(method string, data []byte) []byte {
	switch method {
	case "lzma":
		var buf bytes.Buffer
		writer, err := lzma.NewWriter(&buf)
		if err != nil {
			panic(err)
		}
		if _, err := writer.Write(data); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
		encoded := buf.Bytes()
		return append(encoded[:5:5], encoded[13:]...)
	case "deflate":
		var buf bytes.Buffer
		writer, _ := flate.NewWriter(&buf, flate.DefaultCompression)
		if _, err := writer.Write(data); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
		return buf.Bytes()
	default:
		return data
	}
}

// --- tests ---

func TestNsisExtractSolidLzma(t *testing.T) {
	files := map[string]string{
		"app.exe":  strings.Repeat("manager payload ", 4096),
		"host.exe": "host payload",
	}
	installer := nsisFixture(t, true, "lzma", files, []string{"app.exe", "host.exe"})
	extracted, err := nsisExtractFiles(installer, "app.exe", "host.exe")
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		if string(extracted[name]) != want {
			t.Fatalf("%s: got %d bytes, want %d", name, len(extracted[name]), len(want))
		}
	}
}

func TestNsisExtractSolidDeflate(t *testing.T) {
	files := map[string]string{
		"app.exe":  strings.Repeat("deflate solid ", 2048),
		"host.exe": "host payload",
	}
	installer := nsisFixture(t, true, "deflate", files, []string{"app.exe", "host.exe"})
	extracted, err := nsisExtractFiles(installer, "app.exe", "host.exe")
	if err != nil {
		t.Fatal(err)
	}
	if string(extracted["app.exe"]) != files["app.exe"] {
		t.Fatal("solid deflate content mismatch")
	}
}

func TestNsisExtractNonSolidStored(t *testing.T) {
	files := map[string]string{
		"app.exe":  "stored payload",
		"host.exe": "other payload",
	}
	installer := nsisFixture(t, false, "", files, []string{"app.exe", "host.exe"})
	extracted, err := nsisExtractFiles(installer, "app.exe", "host.exe")
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		if string(extracted[name]) != want {
			t.Fatalf("%s: %q", name, extracted[name])
		}
	}
}

func TestNsisExtractNonSolidDeflate(t *testing.T) {
	files := map[string]string{
		"app.exe":  strings.Repeat("non-solid deflate ", 1024),
		"host.exe": "host payload",
	}
	installer := nsisFixture(t, false, "deflate", files, []string{"app.exe", "host.exe"})
	extracted, err := nsisExtractFiles(installer, "app.exe", "host.exe")
	if err != nil {
		t.Fatal(err)
	}
	if string(extracted["app.exe"]) != files["app.exe"] {
		t.Fatal("non-solid deflate content mismatch")
	}
}

func TestNsisExtractMissingFile(t *testing.T) {
	installer := nsisFixture(t, true, "lzma", map[string]string{"app.exe": "x"}, []string{"app.exe"})
	if _, err := nsisExtractFiles(installer, "other.exe"); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestNsisRejectsGarbage(t *testing.T) {
	if _, err := nsisExtractFiles([]byte("not an installer at all"), "app.exe"); err == nil {
		t.Fatal("garbage accepted as an installer")
	}
	installer := nsisFixture(t, true, "lzma", map[string]string{"app.exe": "x"}, []string{"app.exe"})
	truncated := installer[:len(installer)-10]
	if _, err := nsisExtractFiles(truncated, "app.exe"); err == nil {
		t.Fatal("truncated archive accepted")
	}
	corrupt := append([]byte{}, installer...)
	corrupt[len(corrupt)-3] ^= 0xff
	if _, err := nsisExtractFiles(corrupt, "app.exe"); err == nil {
		t.Fatal("corrupt stream accepted")
	}
}

func TestNsisStrings(t *testing.T) {
	// A unicode string table: empty string, "app.exe", then a name that
	// begins with a variable reference.
	table := []byte{0, 0}
	putUTF16 := func(s string) {
		for _, r := range s {
			table = append(table, byte(r), 0)
		}
		table = append(table, 0, 0)
	}
	putUTF16("app.exe")
	varRef := len(table) / 2
	table = append(table, 0x03, 0x00, 0x02, 0x00) // var reference code with its argument
	putUTF16("\\dir\\name.exe")

	archive := &nsisArchive{strings: table}
	if got := archive.string(1); got != "app.exe" {
		t.Fatalf("unicode name = %q", got)
	}
	// The variable reference is dropped; the literal tail keeps its name.
	if got := archive.string(int32(varRef)); got != "name.exe" {
		t.Fatalf("name with variable = %q", got)
	}

	// ANSI table: offsets are byte positions.
	ansi := []byte{0}
	ansi = append(ansi, []byte("app.exe")...)
	ansi = append(ansi, 0)
	ansi = append(ansi, 0x03, 0x80, 0x80) // var: skip the two argument bytes
	ansi = append(ansi, []byte("\\name.exe")...)
	ansi = append(ansi, 0)
	ansiArchive := &nsisArchive{strings: ansi}
	if got := ansiArchive.string(1); got != "app.exe" {
		t.Fatalf("ansi name = %q", got)
	}
	if got := ansiArchive.string(int32(9)); got != "name.exe" {
		t.Fatalf("ansi name with variable = %q", got)
	}
}

func TestNsisBaseName(t *testing.T) {
	cases := map[string]string{
		"app.exe":            "app.exe",
		`$INSTDIR\app.exe`:   "app.exe",
		`C:\dir\sub\App.EXE`: "app.exe",
		"dir/name.exe":       "name.exe",
		`  spaced.exe `:      "spaced.exe",
	}
	for name, want := range cases {
		if got := baseName(name); got != want {
			t.Fatalf("baseName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestNsisLzmaStreamRoundTrip(t *testing.T) {
	data := []byte(strings.Repeat("round trip payload ", 1000))
	compressed := nsisCompress("lzma", data)
	decoded, err := nsisLzmaStream(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, data) {
		t.Fatalf("round trip mismatch: %d bytes vs %d", len(decoded), len(data))
	}
}
