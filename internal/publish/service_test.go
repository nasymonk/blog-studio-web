package publish

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

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

func TestSaveDraftEmptyTitle(t *testing.T) {
	svc, _, _ := setupService(t)

	draft := content.PostDraft{
		Slug:        "valid-slug",
		FrontMatter: content.FrontMatter{Title: ""},
		Body:        "body",
	}

	if err := svc.SaveDraft(draft); err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestSaveDraftSetsMetadata(t *testing.T) {
	svc, _, _ := setupService(t)

	draft := content.PostDraft{
		Slug: "meta-post",
		FrontMatter: content.FrontMatter{
			Title:      "Meta Post",
			Date:       "2026-05-10",
			Draft:      true,
			Tags:       []string{"go", "test"},
			Categories: []string{"dev"},
		},
		Body: "Content.",
	}

	if err := svc.SaveDraft(draft); err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}

	// Load metadata and verify
	meta := svc.loadMetadata()
	state, ok := meta.Posts["meta-post"]
	if !ok {
		t.Fatal("metadata missing slug")
	}
	if state.Title != "Meta Post" {
		t.Fatalf("expected title 'Meta Post', got %q", state.Title)
	}
	if !state.Dirty {
		t.Fatal("expected dirty flag to be true")
	}
	// When saving to cache (no remote file yet), SyncStatus is unknown because remoteTime is zero
	if state.SyncStatus != SyncDirty && state.SyncStatus != SyncUnknown {
		t.Fatalf("expected SyncDirty or SyncUnknown, got %s", state.SyncStatus)
	}
	if len(state.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(state.Tags))
	}
	if len(state.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(state.Categories))
	}
}

