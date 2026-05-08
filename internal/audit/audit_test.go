package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blog-studio-web/internal/config"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	dir := t.TempDir()
	paths := config.Paths{
		Logs:  filepath.Join(dir, "logs"),
		Diffs: filepath.Join(dir, "diffs"),
	}
	return NewLogger(paths)
}

func TestAppendAndRecent(t *testing.T) {
	l := newTestLogger(t)

	entries := []Entry{
		{Operation: "publish", Slug: "a", Result: "ok"},
		{Operation: "rollback", Slug: "b", Result: "ok"},
		{Operation: "publish", Slug: "c", Result: "error"},
	}
	for _, e := range entries {
		if err := l.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	recent, err := l.Recent(10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("got %d entries want 3", len(recent))
	}
	// Recent returns newest-first.
	if recent[0].Slug != "c" {
		t.Errorf("first entry slug = %q want c", recent[0].Slug)
	}
}

func TestRecentEmpty(t *testing.T) {
	l := newTestLogger(t)
	entries, err := l.Recent(10)
	if err != nil {
		t.Fatalf("Recent on empty: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("want empty slice, got %d entries", len(entries))
	}
}

func TestRecentLimit(t *testing.T) {
	l := newTestLogger(t)
	for i := 0; i < 10; i++ {
		_ = l.Append(Entry{Operation: "publish", Slug: "x"})
	}
	recent, _ := l.Recent(3)
	if len(recent) != 3 {
		t.Errorf("got %d want 3", len(recent))
	}
}

func TestRotate(t *testing.T) {
	l := newTestLogger(t)
	// Write 120 entries.
	for i := 0; i < 120; i++ {
		_ = l.Append(Entry{Operation: "publish", Slug: "x"})
	}
	if err := l.Rotate(100); err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(l.paths.Logs, "audit.log"))
	lines := splitLines(string(data))
	if len(lines) != 100 {
		t.Errorf("after rotate got %d lines want 100", len(lines))
	}
}

func TestRotateNoOpWhenUnderLimit(t *testing.T) {
	l := newTestLogger(t)
	for i := 0; i < 50; i++ {
		_ = l.Append(Entry{Operation: "publish", Slug: "x"})
	}
	_ = l.Rotate(100)
	data, _ := os.ReadFile(filepath.Join(l.paths.Logs, "audit.log"))
	lines := splitLines(string(data))
	if len(lines) != 50 {
		t.Errorf("got %d lines want 50", len(lines))
	}
}

func TestPruneDiffs(t *testing.T) {
	l := newTestLogger(t)

	// Write one entry with a diff path.
	auditID := "audit-keep"
	diffPath, err := l.WriteDiff(auditID, "--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new")
	if err != nil {
		t.Fatalf("WriteDiff: %v", err)
	}
	_ = l.Append(Entry{Operation: "publish", Slug: "a", DiffPath: diffPath})

	// Write an orphan diff file not referenced by any audit entry.
	orphanPath := filepath.Join(l.paths.Diffs, "orphan.diff")
	_ = os.WriteFile(orphanPath, []byte("orphan"), 0600)

	if err := l.PruneDiffs(); err != nil {
		t.Fatalf("PruneDiffs: %v", err)
	}

	// Orphan should be gone, referenced diff should remain.
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Error("orphan diff should have been deleted")
	}
	if _, err := os.Stat(diffPath); err != nil {
		t.Errorf("referenced diff should remain: %v", err)
	}
}

func TestSplitLinesEdgeCases(t *testing.T) {
	tests := []struct {
		input string
		count int
	}{
		{"", 0},
		{"a", 1},
		{"a\n", 1},
		{"a\nb", 2},
		{"a\nb\n", 2},
	}
	for _, tt := range tests {
		lines := splitLines(tt.input)
		if len(lines) != tt.count {
			t.Errorf("splitLines(%q) = %d lines want %d", tt.input, len(lines), tt.count)
		}
	}
}

func TestAppendSecretsRedacted(t *testing.T) {
	l := newTestLogger(t)
	secret := "super-secret-token"
	_ = l.Append(Entry{Operation: "publish", Slug: "x", ErrorBrief: secret}, secret)

	data, _ := os.ReadFile(filepath.Join(l.paths.Logs, "audit.log"))
	if strings.Contains(string(data), secret) {
		t.Error("secret should have been redacted in audit.log")
	}
	var e Entry
	lines := splitLines(string(data))
	if len(lines) == 0 {
		t.Fatal("no lines written")
	}
	_ = json.Unmarshal([]byte(lines[0]), &e)
	if e.ErrorBrief != "[REDACTED]" {
		t.Errorf("ErrorBrief = %q want [REDACTED]", e.ErrorBrief)
	}
}
