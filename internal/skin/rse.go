package skin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// Injected at build time via -ldflags
var (
	rseKeyURL      string
	rseSkinBaseURL string
)

var (
	rseKeyCache []byte
	rseKeyMu    sync.Mutex
)

var rseMagic = []byte{'R', 'S', 'E', 0x01}

const rseNonceSize = 16
const hmacBlockSize = 32

var rseClient = &http.Client{Timeout: 30 * time.Second}

func init() {
	envVars := loadEnvFile()
	if rseKeyURL == "" {
		rseKeyURL = envVars["RSE_KEY_URL"]
	}
	if rseSkinBaseURL == "" {
		rseSkinBaseURL = envVars["RSE_SKIN_BASE_URL"]
	}
}

func loadEnvFile() map[string]string {
	vars := make(map[string]string)
	exe, err := os.Executable()
	if err != nil {
		return vars
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(exe), ".env"))
	if err != nil {
		return vars
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
		}
		vars[strings.TrimSpace(k)] = v
	}
	return vars
}

func rseAvailable() bool {
	return rseKeyURL != "" && rseSkinBaseURL != ""
}

func fetchRSEKey() ([]byte, error) {
	rseKeyMu.Lock()
	defer rseKeyMu.Unlock()

	if rseKeyCache != nil {
		return rseKeyCache, nil
	}

	req, err := http.NewRequest("GET", rseKeyURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rose")

	resp, err := rseClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("key server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, fmt.Errorf("invalid key format: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d", len(key))
	}

	rseKeyCache = key
	return rseKeyCache, nil
}

func rseKeystream(key, nonce []byte, length int) []byte {
	stream := make([]byte, 0, length+hmacBlockSize)
	buf := make([]byte, rseNonceSize+8)
	copy(buf, nonce)

	mac := hmac.New(sha256.New, key)
	counter := uint64(0)
	for len(stream) < length {
		binary.LittleEndian.PutUint64(buf[rseNonceSize:], counter)
		mac.Reset()
		mac.Write(buf)
		stream = append(stream, mac.Sum(nil)...)
		counter++
	}

	return stream[:length]
}

func decryptRSE(data []byte) ([]byte, error) {
	headerSize := len(rseMagic) + rseNonceSize
	if len(data) < headerSize {
		return nil, fmt.Errorf("rse data too short")
	}

	for i, b := range rseMagic {
		if data[i] != b {
			return nil, fmt.Errorf("invalid rse magic")
		}
	}

	key, err := fetchRSEKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get decryption key: %w", err)
	}

	nonce := data[len(rseMagic) : len(rseMagic)+rseNonceSize]
	ciphertext := data[headerSize:]
	ks := rseKeystream(key, nonce, len(ciphertext))

	for i := range ciphertext {
		ciphertext[i] ^= ks[i]
	}

	return ciphertext, nil
}

func downloadRSE(championID, skinID, baseSkinID string) (string, error) {
	var downloadURL string
	if baseSkinID != "" && baseSkinID != "0" {
		downloadURL = fmt.Sprintf("%s/%s/%s/%s/%s.rse", rseSkinBaseURL, championID, baseSkinID, skinID, skinID)
	} else {
		downloadURL = fmt.Sprintf("%s/%s/%s/%s.rse", rseSkinBaseURL, championID, skinID, skinID)
	}

	resp, err := rseClient.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("rse download returned status %d", resp.StatusCode)
	}

	encrypted, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decrypted, err := decryptRSE(encrypted)
	if err != nil {
		return "", err
	}

	skinDir := filepath.Join(config.SkinsDir, championID, skinID)
	if err := os.MkdirAll(skinDir, os.ModePerm); err != nil {
		return "", err
	}

	filePath := filepath.Join(skinDir, skinID+".zip")
	if err := os.WriteFile(filePath, decrypted, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}
