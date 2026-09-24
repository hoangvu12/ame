package celestial

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestSealRootUsesTruncatedSHA512(t *testing.T) {
	// SHA-512("abc") first 32 bytes; SHA-512/256 uses different initial state.
	want, err := hex.DecodeString("ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a")
	if err != nil {
		t.Fatal(err)
	}
	got := SealRoot([]byte("abc"))
	if !bytes.Equal(got[:], want) {
		t.Fatal("incorrect root derivation")
	}
	if _, err := IndexSeal([]byte("invalid")); err == nil {
		t.Fatal("sealed an invalid index")
	}
}

func TestInstalledSeal(t *testing.T) {
	name := os.Getenv("AME_TEST_CELESTIAL_INDEX")
	if name == "" {
		t.Skip("set AME_TEST_CELESTIAL_INDEX for installed-file validation")
	}
	encoded, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(filepath.Dir(name), "celestial-overlay.seal"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := IndexSeal(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("generated seal differs from installed single-file fixture")
	}
	t.Logf("Generated %d-byte seal matches Celestial exactly; live session root not captured", len(got))
}
