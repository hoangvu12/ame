package roomparty

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hoangvu12/ame/internal/config"
)

func TestPartyOnlyModNamesExcludeLocalDefault(t *testing.T) {
	originalModsDir := config.ModsDir
	config.ModsDir = t.TempDir()
	t.Cleanup(func() { config.ModsDir = originalModsDir })

	teammateMod := filepath.Join(config.ModsDir, "skin_2002")
	if err := os.MkdirAll(teammateMod, 0755); err != nil {
		t.Fatal(err)
	}

	state := NewRoomState()
	state.teammates = []Member{{SkinInfo: SkinInfo{SkinID: "2002"}}}

	if !state.HasTeammateSkins() {
		t.Fatal("expected teammate skin to be detected")
	}
	if got := state.GetAllModNames("0"); got != "skin_2002" {
		t.Fatalf("expected party-only mod list, got %q", got)
	}
	if got := state.BuiltTeammateSkinCount(); got != 1 {
		t.Fatalf("expected one built teammate skin, got %d", got)
	}
}
