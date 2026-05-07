package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"blog-studio-web/internal/config"
)

func TestBackupPruneKeepsNewest(t *testing.T) {
	root := t.TempDir()
	paths := config.Paths{Backups: filepath.Join(root, "backups")}
	store := NewStore(paths, 2)
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "index.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, _, err := store.Create("site", "slug", src); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1100 * time.Millisecond)
	}
	entries, err := os.ReadDir(filepath.Join(paths.Backups, "site", "slug"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(entries))
	}
}
