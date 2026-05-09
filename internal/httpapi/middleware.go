package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/logging"
	"blog-studio-web/internal/metrics"

	"golang.org/x/time/rate"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	n, err := sw.ResponseWriter.Write(b)
	sw.bytes += n
	return n, err
}

func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"panic", fmt.Sprintf("%v", rec),
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
					)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(APIResponse{OK: false, Error: apperror.New("INTERNAL", "服务器内部错误。", fmt.Sprintf("%v", rec), "请稍后重试。", true)})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = auth.RandHex(8)
		}
		w.Header().Set("X-Request-ID", id)
		logger := logging.FromContext(r.Context()).With("request_id", id)
		r = r.WithContext(logging.WithLogger(r.Context(), logger))
		next.ServeHTTP(w, r)
	})
}

// normalizedPaths maps exact API paths to stable label strings to avoid metric cardinality explosion.
var normalizedPaths = map[string]string{
	"/posts":                   "/posts",
	"/posts/":                  "/posts",
	"/config":                  "/config",
	"/health":                  "/health",
	"/health/full":             "/health/full",
	"/auth/login":              "/auth/login",
	"/auth/logout":             "/auth/logout",
	"/auth/session":            "/auth/session",
	"/auth/password":           "/auth/password",
	"/site":                    "/site",
	"/now":                     "/now",
	"/audit":                   "/audit",
	"/wechat/draft":            "/wechat/draft",
	"/metrics":                 "/metrics",
}

func normalizePath(path string) string {
	// Strip /studio/api prefix if present.
	p := path
	for _, prefix := range []string{"/studio/api", "/api"} {
		if len(p) > len(prefix) && p[:len(prefix)] == prefix {
			p = p[len(prefix):]
			break
		}
	}
	if label, ok := normalizedPaths[p]; ok {
		return label
	}
	// Collapse slug-like segments: /posts/:slug/*, /trash/:id/*
	for _, pattern := range []struct{ prefix, label string }{
		{"/posts/", "/posts/:slug"},
		{"/trash/", "/trash/:id"},
		{"/preview/", "/preview/:id"},
	} {
		if len(p) > len(pattern.prefix) && p[:len(pattern.prefix)] == pattern.prefix {
			return pattern.label
		}
	}
	return p
}

func accessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(sw, r)
			elapsed := time.Since(start)
			remote := r.RemoteAddr
			if host, _, err := net.SplitHostPort(remote); err == nil {
				remote = host
			}
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				remote = fwd
			}
			label := normalizePath(r.URL.Path)
			status := strconv.Itoa(sw.status)
			metrics.HTTPRequests.WithLabelValues(r.Method, label, status).Inc()
			metrics.HTTPDuration.WithLabelValues(r.Method, label).Observe(elapsed.Seconds())
			logging.FromContext(r.Context()).Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"bytes", sw.bytes,
				"latency_ms", elapsed.Milliseconds(),
				"remote", remote,
			)
			_ = logger
		})
	}
}

const maxRequestBodyBytes = 10 << 20 // 10 MB

// withMaxBytes wraps request bodies with http.MaxBytesReader for write methods
// (POST, PUT, DELETE) to prevent oversized payloads.
func withMaxBytes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// writeRateEntry holds a per-IP rate limiter for write endpoints.
type writeRateEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// writeRateLimiter provides per-IP rate limiting for write API endpoints.
type writeRateLimiter struct {
	mu   sync.Mutex
	ips  map[string]*writeRateEntry
	rate rate.Limit
	burst int
}

func newWriteRateLimiter(rps float64, burst int) *writeRateLimiter {
	rl := &writeRateLimiter{
		ips:   make(map[string]*writeRateEntry),
		rate:  rate.Limit(rps),
		burst: burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *writeRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, ok := rl.ips[ip]
	if !ok {
		if len(rl.ips) >= 1024 {
			rl.evictOldestLocked()
		}
		entry = &writeRateEntry{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.ips[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (rl *writeRateLimiter) evictOldestLocked() {
	var oldest string
	var oldestTime time.Time
	for ip, e := range rl.ips {
		if oldest == "" || e.lastSeen.Before(oldestTime) {
			oldest = ip
			oldestTime = e.lastSeen
		}
	}
	if oldest != "" {
		delete(rl.ips, oldest)
	}
}

func (rl *writeRateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		cutoff := time.Now().Add(-30 * time.Minute)
		rl.mu.Lock()
		for ip, e := range rl.ips {
			if e.lastSeen.Before(cutoff) {
				delete(rl.ips, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// withWriteRateLimit applies per-IP rate limiting to write methods (POST, PUT, DELETE).
func (s *Server) withWriteRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			ip := auth.RealIP(r)
			if !s.writeLimiter.Allow(ip) {
				writeError(w, http.StatusTooManyRequests, apperror.New("RATE_LIMITED", "请求过于频繁，请稍后重试。", "rate limit exceeded for "+ip, "请稍后重试。", true))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to compress writes with gzip.
type gzipResponseWriter struct {
	http.ResponseWriter
	gz       *gzip.Writer
	wroteHdr bool
}

func (gw *gzipResponseWriter) WriteHeader(code int) {
	if !gw.wroteHdr {
		gw.wroteHdr = true
		// Remove Content-Length since compressed size differs.
		gw.ResponseWriter.Header().Del("Content-Length")
		gw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		gw.ResponseWriter.Header().Del("Vary")
		gw.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
	}
	gw.ResponseWriter.WriteHeader(code)
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gw.wroteHdr {
		gw.WriteHeader(http.StatusOK)
	}
	return gw.gz.Write(b)
}

func (gw *gzipResponseWriter) Flush() {
	if f, ok := gw.ResponseWriter.(http.Flusher); ok {
		_ = gw.gz.Flush()
		f.Flush()
	}
}

func (gw *gzipResponseWriter) Close() error {
	return gw.gz.Close()
}

// withGzip compresses JSON API responses when the client accepts gzip encoding.
func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// Only compress JSON responses. Content-Type is set after the handler runs,
		// so we check the path prefix instead: API routes always return JSON.
		isAPI := strings.HasPrefix(r.URL.Path, "/studio/api") || strings.HasPrefix(r.URL.Path, "/api")
		if !isAPI {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipResponseWriter{ResponseWriter: w, gz: gz}
		defer func() {
			_ = gw.Close()
		}()
		next.ServeHTTP(gw, r)
	})
}
