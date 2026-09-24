// Package celestial contains experimental format support for the inspected
// Celestial runtime. It is not yet connected to AME's overlay backend.
package celestial

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

const MaxIndexBytes = 64 * 1024 * 1024

type Index struct {
	Sources []string
	Targets []Target
}

type Target struct {
	Name   string
	Chunks []Chunk
}

type Chunk struct {
	Hash           uint64
	Source         uint32
	Offset         uint64
	CompressedSize uint32
	Size           uint32
	Checksum       uint64
	// Preserve the four trailing bytes verbatim. Their complete semantics
	// have not been verified; they are not a copied WAD duplicate flag.
	Trailer [4]byte
}

// transform applies the inspected runtime's reversible index byte transform.
// This is not authentication and does not produce a runtime session record.
func transform(data []byte) {
	state := uint64(0x4f7be74b9041d291)
	for off := 0; off < len(data); off += 8 {
		state += 0x9e3779b97f4a7c15
		z := (state ^ (state >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		z ^= z >> 31
		for j := 0; j < 8 && off+j < len(data); j++ {
			data[off+j] ^= byte(z >> (8 * j))
		}
	}
}

// DecodeIndex decodes bytes only: it never opens paths named by the index.
func DecodeIndex(encoded []byte) (*Index, error) {
	if len(encoded) < 12 || len(encoded) > MaxIndexBytes {
		return nil, fmt.Errorf("invalid index size: %d", len(encoded))
	}
	data := append([]byte(nil), encoded...)
	transform(data)
	if string(data[:4]) != "CHK3" {
		return nil, fmt.Errorf("unsupported index magic")
	}
	r := bytes.NewReader(data[4:])
	read := func(value interface{}) error { return binary.Read(r, binary.LittleEndian, value) }
	var count uint32
	if err := read(&count); err != nil {
		return nil, err
	}
	if uint64(count)*2+4 > uint64(r.Len()) {
		return nil, io.ErrUnexpectedEOF
	}
	index := &Index{}
	for n := uint32(0); n < count; n++ {
		var length uint16
		if err := read(&length); err != nil {
			return nil, err
		}
		if int(length)*2 > r.Len() {
			return nil, io.ErrUnexpectedEOF
		}
		units := make([]uint16, length)
		if err := read(units); err != nil {
			return nil, err
		}
		text := string(utf16.Decode(units))
		// Decode replaces invalid surrogates. Reject them instead of silently
		// changing the path during a later encode.
		if !equalUnits(units, utf16.Encode([]rune(text))) {
			return nil, fmt.Errorf("invalid UTF-16 source path")
		}
		index.Sources = append(index.Sources, text)
	}
	if err := read(&count); err != nil {
		return nil, err
	}
	if uint64(count)*6 > uint64(r.Len()) {
		return nil, io.ErrUnexpectedEOF
	}
	for n := uint32(0); n < count; n++ {
		var length uint16
		if err := read(&length); err != nil {
			return nil, err
		}
		if int(length) > r.Len() {
			return nil, io.ErrUnexpectedEOF
		}
		name := make([]byte, length)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}
		if !utf8.Valid(name) {
			return nil, fmt.Errorf("invalid UTF-8 target name")
		}
		var chunks uint32
		if err := read(&chunks); err != nil {
			return nil, err
		}
		if uint64(chunks)*40 > uint64(r.Len()) {
			return nil, io.ErrUnexpectedEOF
		}
		target := Target{Name: string(name)}
		for i := uint32(0); i < chunks; i++ {
			var chunk Chunk
			if err := read(&chunk); err != nil {
				return nil, err
			}
			if uint64(chunk.Source) >= uint64(len(index.Sources)) {
				return nil, fmt.Errorf("source index out of range")
			}
			if chunk.Offset+uint64(chunk.CompressedSize) < chunk.Offset {
				return nil, fmt.Errorf("chunk offset overflow")
			}
			target.Chunks = append(target.Chunks, chunk)
		}
		index.Targets = append(index.Targets, target)
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("unexpected trailing index bytes: %d", r.Len())
	}
	return index, nil
}

func equalUnits(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EncodeIndex preserves caller-supplied order and trailer bytes. It does not
// select winners for conflicting mods or infer unverified trailer semantics.
func EncodeIndex(index *Index) ([]byte, error) {
	if index == nil {
		return nil, fmt.Errorf("nil index")
	}
	// Bound the full encoded size before allocating the output buffer.
	size := uint64(12)
	for _, source := range index.Sources {
		if !utf8.ValidString(source) {
			return nil, fmt.Errorf("invalid source path")
		}
		length := len(utf16.Encode([]rune(source)))
		if length > 65535 {
			return nil, fmt.Errorf("source path too long")
		}
		size += uint64(2 + length*2)
	}
	for _, target := range index.Targets {
		if !utf8.ValidString(target.Name) || len(target.Name) > 65535 {
			return nil, fmt.Errorf("invalid target name")
		}
		size += uint64(6+len(target.Name)) + uint64(len(target.Chunks))*40
	}
	if size > MaxIndexBytes {
		return nil, fmt.Errorf("index exceeds 64 MiB")
	}
	var out bytes.Buffer
	out.Grow(int(size))
	out.WriteString("CHK3")
	write := func(value interface{}) { _ = binary.Write(&out, binary.LittleEndian, value) }
	write(uint32(len(index.Sources)))
	for _, source := range index.Sources {
		units := utf16.Encode([]rune(source))
		write(uint16(len(units)))
		write(units)
	}
	write(uint32(len(index.Targets)))
	for _, target := range index.Targets {
		write(uint16(len(target.Name)))
		out.WriteString(target.Name)
		write(uint32(len(target.Chunks)))
		for _, chunk := range target.Chunks {
			if uint64(chunk.Source) >= uint64(len(index.Sources)) {
				return nil, fmt.Errorf("source index out of range")
			}
			if chunk.Offset+uint64(chunk.CompressedSize) < chunk.Offset {
				return nil, fmt.Errorf("chunk offset overflow")
			}
			write(chunk)
		}
	}
	data := out.Bytes()
	transform(data)
	return data, nil
}
