package setup

import (
	"archive/zip"
	"fmt"
	"io"
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

// isPenguActivated checks if Pengu Loader is activated via registry
func isPenguActivated() bool {
	cmd := exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe`,
		"/v", "Debugger")
	cmd.SysProcAttr = getSysProcAttr()
	err := cmd.Run()
	return err == nil
}

// launchPenguForActivation launches Pengu Loader and waits for user to close it
func launchPenguForActivation() error {
	penguExe := filepath.Join(PENGU_DIR, "Pengu Loader.exe")
	if _, err := os.Stat(penguExe); os.IsNotExist(err) {
		return err
	}

	fmt.Println()
	fmt.Println("  +-----------------------------------------+")
	fmt.Println("  |  Click 'Activate' in Pengu Loader       |")
	fmt.Println("  |  then close the window to continue      |")
	fmt.Println("  +-----------------------------------------+")
	fmt.Println()

	cmd := exec.Command(penguExe)
	cmd.SysProcAttr = getSysProcAttr()
	err := cmd.Run()
	return err
}

// createDirectories creates base directories (not Pengu-related, those are created after detection)
func createDirectories() {
	dirs := []string{config.AmeDir, config.ToolsDir, config.SkinsDir, config.ModsDir, config.OverlayDir}
	for _, dir := range dirs {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// setupModTools downloads mod-tools if not present
func setupModTools(toolsURL string) bool {
	modToolsPath := filepath.Join(config.ToolsDir, "mod-tools.exe")
	if _, err := os.Stat(modToolsPath); err == nil {
		return true
	}

	info("Downloading mod-tools...")

	for _, file := range TOOL_FILES {
		url := toolsURL + "/" + file
		dest := filepath.Join(config.ToolsDir, file)
		if err := downloadFile(url, dest); err != nil {
			statusFail("mod-tools download")
			return false
		}
	}

	return true
}

// setupPenguLoader detects existing installation or downloads Pengu Loader
func setupPenguLoader(penguURL string) bool {
	// Check for existing Pengu installation via registry
	existingDir := GetExistingPenguDir()
	if existingDir != "" {
		PENGU_DIR = existingDir
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
		return true
	}

	// Check if we already downloaded Pengu to our directory
	penguExe := filepath.Join(PENGU_DIR, "Pengu Loader.exe")
	if _, err := os.Stat(penguExe); err == nil {
		PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
		return true
	}

	// Fresh download needed
	os.MkdirAll(PENGU_DIR, os.ModePerm)
	info("Downloading Pengu Loader...")
	zipPath := filepath.Join(config.AmeDir, "pengu.zip")

	if err := downloadFile(penguURL, zipPath); err != nil {
		statusFail("Pengu Loader download")
		return false
	}

	if err := extractZip(zipPath, PENGU_DIR); err != nil {
		statusFail("Pengu Loader setup")
		return false
	}

	os.Remove(zipPath)
	PLUGIN_DIR = filepath.Join(PENGU_DIR, "plugins", "ame")
	return true
}

// SetupPluginFromLocal copies plugin source files from a local directory
func SetupPluginFromLocal(srcDir string) bool {
	os.RemoveAll(PLUGIN_DIR)
	os.MkdirAll(PLUGIN_DIR, os.ModePerm)

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		statusFail("Plugin setup")
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(PLUGIN_DIR, entry.Name())

		data, err := os.ReadFile(srcPath)
		if err != nil {
			statusFail("Plugin setup")
			return false
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			statusFail("Plugin setup")
			return false
		}
	}

	return true
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

// checkPenguActivation checks and prompts for Pengu activation
func checkPenguActivation() bool {
	if isPenguActivated() {
		return true
	}

	launchPenguForActivation()

	if isPenguActivated() {
		return true
	}

	statusFail("Pengu activation")
	return false
}

// RunSetup runs the full setup process (version can be passed for display)
func RunSetup(config Config) bool {
	// Create directories
	createDirectories()

	// Setup mod-tools
	if !setupModTools(config.ToolsURL) {
		return false
	}

	// Setup Pengu Loader
	if !setupPenguLoader(config.PenguURL) {
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

	// Check activation
	checkPenguActivation()

	return true
}
