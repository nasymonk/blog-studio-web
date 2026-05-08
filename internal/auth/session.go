package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	CSRF      string    `json:"csrf"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Store struct {
	mu          sync.Mutex
	secret      []byte
	sessions    map[string]Session
	ttl         time.Duration
	basePath    string
	persistPath string
	saveCh      chan struct{}
}

func NewStore(secret string, ttl time.Duration, basePath string) *Store {
	s := &Store{
		secret:   []byte(secret),
		sessions: map[string]Session{},
		ttl:      ttl,
		basePath: basePath,
		saveCh:   make(chan struct{}, 1),
	}
	return s
}

func NewStoreWithPersist(secret string, ttl time.Duration, basePath, persistPath string) *Store {
	s := &Store{
		secret:      []byte(secret),
		sessions:    map[string]Session{},
		ttl:         ttl,
		basePath:    basePath,
		persistPath: persistPath,
		saveCh:      make(chan struct{}, 1),
	}
	s.load()
	go s.reaper()
	go s.saver()
	return s
}

func (s *Store) Close(ctx context.Context) {
	s.snapshot()
}

func (s *Store) Create(w http.ResponseWriter) Session {
	session := Session{ID: uuid.NewString(), CSRF: randomToken(32), ExpiresAt: time.Now().Add(s.ttl)}
	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()
	secure := os.Getenv("BLOG_STUDIO_COOKIE_INSECURE") != "1"
	http.SetCookie(w, &http.Cookie{
		Name:     "blog_studio_session",
		Value:    s.sign(session.ID),
		Path:     cookiePath(s.basePath),
		MaxAge:   int(s.ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
	s.triggerSave()
	return session
}

func (s *Store) Destroy(w http.ResponseWriter, r *http.Request) {
	if session, ok := s.FromRequest(r); ok {
		s.mu.Lock()
		delete(s.sessions, session.ID)
		s.mu.Unlock()
	}
	secure := os.Getenv("BLOG_STUDIO_COOKIE_INSECURE") != "1"
	http.SetCookie(w, &http.Cookie{
		Name:     "blog_studio_session",
		Value:    "",
		Path:     cookiePath(s.basePath),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
	s.triggerSave()
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

func (s *Store) triggerSave() {
	select {
	case s.saveCh <- struct{}{}:
	default:
	}
}

func (s *Store) reaper() {
	for range time.Tick(5 * time.Minute) {
		now := time.Now()
		s.mu.Lock()
		for id, sess := range s.sessions {
			if now.After(sess.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
		s.triggerSave()
	}
}

func (s *Store) saver() {
	for range s.saveCh {
		time.Sleep(500 * time.Millisecond)
		s.snapshot()
	}
}

func (s *Store) snapshot() {
	if s.persistPath == "" {
		return
	}
	s.mu.Lock()
	data, err := json.Marshal(s.sessions)
	s.mu.Unlock()
	if err != nil {
		return
	}
	_ = os.MkdirAll(strings.TrimSuffix(s.persistPath, "/sessions.json"), 0700)
	tmp := s.persistPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return
	}
	_ = os.Rename(tmp, s.persistPath)
}

func (s *Store) load() {
	if s.persistPath == "" {
		return
	}
	data, err := os.ReadFile(s.persistPath)
	if err != nil {
		return
	}
	var loaded map[string]Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}
	now := time.Now()
	s.mu.Lock()
	for id, sess := range loaded {
		if now.Before(sess.ExpiresAt) {
			s.sessions[id] = sess
		}
	}
	s.mu.Unlock()
}

func randomToken(size int) string {
	data := make([]byte, size)
	_, _ = rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}

func RandHex(size int) string {
	data := make([]byte, size)
	_, _ = rand.Read(data)
	return hex.EncodeToString(data)
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