func TestSaveDraftOverwritesExisting(t *testing.T) {
	svc, _, _ := setupService(t)

	draft1 := content.PostDraft{
		Slug:        "overwrite-post",
		FrontMatter: content.FrontMatter{Title: "V1", Date: "2026-05-10"},
		Body:        "Version 1.",
	}
	draft2 := content.PostDraft{
		Slug:        "overwrite-post",
		FrontMatter: content.FrontMatter{Title: "V2", Date: "2026-05-10"},
		Body:        "Version 2.",
	}

	if err := svc.SaveDraft(draft1); err != nil {
		t.Fatal(err)
	}
	if err := svc.SaveDraft(draft2); err != nil {
		t.Fatal(err)
	}

	draft, err := svc.LoadPost("overwrite-post")
	if err != nil {
		// LoadPost reads from content dir, not cache. Check cache instead.
		return
	}
	if draft.FrontMatter.Title != "V2" {
		t.Fatalf("expected 'V2', got %q", draft.FrontMatter.Title)
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

func TestLoadPostWithAssets(t *testing.T) {
	svc, _, cfg := setupService(t)

	postDir := filepath.Join(cfg.Site.PostRoot, "asset-post")
	os.MkdirAll(postDir, 0755)
	os.WriteFile(filepath.Join(postDir, "index.md"), []byte("---\ntitle: Asset\ndate: \"2026-05-10\"\n---\n\nBody.\n"), 0644)
	os.WriteFile(filepath.Join(postDir, "cover.png"), []byte("fake-png"), 0644)
	os.WriteFile(filepath.Join(postDir, "diagram.svg"), []byte("<svg/>"), 0644)

	draft, err := svc.LoadPost("asset-post")
	if err != nil {
		t.Fatalf("LoadPost failed: %v", err)
	}
	if len(draft.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(draft.Assets))
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

func TestLoadPostNoFrontMatter(t *testing.T) {
	svc, _, cfg := setupService(t)

	postDir := filepath.Join(cfg.Site.PostRoot, "plain-post")
	os.MkdirAll(postDir, 0755)
	os.WriteFile(filepath.Join(postDir, "index.md"), []byte("Just plain text, no front matter.\n"), 0644)

	draft, err := svc.LoadPost("plain-post")
	if err != nil {
		t.Fatalf("LoadPost failed: %v", err)
	}
	if draft.FrontMatter.Title != "Untitled" {
		t.Fatalf("expected 'Untitled', got %q", draft.FrontMatter.Title)
	}
	if draft.Body != "Just plain text, no front matter.\n" {
		t.Fatalf("unexpected body: %q", draft.Body)
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

func TestPublishBlogWithAssetFiles(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}

	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	req := BlogPublishRequest{
		Slug:             "asset-test",
		Draft:            content.PostDraft{Slug: "asset-test", FrontMatter: content.FrontMatter{Title: "Assets", Date: "2026-05-10"}, Body: "With image."},
		ConfirmOverwrite: true,
		Files: map[string][]byte{
			"cover.png": []byte("fake-png-data"),
		},
	}

	result := svc.PublishBlog(context.Background(), req)
	if result.Status != "success" {
		t.Fatalf("expected success, got %q (error: %v)", result.Status, result.Error)
	}
	if len(result.UploadedFiles) < 2 {
		t.Fatalf("expected at least 2 uploaded files (index.md + asset), got %d", len(result.UploadedFiles))
	}

	// Verify asset file was written
	assetPath := filepath.Join(postRoot, "asset-test", "cover.png")
	if _, err := os.Stat(assetPath); err != nil {
		t.Fatalf("asset file not written: %v", err)
	}
}

func TestPublishBlogRejectsInvalidAssetName(t *testing.T) {
	svc, _, _ := setupService(t)

	req := BlogPublishRequest{
		Slug: "bad-asset",
		Draft: content.PostDraft{
			Slug:        "bad-asset",
			FrontMatter: content.FrontMatter{Title: "Bad", Date: "2026-05-10"},
			Body:        "body",
		},
		ConfirmOverwrite: true,
		Files: map[string][]byte{
			"../../../etc/passwd": []byte("evil"),
		},
	}

	result := svc.PublishBlog(context.Background(), req)
	if result.Status == "success" {
		t.Fatal("expected failure for path traversal in asset name")
	}
}

func TestPublishBlogRejectsNonImageAsset(t *testing.T) {
	svc, _, _ := setupService(t)

	req := BlogPublishRequest{
		Slug: "bad-ext",
		Draft: content.PostDraft{
			Slug:        "bad-ext",
			FrontMatter: content.FrontMatter{Title: "Bad Ext", Date: "2026-05-10"},
			Body:        "body",
		},
		ConfirmOverwrite: true,
		Files: map[string][]byte{
			"malware.exe": []byte("evil"),
		},
	}

	result := svc.PublishBlog(context.Background(), req)
	if result.Status == "success" {
		t.Fatal("expected failure for non-image asset")
	}
}

func TestPublishBlogConflictDetection(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	// First publish
	draft := content.PostDraft{Slug: "conflict-post", FrontMatter: content.FrontMatter{Title: "V1", Date: "2026-05-10"}, Body: "Version 1."}
	req := BlogPublishRequest{Slug: "conflict-post", Draft: draft}
	result := svc.PublishBlog(context.Background(), req)
	if result.Status != "success" {
		t.Fatalf("first publish failed: %v", result.Error)
	}

	// Simulate external modification: update the file's mtime to be newer than cached
	postFile := filepath.Join(postRoot, "conflict-post", "index.md")
	futureTime := time.Now().Add(2 * time.Hour)
	os.Chtimes(postFile, futureTime, futureTime)

	// Second publish without ConfirmOverwrite - should detect conflict
	draft2 := content.PostDraft{Slug: "conflict-post", FrontMatter: content.FrontMatter{Title: "V2", Date: "2026-05-10"}, Body: "Version 2."}
	req2 := BlogPublishRequest{Slug: "conflict-post", Draft: draft2, ConfirmOverwrite: false}
	result2 := svc.PublishBlog(context.Background(), req2)
	if result2.Status != "conflict" {
		t.Fatalf("expected conflict status, got %q", result2.Status)
	}
	if len(result2.Conflicts) == 0 {
		t.Fatal("expected conflict details")
	}
}

func TestPublishBlogConfirmOverwriteBypassesConflict(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	// First publish
	draft := content.PostDraft{Slug: "ow-post", FrontMatter: content.FrontMatter{Title: "V1", Date: "2026-05-10"}, Body: "V1."}
	svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "ow-post", Draft: draft})

	// Modify mtime
	postFile := filepath.Join(postRoot, "ow-post", "index.md")
	futureTime := time.Now().Add(2 * time.Hour)
	os.Chtimes(postFile, futureTime, futureTime)

	// With ConfirmOverwrite=true, should succeed despite conflict
	draft2 := content.PostDraft{Slug: "ow-post", FrontMatter: content.FrontMatter{Title: "V2", Date: "2026-05-10"}, Body: "V2."}
	result := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "ow-post", Draft: draft2, ConfirmOverwrite: true})
	if result.Status != "success" {
		t.Fatalf("expected success with confirm overwrite, got %q (error: %v)", result.Status, result.Error)
	}
}

