package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/content"
)

func SafeJoin(root string, parts ...string) (string, *apperror.AppError) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", apperror.Wrap("PATH_INVALID", "根路径无效。", err, "检查挂载目录配置。", false)
	}
	target := filepath.Join(append([]string{rootAbs}, parts...)...)
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", apperror.Wrap("PATH_INVALID", "目标路径无效。", err, "检查路径参数。", false)
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", apperror.New("PATH_TRAVERSAL", "路径越界。", targetAbs, "检查 slug 或文件名，禁止使用上级目录。", false)
	}
	return targetAbs, nil
}

func PostDir(postRoot, slug string) (string, *apperror.AppError) {
	if err := content.ValidateSlug(slug); err != nil {
		return "", apperror.Wrap("SLUG_INVALID", "文章 slug 不合法。", err, "仅使用字母、数字、横线和下划线。", false)
	}
	return SafeJoin(postRoot, slug)
}

func AtomicWriteFile(pathName string, data []byte, perm os.FileMode) *apperror.AppError {
	if err := os.MkdirAll(filepath.Dir(pathName), 0755); err != nil {
		return apperror.Wrap("WRITE_DIR_FAILED", "无法创建目录。", err, "检查容器挂载目录权限。", true)
	}
	tmp, err := os.CreateTemp(filepath.Dir(pathName), ".blog-studio-*.tmp")
	if err != nil {
		return apperror.Wrap("WRITE_TEMP_FAILED", "无法创建临时文件。", err, "检查目录权限。", true)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return apperror.Wrap("WRITE_FAILED", "写入临时文件失败。", err, "检查磁盘空间。", true)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return apperror.Wrap("WRITE_CLOSE_FAILED", "关闭临时文件失败。", err, "检查磁盘状态。", true)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return apperror.Wrap("WRITE_CHMOD_FAILED", "设置文件权限失败。", err, "检查文件系统权限。", true)
	}
	if err := os.Rename(tmpName, pathName); err != nil {
		_ = os.Remove(tmpName)
		return apperror.Wrap("WRITE_RENAME_FAILED", "无法原子替换文件。", err, "检查目标文件权限。", true)
	}
	return nil
}

func CopyDir(src, dst string, exclude func(string, os.DirEntry) bool) *apperror.AppError {
	return wrap(filepath.WalkDir(src, func(pathName string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, pathName)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if exclude != nil && exclude(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		in, err := os.Open(pathName)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}), "COPY_FAILED", "复制目录失败。")
}

func wrap(err error, code, message string) *apperror.AppError {
	if err == nil {
		return nil
	}
	return apperror.Wrap(code, message, err, "检查源目录、目标目录和磁盘空间。", true)
}
