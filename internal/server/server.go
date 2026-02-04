package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/modtools"
	"github.com/hoangvu12/ame/internal/lcu"
	"github.com/hoangvu12/ame/internal/roomparty"
	"github.com/hoangvu12/ame/internal/setup"
	"github.com/hoangvu12/ame/internal/skin"
	"github.com/hoangvu12/ame/internal/startup"
	"github.com/hoangvu12/ame/internal/suspend"
)

// ApplyMessage represents an apply skin request
type ApplyMessage struct {
	Type         string      `json:"type"`
	ChampionID   interface{} `json:"championId"`
	SkinID       interface{} `json:"skinId"`
	BaseSkinID   interface{} `json:"baseSkinId,omitempty"`
	ChampionName string      `json:"championName,omitempty"`
	SkinName     string      `json:"skinName,omitempty"`
	ChromaName   string      `json:"chromaName,omitempty"`
}

// CleanupMessage represents a cleanup request
type CleanupMessage struct {
	Type string `json:"type"`
}

// StatusMessage represents a status response
type StatusMessage struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StateMessage represents the current overlay state sent in response to a query
type StateMessage struct {
	Type          string `json:"type"`
	ChampionID    string `json:"championId,omitempty"`
	SkinID        string `json:"skinId,omitempty"`
	BaseSkinID    string `json:"baseSkinId,omitempty"`
	ChampionName  string `json:"championName,omitempty"`
	SkinName      string `json:"skinName,omitempty"`
	ChromaName    string `json:"chromaName,omitempty"`
	OverlayActive bool   `json:"overlayActive"`
}

// GamePathMessage represents a game path request/response
type GamePathMessage struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// BoolSettingMessage is a generic message for boolean setting get/set
type BoolSettingMessage struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// AutoSelectRoleMessage represents a per-role auto-select config update
type AutoSelectRoleMessage struct {
	Type  string `json:"type"`
	Role  string `json:"role"`
	Picks []int  `json:"picks"`
	Bans  []int  `json:"bans"`
}

// RoomPartyJoinMessage is sent by the plugin when champ select starts with room party enabled
type RoomPartyJoinMessage struct {
	Type       string   `json:"type"`
	RoomKey    string   `json:"roomKey"`
	Puuid      string   `json:"puuid"`
	TeamPuuids []string `json:"teamPuuids"`
}

// RoomPartySkinMessage is sent by the plugin when the user's skin selection changes
type RoomPartySkinMessage struct {
	Type         string      `json:"type"`
	ChampionID   interface{} `json:"championId"`
	SkinID       interface{} `json:"skinId"`
	BaseSkinID   interface{} `json:"baseSkinId,omitempty"`
	ChampionName string      `json:"championName,omitempty"`
	SkinName     string      `json:"skinName,omitempty"`
	ChromaName   string      `json:"chromaName,omitempty"`
}

// RoomPartyUpdateMessage is sent TO the plugin with teammate info
type RoomPartyUpdateMessage struct {
	Type      string             `json:"type"`
	Teammates []roomparty.Member `json:"teammates"`
}

// ChatStatusSettingMessage represents a chat status get/set
type ChatStatusSettingMessage struct {
	Type          string `json:"type"`
	Availability  string `json:"availability"`
	StatusMessage string `json:"statusMessage"`
}

