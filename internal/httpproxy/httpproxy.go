package httpproxy

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/display"
)

// No per-domain rate limiting — extensions need to make many concurrent
// requests (thumbnails, API calls, etc.) and the upstream servers handle
// their own throttling.

// Request is the WS message from the frontend asking to proxy an HTTP request.
type Request struct {
	Type      string            `json:"type"` // "httpProxy"
	RequestID string            `json:"requestId"`
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
}

// Response is the WS message sent back with the proxied result.
type Response struct {
	Type      string            `json:"type"` // "httpProxyResult"
	RequestID string            `json:"requestId"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body"`
	IsBase64  bool              `json:"isBase64"`
	Error     string            `json:"error,omitempty"`
}

const (
	maxResponseBody = 5 * 1024 * 1024 // 5MB
	requestTimeout  = 30 * time.Second
)

// httpClient forces IPv4 + HTTP/1.1 to avoid Cloudflare connection resets.
var httpClient = &http.Client{
	Timeout: requestTimeout,
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

// HandleProxy handles a proxied HTTP request from the frontend via WebSocket.
func HandleProxy(conn *websocket.Conn, raw json.RawMessage) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return
	}

	go func() {
		resp := doProxy(req)
		data, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, data)
	}()
}

func doProxy(req Request) Response {
	resp := Response{Type: "httpProxyResult", RequestID: req.RequestID}

	if _, err := url.Parse(req.URL); err != nil {
		resp.Error = "invalid URL"
		return resp
	}

	method := req.Method
	if method == "" {
		method = "GET"
	}

	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		var bodyReader io.Reader
		if req.Body != "" {
			bodyReader = strings.NewReader(req.Body)
		}

		httpReq, err := http.NewRequest(method, req.URL, bodyReader)
		if err != nil {
			resp.Error = fmt.Sprintf("request error: %v", err)
			return resp
		}

		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
		if httpReq.Header.Get("User-Agent") == "" {
			httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		}

		httpResp, err := httpClient.Do(httpReq)
		if err != nil {
			if attempt < maxRetries-1 && isRetryableErr(err) {
				continue
			}
			resp.Error = fmt.Sprintf("fetch error: %v", err)
			return resp
		}

		body, err := io.ReadAll(io.LimitReader(httpResp.Body, maxResponseBody))
		httpResp.Body.Close()
		if err != nil {
			if attempt < maxRetries-1 {
				continue
			}
			resp.Error = fmt.Sprintf("read error: %v", err)
			return resp
		}

		resp.Status = httpResp.StatusCode
		resp.Headers = make(map[string]string)
		for k := range httpResp.Header {
			resp.Headers[strings.ToLower(k)] = httpResp.Header.Get(k)
		}

		ct := httpResp.Header.Get("Content-Type")
		if isTextContent(ct) {
			resp.Body = string(body)
			resp.IsBase64 = false
		} else {
			resp.Body = base64.StdEncoding.EncodeToString(body)
			resp.IsBase64 = true
		}

		return resp
	}

	return resp
}

func isRetryableErr(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "closed") ||
		strings.Contains(msg, "timeout")
}

func isTextContent(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.HasPrefix(ct, "text/") ||
		strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "application/xml") ||
		strings.Contains(ct, "application/javascript")
}

// --- Image proxy HTTP endpoint ---

type cacheEntry struct {
	data      []byte
	ct        string
	expiresAt time.Time
}

var (
	imageCache   = make(map[string]*cacheEntry)
	imageCacheMu sync.Mutex
)

// ServeImageProxy is the HTTP handler for GET /proxy-image?url=<encoded>
func ServeImageProxy(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check cache
	imageCacheMu.Lock()
	entry, ok := imageCache[rawURL]
	if ok && time.Now().Before(entry.expiresAt) {
		imageCacheMu.Unlock()
		w.Header().Set("Content-Type", entry.ct)
		w.Header().Set("Cache-Control", "public, max-age=600")
		w.Write(entry.data)
		return
	}
	imageCacheMu.Unlock()

	if _, err := url.Parse(rawURL); err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	var body []byte
	var ct string
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			http.Error(w, "request error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries-1 && isRetryableErr(err) {
				continue
			}
			display.Log(fmt.Sprintf("Image proxy error: %v", err))
			http.Error(w, "fetch error", http.StatusBadGateway)
			return
		}

		body, err = io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries-1 {
				continue
			}
			http.Error(w, "read error", http.StatusBadGateway)
			return
		}

		ct = resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}
		break
	}

	// Cache for 10 minutes
	imageCacheMu.Lock()
	imageCache[rawURL] = &cacheEntry{
		data:      body,
		ct:        ct,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	// Evict expired entries (simple cleanup)
	now := time.Now()
	for k, v := range imageCache {
		if now.After(v.expiresAt) {
			delete(imageCache, k)
		}
	}
	imageCacheMu.Unlock()

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Write(body)
}
