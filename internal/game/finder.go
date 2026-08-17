package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/winproc"
)

// cachedGameDir is read and written from request-handling goroutines
// (handleApply, handlePrefetch), so it is guarded.
var (
	cachedGameDirMu sync.RWMutex
	cachedGameDir   string
)

// loadCachedGameDir returns the cached directory if it is still valid.
func loadCachedGameDir() string {
	cachedGameDirMu.RLock()
	dir := cachedGameDir
	cachedGameDirMu.RUnlock()
	if dir != "" && isValidGameDir(dir) {
		return dir
	}
	return ""
}

// storeCachedGameDir records dir as the resolved game directory.
func storeCachedGameDir(dir string) {
	cachedGameDirMu.Lock()
	cachedGameDir = dir
	cachedGameDirMu.Unlock()
}

// commonPathSuffixes lists directory patterns (relative to a drive root) where
// League of Legends may be installed.
var commonPathSuffixes = []string{
	filepath.Join("Riot Games", "League of Legends", "Game"),
	filepath.Join("Riot", "Riot Games", "League of Legends", "Game"),
	filepath.Join("Games", "League of Legends", "Game"),
	filepath.Join("Program Files", "Riot Games", "League of Legends", "Game"),
	filepath.Join("Program Files (x86)", "Riot Games", "League of Legends", "Game"),
}

// isValidGameDir checks if a directory contains League of Legends.exe
func isValidGameDir(dir string) bool {
	return fileExists(filepath.Join(dir, "League of Legends.exe"))
}

// getSavedGameDir reads the saved game directory from config
func getSavedGameDir() string {
	dir := config.GamePath()
	if dir != "" && isValidGameDir(dir) {
		return dir
	}
	return ""
}

// SaveGameDir persists the game directory to config
func SaveGameDir(dir string) {
	config.SetGamePath(dir)
}

// addCandidate appends dir to candidates if it's a valid game dir and not already present.
func addCandidate(candidates []string, dir string) []string {
	dir = filepath.Clean(dir)
	if !isValidGameDir(dir) {
		return candidates
	}
	for _, c := range candidates {
		if strings.EqualFold(c, dir) {
			return candidates
		}
	}
	return append(candidates, dir)
}

// newestGameDir returns the candidate whose League of Legends.exe was most
// recently modified. This picks the actively-patched installation over stale ones.
func newestGameDir(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	best := candidates[0]
	bestTime := modTime(filepath.Join(best, "League of Legends.exe"))

	for _, c := range candidates[1:] {
		t := modTime(filepath.Join(c, "League of Legends.exe"))
		if t.After(bestTime) {
			best = c
			bestTime = t
		}
	}
	return best
}

// FindGameDir finds League of Legends Game directory using multiple detection methods.
// When multiple installations exist it picks the one with the most recently
// modified League of Legends.exe (i.e. the actively-patched installation).
func FindGameDir() string {
	// 0. In-memory cache
	if dir := loadCachedGameDir(); dir != "" {
		return dir
	}

	// Collect all candidate directories, then pick the newest.
	var candidates []string

	// 1. Saved path from previous run (as candidate, not final answer)
	if dir := getSavedGameDir(); dir != "" {
		candidates = addCandidate(candidates, dir)
	}

	// 2. RiotClientInstalls.json (authoritative — checked early)
	candidates = appendFromRiotClientInstalls(candidates)

	// 3. Common paths on all fixed drives
	for _, drive := range getFixedDrives() {
		for _, suffix := range commonPathSuffixes {
			candidates = addCandidate(candidates, filepath.Join(drive, suffix))
		}
	}

	if best := newestGameDir(candidates); best != "" {
		storeCachedGameDir(best)
		SaveGameDir(best)
		return best
	}

	// 4. Running LeagueClientUx.exe process (expensive — only if nothing above matched)
	if dir := findFromRunningProcess(); dir != "" {
		storeCachedGameDir(dir)
		SaveGameDir(dir)
		return dir
	}

	// 5. Registry lookup
	if dir := findFromRegistry(); dir != "" {
		storeCachedGameDir(dir)
		SaveGameDir(dir)
		return dir
	}

	return ""
}

