package skin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

type testObject struct {
	kind uint8
	id   uint32
	body []byte
}

func testBank(bankID uint32, objects ...testObject) []byte {
	var hirc bytes.Buffer
	binary.Write(&hirc, binary.LittleEndian, uint32(len(objects)))
	for _, o := range objects {
		hirc.WriteByte(o.kind)
		binary.Write(&hirc, binary.LittleEndian, uint32(4+len(o.body)))
		binary.Write(&hirc, binary.LittleEndian, o.id)
		hirc.Write(o.body)
	}
	var out bytes.Buffer
	out.WriteString("BKHD")
	binary.Write(&out, binary.LittleEndian, uint32(12))
	binary.Write(&out, binary.LittleEndian, [3]uint32{145, bankID, 0x393e9711})
	out.WriteString("HIRC")
	binary.Write(&out, binary.LittleEndian, uint32(hirc.Len()))
	out.Write(hirc.Bytes())
	return out.Bytes()
}

func actions(ids ...uint32) []byte {
	var b bytes.Buffer
	b.WriteByte(byte(len(ids)))
	for _, id := range ids {
		binary.Write(&b, binary.LittleEndian, id)
	}
	return b.Bytes()
}

// testWAD writes a v3 WAD; zstd chunks when compress is set.
func testWAD(t *testing.T, path string, chunks map[uint64][]byte, compress bool) {
	t.Helper()
	encoder, _ := zstd.NewWriter(nil)
	var hashes []uint64
	for hash := range chunks {
		hashes = append(hashes, hash)
	}
	header := make([]byte, 272)
	copy(header, "RW")
	header[2], header[3] = 3, 4
	binary.LittleEndian.PutUint32(header[268:], uint32(len(hashes)))
	toc := make([]byte, 32*len(hashes))
	var data bytes.Buffer
	offset := 272 + len(toc)
	for i, hash := range hashes {
		raw, kind := chunks[hash], byte(0)
		stored := raw
		if compress {
			stored, kind = encoder.EncodeAll(raw, nil), 3
		}
		e := toc[32*i:]
		binary.LittleEndian.PutUint64(e, hash)
		binary.LittleEndian.PutUint32(e[8:], uint32(offset+data.Len()))
		binary.LittleEndian.PutUint32(e[12:], uint32(len(stored)))
		binary.LittleEndian.PutUint32(e[16:], uint32(len(raw)))
		e[20] = kind
		data.Write(stored)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append(header, toc...), data.Bytes()...), 0644); err != nil {
		t.Fatal(err)
	}
}

// The shipped bank: built on another locale, events renamed (0xbbbb0001).
// The player's bank: own media, original event IDs, different object order.
var (
	shippedBank = testBank(7,
		testObject{2, 100, []byte("media-vi")},
		testObject{3, 200, []byte("play 100")},
		testObject{4, 0xbbbb0001, actions(200)},
		testObject{3, 201, []byte("play 100 again")},
		testObject{4, 0xaaaa0002, actions(201)},
	)
	userBank = testBank(7,
		testObject{3, 201, []byte("play 100 again")},
		testObject{2, 100, []byte("media-pt")},
		testObject{3, 200, []byte("play 100")},
		testObject{4, 0xaaaa0001, actions(200)},
		testObject{4, 0xaaaa0002, actions(201)},
	)
	wantBank = testBank(7,
		testObject{3, 201, []byte("play 100 again")},
		testObject{2, 100, []byte("media-pt")},
		testObject{3, 200, []byte("play 100")},
		testObject{4, 0xbbbb0001, actions(200)},
		testObject{4, 0xaaaa0002, actions(201)},
	)
)

func TestTransplantKeepsOwnMediaAndTakesRenamedEvents(t *testing.T) {
	out, renamed, err := transplantBank(shippedBank, userBank)
	if err != nil {
		t.Fatal(err)
	}
	if renamed != 1 || !bytes.Equal(out, wantBank) {
		t.Fatalf("renamed=%d\n got %x\nwant %x", renamed, out, wantBank)
	}
}

func TestTransplantPairsEventsSharingActions(t *testing.T) {
	shipped := testBank(7, testObject{3, 200, nil}, testObject{4, 0xbbbb0001, actions(200)}, testObject{4, 0xaaaa0002, actions(200)})
	user := testBank(7, testObject{3, 200, nil}, testObject{4, 0xaaaa0001, actions(200)}, testObject{4, 0xaaaa0002, actions(200)})
	out, _, err := transplantBank(shipped, user)
	if err != nil || !bytes.Equal(out, shipped) {
		t.Fatalf("got %x err=%v", out, err)
	}
}

