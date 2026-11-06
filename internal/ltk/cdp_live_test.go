package ltk

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// TestLiveManagerSession launches the real LTK Manager, drives it over the
// DevTools protocol, and installs a skin package into its library. The
// patcher itself is only started when a League installation is configured.
// Enabled only with AME_LIVE_CDP=1.
func TestLiveManagerSession(t *testing.T) {
	if os.Getenv("AME_LIVE_CDP") != "1" {
		t.Skip("set AME_LIVE_CDP=1 to run the live manager session test")
	}
	old := config.AmeDir
	config.AmeDir = t.TempDir()
	defer func() { config.AmeDir = old }()
	fetchStub := managerFetch
	managerFetch = ManagerFetch
	defer func() { managerFetch = fetchStub }()

	// Pack a real skin mod: the importer mounts packed WADs, so a synthetic
	// archive cannot be installed.
	modDir := os.Getenv("AME_LIVE_MOD")
	if modDir == "" {
		modDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame", "mods", "skin_103088")
	}
	if _, err := os.Stat(filepath.Join(modDir, "WAD")); err != nil {
		t.Skipf("no real mod to install at %s: %v", modDir, err)
	}
	pkg := filepath.Join(t.TempDir(), "live.fantome")
	if err := packMod(modDir, pkg); err != nil {
		t.Fatal(err)
	}
	if err := mergeFantomes([]string{pkg}, pkg+".merged", nil); err != nil {
		t.Fatal(err)
	}
	pkg = pkg + ".merged"

	if game := os.Getenv("AME_LIVE_GAME_DIR"); game != "" {
		gameDir = game
	} else {
		gameDir = `C:\Riot Games\League of Legends\Game`
	}

	managerMu.Lock()
	m, err := ensureManagerLocked()
	managerMu.Unlock()
	defer Stop()
	if err != nil {
		t.Fatalf("manager session: %v", err)
	}

	mods, err := m.cdp.installedMods()
	if err != nil {
		t.Fatalf("get_installed_mods: %v", err)
	}
	t.Logf("library holds %d mods before install", len(mods))

	status, err := m.cdp.patcherStatus()
	if err != nil {
		t.Fatalf("get_patcher_status: %v", err)
	}
	t.Logf("patcher before start: running=%v phase=%s", status.Running, status.Phase)

	modID, err := m.cdp.installMod(pkg)
	if err != nil {
		t.Fatalf("install_mod: %v", err)
	}
	t.Logf("installed mod id=%s", modID)

	if err := m.cdp.toggleMod(modID, true); err != nil {
		t.Fatalf("toggle_mod: %v", err)
	}
	mods, err = m.cdp.installedMods()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, mod := range mods {
		if mod.ID == modID {
			found = true
			t.Logf("mod enabled=%v name=%q", mod.Enabled, mod.Name)
		}
	}
	if !found {
		t.Fatal("installed mod missing from the library")
	}

	if os.Getenv("AME_LIVE_START_PATCHER") == "1" {
		if err := m.cdp.startPatcher(); err != nil {
			t.Fatalf("start_patcher: %v", err)
		}
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			s := currentSnapshot()
			if s.phase == "patching" || s.failure != "" {
				t.Logf("phase=%s attached=%v failure=%q", s.phase, s.attached, s.failure)
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
		status, err := m.cdp.patcherStatus()
		if err == nil {
			t.Logf("patcher after start: running=%v phase=%s prefix=%v", status.Running, status.Phase, status.Session)
		}
	}

	if err := m.cdp.uninstallMod(modID); err != nil {
		t.Fatalf("uninstall_mod: %v", err)
	}
}
