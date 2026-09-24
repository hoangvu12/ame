package celestial

import (
	"bytes"
	"encoding/binary"
	"os"
	"reflect"
	"testing"
)

func TestIndexRoundTrip(t *testing.T) {
	want := &Index{Sources: []string{`C:\mods\skin-水-😀.wad.client`}, Targets: []Target{{
		Name: "data/final/champions/ahri.wad.client",
		Chunks: []Chunk{{Hash: 91, Offset: 1 << 33, CompressedSize: 12, Size: 42,
			Checksum: 83, Trailer: [4]byte{3, 7, 21, 4}}},
	}}}
	encoded, err := EncodeIndex(want)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), encoded...)
	got, err := DecodeIndex(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip changed index: %#v", got)
	}
	if !bytes.Equal(before, encoded) {
		t.Fatal("decode mutated input")
	}
	for length := 0; length < len(encoded); length++ {
		if _, err := DecodeIndex(encoded[:length]); err == nil {
			t.Fatalf("accepted truncation at %d", length)
		}
	}
}

func TestIndexRejectsMalformedInput(t *testing.T) {
	for _, tc := range []struct {
		name  string
		plain []byte
	}{
		{"huge source count", []byte{'C', 'H', 'K', '3', 255, 255, 255, 255, 0, 0, 0, 0}},
		{"invalid surrogate", []byte{'C', 'H', 'K', '3', 1, 0, 0, 0, 1, 0, 0, 216, 0, 0, 0, 0}},
		{"extra bytes", []byte{'C', 'H', 'K', '3', 0, 0, 0, 0, 0, 0, 0, 0, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transform(tc.plain)
			if _, err := DecodeIndex(tc.plain); err == nil {
				t.Fatal("accepted malformed input")
			}
		})
	}
	for _, chunk := range []Chunk{{Source: 1}, {Offset: ^uint64(0), CompressedSize: 1}} {
		if _, err := EncodeIndex(&Index{Sources: []string{"source"}, Targets: []Target{{Name: "target", Chunks: []Chunk{chunk}}}}); err == nil {
			t.Fatal("accepted invalid chunk")
		}
	}
}

// This opt-in test reads an existing vendor artifact, never rewrites it, and
// independently compares core chunk fields against the referenced WAD tables.
func TestInstalledIndex(t *testing.T) {
	path := os.Getenv("AME_TEST_CELESTIAL_INDEX")
	if path == "" {
		t.Skip("set AME_TEST_CELESTIAL_INDEX for read-only installed-file validation")
	}
	encoded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	index, err := DecodeIndex(encoded)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := EncodeIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, rebuilt) {
		t.Fatal("rebuilt index differs from installed artifact")
	}
	wads := make([][]byte, len(index.Sources))
	for i, source := range index.Sources {
		wads[i], err = os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
	}
	checked := 0
	for _, target := range index.Targets {
		for _, c := range target.Chunks {
			wad := wads[c.Source]
			if len(wad) < 272 || string(wad[:2]) != "RW" || wad[2] != 3 {
				t.Fatal("unsupported source WAD")
			}
			n := binary.LittleEndian.Uint32(wad[268:272])
			if uint64(n)*32+272 > uint64(len(wad)) {
				t.Fatal("truncated WAD directory")
			}
			found := false
			for i := uint32(0); i < n; i++ {
				entry := wad[272+int(i)*32 : 272+int(i+1)*32]
				if binary.LittleEndian.Uint64(entry) != c.Hash {
					continue
				}
				found = true
				if uint64(binary.LittleEndian.Uint32(entry[8:])) != c.Offset ||
					binary.LittleEndian.Uint32(entry[12:]) != c.CompressedSize ||
					binary.LittleEndian.Uint32(entry[16:]) != c.Size ||
					binary.LittleEndian.Uint64(entry[24:]) != c.Checksum || entry[20]&15 != c.Trailer[0] {
					t.Fatal("index chunk differs from source WAD")
				}
				if c.Offset+uint64(c.CompressedSize) > uint64(len(wad)) {
					t.Fatal("chunk outside source WAD")
				}
				break
			}
			if !found {
				t.Fatal("index hash missing from source WAD")
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("fixture has no chunks to validate")
	}
	t.Logf("Exact %d-byte index round trip; %d source WAD chunk records verified", len(encoded), checked)
}
