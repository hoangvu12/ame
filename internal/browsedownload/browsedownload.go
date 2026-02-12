package browsedownload

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/custommods"
	"github.com/hoangvu12/ame/internal/display"
)

// browserClient forces IPv4 + HTTP/1.1 to avoid Cloudflare connection resets.
var browserClient = &http.Client{
	Timeout: 5 * time.Minute,
	Transport: &http.Transport{
		TLSClientConfig:   &tls.Config{},
		ForceAttemptHTTP2:  false,
		DisableKeepAlives:  false,
		TLSNextProto:       make(map[string]func(string, *tls.Conn) http.RoundTripper),
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, "tcp4", addr)
		},
	},
}

// DownloadRequest is the WS message from frontend to start a mod download.
type DownloadRequest struct {
	Type         string `json:"type"` // "browseDownload"
	RequestID    string `json:"requestId"`
	URL          string `json:"url"`
	Name         string `json:"name"`
	Author       string `json:"author"`
	ChampionID   int    `json:"championId"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

// ProgressMessage is sent to frontend to report download/import progress.
type ProgressMessage struct {
	Type      string      `json:"type"` // "browseDownloadProgress"
	RequestID string      `json:"requestId"`
	Phase     string      `json:"phase"` // "downloading", "importing", "done", "error"
	Progress  float64     `json:"progress"`
	Error     string      `json:"error,omitempty"`
	Mod       interface{} `json:"mod,omitempty"`
}

// HandleDownload starts downloading a mod file from URL and importing it.
func HandleDownload(conn *websocket.Conn, raw json.RawMessage) {
	var req DownloadRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return
	}

	go func() {
		sendProgress(conn, req.RequestID, "downloading", 0, "", nil)

		tmpDir := filepath.Join(config.AmeDir, "tmp")
		os.MkdirAll(tmpDir, os.ModePerm)
		tmpFile, err := os.CreateTemp(tmpDir, "browse-*.zip")
		if err != nil {
			sendProgress(conn, req.RequestID, "error", 0, "Failed to create temp file", nil)
			return
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		const maxRetries = 3
		var dlErr error

		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * time.Second)
				sendProgress(conn, req.RequestID, "downloading", 0, "", nil)
				// Reset temp file for retry
				tmpFile.Seek(0, 0)
				tmpFile.Truncate(0)
			}

			httpReq, _ := http.NewRequest("GET", req.URL, nil)
			httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
			httpReq.Header.Set("Accept", "*/*")

			resp, err := browserClient.Do(httpReq)
			if err != nil {
				dlErr = err
				if attempt < maxRetries-1 && isRetryable(err) {
					display.Log(fmt.Sprintf("Browse download retry %d: %v", attempt+1, err))
					continue
				}
				tmpFile.Close()
				display.Log(fmt.Sprintf("Browse download error: %v", err))
				sendProgress(conn, req.RequestID, "error", 0, "Download failed", nil)
				return
			}

			totalSize := int64(-1)
			if cl := resp.Header.Get("Content-Length"); cl != "" {
				if n, err := strconv.ParseInt(cl, 10, 64); err == nil {
					totalSize = n
				}
			}

			var downloaded int64
			buf := make([]byte, 32*1024)
			lastReport := time.Now()
			readFailed := false

			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
						resp.Body.Close()
						tmpFile.Close()
						sendProgress(conn, req.RequestID, "error", 0, "Write error", nil)
						return
					}
					downloaded += int64(n)

					if time.Since(lastReport) > 200*time.Millisecond {
						var pct float64 = -1
						if totalSize > 0 {
							pct = float64(downloaded) / float64(totalSize) * 100
						}
						sendProgress(conn, req.RequestID, "downloading", pct, "", nil)
						lastReport = time.Now()
					}
				}
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					resp.Body.Close()
					dlErr = readErr
					if attempt < maxRetries-1 {
						display.Log(fmt.Sprintf("Browse download retry %d: %v", attempt+1, readErr))
						readFailed = true
						break
					}
					tmpFile.Close()
					sendProgress(conn, req.RequestID, "error", 0, "Download interrupted", nil)
					return
				}
			}
			resp.Body.Close()

			if readFailed {
				continue
			}

			dlErr = nil
			break
		}

		tmpFile.Close()

		if dlErr != nil {
			display.Log(fmt.Sprintf("Browse download failed after retries: %v", dlErr))
			sendProgress(conn, req.RequestID, "error", 0, "Download failed after retries", nil)
			return
		}

		// Optionally download thumbnail
		var thumbnailBase64 string
		if req.ThumbnailURL != "" {
			thumbnailBase64 = downloadThumbnail(req.ThumbnailURL)
		}

		sendProgress(conn, req.RequestID, "importing", -1, "", nil)

		mod, err := custommods.ImportMod(tmpPath, req.Name, req.Author, req.ChampionID, thumbnailBase64)
		if err != nil {
			display.Log(fmt.Sprintf("Browse import error: %v", err))
			sendProgress(conn, req.RequestID, "error", 0, fmt.Sprintf("Import failed: %v", err), nil)
			return
		}

		display.Log(fmt.Sprintf("Browse: imported mod %s (%s)", mod.Name, mod.ID))
		sendProgress(conn, req.RequestID, "done", 100, "", mod)

		added := map[string]interface{}{
			"type": "customModAdded",
			"mod":  mod,
		}
		data, _ := json.Marshal(added)
		conn.WriteMessage(websocket.TextMessage, data)
	}()
}

func sendProgress(conn *websocket.Conn, requestID, phase string, progress float64, errMsg string, mod interface{}) {
	msg := ProgressMessage{
		Type:      "browseDownloadProgress",
		RequestID: requestID,
		Phase:     phase,
		Progress:  progress,
		Error:     errMsg,
		Mod:       mod,
	}
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func downloadThumbnail(imgURL string) string {
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		req, _ := http.NewRequest("GET", imgURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := browserClient.Do(req)
		if err != nil {
			if attempt < 2 && isRetryable(err) {
				continue
			}
			return ""
		}

		data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
		resp.Body.Close()
		if err != nil {
			if attempt < 2 {
				continue
			}
			return ""
		}

		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "image/png"
		}

		return fmt.Sprintf("data:%s;base64,%s", ct, base64.StdEncoding.EncodeToString(data))
	}

	return ""
}

func isRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "closed") ||
		strings.Contains(msg, "timeout")
}
