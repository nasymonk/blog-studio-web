package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- normalize (tested indirectly via Store.Load) ---

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg, appErr := store.Load()
	if appErr != nil {
		t.Fatalf("expected nil error, got %v", appErr)
	}

	if cfg.BasePath != "/studio" {
		t.Errorf("BasePath = %q, want %q", cfg.BasePath, "/studio")
	}
	if cfg.Site.PostSection != "post" {
		t.Errorf("PostSection = %q, want %q", cfg.Site.PostSection, "post")
	}
	if cfg.Preview.TTLMinutes != 120 {
		t.Errorf("TTLMinutes = %d, want %d", cfg.Preview.TTLMinutes, 120)
	}
	if cfg.MaxUploadBytes != 10*1024*1024 {
		t.Errorf("MaxUploadBytes = %d, want %d", cfg.MaxUploadBytes, 10*1024*1024)
	}
	if cfg.Site.BuildCommand != "hugo --minify" {
		t.Errorf("BuildCommand = %q, want %q", cfg.Site.BuildCommand, "hugo --minify")
	}
	if cfg.Site.BlogRoot != paths.BlogRoot {
		t.Errorf("BlogRoot = %q, want %q", cfg.Site.BlogRoot, paths.BlogRoot)
	}
}

func TestNormalize_FillsEmptyFields(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	// Save a config with only BlogRoot set (everything else empty).
	// normalize should fill in defaults on Load.
	partial := Config{
		BasePath: "/studio",
		Site:     SiteConfig{BlogRoot: paths.BlogRoot},
	}
	writeConfigJSON(t, paths.Config, partial)

	cfg, appErr := store.Load()
	if appErr != nil {
		t.Fatalf("Load error: %v", appErr)
	}

	if cfg.Site.PostSection != "post" {
		t.Errorf("PostSection = %q, want %q", cfg.Site.PostSection, "post")
	}
	if cfg.Site.BuildCommand != "hugo --minify" {
		t.Errorf("BuildCommand = %q, want %q", cfg.Site.BuildCommand, "hugo --minify")
	}
	if cfg.Site.ContentRoot == "" {
		t.Error("ContentRoot should be filled by normalize")
	}
	if cfg.Site.PostRoot == "" {
		t.Error("PostRoot should be filled by normalize")
	}
	if cfg.Site.PublicRoot == "" {
		t.Error("PublicRoot should be filled by normalize")
	}
	if cfg.Preview.TTLMinutes != 120 {
		t.Errorf("TTLMinutes = %d, want 120 (default)", cfg.Preview.TTLMinutes)
	}
	if cfg.MaxUploadBytes != 10*1024*1024 {
		t.Errorf("MaxUploadBytes = %d, want default", cfg.MaxUploadBytes)
	}
}

func TestNormalize_TrimsTrailingSlashFromBasePath(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio/",
		Site:     SiteConfig{BlogRoot: paths.BlogRoot},
	}
	writeConfigJSON(t, paths.Config, cfg)

	loaded, appErr := store.Load()
	if appErr != nil {
		t.Fatalf("Load error: %v", appErr)
	}
	if loaded.BasePath != "/studio" {
		t.Errorf("BasePath = %q, want %q (trailing slash trimmed)", loaded.BasePath, "/studio")
	}
}

