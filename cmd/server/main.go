package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"log/slog"

	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/httpapi"
	"blog-studio-web/internal/logging"
	"blog-studio-web/internal/trash"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "hash-password" {
		hash, err := auth.HashPassword(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(hash)
		return
	}

	appEnv := env("APP_ENV", "development")
	logger := logging.New(appEnv)

	adminHash := os.Getenv("BLOG_STUDIO_ADMIN_PASSWORD_HASH")
	sessionSecret := os.Getenv("BLOG_STUDIO_SESSION_SECRET")
	if adminHash == "" || sessionSecret == "" || len(sessionSecret) < 32 {
		logger.Error("missing required env vars", "note", "BLOG_STUDIO_ADMIN_PASSWORD_HASH and BLOG_STUDIO_SESSION_SECRET (32+ chars) required")
		os.Exit(1)
	}

	paths := config.DefaultPaths()
	if data, err := os.ReadFile(paths.AdminHash); err == nil && strings.TrimSpace(string(data)) != "" {
		adminHash = strings.TrimSpace(string(data))
	}

	store := config.NewStore(paths)
	cfg, cfgErr := store.Load()
	if cfgErr != nil {
		logger.Error("failed to load config", "error", cfgErr.Error())
		os.Exit(1)
	}
	if err := store.Save(cfg); err != nil {
		logger.Error("failed to save config", "error", err.Error())
		os.Exit(1)
	}

	persistPath := filepath.Join(paths.DataRoot, "sessions.json")
	sessions := auth.NewStoreWithPersist(sessionSecret, 12*time.Hour, cfg.BasePath, persistPath)
	auditLogger := audit.NewLogger(paths)
	server := httpapi.NewWithAudit(paths, cfg, store, adminHash, sessions, auditLogger, logger)

	addr := ":" + env("PORT", "8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("blog-studio-web starting", "addr", addr, "base", strings.TrimRight(cfg.BasePath, "/"), "env", appEnv)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	go startBackgroundWorkers(ctx, server, auditLogger, paths, logger)

	<-ctx.Done()
	logger.Info("shutdown signal received, draining connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sessions.Close(shutdownCtx)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func startBackgroundWorkers(ctx context.Context, server *httpapi.Server, auditLog *audit.Logger, paths config.Paths, logger *slog.Logger) {
	rotateAudit := func() {
		if err := auditLog.Rotate(audit.MaxAuditLines); err != nil {
			logger.Warn("audit rotate failed", "error", err)
		}
		if err := auditLog.PruneDiffs(); err != nil {
			logger.Warn("audit prune diffs failed", "error", err)
		}
	}
	cleanPreview := func() {
		logger.Info("preview cleanup running")
		if prev := server.PreviewService(); prev != nil {
			if err := prev.Cleanup(); err != nil {
				logger.Warn("preview cleanup failed", "error", err)
			}
		}
	}
	pruneTrash := func() {
		if err := trash.NewStore(paths).PruneOlderThan(30 * 24 * time.Hour); err != nil {
			logger.Warn("trash prune failed", "error", err)
		}
	}
	rotateAudit()
	cleanPreview()
	pruneTrash()
	dailyTick := time.NewTicker(24 * time.Hour)
	previewTick := time.NewTicker(10 * time.Minute)
	defer dailyTick.Stop()
	defer previewTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-dailyTick.C:
			rotateAudit()
			pruneTrash()
		case <-previewTick.C:
			cleanPreview()
		}
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
