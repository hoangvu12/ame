package lcu

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hoangvu12/ame/internal/game"
)

// lcuClient trusts the local LCU self-signed certificate.
var lcuClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// readLockfile returns port and token from the LCU lockfile, which the client
// writes on startup and removes on shutdown. Format: name:pid:port:password:protocol.
func readLockfile() (port, token string, ok bool) {
	gameDir := game.FindGameDir()
	if gameDir == "" {
		return "", "", false
	}
	data, err := os.ReadFile(filepath.Join(gameDir, "..", "lockfile"))
	if err != nil {
		return "", "", false
	}
	parts := strings.Split(strings.TrimSpace(string(data)), ":")
	if len(parts) < 5 {
		return "", "", false
	}
	return parts[2], parts[3], true
}

// IsClientRunning checks if the League client is currently running.
func IsClientRunning() bool {
	if _, _, ok := readLockfile(); ok {
		return true
	}
	cmd := exec.Command("wmic", "process", "where", "name='LeagueClientUx.exe'", "get", "ProcessId", "/value")
	cmd.SysProcAttr = getSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "ProcessId=")
}

// RestartClient restarts the League Client UX by calling the LCU API.
// It extracts auth credentials from the running process command line
// and calls POST /riotclient/kill-and-restart-ux.
func RestartClient() error {
	port, token, err := getCredentials()
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + token))

	req, err := http.NewRequest("POST", fmt.Sprintf("https://127.0.0.1:%s/riotclient/kill-and-restart-ux", port), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := lcuClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("LCU returned status %d", resp.StatusCode)
	}

	return nil
}

type regionLocaleResponse struct {
	Locale string `json:"locale"`
}

// GetRegionLocale fetches the client locale (e.g. en_US, vi_VN) from the LCU API.
func GetRegionLocale() (string, error) {
	port, token, err := getCredentials()
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + token))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:%s/riotclient/region-locale", port), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := lcuClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LCU returned status %d", resp.StatusCode)
	}

	var payload regionLocaleResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	if payload.Locale == "" {
		return "", fmt.Errorf("LCU returned empty locale")
	}

	return payload.Locale, nil
}

var (
	tokenRe = regexp.MustCompile(`--remoting-auth-token=(\S+)`)
	portRe  = regexp.MustCompile(`--app-port=(\S+)`)
)

// getCredentials extracts the LCU auth token and port. It prefers the lockfile
// (instant) and falls back to parsing LeagueClientUx.exe's command line via wmic.
func getCredentials() (port string, token string, err error) {
	if p, t, ok := readLockfile(); ok {
		return p, t, nil
	}

	cmd := exec.Command("wmic", "process", "where", "name='LeagueClientUx.exe'", "get", "CommandLine", "/value")
	cmd.SysProcAttr = getSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to query process: %w", err)
	}

	line := string(output)

	if m := tokenRe.FindStringSubmatch(line); len(m) > 1 {
		token = strings.Trim(m[1], `"`)
	}
	if m := portRe.FindStringSubmatch(line); len(m) > 1 {
		port = strings.Trim(m[1], `"`)
	}

	if token == "" || port == "" {
		return "", "", fmt.Errorf("could not find LCU credentials from process")
	}

	return port, token, nil
}
