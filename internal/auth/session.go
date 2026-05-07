package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	CSRF      string
	ExpiresAt time.Time
}

type Store struct {
	mu       sync.Mutex
	secret   []byte
	sessions map[string]Session
	ttl      time.Duration
	basePath string
}

func NewStore(secret string, ttl time.Duration, basePath string) *Store {
	return &Store{secret: []byte(secret), sessions: map[string]Session{}, ttl: ttl, basePath: basePath}
}

func (s *Store) Create(w http.ResponseWriter) Session {
	session := Session{ID: uuid.NewString(), CSRF: randomToken(32), ExpiresAt: time.Now().Add(s.ttl)}
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "blog_studio_session",
		Value:    s.sign(session.ID),
		Path:     cookiePath(s.basePath),
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return session
}

func (s *Store) Destroy(w http.ResponseWriter, r *http.Request) {
	if session, ok := s.FromRequest(r); ok {
		s.mu.Lock()
		delete(s.sessions, session.ID)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "blog_studio_session", Value: "", Path: cookiePath(s.basePath), Expires: time.Unix(0, 0), HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (s *Store) FromRequest(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie("blog_studio_session")
	if err != nil {
		return Session{}, false
	}
	id, ok := s.verify(cookie.Value)
	if !ok {
		return Session{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok || time.Now().After(session.ExpiresAt) {
		delete(s.sessions, id)
		return Session{}, false
	}
	return session, true
}

func (s *Store) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.FromRequest(r); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Store) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		session, ok := s.FromRequest(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-CSRF-Token") != session.CSRF {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Store) sign(id string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(id))
	return id + "." + hex.EncodeToString(mac.Sum(nil))
}

func (s *Store) verify(value string) (string, bool) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	expected := s.sign(parts[0])
	return parts[0], hmac.Equal([]byte(value), []byte(expected))
}

func randomToken(size int) string {
	data := make([]byte, size)
	_, _ = rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}

func cookiePath(basePath string) string {
	if basePath == "" {
		return "/"
	}
	if !strings.HasSuffix(basePath, "/") {
		return basePath + "/"
	}
	return basePath
}
