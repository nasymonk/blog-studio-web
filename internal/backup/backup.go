package backup

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/storage"
)

type Store struct {
	paths config.Paths
	keep  int
}

func NewStore(paths config.Paths, keep int) *Store {
	return &Store{paths: paths, keep: keep}
}

func (s *Store) Create(siteID, slug, sourceDir string) (string, string, *apperror.AppError) {
	backupID := time.Now().Format("20060102-150405")
	target := filepath.Join(s.paths.Backups, siteID, slug, backupID)
	if _, err := os.Stat(sourceDir); err == nil {
		if err := storage.CopyDir(sourceDir, target, nil); err != nil {
			return "", "", err
		}
	} else if err := os.MkdirAll(target, 0700); err != nil {
		return "", "", apperror.Wrap("BACKUP_DIR_FAILED", "无法创建空备份目录。", err, "检查 /data/backups 权限。", true)
	}
	if err := s.prune(siteID, slug); err != nil {
		return "", "", err
	}
	return backupID, target, nil
}

func (s *Store) Latest(siteID, slug string) (string, string, *apperror.AppError) {
	root := filepath.Join(s.paths.Backups, siteID, slug)
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", "", apperror.Wrap("BACKUP_NOT_FOUND", "没有可回滚的备份。", err, "先发布一次文章以生成备份。", false)
	}
	ids := backupIDs(entries)
	if len(ids) == 0 {
		return "", "", apperror.New("BACKUP_NOT_FOUND", "没有可回滚的备份。", "backup directory empty", "先发布一次文章以生成备份。", false)
	}
	id := ids[len(ids)-1]
	return id, filepath.Join(root, id), nil
}

func (s *Store) prune(siteID, slug string) *apperror.AppError {
	root := filepath.Join(s.paths.Backups, siteID, slug)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	ids := backupIDs(entries)
	for len(ids) > s.keep {
		oldest := ids[0]
		if err := os.RemoveAll(filepath.Join(root, oldest)); err != nil {
			return apperror.Wrap("BACKUP_PRUNE_FAILED", "无法清理旧备份。", err, "手动检查 /data/backups。", true)
		}
		ids = ids[1:]
	}
	return nil
}

func backupIDs(entries []os.DirEntry) []string {
	ids := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	return ids
}
