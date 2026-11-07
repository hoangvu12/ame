package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/browsedownload"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/custommods"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/extensions"
	"github.com/hoangvu12/ame/internal/formicons"
	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/httpproxy"
	"github.com/hoangvu12/ame/internal/lcu"
	"github.com/hoangvu12/ame/internal/ltk"
	"github.com/hoangvu12/ame/internal/modtools"
	"github.com/hoangvu12/ame/internal/roomparty"
	"github.com/hoangvu12/ame/internal/setup"
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

// RandomSkinMessage represents a random skin mode get/set
type RandomSkinMessage struct {
	Type string `json:"type"`
	Mode string `json:"mode"`
}

// ChatStatusSettingMessage represents a chat status get/set
type ChatStatusSettingMessage struct {
	Type          string `json:"type"`
	Availability  string `json:"availability"`
	StatusMessage string `json:"statusMessage"`
}

// CustomModImportMessage represents a custom mod import request
type CustomModImportMessage struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	ChampionID  int    `json:"championId"`
	FilePath    string `json:"filePath"`
	ImageBase64 string `json:"imageBase64,omitempty"`
}

// CustomModUpdateMessage represents a custom mod metadata update
type CustomModUpdateMessage struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Author     string `json:"author"`
	ChampionID int    `json:"championId"`
}

// CustomModToggleMessage represents a custom mod enable/disable toggle
type CustomModToggleMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// CustomModIDMessage represents a message with just a mod ID
type CustomModIDMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
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

var clients = make(map[*websocket.Conn]*sync.Mutex)
var clientsMu sync.Mutex

const clientWriteTimeout = 5 * time.Second

// Snapshot membership without holding the registry lock during network I/O.
func connectedClients() []*websocket.Conn {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	connections := make([]*websocket.Conn, 0, len(clients))
	for conn := range clients {
		connections = append(connections, conn)
	}
	return connections
}

// All application writes (including deadline changes) share one lock per client.
// Missing clients are disconnected: never recreate their lock for a late sender.
func writeClientMessage(conn *websocket.Conn, messageType int, data []byte) error {
	clientsMu.Lock()
	mu := clients[conn]
	clientsMu.Unlock()
	if mu == nil {
		return net.ErrClosed
	}
	mu.Lock()
	defer mu.Unlock()
	if err := conn.SetWriteDeadline(time.Now().Add(clientWriteTimeout)); err != nil {
		return err
	}
	return conn.WriteMessage(messageType, data)
}

// Last applied skin state — survives client reconnects
var lastChampionID string
var lastSkinID string
var lastBaseSkinID string
var lastChampionName string
var lastSkinName string
var lastChromaName string
var preparationInvalid atomic.Bool

var lastModKey string
var stateMu sync.Mutex

// Prebuild state — tracks overlay pre-built during champion select
var overlayBuildMu sync.Mutex
var prebuiltModKey string

// Serialize preparation and runtime replacement across prefetch, apply and cleanup.
var skinPreparationMu sync.Mutex
var selectionRevision atomic.Uint64

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
	writeClientMessage(conn, websocket.TextMessage, data)

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
	ltk.RemoveMapCopies()

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
	custommods.CleanModsDir()
	removeOverlayState(config.OverlayDir)
	os.RemoveAll(config.CustomModsDir)

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

// customModKey returns a string suffix representing enabled custom mods, for cache keying.
func customModKey() string {
	mods := config.GetEnabledCustomMods()
	if len(mods) == 0 {
		return ""
	}
	var ids []string
	for _, m := range mods {
		ids = append(ids, m.ID)
	}
	return "cm:" + strings.Join(ids, ",")
}

