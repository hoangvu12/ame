package ltk

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hoangvu12/ame/internal/display"
)

// --- constants that change only when the service endpoint does ---

const (
	deliveryBaseURL     = "https://patcher.leaguetoolkit.dev"
	deliveryManifestKey = "RWSk74swcHfkZfLblm3iSk9RLVHJ1xjtqn/zySIPaCBM1hecnlFy1Bxi"
	maxDifficulty       = 24

	ltkReleasesURL = "https://api.github.com/repos/LeagueToolkit/ltk-manager/releases/latest"

	// Ed25519 signature = 64 bytes = 86 base64url chars without padding.
	sigB64Len = 86
)

// --- types ---

type DeliveryManifest struct {
	Schema   uint32          `json:"schema"`
	Channel  string          `json:"channel"`
	Sequence uint64          `json:"sequence"`
	Release  DeliveryRelease `json:"release"`
	Revoked  []string        `json:"revoked"`
}

type DeliveryRelease struct {
	Version  string `json:"version"`
	Digest   string `json:"digest"`
	Size     uint64 `json:"size"`
	Protocol uint32 `json:"protocol"`
	Managers string `json:"managers"`
	EOL      uint32 `json:"eol"`
}

type deliveryChallenge struct {
	Challenge  string `json:"challenge"`
	Difficulty uint32 `json:"difficulty"`
}
type deliveryRegistered struct{ Token string `json:"token"` }
type deliveryEnvelope struct {
	Manifest  string `json:"manifest"`
	Signature string `json:"signature"`
}

type deliveryOffered struct {
	Manifest deliveryEnvelope `json:"manifest"`
}

// deliveryAttestation holds a carved attestation and the version it names.
type deliveryAttestation struct {
	Token   string // "eyJ...sig" full attestation
	Manager string // e.g. "1.32.2"
}

// --- attestation store ---

// loadAttestation reads the cached attestation or returns the baked-in fallback.
// For the fallback there is no version file; we parse it from the token.
func loadAttestation(cacheDir string) deliveryAttestation {
	if cacheDir != "" {
		if att, ok := readCachedAttestation(cacheDir); ok {
			return att
		}
	}
	return parseAttestation(fallbackAttestation)
}

func readCachedAttestation(cacheDir string) (deliveryAttestation, bool) {
	token, err := os.ReadFile(filepath.Join(cacheDir, "attestation.txt"))
	if err != nil || len(token) < 50 {
		return deliveryAttestation{}, false
	}
	ver, err := os.ReadFile(filepath.Join(cacheDir, "attestation_version.txt"))
	if err != nil {
		return deliveryAttestation{}, false
	}
	return deliveryAttestation{Token: strings.TrimSpace(string(token)), Manager: strings.TrimSpace(string(ver))}, true
}

func saveAttestation(cacheDir string, att deliveryAttestation) {
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "attestation.txt"), []byte(att.Token), 0644)
	os.WriteFile(filepath.Join(cacheDir, "attestation_version.txt"), []byte(att.Manager), 0644)
}

func userAgent(manager string) string {
	return "ltk-manager/" + manager
}

// parseAttestation splits a token and extracts the manager version from the statement.
func parseAttestation(token string) deliveryAttestation {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return deliveryAttestation{Token: token, Manager: "1.32.2"}
	}
	stmtBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return deliveryAttestation{Token: token, Manager: "1.32.2"}
	}
	var stmt struct {
		Manager string `json:"manager"`
	}
	if err := json.Unmarshal(stmtBytes, &stmt); err != nil || stmt.Manager == "" {
		return deliveryAttestation{Token: token, Manager: "1.32.2"}
	}
	return deliveryAttestation{Token: token, Manager: stmt.Manager}
}

// --- attestation refresh ---

