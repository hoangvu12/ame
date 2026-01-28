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
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/modtools"
	"github.com/hoangvu12/ame/internal/skin"
)

var (
	AME_DIR     = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame")
	MODS_DIR    = filepath.Join(AME_DIR, "mods")
	OVERLAY_DIR = filepath.Join(AME_DIR, "overlay")
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

// sendStatus sends a status message to the WebSocket client
func sendStatus(conn *websocket.Conn, status, message string) {
	payload := StatusMessage{
		Type:    "status",
		Status:  status,
		Message: message,
	}
	data, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, data)
	fmt.Printf("[ame] %s: %s\n", status, message)
}

// handleApply handles skin apply request
func handleApply(conn *websocket.Conn, championID, skinID, baseSkinID string) {
	// Find game directory
	gameDir := game.FindGameDir()
	if gameDir == "" {
		sendStatus(conn, "error", "League of Legends Game directory not found")
		return
	}
	fmt.Printf("[ame] Game dir: %s\n", gameDir)

	// Check mod-tools exists
	if !modtools.Exists() {
		sendStatus(conn, "error", "mod-tools.exe not found. Please restart ame.")
		return
	}

	// Check for cached skin file
	zipPath := skin.GetCachedPath(championID, skinID)

	// Download if not cached
	if zipPath == "" {
		sendStatus(conn, "downloading", "Downloading skin...")
		downloaded, err := skin.Download(championID, skinID, baseSkinID)
		if err != nil {
			sendStatus(conn, "error", "Skin not available for download")
			return
		}
		zipPath = downloaded
	} else {
		fmt.Printf("[ame] Skin %s already cached\n", skinID)
	}

	// Kill any previous runoverlay
	modtools.KillModTools()
	time.Sleep(300 * time.Millisecond)

	// Clean and extract to mods dir
	os.RemoveAll(MODS_DIR)
	modSubDir := filepath.Join(MODS_DIR, fmt.Sprintf("skin_%s", skinID))
	os.MkdirAll(modSubDir, os.ModePerm)

	if err := skin.Extract(zipPath, modSubDir); err != nil {
		sendStatus(conn, "error", "Failed to extract skin archive")
		return
	}

	// Clean overlay dir
	os.RemoveAll(OVERLAY_DIR)
	os.MkdirAll(OVERLAY_DIR, os.ModePerm)

	// Run mkoverlay
	sendStatus(conn, "injecting", "Applying skin...")
	modName := fmt.Sprintf("skin_%s", skinID)
	success, exitCode := modtools.RunMkOverlay(MODS_DIR, OVERLAY_DIR, gameDir, modName)

	if !success {
		sendStatus(conn, "error", fmt.Sprintf("Failed to apply skin (exit code %d)", exitCode))
		return
	}

	// Run runoverlay
	configPath := filepath.Join(OVERLAY_DIR, "cslol-config.json")
	modtools.RunOverlay(OVERLAY_DIR, configPath, gameDir)

	sendStatus(conn, "ready", "Skin applied!")
}

// HandleCleanup handles cleanup request
func HandleCleanup() {
	fmt.Println("[ame] Cleanup: stopping runoverlay")
	modtools.KillModTools()
	os.RemoveAll(OVERLAY_DIR)
}

// handleConnection handles a single WebSocket connection
func handleConnection(conn *websocket.Conn) {
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
		fmt.Println("[ame] Client disconnected")
	}()

	fmt.Println("[ame] Client connected")
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse message type first
		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			fmt.Printf("[ame] Message parse error: %v\n", err)
			continue
		}

		switch incoming.Type {
		case "apply":
			var applyMsg ApplyMessage
			if err := json.Unmarshal(message, &applyMsg); err != nil {
				fmt.Printf("[ame] Apply message parse error: %v\n", err)
				continue
			}
			championID := toString(applyMsg.ChampionID)
			skinID := toString(applyMsg.SkinID)
			baseSkinID := toString(applyMsg.BaseSkinID)
			fmt.Printf("[ame] Apply requested: champion=%s skin=%s base=%s\n",
				championID, skinID, baseSkinID)
			handleApply(conn, championID, skinID, baseSkinID)

		case "cleanup":
			HandleCleanup()

		default:
			fmt.Printf("[ame] Unknown message type: %s\n", incoming.Type)
		}
	}
}

// wsHandler handles WebSocket upgrade and connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[ame] WebSocket upgrade error: %v\n", err)
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
		// Check if it's a WebSocket upgrade request
		if r.Header.Get("Upgrade") == "websocket" {
			wsHandler(w, r)
		} else {
			httpHandler(w, r)
		}
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("[ame] Listening on ws://localhost%s\n", addr)
	fmt.Println("[ame] Open League client to see the skin selector.")
	fmt.Println("[ame] Press Ctrl+C to stop.")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("[ame] Server error: %v\n", err)
	}
}
