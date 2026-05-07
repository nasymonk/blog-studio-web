package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeJoinRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	if _, err := SafeJoin(root, "post", "hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := SafeJoin(root, "..", "secret"); err == nil {
		t.Fatal("expected traversal to fail")
	}
}

func TestAtomicWriteFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "nested", "index.md")
	if err := AtomicWriteFile(target, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}
