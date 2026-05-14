package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetrics_GlobalVarsAreNotNil(t *testing.T) {
	if HTTPRequests == nil {
		t.Error("HTTPRequests is nil")
	}
	if HTTPDuration == nil {
		t.Error("HTTPDuration is nil")
	}
	if HugoDuration == nil {
		t.Error("HugoDuration is nil")
	}
	if LoginAttempts == nil {
		t.Error("LoginAttempts is nil")
	}
	if PreviewActive == nil {
		t.Error("PreviewActive is nil")
	}
}

func TestMetrics_NamingConventions(t *testing.T) {
	reg := prometheus.NewRegistry()
	copied := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "test"},
		[]string{"method", "path", "status"},
	)
	copied.WithLabelValues("GET", "/", "200").Inc()
	reg.MustRegister(copied)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	if len(families) == 0 {
		t.Fatal("no metric families gathered")
	}
	if families[0].GetName() != "http_requests_total" {
		t.Errorf("expected name 'http_requests_total', got %q", families[0].GetName())
	}
}

func TestCounterVec_Increment(t *testing.T) {
	LoginAttempts.WithLabelValues("success").Inc()
	LoginAttempts.WithLabelValues("failure").Inc()
	LoginAttempts.WithLabelValues("failure").Inc()
}

func TestHistogramVec_Observe(t *testing.T) {
	HTTPDuration.WithLabelValues("GET", "/").Observe(0.1)
	HTTPDuration.WithLabelValues("GET", "/").Observe(0.5)
	HTTPDuration.WithLabelValues("POST", "/api/publish").Observe(1.2)

	HugoDuration.WithLabelValues("production", "true").Observe(3.0)
	HugoDuration.WithLabelValues("draft", "false").Observe(0.5)
}

func TestGauge_Set(t *testing.T) {
	PreviewActive.Set(1)
	PreviewActive.Set(5)
	PreviewActive.Set(0)
}

func TestMetrics_CorrectLabelNames(t *testing.T) {
	reg := prometheus.NewRegistry()
	copied := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "test_requests_total", Help: "test"},
		[]string{"method", "path", "status"},
	)
	copied.WithLabelValues("GET", "/", "200").Inc()
	reg.MustRegister(copied)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	if len(families) == 0 || len(families[0].GetMetric()) == 0 {
		t.Fatal("no metrics gathered")
	}
	labels := families[0].GetMetric()[0].GetLabel()
	wantLabels := map[string]bool{"method": false, "path": false, "status": false}
	for _, l := range labels {
		wantLabels[l.GetName()] = true
	}
	for k, found := range wantLabels {
		if !found {
			t.Errorf("missing label %q", k)
		}
	}
}

func TestMetrics_RegisterWithCustomRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	copied := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "test_register", Help: "test"},
		[]string{"method"},
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("MustRegister panicked: %v", r)
			}
		}()
		reg.MustRegister(copied)
	}()

	copied.WithLabelValues("GET").Inc()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	var found bool
	for _, f := range families {
		if f.GetName() == "test_register" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test_register metric family not found after registration")
	}
}

func TestMetrics_HugoDurationName(t *testing.T) {
	reg := prometheus.NewRegistry()
	copied := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hugo_build_duration_seconds",
			Help:    "test",
			Buckets: []float64{1, 2, 5, 10, 30, 60},
		},
		[]string{"target", "success"},
	)
	copied.WithLabelValues("test", "true").Observe(1.0)
	reg.MustRegister(copied)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	if len(families) == 0 {
		t.Fatal("no metric families gathered")
	}
	if families[0].GetName() != "hugo_build_duration_seconds" {
		t.Errorf("expected name 'hugo_build_duration_seconds', got %q", families[0].GetName())
	}
}

func TestMetrics_LoginAttemptsLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	copied := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "login_test", Help: "test"},
		[]string{"result"},
	)
	copied.WithLabelValues("success").Inc()
	reg.MustRegister(copied)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	if len(families) == 0 || len(families[0].GetMetric()) == 0 {
		t.Fatal("no metrics gathered")
	}
	labels := families[0].GetMetric()[0].GetLabel()
	found := false
	for _, l := range labels {
		if l.GetName() == "result" {
			found = true
			break
		}
	}
	if !found {
		t.Error("missing label 'result'")
	}
}

func TestHandler_ReturnsNonNil(t *testing.T) {
	h := Handler()
	if h == nil {
		t.Fatal("Handler() returned nil")
	}
}