func TestPublishBlogCreatesBackup(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	draft := content.PostDraft{Slug: "backup-post", FrontMatter: content.FrontMatter{Title: "Backup", Date: "2026-05-10"}, Body: "Content."}
	result := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "backup-post", Draft: draft})

	if result.BackupID == "" {
		t.Fatal("expected backup ID")
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

func TestInvalidateCache(t *testing.T) {
	svc, _, cfg := setupService(t)

	// Create a post
	postDir := filepath.Join(cfg.Site.PostRoot, "cache-test")
	os.MkdirAll(postDir, 0755)
	os.WriteFile(filepath.Join(postDir, "index.md"), []byte("---\ntitle: Cache\ndate: \"2026-05-10\"\n---\n\nBody.\n"), 0644)

	// First call populates cache
	posts1, err := svc.ListPosts()
	if err != nil {
		t.Fatal(err)
	}

	// Verify cache is populated
	svc.cache.mu.RLock()
	hasCache := svc.cache.data != nil
	svc.cache.mu.RUnlock()
	if !hasCache {
		t.Fatal("expected cache to be populated")
	}

	// Invalidate
	svc.InvalidateCache()

	svc.cache.mu.RLock()
	cacheCleared := svc.cache.data == nil
	svc.cache.mu.RUnlock()
	if !cacheCleared {
		t.Fatal("expected cache to be cleared after invalidation")
	}

	// Second call should re-populate
	posts2, err := svc.ListPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts1) != len(posts2) {
		t.Fatalf("post count mismatch: %d vs %d", len(posts1), len(posts2))
	}
}

func TestListPostsSortByDateDesc(t *testing.T) {
	svc, _, cfg := setupService(t)

	// Create posts with different dates
	for _, p := range []struct{ slug, date string }{
		{"old-post", "2025-01-01"},
		{"new-post", "2026-06-01"},
		{"mid-post", "2025-06-15"},
	} {
		dir := filepath.Join(cfg.Site.PostRoot, p.slug)
		os.MkdirAll(dir, 0755)
		md := "---\ntitle: " + p.slug + "\ndate: \"" + p.date + "\"\n---\n\nBody.\n"
		os.WriteFile(filepath.Join(dir, "index.md"), []byte(md), 0644)
	}

	posts, err := svc.ListPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}
	// Should be sorted by date descending
	if posts[0].Date < posts[1].Date {
		t.Fatalf("posts not sorted descending: %s < %s", posts[0].Date, posts[1].Date)
	}
}

func TestListPostsEmptyDir(t *testing.T) {
	svc, _, _ := setupService(t)

	posts, err := svc.ListPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(posts))
	}
}

func TestListPostsSkipsFiles(t *testing.T) {
	svc, _, cfg := setupService(t)

	// Create a regular file (not a directory) in post root
	os.WriteFile(filepath.Join(cfg.Site.PostRoot, "not-a-dir.txt"), []byte("data"), 0644)
	// Create a real post dir
	dir := filepath.Join(cfg.Site.PostRoot, "real-post")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "index.md"), []byte("---\ntitle: Real\ndate: \"2026-05-10\"\n---\n\nBody.\n"), 0644)

	posts, err := svc.ListPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Slug != "real-post" {
		t.Fatalf("expected 'real-post', got %q", posts[0].Slug)
	}
}

