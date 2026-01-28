package modtools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var TOOLS_DIR = filepath.Join(os.Getenv("LOCALAPPDATA"), "ame", "tools")

// Running process reference (like bocchi's this.runningProcess)
var runningProcess *exec.Cmd

// KillModTools kills any running mod-tools processes
func KillModTools() {
	// First try to gracefully stop by writing to stdin (like bocchi)
	if stdinPipe != nil {
		stdinPipe.Write([]byte("\n"))
		stdinPipe.Close()
		stdinPipe = nil
	}

	// Give it a moment to exit gracefully
	if runningProcess != nil && runningProcess.Process != nil {
		// Wait a bit for graceful exit
		done := make(chan struct{})
		go func() {
			runningProcess.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(1 * time.Second):
			// Force kill if still running
			runningProcess.Process.Kill()
		}
		runningProcess = nil
	}

	// Then force kill any remaining mod-tools
	cmd := exec.Command("taskkill", "/F", "/IM", "mod-tools.exe")
	cmd.SysProcAttr = getSysProcAttr()
	cmd.Run()
}

// RunMkOverlay runs mod-tools mkoverlay command
func RunMkOverlay(modsDir, overlayDir, gameDir, modName string) (bool, int) {
	modTools := filepath.Join(TOOLS_DIR, "mod-tools.exe")

	if _, err := os.Stat(modTools); os.IsNotExist(err) {
		return false, 1
	}

	args := fmt.Sprintf(`mkoverlay "%s" "%s" "--game:%s" --mods:%s --noTFT --ignoreConflict`,
		modsDir, overlayDir, gameDir, modName)

	fmt.Printf("[ame] Running: %s %s\n", modTools, args)

	cmd := exec.Command(modTools, "mkoverlay", modsDir, overlayDir,
		fmt.Sprintf("--game:%s", gameDir),
		fmt.Sprintf("--mods:%s", modName),
		"--noTFT", "--ignoreConflict")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = TOOLS_DIR

	err := cmd.Run()
	if err != nil {
		fmt.Printf("[ame] mkoverlay error: %v\n", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return false, exitErr.ExitCode()
		}
		return false, 1
	}

	return true, 0
}

// stdinPipe holds the stdin writer for runoverlay (to keep it alive and stop gracefully)
var stdinPipe io.WriteCloser

// RunOverlay runs mod-tools runoverlay command (NOT detached, like bocchi)
func RunOverlay(overlayDir, configPath, gameDir string) error {
	modTools := filepath.Join(TOOLS_DIR, "mod-tools.exe")

	if _, err := os.Stat(modTools); os.IsNotExist(err) {
		return fmt.Errorf("mod-tools.exe not found")
	}

	fmt.Printf("[ame] Running: %s runoverlay \"%s\" \"%s\" \"--game:%s\" --opts:none\n",
		modTools, overlayDir, configPath, gameDir)

	// Like bocchi: detached: false, stdio: ['pipe', 'pipe', 'pipe']
	cmd := exec.Command(modTools, "runoverlay",
		overlayDir, configPath,
		fmt.Sprintf("--game:%s", gameDir),
		"--opts:none") // bocchi uses --opts:none

	cmd.Dir = TOOLS_DIR

	// Create pipes for stdin/stdout/stderr (like bocchi's stdio: ['pipe', 'pipe', 'pipe'])
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	// Keep stdin pipe open (bocchi keeps reference and writes to it to stop)
	stdinPipe = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the process (NOT detached)
	err = cmd.Start()
	if err != nil {
		fmt.Printf("[ame] Failed to start runoverlay: %v\n", err)
		return err
	}

	// Keep reference like bocchi's this.runningProcess
	runningProcess = cmd

	fmt.Printf("[ame] runoverlay started with PID %d\n", cmd.Process.Pid)

	// Read stdout in goroutine (keeps pipe alive, like bocchi's event listeners)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("[MOD-TOOLS]: %s\n", line)
		}
	}()

	// Read stderr in goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("[MOD-TOOLS ERROR]: %s\n", line)
		}
	}()

	// Monitor process exit in goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("[ame] runoverlay exited with error: %v\n", err)
		} else {
			fmt.Println("[ame] runoverlay exited normally")
		}
		runningProcess = nil
		stdinPipe = nil
	}()

	return nil
}

// Exists checks if mod-tools.exe exists
func Exists() bool {
	modTools := filepath.Join(TOOLS_DIR, "mod-tools.exe")
	_, err := os.Stat(modTools)
	return err == nil
}

// IsRunning checks if runoverlay is currently running
func IsRunning() bool {
	return runningProcess != nil && runningProcess.Process != nil
}
