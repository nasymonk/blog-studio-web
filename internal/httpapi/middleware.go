package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/logging"
	"blog-studio-web/internal/metrics"
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
