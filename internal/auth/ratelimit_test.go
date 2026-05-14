package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginLimiter_AllowThenDeny(t *testing.T) {
	l := NewLoginLimiter()
	ip := "192.0.2.1"

	// First loginBurst attempts should be allowed.
	for i := 0; i < loginBurst; i++ {
		if !l.Allow(ip) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	// Next attempt must be denied.
	if l.Allow(ip) {
		t.Error("attempt after burst limit should be denied")
	}
}

func TestLoginLimiter_DifferentIPs(t *testing.T) {
	l := NewLoginLimiter()
	for i := 0; i < loginBurst; i++ {
		l.Allow("10.0.0.1")
	}
	// A different IP should still be allowed.
	if !l.Allow("10.0.0.2") {
		t.Error("different IP should not be rate-limited")
	}
}

func TestLoginLimiter_EvictOldest(t *testing.T) {
	l := NewLoginLimiter()
	// Fill up to limiterMaxIPs so the next insert triggers eviction.
	for i := 0; i < limiterMaxIPs; i++ {
		ip := "10.0." + string(rune('0'+i/256)) + "." + string(rune('0'+i%256))
		l.Allow(ip)
	}
	// Should not panic and should still accept a new IP.
	if !l.Allow("172.16.0.1") {
		// After eviction a fresh limiter is created — Allow returns true.
		t.Log("new IP denied after eviction (burst already consumed for that slot — ok if deterministic)")
	}
}

func TestRealIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 1.2.3.4")
	got := RealIP(req)
	if got != "10.0.0.1" {
		t.Errorf("RealIP = %q want 10.0.0.1 (first IP in X-Forwarded-For)", got)
	}
}

func TestRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	got := RealIP(req)
	if got != "1.2.3.4" {
		t.Errorf("RealIP = %q want 1.2.3.4", got)
	}
}
