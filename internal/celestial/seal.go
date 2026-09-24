package celestial

import (
	"crypto/sha512"
	"encoding/hex"
)

// IndexSeal constructs the observed single-file overlay seal. It covers the
// encoded wad.idx bytes, not the decoded CHK3 payload or its referenced WADs.
// Overlays containing additional files need a separately verified manifest
// ordering/selection policy; this function must not be used to omit them.
func IndexSeal(encodedIndex []byte) ([]byte, error) {
	if _, err := DecodeIndex(encodedIndex); err != nil {
		return nil, err
	}
	digest := sha512.Sum512(encodedIndex)
	seal := make([]byte, 0, 137)
	seal = append(seal, "wad.idx\t"...)
	seal = append(seal, hex.EncodeToString(digest[:])...)
	seal = append(seal, '\n')
	return seal, nil
}

// SealRoot follows the inspected producer: the first 32 bytes of SHA-512 over
// the exact seal bytes. This is deliberately NOT SHA-512/256. The derivation
// is established by static analysis; no live host authentication is implied.
// A root is an integrity value, not a signature or a valid session record.
func SealRoot(seal []byte) [32]byte {
	digest := sha512.Sum512(seal)
	var root [32]byte
	copy(root[:], digest[:32])
	return root
}
