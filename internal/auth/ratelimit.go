package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	loginBurst    = 5
	limiterMaxIPs = 1024
)

var loginRate = rate.Every(3 * time.Minute) // 5 attempts per 15 minutes

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type LoginLimiter struct {
	mu   sync.Mutex
	ips  map[string]*ipEntry
}

func NewLoginLimiter() *LoginLimiter {
	l := &LoginLimiter{ips: make(map[string]*ipEntry)}
	go l.cleanup()
	return l
}

func (l *LoginLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.ips[ip]
	if !ok {
		if len(l.ips) >= limiterMaxIPs {
			l.evictOldestLocked()
		}
		entry = &ipEntry{limiter: rate.NewLimiter(loginRate, loginBurst)}
		l.ips[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (l *LoginLimiter) evictOldestLocked() {
	var oldest string
	var oldestTime time.Time
	for ip, e := range l.ips {
		if oldest == "" || e.lastSeen.Before(oldestTime) {
			oldest = ip
			oldestTime = e.lastSeen
		}
	}
	if oldest != "" {
		delete(l.ips, oldest)
	}
}

func (l *LoginLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		cutoff := time.Now().Add(-30 * time.Minute)
		l.mu.Lock()
		for ip, e := range l.ips {
			if e.lastSeen.Before(cutoff) {
				delete(l.ips, ip)
			}
		}
		l.mu.Unlock()
	}
}

func RealIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	if rip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return rip
	}
	return r.RemoteAddr
}
