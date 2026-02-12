package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Paths - single source of truth (previously duplicated across packages)
var (
	AmeDir        = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame")
	ToolsDir      = filepath.Join(AmeDir, "tools")
	SkinsDir      = filepath.Join(AmeDir, "skins")
	ModsDir       = filepath.Join(AmeDir, "mods")
	OverlayDir    = filepath.Join(AmeDir, "overlay")
	PenguDir      = filepath.Join(AmeDir, "pengu")
	CustomModsDir = filepath.Join(AmeDir, "custom-mods")
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

// CustomMod represents a user-imported custom mod (.fantome/.zip).
type CustomMod struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Author     string `json:"author"`
	ChampionID int    `json:"championId"`
	Enabled    bool   `json:"enabled"`
	HasImage   bool   `json:"hasImage"`
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
	RandomSkin        string                `json:"randomSkin"`
	CustomMods        []CustomMod           `json:"customMods"`
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

// RandomSkin returns the current random-skin mode ("", "all", or "top3").
func RandomSkin() string {
	mu.RLock()
	defer mu.RUnlock()
	return settings.RandomSkin
}

// SetRandomSkin updates and persists the random-skin mode.
func SetRandomSkin(mode string) error {
	mu.Lock()
	defer mu.Unlock()
	settings.RandomSkin = mode
	return save()
}

// SetChatStatus updates and persists both chat availability and status message.
func SetChatStatus(availability, statusMessage string) error {
	mu.Lock()
	defer mu.Unlock()
	settings.ChatAvailability = availability
	settings.ChatStatusMessage = statusMessage
	return save()
}

// GetCustomMods returns a copy of the custom mods list.
func GetCustomMods() []CustomMod {
	mu.RLock()
	defer mu.RUnlock()
	cp := make([]CustomMod, len(settings.CustomMods))
	copy(cp, settings.CustomMods)
	return cp
}

// GetEnabledCustomMods returns only the enabled custom mods.
func GetEnabledCustomMods() []CustomMod {
	mu.RLock()
	defer mu.RUnlock()
	var enabled []CustomMod
	for _, m := range settings.CustomMods {
		if m.Enabled {
			enabled = append(enabled, m)
		}
	}
	return enabled
}

// AddCustomMod appends a new custom mod and persists.
func AddCustomMod(mod CustomMod) error {
	mu.Lock()
	defer mu.Unlock()
	settings.CustomMods = append(settings.CustomMods, mod)
	return save()
}

// UpdateCustomMod updates name, author, championId for an existing mod.
func UpdateCustomMod(id, name, author string, championID int) (*CustomMod, error) {
	mu.Lock()
	defer mu.Unlock()
	for i := range settings.CustomMods {
		if settings.CustomMods[i].ID == id {
			settings.CustomMods[i].Name = name
			settings.CustomMods[i].Author = author
			settings.CustomMods[i].ChampionID = championID
			cp := settings.CustomMods[i]
			return &cp, save()
		}
	}
	return nil, nil
}

// DeleteCustomMod removes a custom mod by ID and persists.
func DeleteCustomMod(id string) error {
	mu.Lock()
	defer mu.Unlock()
	for i := range settings.CustomMods {
		if settings.CustomMods[i].ID == id {
			settings.CustomMods = append(settings.CustomMods[:i], settings.CustomMods[i+1:]...)
			return save()
		}
	}
	return nil
}

// ToggleCustomMod sets the enabled state for a custom mod.
func ToggleCustomMod(id string, enabled bool) (*CustomMod, error) {
	mu.Lock()
	defer mu.Unlock()
	for i := range settings.CustomMods {
		if settings.CustomMods[i].ID == id {
			settings.CustomMods[i].Enabled = enabled
			cp := settings.CustomMods[i]
			return &cp, save()
		}
	}
	return nil, nil
}

// SetCustomModHasImage updates the hasImage flag for a mod.
func SetCustomModHasImage(id string, hasImage bool) error {
	mu.Lock()
	defer mu.Unlock()
	for i := range settings.CustomMods {
		if settings.CustomMods[i].ID == id {
			settings.CustomMods[i].HasImage = hasImage
			return save()
		}
	}
	return nil
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
