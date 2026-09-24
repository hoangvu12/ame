package celestial

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverTargets(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "DATA", "FINAL", "Champions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"A.wad.client", "B.wad.client"} {
		if err := os.WriteFile(filepath.Join(dir, name), fixtureWAD(51, 1), 0600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := DiscoverTargets(root, []uint64{51, 52})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"data/final/champions/a.wad.client", "data/final/champions/b.wad.client"}
	if !reflect.DeepEqual(got[51], want) {
		t.Fatalf("cross-WAD holders: %v", got[51])
	}
	if len(got[52]) != 0 {
		t.Fatal("invented destination for missing hash")
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.wad.client"), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := DiscoverTargets(root, []uint64{51}); err == nil {
		t.Fatal("returned partial routing for corrupt install")
	}
}

func TestInstalledRouting(t *testing.T) {
	game, indexPath := os.Getenv("AME_TEST_GAME_DIR"), os.Getenv("AME_TEST_CELESTIAL_INDEX")
	if game == "" || indexPath == "" {
		t.Skip("set installed-game and index fixture environment variables")
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	index, err := DecodeIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	var hashes []uint64
	for _, target := range index.Targets {
		for _, chunk := range target.Chunks {
			hashes = append(hashes, chunk.Hash)
		}
	}
	holders, err := DiscoverTargets(game, hashes)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range index.Targets {
		for _, chunk := range target.Chunks {
			found := false
			for _, holder := range holders[chunk.Hash] {
				if holder == target.Name {
					found = true
				}
			}
			if !found {
				t.Fatalf("fixture target is not a game holder for %x", chunk.Hash)
			}
		}
	}
	t.Logf("Verified installed-game routing for %d fixture chunk hashes", len(hashes))
}
