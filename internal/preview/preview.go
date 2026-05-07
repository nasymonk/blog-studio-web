package preview

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/publish"
	"blog-studio-web/internal/storage"
	"github.com/google/uuid"
)

type Result struct {
	PreviewID   string                `json:"previewId"`
	URL         string                `json:"url"`
	BuildResult publish.CommandResult `json:"buildResult"`
	ExpiresAt   string                `json:"expiresAt"`
	Error       *apperror.AppError    `json:"error,omitempty"`
}

type Service struct {
	paths   config.Paths
	cfg     config.Config
	content *content.Service
}

func NewService(paths config.Paths, cfg config.Config) *Service {
	return &Service{paths: paths, cfg: cfg, content: content.NewService()}
}

func (s *Service) Create(ctx context.Context, draft content.PostDraft) Result {
	id := uuid.NewString()
	workDir := filepath.Join(s.paths.Preview, "work", id)
	publicDir := filepath.Join(s.paths.Preview, "public", id)
	expiresAt := time.Now().Add(time.Duration(s.cfg.Preview.TTLMinutes) * time.Minute)
	result := Result{PreviewID: id, URL: s.cfg.BasePath + "/preview/" + id + "/", ExpiresAt: expiresAt.Format(time.RFC3339Nano)}
	_ = os.RemoveAll(workDir)
	_ = os.RemoveAll(publicDir)
	if err := storage.CopyDir(s.cfg.Site.BlogRoot, workDir, func(rel string, entry os.DirEntry) bool {
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
	result.BuildResult = runPreviewHugo(ctx, workDir, publicDir, result.URL)
	if !result.BuildResult.Success {
		result.Error = apperror.New("PREVIEW_BUILD_FAILED", "Hugo 预览构建失败。", result.BuildResult.Stderr, "检查草稿内容和 Hugo 配置。", true)
		return result
	}
	return result
}

func (s *Service) Cleanup() *apperror.AppError {
	root := filepath.Join(s.paths.Preview, "public")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperror.Wrap("PREVIEW_CLEAN_FAILED", "无法读取预览目录。", err, "检查 /data/preview 权限。", true)
	}
	cutoff := time.Now().Add(-time.Duration(s.cfg.Preview.TTLMinutes) * time.Minute)
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(root, entry.Name()))
			_ = os.RemoveAll(filepath.Join(s.paths.Preview, "work", entry.Name()))
		}
	}
	return nil
}

func runPreviewHugo(ctx context.Context, workDir, publicDir, baseURL string) publish.CommandResult {
	start := time.Now()
	result := publish.CommandResult{StartedAt: start.Format(time.RFC3339Nano), ExitCode: -1}
	cmd := exec.CommandContext(ctx, "hugo", "--destination", publicDir, "--baseURL", baseURL, "--buildDrafts", "--buildFuture")
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.FinishedAt = time.Now().Format(time.RFC3339Nano)
	if err == nil {
		result.Success = true
		result.ExitCode = 0
		return result
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	return result
}
