package storage

import (
	"io"
	"os"
	"path/filepath"

	"blog-studio-web/internal/apperror"
)

// SyncDir incrementally mirrors src into dst, only writing files whose
// mtime+size differ. Files in dst that are absent from src are removed.
func SyncDir(src, dst string, exclude func(string, os.DirEntry) bool) *apperror.AppError {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return apperror.Wrap("SYNC_MKDIR_FAILED", "无法创建目标目录。", err, "检查目录权限。", true)
	}
	if err := syncDown(src, dst, exclude); err != nil {
		return apperror.Wrap("SYNC_FAILED", "目录同步失败。", err, "检查源目录、目标目录和磁盘空间。", true)
	}
	return nil
}

func syncDown(src, dst string, exclude func(string, os.DirEntry) bool) error {
	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	srcNames := make(map[string]struct{}, len(srcEntries))
	for _, entry := range srcEntries {
		rel := entry.Name()
		if exclude != nil && exclude(rel, entry) {
			continue
		}
		srcNames[rel] = struct{}{}
		srcPath := filepath.Join(src, rel)
		dstPath := filepath.Join(dst, rel)
		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := syncDown(srcPath, dstPath, func(r string, e os.DirEntry) bool {
				return exclude != nil && exclude(rel+"/"+r, e)
			}); err != nil {
				return err
			}
		} else {
			if err := syncFile(srcPath, dstPath, entry); err != nil {
				return err
			}
		}
	}
	// Remove dst entries not present in src.
	dstEntries, err := os.ReadDir(dst)
	if err != nil {
		return err
	}
	for _, entry := range dstEntries {
		if _, keep := srcNames[entry.Name()]; !keep {
			_ = os.RemoveAll(filepath.Join(dst, entry.Name()))
		}
	}
	return nil
}

func syncFile(src, dst string, srcEntry os.DirEntry) error {
	srcInfo, err := srcEntry.Info()
	if err != nil {
		return err
	}
	if dstInfo, err := os.Stat(dst); err == nil {
		if dstInfo.Size() == srcInfo.Size() && !srcInfo.ModTime().After(dstInfo.ModTime()) {
			return nil // unchanged
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