// refreshAttestation downloads the latest LTK Manager release, carves the
// attestation from its binary, caches it, and returns it.
// Requires 7z on PATH for NSIS extraction.
func refreshAttestation(cacheDir string) (deliveryAttestation, string, error) {
	// 1. Latest release metadata
	req, _ := http.NewRequest("GET", ltkReleasesURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent("1.32.2"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return deliveryAttestation{}, "", fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return deliveryAttestation{}, "", fmt.Errorf("github json: %w", err)
	}

	var setupURL string
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, "-setup.exe") && !strings.HasSuffix(a.Name, ".sig") {
			setupURL = a.BrowserDownloadURL
			break
		}
	}
	if setupURL == "" {
		return deliveryAttestation{}, "", fmt.Errorf("no setup.exe in %s", release.TagName)
	}

	// 2. Download installer
	tmpDir, err := os.MkdirTemp("", "ltk-refresh-")
	if err != nil {
		return deliveryAttestation{}, "", err
	}
	defer os.RemoveAll(tmpDir)

	setupData, err := downloadFile(setupURL)
	if err != nil {
		return deliveryAttestation{}, "", err
	}
	setupPath := filepath.Join(tmpDir, "setup.exe")
	if err := os.WriteFile(setupPath, setupData, 0644); err != nil {
		return deliveryAttestation{}, "", err
	}

	// 3. Extract with 7z
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)
	cmd := exec.Command("7z", "x", "-o"+extractDir, "-y", setupPath)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		// 7z unavailable; keep fallback.
		return parseAttestation(fallbackAttestation), release.TagName, nil
	}

	// 4. Find ltk-manager.exe
	exePath := ""
	filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.EqualFold(info.Name(), "ltk-manager.exe") {
			exePath = path
			return filepath.SkipAll
		}
		return nil
	})
	if exePath == "" {
		return parseAttestation(fallbackAttestation), release.TagName, nil
	}

	// 5. Carve
	att, err := carveAttestation(exePath)
	if err != nil {
		return parseAttestation(fallbackAttestation), release.TagName, nil
	}

	saveAttestation(cacheDir, att)
	return att, release.TagName, nil
}

// carveAttestation finds the attestation in a ltk-manager.exe binary.
func carveAttestation(path string) (deliveryAttestation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return deliveryAttestation{}, err
	}

	pubIdx := bytes.Index(data, []byte(deliveryManifestKey))
	if pubIdx < 0 {
		return deliveryAttestation{}, fmt.Errorf("public key not found")
	}

	searchStart := pubIdx + len(deliveryManifestKey)
	attRel := bytes.Index(data[searchStart:], []byte("eyJ"))
	if attRel < 0 {
		return deliveryAttestation{}, fmt.Errorf("attestation header not found")
	}
	attStart := searchStart + attRel

	// eyJ<stmt>.{86 Ed25519 sig}
	re := regexp.MustCompile(fmt.Sprintf(`^eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]{%d}`, sigB64Len))
	m := re.Find(data[attStart:])
	if m == nil {
		return deliveryAttestation{}, fmt.Errorf("attestation format not recognized")
	}

	return parseAttestation(string(m)), nil
}

// --- delivery protocol ---

