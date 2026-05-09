package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/config"
)

// newTestServer creates a Server with temp directories and returns the handler and a
// helper to perform authenticated requests.
func newTestServer(t *testing.T) (http.Handler, *auth.Store) {
	t.Helper()

	tmpDir := t.TempDir()
	blogRoot := filepath.Join(tmpDir, "blog")
	dataRoot := filepath.Join(tmpDir, "data")
	staticDir := filepath.Join(tmpDir, "static")

	// Create required directories.
	for _, d := range []string{blogRoot, dataRoot, staticDir, filepath.Join(blogRoot, "content", "post")} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	// Create a minimal hugo.toml so site-related endpoints don't error.
	hugoToml := "baseURL = \"http://localhost\"\ntitle = \"Test\"\n[params]\n  description = \"\"\n  profileImage = \"\"\n"
	if err := os.WriteFile(filepath.Join(blogRoot, "hugo.toml"), []byte(hugoToml), 0644); err != nil {
		t.Fatalf("write hugo.toml: %v", err)
	}

	// Create a minimal static/index.html for the SPA handler.
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!DOCTYPE html>"), 0644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	paths := config.Paths{
		BlogRoot:  blogRoot,
		DataRoot:  dataRoot,
		Config:    filepath.Join(dataRoot, "config.json"),
		Cache:     filepath.Join(dataRoot, "cache"),
		Backups:   filepath.Join(dataRoot, "backups"),
		Logs:      filepath.Join(dataRoot, "logs"),
		Diffs:     filepath.Join(dataRoot, "logs", "diffs"),
		Preview:   filepath.Join(dataRoot, "preview"),
		Static:    staticDir,
		AdminHash: filepath.Join(dataRoot, "admin-password.hash"),
		Trash:     filepath.Join(dataRoot, "trash"),
	}

	cfg := config.DefaultConfig(paths)
	cfgStore := config.NewStore(paths)

	adminHash, err := auth.HashPassword("testpass123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	sessions := auth.NewStore("test-secret-for-session-store-32b!", 1*time.Hour, "/studio")

	// Disable cookie Secure flag for httptest.
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")

	logger := slog.Default()
	srv := New(paths, cfg, cfgStore, adminHash, sessions, logger)
	return srv.Handler(), sessions
}

// login performs a POST /studio/api/auth/login and returns the response recorder.
// The session cookie is set on the recorder if login succeeds.
func login(t *testing.T, handler http.Handler, password string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": password})
	req := httptest.NewRequest(http.MethodPost, "/studio/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// doRequest is a convenience for making an HTTP request against the handler.
func doRequest(t *testing.T, handler http.Handler, method, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHealthEndpoint(t *testing.T) {
	handler, _ := newTestServer(t)

	rec := doRequest(t, handler, http.MethodGet, "/studio/api/health", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got false: %s", rec.Body.String())
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", resp.Data)
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", data["status"])
	}
}

func TestLoginWrongPassword(t *testing.T) {
	handler, _ := newTestServer(t)

	rec := login(t, handler, "wrongpassword")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	if resp.Error == nil {
		t.Fatalf("expected error in response")
	}
	if resp.Error.Code != "LOGIN_FAILED" {
		t.Fatalf("expected error code LOGIN_FAILED, got %s", resp.Error.Code)
	}
}

func TestPostsRequiresAuth(t *testing.T) {
	handler, _ := newTestServer(t)

	rec := doRequest(t, handler, http.MethodGet, "/studio/api/posts", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	if resp.Error == nil {
		t.Fatalf("expected error in response")
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected error code UNAUTHORIZED, got %s", resp.Error.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler, _ := newTestServer(t)

	rec := doRequest(t, handler, http.MethodGet, "/studio/api/health", nil)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "SAMEORIGIN"},
		{"Referrer-Policy", "same-origin"},
		{"Cross-Origin-Opener-Policy", "same-origin"},
	}
	for _, tt := range tests {
		got := rec.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, got)
		}
	}

	// Content-Security-Policy should be present (non-empty).
	if csp := rec.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("expected Content-Security-Policy header to be present")
	}

	// Permissions-Policy should be present.
	if pp := rec.Header().Get("Permissions-Policy"); pp == "" {
		t.Error("expected Permissions-Policy header to be present")
	}
}

func TestLoginSuccessAndAuthenticatedRequest(t *testing.T) {
	handler, _ := newTestServer(t)

	// Login with correct password.
	rec := login(t, handler, "testpass123")
	if rec.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Extract session cookie.
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "blog_studio_session" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("expected session cookie from login response")
	}

	// Use session cookie to access a protected endpoint.
	rec2 := doRequest(t, handler, http.MethodGet, "/studio/api/posts", cookie)
	if rec2.Code != http.StatusOK {
		t.Fatalf("posts with auth: expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var resp APIResponse
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got false: %s", rec2.Body.String())
	}
}

func TestRequestIDHeader(t *testing.T) {
	handler, _ := newTestServer(t)

	rec := doRequest(t, handler, http.MethodGet, "/studio/api/health", nil)

	reqID := rec.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header to be present")
	}
}
