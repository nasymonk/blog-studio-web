package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/hugobuild"
	"blog-studio-web/internal/metrics"
	"blog-studio-web/internal/storage"
)

type Result struct {
	PreviewID   string                  `json:"previewId"`
	URL         string                  `json:"url"`
	BuildResult hugobuild.CommandResult `json:"buildResult"`
	ExpiresAt   string                  `json:"expiresAt"`
	Error       *apperror.AppError      `json:"error,omitempty"`
}

type Service struct {
	paths   config.Paths
	cfg     config.Config
	content *content.Service
	runner  *hugobuild.Runner
}

func NewService(paths config.Paths, cfg config.Config, runner *hugobuild.Runner) *Service {
	return &Service{paths: paths, cfg: cfg, content: content.NewService(), runner: runner}
}

func (s *Service) Create(ctx context.Context, draft content.PostDraft) Result {
	id := slugID(draft.Slug)
	workDir := filepath.Join(s.paths.Preview, "work", id)
	publicDir := filepath.Join(s.paths.Preview, "public", id)
	expiresAt := time.Now().Add(time.Duration(s.cfg.Preview.TTLMinutes) * time.Minute)
	result := Result{
		PreviewID: id,
		URL:       s.cfg.BasePath + "/preview/" + id + "/",
		ExpiresAt: expiresAt.Format(time.RFC3339Nano),
	}

	if err := storage.SyncDir(s.cfg.Site.BlogRoot, workDir, func(rel string, entry os.DirEntry) bool {
		return rel == ".git" || rel == "public" || rel == "resources" || rel == ".hugo_build.lock"
	}); err != nil {
		result.Error = err
		return result
	}

	raw, err := s.content.Compose(draft)
	if err != nil {
		result.Error = apperror.Wrap("PREVIEW_DRAFT_INVALID", "预览草稿无效。", err, "检查文章标题、slug 和 front matter。", false)
		return result
	}
	postDir, postErr := storage.SafeJoin(workDir, "content", s.cfg.Site.PostSection, draft.Slug)
	if postErr != nil {
		result.Error = postErr
		return result
	}
	if err := storage.AtomicWriteFile(filepath.Join(postDir, "index.md"), raw, 0644); err != nil {
		result.Error = err
		return result
	}

	result.BuildResult = s.runner.RunWithLabel(ctx, "preview", workDir,
		"--destination", publicDir,
		"--baseURL", result.URL,
		"--buildDrafts",
		"--buildFuture",
	)
	if !result.BuildResult.Success {
		result.Error = apperror.New("PREVIEW_BUILD_FAILED", "Hugo 预览构建失败。", result.BuildResult.Stderr, "检查草稿内容和 Hugo 配置。", true)
	}
	return result
}

func (s *Service) Cleanup() *apperror.AppError {
	publicRoot := filepath.Join(s.paths.Preview, "public")
	workRoot := filepath.Join(s.paths.Preview, "work")
	entries, err := os.ReadDir(publicRoot)
	if err != nil {
		if os.IsNotExist(err) {
			s.pruneOrphanWork(workRoot, nil)
			return nil
		}
		return apperror.Wrap("PREVIEW_CLEAN_FAILED", "无法读取预览目录。", err, "检查 /data/preview 权限。", true)
	}
	cutoff := time.Now().Add(-time.Duration(s.cfg.Preview.TTLMinutes) * time.Minute)
	alive := map[string]struct{}{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(publicRoot, entry.Name()))
			_ = os.RemoveAll(filepath.Join(workRoot, entry.Name()))
		} else {
			alive[entry.Name()] = struct{}{}
		}
	}
	s.pruneOrphanWork(workRoot, alive)
	metrics.PreviewActive.Set(float64(len(alive)))
	return nil
}

func (s *Service) pruneOrphanWork(workRoot string, alive map[string]struct{}) {
	entries, err := os.ReadDir(workRoot)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if _, keep := alive[entry.Name()]; !keep {
			_ = os.RemoveAll(filepath.Join(workRoot, entry.Name()))
		}
	}
}

func slugID(slug string) string {
	h := sha256.Sum256([]byte(slug))
	return fmt.Sprintf("prev-%s", hex.EncodeToString(h[:4]))
}
