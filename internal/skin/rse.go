package skin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

// Injected at build time via -ldflags
var (
	rseKeyURL       string
	rseSkinBaseURL  string
	rseSkinIDsURL   string
	rseAuthSecret   string
	rseClientHalf   string
)

var (
	rseKeyCache []byte
	rseKeyMu    sync.Mutex

	rseClockOnce  sync.Once
	rseClockDrift int64
)

var rseMagic = []byte{'R', 'S', 'E', 0x01}

const (
	rseNonceSize    = 16
	hmacBlockSize   = 32
	rseResponseSize = 48
	rseUserAgent    = "Rose"
	rseSkinKeyPath  = "/skin-key"
	rseSignSuffix   = ":" + rseSkinKeyPath
	rseHeaderNonce  = "X-Rose-N"
	rseHeaderTime   = "X-Rose-T"
	rseHeaderSig    = "X-Rose-S"
	rseHeaderSrvTm  = "X-Server-Time"
)

var rseClient = &http.Client{Timeout: 30 * time.Second}

func init() {
	envVars := loadEnvFile()
	if rseKeyURL == "" {
		rseKeyURL = envVars["RSE_KEY_URL"]
	}
	if rseSkinBaseURL == "" {
		rseSkinBaseURL = envVars["RSE_SKIN_BASE_URL"]
	}
	if rseSkinIDsURL == "" {
		rseSkinIDsURL = envVars["RSE_SKIN_IDS_URL"]
	}
	if rseAuthSecret == "" {
		rseAuthSecret = envVars["RSE_AUTH_SECRET"]
	}
	if rseClientHalf == "" {
		rseClientHalf = envVars["RSE_CLIENT_HALF"]
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
	return rseKeyURL != "" && rseSkinBaseURL != "" && rseAuthSecret != "" && rseClientHalf != ""
}

func rseKeyBaseURL() string {
	return strings.TrimSuffix(strings.TrimSuffix(rseKeyURL, rseSkinKeyPath), "/")
}

func syncRSEClock() {
	rseClockOnce.Do(func() {
		req, err := http.NewRequest("GET", rseKeyBaseURL()+"/", nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", rseUserAgent)

		resp, err := rseClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		serverTimeStr := strings.TrimSpace(resp.Header.Get(rseHeaderSrvTm))
		if serverTimeStr == "" {
			return
		}
		serverTime, err := strconv.ParseInt(serverTimeStr, 10, 64)
		if err != nil {
			return
		}
		rseClockDrift = serverTime - time.Now().Unix()
		if rseClockDrift != 0 {
			logSkin(fmt.Sprintf("RSE clock drift detected: %+ds (compensating)", rseClockDrift))
		}
	})
}

func fetchRSEKey() ([]byte, error) {
	rseKeyMu.Lock()
	defer rseKeyMu.Unlock()

	if rseKeyCache != nil {
		return rseKeyCache, nil
	}
	logSkin("fetching RSE key")

	authSecret, err := hex.DecodeString(rseAuthSecret)
	if err != nil || len(authSecret) != 32 {
		return nil, fmt.Errorf("invalid auth secret")
	}
	clientHalf, err := hex.DecodeString(rseClientHalf)
	if err != nil || len(clientHalf) != 32 {
		return nil, fmt.Errorf("invalid client half")
	}

	syncRSEClock()

	nonceBytes := make([]byte, rseNonceSize)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("nonce generation failed: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	timestamp := strconv.FormatInt(time.Now().Unix()+rseClockDrift, 10)

	signMsg := timestamp + ":" + nonce + rseSignSuffix
	mac := hmac.New(sha256.New, authSecret)
	mac.Write([]byte(signMsg))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", rseKeyURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", rseUserAgent)
	req.Header.Set(rseHeaderNonce, nonce)
	req.Header.Set(rseHeaderTime, timestamp)
	req.Header.Set(rseHeaderSig, signature)

	resp, err := rseClient.Do(req)
	if err != nil {
		logSkin(fmt.Sprintf("RSE key request failed: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logSkin(fmt.Sprintf("RSE key request failed (status %d)", resp.StatusCode))
		return nil, fmt.Errorf("key server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return nil, err
	}

	respBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, fmt.Errorf("invalid base64 response: %w", err)
	}
	if len(respBytes) != rseResponseSize {
		return nil, fmt.Errorf("unexpected response size: %d (expected %d)", len(respBytes), rseResponseSize)
	}

	responseNonce := respBytes[:rseNonceSize]
	encryptedHalf := respBytes[rseNonceSize:]

	rkMac := hmac.New(sha256.New, authSecret)
	rkMac.Write([]byte(nonce + ":" + timestamp + ":rk"))
	responseKey := rkMac.Sum(nil)

	keystream := rseKeystream(responseKey, responseNonce, hmacBlockSize)

	skinKey := make([]byte, hmacBlockSize)
	for i := 0; i < hmacBlockSize; i++ {
		skinKey[i] = encryptedHalf[i] ^ keystream[i] ^ clientHalf[i]
	}

	rseKeyCache = skinKey
	logSkin("RSE key ready")
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
	path := rseSkinPath(championID, skinID, baseSkinID)
	logSkin(fmt.Sprintf("downloading %s", path))
	remote, _ := fetchRSESkinVersion(championID, skinID, baseSkinID)

	resp, err := rseClient.Get(rseSkinURL(championID, skinID, baseSkinID))
	if err != nil {
		logSkin(fmt.Sprintf("download request failed for %s: %v", path, err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logSkin(fmt.Sprintf("download request failed for %s (status %d)", path, resp.StatusCode))
		return "", fmt.Errorf("rse download returned status %d", resp.StatusCode)
	}
	if remote == nil {
		remote = &remoteSkinVersion{
			Path:         rseSkinPath(championID, skinID, baseSkinID),
			ETag:         resp.Header.Get("ETag"),
			LastModified: resp.Header.Get("Last-Modified"),
			Size:         resp.ContentLength,
		}
	}

	encrypted, err := io.ReadAll(resp.Body)
	if err != nil {
		logSkin(fmt.Sprintf("failed reading %s: %v", path, err))
		return "", err
	}
	logSkin(fmt.Sprintf("downloaded %s (%d bytes encrypted)", path, len(encrypted)))

	decrypted, err := decryptRSE(encrypted)
	if err != nil {
		logSkin(fmt.Sprintf("failed decrypting %s: %v", path, err))
		return "", err
	}

	skinDir := filepath.Join(config.SkinsDir, championID, skinID)
	if err := os.MkdirAll(skinDir, os.ModePerm); err != nil {
		logSkin(fmt.Sprintf("failed creating cache dir for %s: %v", path, err))
		return "", err
	}

	filePath := filepath.Join(skinDir, skinID+".zip")
	if err := os.WriteFile(filePath, decrypted, 0644); err != nil {
		logSkin(fmt.Sprintf("failed writing cache for %s: %v", path, err))
		return "", err
	}
	writeCacheMetadata(championID, skinID, remote)
	logSkin(fmt.Sprintf("cached %s", path))

	return filePath, nil
}