// appendFromRiotClientInstalls parses RiotClientInstalls.json and appends any
// valid League game directories it finds to candidates.
func appendFromRiotClientInstalls(candidates []string) []string {
	rcPath := `C:\ProgramData\Riot Games\RiotClientInstalls.json`
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return candidates
	}
	var rc map[string]interface{}
	if err := json.Unmarshal(data, &rc); err != nil {
		return candidates
	}

	var paths []string
	if assoc, ok := rc["associated_client"].(map[string]interface{}); ok {
		for key := range assoc {
			paths = append(paths, key)
		}
	}
	if def, ok := rc["rc_default"].(string); ok {
		paths = append(paths, def)
	}
	if live, ok := rc["rc_live"].(string); ok {
		paths = append(paths, live)
	}

	leagueRe := regexp.MustCompile(`(?i)league`)
	for _, p := range paths {
		if leagueRe.MatchString(p) {
			// associated_client keys are League root dirs (e.g. "D:/Riot/Riot Games/League of Legends/")
			// so "Game" is a direct child. Also try parent in case the path points to an exe.
			candidates = addCandidate(candidates, filepath.Join(p, "Game"))
			candidates = addCandidate(candidates, filepath.Join(p, "..", "Game"))
		}
	}
	return candidates
}

// findFromRunningProcess tries to locate the game dir from a running LeagueClientUx.exe.
//
// This used to shell out to wmic, which no longer exists on Windows 11 24H2+.
// QueryFullProcessImageName via winproc returns the same executable path
// without a process spawn.
func findFromRunningProcess() string {
	exePath := winproc.FindImagePath("LeagueClientUx.exe")
	if exePath == "" {
		return ""
	}
	gameDir := filepath.Join(exePath, "..", "Game")
	if isValidGameDir(gameDir) {
		return filepath.Clean(gameDir)
	}
	return ""
}

// findFromRegistry tries to locate the game dir from the Windows registry.
func findFromRegistry() string {
	output := runCommand("reg", "query", `HKLM\SOFTWARE\WOW6432Node\Riot Games, Inc\League of Legends`, "/v", "Location")
	if output == "" {
		return ""
	}
	re := regexp.MustCompile(`Location\s+REG_SZ\s+(.+)`)
	if match := re.FindStringSubmatch(output); len(match) > 1 {
		loc := strings.TrimSpace(match[1])
		gameDir := filepath.Join(loc, "Game")
		if isValidGameDir(gameDir) {
			return filepath.Clean(gameDir)
		}
	}
	return ""
}

// PromptGameDir asks the user to manually enter the game directory.
// Returns the validated path, or empty string if the user skips.
func PromptGameDir() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("  Enter the path to the Game folder, for example:")
	fmt.Println("  C:\\Riot Games\\League of Legends\\Game")
	fmt.Println()

	for {
		fmt.Print("  Path (or press Enter to skip): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return ""
		}

		// Remove surrounding quotes if user pasted a quoted path
		input = strings.Trim(input, `"'`)

		if isValidGameDir(input) {
			storeCachedGameDir(input)
			SaveGameDir(input)
			fmt.Printf("  > %s\n", input)
			return input
		}

		fmt.Println("  ! League of Legends.exe not found in that folder.")
		fmt.Println()
	}
}

var (
	fixedDrivesOnce  sync.Once
	fixedDrivesCache []string
)

// getFixedDrives returns the local fixed drives (C:, D:, ...) that are readable.
//
// Two things matter here. It asks the OS which letters exist and filters to
// DRIVE_FIXED, because the previous A-Z probe also touched optical drives with
// no media and disconnected network mappings — the latter block for the full
// SMB timeout, stalling game detection for tens of seconds. And it reads a
// single directory entry to test readability rather than materializing the
// whole drive root. The result is memoized because drive letters do not
// meaningfully change within one run.
func getFixedDrives() []string {
	fixedDrivesOnce.Do(func() {
		for _, drive := range logicalFixedDrives() {
			if isReadableDir(drive) {
				fixedDrivesCache = append(fixedDrivesCache, drive)
			}
		}
	})
	return fixedDrivesCache
}

// isReadableDir reports whether dir can be opened and listed, reading at most
// one entry so an enormous drive root costs the same as an empty one.
func isReadableDir(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()
	if _, err := f.Readdirnames(1); err != nil && err != io.EOF {
		return false
	}
	return true
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// modTime returns the modification time of a file, or zero time on error.
func modTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// runCommand runs a command and returns its output
func runCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = getSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