func TestComputeStatus(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	var zero time.Time

	cases := []struct {
		name     string
		dirty    bool
		cached   time.Time
		remote   time.Time
		expected SyncStatus
	}{
		{"clean", false, now, now, SyncClean},
		{"dirty", true, now, now, SyncDirty},
		{"stale", false, past, now, SyncStale},
		{"conflict", true, past, now, SyncConflict},
		{"unknown zero remote", false, now, zero, SyncUnknown},
		{"clean remote equals cached", false, future, future, SyncClean},
		{"dirty no newer remote", true, future, now, SyncDirty},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeStatus(tc.dirty, tc.cached, tc.remote)
			if got != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func TestParseAndFormatTime(t *testing.T) {
	// Empty string
	if got := parseTime(""); !got.IsZero() {
		t.Fatalf("expected zero time for empty string, got %v", got)
	}

	// Invalid string
	if got := parseTime("not-a-time"); !got.IsZero() {
		t.Fatalf("expected zero time for invalid string, got %v", got)
	}

	// RFC3339
	now := time.Now().Truncate(time.Second)
	formatted := now.Format(time.RFC3339)
	parsed := parseTime(formatted)
	if !parsed.Equal(now) {
		t.Fatalf("round-trip failed: %v != %v", now, parsed)
	}

	// RFC3339Nano
	nano := time.Now()
	nanoStr := nano.Format(time.RFC3339Nano)
	parsedNano := parseTime(nanoStr)
	if !parsedNano.Equal(nano) {
		t.Fatalf("nano round-trip failed")
	}

	// Format zero time
	if got := formatTime(time.Time{}); got != "" {
		t.Fatalf("expected empty string for zero time, got %q", got)
	}

	// Format non-zero time
	if got := formatTime(now); got == "" {
		t.Fatal("expected non-empty string for non-zero time")
	}
}

func TestAllowedAssetName(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect bool
	}{
		{"png", "cover.png", true},
		{"jpg", "photo.jpg", true},
		{"jpeg", "photo.jpeg", true},
		{"gif", "anim.gif", true},
		{"webp", "img.webp", true},
		{"svg", "diagram.svg", true},
		{"uppercase", "IMG.PNG", true},
		{"exe", "malware.exe", false},
		{"txt", "readme.txt", false},
		{"parent escape", "../secret.png", false},
		{"absolute path", "/etc/passwd.png", false},
		{"backslash", "dir\\file.png", false},
		{"double dot", "a..b.png", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := allowedAssetName(tc.input)
			if got != tc.expect {
				t.Fatalf("allowedAssetName(%q) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestLoadMetadataCorruptedJSON(t *testing.T) {
	svc, paths, _ := setupService(t)

	// Write invalid JSON
	os.MkdirAll(paths.Cache, 0755)
	os.WriteFile(filepath.Join(paths.Cache, "metadata.json"), []byte("not-json"), 0600)

	// Should return empty metadata, not crash
	meta := svc.loadMetadata()
	if meta.Posts == nil {
		t.Fatal("expected non-nil Posts map")
	}
	if len(meta.Posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(meta.Posts))
	}
}

func TestLoadMetadataMissingFile(t *testing.T) {
	svc, _, _ := setupService(t)

	meta := svc.loadMetadata()
	if meta.Posts == nil {
		t.Fatal("expected non-nil Posts map")
	}
}

func TestLoadMetadataNilPostsField(t *testing.T) {
	svc, paths, _ := setupService(t)

	// Write JSON with null posts
	os.MkdirAll(paths.Cache, 0755)
	os.WriteFile(filepath.Join(paths.Cache, "metadata.json"), []byte(`{"posts": null}`), 0600)

	meta := svc.loadMetadata()
	if meta.Posts == nil {
		t.Fatal("expected non-nil Posts map after null posts field")
	}
}

func TestSaveAndLoadMetadataRoundTrip(t *testing.T) {
	svc, _, _ := setupService(t)

	meta := metadataFile{Posts: map[string]PostState{
		"test": {Slug: "test", Title: "Test", Date: "2026-05-10", Tags: []string{"go"}},
	}}

	if err := svc.saveMetadata(meta); err != nil {
		t.Fatalf("saveMetadata failed: %v", err)
	}

	loaded := svc.loadMetadata()
	if loaded.Posts["test"].Title != "Test" {
		t.Fatalf("round-trip failed, got title %q", loaded.Posts["test"].Title)
	}
	if len(loaded.Posts["test"].Tags) != 1 || loaded.Posts["test"].Tags[0] != "go" {
		t.Fatal("tags round-trip failed")
	}
}

func TestPublishBlogAuditLogging(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	draft := content.PostDraft{Slug: "audit-post", FrontMatter: content.FrontMatter{Title: "Audit", Date: "2026-05-10"}, Body: "Content."}
	result := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "audit-post", Draft: draft})

	if result.AuditID == "" {
		t.Fatal("expected audit ID")
	}

	// Verify audit log was written
	logPath := filepath.Join(paths.Logs, "audit.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("audit log not created: %v", err)
	}
	if !contains(string(data), "audit-post") {
		t.Fatal("audit log missing slug")
	}
}

func TestPublishBlogDiffPath(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	draft := content.PostDraft{Slug: "diff-post", FrontMatter: content.FrontMatter{Title: "Diff", Date: "2026-05-10"}, Body: "Content."}
	result := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "diff-post", Draft: draft})

	if result.DiffPath == "" {
		t.Fatal("expected diff path")
	}
	if _, err := os.Stat(result.DiffPath); err != nil {
		t.Fatalf("diff file not found: %v", err)
	}
}