// handleApply handles skin apply request
func handleApply(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string) {
	skinPreparationMu.Lock()
	defer skinPreparationMu.Unlock()
	if preparationInvalid.Swap(false) {
		stateMu.Lock()
		lastModKey = ""
		stateMu.Unlock()
		prebuiltModKey = ""
	}
	started := time.Now()
	display.Log(fmt.Sprintf("Apply: begin champion=%s skin=%s sharing=%v", championID, skinID, roomState.IsActive()))
	defer func() {
		display.Log(fmt.Sprintf("Apply: handler returned duration_ms=%d", time.Since(started).Milliseconds()))
	}()
	hasSkin := skinID != "" && skinID != "0"
	enabledCustomMods := config.GetEnabledCustomMods()
	hasCustomMods := len(enabledCustomMods) > 0
	hasTeammateSkins := roomState.IsActive() && roomState.HasTeammateSkins()

	// Nothing to apply: default skin with no auxiliary mods.
	if !hasSkin && !hasCustomMods && !hasTeammateSkins {
		sendStatus(conn, "noop", "Nothing to apply")
		return
	}

	// Compute mod key early: includes own skin + teammate skins + custom mods if active
	currentModKey := skinID
	if roomState.IsActive() {
		currentModKey = roomState.ComputeModKey(skinID)
	}
	if cmk := customModKey(); cmk != "" {
		currentModKey += ";" + cmk
	}

	// If runoverlay is already running for this exact mod set, skip — nothing to do
	stateMu.Lock()
	alreadyActive := !preparationInvalid.Load() && ltk.Ready() && modtools.IsRunning() && lastModKey == skin.OverlayKey(currentModKey)
	display.Log(fmt.Sprintf("Apply: modKey=%s lastModKey=%s running=%v alreadyActive=%v", currentModKey, lastModKey, modtools.IsRunning(), alreadyActive))
	stateMu.Unlock()
	if alreadyActive {
		if failure := ltk.Failure(); failure != "" {
			sendStatus(conn, "error", "Skin runtime could not overlay this game. Start a fresh match after selecting the skin.")
			return
		}
		display.SetSkin(skinName, chromaName)
		display.SetOverlayKey("display.value.overlay_active", nil)
		display.Log("Apply: reusing runtime armed during champion select")
		sendStatus(conn, "ready", "Skin prepared for the match")
		return
	}

	// Too late to change what the manager holds for this match (League
	// runs, or champion select has less time left than this PC's rebuild):
	// decided before any download or build, and again before the edit.
	if verdict, reason := lateChange(championID, hasSkin); verdict != changeAllowed {
		reportBlockedChange(conn, "Apply", verdict, reason)
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
		sendStatus(conn, "error", "Skin runtime unavailable. Please restart ame.")
		return
	}

	// Download skin if not default
	display.Log("Apply: preparing own skin")
	resumeScanner, err := ltk.BeginPreparation(config.OverlayDir)
	if err != nil {
		sendStatus(conn, "error", fmt.Sprintf("Could not prepare skin runtime: %v", err))
		return
	}
	defer func() {
		if err := resumeScanner(); err != nil {
			display.Log("Preparation scanner release failed: " + err.Error())
		}
	}()
	var zipPath string
	if hasSkin {
		zipPath = skin.GetValidCachedPath(championID, skinID, baseSkinID)
		if zipPath == "" {
			downloaded, err := skin.Download(championID, skinID, baseSkinID)
			if err != nil {
				sendStatus(conn, "error", fmt.Sprintf("Could not prepare skin: %v", err))
				return
			}
			zipPath = downloaded
		}
	}

	// Build overlay (or reuse pre-built one from prefetch).
	display.Log("Apply: waiting for overlay build lock")
	overlayBuildMu.Lock()
	display.Log("Apply: acquired overlay build lock")

	teammateSkinCount := 0
	if roomState.IsActive() {
		display.Log(fmt.Sprintf("Apply: room party active, mod key: %s", currentModKey))
	} else {
		display.Log("Apply: room party inactive, own skin only")
	}

	prebuilt := !preparationInvalid.Load() && ltk.Ready() && prebuiltModKey == skin.OverlayKey(currentModKey)
	if prebuilt {
		// The prebuilt overlay is the published skin package: the manager
		// builds its own overlay from it when Start installs it.
		if !skinPackageReady(config.OverlayDir + ".fantome") {
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
		if err := refreshModsDir(config.ModsDir, skinID, zipPath); err != nil {
			overlayBuildMu.Unlock()
			sendStatus(conn, "error", "Failed to extract skin archive")
			return
		}

		// Download and extract teammate skins if room party is active
		if roomState.IsActive() {
			display.Log("Apply: preparing teammate skins")
			roomState.DownloadTeammateSkins()
			display.Log("Apply: teammate preparation finished")
		}

		// Create junctions for enabled custom mods
		customModNames := custommods.SetupCustomModJunctions()

		// Build mod list: own skin + teammate skins + custom mods
		var modName string
		if roomState.IsActive() {
			modName = roomState.GetAllModNames(skinID)
		} else if hasSkin {
			modName = fmt.Sprintf("skin_%s", skinID)
		}
		// Append custom mod names (highest priority — last in list)
		if len(customModNames) > 0 {
			if modName != "" {
				modName += "/" + strings.Join(customModNames, "/")
			} else {
				modName = strings.Join(customModNames, "/")
			}
		}

		display.Log(fmt.Sprintf("Apply: building skin package with mods: %s", modName))
		teammateSkinCount = strings.Count(modName, "/")
		if hasCustomMods {
			teammateSkinCount -= len(customModNames)
			if !hasSkin {
				teammateSkinCount = 0
			}
		}

		success, exitCode := modtools.RunMkOverlay(config.ModsDir, config.OverlayDir, gameDir, modName)

		if !success {
			overlayBuildMu.Unlock()
			sendStatus(conn, "error", fmt.Sprintf("Failed to apply skin (code %d)", exitCode))
			return
		}
	}
	overlayBuildMu.Unlock()
	if roomState.IsActive() {
		teammateSkinCount = roomState.BuiltTeammateSkinCount()
	}

	// The last look before the manager's library changes: downloads and the
	// package build may have taken seconds.
	if verdict, reason := lateChange(championID, hasSkin); verdict != changeAllowed {
		reportBlockedChange(conn, "Apply", verdict, reason)
		return
	}

	// Hand the package to the manager (its patcher hooks the game when it
	// finds it)
	err = startRuntime(config.OverlayDir)
	if errors.Is(err, ltk.ErrGameRunning) {
		verdict, reason := lateChange(championID, hasSkin)
		reportBlockedChange(conn, "Apply", verdict, reason)
		return
	}
	setPreparedChampion(championID, hasSkin)
	if err != nil {
		preparationInvalid.Store(true)
		stateMu.Lock()
		lastModKey = ""
		stateMu.Unlock()
		prebuiltModKey = ""
		if errors.Is(err, ltk.ErrRebuildPending) {
			sendStatus(conn, "error", "Skins are still being prepared (this PC rebuilds them slowly). They apply this match only if that finishes before the game starts.")
			return
		}
		sendStatus(conn, "error", fmt.Sprintf("Failed to start overlay: %v", err))
		return
	}

	// Track last applied state — use the actual built key, not the theoretical one,
	// so that a later apply with new teammates isn't short-circuited.
	actualModKey := skinID
	if roomState.IsActive() {
		actualModKey = roomState.ComputeBuiltModKey(skinID)
		// Fallback: if teammates were part of the requested mod key but
		// ComputeBuiltModKey didn't pick them up (e.g., timing/race), store the
		// requested key so we don't thrash with redundant rebuilds.
		if strings.Contains(currentModKey, ",") && !strings.Contains(actualModKey, ",") {
			actualModKey = currentModKey
		}
	}
	if cmk := customModKey(); cmk != "" {
		actualModKey += ";" + cmk
	}
	stateMu.Lock()
	lastChampionID = championID
	lastSkinID = skinID
	lastBaseSkinID = baseSkinID
	lastChampionName = championName
	lastSkinName = skinName
	lastChromaName = chromaName
	lastModKey = skin.OverlayKey(actualModKey)
	stateMu.Unlock()
	display.Log(fmt.Sprintf("Apply: stored lastModKey=%s (wanted=%s)", actualModKey, currentModKey))

	display.SetSkin(skinName, chromaName)
	display.SetOverlayKey("display.value.overlay_active", nil)

	// "Prepared": the manager scans with the set. Only the patcher DLL's
	// verdict on the game (logged as "LTK: overlay verdict") confirms skins.
	if teammateSkinCount > 0 {
		sendStatus(conn, "ready", fmt.Sprintf("Skin prepared for the match (+%d teammate skins)", teammateSkinCount))
	} else if !hasSkin && hasCustomMods {
		sendStatus(conn, "ready", "Custom mods prepared for the match")
	} else {
		sendStatus(conn, "ready", "Skin prepared for the match")
	}
}

// handlePrefetch pre-downloads a skin and pre-builds the overlay during champion select
func handlePrefetch(conn *websocket.Conn, championID, skinID, baseSkinID, championName, skinName, chromaName string, revision uint64) {
	skinPreparationMu.Lock()
	defer skinPreparationMu.Unlock()
	if preparationInvalid.Swap(false) {
		stateMu.Lock()
		lastModKey = ""
		stateMu.Unlock()
		prebuiltModKey = ""
	}
	if revision != selectionRevision.Load() {
		display.Log("Prefetch: skipped superseded queued selection")
		return
	}
	hasSkin := skinID != "" && skinID != "0"
	hasCustomMods := len(config.GetEnabledCustomMods()) > 0
	hasTeammateSkins := roomState.IsActive() && roomState.HasTeammateSkins()

	// Nothing to prefetch: disarm a previous selection before a game starts.
	// Never while League runs: stopping the patcher cannot take skins out of
	// a match that loaded them. The manager stays for the next selection.
	if !hasSkin && !hasCustomMods && !hasTeammateSkins {
		if patcherWaiting() && !gameRunning() {
			disarmRuntime()
			stateMu.Lock()
			lastModKey = ""
			stateMu.Unlock()
		}
		return
	}

	// Too late to change what the manager holds for this match: decided
	// before any download or build, and again before the edit.
	if verdict, reason := lateChange(championID, hasSkin); verdict != changeAllowed {
		reportBlockedChange(conn, "Prefetch", verdict, reason)
		return
	}

	resumeScanner, err := ltk.BeginPreparation(config.OverlayDir)
	if err != nil {
		display.Log("Prefetch: cannot prepare scanner: " + err.Error())
		return
	}
	defer func() {
		if err := resumeScanner(); err != nil {
			display.Log("Preparation scanner release failed: " + err.Error())
		}
	}()

	// Download skin if not default
	var zipPath string
	if hasSkin {
		zipPath = skin.GetValidCachedPath(championID, skinID, baseSkinID)
		if zipPath == "" {
			downloaded, err := skin.Download(championID, skinID, baseSkinID)
			if err != nil {
				return
			}
			zipPath = downloaded
		}
	}

	if revision != selectionRevision.Load() {
		display.Log("Prefetch: superseded during generation; keeping hold for queued selection")
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
	if cmk := customModKey(); cmk != "" {
		currentModKey += ";" + cmk
	}

	// Skip if already pre-built for this exact set of skins
	if !preparationInvalid.Load() && ltk.Ready() && prebuiltModKey == skin.OverlayKey(currentModKey) && modtools.IsRunning() {
		display.Log(fmt.Sprintf("Prefetch: skipped (already built), modKey=%s", currentModKey))
		return
	}
	display.Log(fmt.Sprintf("Prefetch: modKey=%s (was %s), roomActive=%v", currentModKey, prebuiltModKey, roomState.IsActive()))

	if err := refreshModsDir(config.ModsDir, skinID, zipPath); err != nil {
		display.Log("Prefetch: skin extraction failed: " + err.Error())
		return
	}

	// Download and extract teammate skins if room party is active
	if roomState.IsActive() {
		roomState.DownloadTeammateSkins()
	}

	// Create junctions for enabled custom mods
	customModNames := custommods.SetupCustomModJunctions()

	// Build mod list: own skin + teammate skins + custom mods
	var modName string
	if roomState.IsActive() {
		modName = roomState.GetAllModNames(skinID)
	} else if hasSkin {
		modName = fmt.Sprintf("skin_%s", skinID)
	}
	if len(customModNames) > 0 {
		if modName != "" {
			modName += "/" + strings.Join(customModNames, "/")
		} else {
			modName = strings.Join(customModNames, "/")
		}
	}

	display.Log(fmt.Sprintf("Prefetch: building skin package with mods: %s", modName))

	success, exitCode := modtools.RunMkOverlay(config.ModsDir, config.OverlayDir, gameDir, modName)
	if !success {
		display.Log(fmt.Sprintf("Prefetch: mkoverlay failed (code %d)", exitCode))
		return
	}

	// Record what was actually built
	builtKey := skinID
	if roomState.IsActive() {
		builtKey = roomState.ComputeBuiltModKey(skinID)
	}
	if cmk := customModKey(); cmk != "" {
		builtKey += ";" + cmk
	}
	builtKey = skin.OverlayKey(builtKey)

	// The last look before the manager's library changes. A kept set leaves
	// the bookkeeping on the armed set: only a package the manager accepted
	// becomes the prebuilt one.
	if verdict, reason := lateChange(championID, hasSkin); verdict != changeAllowed {
		reportBlockedChange(conn, "Prefetch", verdict, reason)
		return
	}
	err = startRuntime(config.OverlayDir)
	if errors.Is(err, ltk.ErrGameRunning) {
		verdict, reason := lateChange(championID, hasSkin)
		reportBlockedChange(conn, "Prefetch", verdict, reason)
		return
	}
	setPreparedChampion(championID, hasSkin)
	if err != nil {
		preparationInvalid.Store(true)
		stateMu.Lock()
		lastModKey = ""
		stateMu.Unlock()
		prebuiltModKey = ""
		if errors.Is(err, ltk.ErrRebuildPending) {
			display.Log("Prefetch: the manager is still rebuilding this skin set; it is armed only if the rebuild ends before League starts")
			return
		}
		display.Log(fmt.Sprintf("Prefetch: runtime startup failed: %v", err))
		return
	}
	prebuiltModKey = builtKey
	stateMu.Lock()
	lastModKey = prebuiltModKey
	lastChampionID, lastSkinID, lastBaseSkinID = championID, skinID, baseSkinID
	lastChampionName, lastSkinName, lastChromaName = championName, skinName, chromaName
	stateMu.Unlock()
	display.Log("Prefetch: runtime scanner started for prepared overlay")
	display.Log(fmt.Sprintf("Skin ready (prebuilt key: %s)", prebuiltModKey))
}

// HandleCleanup stops the skin runtime, manager included, and forgets the
// prepared selection. It is unconditional: uninstall, the stuck-game button
// and shutting ame down call it directly. The plugin's cleanup message goes
// through handleCleanupRequest, which declines while League runs and only
// disarms the patcher.
func HandleCleanup() {
	resetSelection(stopRuntime)
}

// resetSelection runs stop (a full stop or a disarm) and forgets the
// prepared selection.
func resetSelection(stop func()) {
	skinPreparationMu.Lock()
	defer skinPreparationMu.Unlock()
	if preparationInvalid.Swap(false) {
		stateMu.Lock()
		lastModKey = ""
		stateMu.Unlock()
		prebuiltModKey = ""
	}
	stop()
	removeOverlayState(config.OverlayDir)

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
	preparedChampionID, preparedOwnSkin = "", false
	stateMu.Unlock()

	display.SetSkin("", "")
	display.SetOverlayKey("display.value.overlay_inactive", nil)
}

// handleUnstuck ends a stuck game at the user's explicit request (the
// in-game "I'm stuck" button). ame never pauses the game anymore, so there
// is nothing to resume first.
func handleUnstuck(conn *websocket.Conn) {
	started := time.Now()
	defer func() {
		display.Log(fmt.Sprintf("Unstuck: handler returned duration_ms=%d", time.Since(started).Milliseconds()))
	}()
	display.Log("Unstuck: ending the game at the user's request")

	// Kill the game process
	display.Log("Unstuck: attempting game termination")
	hiddenAttr := &syscall.SysProcAttr{HideWindow: true}
	kill := exec.Command("taskkill", "/F", "/IM", "League of Legends.exe")
	kill.SysProcAttr = hiddenAttr
	killErr := kill.Run()

	// Cleanup overlay state regardless
	HandleCleanup()

	if killErr != nil {
		display.Log(fmt.Sprintf("Unstuck: taskkill failed: %v (process may already be dead)", killErr))
	} else {
		display.Log("Unstuck: game process killed")
	}

	sendStatus(conn, "ready", "Game released")
}

// broadcastRoomUpdate sends room party teammate info to all connected clients.
func broadcastRoomUpdate(teammates []roomparty.Member) {
	msg := RoomPartyUpdateMessage{
		Type:      "roomPartyUpdate",
		Teammates: teammates,
	}
	data, _ := json.Marshal(msg)

	for _, conn := range connectedClients() {
		writeClientMessage(conn, websocket.TextMessage, data)
	}
}

// broadcastStatus sends a status message to every connected client.
func broadcastStatus(status, message string) {
	for _, conn := range connectedClients() {
		sendStatus(conn, status, message)
	}
}

func init() {
	roomState.OnUpdate = broadcastRoomUpdate
	ltk.OnOverlayVerdict = reportOverlayVerdict
	ltk.OnMapPrep = broadcastMapPrep
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
			display.SetStatusKey("display.value.waiting_client", nil)
		}
		display.Log("Client disconnected")
	}()
	clientsMu.Lock()
	clients[conn] = &sync.Mutex{}
	clientsMu.Unlock()
	display.SetStatusKey("display.value.connected", nil)
	display.Log("Client connected")

	// Tell the client where startup got to, so a plugin that connects during
	// setup knows to wait rather than showing itself as fully operational.
	readyMsg, _ := json.Marshal(map[string]interface{}{"type": "ready", "ready": ready.Load()})
	writeClientMessage(conn, websocket.TextMessage, readyMsg)
	sendMapPrep(conn)
	// The League client just started or reloaded, often right after an
	// update replaced the maps.
	scheduleMapPrep("client connected", mapPrepClientDelay)
	warmManagerSoon("client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		// Hold back anything that needs mod-tools or the game install until
		// setup has finished. The plugin is allowed to connect and read state
		// well before that point, so it can be told what is going on.
		if commandsNeedingSetup[incoming.Type] && !ready.Load() {
			sendStatus(conn, "starting", "ame is still starting up")
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
			selectionRevision.Add(1)
			handleApply(conn, championID, skinID, baseSkinID, applyMsg.ChampionName, applyMsg.SkinName, applyMsg.ChromaName)

		case "prefetch":
			var prefetchMsg ApplyMessage
			if err := json.Unmarshal(message, &prefetchMsg); err != nil {
				continue
			}
			championID := toString(prefetchMsg.ChampionID)
			skinID := toString(prefetchMsg.SkinID)
			baseSkinID := toString(prefetchMsg.BaseSkinID)
			revision := selectionRevision.Add(1)
			go handlePrefetch(conn, championID, skinID, baseSkinID, prefetchMsg.ChampionName, prefetchMsg.SkinName, prefetchMsg.ChromaName, revision)

		case "cleanup":
			handleCleanupRequest(conn)

		case "getGamePath":
			path := config.GamePath()
			if path == "" {
				path = game.FindGameDir()
			}
			resp := GamePathMessage{Type: "gamePath", Path: path}
			data, _ := json.Marshal(resp)
			writeClientMessage(conn, websocket.TextMessage, data)

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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				"randomSkin":            s.RandomSkin,
				"mapPrep":               s.MapPrep,
			}
			data, _ := json.Marshal(resp)
			writeClientMessage(conn, websocket.TextMessage, data)

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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
			}

		case "setMapPrep":
			var msg BoolSettingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetMapPrep(msg.Enabled); err != nil {
				sendStatus(conn, "error", "Failed to save map preparation setting")
			} else {
				resp := BoolSettingMessage{Type: "mapPrep", Enabled: msg.Enabled}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
				// Off: later sets stop keeping the copies, and the manager
				// deletes them on its next build. On: prepare them soon.
				ltk.RefreshMapPrepSetting()
				if msg.Enabled {
					scheduleMapPrep("setting enabled", mapPrepStartupDelay)
				}
			}

		case "setRandomSkin":
			var msg RandomSkinMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := config.SetRandomSkin(msg.Mode); err != nil {
				sendStatus(conn, "error", "Failed to save random skin setting")
			} else {
				resp := RandomSkinMessage{Type: "randomSkin", Mode: msg.Mode}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
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
				writeClientMessage(conn, websocket.TextMessage, data)
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

		case "unstuck":
			go handleUnstuck(conn)

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
			writeClientMessage(conn, websocket.TextMessage, data)

		case "getLogs":
			logs := display.GetLogsJSON()
			version := display.GetVersion()
			resp := map[string]interface{}{
				"type":    "logs",
				"version": version,
				"entries": logs,
			}
			data, _ := json.Marshal(resp)
			writeClientMessage(conn, websocket.TextMessage, data)

		case "pickCustomModFile":
			go func() {
				result, err := custommods.PickModFile()
				if err != nil {
					sendStatus(conn, "error", "Failed to open file picker")
					return
				}
				if result == nil {
					return // User cancelled
				}
				resp := map[string]interface{}{
					"type":                "customModFilePicked",
					"filePath":            result.FilePath,
					"suggestedName":       result.SuggestedName,
					"suggestedAuthor":     result.SuggestedAuthor,
					"suggestedChampionId": result.SuggestedChampionID,
				}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
			}()

		case "importCustomMod":
			var msg CustomModImportMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			go func() {
				mod, err := custommods.ImportMod(msg.FilePath, msg.Name, msg.Author, msg.ChampionID, msg.ImageBase64)
				if err != nil {
					sendStatus(conn, "error", fmt.Sprintf("Failed to import mod: %v", err))
					return
				}
				resp := map[string]interface{}{
					"type": "customModAdded",
					"mod":  mod,
				}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
			}()

		case "getCustomMods":
			mods := config.GetCustomMods()
			if mods == nil {
				mods = []config.CustomMod{}
			}
			resp := map[string]interface{}{
				"type": "customMods",
				"mods": mods,
			}
			data, _ := json.Marshal(resp)
			writeClientMessage(conn, websocket.TextMessage, data)

		case "deleteCustomMod":
			var msg CustomModIDMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := custommods.DeleteMod(msg.ID); err != nil {
				sendStatus(conn, "error", "Failed to delete mod")
			} else {
				resp := map[string]interface{}{
					"type": "customModDeleted",
					"id":   msg.ID,
				}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
			}

		case "updateCustomMod":
			var msg CustomModUpdateMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			mod, err := config.UpdateCustomMod(msg.ID, msg.Name, msg.Author, msg.ChampionID)
			if err != nil {
				sendStatus(conn, "error", "Failed to update mod")
			} else if mod != nil {
				resp := map[string]interface{}{
					"type": "customModUpdated",
					"mod":  mod,
				}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
			}

		case "toggleCustomMod":
			var msg CustomModToggleMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			mod, err := config.ToggleCustomMod(msg.ID, msg.Enabled)
			if err != nil {
				sendStatus(conn, "error", "Failed to toggle mod")
			} else if mod != nil {
				resp := map[string]interface{}{
					"type": "customModUpdated",
					"mod":  mod,
				}
				data, _ := json.Marshal(resp)
				writeClientMessage(conn, websocket.TextMessage, data)
			}

		case "pickCustomModImage":
			var msg CustomModIDMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			go func() {
				saved, err := custommods.SaveImageFromPicker(msg.ID)
				if err != nil {
					sendStatus(conn, "error", "Failed to save image")
					return
				}
				if !saved {
					return // User cancelled
				}
				// Re-read the mod to get updated state
				for _, m := range config.GetCustomMods() {
					if m.ID == msg.ID {
						resp := map[string]interface{}{
							"type": "customModUpdated",
							"mod":  m,
						}
						data, _ := json.Marshal(resp)
						writeClientMessage(conn, websocket.TextMessage, data)
						break
					}
				}
			}()

		case "deleteCustomModImage":
			var msg CustomModIDMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			if err := custommods.DeleteImage(msg.ID); err != nil {
				sendStatus(conn, "error", "Failed to delete image")
			} else {
				for _, m := range config.GetCustomMods() {
					if m.ID == msg.ID {
						resp := map[string]interface{}{
							"type": "customModUpdated",
							"mod":  m,
						}
						data, _ := json.Marshal(resp)
						writeClientMessage(conn, websocket.TextMessage, data)
						break
					}
				}
			}

		case "httpProxy":
			httpproxy.HandleProxy(conn, message)

		case "browseDownload":
			browsedownload.HandleDownload(conn, message)

		case "getExtensions":
			extensions.HandleGetExtensions(conn)

		case "addExtension":
			extensions.HandleAddExtension(conn, message)

		case "addExtensionFromFile":
			extensions.HandleAddExtensionFromFile(conn)

		case "removeExtension":
			extensions.HandleRemoveExtension(conn, message)
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

// ShowAck is the exact body served from /show. A second ame process checks for
// it so it can distinguish a running ame from an unrelated program that
// happens to hold the port.
const ShowAck = "ame-show-ok"

// OnShowRequested is invoked when another ame process asks this instance to
// reveal itself. Set by main.
var OnShowRequested func()

// showHandler lets a second ame instance surface the one already running,
// instead of starting a duplicate that would fight over settings, the overlay
// directory and mod-tools.
//
// focus=0 asks only for the acknowledgement, so a second launch that was itself
// started minimized (the logon task firing while ame already runs) can confirm
// an instance exists without yanking its console into view.
func showHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("focus") != "0" && OnShowRequested != nil {
		OnShowRequested()
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ShowAck))
}

