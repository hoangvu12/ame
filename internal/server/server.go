package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/modtools"
	"github.com/hoangvu12/ame/internal/skin"
	"github.com/hoangvu12/ame/internal/startup"
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
var stateMu sync.Mutex

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

// handleApply handles skin apply request
func handleApply(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string) {
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

	// Clean and extract to mods dir
	os.RemoveAll(config.ModsDir)
	modSubDir := filepath.Join(config.ModsDir, fmt.Sprintf("skin_%s", skinID))
	os.MkdirAll(modSubDir, os.ModePerm)

	if err := skin.Extract(zipPath, modSubDir); err != nil {
		sendStatus(conn, "error", "Failed to extract skin archive")
		return
	}

	// Clean overlay dir
	os.RemoveAll(config.OverlayDir)
	os.MkdirAll(config.OverlayDir, os.ModePerm)

	// Run mkoverlay
	modName := fmt.Sprintf("skin_%s", skinID)
	success, exitCode := modtools.RunMkOverlay(config.ModsDir, config.OverlayDir, gameDir, modName)

	if !success {
		sendStatus(conn, "error", fmt.Sprintf("Failed to apply skin (exit code %d)", exitCode))
		return
	}

	// Run runoverlay
	configPath := filepath.Join(config.OverlayDir, "cslol-config.json")
	if err := modtools.RunOverlay(config.OverlayDir, configPath, gameDir); err != nil {
		sendStatus(conn, "error", fmt.Sprintf("Failed to start overlay: %v", err))
		return
	}

	// Track last applied state
	stateMu.Lock()
	lastChampionID = championID
	lastSkinID = skinID
	lastBaseSkinID = baseSkinID
	lastChampionName = championName
	lastSkinName = skinName
	lastChromaName = chromaName
	stateMu.Unlock()

	display.SetSkin(skinName, chromaName)
	display.SetOverlay("Active")

	sendStatus(conn, "ready", "Skin applied!")
}

// handlePrefetch pre-downloads a skin to cache (no mkoverlay/runoverlay)
func handlePrefetch(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string) {
	// Check for cached skin file — if already cached, nothing to do
	if skin.GetCachedPath(championID, skinID) != "" {
		return
	}

	skin.Download(championID, skinID, baseSkinID, championName, skinName, chromaName)
}

// HandleCleanup handles cleanup request
func HandleCleanup() {
	modtools.KillModTools()
	os.RemoveAll(config.OverlayDir)

	stateMu.Lock()
	lastChampionID = ""
	lastSkinID = ""
	lastBaseSkinID = ""
	lastChampionName = ""
	lastSkinName = ""
	lastChromaName = ""
	stateMu.Unlock()

	display.SetSkin("", "")
	display.SetOverlay("Inactive")
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
		display.Log(fmt.Sprintf("Client disconnected (%d active)", remaining))
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
				"type":             "settings",
				"autoAccept":       s.AutoAccept,
				"benchSwap":             s.BenchSwap,
			"benchSwapSkipCooldown": s.BenchSwapSkipCooldown,
				"startWithWindows": s.StartWithWindows,
				"autoUpdate":       s.AutoUpdate,
				"autoSelect":       s.AutoSelect,
				"autoSelectRoles":  roles,
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
