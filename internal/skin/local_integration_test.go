package skin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/ltk"
)

// Opt-in offline acceptance. Never launches League or runoverlay, and writes
// only to t.TempDir. Build the embedded helper before running this test.
func TestLocalInstalledGame(t *testing.T) {
	gameDir := os.Getenv("AME_TEST_GAME_DIR")
	if gameDir == "" {
		t.Skip("set AME_TEST_GAME_DIR for installed-file/offline overlay validation")
	}
	oldAme, oldSkins := config.AmeDir, config.SkinsDir
	root := t.TempDir()
	config.AmeDir = root
	config.SkinsDir = filepath.Join(root, "skins")
	defer func() { config.AmeDir, config.SkinsDir = oldAme, oldSkins }()
	start := time.Now()
	if _, err := helperDir(); err != nil {
		t.Fatal(err)
	}
	t.Logf("helper preparation: %s; installed build: %s", time.Since(start), installedBuild(filepath.Join(gameDir, "League of Legends.exe")))
	requests := []PackageRequest{{"103", "103001", ""}, {"1", "1001", ""}, {"22", "22001", ""}, {"86", "86001", ""}, {"99", "99001", ""}, {"22", "22008", ""}, {"21", "21069", ""}, {"238", "238068", ""}, {"99", "99996", "99007"}, {"99", "99020", "99019"}, {"37", "37998", "37006"}, {"222", "222999", "222060"}, {"875", "875999", "875066"}}
	paths := map[string]string{}
	mods := filepath.Join(root, "mods")
	var names []string
	start = time.Now()
	for _, r := range requests {
		path, err := EnsurePackage(context.Background(), r, gameDir)
		if err != nil {
			t.Fatal(err)
		}
		meta, err := readCacheMetadata(r.ChampionID, r.SkinID)
		if err != nil || meta.Source != "local" {
			t.Fatalf("not locally generated: %+v %v", meta, err)
		}
		paths[r.SkinID] = path
		name := "skin_" + r.SkinID
		names = append(names, name)
		if err := ExtractPackage(path, filepath.Join(mods, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d fresh packages + extraction: %s", len(requests), time.Since(start))
	start = time.Now()
	for round := 0; round < 3; round++ {
		for _, r := range requests {
			path, err := EnsurePackage(context.Background(), r, gameDir)
			if err != nil || path != paths[r.SkinID] {
				t.Fatalf("unchanged selection not reused: %s %v", path, err)
			}
		}
	}
	t.Logf("three rounds of %d cached selections: %s", len(requests), time.Since(start))
	for _, r := range []PackageRequest{{"103", "103002", ""}, {"1", "1002", ""}, {"103", "103001", ""}} {
		path, err := EnsurePackage(context.Background(), r, gameDir)
		if err != nil {
			t.Fatal(err)
		}
		if old := paths[r.SkinID]; old != "" && path != old {
			t.Fatal("return selection regenerated")
		}
	}
	if err := ltk.Build(mods, filepath.Join(root, "ltk-overlay"), gameDir, strings.Join(names, "/")); err != nil {
		t.Fatal(err)
	}
	t.Log("combined LTK overlay built, including localized PROJECT Ashe voice")
	if tool := os.Getenv("AME_TEST_MOD_TOOLS"); tool != "" {
		start = time.Now()
		cmd := exec.Command(tool, "mkoverlay", mods, filepath.Join(root, "overlay"), "--game:"+gameDir, "--mods:"+strings.Join(names, "/"), "--noTFT", "--ignoreConflict")
		hideHelper(cmd)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("mkoverlay: %v\n%s", err, out)
		}
		t.Logf("combined offline overlay: %s", time.Since(start))
	}
}
