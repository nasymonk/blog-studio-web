package logging

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey int

const keyLogger ctxKey = 0

func New(env string) *slog.Logger {
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, keyLogger, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(keyLogger).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
