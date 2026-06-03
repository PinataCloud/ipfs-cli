// Package httpclient provides a shared HTTP client that identifies the CLI
// with a User-Agent header on every outgoing request.
package httpclient

import (
	"net/http"
	"time"

	"pinata/internal/version"
)

// userAgentTransport wraps a base RoundTripper and sets the CLI User-Agent
// on every request.
type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("User-Agent", version.UserAgent())
	return t.base.RoundTrip(r)
}

// Transport is the shared RoundTripper used by all CLI HTTP clients.
var Transport http.RoundTripper = &userAgentTransport{base: http.DefaultTransport}

// Client is the shared HTTP client for Pinata API requests.
var Client = &http.Client{Transport: Transport}

// WithTimeout returns a client that uses the shared transport with a timeout.
func WithTimeout(d time.Duration) *http.Client {
	return &http.Client{Transport: Transport, Timeout: d}
}
