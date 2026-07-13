// Package client talks to the homelab backend HTTP API.
//
// The client is intentionally thin: it owns the base URL, the optional API key,
// the require-auth gate, and the single Do() entry point used by every command.
// Auth handling lives here so callers never think about headers. The Bearer
// header is attached only when requireAuth is true (production) AND the request
// is not to /api/health. In dev (requireAuth=false) no Authorization header is
// ever sent, mirroring the backend's dev=open, prod=authenticated gate.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const healthPath = "/api/health"

// Client is a minimal HTTP client for the homelab backend.
type Client struct {
	baseURL     string
	apiKey      string
	requireAuth bool
	http        *http.Client
}

// New returns a client rooted at the given base URL. requireAuth mirrors the
// backend's auth gate: when true (production) Do attaches
// "Authorization: Bearer <apiKey>" to every request except /api/health; when
// false (development) no Authorization header is ever sent, even if apiKey is
// non-empty — dev is fully open. An empty apiKey is permitted in dev; in prod
// callers are expected to enforce a non-empty key before constructing a client.
func New(baseURL, apiKey string, requireAuth bool) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      apiKey,
		requireAuth: requireAuth,
		http:        &http.Client{},
	}
}

// Do performs an HTTP request against path (e.g. "/api/todos"). body is sent as
// JSON; pass nil for requests without a body. The Authorization header is
// attached only when requireAuth is true and path is not /api/health; /api/health
// is always unauthenticated regardless of mode. Non-2xx responses are returned
// as an error carrying the server's {"error": "..."} message when present.
func (c *Client) Do(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.requireAuth && path != healthPath {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return raw, resp.StatusCode, errorFromResponse(raw, resp.StatusCode)
	}

	return raw, resp.StatusCode, nil
}

// errorFromResponse turns a non-2xx body into an actionable error. The backend
// emits {"error": "<msg>"} for JSON errors and plain text "method not allowed"
// for 405s; both shapes are surfaced with the status code prefixed for context.
func errorFromResponse(body []byte, status int) error {
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err == nil && env.Error != "" {
		return fmt.Errorf("server returned %d: %s", status, env.Error)
	}
	if msg := strings.TrimSpace(string(body)); msg != "" {
		return fmt.Errorf("server returned %d: %s", status, msg)
	}
	return fmt.Errorf("server returned %d", status)
}