package hugobuild

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"

	"blog-studio-web/internal/metrics"
)

type CommandResult struct {
	Success    bool   `json:"success"`
	ExitCode   int    `json:"exitCode"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

type Runner struct {
	mu     sync.Mutex
	logger *slog.Logger
}

func NewRunner(logger *slog.Logger) *Runner {
	return &Runner{logger: logger}
}

func (r *Runner) RunWithLabel(ctx context.Context, target, workDir string, args ...string) CommandResult {
	result := r.Run(ctx, workDir, args...)
	metrics.HugoDuration.WithLabelValues(target, fmt.Sprintf("%t", result.Success)).Observe(
		func() float64 {
			start, _ := time.Parse(time.RFC3339Nano, result.StartedAt)
			end, _ := time.Parse(time.RFC3339Nano, result.FinishedAt)
			return end.Sub(start).Seconds()
		}(),
	)
	return result
}

func (r *Runner) Run(ctx context.Context, workDir string, args ...string) CommandResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := time.Now()
	result := CommandResult{StartedAt: start.Format(time.RFC3339Nano), ExitCode: -1}
	if r.logger != nil {
		r.logger.Info("hugo build starting", "workDir", workDir, "args", args)
	}
	cmd := exec.CommandContext(ctx, "hugo", args...)
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.FinishedAt = time.Now().Format(time.RFC3339Nano)
	if err == nil {
		result.Success = true
		result.ExitCode = 0
		if r.logger != nil {
			r.logger.Info("hugo build ok", "duration_ms", time.Since(start).Milliseconds())
		}
		return result
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	if r.logger != nil {
		r.logger.Warn("hugo build failed", "exit_code", result.ExitCode, "stderr", result.Stderr, "duration_ms", time.Since(start).Milliseconds())
	}
	return result
}
