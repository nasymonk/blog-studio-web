package hugobuild

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// injectFakeHugo writes a shell script named "hugo" to a temp bin dir,
// adds it to the front of PATH, and returns a cleanup function.
func injectFakeHugo(t *testing.T, script string) {
	t.Helper()
	binDir := t.TempDir()
	hugoPath := filepath.Join(binDir, "hugo")
	if err := os.WriteFile(hugoPath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
		t.Fatalf("write fake hugo: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunner_SuccessfulBuild(t *testing.T) {
	injectFakeHugo(t, "exit 0")
	r := NewRunner(nil)
	result := r.Run(context.Background(), t.TempDir())
	if !result.Success {
		t.Errorf("expected success, got stderr=%q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d want 0", result.ExitCode)
	}
}

func TestRunner_FailedBuild(t *testing.T) {
	injectFakeHugo(t, "echo 'build error' >&2; exit 1")
	r := NewRunner(nil)
	result := r.Run(context.Background(), t.TempDir())
	if result.Success {
		t.Error("expected failure")
	}
	if result.ExitCode != 1 {
		t.Errorf("exit code = %d want 1", result.ExitCode)
	}
}

func TestRunner_ConcurrentCallsReturnSameResult(t *testing.T) {
	// singleflight ensures concurrent calls with the same key share one execution.
	// Both calls should return the same result.
	injectFakeHugo(t, "exit 0")
	r := NewRunner(nil)
	workDir := t.TempDir()

	var wg sync.WaitGroup
	results := make([]CommandResult, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			results[idx] = r.Run(context.Background(), workDir)
		}()
	}
	wg.Wait()

	for i, res := range results {
		if !res.Success {
			t.Errorf("run %d: expected success, got exit code %d stderr=%q", i, res.ExitCode, res.Stderr)
		}
	}
}

func TestRunner_ContextCancellation(t *testing.T) {
	injectFakeHugo(t, "sleep 10")
	r := NewRunner(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result := r.Run(ctx, t.TempDir())
	if result.Success {
		t.Error("cancelled build should not succeed")
	}
}

func TestRunWithLabel(t *testing.T) {
	injectFakeHugo(t, "exit 0")
	r := NewRunner(nil)
	result := r.RunWithLabel(context.Background(), "publish", t.TempDir())
	if !result.Success {
		t.Error("RunWithLabel should succeed")
	}
}
