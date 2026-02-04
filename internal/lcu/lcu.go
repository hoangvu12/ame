package lcu

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

// IsClientRunning checks if LeagueClientUx.exe is currently running.
func IsClientRunning() bool {
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

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://127.0.0.1:%s/riotclient/kill-and-restart-ux", port), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := client.Do(req)
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

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:%s/riotclient/region-locale", port), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := client.Do(req)
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

// getCredentials extracts the auth token and port from the LeagueClientUx.exe command line.
func getCredentials() (port string, token string, err error) {
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
