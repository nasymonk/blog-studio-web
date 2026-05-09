package publish

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/backup"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/hugobuild"
)

// setupService creates a Service with temp directories for testing.
func setupService(t *testing.T) (*Service, config.Paths, config.Config) {
	t.Helper()
	root := t.TempDir()
	dataDir := t.TempDir()

	postRoot := filepath.Join(root, "content", "post")
	if err := os.MkdirAll(postRoot, 0755); err != nil {
		t.Fatal(err)
	}

	paths := config.Paths{
		BlogRoot: root,
		DataRoot: dataDir,
		Cache:    filepath.Join(dataDir, "cache"),
		Backups:  filepath.Join(dataDir, "backups"),
		Logs:     filepath.Join(dataDir, "logs"),
		Diffs:    filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{
		Site: config.SiteConfig{
			ID:           "test",
			BlogRoot:     root,
			PostRoot:     postRoot,
			BuildCommand: "hugo --minify",
		},
	}

	backupStore := backup.NewStore(paths, 5)
	auditLogger := audit.NewLogger(paths)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	runner := hugobuild.NewRunner(logger)

	svc := NewService(paths, cfg, backupStore, auditLogger, runner)
	return svc, paths, cfg
}

// initHugoSite runs "hugo new site" in the given directory to create a minimal Hugo project.
func initHugoSite(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}
	cmd := exec.Command("hugo", "new", "site", dir, "--format", "toml")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hugo new site failed: %v\n%s", err, out)
	}
}

func TestSaveDraftCreatesCacheFile(t *testing.T) {
	svc, paths, _ := setupService(t)

	draft := content.PostDraft{
		Slug: "test-post",
		FrontMatter: content.FrontMatter{
			Title: "Test Post",
			Date:  "2026-05-10",
			Draft: false,
		},
		Body: "Hello world.",
	}

	if err := svc.SaveDraft(draft); err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}

	cached := filepath.Join(paths.Cache, "posts", "test-post", "index.md")
	raw, err := os.ReadFile(cached)
	if err != nil {
		t.Fatalf("cache file not created: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("cache file is empty")
	}
	// Verify content contains the title
	text := string(raw)
	if !contains(text, "title: Test Post") {
		t.Fatalf("cache file missing title, got: %s", text)
	}
}

func TestSaveDraftInvalidSlug(t *testing.T) {
	svc, _, _ := setupService(t)

	draft := content.PostDraft{
		Slug:        "../escape",
		FrontMatter: content.FrontMatter{Title: "Bad"},
		Body:        "body",
	}

	if err := svc.SaveDraft(draft); err == nil {
		t.Fatal("expected error for invalid slug")
	}
}

func TestLoadPostFromContentDir(t *testing.T) {
	svc, _, cfg := setupService(t)

	// Create a post file in the content directory
	postDir := filepath.Join(cfg.Site.PostRoot, "my-post")
	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatal(err)
	}
	md := "---\ntitle: My Post\ndate: \"2026-05-10\"\ndraft: false\ntags:\n- go\n---\n\nBody here.\n"
	if err := os.WriteFile(filepath.Join(postDir, "index.md"), []byte(md), 0644); err != nil {
		t.Fatal(err)
	}

	draft, err := svc.LoadPost("my-post")
	if err != nil {
		t.Fatalf("LoadPost failed: %v", err)
	}
	if draft.Slug != "my-post" {
		t.Fatalf("expected slug 'my-post', got %q", draft.Slug)
	}
	if draft.FrontMatter.Title != "My Post" {
		t.Fatalf("expected title 'My Post', got %q", draft.FrontMatter.Title)
	}
	if draft.Body != "Body here.\n" {
		t.Fatalf("expected body 'Body here.\\n', got %q", draft.Body)
	}
}

func TestLoadPostNotFound(t *testing.T) {
	svc, _, _ := setupService(t)

	_, err := svc.LoadPost("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent post")
	}
}

func TestLoadPostInvalidSlug(t *testing.T) {
	svc, _, _ := setupService(t)

	_, err := svc.LoadPost("../escape")
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
}

