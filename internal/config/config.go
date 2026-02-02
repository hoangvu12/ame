package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Paths - single source of truth (previously duplicated across packages)
var (
	AmeDir     = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame")
	ToolsDir   = filepath.Join(AmeDir, "tools")
	SkinsDir   = filepath.Join(AmeDir, "skins")
	ModsDir    = filepath.Join(AmeDir, "mods")
	OverlayDir = filepath.Join(AmeDir, "overlay")
	PenguDir   = filepath.Join(AmeDir, "pengu")
)

var (
	settingsPath = filepath.Join(AmeDir, "settings.json")
	legacyPath   = filepath.Join(AmeDir, "gamedir.txt")

	mu       sync.RWMutex
	settings Settings
)

// RoleConfig holds the pick/ban champion priority lists for a single role.
type RoleConfig struct {
	Picks []int `json:"picks"`
	Bans  []int `json:"bans"`
}

// Settings holds persisted application settings.
type Settings struct {
	GamePath         string                `json:"gamePath"`
	AutoAccept       bool                  `json:"autoAccept"`
	BenchSwap             bool                  `json:"benchSwap"`
	BenchSwapSkipCooldown bool                  `json:"benchSwapSkipCooldown"`
	StartWithWindows bool                  `json:"startWithWindows"`
	AutoUpdate       bool                  `json:"autoUpdate"`
	AutoSelect       bool                  `json:"autoSelect"`
	AutoSelectRoles   map[string]RoleConfig `json:"autoSelectRoles"`
	RoomParty         bool                  `json:"roomParty"`
	ChatAvailability  string                `json:"chatAvailability"`
	ChatStatusMessage string                `json:"chatStatusMessage"`
}

// Init loads settings from disk.
// It migrates from the legacy gamedir.txt file if settings.json does not exist.
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	// Set defaults before loading (fields missing from JSON keep these values)
	settings.AutoUpdate = true

	// Try settings.json first
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err == nil {
			ensureAutoSelectRoles()
			return nil
		}
	}

	// Migrate from legacy gamedir.txt
	if legacy, err := os.ReadFile(legacyPath); err == nil {
		dir := trimSpace(string(legacy))
		if dir != "" {
			settings.GamePath = dir
			if err := save(); err != nil {
				return err
			}
			os.Remove(legacyPath)
			return nil
		}
	}

	// Neither file exists — start with defaults
	settings.AutoUpdate = true
	ensureAutoSelectRoles()
	return nil
}

func ensureAutoSelectRoles() {
	if settings.AutoSelectRoles == nil {
		settings.AutoSelectRoles = make(map[string]RoleConfig)
	}
}

// Get returns a copy of the current settings.
func Get() Settings {
	mu.RLock()
	defer mu.RUnlock()
	return settings
}

// GamePath returns the configured game directory path.
func GamePath() string {
	mu.RLock()
	defer mu.RUnlock()
	return settings.GamePath
}

// SetGamePath updates and persists the game directory path.
func SetGamePath(path string) error {
	mu.Lock()
	defer mu.Unlock()
	settings.GamePath = path
	return save()
}

// AutoAccept returns the current auto-accept setting.
func AutoAccept() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.AutoAccept
}

// SetAutoAccept updates and persists the auto-accept setting.
func SetAutoAccept(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.AutoAccept = enabled
	return save()
}

// BenchSwap returns the current bench-swap setting.
func BenchSwap() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.BenchSwap
}

// SetBenchSwap updates and persists the bench-swap setting.
func SetBenchSwap(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.BenchSwap = enabled
	return save()
}

// BenchSwapSkipCooldown returns the current bench-swap-skip-cooldown setting.
func BenchSwapSkipCooldown() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.BenchSwapSkipCooldown
}

// SetBenchSwapSkipCooldown updates and persists the bench-swap-skip-cooldown setting.
func SetBenchSwapSkipCooldown(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.BenchSwapSkipCooldown = enabled
	return save()
}

// StartWithWindows returns the current start-with-windows setting.
func StartWithWindows() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.StartWithWindows
}

// SetStartWithWindows updates and persists the start-with-windows setting.
func SetStartWithWindows(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.StartWithWindows = enabled
	return save()
}

// AutoUpdate returns the current auto-update setting.
func AutoUpdate() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.AutoUpdate
}

// SetAutoUpdate updates and persists the auto-update setting.
func SetAutoUpdate(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.AutoUpdate = enabled
	return save()
}

// AutoSelect returns the current auto-select setting.
func AutoSelect() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.AutoSelect
}

// SetAutoSelect updates and persists the auto-select setting.
func SetAutoSelect(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.AutoSelect = enabled
	return save()
}

// AutoSelectRoles returns a copy of the auto-select role configurations.
func AutoSelectRoles() map[string]RoleConfig {
	mu.RLock()
	defer mu.RUnlock()
	cp := make(map[string]RoleConfig, len(settings.AutoSelectRoles))
	for k, v := range settings.AutoSelectRoles {
		picks := make([]int, len(v.Picks))
		copy(picks, v.Picks)
		bans := make([]int, len(v.Bans))
		copy(bans, v.Bans)
		cp[k] = RoleConfig{Picks: picks, Bans: bans}
	}
	return cp
}

// RoomParty returns the current room-party setting.
func RoomParty() bool {
	mu.RLock()
	defer mu.RUnlock()
	return settings.RoomParty
}

// SetRoomParty updates and persists the room-party setting.
func SetRoomParty(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	settings.RoomParty = enabled
	return save()
}

// SetAutoSelectRole updates and persists the pick/ban config for a single role.
func SetAutoSelectRole(role string, picks, bans []int) error {
	mu.Lock()
	defer mu.Unlock()
	ensureAutoSelectRoles()
	settings.AutoSelectRoles[role] = RoleConfig{Picks: picks, Bans: bans}
	return save()
}

// ChatAvailability returns the current chat availability override.
func ChatAvailability() string {
	mu.RLock()
	defer mu.RUnlock()
	return settings.ChatAvailability
}

// ChatStatusMessage returns the current chat status message.
func ChatStatusMessage() string {
	mu.RLock()
	defer mu.RUnlock()
	return settings.ChatStatusMessage
}

// SetChatStatus updates and persists both chat availability and status message.
func SetChatStatus(availability, statusMessage string) error {
	mu.Lock()
	defer mu.Unlock()
	settings.ChatAvailability = availability
	settings.ChatStatusMessage = statusMessage
	return save()
}

// save writes the current settings to disk. Caller must hold mu.
func save() error {
	os.MkdirAll(AmeDir, os.ModePerm)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

// trimSpace trims whitespace and newlines (avoids importing strings for one call).
func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
