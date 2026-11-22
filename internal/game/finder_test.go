package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hoangvu12/ame/internal/config"
)

func TestActiveClientOverridesCachedInstall(t *testing.T) {
	oldActive, oldCache, oldSaved := activeClientGameDir, loadCachedGameDir(), config.GamePath()
	t.Cleanup(func() { activeClientGameDir = oldActive; storeCachedGameDir(oldCache); config.SetGamePath(oldSaved) })
	install := func() string {
		d := t.TempDir()
		if err := os.WriteFile(filepath.Join(d, "League of Legends.exe"), nil, 0600); err != nil {
			t.Fatal(err)
		}
		return d
	}
	stale, active := install(), install()
	storeCachedGameDir(stale)
	activeClientGameDir = func() string { return "" }
	if got := FindGameDir(); got != stale {
		t.Fatalf("before client: %s", got)
	}
	activeClientGameDir = func() string { return active }
	if got := FindGameDir(); got != active {
		t.Fatalf("cached installation defeated active client: %s", got)
	}
}

func TestManualSelectionValidatesAndRefreshesCache(t *testing.T) {
	oldCache, oldSaved := loadCachedGameDir(), config.GamePath()
	t.Cleanup(func() { storeCachedGameDir(oldCache); config.SetGamePath(oldSaved) })
	d := t.TempDir()
	if err := SelectGameDir(d); err == nil {
		t.Fatal("accepted folder without League")
	}
	if err := os.WriteFile(filepath.Join(d, "League of Legends.exe"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := SelectGameDir(d); err != nil {
		t.Fatal(err)
	}
	if got := loadCachedGameDir(); got != d {
		t.Fatalf("selection did not update cache: %s", got)
	}
}
