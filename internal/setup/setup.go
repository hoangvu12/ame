package setup

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hoangvu12/ame/internal/config"
)

var (
	PENGU_DIR  = config.PenguDir
	PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
)

// GetExistingPenguDir detects an existing Pengu Loader installation from the registry
// Returns the installation directory path, or empty string if not found
// Exported for use by other packages
func GetExistingPenguDir() string {
	// Query the IFEO Debugger value
	cmd := exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe`,
		"/v", "Debugger")
	cmd.SysProcAttr = getSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse the output to extract the path
	// Format: rundll32 "C:\path\to\core.dll", #6000
	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "rundll32") {
		return ""
	}

	// Find the path between quotes
	startQuote := strings.Index(outputStr, `"`)
	if startQuote == -1 {
		return ""
	}
	endQuote := strings.Index(outputStr[startQuote+1:], `"`)
	if endQuote == -1 {
		return ""
	}

	corePath := outputStr[startQuote+1 : startQuote+1+endQuote]

	// Verify the core.dll exists
	if _, err := os.Stat(corePath); os.IsNotExist(err) {
		return ""
	}

	// Return the parent directory (Pengu installation dir)
	return filepath.Dir(corePath)
}

// IsUsingExternalPengu checks if we're using a Pengu installation outside of ame's directory
func IsUsingExternalPengu() bool {
	existingDir := GetExistingPenguDir()
	if existingDir == "" {
		return false
	}
	// Check if it's NOT in our default location
	defaultPenguDir := config.PenguDir
	return !strings.EqualFold(filepath.Clean(existingDir), filepath.Clean(defaultPenguDir))
}

// DetectAndSetPenguPaths detects existing Pengu installation and updates PENGU_DIR/PLUGIN_DIR
// Call this before SetupPlugin when you need to install plugins without full setup
func DetectAndSetPenguPaths() {
	existingDir := GetExistingPenguDir()
	if existingDir != "" {
		PENGU_DIR = existingDir
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
	} else {
		// Ensure PLUGIN_DIR is set correctly for default location
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
	}
}

// GetPluginDir returns the current plugin directory path
func GetPluginDir() string {
	return PLUGIN_DIR
}

// Tool files to download
var TOOL_FILES = []string{
	"mod-tools.exe",
	"cslol-diag.exe",
	"cslol-dll.dll",
	"wad-extract.exe",
	"wad-make.exe",
}


// Config holds setup URLs
type Config struct {
	ToolsURL  string
	DllURL    string // Separate URL for cslol-dll.dll
	PenguURL  string
	PluginURL string
	DevSrcDir string // When non-empty, copy plugin from this local dir instead of downloading
}

// statusFail prints a failure message (success is silent)
func statusFail(name string) {
	fmt.Printf("  ! %s failed\n", name)
}

// info prints an info message
func info(message string) {
	fmt.Printf("  %s\n", message)
}

