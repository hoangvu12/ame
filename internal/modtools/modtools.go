package modtools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

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
	modTools := filepath.Join(config.ToolsDir, "mod-tools.exe")

	if _, err := os.Stat(modTools); os.IsNotExist(err) {
		return false, 1
	}

	cmd := exec.Command(modTools, "mkoverlay", modsDir, overlayDir,
		fmt.Sprintf("--game:%s", gameDir),
		fmt.Sprintf("--mods:%s", modName),
		"--noTFT", "--ignoreConflict")
	cmd.Dir = config.ToolsDir

	err := cmd.Run()
	if err != nil {
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
	modTools := filepath.Join(config.ToolsDir, "mod-tools.exe")

	if _, err := os.Stat(modTools); os.IsNotExist(err) {
		return fmt.Errorf("mod-tools.exe not found")
	}

	cmd := exec.Command(modTools, "runoverlay",
		overlayDir, configPath,
		fmt.Sprintf("--game:%s", gameDir),
		"--opts:none") // bocchi uses --opts:none

	cmd.Dir = config.ToolsDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	stdinPipe = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	runningProcess = cmd

	// Drain stdout/stderr to keep pipes alive
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
		}
	}()

	// Monitor process exit
	go func() {
		cmd.Wait()
		runningProcess = nil
		stdinPipe = nil
	}()

	return nil
}

// Exists checks if mod-tools.exe exists
func Exists() bool {
	modTools := filepath.Join(config.ToolsDir, "mod-tools.exe")
	_, err := os.Stat(modTools)
	return err == nil
}

// IsRunning checks if runoverlay is currently running
func IsRunning() bool {
	return runningProcess != nil && runningProcess.Process != nil
}
