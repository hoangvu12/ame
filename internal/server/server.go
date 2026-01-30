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
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/modtools"
	"github.com/hoangvu12/ame/internal/skin"
)

// ApplyMessage represents an apply skin request
type ApplyMessage struct {
	Type       string      `json:"type"`
	ChampionID interface{} `json:"championId"`
	SkinID     interface{} `json:"skinId"`
	BaseSkinID interface{} `json:"baseSkinId,omitempty"`
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
	OverlayActive bool   `json:"overlayActive"`
}

// GamePathMessage represents a game path request/response
type GamePathMessage struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// AutoAcceptMessage represents an auto-accept setting request/response
type AutoAcceptMessage struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
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

	// Only log important status changes
	switch status {
	case "ready":
		fmt.Printf("  > %s\n", message)
	case "error":
		fmt.Printf("  ! %s\n", message)
	}
}

// handleApply handles skin apply request
func handleApply(conn *websocket.Conn, championID, skinID, baseSkinID string) {
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
		downloaded, err := skin.Download(championID, skinID, baseSkinID)
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
	stateMu.Unlock()

	sendStatus(conn, "ready", "Skin applied!")
}

// handlePrefetch pre-downloads a skin to cache (no mkoverlay/runoverlay)
func handlePrefetch(conn *websocket.Conn, championID, skinID, baseSkinID string) {
	// Check for cached skin file — if already cached, nothing to do
	if skin.GetCachedPath(championID, skinID) != "" {
		return
	}

	skin.Download(championID, skinID, baseSkinID)
}

// HandleCleanup handles cleanup request
func HandleCleanup() {
	modtools.KillModTools()
	os.RemoveAll(config.OverlayDir)

	stateMu.Lock()
	lastChampionID = ""
	lastSkinID = ""
	lastBaseSkinID = ""
	stateMu.Unlock()
}

// handleConnection handles a single WebSocket connection
func handleConnection(conn *websocket.Conn) {
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

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
			handleApply(conn, championID, skinID, baseSkinID)

		case "prefetch":
			var prefetchMsg ApplyMessage
			if err := json.Unmarshal(message, &prefetchMsg); err != nil {
				continue
			}
			championID := toString(prefetchMsg.ChampionID)
			skinID := toString(prefetchMsg.SkinID)
			baseSkinID := toString(prefetchMsg.BaseSkinID)
			go handlePrefetch(conn, championID, skinID, baseSkinID)

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

		case "getAutoAccept":
			resp := AutoAcceptMessage{Type: "autoAccept", Enabled: config.AutoAccept()}
			data, _ := json.Marshal(resp)
			conn.WriteMessage(websocket.TextMessage, data)

		case "setAutoAccept":
			var msg AutoAcceptMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetAutoAccept(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save auto-accept setting")
			} else {
				resp := AutoAcceptMessage{Type: "autoAccept", Enabled: msg.Enabled}
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
		fmt.Printf("  ! Server error: %v\n", err)
	}
}
