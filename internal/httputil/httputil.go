package httputil

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"
)

// NewIPv4Client returns an HTTP client that forces IPv4 + HTTP/1.1 to avoid
// Cloudflare connection resets.
func NewIPv4Client(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig:  &tls.Config{},
			ForceAttemptHTTP2: false,
			DisableKeepAlives: false,
			TLSNextProto:      make(map[string]func(string, *tls.Conn) http.RoundTripper),
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, "tcp4", addr)
			},
		},
	}
}

// IsRetryableErr returns true if the error looks like a transient network issue
// worth retrying.
func IsRetryableErr(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "closed") ||
		strings.Contains(msg, "timeout")
}
