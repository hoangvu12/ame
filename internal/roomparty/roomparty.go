package roomparty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/skin"
)

const workerURL = "wss://ame-rooms-ws.kaguya-gindex.workers.dev"
const heartbeatInterval = 30 * time.Second
const reconnectDelay = 3 * time.Second

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
	mu         sync.Mutex
	active     bool
	roomKey    string
	puuid      string
	teamPuuids []string
	mySkinInfo SkinInfo
	teammates  []Member
	conn       *websocket.Conn
	cancelRead context.CancelFunc
	OnUpdate   OnUpdateFunc
	dialer     *websocket.Dialer
}

// NewRoomState creates a new room state manager.
func NewRoomState() *RoomState {
	return &RoomState{
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Join starts a room party session. Connects WS and sends join message.
func (rs *RoomState) Join(roomKey, puuid string, teamPuuids []string) {
	rs.mu.Lock()

	// If already in a room, close existing connection
	if rs.active && rs.cancelRead != nil {
		rs.cancelRead()
	}
	if rs.conn != nil {
		rs.conn.Close()
		rs.conn = nil
	}

	rs.active = true
	rs.roomKey = roomKey
	rs.puuid = puuid
	rs.teamPuuids = teamPuuids
	rs.teammates = nil
	rs.mu.Unlock()

	display.SetPartyKey("display.value.party_in_room_teammates", map[string]interface{}{"count": len(teamPuuids)})
	display.Log(fmt.Sprintf("Room key: %s", roomKey))

	go rs.connectAndRun(roomKey, puuid)
}

// UpdateSkin updates the local user's skin info and sends it over WS.
func (rs *RoomState) UpdateSkin(info SkinInfo) {
	rs.mu.Lock()
	rs.mySkinInfo = info
	active := rs.active
	rs.mu.Unlock()

	if active {
		rs.sendJSON(map[string]interface{}{
			"type":     "skin",
			"skinInfo": info,
		})
	}
}

// Leave stops the WS connection and sends a leave message.
func (rs *RoomState) Leave() {
	rs.mu.Lock()
	if !rs.active {
		rs.mu.Unlock()
		return
	}
	rs.active = false
	if rs.cancelRead != nil {
		rs.cancelRead()
		rs.cancelRead = nil
	}
	conn := rs.conn
	rs.conn = nil
	rs.teammates = nil
	rs.mu.Unlock()

	display.SetPartyKey("display.value.party_off", nil)

	if conn != nil {
		// Best-effort leave message
		data, _ := json.Marshal(map[string]string{"type": "leave"})
		conn.WriteMessage(websocket.TextMessage, data)
		conn.Close()
	}
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

// connectAndRun connects the WS, sends join, runs readLoop, and auto-reconnects on drop.
func (rs *RoomState) connectAndRun(roomKey, puuid string) {
	for {
		rs.mu.Lock()
		if !rs.active || rs.roomKey != roomKey {
			rs.mu.Unlock()
			return
		}
		skinInfo := rs.mySkinInfo
		rs.mu.Unlock()

		conn, err := rs.connectWS(roomKey)
		if err != nil {
			time.Sleep(reconnectDelay)
			continue
		}

		rs.mu.Lock()
		if !rs.active || rs.roomKey != roomKey {
			rs.mu.Unlock()
			conn.Close()
			return
		}
		rs.conn = conn
		ctx, cancel := context.WithCancel(context.Background())
		rs.cancelRead = cancel
		rs.mu.Unlock()

		// Send join message
		joinMsg, _ := json.Marshal(map[string]interface{}{
			"type":     "join",
			"puuid":    puuid,
			"skinInfo": skinInfo,
		})
		rs.mu.Lock()
		if rs.conn != nil {
			rs.conn.WriteMessage(websocket.TextMessage, joinMsg)
		}
		rs.mu.Unlock()

		rs.readLoop(ctx, roomKey)

		// readLoop exited — check if we should reconnect
		rs.mu.Lock()
		shouldReconnect := rs.active && rs.roomKey == roomKey
		rs.mu.Unlock()

		if !shouldReconnect {
			return
		}

		time.Sleep(reconnectDelay)
	}
}

// connectWS dials the WebSocket endpoint.
func (rs *RoomState) connectWS(roomKey string) (*websocket.Conn, error) {
	u := fmt.Sprintf("%s/?roomKey=%s", workerURL, roomKey)
	header := http.Header{}
	conn, _, err := rs.dialer.Dial(u, header)
	return conn, err
}

// readLoop reads WS messages and runs a heartbeat goroutine.
func (rs *RoomState) readLoop(ctx context.Context, roomKey string) {
	// Start heartbeat goroutine
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rs.mu.Lock()
				if rs.conn == nil {
					rs.mu.Unlock()
					return
				}
				err := rs.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				rs.mu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

	defer func() {
		rs.mu.Lock()
		if rs.cancelRead != nil {
			rs.cancelRead()
		}
		rs.mu.Unlock()
		<-heartbeatDone
	}()

	for {
		rs.mu.Lock()
		conn := rs.conn
		rs.mu.Unlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		text := string(message)
		if text == "pong" {
			continue
		}

		var msg struct {
			Type    string   `json:"type"`
			Members []Member `json:"members"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.Type != "members" {
			continue
		}

		rs.mu.Lock()
		if !rs.active || rs.roomKey != roomKey {
			rs.mu.Unlock()
			return
		}
		oldTeammates := rs.teammates
		rs.teammates = filterTeammates(msg.Members, rs.teamPuuids)
		newTeammates := make([]Member, len(rs.teammates))
		copy(newTeammates, rs.teammates)
		onUpdate := rs.OnUpdate
		rs.mu.Unlock()

		if len(newTeammates) != len(oldTeammates) {
			if len(newTeammates) > 0 {
				display.SetPartyKey("display.value.party_active_users", map[string]interface{}{"count": len(newTeammates)})
			} else {
				display.SetPartyKey("display.value.party_in_room_waiting", nil)
			}
		}

		if onUpdate != nil {
			onUpdate(newTeammates)
		}

		go rs.prefetchTeammateSkins(oldTeammates, newTeammates)
	}
}

// sendJSON marshals v and writes it to the WS connection under lock.
func (rs *RoomState) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.conn == nil {
		return
	}
	rs.conn.WriteMessage(websocket.TextMessage, data)
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