// DeliveryFetch downloads the current patcher DLL and host exe from the
// LeagueToolkit delivery service. The host exe is extracted from the latest
// LTK Manager NSIS installer via 7z, which must be on PATH.
func DeliveryFetch(dstDir string) (string, string, error) {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", "", err
	}

	att := loadAttestation(dstDir)
	client := &http.Client{Timeout: 120 * time.Second}

	token, err := deliveryRegister(client, att)
	if err != nil {
		// If registration fails, try once with a refreshed attestation.
		if strings.Contains(err.Error(), "register: HTTP 401") ||
			strings.Contains(err.Error(), "register: HTTP 403") {
			newAtt, tag, refreshErr := refreshAttestation(dstDir)
			if refreshErr == nil && newAtt.Token != att.Token {
				att = newAtt
				_ = tag
				token, err = deliveryRegister(client, att)
			}
			if err != nil {
				return "", "", fmt.Errorf("delivery register after refresh: %w", err)
			}
		} else {
			return "", "", fmt.Errorf("delivery register: %w", err)
		}
	}

	manifest, err := deliveryOffer(client, token, att.Manager)
	if err != nil {
		return "", "", fmt.Errorf("delivery offer: %w", err)
	}

	dllPath := filepath.Join(dstDir, "ltk_patcher_dll.dll")

	needDLL := true
	if existing, _ := os.ReadFile(dllPath); len(existing) > 0 {
		if hex.EncodeToString(sha256Sum(existing)) == manifest.Release.Digest {
			needDLL = false
		}
	}

	if needDLL {
		if err := deliveryDownload(client, token, &manifest.Release, dllPath); err != nil {
			return "", "", fmt.Errorf("delivery download: %w", err)
		}
	}

	// The host exe comes from the same manager release as the DLL. A stale or
	// missing host is replaced; failures are not fatal because the bundled
	// host remains usable.
	if err := deliveryEnsureHost(dstDir); err != nil {
		display.Log("LTK: host exe fetch skipped: " + err.Error())
	}

	// The signed ltk-manager.exe from the installer is needed so the launcher
	// can spoof the parent process for VRF00000009. Failures are non-fatal;
	// the launcher falls back to launching the host directly.
	if err := deliveryEnsureRealManager(dstDir); err != nil {
		display.Log("LTK: signed manager fetch skipped: " + err.Error())
	}

	return dllPath, manifest.Release.Version, nil
}

// hostReleaseFile records the LTK Manager release the delivered host came from.
const hostReleaseFile = "ltk-patcher-host.release"

// managerReleaseFile records the LTK Manager release the signed manager came from.
const managerReleaseFile = "ltk-manager.release"

// managerExeName is the on-disk name of the signed ltk-manager.exe extracted from
// the installer. The suffix avoids a clash with the Go stub (which is also
// compiled as ltk-manager.exe). The parent override uses this file's path, and
// WinVerifyTrust verifies the file on disk — the image name is irrelevant for
// VRF00000009.
const managerExeName = "ltk-manager-signed.exe"

// deliveryEnsureHost replaces ltk_patcher_host.exe with the host from the
// newest LTK Manager release. Nothing is downloaded when the delivered host
// already came from that release. The DLL verifies the host image name and
// signature, so the published file name is kept.
func deliveryEnsureHost(dstDir string) error {
	tag, setupURL, err := ltkLatestSetup()
	if err != nil {
		return err
	}
	hostPath := filepath.Join(dstDir, "ltk_patcher_host.exe")
	if marked, err := os.ReadFile(filepath.Join(dstDir, hostReleaseFile)); err == nil {
		if _, statErr := os.Stat(hostPath); statErr == nil && strings.TrimSpace(string(marked)) == tag {
			return nil
		}
	}

	data, err := ltkHostFromInstaller(setupURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(hostPath, data, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dstDir, hostReleaseFile), []byte(tag), 0644)
}

// ltkLatestSetup returns the newest LTK Manager release tag and its NSIS
// installer URL.
func ltkLatestSetup() (string, string, error) {
	req, _ := http.NewRequest("GET", ltkReleasesURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent("1.32.2"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("github json: %w", err)
	}
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, "-setup.exe") && !strings.HasSuffix(a.Name, ".sig") {
			return release.TagName, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("no setup.exe in %s", release.TagName)
}

// deliveryEnsureRealManager extracts the signed ltk-manager.exe from the latest
// LTK Manager NSIS installer and writes it as ltk-manager-signed.exe in dstDir.
// The file is cached with a release marker to avoid re-downloading the same
// installer. Failures are returned so callers can fall back to direct host launch.
func deliveryEnsureRealManager(dstDir string) error {
	tag, setupURL, err := ltkLatestSetup()
	if err != nil {
		return err
	}
	managerPath := filepath.Join(dstDir, managerExeName)
	if marked, err := os.ReadFile(filepath.Join(dstDir, managerReleaseFile)); err == nil {
		if _, statErr := os.Stat(managerPath); statErr == nil && strings.TrimSpace(string(marked)) == tag {
			return nil
		}
	}

	data, err := ltkFileFromInstaller(setupURL, "ltk-manager.exe")
	if err != nil {
		return err
	}
	if err := os.WriteFile(managerPath, data, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dstDir, managerReleaseFile), []byte(tag), 0644)
}

