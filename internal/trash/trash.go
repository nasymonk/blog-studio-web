package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
)

type Item struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	SiteID    string `json:"siteId"`
	DeletedAt string `json:"deletedAt"`
	Size      int64  `json:"size"`
}

type Store struct {
	paths config.Paths
}

func NewStore(paths config.Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) Move(siteID, slug, srcDir string) (string, *apperror.AppError) {
	if err := os.MkdirAll(filepath.Join(s.paths.Trash, siteID), 0700); err != nil {
		return "", apperror.Wrap("TRASH_DIR_FAILED", "无法创建回收站目录。", err, "检查 /data/trash 权限。", true)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	id := ts + "-" + slug
	dst := filepath.Join(s.paths.Trash, siteID, id)
	if err := os.Rename(srcDir, dst); err != nil {
		return "", apperror.Wrap("TRASH_MOVE_FAILED", "无法移动文章到回收站。", err, "检查文件权限。", true)
	}
	return id, nil
}

func (s *Store) List(siteID string) ([]Item, *apperror.AppError) {
	dir := filepath.Join(s.paths.Trash, siteID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, apperror.Wrap("TRASH_LIST_FAILED", "无法读取回收站。", err, "检查权限。", true)
	}
	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := parseSlug(entry.Name())
		deletedAt := parseTime(entry.Name())
		size := dirSize(filepath.Join(dir, entry.Name()))
		items = append(items, Item{
			ID:        entry.Name(),
			Slug:      slug,
			SiteID:    siteID,
			DeletedAt: deletedAt,
			Size:      size,
		})
	}
	return items, nil
}

func (s *Store) Restore(siteID, id, postRoot string) *apperror.AppError {
	src := filepath.Join(s.paths.Trash, siteID, id)
	slug := parseSlug(id)
	dst := filepath.Join(postRoot, slug)
	if _, err := os.Stat(dst); err == nil {
		return apperror.New("TRASH_RESTORE_CONFLICT", fmt.Sprintf("文章 %s 已存在，无法还原。", slug), id, "先删除现有文章再还原。", false)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return apperror.Wrap("TRASH_RESTORE_MKDIR", "无法创建目录。", err, "检查权限。", true)
	}
	if err := os.Rename(src, dst); err != nil {
		return apperror.Wrap("TRASH_RESTORE_FAILED", "无法从回收站还原文章。", err, "检查文件权限。", true)
	}
	return nil
}

func (s *Store) Purge(siteID, id string) *apperror.AppError {
	target := filepath.Join(s.paths.Trash, siteID, id)
	if err := os.RemoveAll(target); err != nil {
		return apperror.Wrap("TRASH_PURGE_FAILED", "无法永久删除文章。", err, "检查权限。", true)
	}
	return nil
}

func (s *Store) PruneOlderThan(d time.Duration) *apperror.AppError {
	sites, err := os.ReadDir(s.paths.Trash)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperror.Wrap("TRASH_PRUNE_READ", "无法扫描回收站。", err, "检查权限。", true)
	}
	cutoff := time.Now().UTC().Add(-d)
	for _, site := range sites {
		siteDir := filepath.Join(s.paths.Trash, site.Name())
		items, err := os.ReadDir(siteDir)
		if err != nil {
			continue
		}
		for _, item := range items {
			t, err := time.Parse("20060102T150405Z", item.Name()[:16])
			if err != nil {
				continue
			}
			if t.Before(cutoff) {
				_ = os.RemoveAll(filepath.Join(siteDir, item.Name()))
			}
		}
	}
	return nil
}

func parseSlug(id string) string {
	if len(id) > 17 {
		return id[17:]
	}
	return id
}

func parseTime(id string) string {
	if len(id) < 16 {
		return ""
	}
	t, err := time.Parse("20060102T150405Z", id[:16])
	if err != nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func dirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
