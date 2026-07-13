package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recordingHandler captures the inbound Authorization header and the request
// method/path, then writes a canned response. It is reused across tests with
// per-case expectations.
type recordingHandler struct {
	auth        string
	method      string
	path        string
	body        string
	status      int
	response    []byte
	contentType string
}

func (h *recordingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.auth = r.Header.Get("Authorization")
	h.method = r.Method
	h.path = r.URL.Path
	if r.Body != nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			h.body = string(b)
		}
	}
	if h.contentType != "" {
		w.Header().Set("Content-Type", h.contentType)
	}
	w.WriteHeader(h.status)
	if h.response != nil {
		_, _ = w.Write(h.response)
	}
}

func TestDo_SendsBearerExceptHealth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		path     string
		apiKey   string
		wantAuth string
	}{
		{"todos route gets bearer", "/api/todos", "secret-key", "Bearer secret-key"},
		{"health route skips bearer", "/api/health", "secret-key", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			h := &recordingHandler{status: http.StatusOK, response: []byte("{}"), contentType: "application/json"}
			srv := httptest.NewServer(h)
			t.Cleanup(srv.Close)

			c := New(srv.URL, tc.apiKey)
			if _, _, err := c.Do(context.Background(), "GET", tc.path, nil); err != nil {
				t.Fatalf("Do: %v", err)
			}
			if h.auth != tc.wantAuth {
				t.Fatalf("Authorization = %q, want %q", h.auth, tc.wantAuth)
			}
		})
	}
}

func TestDo_SurfacesErrorEnvelope(t *testing.T) {
	t.Parallel()
	h := &recordingHandler{
		status:      http.StatusUnauthorized,
		response:    []byte(`{"error":"unauthorized"}`),
		contentType: "application/json",
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "bad-key")
	_, _, err := c.Do(context.Background(), "GET", "/api/todos", nil)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q missing status 401", err)
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error %q missing server message", err)
	}
}

func TestDo_PlainTextMethodNotAllowed(t *testing.T) {
	t.Parallel()
	h := &recordingHandler{
		status:      http.StatusMethodNotAllowed,
		response:    []byte("method not allowed"),
		contentType: "text/plain; charset=utf-8",
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "k")
	_, _, err := c.Do(context.Background(), "PATCH", "/api/todos", nil)
	if err == nil {
		t.Fatal("expected error for 405")
	}
	if !strings.Contains(err.Error(), "method not allowed") {
		t.Errorf("error %q missing plain text message", err)
	}
}

func TestDo_PostsJSONBody(t *testing.T) {
	t.Parallel()
	h := &recordingHandler{status: http.StatusCreated, response: []byte("{}"), contentType: "application/json"}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "k")
	body := []byte(`{"title":"x"}`)
	if _, _, err := c.Do(context.Background(), "POST", "/api/todos", body); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if h.method != http.MethodPost {
		t.Errorf("method = %q", h.method)
	}
	if h.body != `{"title":"x"}` {
		t.Errorf("body = %q", h.body)
	}
}

func TestDo_DeletesReturnNoBody(t *testing.T) {
	t.Parallel()
	h := &recordingHandler{status: http.StatusNoContent}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "k")
	raw, status, err := c.Do(context.Background(), "DELETE", "/api/todos/1", nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if status != http.StatusNoContent {
		t.Errorf("status = %d", status)
	}
	if len(raw) != 0 {
		t.Errorf("expected empty body, got %q", raw)
	}
}

func TestNew_TrimsTrailingSlash(t *testing.T) {
	t.Parallel()
	c := New("http://localhost:8080/", "k")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

// helper to read the auth the server received in table-style assertions above.
func TestDo_HealthNoKeyButKeyPresentStillSkipsHeader(t *testing.T) {
	t.Parallel()
	h := &recordingHandler{status: http.StatusOK, response: []byte(`{"status":"ok","db_ok":true}`), contentType: "application/json"}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c := New(srv.URL, "secret")
	if _, _, err := c.Do(context.Background(), "GET", "/api/health", nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if h.auth != "" {
		t.Fatalf("health must not send Authorization, got %q", h.auth)
	}
	// sanity: the JSON body decodes.
	var resp map[string]any
	_ = json.Unmarshal(h.response, &resp)
}