// ready reports whether startup has progressed far enough to run commands that
// depend on mod-tools and the game install being in place.
var ready atomic.Bool

// SetReady marks setup as complete, unblocking the gated commands.
func SetReady() {
	ready.Store(true)
	broadcastReady()
	if clientRunning() {
		warmManagerSoon("startup")
	}
	scheduleMapPrep("startup", mapPrepStartupDelay)
}

// IsReady reports whether gated commands are currently accepted.
func IsReady() bool { return ready.Load() }

// commandsNeedingSetup lists the commands that touch mod-tools, the overlay
// directory or the game install. The server now starts listening before setup
// runs so the plugin can connect immediately, which means these have to be
// held back until the things they depend on actually exist.
var commandsNeedingSetup = map[string]bool{
	"apply":           true,
	"prefetch":        true,
	"cleanup":         true,
	"unstuck":         true,
	"roomPartySkin":   true,
	"importCustomMod": true,
}

// broadcastReady tells every connected client that setup has finished.
func broadcastReady() {
	payload := map[string]interface{}{"type": "ready", "ready": true}
	data, _ := json.Marshal(payload)

	for _, conn := range connectedClients() {
		writeClientMessage(conn, websocket.TextMessage, data)
	}
}

// newMux builds the HTTP routes. It uses a dedicated mux rather than
// http.DefaultServeMux so that retrying a failed bind cannot panic on
// duplicate route registration.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/custom-mod-image/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/custom-mod-image/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		custommods.ServeImage(w, r, id)
	})

	mux.HandleFunc("/form-icon/", formicons.ServeFormIcon)
	mux.HandleFunc("/proxy-image", httpproxy.ServeImageProxy)
	mux.HandleFunc("/show", showHandler)
	mux.HandleFunc("/diagnostics.zip", handleDiagnosticsDownload)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			wsHandler(w, r)
		} else {
			httpHandler(w, r)
		}
	})

	return mux
}

