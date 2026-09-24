package celestial

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func fixtureWAD(hash uint64, payload byte) []byte {
	b := make([]byte, 305)
	copy(b, []byte{'R', 'W', 3, 4})
	binary.LittleEndian.PutUint32(b[268:], 1)
	binary.LittleEndian.PutUint64(b[272:], hash)
	binary.LittleEndian.PutUint32(b[280:], 304)
	binary.LittleEndian.PutUint32(b[284:], 1)
	binary.LittleEndian.PutUint32(b[288:], 1)
	b[292] = 3
	binary.LittleEndian.PutUint64(b[296:], uint64(payload))
	b[304] = payload
	return b
}

func TestBuildIndexPriorityAndRouting(t *testing.T) {
	dir := t.TempDir()
	first, last := filepath.Join(dir, "first.wad.client"), filepath.Join(dir, "last.wad.client")
	for name, contents := range map[string][]byte{first: fixtureWAD(51, 1), last: fixtureWAD(51, 2)} {
		if err := os.WriteFile(name, contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	a, b := "data/final/a.wad.client", "data/final/b.wad.client"
	index, err := BuildIndex([]WADInput{{first, []string{b, a}}, {last, []string{a}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Targets) != 2 || index.Targets[0].Name != a || index.Targets[1].Name != b {
		t.Fatal("incorrect destination routing")
	}
	if index.Targets[0].Chunks[0].Source != 1 || index.Targets[0].Chunks[0].Checksum != 2 {
		t.Fatal("later mod did not win conflict")
	}
	if index.Targets[1].Chunks[0].Source != 0 {
		t.Fatal("override leaked into unrelated target")
	}
	if _, err := BuildIndex([]WADInput{{first, []string{"data/final/../bad.wad.client"}}}); err == nil {
		t.Fatal("accepted escaping target")
	}
}

func TestWADDirectoryRejectsUnsupportedAndInvalid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]byte)
	}{
		{"subchunks", func(b []byte) { b[292] = 4 }},
		{"metadata", func(b []byte) { b[294] = 1 }},
		{"bad offset", func(b []byte) { binary.LittleEndian.PutUint32(b[280:], 999) }},
		{"header overlap", func(b []byte) { binary.LittleEndian.PutUint32(b[280:], 0) }},
		{"bad count", func(b []byte) { binary.LittleEndian.PutUint32(b[268:], 999) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := fixtureWAD(51, 1)
			tc.mutate(b)
			if _, err := readWADDirectory(bytes.NewReader(b), int64(len(b))); err == nil {
				t.Fatal("accepted invalid directory")
			}
		})
	}
}

func TestBuildInstalledIndex(t *testing.T) {
	name := os.Getenv("AME_TEST_CELESTIAL_INDEX")
	if name == "" {
		t.Skip("set AME_TEST_CELESTIAL_INDEX for installed Ahri fixture")
	}
	encoded, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := DecodeIndex(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixture.Sources) != 1 || len(fixture.Targets) != 1 {
		t.Fatal("this comparison requires the one-source Ahri fixture")
	}
	// Only routing/path information comes from the fixture. All chunk fields
	// and ordering are generated independently from the source WAD directory.
	built, err := BuildIndex([]WADInput{{fixture.Sources[0], []string{fixture.Targets[0].Name}}})
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := EncodeIndex(built)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, rebuilt) {
		t.Fatal("independently built index differs from Celestial fixture")
	}
	t.Logf("Independently built %d-byte index matches Celestial exactly", len(rebuilt))
}