func TestPublishBlogWritesToContentDir(t *testing.T) {
	root := t.TempDir()
	dataDir := t.TempDir()

	// Initialize a Hugo site so the build step succeeds
	initHugoSite(t, root)

	postRoot := filepath.Join(root, "content", "post")
	if err := os.MkdirAll(postRoot, 0755); err != nil {
		t.Fatal(err)
	}

	paths := config.Paths{
		BlogRoot: root,
		DataRoot: dataDir,
		Cache:    filepath.Join(dataDir, "cache"),
		Backups:  filepath.Join(dataDir, "backups"),
		Logs:     filepath.Join(dataDir, "logs"),
		Diffs:    filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{
		Site: config.SiteConfig{
			ID:           "test",
			BlogRoot:     root,
			PostRoot:     postRoot,
			BuildCommand: "hugo --minify",
		},
	}

	backupStore := backup.NewStore(paths, 5)
	auditLogger := audit.NewLogger(paths)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	runner := hugobuild.NewRunner(logger)
	svc := NewService(paths, cfg, backupStore, auditLogger, runner)

	draft := content.PostDraft{
		Slug: "pub-test",
		FrontMatter: content.FrontMatter{
			Title: "Publish Test",
			Date:  "2026-05-10",
			Draft: false,
		},
		Body: "Published content.",
	}

	req := BlogPublishRequest{
		Slug:             "pub-test",
		Draft:            draft,
		ConfirmOverwrite: false,
	}

	result := svc.PublishBlog(context.Background(), req)

	// The file should be written to content directory
	postFile := filepath.Join(postRoot, "pub-test", "index.md")
	raw, err := os.ReadFile(postFile)
	if err != nil {
		t.Fatalf("post file not written to content dir: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("post file is empty")
	}
	text := string(raw)
	if !contains(text, "title: Publish Test") {
		t.Fatalf("post file missing title, got: %s", text)
	}

	// Verify audit ID is set
	if result.AuditID == "" {
		t.Fatal("expected non-empty audit ID")
	}

	// Hugo should succeed with a valid site
	if result.Status != "success" {
		t.Fatalf("expected status 'success', got %q (error: %v)", result.Status, result.Error)
	}
}

func TestPublishBlogReturnsOKWhenHugoAvailable(t *testing.T) {
	// Only run this test if Hugo is actually installed
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH, skipping full publish test")
	}

	root := t.TempDir()
	initHugoSite(t, root)

	// Hugo new site creates content/ but not content/post/
	postRoot := filepath.Join(root, "content", "post")
	if err := os.MkdirAll(postRoot, 0755); err != nil {
		t.Fatal(err)
	}

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root,
		DataRoot: dataDir,
		Cache:    filepath.Join(dataDir, "cache"),
		Backups:  filepath.Join(dataDir, "backups"),
		Logs:     filepath.Join(dataDir, "logs"),
		Diffs:    filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{
		Site: config.SiteConfig{
			ID:           "test",
			BlogRoot:     root,
			PostRoot:     postRoot,
			BuildCommand: "hugo --minify",
		},
	}

	backupStore := backup.NewStore(paths, 5)
	auditLogger := audit.NewLogger(paths)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	runner := hugobuild.NewRunner(logger)
	svc := NewService(paths, cfg, backupStore, auditLogger, runner)

	draft := content.PostDraft{
		Slug: "ok-test",
		FrontMatter: content.FrontMatter{
			Title: "OK Test",
			Date:  "2026-05-10",
			Draft: false,
		},
		Body: "Full publish.",
	}

	req := BlogPublishRequest{
		Slug:  "ok-test",
		Draft: draft,
	}

	result := svc.PublishBlog(context.Background(), req)
	if result.Status != "success" {
		t.Fatalf("expected status 'success', got %q (error: %v)", result.Status, result.Error)
	}
	if result.UploadedFiles == nil || len(result.UploadedFiles) == 0 {
		t.Fatal("expected uploaded files list")
	}
}

func TestSaveDraftCreatesMetadata(t *testing.T) {
	svc, paths, _ := setupService(t)

	draft := content.PostDraft{
		Slug: "meta-test",
		FrontMatter: content.FrontMatter{
			Title: "Meta Test",
			Date:  "2026-05-10",
			Draft: true,
			Tags:  []string{"go"},
		},
		Body: "Content.",
	}

	if err := svc.SaveDraft(draft); err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}

	// Verify metadata.json was created
	metaPath := filepath.Join(paths.Cache, "metadata.json")
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("metadata.json not created: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("metadata.json is empty")
	}
	text := string(raw)
	if !contains(text, "meta-test") {
		t.Fatalf("metadata.json missing slug, got: %s", text)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