// listenTimeout bounds how long StartServer waits for the port to free up.
// It exists for the update handoff, where the outgoing core's established
// connections can hold the port in TIME_WAIT for a moment after it exits.
const listenTimeout = 10 * time.Second

// StartServer binds the loopback port and serves until the process exits.
//
// It binds explicitly to 127.0.0.1. Binding all interfaces exposed the whole
// command surface — including uninstall, which deletes directories and rewrites
// HKLM keys — to anyone on the same network.
//
// A bind failure is returned rather than logged and forgotten: the port is how
// the plugin reaches ame, so a process that cannot bind it is useless and must
// not keep running as a silent duplicate.
func StartServer(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	var ln net.Listener
	var err error
	deadline := time.Now().Add(listenTimeout)
	for {
		ln, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("listen on %s: %w", addr, err)
		}
		time.Sleep(250 * time.Millisecond)
	}

	srv := &http.Server{Handler: newMux()}
	return srv.Serve(ln)
}

// skinPackageReady reports whether the published skin package for an overlay
// prefix exists and holds content. The manager's patcher installs this
// package; the extracted overlay it used to sit beside is no longer built.
func skinPackageReady(pkg string) bool {
	info, err := os.Stat(pkg)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

// removeOverlayState removes the published overlay artifacts: the legacy
// overlay directory, the skin package, and any package a failed build left
// staged.
func removeOverlayState(overlay string) {
	os.RemoveAll(overlay)
	os.Remove(overlay + ".fantome")
	os.Remove(overlay + ".fantome.new")
}

// refreshModsDir prunes mod directories the current selection no longer uses
// and refreshes the own skin's mod directory only when its package bytes
// changed, so re-applying the same selection costs no re-extraction and no
// voice-bank relocalization. Teammate mods are refreshed (also digest-gated)
// by DownloadTeammateSkins right after this runs.
func refreshModsDir(modsDir, skinID, zipPath string) error {
	keep := map[string]bool{}
	hasSkin := skinID != "" && skinID != "0"
	if hasSkin {
		keep[fmt.Sprintf("skin_%s", skinID)] = true
	}
	if roomState.IsActive() {
		for _, tm := range roomState.GetTeammates() {
			if tm.SkinInfo.SkinID != "" && tm.SkinInfo.SkinID != "0" {
				keep[fmt.Sprintf("skin_%s", tm.SkinInfo.SkinID)] = true
			}
		}
	}
	for _, mod := range config.GetEnabledCustomMods() {
		keep["custom_"+mod.ID] = true
	}
	entries, err := os.ReadDir(modsDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, entry := range entries {
		if keep[entry.Name()] {
			continue
		}
		entryPath := filepath.Join(modsDir, entry.Name())
		// Junctions (custom mods) unlink without touching their target.
		if entry.Type()&os.ModeSymlink != 0 {
			_ = os.Remove(entryPath)
		} else {
			_ = os.RemoveAll(entryPath)
		}
	}
	if !hasSkin {
		return nil
	}
	if err := os.MkdirAll(modsDir, os.ModePerm); err != nil {
		return err
	}
	return skin.ExtractPackage(zipPath, filepath.Join(modsDir, fmt.Sprintf("skin_%s", skinID)))
}