func TestNormalize_PreservesExistingValues(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath:       "/custom",
		MaxUploadBytes: 5 * 1024 * 1024,
		Preview:        PreviewConfig{TTLMinutes: 30},
		Site: SiteConfig{
			BlogRoot:    paths.BlogRoot,
			PostSection: "articles",
			BuildCommand: "hugo",
		},
	}
	writeConfigJSON(t, paths.Config, cfg)

	loaded, appErr := store.Load()
	if appErr != nil {
		t.Fatalf("Load error: %v", appErr)
	}
	if loaded.BasePath != "/custom" {
		t.Errorf("BasePath = %q, want %q", loaded.BasePath, "/custom")
	}
	if loaded.MaxUploadBytes != 5*1024*1024 {
		t.Errorf("MaxUploadBytes = %d, want %d", loaded.MaxUploadBytes, 5*1024*1024)
	}
	if loaded.Preview.TTLMinutes != 30 {
		t.Errorf("TTLMinutes = %d, want 30", loaded.Preview.TTLMinutes)
	}
	if loaded.Site.PostSection != "articles" {
		t.Errorf("PostSection = %q, want %q", loaded.Site.PostSection, "articles")
	}
	if loaded.Site.BuildCommand != "hugo" {
		t.Errorf("BuildCommand = %q, want %q", loaded.Site.BuildCommand, "hugo")
	}
}

// --- validate (tested indirectly via Store.Save) ---

func TestSave_RejectsBasePathWithoutLeadingSlash(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "studio",
		Site:     defaultSiteConfig(paths),
	}
	appErr := store.Save(cfg)
	if appErr == nil {
		t.Fatal("expected error for basePath without leading slash, got nil")
	}
	if appErr.Code != "CONFIG_INVALID" {
		t.Errorf("error code = %q, want %q", appErr.Code, "CONFIG_INVALID")
	}
}

func TestSave_RejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio",
		Site: SiteConfig{
			BlogRoot:    paths.BlogRoot,
			ContentRoot: filepath.Join(paths.BlogRoot, "content"),
			PostRoot:    filepath.Join(dir, "evil"), // outside blog root
			PublicRoot:  filepath.Join(paths.BlogRoot, "public"),
			BuildCommand: "hugo --minify",
		},
	}
	appErr := store.Save(cfg)
	if appErr == nil {
		t.Fatal("expected error for path outside blog root, got nil")
	}
	if appErr.Code != "CONFIG_INVALID" {
		t.Errorf("error code = %q, want %q", appErr.Code, "CONFIG_INVALID")
	}
}

func TestSave_RejectsInvalidBuildCommand(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio",
		Site: SiteConfig{
			BlogRoot:     paths.BlogRoot,
			ContentRoot:  filepath.Join(paths.BlogRoot, "content"),
			PostRoot:     filepath.Join(paths.BlogRoot, "content", "post"),
			PublicRoot:   filepath.Join(paths.BlogRoot, "public"),
			BuildCommand: "rm -rf /",
		},
	}
	appErr := store.Save(cfg)
	if appErr == nil {
		t.Fatal("expected error for invalid buildCommand, got nil")
	}
}

func TestSave_AcceptsValidConfig(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio",
		Site:     defaultSiteConfig(paths),
	}
	appErr := store.Save(cfg)
	if appErr != nil {
		t.Fatalf("expected nil error, got %v", appErr)
	}
}

