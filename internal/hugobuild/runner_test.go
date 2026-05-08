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

func TestRunner_Serialization(t *testing.T) {
	// Each fake hugo sleeps 50ms. If serialized, 2 runs take ≥100ms total.
	// If they ran concurrently, total would be ~50ms.
	injectFakeHugo(t, "sleep 0.05")
	r := NewRunner(nil)
	workDir := t.TempDir()

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Run(context.Background(), workDir)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	if elapsed < 90*time.Millisecond {
		t.Errorf("runs appear to have overlapped (elapsed %v < 90ms, expected serialized ≥100ms)", elapsed)
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
