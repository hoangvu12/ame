package roomparty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/skin"
)

const workerURL = "https://ame-rooms.kaguya-gindex.workers.dev"
const pollInterval = 3 * time.Second

// SkinInfo describes a skin selection shared between Ame users.
type SkinInfo struct {
	ChampionID   string `json:"championId"`
	SkinID       string `json:"skinId"`
	BaseSkinID   string `json:"baseSkinId"`
	ChampionName string `json:"championName"`
	SkinName     string `json:"skinName"`
	ChromaName   string `json:"chromaName"`
}

// Member represents another Ame user in the same game.
type Member struct {
	Puuid    string   `json:"puuid"`
	SkinInfo SkinInfo `json:"skinInfo"`
}

// OnUpdateFunc is called when the teammate list changes.
type OnUpdateFunc func(teammates []Member)

// RoomState manages a room party session.
type RoomState struct {
	mu           sync.Mutex
	active       bool
	roomKey      string
	puuid        string
	teamPuuids   []string
	mySkinInfo   SkinInfo
	teammates    []Member
	cancelPoll   context.CancelFunc
	OnUpdate     OnUpdateFunc
	client       *http.Client
}

// NewRoomState creates a new room state manager.
func NewRoomState() *RoomState {
	return &RoomState{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Join starts a room party session. Registers with the CF Worker and starts polling.
func (rs *RoomState) Join(roomKey, puuid string, teamPuuids []string) {
	rs.mu.Lock()

	// If already in a room, leave first
	if rs.active && rs.cancelPoll != nil {
		rs.cancelPoll()
	}

	rs.active = true
	rs.roomKey = roomKey
	rs.puuid = puuid
	rs.teamPuuids = teamPuuids
	rs.teammates = nil

	ctx, cancel := context.WithCancel(context.Background())
	rs.cancelPoll = cancel
	rs.mu.Unlock()

	display.SetParty(fmt.Sprintf("In room (%d teammates)", len(teamPuuids)))
	display.Log(fmt.Sprintf("Room key: %s", roomKey))

	// Register with CF Worker
	rs.postJoin()

	// Start polling
	go rs.pollLoop(ctx, roomKey, puuid)
}

// UpdateSkin updates the local user's skin info and re-registers with the CF Worker.
func (rs *RoomState) UpdateSkin(info SkinInfo) {
	rs.mu.Lock()
	rs.mySkinInfo = info
	active := rs.active
	rs.mu.Unlock()

	if active {
		rs.postJoin()
	}
}

// Leave stops polling and sends a leave request.
func (rs *RoomState) Leave() {
	rs.mu.Lock()
	if !rs.active {
		rs.mu.Unlock()
		return
	}
	rs.active = false
	if rs.cancelPoll != nil {
		rs.cancelPoll()
		rs.cancelPoll = nil
	}
	rs.teammates = nil
	roomKey := rs.roomKey
	puuid := rs.puuid
	rs.mu.Unlock()

	display.SetParty("Off")

	// Best-effort leave — use captured values to avoid racing with a
	// subsequent Join() that may overwrite rs.roomKey before the goroutine runs.
	go rs.postLeaveFor(roomKey, puuid)
}

// GetTeammates returns the current list of Ame-using teammates.
func (rs *RoomState) GetTeammates() []Member {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	cp := make([]Member, len(rs.teammates))
	copy(cp, rs.teammates)
	return cp
}

// IsActive returns whether a room party session is active.
func (rs *RoomState) IsActive() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.active
}

// GetAllModNames returns a slash-separated mod directory name list for mkoverlay --mods.
// mod-tools expects slash-separated names (e.g. "skin_1/skin_2/skin_3").
// It includes the user's own skin and all teammate skins that exist in ModsDir.
func (rs *RoomState) GetAllModNames(ownSkinID string) string {
	rs.mu.Lock()
	teammates := make([]Member, len(rs.teammates))
	copy(teammates, rs.teammates)
	rs.mu.Unlock()

	names := []string{fmt.Sprintf("skin_%s", ownSkinID)}
	seen := map[string]bool{ownSkinID: true}

	for _, tm := range teammates {
		sid := tm.SkinInfo.SkinID
		if sid == "" || seen[sid] {
			continue
		}
		modDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", sid))
		if _, err := os.Stat(modDir); err == nil {
			names = append(names, fmt.Sprintf("skin_%s", sid))
			seen[sid] = true
		}
	}

	return strings.Join(names, "/")
}

// ComputeModKey returns a sorted, deduplicated, comma-separated skin ID list
// for cache-keying the prebuilt overlay.
func (rs *RoomState) ComputeModKey(ownSkinID string) string {
	rs.mu.Lock()
	teammates := make([]Member, len(rs.teammates))
	copy(teammates, rs.teammates)
	rs.mu.Unlock()

	ids := []string{ownSkinID}
	seen := map[string]bool{ownSkinID: true}
	for _, tm := range teammates {
		sid := tm.SkinInfo.SkinID
		if sid == "" || seen[sid] {
			continue
		}
		ids = append(ids, sid)
		seen[sid] = true
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

// ComputeBuiltModKey returns a sorted, comma-separated skin ID list containing only
// the own skin and teammate skins whose mod directories actually exist in ModsDir.
// Use this (instead of ComputeModKey) when recording what was actually built into the overlay.
func (rs *RoomState) ComputeBuiltModKey(ownSkinID string) string {
	rs.mu.Lock()
	teammates := make([]Member, len(rs.teammates))
	copy(teammates, rs.teammates)
	rs.mu.Unlock()

	ids := []string{ownSkinID}
	seen := map[string]bool{ownSkinID: true}
	for _, tm := range teammates {
		sid := tm.SkinInfo.SkinID
		if sid == "" || seen[sid] {
			continue
		}
		modDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", sid))
		if _, err := os.Stat(modDir); err == nil {
			ids = append(ids, sid)
			seen[sid] = true
		}
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

// DownloadTeammateSkins downloads all teammate skins that are not yet cached.
// It also extracts them into ModsDir for overlay building.
func (rs *RoomState) DownloadTeammateSkins() {
	teammates := rs.GetTeammates()
	var wg sync.WaitGroup

	for _, tm := range teammates {
		info := tm.SkinInfo
		if info.SkinID == "" {
			continue
		}
		// Skip if already extracted in mods dir
		modDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", info.SkinID))
		if _, err := os.Stat(modDir); err == nil {
			continue
		}

		wg.Add(1)
		go func(si SkinInfo) {
			defer wg.Done()
			zipPath := skin.GetCachedPath(si.ChampionID, si.SkinID)
			if zipPath == "" {
				downloaded, err := skin.Download(si.ChampionID, si.SkinID, si.BaseSkinID, si.ChampionName, si.SkinName, si.ChromaName)
				if err != nil {
					display.Log(fmt.Sprintf("! Teammate skin unavailable: %s", si.SkinName))
					return
				}
				zipPath = downloaded
			}

			os.MkdirAll(modDir, os.ModePerm)
			if err := skin.Extract(zipPath, modDir); err != nil {
				display.Log(fmt.Sprintf("! Failed to extract teammate skin: %s", si.SkinName))
				return
			}
			display.Log(fmt.Sprintf("Teammate skin ready: %s", si.SkinName))
		}(info)
	}

	wg.Wait()
}

// pollLoop polls the CF Worker for room members every pollInterval.
// roomKey identifies which room this loop belongs to; if rs.roomKey has
// changed (rapid Leave→Join), the stale result is discarded.
func (rs *RoomState) pollLoop(ctx context.Context, roomKey, puuid string) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			members, err := rs.fetchMembersFor(roomKey, puuid)
			if err != nil {
				continue
			}

			rs.mu.Lock()
			if !rs.active || rs.roomKey != roomKey {
				rs.mu.Unlock()
				return
			}
			oldTeammates := rs.teammates
			rs.teammates = filterTeammates(members, rs.teamPuuids)
			newTeammates := make([]Member, len(rs.teammates))
			copy(newTeammates, rs.teammates)
			onUpdate := rs.OnUpdate
			rs.mu.Unlock()

			// Update display if teammate count changed
			if len(newTeammates) != len(oldTeammates) {
				if len(newTeammates) > 0 {
					display.SetParty(fmt.Sprintf("Active (%d Ame users)", len(newTeammates)))
				} else {
					display.SetParty("In room (waiting)")
				}
			}

			// Notify plugin of updates
			if onUpdate != nil {
				onUpdate(newTeammates)
			}

			// Prefetch new/changed teammate skins
			go rs.prefetchTeammateSkins(oldTeammates, newTeammates)
		}
	}
}

// prefetchTeammateSkins downloads skins for newly discovered teammates or changed skins.
func (rs *RoomState) prefetchTeammateSkins(old, current []Member) {
	oldMap := make(map[string]string, len(old))
	for _, m := range old {
		oldMap[m.Puuid] = m.SkinInfo.SkinID
	}

	for _, m := range current {
		if m.SkinInfo.SkinID == "" {
			continue
		}
		// Skip if skin is same as before
		if oldMap[m.Puuid] == m.SkinInfo.SkinID {
			continue
		}
		// Skip if already cached
		if skin.GetCachedPath(m.SkinInfo.ChampionID, m.SkinInfo.SkinID) != "" {
			continue
		}
		// Download in background
		go func(si SkinInfo) {
			_, err := skin.Download(si.ChampionID, si.SkinID, si.BaseSkinID, si.ChampionName, si.SkinName, si.ChromaName)
			if err == nil {
				display.Log(fmt.Sprintf("Prefetched teammate skin: %s", si.SkinName))
			}
		}(m.SkinInfo)
	}
}

// filterTeammates filters CF Worker members to only those in myTeam.
func filterTeammates(members []Member, teamPuuids []string) []Member {
	teamSet := make(map[string]bool, len(teamPuuids))
	for _, p := range teamPuuids {
		teamSet[p] = true
	}
	var result []Member
	for _, m := range members {
		if teamSet[m.Puuid] {
			result = append(result, m)
		}
	}
	return result
}

// --- HTTP client methods ---

func (rs *RoomState) postJoin() {
	rs.mu.Lock()
	roomKey := rs.roomKey
	puuid := rs.puuid
	skinInfo := rs.mySkinInfo
	rs.mu.Unlock()

	body := map[string]interface{}{
		"roomKey":  roomKey,
		"puuid":    puuid,
		"skinInfo": skinInfo,
	}
	data, _ := json.Marshal(body)

	resp, err := rs.client.Post(workerURL+"/rooms/join", "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	resp.Body.Close()
}

type membersResponse struct {
	Members []Member `json:"members"`
}

func (rs *RoomState) fetchMembersFor(roomKey, puuid string) ([]Member, error) {
	u := fmt.Sprintf("%s/rooms/members?roomKey=%s&puuid=%s",
		workerURL,
		url.QueryEscape(roomKey),
		url.QueryEscape(puuid),
	)

	resp, err := rs.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result membersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Members, nil
}

func (rs *RoomState) postLeaveFor(roomKey, puuid string) {
	body := map[string]interface{}{
		"roomKey": roomKey,
		"puuid":   puuid,
	}
	data, _ := json.Marshal(body)

	resp, err := rs.client.Post(workerURL+"/rooms/leave", "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	resp.Body.Close()
}