// IncomingMessage is used for parsing the message type first
type IncomingMessage struct {
	Type string `json:"type"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// toString converts interface{} to string (handles both string and number types)
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// Last applied skin state — survives client reconnects
var lastChampionID string
var lastSkinID string
var lastBaseSkinID string
var lastChampionName string
var lastSkinName string
var lastChromaName string
var lastModKey string
var stateMu sync.Mutex

// Prebuild state — tracks overlay pre-built during champion select
var overlayBuildMu sync.Mutex
var prebuiltModKey string

// Room party state
var roomState = roomparty.NewRoomState()

// sendStatus sends a status message to the WebSocket client
func sendStatus(conn *websocket.Conn, status, message string) {
	payload := StatusMessage{
		Type:    "status",
		Status:  status,
		Message: message,
	}
	data, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, data)

	// Log important status changes to the live display
	switch status {
	case "ready":
		display.Log(message)
	case "error":
		display.Log("! " + message)
	}
}

// OnUninstall is called after uninstall cleanup to trigger app exit.
var OnUninstall func()

// handleUninstall performs a full uninstall: deactivates Pengu, removes
// files, schedules ame directory deletion, then exits.
func handleUninstall(conn *websocket.Conn) {
	display.Log("Uninstalling...")
	hiddenAttr := &syscall.SysProcAttr{HideWindow: true}

	// Check if Pengu is external before we touch the registry
	isExternalPengu := setup.IsUsingExternalPengu()

	// Deactivate Pengu (remove IFEO registry key)
	setup.DeactivatePengu()

	// Restart League Client so it releases core.dll and unloads Pengu
	if lcu.IsClientRunning() {
		display.Log("Restarting League Client...")
		lcu.RestartClient()
	}

	// Remove startup scheduled task
	startup.Disable()

	// Clean up overlay state and kill mod-tools
	HandleCleanup()

	// Kill Pengu Loader process
	kill := exec.Command("taskkill", "/F", "/IM", "Pengu Loader.exe")
	kill.SysProcAttr = hiddenAttr
	kill.Run()

	// Handle Pengu cleanup
	if isExternalPengu {
		// External Pengu: only remove ame plugin, leave Pengu intact
		os.RemoveAll(setup.GetPluginDir())
	}

	// Remove ame subdirectories we can delete now
	os.RemoveAll(config.ToolsDir)
	os.RemoveAll(config.SkinsDir)
	os.RemoveAll(config.ModsDir)
	os.RemoveAll(config.OverlayDir)

	// Schedule deletion of ame directory after process exits.
	// Retries for 30s to handle locked files (core.dll released after client restart).
	ameDir := strings.ReplaceAll(config.AmeDir, "'", "''")
	selfDestruct := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden",
		"-Command", fmt.Sprintf(
			"for ($i = 0; $i -lt 10; $i++) { Start-Sleep 3; Remove-Item -Recurse -Force '%s' -ErrorAction SilentlyContinue; if (!(Test-Path '%s')) { break } }",
			ameDir, ameDir,
		))
	selfDestruct.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	selfDestruct.Start()

	display.Log("Uninstall complete")

	if OnUninstall != nil {
		OnUninstall()
	}
}

// handleApply handles skin apply request
func handleApply(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string) {
	// Compute mod key early: includes own skin + teammate skins if room party is active
	currentModKey := skinID
	if roomState.IsActive() {
		currentModKey = roomState.ComputeModKey(skinID)
	}

	// If runoverlay is already running for this exact mod set, skip — nothing to do
	stateMu.Lock()
	alreadyActive := modtools.IsRunning() && lastModKey == currentModKey
	display.Log(fmt.Sprintf("Apply: modKey=%s lastModKey=%s running=%v alreadyActive=%v", currentModKey, lastModKey, modtools.IsRunning(), alreadyActive))
	stateMu.Unlock()
	if alreadyActive {
		sendStatus(conn, "ready", "Skin applied!")
		return
	}

	// Find game directory
	gameDir := game.FindGameDir()
	if gameDir == "" {
		sendStatus(conn, "error", "League of Legends Game directory not found")
		return
	}

	// Check mod-tools exists
	if !modtools.Exists() {
		sendStatus(conn, "error", "mod-tools.exe not found. Please restart ame.")
		return
	}

	// Check for cached skin file
	zipPath := skin.GetCachedPath(championID, skinID)

	// Download if not cached
	if zipPath == "" {
		downloaded, err := skin.Download(championID, skinID, baseSkinID, championName, skinName, chromaName)
		if err != nil {
			sendStatus(conn, "error", "Skin not available for download")
			return
		}
		zipPath = downloaded
	}

	// Kill any previous runoverlay
	modtools.KillModTools()
	time.Sleep(300 * time.Millisecond)

	// applyDone signals that overlay build + runoverlay start are complete.
	// The suspend goroutine only freezes the game if it appears BEFORE this closes.
	applyDone := make(chan struct{})
	defer close(applyDone)

	// Background: if game appears while we're still building/starting, freeze it.
	go func() {
		// Poll for game process, but stop early if apply finishes first
		var pid uint32
		deadline := time.Now().Add(120 * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-applyDone:
				return // Apply finished, no freeze needed
			default:
			}
			if p := suspend.FindProcess("League of Legends.exe"); p != 0 {
				pid = p
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if pid == 0 {
			return
		}

		// Double-check apply hasn't finished during the FindProcess call
		select {
		case <-applyDone:
			return
		default:
		}

		display.Log("Game detected, holding until ready...")
		s, err := suspend.NewSuspender(pid)
		if err != nil {
			return
		}
		count, suspendErr := s.Suspend()
		if suspendErr != nil || count == 0 {
			s.Close()
			return
		}

		// Wait for apply to finish, with a safety timeout
		select {
		case <-applyDone:
		case <-time.After(30 * time.Second):
		}
		s.Resume()
		display.Log("Game released")
	}()

	// Build overlay (or reuse pre-built one from prefetch).
	overlayBuildMu.Lock()

	teammateSkinCount := 0
	if roomState.IsActive() {
		display.Log(fmt.Sprintf("Apply: room party active, mod key: %s", currentModKey))
	} else {
		display.Log("Apply: room party inactive, own skin only")
	}

	prebuilt := prebuiltModKey == currentModKey
	if prebuilt {
		// Verify overlay dir still exists
		configCheck := filepath.Join(config.OverlayDir, "cslol-config.json")
		if _, err := os.Stat(configCheck); os.IsNotExist(err) {
			prebuilt = false
		}
	}
	if prebuilt {
		display.Log("Apply: using prebuilt overlay")
		prebuiltModKey = ""
		// Count teammate skins from the mod key
		teammateSkinCount = strings.Count(currentModKey, ",")
	} else {
		prebuiltModKey = ""
		os.RemoveAll(config.ModsDir)
		modSubDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", skinID))
		os.MkdirAll(modSubDir, os.ModePerm)

		if err := skin.Extract(zipPath, modSubDir); err != nil {
			overlayBuildMu.Unlock()
			sendStatus(conn, "error", "Failed to extract skin archive")
			return
		}

		// Download and extract teammate skins if room party is active
		if roomState.IsActive() {
			roomState.DownloadTeammateSkins()
		}

		os.RemoveAll(config.OverlayDir)
		os.MkdirAll(config.OverlayDir, os.ModePerm)

		// Build mod list: own skin + teammate skins
		modName := fmt.Sprintf("skin_%s", skinID)
		if roomState.IsActive() {
			modName = roomState.GetAllModNames(skinID)
		}

		display.Log(fmt.Sprintf("Apply: building overlay with mods: %s", modName))
		teammateSkinCount = strings.Count(modName, "/")

		success, exitCode := modtools.RunMkOverlay(config.ModsDir, config.OverlayDir, gameDir, modName)

		if !success {
			overlayBuildMu.Unlock()
			sendStatus(conn, "error", fmt.Sprintf("Failed to apply skin (code %d)", exitCode))
			return
		}
	}
	overlayBuildMu.Unlock()

	// Start runoverlay (hooks game process when it finds it)
	configPath := filepath.Join(config.OverlayDir, "cslol-config.json")
	if err := modtools.RunOverlay(config.OverlayDir, configPath, gameDir); err != nil {
		sendStatus(conn, "error", fmt.Sprintf("Failed to start overlay: %v", err))
		return
	}

	// Track last applied state — use the actual built key, not the theoretical one,
	// so that a later apply with new teammates isn't short-circuited.
	actualModKey := skinID
	if roomState.IsActive() {
		actualModKey = roomState.ComputeBuiltModKey(skinID)
	}
	stateMu.Lock()
	lastChampionID = championID
	lastSkinID = skinID
	lastBaseSkinID = baseSkinID
	lastChampionName = championName
	lastSkinName = skinName
	lastChromaName = chromaName
	lastModKey = actualModKey
	stateMu.Unlock()
	display.Log(fmt.Sprintf("Apply: stored lastModKey=%s (wanted=%s)", actualModKey, currentModKey))

	display.SetSkin(skinName, chromaName)
	display.SetOverlay("Active")

	if teammateSkinCount > 0 {
		sendStatus(conn, "ready", fmt.Sprintf("Skin applied! (+%d teammate skins)", teammateSkinCount))
	} else {
		sendStatus(conn, "ready", "Skin applied!")
	}
}

// handlePrefetch pre-downloads a skin and pre-builds the overlay during champion select
func handlePrefetch(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string) {
	// Download if not cached
	zipPath := skin.GetCachedPath(championID, skinID)
	if zipPath == "" {
		downloaded, err := skin.Download(championID, skinID, baseSkinID, championName, skinName, chromaName)
		if err != nil {
			return
		}
		zipPath = downloaded
	}

	// Don't build if overlay is already active (mid-game)
	if modtools.IsRunning() {
		return
	}

	// Find game directory for mkoverlay
	gameDir := game.FindGameDir()
	if gameDir == "" {
		return
	}

	if !modtools.Exists() {
		return
	}

	// Build overlay so handleApply can skip this step
	overlayBuildMu.Lock()
	defer overlayBuildMu.Unlock()

	// Compute mod key for cache comparison
	currentModKey := skinID
	if roomState.IsActive() {
		currentModKey = roomState.ComputeModKey(skinID)
	}

	// Skip if already pre-built for this exact set of skins
	if prebuiltModKey == currentModKey {
		display.Log(fmt.Sprintf("Prefetch: skipped (already built), modKey=%s", currentModKey))
		return
	}
	display.Log(fmt.Sprintf("Prefetch: modKey=%s (was %s), roomActive=%v", currentModKey, prebuiltModKey, roomState.IsActive()))

	os.RemoveAll(config.ModsDir)
	modSubDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", skinID))
	os.MkdirAll(modSubDir, os.ModePerm)

	if err := skin.Extract(zipPath, modSubDir); err != nil {
		return
	}

	// Download and extract teammate skins if room party is active
	if roomState.IsActive() {
		roomState.DownloadTeammateSkins()
	}

	os.RemoveAll(config.OverlayDir)
	os.MkdirAll(config.OverlayDir, os.ModePerm)

	// Build mod list: own skin + teammate skins
	modName := fmt.Sprintf("skin_%s", skinID)
	if roomState.IsActive() {
		modName = roomState.GetAllModNames(skinID)
	}

	display.Log(fmt.Sprintf("Prefetch: building overlay with mods: %s", modName))

	success, _ := modtools.RunMkOverlay(config.ModsDir, config.OverlayDir, gameDir, modName)
	if !success {
		display.Log("Prefetch: mkoverlay failed")
		return
	}

	// Record what was actually built (only skins whose mod dirs exist),
	// not the theoretical set. This avoids false cache hits when a
	// teammate skin download failed.
	if roomState.IsActive() {
		prebuiltModKey = roomState.ComputeBuiltModKey(skinID)
	} else {
		prebuiltModKey = skinID
	}
	display.Log(fmt.Sprintf("Skin ready (prebuilt key: %s)", prebuiltModKey))
}

// HandleCleanup handles cleanup request
func HandleCleanup() {
	modtools.KillModTools()
	os.RemoveAll(config.OverlayDir)

	overlayBuildMu.Lock()
	prebuiltModKey = ""
	overlayBuildMu.Unlock()

	stateMu.Lock()
	lastChampionID = ""
	lastSkinID = ""
	lastBaseSkinID = ""
	lastChampionName = ""
	lastSkinName = ""
	lastChromaName = ""
	lastModKey = ""
	stateMu.Unlock()

	display.SetSkin("", "")
	display.SetOverlay("Inactive")
}

// broadcastRoomUpdate sends room party teammate info to all connected clients.
func broadcastRoomUpdate(teammates []roomparty.Member) {
	msg := RoomPartyUpdateMessage{
		Type:      "roomPartyUpdate",
		Teammates: teammates,
	}
	data, _ := json.Marshal(msg)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn := range clients {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func init() {
	roomState.OnUpdate = broadcastRoomUpdate
}

// handleConnection handles a single WebSocket connection
func handleConnection(conn *websocket.Conn) {
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		remaining := len(clients)
		clientsMu.Unlock()
		conn.Close()
		if remaining == 0 {
			display.SetStatus("Waiting for client")
		}
		display.Log("Client disconnected")
	}()
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()
	display.SetStatus("Connected")
	display.Log("Client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		switch incoming.Type {
		case "apply":
			var applyMsg ApplyMessage
			if err := json.Unmarshal(message, &applyMsg); err != nil {
				continue
			}
			championID := toString(applyMsg.ChampionID)
			skinID := toString(applyMsg.SkinID)
			baseSkinID := toString(applyMsg.BaseSkinID)
			handleApply(conn, championID, skinID, baseSkinID, applyMsg.ChampionName, applyMsg.SkinName, applyMsg.ChromaName)

		case "prefetch":
			var prefetchMsg ApplyMessage
			if err := json.Unmarshal(message, &prefetchMsg); err != nil {
				continue
			}
			championID := toString(prefetchMsg.ChampionID)
			skinID := toString(prefetchMsg.SkinID)
			baseSkinID := toString(prefetchMsg.BaseSkinID)
			go handlePrefetch(conn, championID, skinID, baseSkinID, prefetchMsg.ChampionName, prefetchMsg.SkinName, prefetchMsg.ChromaName)

		case "cleanup":
			HandleCleanup()

		case "getGamePath":
			path := config.GamePath()
			if path == "" {
				path = game.FindGameDir()
			}
			resp := GamePathMessage{Type: "gamePath", Path: path}
			data, _ := json.Marshal(resp)
			conn.WriteMessage(websocket.TextMessage, data)

		case "setGamePath":
			var msg GamePathMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetGamePath(msg.Path); err != nil {
				sendStatus(conn, "error", "Failed to save game path")
			} else {
				resp := GamePathMessage{Type: "gamePath", Path: msg.Path}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "getSettings":
			s := config.Get()
			roles := config.AutoSelectRoles()
			resp := map[string]interface{}{
				"type":                  "settings",
				"autoAccept":            s.AutoAccept,
				"benchSwap":             s.BenchSwap,
				"benchSwapSkipCooldown": s.BenchSwapSkipCooldown,
				"startWithWindows":      s.StartWithWindows,
				"autoUpdate":            s.AutoUpdate,
				"autoSelect":            s.AutoSelect,
				"autoSelectRoles":       roles,
				"roomParty":             s.RoomParty,
				"chatAvailability":      s.ChatAvailability,
				"chatStatusMessage":     s.ChatStatusMessage,
			}
			data, _ := json.Marshal(resp)
			conn.WriteMessage(websocket.TextMessage, data)

		case "setAutoAccept":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetAutoAccept(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save auto-accept setting")
			} else {
				resp := BoolSettingMessage{Type: "autoAccept", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setBenchSwap":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetBenchSwap(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save bench swap setting")
			} else {
				resp := BoolSettingMessage{Type: "benchSwap", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setBenchSwapSkipCooldown":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetBenchSwapSkipCooldown(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save bench swap skip cooldown setting")
			} else {
				resp := BoolSettingMessage{Type: "benchSwapSkipCooldown", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setStartWithWindows":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			var actionErr error
			if msg.Enabled {
				actionErr = startup.Enable()
			} else {
				actionErr = startup.Disable()
			}
			if actionErr != nil {
				sendStatus(conn, "error", "Failed to register startup task")
			} else if err := config.SetStartWithWindows(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save startup setting")
			} else {
				resp := BoolSettingMessage{Type: "startWithWindows", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setAutoUpdate":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetAutoUpdate(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save auto-update setting")
			} else {
				resp := BoolSettingMessage{Type: "autoUpdate", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setAutoSelect":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetAutoSelect(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save auto-select setting")
			} else {
				resp := BoolSettingMessage{Type: "autoSelect", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setAutoSelectRole":
			var msg AutoSelectRoleMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if msg.Picks == nil {
				msg.Picks = []int{}
			}
			if msg.Bans == nil {
				msg.Bans = []int{}
			}
			if err := config.SetAutoSelectRole(msg.Role, msg.Picks, msg.Bans); err != nil {
				sendStatus(conn, "error", "Failed to save auto-select role config")
			} else {
				resp := AutoSelectRoleMessage{Type: "autoSelectRole", Role: msg.Role, Picks: msg.Picks, Bans: msg.Bans}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setRoomParty":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetRoomParty(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save room party setting")
			} else {
				resp := BoolSettingMessage{Type: "roomParty", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "setChatStatus":
			var msg ChatStatusSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetChatStatus(msg.Availability, msg.StatusMessage); err != nil {
				sendStatus(conn, "error", "Failed to save chat status setting")
			} else {
				resp := ChatStatusSettingMessage{Type: "chatStatus", Availability: msg.Availability, StatusMessage: msg.StatusMessage}
				data, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, data)
			}

		case "roomPartyJoin":
			var msg RoomPartyJoinMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if !config.RoomParty() {
				display.Log("Room Party: join ignored (setting disabled)")
				continue
			}
			roomState.Join(msg.RoomKey, msg.Puuid, msg.TeamPuuids)
			display.Log(fmt.Sprintf("Room Party: joined room %s with %d teammates", msg.RoomKey, len(msg.TeamPuuids)))

		case "roomPartySkin":
			var msg RoomPartySkinMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if !config.RoomParty() {
				continue
			}
			display.Log(fmt.Sprintf("Room Party: own skin updated to %s (%s)", msg.SkinName, toString(msg.SkinID)))
			roomState.UpdateSkin(roomparty.SkinInfo{
				ChampionID:   toString(msg.ChampionID),
				SkinID:       toString(msg.SkinID),
				BaseSkinID:   toString(msg.BaseSkinID),
				ChampionName: msg.ChampionName,
				SkinName:     msg.SkinName,
				ChromaName:   msg.ChromaName,
			})

		case "roomPartyLeave":
			roomState.Leave()
			display.Log("Room Party: left room")

		case "uninstall":
			go handleUninstall(conn)

		case "query":
			stateMu.Lock()
			state := StateMessage{
				Type:          "state",
				ChampionID:    lastChampionID,
				SkinID:        lastSkinID,
				BaseSkinID:    lastBaseSkinID,
				ChampionName:  lastChampionName,
				SkinName:      lastSkinName,
				ChromaName:    lastChromaName,
				OverlayActive: modtools.IsRunning() && lastSkinID != "",
			}
			stateMu.Unlock()
			data, _ := json.Marshal(state)
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// wsHandler handles WebSocket upgrade and connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	handleConnection(conn)
}

// httpHandler handles regular HTTP requests
func httpHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ame server running - connect via ws://localhost:18765"))
}

// StartServer starts the WebSocket server
func StartServer(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			wsHandler(w, r)
		} else {
			httpHandler(w, r)
		}
	})

	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		display.Log(fmt.Sprintf("! Server error: %v", err))
	}
}