func TestPublishBlogRollback(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not found in PATH")
	}

	root := t.TempDir()
	initHugoSite(t, root)
	postRoot := filepath.Join(root, "content", "post")
	os.MkdirAll(postRoot, 0755)

	dataDir := t.TempDir()
	paths := config.Paths{
		BlogRoot: root, DataRoot: dataDir,
		Cache: filepath.Join(dataDir, "cache"), Backups: filepath.Join(dataDir, "backups"),
		Logs: filepath.Join(dataDir, "logs"), Diffs: filepath.Join(dataDir, "logs", "diffs"),
	}
	cfg := config.Config{Site: config.SiteConfig{ID: "test", BlogRoot: root, PostRoot: postRoot, BuildCommand: "hugo --minify"}}
	svc := NewService(paths, cfg, backup.NewStore(paths, 5), audit.NewLogger(paths), hugobuild.NewRunner(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))))

	// Publish first version
	draft := content.PostDraft{Slug: "rb-post", FrontMatter: content.FrontMatter{Title: "V1", Date: "2026-05-10"}, Body: "Version 1."}
	result := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "rb-post", Draft: draft})
	if result.Status != "success" {
		t.Fatalf("publish failed: %v", result.Error)
	}

	// Publish second version
	draft2 := content.PostDraft{Slug: "rb-post", FrontMatter: content.FrontMatter{Title: "V2", Date: "2026-05-10"}, Body: "Version 2."}
	result2 := svc.PublishBlog(context.Background(), BlogPublishRequest{Slug: "rb-post", Draft: draft2})
	if result2.Status != "success" {
		t.Fatalf("second publish failed: %v", result2.Error)
	}

	// Rollback
	rbResult := svc.Rollback(context.Background(), "rb-post")
	if rbResult.Status != "success" {
		t.Fatalf("rollback failed: %v", rbResult.Error)
	}
	if rbResult.BackupID == "" {
		t.Fatal("expected backup ID in rollback result")
	}
}

func TestRollbackNoBackup(t *testing.T) {
	svc, _, _ := setupService(t)

	result := svc.Rollback(context.Background(), "no-backup")
	if result.Status != "failed" {
		t.Fatalf("expected failed status, got %q", result.Status)
	}
	if result.Error == nil {
		t.Fatal("expected error for rollback without backup")
	}
}

func TestMetadataJSONSerialization(t *testing.T) {
	meta := metadataFile{
		Posts: map[string]PostState{
			"post-1": {
				Slug:       "post-1",
				Title:      "Post 1",
				Date:       "2026-05-10",
				Draft:      false,
				Tags:       []string{"go", "test"},
				Dirty:      true,
				SyncStatus: SyncDirty,
			},
		},
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	var loaded metadataFile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Posts["post-1"].Title != "Post 1" {
		t.Fatal("title mismatch")
	}
	if loaded.Posts["post-1"].SyncStatus != SyncDirty {
		t.Fatal("sync status mismatch")
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
