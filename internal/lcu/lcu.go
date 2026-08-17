package lcu

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/game"
	"github.com/hoangvu12/ame/internal/winproc"
)

// clientProcess is the League client process that owns the LCU API.
const clientProcess = "LeagueClientUx.exe"

// lcuClient trusts the local LCU self-signed certificate.
//
// The timeouts are not optional: a client that is mid-patch or waiting on
// Vanguard accepts the connection and then never answers, and this client sits
// on the synchronous startup path. Without a deadline that hangs ame forever
// before it can serve anything.
var lcuClient = &http.Client{
	Timeout: 3 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 2 * time.Second,
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
//
// This asks the OS for the process rather than trusting the lockfile: the
// client only deletes its lockfile on a clean shutdown, so after a crash a
// stale lockfile would report a dead client as running.
func IsClientRunning() bool {
	return winproc.IsRunning(clientProcess)
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

// getCredentials extracts the LCU auth token and port from the lockfile.
//
// There used to be a wmic fallback that parsed --app-port/--remoting-auth-token
// off the client's command line. It has been removed: wmic no longer ships with
// Windows, so it could only ever fail, and it failed silently in a way that was
// indistinguishable from "the client isn't running". The lockfile carries both
// values anyway, and reading it is a single file read.
func getCredentials() (port string, token string, err error) {
	if p, t, ok := readLockfile(); ok {
		return p, t, nil
	}
	if !IsClientRunning() {
		return "", "", fmt.Errorf("League client is not running")
	}
	return "", "", fmt.Errorf("could not read LCU lockfile")
}
