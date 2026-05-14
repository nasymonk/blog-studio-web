package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestNew_Production_ReturnsJSONHandler(t *testing.T) {
	logger := New("production")
	if logger == nil {
		t.Fatal("New(\"production\") returned nil")
	}
	_, ok := logger.Handler().(*slog.JSONHandler)
	if !ok {
		t.Errorf("expected *slog.JSONHandler, got %T", logger.Handler())
	}
}

func TestNew_Development_ReturnsTextHandler(t *testing.T) {
	logger := New("development")
	if logger == nil {
		t.Fatal("New(\"development\") returned nil")
	}
	_, ok := logger.Handler().(*slog.TextHandler)
	if !ok {
		t.Errorf("expected *slog.TextHandler, got %T", logger.Handler())
	}
}

func TestNew_EmptyEnv_ReturnsTextHandler(t *testing.T) {
	logger := New("")
	if logger == nil {
		t.Fatal("New(\"\") returned nil")
	}
	_, ok := logger.Handler().(*slog.TextHandler)
	if !ok {
		t.Errorf("expected *slog.TextHandler for empty env, got %T", logger.Handler())
	}
}

func TestNew_AnyNonProductionEnv_ReturnsTextHandler(t *testing.T) {
	for _, env := range []string{"staging", "test", "local", "dev"} {
		logger := New(env)
		_, ok := logger.Handler().(*slog.TextHandler)
		if !ok {
			t.Errorf("env %q: expected *slog.TextHandler, got %T", env, logger.Handler())
		}
	}
}

func TestWithLogger_StoresLoggerInContext(t *testing.T) {
	logger := slog.Default()
	ctx := WithLogger(context.Background(), logger)

	val := ctx.Value(keyLogger)
	if val == nil {
		t.Fatal("logger not stored in context")
	}
	if val != logger {
		t.Errorf("stored logger does not match input")
	}
}

func TestFromContext_ReturnsStoredLogger(t *testing.T) {
	logger := slog.Default()
	ctx := WithLogger(context.Background(), logger)

	got := FromContext(ctx)
	if got != logger {
		t.Errorf("FromContext = %v, want %v", got, logger)
	}
}

func TestFromContext_NoLoggerInContext_ReturnsDefault(t *testing.T) {
	ctx := context.Background()
	got := FromContext(ctx)

	if got != slog.Default() {
		t.Errorf("FromContext without logger = %v, want %v", got, slog.Default())
	}
}

func TestFromContext_NilLoggerStored_ReturnsDefault(t *testing.T) {
	ctx := WithLogger(context.Background(), nil)
	got := FromContext(ctx)

	if got != slog.Default() {
		t.Errorf("FromContext with nil stored = %v, want %v", got, slog.Default())
	}
}

func TestFromContext_DifferentContextsAreIndependent(t *testing.T) {
	l1 := slog.Default()
	l2 := slog.New(slog.NewTextHandler(nil, nil))

	ctx1 := WithLogger(context.Background(), l1)
	ctx2 := WithLogger(context.Background(), l2)

	if FromContext(ctx1) != l1 {
		t.Error("ctx1 should return l1")
	}
	if FromContext(ctx2) != l2 {
		t.Error("ctx2 should return l2")
	}
}

func TestFromContext_LoggerLoggedToStderr(t *testing.T) {
	// Verify returned logger is functional by logging at Info level (no crash).
	logger := FromContext(context.Background())
	logger.Info("test log message")
}
