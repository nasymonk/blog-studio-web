package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	HugoDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hugo_build_duration_seconds",
		Help:    "Hugo build duration.",
		Buckets: []float64{1, 2, 5, 10, 30, 60},
	}, []string{"target", "success"})

	LoginAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "login_attempts_total",
		Help: "Login attempt results.",
	}, []string{"result"})

	PreviewActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "preview_active",
		Help: "Number of active preview directories.",
	})
)

func init() {
	prometheus.MustRegister(HTTPRequests, HTTPDuration, HugoDuration, LoginAttempts, PreviewActive)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