// downloadFile downloads a file from URL to destination
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractZip extracts a zip file to destination directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Prevent zip slip vulnerability
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// IsPenguActivated checks if Pengu Loader is activated via registry
func IsPenguActivated() bool {
	cmd := exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe`,
		"/v", "Debugger")
	cmd.SysProcAttr = getSysProcAttr()
	err := cmd.Run()
	return err == nil
}

const ifeoKey = `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe`

// ActivatePengu sets the IFEO Debugger registry value so PenguLoader
// injects into LeagueClientUx.exe on launch.
func ActivatePengu() error {
	coreDll := filepath.Join(PENGU_DIR, "core.dll")
	if _, err := os.Stat(coreDll); os.IsNotExist(err) {
		return fmt.Errorf("core.dll not found at %s", coreDll)
	}

	value := fmt.Sprintf(`rundll32 "%s", #6000`, coreDll)

	cmd := exec.Command("reg", "add", ifeoKey, "/v", "Debugger", "/t", "REG_SZ", "/d", value, "/f")
	cmd.SysProcAttr = getSysProcAttr()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reg add failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// DeactivatePengu removes the IFEO Debugger registry value.
func DeactivatePengu() {
	cmd := exec.Command("reg", "delete", ifeoKey, "/v", "Debugger", "/f")
	cmd.SysProcAttr = getSysProcAttr()
	cmd.Run() // ignore errors — value may already be absent
}

// createDirectories creates base directories (not Pengu-related, those are created after detection)
func createDirectories() {
	dirs := []string{config.AmeDir, config.ToolsDir, config.SkinsDir, config.ModsDir, config.OverlayDir}
	for _, dir := range dirs {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// setupModTools downloads mod-tools if not present
func setupModTools(toolsURL string, dllURL string) bool {
	modToolsPath := filepath.Join(config.ToolsDir, "mod-tools.exe")
	if _, err := os.Stat(modToolsPath); err == nil {
		return true
	}

	info("Downloading mod-tools...")

	for _, file := range TOOL_FILES {
		var url string
		if file == "cslol-dll.dll" && dllURL != "" {
			url = dllURL
		} else {
			url = toolsURL + "/" + file
		}
		dest := filepath.Join(config.ToolsDir, file)
		if err := downloadFile(url, dest); err != nil {
			statusFail("mod-tools download")
			return false
		}
	}

	return true
}

// setupPenguLoader detects existing installation or downloads Pengu Loader.
// Returns (ok, freshInstall).
func setupPenguLoader(penguURL string) (bool, bool) {
	// Check for existing Pengu installation via registry
	existingDir := GetExistingPenguDir()
	if existingDir != "" {
		PENGU_DIR = existingDir
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
		return true, false
	}

	// Check if we already downloaded Pengu to our directory
	penguExe := filepath.Join(PENGU_DIR, "Pengu Loader.exe")
	if _, err := os.Stat(penguExe); err == nil {
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
		return true, false
	}

	// Fresh download needed
	os.MkdirAll(PENGU_DIR, os.ModePerm)
	info("Downloading Pengu Loader...")
	zipPath := filepath.Join(config.AmeDir, "pengu.zip")

	if err := downloadFile(penguURL, zipPath); err != nil {
		statusFail("Pengu Loader download")
		return false, false
	}

	if err := extractZip(zipPath, PENGU_DIR); err != nil {
		statusFail("Pengu Loader setup")
		return false, false
	}

	os.Remove(zipPath)
	PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
	return true, true
}

// SetupPluginFromLocal copies plugin source files from a local directory
func SetupPluginFromLocal(srcDir string) bool {
	os.RemoveAll(PLUGIN_DIR)
	os.MkdirAll(PLUGIN_DIR, os.ModePerm)

	if err := copyPluginDir(srcDir); err != nil {
		statusFail("Plugin setup")
		return false
	}

	return true
}

func copyPluginDir(srcDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dstPath := filepath.Join(PLUGIN_DIR, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, os.ModePerm)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}

// SetupPlugin downloads and extracts plugin zip (exported for use after updates)
func SetupPlugin(pluginZipURL string) bool {
	info("Downloading plugin...")
	zipPath := filepath.Join(config.AmeDir, "plugin.zip")

	if err := downloadFile(pluginZipURL, zipPath); err != nil {
		statusFail("Plugin download")
		return false
	}

	os.RemoveAll(PLUGIN_DIR)
	os.MkdirAll(PLUGIN_DIR, os.ModePerm)

	if err := extractZip(zipPath, PLUGIN_DIR); err != nil {
		statusFail("Plugin setup")
		return false
	}

	os.Remove(zipPath)
	return true
}

// checkPenguActivation activates Pengu via registry.
// Returns true if activation succeeded or was already active.
func checkPenguActivation() bool {
	if IsPenguActivated() {
		return true
	}

	if err := ActivatePengu(); err != nil {
		statusFail("Pengu activation")
		return false
	}

	return true
}

// RunSetup runs the full setup process (version can be passed for display)
func RunSetup(config Config) bool {
	// Create directories
	createDirectories()

	// Setup mod-tools
	if !setupModTools(config.ToolsURL, config.DllURL) {
		return false
	}

	// Setup Pengu Loader
	ok, freshInstall := setupPenguLoader(config.PenguURL)
	if !ok {
		return false
	}

	// Setup plugin
	if config.DevSrcDir != "" {
		if !SetupPluginFromLocal(config.DevSrcDir) {
			return false
		}
	} else {
		if !SetupPlugin(config.PluginURL) {
			return false
		}
	}

	// Only auto-activate on fresh install
	if freshInstall {
		checkPenguActivation()
	}

	return true
}