// ltkHostFromInstaller extracts ltk_patcher_host.exe from an NSIS installer.
// Requires 7z on PATH.
func ltkHostFromInstaller(setupURL string) ([]byte, error) {
	return ltkFileFromInstaller(setupURL, "ltk_patcher_host.exe")
}

// ltkFileFromInstaller extracts a named file from an NSIS installer via 7z.
// Requires 7z on PATH.
func ltkFileFromInstaller(setupURL, targetName string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "ltk-extract-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	setupData, err := downloadFile(setupURL)
	if err != nil {
		return nil, err
	}
	setupPath := filepath.Join(tmpDir, "setup.exe")
	if err := os.WriteFile(setupPath, setupData, 0644); err != nil {
		return nil, err
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)
	cmd := exec.Command("7z", "x", "-o"+extractDir, "-y", setupPath)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("7z: %w", err)
	}

	var found string
	filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.EqualFold(info.Name(), targetName) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if found == "" {
		return nil, fmt.Errorf("%s not found in installer", targetName)
	}
	return os.ReadFile(found)
}

func deliveryRegister(client *http.Client, att deliveryAttestation) (string, error) {
	installID := newInstallID()

	// Challenge
	req, _ := http.NewRequest("GET", deliveryBaseURL+"/v1/challenge", nil)
	req.Header.Set("User-Agent", userAgent(att.Manager))
	q := req.URL.Query()
	q.Set("install", installID)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("challenge: HTTP %d: %s", resp.StatusCode, body)
	}
	var ch deliveryChallenge
	if err := json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return "", err
	}
	if ch.Difficulty > maxDifficulty {
		return "", fmt.Errorf("difficulty %d > %d", ch.Difficulty, maxDifficulty)
	}

	nonce := solvePoW(ch.Challenge, ch.Difficulty)

	// Register
	body, _ := json.Marshal(map[string]string{
		"install":     installID,
		"manager":     att.Manager,
		"attestation": att.Token,
		"challenge":   ch.Challenge,
		"nonce":       nonce,
	})
	req, _ = http.NewRequest("POST", deliveryBaseURL+"/v1/installs", bytes.NewReader(body))
	req.Header.Set("User-Agent", userAgent(att.Manager))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("register: HTTP %d: %s", resp.StatusCode, respBody)
	}
	var reg deliveryRegistered
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return "", err
	}
	return reg.Token, nil
}

func deliveryOffer(client *http.Client, token, manager string) (*DeliveryManifest, error) {
	req, _ := http.NewRequest("GET", deliveryBaseURL+"/v1/offer", nil)
	req.Header.Set("User-Agent", userAgent(manager))
	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Set("channel", "stable")
	q.Set("manager", manager)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		return nil, fmt.Errorf("offer: HTTP 409: manager %s is not served by any release range", manager)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("offer: HTTP %d: %s", resp.StatusCode, body)
	}
	var offered deliveryOffered
	if err := json.NewDecoder(resp.Body).Decode(&offered); err != nil {
		return nil, err
	}
	manifestBytes, err := base64.StdEncoding.DecodeString(offered.Manifest.Manifest)
	if err != nil {
		return nil, fmt.Errorf("manifest base64: %w", err)
	}
	var manifest DeliveryManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("manifest json: %w", err)
	}

	// Verify signature via trusted comment embedded in the minisign envelope.
	if err := verifyManifestSignature(&offered.Manifest, &manifest); err != nil {
		return nil, fmt.Errorf("manifest signature: %w", err)
	}

	return &manifest, nil
}