// --- Load / Save round-trip ---

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	original := Config{
		BasePath:       "/blog-admin",
		MaxUploadBytes: 20 * 1024 * 1024,
		Preview:        PreviewConfig{TTLMinutes: 60},
		Site: SiteConfig{
			ID:           "my-blog",
			Name:         "Test Blog",
			Theme:        "PaperMod",
			BlogRoot:     paths.BlogRoot,
			ContentRoot:  filepath.Join(paths.BlogRoot, "content"),
			PostSection:  "posts",
			PostRoot:     filepath.Join(paths.BlogRoot, "content", "posts"),
			BuildCommand: "hugo",
			PublicRoot:   filepath.Join(paths.BlogRoot, "public"),
		},
	}

	// Save
	if err := store.Save(original); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Load
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	// Verify key fields survived the round-trip
	if loaded.BasePath != original.BasePath {
		t.Errorf("BasePath = %q, want %q", loaded.BasePath, original.BasePath)
	}
	if loaded.Site.ID != original.Site.ID {
		t.Errorf("ID = %q, want %q", loaded.Site.ID, original.Site.ID)
	}
	if loaded.Site.Name != original.Site.Name {
		t.Errorf("Name = %q, want %q", loaded.Site.Name, original.Site.Name)
	}
	if loaded.Site.PostSection != original.Site.PostSection {
		t.Errorf("PostSection = %q, want %q", loaded.Site.PostSection, original.Site.PostSection)
	}
	if loaded.Site.BuildCommand != original.Site.BuildCommand {
		t.Errorf("BuildCommand = %q, want %q", loaded.Site.BuildCommand, original.Site.BuildCommand)
	}
	if loaded.MaxUploadBytes != original.MaxUploadBytes {
		t.Errorf("MaxUploadBytes = %d, want %d", loaded.MaxUploadBytes, original.MaxUploadBytes)
	}
	if loaded.Preview.TTLMinutes != original.Preview.TTLMinutes {
		t.Errorf("TTLMinutes = %d, want %d", loaded.Preview.TTLMinutes, original.Preview.TTLMinutes)
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	// Write garbage JSON
	if err := os.MkdirAll(filepath.Dir(paths.Config), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.Config, []byte("{invalid"), 0600); err != nil {
		t.Fatal(err)
	}

	_, appErr := store.Load()
	if appErr == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if appErr.Code != "CONFIG_PARSE_FAILED" {
		t.Errorf("error code = %q, want %q", appErr.Code, "CONFIG_PARSE_FAILED")
	}
}

func TestSave_CreatesConfigDirectory(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	// Config path nested deep so directory doesn't exist yet
	paths.Config = filepath.Join(dir, "data", "sub", "config.json")
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio",
		Site:     defaultSiteConfig(paths),
	}
	appErr := store.Save(cfg)
	if appErr != nil {
		t.Fatalf("Save should create dirs, got error: %v", appErr)
	}

	if _, err := os.Stat(paths.Config); err != nil {
		t.Fatalf("config file should exist: %v", err)
	}
}

// --- inside() tested indirectly ---

func TestSave_AcceptBlogRootEqualToRoot(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	store := NewStore(paths)

	cfg := Config{
		BasePath: "/studio",
		Site: SiteConfig{
			BlogRoot:     paths.BlogRoot,
			ContentRoot:  paths.BlogRoot, // equal to root is allowed
			PostRoot:     paths.BlogRoot,
			PublicRoot:   paths.BlogRoot,
			BuildCommand: "hugo --minify",
		},
	}
	appErr := store.Save(cfg)
	if appErr != nil {
		t.Fatalf("paths equal to blog root should be accepted, got: %v", appErr)
	}
}

// --- helpers ---

func testPaths(dir string) Paths {
	blogRoot := filepath.Join(dir, "blog")
	return Paths{
		BlogRoot: blogRoot,
		DataRoot: filepath.Join(dir, "data"),
		Config:   filepath.Join(dir, "data", "config.json"),
		Cache:    filepath.Join(dir, "data", "cache"),
		Backups:  filepath.Join(dir, "data", "backups"),
		Logs:     filepath.Join(dir, "data", "logs"),
		Diffs:    filepath.Join(dir, "data", "logs", "diffs"),
		Preview:  filepath.Join(dir, "data", "preview"),
		Static:   filepath.Join(dir, "web", "dist"),
		Trash:    filepath.Join(dir, "data", "trash"),
	}
}

func defaultSiteConfig(paths Paths) SiteConfig {
	return SiteConfig{
		BlogRoot:     paths.BlogRoot,
		ContentRoot:  filepath.Join(paths.BlogRoot, "content"),
		PostSection:  "post",
		PostRoot:     filepath.Join(paths.BlogRoot, "content", "post"),
		PublicRoot:   filepath.Join(paths.BlogRoot, "public"),
		BuildCommand: "hugo --minify",
	}
}

func writeConfigJSON(t *testing.T, path string, cfg Config) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
