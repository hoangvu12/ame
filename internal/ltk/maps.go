package ltk

// Map copies kept ready between skin sets.
//
// The manager writes its own copy of every game WAD a skin set changes, and
// a skin entry also lands in every map WAD that holds the same entry, so
// some champions touch the multi-gigabyte map WADs. A copy that is still on
// disk is updated by rewriting its end (milliseconds); a missing one is
// copied in full (seconds on an NVMe drive, over a minute on a slow disk).
// After every build the manager deletes the copies the current set does not
// use. So every skin set carries one tiny constant entry per map whose copy
// is ready: the copy stays, and a later change that touches the map stays a
// fast end-of-file rewrite.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hoangvu12/ame/internal/config"
)

// keepAliveMaps are the map WADs skin entries are routed into (Summoner's
// Rift, ARAM, Arena). The others hold no entry any measured skin touches.
var keepAliveMaps = []string{"Map11.wad.client", "Map12.wad.client", "Map30.wad.client"}

// keepAliveContent is constant, so a map counts as changed only when a real
// skin entry touches it.
const keepAliveContent = "ame keep-alive v1\n"

// packageEntry is one file ame adds to a skin package beside the mods.
type packageEntry struct {
	name string
	data []byte
}

// mapPrepEnabled reports whether map copies are kept ready: the ame setting
// (on by default) and the AME_MAP_PREP=0 kill switch. A variable for tests.
var mapPrepEnabled = func() bool {
	return config.MapPrep() && os.Getenv("AME_MAP_PREP") != "0"
}

func gameMapsDir(game string) string {
	return filepath.Join(game, "DATA", "FINAL", "Maps", "Shipping")
}

// managerOverlayRoot is the overlay folder of the manager's default profile.
// A variable for tests.
var managerOverlayRoot = func() string {
	settings := managerSettingsFile()
	if settings == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(settings), "profiles", "default", "overlay")
}

// keepAliveEntry is the directory-layout entry that routes a constant file
// into one map WAD. The path is in no game WAD, so the builder treats it as
// a new entry for the declared WAD; a copy of a real game entry would be
// dropped as identical to the game.
func keepAliveEntry(wad string) packageEntry {
	base := strings.ToLower(strings.TrimSuffix(wad, ".wad.client"))
	return packageEntry{
		name: "WAD/" + wad + "/ame/keepalive/" + base + ".txt",
		data: []byte(keepAliveContent),
	}
}

// mapCopyState compares the manager's copy of one map WAD with League's
// file. An empty result means the copy is ready: present, at least as large
// as the game's file (the copy holds all of it plus the skin data) and
// written after it. Otherwise it says why not, for the logs.
func mapCopyState(game, overlay, wad string) (present bool, problem string) {
	original, err := os.Stat(filepath.Join(gameMapsDir(game), wad))
	if err != nil || !original.Mode().IsRegular() {
		return false, ""
	}
	name := strings.TrimSuffix(wad, ".wad.client")
	if overlay == "" {
		return true, name + " has no manager overlay folder"
	}
	copied, err := os.Stat(filepath.Join(gameMapsDir(overlay), wad))
	switch {
	case err != nil:
		return true, name + " missing"
	case copied.Size() < original.Size():
		return true, name + " smaller than League's file"
	case !copied.ModTime().After(original.ModTime()):
		return true, name + " older than League's file"
	}
	return true, ""
}

// readyMaps lists the keep-alive maps whose manager copy is ready.
func readyMaps(game string) []string {
	if game == "" {
		return nil
	}
	overlay := managerOverlayRoot()
	var ready []string
	for _, wad := range keepAliveMaps {
		if present, problem := mapCopyState(game, overlay, wad); present && problem == "" {
			ready = append(ready, wad)
		}
	}
	return ready
}

// MapsPrepared reports whether the manager holds a ready copy of every
// keep-alive map League has. The reason names what is missing.
func MapsPrepared(game string) (bool, string) {
	if game == "" {
		return false, "League not found"
	}
	overlay := managerOverlayRoot()
	var problems []string
	found := 0
	for _, wad := range keepAliveMaps {
		present, problem := mapCopyState(game, overlay, wad)
		if !present {
			continue
		}
		found++
		if problem != "" {
			problems = append(problems, problem)
		}
	}
	if found == 0 {
		return false, "League has none of the maps"
	}
	if len(problems) > 0 {
		return false, strings.Join(problems, ", ")
	}
	return true, ""
}

// keepAliveEntries are the entries a skin set built against game carries:
// one per map whose copy is ready. A map without a ready copy gets none, so
// a set that does not touch it never pays for a full copy; preparation
// (PrepareMaps) makes the copies outside champion select.
func keepAliveEntries(game string) []packageEntry {
	if !mapPrepEnabled() {
		return nil
	}
	var entries []packageEntry
	for _, wad := range readyMaps(game) {
		entries = append(entries, keepAliveEntry(wad))
	}
	return entries
}
