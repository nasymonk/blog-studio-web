package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore("test-secret-key-at-least-32-chars!!", 1*time.Hour, "/studio")
}

func TestSessionCreateAndVerify(t *testing.T) {
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")
	s := newTestStore(t)
	w := httptest.NewRecorder()
	session := s.Create(w)

	if session.ID == "" {
		t.Error("session ID should not be empty")
	}
	if session.CSRF == "" {
		t.Error("CSRF token should not be empty")
	}

	// Build a request with the cookie set by Create.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	got, ok := s.FromRequest(req)
	if !ok {
		t.Fatal("FromRequest should return session after Create")
	}
	if got.ID != session.ID {
		t.Errorf("session ID mismatch: got %q want %q", got.ID, session.ID)
	}
}

func TestSessionExpiry(t *testing.T) {
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")
	s := NewStore("test-secret-key-at-least-32-chars!!", 1*time.Millisecond, "/studio")
	w := httptest.NewRecorder()
	_ = s.Create(w)
	time.Sleep(5 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	if _, ok := s.FromRequest(req); ok {
		t.Error("expired session should not be valid")
	}
}

func TestSessionDestroy(t *testing.T) {
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")
	s := newTestStore(t)
	w := httptest.NewRecorder()
	_ = s.Create(w)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	wDel := httptest.NewRecorder()
	s.Destroy(wDel, req)

	// Now using the original cookie should fail.
	if _, ok := s.FromRequest(req); ok {
		t.Error("session should be invalid after Destroy")
	}
}

func TestSessionInvalidSignature(t *testing.T) {
	s := newTestStore(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "blog_studio_session", Value: "fakeid.invalidsig"})
	if _, ok := s.FromRequest(req); ok {
		t.Error("tampered cookie should be rejected")
	}
}

func TestSessionPersistence(t *testing.T) {
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")
	dir := t.TempDir()
	persistPath := filepath.Join(dir, "sessions.json")

	// Create session and persist it.
	s1 := NewStoreWithPersist("test-secret-key-at-least-32-chars!!", 1*time.Hour, "/studio", persistPath)
	w := httptest.NewRecorder()
	session := s1.Create(w)
	// Flush synchronously.
	s1.snapshot()

	if _, err := os.Stat(persistPath); err != nil {
		t.Fatalf("sessions.json not created: %v", err)
	}

	// Load into a new store (simulates restart).
	s2 := NewStoreWithPersist("test-secret-key-at-least-32-chars!!", 1*time.Hour, "/studio", persistPath)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	got, ok := s2.FromRequest(req)
	if !ok {
		t.Fatal("session should survive restart via persistence")
	}
	if got.ID != session.ID {
		t.Errorf("loaded session ID %q != original %q", got.ID, session.ID)
	}
}

func TestCSRFMismatch(t *testing.T) {
	t.Setenv("BLOG_STUDIO_COOKIE_INSECURE", "1")
	s := newTestStore(t)
	w := httptest.NewRecorder()
	_ = s.Create(w)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	req.Header.Set("X-CSRF-Token", "wrong-token")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	rr := httptest.NewRecorder()
	s.RequireCSRF(next).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rr.Code)
	}
	if called {
		t.Error("next handler should not be called on CSRF mismatch")
	}
}
