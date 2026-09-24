// celestial-prepare builds experimental Celestial overlay files without
// launching the host or DLL. It is deliberately separate from AME's backend.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hoangvu12/ame/internal/celestial"
)

func run() error {
	manifest := flag.String("manifest", "", "JSON array of WADInput objects in priority order (last wins)")
	parent := flag.String("parent", "", "existing absolute directory for a fresh overlay subdirectory")
	flag.Parse()
	if *manifest == "" || *parent == "" || flag.NArg() != 0 {
		return fmt.Errorf("usage: celestial-prepare -manifest inputs.json -parent <absolute directory>")
	}
	f, err := os.Open(*manifest)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > 8<<20 {
		return fmt.Errorf("manifest must be a regular file no larger than 8 MiB")
	}
	decoder := json.NewDecoder(io.LimitReader(f, 8<<20))
	decoder.DisallowUnknownFields()
	var inputs []celestial.WADInput
	if err := decoder.Decode(&inputs); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("manifest must contain exactly one JSON array")
	}
	prepared, err := celestial.PrepareOverlay(*parent, inputs)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Directory      string `json:"directory"`
		SealRoot       string `json:"sealRoot"`
		Sources        int    `json:"sources"`
		Targets        int    `json:"targets"`
		Chunks         int    `json:"chunks"`
		RuntimeStarted bool   `json:"runtimeStarted"`
	}{prepared.Directory, hex.EncodeToString(prepared.SealRoot[:]), prepared.Sources, prepared.Targets, prepared.Chunks, false})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