func TestTransplantRefusesDifferentBanks(t *testing.T) {
	for name, user := range map[string][]byte{
		"bank id":     testBank(8, testObject{2, 100, []byte("media-pt")}, testObject{3, 200, []byte("play 100")}, testObject{4, 1, actions(200)}, testObject{3, 201, []byte("play 100 again")}, testObject{4, 2, actions(201)}),
		"extra sound": testBank(7, testObject{2, 101, nil}, testObject{2, 100, []byte("media-pt")}, testObject{3, 200, []byte("play 100")}, testObject{4, 1, actions(200)}, testObject{3, 201, []byte("play 100 again")}, testObject{4, 2, actions(201)}),
		"actions":     testBank(7, testObject{2, 100, []byte("media-pt")}, testObject{3, 200, []byte("play 100")}, testObject{4, 1, actions(201)}, testObject{3, 201, []byte("play 100 again")}, testObject{4, 2, actions(201)}),
		"truncated":   userBank[:len(userBank)-3],
	} {
		if _, _, err := transplantBank(shippedBank, user); err == nil {
			t.Errorf("%s: transplant accepted a different bank", name)
		}
	}
}

func TestLocalizeVoiceWADsRebuildsForInstalledLocale(t *testing.T) {
	gameDir, modDir := t.TempDir(), t.TempDir()
	const bankHash, otherHash = 0xdcb5928df1ce6fe3, 0x0123456789abcdef
	champions := filepath.Join(gameDir, "DATA", "FINAL", "Champions")
	testWAD(t, filepath.Join(champions, "Ahri.pt_BR.wad.client"), map[uint64][]byte{bankHash: userBank}, true)
	testWAD(t, filepath.Join(champions, "Ahri.wad.client"), map[uint64][]byte{1: []byte("main")}, true)
	testWAD(t, filepath.Join(champions, "Zed.ko_KR.wad.client"), map[uint64][]byte{bankHash: userBank}, true)
	// A second shipped chunk the installed locale lacks stays as shipped.
	other := []byte("not a bank")
	testWAD(t, filepath.Join(modDir, "WAD", "Ahri.vi_VN.wad.client"), map[uint64][]byte{bankHash: shippedBank, otherHash: other}, true)

	localizeVoiceWADs(modDir, gameDir)

	if _, err := os.Stat(filepath.Join(modDir, "WAD", "Ahri.vi_VN.wad.client")); !os.IsNotExist(err) {
		t.Fatal("shipped locale WAD kept although the player has another locale")
	}
	dir := filepath.Join(modDir, "WAD", "Ahri.pt_BR.wad.client")
	got, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%016x.bnk", uint64(bankHash))))
	if err != nil || !bytes.Equal(got, wantBank) {
		t.Fatalf("rebuilt bank %x err=%v", got, err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%016x.bnk", uint64(otherHash)))); !bytes.Equal(got, other) {
		t.Fatal("unmatched chunk not kept as shipped")
	}
	if _, err := os.Stat(filepath.Join(modDir, "WAD", "Zed.ko_KR.wad.client")); !os.IsNotExist(err) {
		t.Fatal("rebuilt a WAD for another champion")
	}
}

func TestLocalizeVoiceWADsLeavesMatchingLocale(t *testing.T) {
	gameDir, modDir := t.TempDir(), t.TempDir()
	champions := filepath.Join(gameDir, "DATA", "FINAL", "Champions")
	testWAD(t, filepath.Join(champions, "Ahri.vi_VN.wad.client"), map[uint64][]byte{5: userBank}, true)
	shipped := filepath.Join(modDir, "WAD", "Ahri.vi_VN.wad.client")
	testWAD(t, shipped, map[uint64][]byte{5: shippedBank}, false)
	before, _ := os.ReadFile(shipped)

	localizeVoiceWADs(modDir, gameDir)

	after, err := os.ReadFile(shipped)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("shipped WAD changed for a player on the same locale")
	}
	entries, _ := os.ReadDir(filepath.Join(modDir, "WAD"))
	if len(entries) != 1 {
		t.Fatalf("unexpected WAD entries: %v", entries)
	}
}
