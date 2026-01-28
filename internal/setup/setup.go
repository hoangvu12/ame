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
)

// Paths
var (
	AME_DIR     = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame")
	TOOLS_DIR   = filepath.Join(AME_DIR, "tools")
	SKINS_DIR   = filepath.Join(AME_DIR, "skins")
	MODS_DIR    = filepath.Join(AME_DIR, "mods")
	OVERLAY_DIR = filepath.Join(AME_DIR, "overlay")
	PENGU_DIR   = filepath.Join(AME_DIR, "pengu")
	PLUGIN_DIR  = filepath.Join(PENGU_DIR, "plugins", "ame")
)

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
}

// status prints a status line with indicator
func status(name string, ok bool) {
	if ok {
		fmt.Printf("  + %s\n", name)
	} else {
		fmt.Printf("  x %s (failed)\n", name)
	}
}

// info prints an info message
func info(message string) {
	fmt.Printf("  > %s\n", message)
}

// errorMsg prints an error message
func errorMsg(message string) {
	fmt.Printf("  ! %s\n", message)
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
		errorMsg("Pengu Loader.exe not found")
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

// createDirectories creates all required directories
func createDirectories() {
	dirs := []string{AME_DIR, TOOLS_DIR, SKINS_DIR, MODS_DIR, OVERLAY_DIR, PENGU_DIR, PLUGIN_DIR}
	for _, dir := range dirs {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// setupModTools downloads mod-tools if not present
func setupModTools(toolsURL string) bool {
	modToolsPath := filepath.Join(TOOLS_DIR, "mod-tools.exe")
	if _, err := os.Stat(modToolsPath); err == nil {
		status("mod-tools", true)
		return true
	}

	info("Downloading mod-tools...")
	allSuccess := true

	for _, file := range TOOL_FILES {
		url := toolsURL + "/" + file
		dest := filepath.Join(TOOLS_DIR, file)
		if err := downloadFile(url, dest); err != nil {
			allSuccess = false
		}
	}

	status("mod-tools", allSuccess)
	return allSuccess
}

// setupPenguLoader downloads and extracts Pengu Loader if not present
func setupPenguLoader(penguURL string) bool {
	penguExe := filepath.Join(PENGU_DIR, "Pengu Loader.exe")
	if _, err := os.Stat(penguExe); err == nil {
		status("Pengu Loader", true)
		return true
	}

	info("Downloading Pengu Loader...")
	zipPath := filepath.Join(AME_DIR, "pengu.zip")

	if err := downloadFile(penguURL, zipPath); err != nil {
		status("Pengu Loader", false)
		return false
	}

	if err := extractZip(zipPath, PENGU_DIR); err != nil {
		status("Pengu Loader", false)
		return false
	}

	os.Remove(zipPath)
	status("Pengu Loader", true)
	return true
}

// SetupPlugin downloads and extracts plugin zip (exported for use after updates)
func SetupPlugin(pluginZipURL string) bool {
	info("Downloading plugin...")
	zipPath := filepath.Join(AME_DIR, "plugin.zip")

	if err := downloadFile(pluginZipURL, zipPath); err != nil {
		status("Plugin", false)
		return false
	}

	// Clear existing plugin files
	os.RemoveAll(PLUGIN_DIR)
	os.MkdirAll(PLUGIN_DIR, os.ModePerm)

	if err := extractZip(zipPath, PLUGIN_DIR); err != nil {
		status("Plugin", false)
		return false
	}

	os.Remove(zipPath)
	status("Plugin", true)
	return true
}

// checkPenguActivation checks and prompts for Pengu activation
func checkPenguActivation() bool {
	if isPenguActivated() {
		status("Pengu activated", true)
		return true
	}

	launchPenguForActivation()

	if isPenguActivated() {
		status("Pengu activated", true)
		return true
	}

	status("Pengu activated", false)
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
	if !SetupPlugin(config.PluginURL) {
		return false
	}

	// Check activation
	checkPenguActivation()

	return true
}
