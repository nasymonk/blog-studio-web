package storage

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestSafeJoinValidPaths(t *testing.T) {
	root := t.TempDir()

	cases := []struct {
		name  string
		parts []string
	}{
		{"single part", []string{"hello"}},
		{"nested", []string{"a", "b", "c"}},
		{"with dots in name", []string{"post.with.dots"}},
		{"with spaces", []string{"my post"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SafeJoin(root, tc.parts...)
			if err != nil {
				t.Fatal(err)
			}
			if !filepath.IsAbs(result) {
				t.Fatalf("expected absolute path, got %s", result)
			}
		})
	}
}

func TestSafeJoinTraversalCases(t *testing.T) {
	root := t.TempDir()

	cases := []struct {
		name  string
		parts []string
	}{
		{"parent escape", []string{"..", "secret"}},
		{"nested escape", []string{"a", "..", "..", "secret"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SafeJoin(root, tc.parts...)
			if err == nil {
				t.Fatal("expected traversal error")
			}
			if err.Code != "PATH_TRAVERSAL" {
				t.Fatalf("expected PATH_TRAVERSAL code, got %s", err.Code)
			}
		})
	}
}

func TestPostDirValidSlug(t *testing.T) {
	root := t.TempDir()
	result, err := PostDir(root, "my-post")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(result, "my-post") {
		t.Fatalf("expected path to end with 'my-post', got %s", result)
	}
}

func TestPostDirInvalidSlug(t *testing.T) {
	root := t.TempDir()

	cases := []struct {
		name string
		slug string
	}{
		{"traversal", "../escape"},
		{"starts with dash", "-bad"},
		{"contains space", "has space"},
		{"empty", ""},
		{"contains slash", "has/slash"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := PostDir(root, tc.slug)
			if err == nil {
				t.Fatalf("expected error for slug %q", tc.slug)
			}
		})
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

func TestAtomicWriteFileOverwrite(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "file.txt")

	if err := AtomicWriteFile(target, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWriteFile(target, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second" {
		t.Fatalf("expected 'second', got %q", string(data))
	}
}

func TestAtomicWriteFilePermissions(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "perms.txt")

	if err := AtomicWriteFile(target, []byte("content"), 0755); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("expected perm 0755, got %o", info.Mode().Perm())
	}
}

func TestAtomicWriteFileEmptyData(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "empty.txt")

	if err := AtomicWriteFile(target, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty file, got %d bytes", len(data))
	}
}

func TestAtomicWriteFileLargeData(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "large.txt")

	// Write 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := AtomicWriteFile(target, data, 0644); err != nil {
		t.Fatal(err)
	}
	read, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(read) != len(data) {
		t.Fatalf("size mismatch: wrote %d, read %d", len(data), len(read))
	}
}

func TestCopyDirBasic(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create source structure
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644)

	if err := CopyDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected content: %q", string(data))
	}

	data, err = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Fatalf("unexpected content: %q", string(data))
	}
}

func TestCopyDirWithExclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(src, "skip.log"), []byte("skip"), 0644)

	exclude := func(rel string, entry os.DirEntry) bool {
		return strings.HasSuffix(rel, ".log")
	}

	if err := CopyDir(src, dst, exclude); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Fatal("keep.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "skip.log")); !os.IsNotExist(err) {
		t.Fatal("skip.log should not exist")
	}
}

func TestCopyDirExcludeDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.MkdirAll(filepath.Join(src, "skipdir"), 0755)
	os.WriteFile(filepath.Join(src, "skipdir", "file.txt"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0644)

	exclude := func(rel string, entry os.DirEntry) bool {
		return entry.IsDir() && rel == "skipdir"
	}

	if err := CopyDir(src, dst, exclude); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Fatal("keep.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "skipdir")); !os.IsNotExist(err) {
		t.Fatal("skipdir should not exist")
	}
}

func TestCopyDirEmptySource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := CopyDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dst)
	if len(entries) != 0 {
		t.Fatalf("expected empty dst, got %d entries", len(entries))
	}
}

func TestCopyDirPreservesFileMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	target := filepath.Join(src, "script.sh")
	os.WriteFile(target, []byte("#!/bin/bash"), 0755)

	if err := CopyDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dst, "script.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("expected 0755, got %o", info.Mode().Perm())
	}
}

func TestSyncDirBasic(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644)

	if err := SyncDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected data: %q", string(data))
	}

	data, err = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestSyncDirRemovesDeletedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Initial sync
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("b"), 0644)
	SyncDir(src, dst, nil)

	// Remove one file from src
	os.Remove(filepath.Join(src, "a.txt"))

	if err := SyncDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "a.txt")); !os.IsNotExist(err) {
		t.Fatal("a.txt should have been removed from dst")
	}
	if _, err := os.Stat(filepath.Join(dst, "b.txt")); err != nil {
		t.Fatal("b.txt should still exist")
	}
}

func TestSyncDirUpdatesChangedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "a.txt"), []byte("original"), 0644)
	SyncDir(src, dst, nil)

	// Update source file
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("updated"), 0644)

	if err := SyncDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "updated" {
		t.Fatalf("expected 'updated', got %q", string(data))
	}
}

func TestSyncDirWithExclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(src, "skip.log"), []byte("skip"), 0644)

	exclude := func(rel string, entry os.DirEntry) bool {
		return strings.HasSuffix(rel, ".log")
	}

	if err := SyncDir(src, dst, exclude); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Fatal("keep.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "skip.log")); !os.IsNotExist(err) {
		t.Fatal("skip.log should not exist")
	}
}

func TestSyncDirEmptySource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(dst, "old.txt"), []byte("old"), 0644)

	if err := SyncDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "old.txt")); !os.IsNotExist(err) {
		t.Fatal("old.txt should have been removed")
	}
}

func TestSyncDirNoChangeSkip(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.WriteFile(filepath.Join(src, "a.txt"), []byte("content"), 0644)
	SyncDir(src, dst, nil)

	// Sync again without changes - should be a no-op (skipped by size+mtime check)
	if err := SyncDir(src, dst, nil); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "content" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestConcurrentAtomicWrite(t *testing.T) {
	root := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			path := filepath.Join(root, "file.txt")
			data := strings.Repeat("x", 100)
			if err := AtomicWriteFile(path, []byte(data), 0644); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent write failed: %v", err)
	}

	// File should exist and be valid
	data, err := os.ReadFile(filepath.Join(root, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 100 {
		t.Fatalf("unexpected data length: %d", len(data))
	}
}
