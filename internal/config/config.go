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

// Settings holds persisted application settings.
type Settings struct {
	GamePath         string `json:"gamePath"`
	AutoAccept       bool   `json:"autoAccept"`
	BenchSwap        bool   `json:"benchSwap"`
	StartWithWindows bool   `json:"startWithWindows"`
}

// Init loads settings from disk.
// It migrates from the legacy gamedir.txt file if settings.json does not exist.
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	// Try settings.json first
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err == nil {
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

	// Neither file exists — start with zero-value settings
	return nil
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
