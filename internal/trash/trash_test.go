package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blog-studio-web/internal/config"
)

// helper creates a temp dir layout and returns a Store + paths.
func setup(t *testing.T) (*Store, string, config.Paths) {
	t.Helper()
	tmp := t.TempDir()
	paths := config.Paths{
		BlogRoot: filepath.Join(tmp, "blog"),
		Trash:    filepath.Join(tmp, "data", "trash"),
	}
	store := NewStore(paths)
	return store, tmp, paths
}

// createPostDir makes a fake post directory with a markdown file.
func createPostDir(t *testing.T, dir, slug string) string {
	t.Helper()
	postDir := filepath.Join(dir, slug)
	if err := os.MkdirAll(postDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := filepath.Join(postDir, "index.md")
	if err := os.WriteFile(md, []byte("# "+slug+"\nHello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	return postDir
}

// --- Move ---

func TestMove_Success(t *testing.T) {
	store, _, _ := setup(t)
	siteID := "default"
	srcDir := createPostDir(t, t.TempDir(), "hello-world")

	id, appErr := store.Move(siteID, "hello-world", srcDir)
	if appErr != nil {
		t.Fatalf("Move returned error: %v", appErr)
	}
	if id == "" {
		t.Fatal("Move returned empty id")
	}
	// ID should end with the slug
	if !strings.HasSuffix(id, "-hello-world") {
		t.Fatalf("id %q does not end with -hello-world", id)
	}
	// Source should no longer exist
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Fatalf("source dir still exists after Move")
	}
	// Destination should exist
	dst := filepath.Join(store.paths.Trash, siteID, id)
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("trash destination missing: %v", err)
	}
	// Verify file content preserved
	data, err := os.ReadFile(filepath.Join(dst, "index.md"))
	if err != nil {
		t.Fatalf("reading moved file: %v", err)
	}
	if !strings.Contains(string(data), "hello-world") {
		t.Fatalf("file content mismatch: %s", data)
	}
}

func TestMove_CreatesTrashDir(t *testing.T) {
	store, _, paths := setup(t)
	srcDir := createPostDir(t, t.TempDir(), "post-a")

	// Trash dir should not exist yet
	if _, err := os.Stat(paths.Trash); !os.IsNotExist(err) {
		t.Fatal("trash dir should not exist before Move")
	}

	_, appErr := store.Move("site1", "post-a", srcDir)
	if appErr != nil {
		t.Fatalf("Move returned error: %v", appErr)
	}

	// Now the site subdirectory should exist
	siteDir := filepath.Join(paths.Trash, "site1")
	if _, err := os.Stat(siteDir); err != nil {
		t.Fatalf("site trash dir not created: %v", err)
	}
}

func TestMove_SrcNotExist(t *testing.T) {
	store, _, _ := setup(t)
	_, appErr := store.Move("default", "missing", "/nonexistent/path")
	if appErr == nil {
		t.Fatal("expected error for nonexistent source")
	}
	if appErr.Code != "TRASH_MOVE_FAILED" {
		t.Fatalf("expected TRASH_MOVE_FAILED, got %s", appErr.Code)
	}
}

// --- List ---

func TestList_Empty(t *testing.T) {
	store, _, _ := setup(t)
	items, appErr := store.List("default")
	if appErr != nil {
		t.Fatalf("List returned error: %v", appErr)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestList_NonexistentSite(t *testing.T) {
	store, _, _ := setup(t)
	items, appErr := store.List("no-such-site")
	if appErr != nil {
		t.Fatalf("List returned error: %v", appErr)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestList_ReturnsMovedItems(t *testing.T) {
	store, _, _ := setup(t)
	siteID := "default"

	// Move two posts
	src1 := createPostDir(t, t.TempDir(), "alpha")
	src2 := createPostDir(t, t.TempDir(), "beta")
	id1, _ := store.Move(siteID, "alpha", src1)
	id2, _ := store.Move(siteID, "beta", src2)

	items, appErr := store.List(siteID)
	if appErr != nil {
		t.Fatalf("List returned error: %v", appErr)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Build a map by ID for order-independent checking
	byID := map[string]Item{}
	for _, it := range items {
		byID[it.ID] = it
	}

	it1, ok := byID[id1]
	if !ok {
		t.Fatalf("item with id %q not found", id1)
	}
	if it1.Slug != "alpha" {
		t.Fatalf("expected slug alpha, got %s", it1.Slug)
	}
	if it1.SiteID != siteID {
		t.Fatalf("expected siteId %s, got %s", siteID, it1.SiteID)
	}
	if it1.DeletedAt == "" {
		t.Fatal("deletedAt should not be empty")
	}
	if it1.Size <= 0 {
		t.Fatalf("size should be > 0, got %d", it1.Size)
	}

	it2, ok := byID[id2]
	if !ok {
		t.Fatalf("item with id %q not found", id2)
	}
	if it2.Slug != "beta" {
		t.Fatalf("expected slug beta, got %s", it2.Slug)
	}
}

func TestList_SkipsFiles(t *testing.T) {
	store, _, _ := setup(t)
	siteID := "default"

	// Move one post
	src := createPostDir(t, t.TempDir(), "post-x")
	store.Move(siteID, "post-x", src)

	// Also create a plain file in the site trash dir (should be ignored)
	trashSiteDir := filepath.Join(store.paths.Trash, siteID)
	os.MkdirAll(trashSiteDir, 0o700)
	os.WriteFile(filepath.Join(trashSiteDir, "readme.txt"), []byte("not a post"), 0o644)

	items, appErr := store.List(siteID)
	if appErr != nil {
		t.Fatalf("List returned error: %v", appErr)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (file should be skipped), got %d", len(items))
	}
	if items[0].Slug != "post-x" {
		t.Fatalf("expected slug post-x, got %s", items[0].Slug)
	}
}

// --- Restore ---

func TestRestore_Success(t *testing.T) {
	store, _, paths := setup(t)
	siteID := "default"
	postRoot := filepath.Join(paths.BlogRoot, "content", "post")
	os.MkdirAll(postRoot, 0o755)

	// Move a post to trash
	srcDir := createPostDir(t, t.TempDir(), "my-post")
	id, _ := store.Move(siteID, "my-post", srcDir)

	// Restore
	appErr := store.Restore(siteID, id, postRoot)
	if appErr != nil {
		t.Fatalf("Restore returned error: %v", appErr)
	}

	// Restored dir should exist at postRoot/slug
	restored := filepath.Join(postRoot, "my-post")
	if _, err := os.Stat(restored); err != nil {
		t.Fatalf("restored dir missing: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(restored, "index.md"))
	if err != nil {
		t.Fatalf("reading restored file: %v", err)
	}
	if !strings.Contains(string(data), "my-post") {
		t.Fatalf("restored file content mismatch: %s", data)
	}

	// Trash entry should be gone
	trashEntry := filepath.Join(store.paths.Trash, siteID, id)
	if _, err := os.Stat(trashEntry); !os.IsNotExist(err) {
		t.Fatal("trash entry should be removed after restore")
	}
}

func TestRestore_Conflict(t *testing.T) {
	store, _, paths := setup(t)
	siteID := "default"
	postRoot := filepath.Join(paths.BlogRoot, "content", "post")

	// Move a post to trash
	srcDir := createPostDir(t, t.TempDir(), "conflict-post")
	id, _ := store.Move(siteID, "conflict-post", srcDir)

	// Create a conflicting post at the destination
	createPostDir(t, postRoot, "conflict-post")

	// Restore should fail with conflict
	appErr := store.Restore(siteID, id, postRoot)
	if appErr == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if appErr.Code != "TRASH_RESTORE_CONFLICT" {
		t.Fatalf("expected TRASH_RESTORE_CONFLICT, got %s", appErr.Code)
	}
	if appErr.Retryable {
		t.Fatal("conflict should not be retryable")
	}
}

func TestRestore_NonexistentTrashEntry(t *testing.T) {
	store, _, paths := setup(t)
	postRoot := filepath.Join(paths.BlogRoot, "content", "post")
	os.MkdirAll(postRoot, 0o755)

	appErr := store.Restore("default", "20060102T150405Z-no-such", postRoot)
	if appErr == nil {
		t.Fatal("expected error for nonexistent trash entry")
	}
	if appErr.Code != "TRASH_RESTORE_FAILED" {
		t.Fatalf("expected TRASH_RESTORE_FAILED, got %s", appErr.Code)
	}
}

func TestRestore_PreservesOtherTrashItems(t *testing.T) {
	store, _, paths := setup(t)
	siteID := "default"
	postRoot := filepath.Join(paths.BlogRoot, "content", "post")
	os.MkdirAll(postRoot, 0o755)

	// Move two posts
	src1 := createPostDir(t, t.TempDir(), "keep-me")
	src2 := createPostDir(t, t.TempDir(), "restore-me")
	store.Move(siteID, "keep-me", src1)
	id2, _ := store.Move(siteID, "restore-me", src2)

	// Restore only the second
	appErr := store.Restore(siteID, id2, postRoot)
	if appErr != nil {
		t.Fatalf("Restore error: %v", appErr)
	}

	// The first should still be in trash
	items, _ := store.List(siteID)
	if len(items) != 1 {
		t.Fatalf("expected 1 remaining trash item, got %d", len(items))
	}
	if items[0].Slug != "keep-me" {
		t.Fatalf("expected slug keep-me, got %s", items[0].Slug)
	}
}