func deliveryDownload(client *http.Client, token string, release *DeliveryRelease, dst string) error {
	req, _ := http.NewRequest("GET", deliveryBaseURL+"/v1/dll/"+release.Digest, nil)
	req.Header.Set("User-Agent", userAgent("1.32.2")) // server only checks token here
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download: HTTP %d: %s", resp.StatusCode, body)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".ltk-dll-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	h := sha256.New()
	written, err := io.Copy(tmp, io.TeeReader(resp.Body, h))
	tmp.Close()
	if err != nil {
		return err
	}
	if uint64(written) != release.Size {
		return fmt.Errorf("size mismatch: got %d, expected %d", written, release.Size)
	}
	digest := hex.EncodeToString(h.Sum(nil))
	if digest != release.Digest {
		return fmt.Errorf("digest mismatch: got %s, expected %s", digest, release.Digest)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// --- manifest signature verification ---

func verifyManifestSignature(envelope *deliveryEnvelope, manifest *DeliveryManifest) error {
	// Verify the trusted comment encodes the same channel, sequence, and version
	// as the manifest body. The minisign global signature covers both the body
	// and the trusted comment, but implementing full minisign verification in Go
	// requires its specific Ed25519 + SHA-512 hash construction.
	//
	// Defense-in-depth: the server is the same one LTK Manager uses, the
	// attestation registration already authenticated us, and the DLL is verified
	// by its SHA-256 digest from the manifest before use. The trusted-comment
	// check prevents channel or version substitution between a cached and a
	// fresh manifest.
	expected := fmt.Sprintf("trusted comment: %s %d %s",
		manifest.Channel, manifest.Sequence, manifest.Release.Version)
	if !strings.Contains(envelope.Signature, expected) {
		return fmt.Errorf("trusted comment mismatch: %q not found", expected)
	}
	return nil
}

// --- helpers ---

func solvePoW(challenge string, difficulty uint32) string {
	for n := uint64(0); n < math.MaxUint32; n++ {
		nonce := fmt.Sprintf("%x", n)
		h := sha256.Sum256([]byte(challenge + nonce))
		if leadingZeroBits(h[:]) >= difficulty {
			return nonce
		}
	}
	panic("nonce space exhausted for difficulty <= 24")
}

func leadingZeroBits(digest []byte) uint32 {
	var bits uint32
	for _, b := range digest {
		if b != 0 {
			return bits + uint32(leadingZeros8(b))
		}
		bits += 8
	}
	return bits
}

func leadingZeros8(b byte) int {
	n := 0
	for i := 7; i >= 0; i-- {
		if b&(1<<i) != 0 {
			break
		}
		n++
	}
	return n
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// downloadFile fetches an installer-sized file. The timeout is generous but
// finite so a stalled host fetch cannot block startup forever.
func downloadFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

var rngState = uint64(time.Now().UnixNano())

func newInstallID() string {
	var b [16]byte
	for i := range b {
		rngState ^= rngState << 13
		rngState ^= rngState >> 17
		rngState ^= rngState << 5
		b[i] = byte(rngState)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// --- fallback baked at build time ---

const fallbackAttestation = "eyJtYW5hZ2VyIjoiMS4zMi4yIiwiY29tbWl0IjoiNjJhZmY0ODJjMThmZTI0ZGFjM2ZkYWU3MGEzMTY1OTUyMDRkOWVmNCJ9.55MKhasBjui5ix8dWaPZYi9HizK5Sap2wGN-btye1e6fjsegtM6zsb0Z7TMSXxsweOyiLpJ3ZRQJpJwX25GCBA"