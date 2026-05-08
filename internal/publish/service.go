package publish

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/backup"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/hugobuild"
	"blog-studio-web/internal/storage"
)

type SyncStatus string

const (
	SyncClean    SyncStatus = "clean"
	SyncDirty    SyncStatus = "dirty"
	SyncStale    SyncStatus = "stale"
	SyncConflict SyncStatus = "conflict"
	SyncUnknown  SyncStatus = "unknown"
)

type PostState struct {
	Slug              string     `json:"slug"`
	Title             string     `json:"title"`
	Date              string     `json:"date"`
	Draft             bool       `json:"draft"`
	Tags              []string   `json:"tags"`
	Categories        []string   `json:"categories"`
	RemotePath        string     `json:"remotePath"`
	Dirty             bool       `json:"dirty"`
	RemoteMtime       string     `json:"remoteMtime"`
	CachedRemoteMtime string     `json:"cachedRemoteMtime"`
	SyncStatus        SyncStatus `json:"syncStatus"`
	LastSyncedAt      string     `json:"lastSyncedAt"`
	LatestBackupID    string     `json:"latestBackupId"`
	Large             bool       `json:"large"`
}

type metadataFile struct {
	Posts map[string]PostState `json:"posts"`
}

type Conflict struct {
	Slug              string `json:"slug"`
	RemoteMtime       string `json:"remoteMtime"`
	CachedRemoteMtime string `json:"cachedRemoteMtime"`
	Message           string `json:"message"`
}

type BlogPublishRequest struct {
	Slug             string            `json:"slug"`
	Draft            content.PostDraft `json:"draft"`
	ConfirmOverwrite bool              `json:"confirmOverwrite"`
	Files            map[string][]byte `json:"files,omitempty"`
}

type CommandResult = hugobuild.CommandResult

type Result struct {
	Target        string             `json:"target"`
	Status        string             `json:"status"`
	BackupID      string             `json:"backupId,omitempty"`
	UploadedFiles []string           `json:"uploadedFiles"`
	Conflicts     []Conflict         `json:"conflicts"`
	DiffPath      string             `json:"diffPath,omitempty"`
	AuditID       string             `json:"auditId"`
	BuildResult   CommandResult      `json:"buildResult"`
	ChannelResult interface{}        `json:"channelResult,omitempty"`
	Error         *apperror.AppError `json:"error,omitempty"`
}

type Service struct {
	paths   config.Paths
	cfg     config.Config
	content *content.Service
	backup  *backup.Store
	audit   *audit.Logger
	runner  *hugobuild.Runner
}

func NewService(paths config.Paths, cfg config.Config, backupStore *backup.Store, auditLogger *audit.Logger, runner *hugobuild.Runner) *Service {
	return &Service{paths: paths, cfg: cfg, content: content.NewService(), backup: backupStore, audit: auditLogger, runner: runner}
}

func (s *Service) ListPosts() ([]PostState, *apperror.AppError) {
	entries, err := os.ReadDir(s.cfg.Site.PostRoot)
	if err != nil {
		return nil, apperror.Wrap("POST_LIST_FAILED", "无法读取文章目录。", err, "检查 /blog/content/post 挂载和权限。", true)
	}
	meta := s.loadMetadata()
	out := []PostState{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		state, err := s.stateFromDir(entry.Name(), meta)
		if err != nil {
			continue
		}
		meta.Posts[state.Slug] = state
		out = append(out, state)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date > out[j].Date })
	_ = s.saveMetadata(meta)
	return out, nil
}

func (s *Service) LoadPost(slug string) (content.PostDraft, *apperror.AppError) {
	postDir, err := storage.PostDir(s.cfg.Site.PostRoot, slug)
	if err != nil {
		return content.PostDraft{}, err
	}
	raw, readErr := os.ReadFile(filepath.Join(postDir, "index.md"))
	if readErr != nil {
		return content.PostDraft{}, apperror.Wrap("POST_READ_FAILED", "无法读取文章。", readErr, "检查文章目录和 index.md。", true)
	}
	fm, body, parseErr := s.content.Parse(raw)
	if parseErr != nil {
		return content.PostDraft{}, apperror.Wrap("POST_PARSE_FAILED", "无法解析文章 front matter。", parseErr, "检查 YAML 格式。", false)
	}
	assets, _ := listAssets(postDir)
	return content.PostDraft{Slug: slug, FrontMatter: fm, Body: body, Assets: assets, Large: s.content.IsLargeMarkdown(body)}, nil
}

func (s *Service) SaveDraft(draft content.PostDraft) *apperror.AppError {
	raw, err := s.content.Compose(draft)
	if err != nil {
		return apperror.Wrap("DRAFT_INVALID", "草稿内容无效。", err, "检查标题、slug 和 front matter。", false)
	}
	dir, appErr := storage.SafeJoin(s.paths.Cache, "posts", draft.Slug)
	if appErr != nil {
		return appErr
	}
	if err := storage.AtomicWriteFile(filepath.Join(dir, "index.md"), raw, 0600); err != nil {
		return err
	}
	meta := s.loadMetadata()
	state := meta.Posts[draft.Slug]
	state.Slug = draft.Slug
	state.Title = draft.FrontMatter.Title
	state.Date = draft.FrontMatter.Date
	state.Draft = draft.FrontMatter.Draft
	state.Tags = draft.FrontMatter.Tags
	state.Categories = draft.FrontMatter.Categories
	state.Dirty = true
	state.SyncStatus = computeStatus(true, parseTime(state.CachedRemoteMtime), parseTime(state.RemoteMtime))
	meta.Posts[draft.Slug] = state
	return s.saveMetadata(meta)
}

func (s *Service) PublishBlog(ctx context.Context, req BlogPublishRequest) Result {
	auditID := audit.NewID("publish")
	result := Result{Target: "blog", Status: "failed", AuditID: auditID}
	if err := s.SaveDraft(req.Draft); err != nil {
		result.Error = err
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	postDir, err := storage.PostDir(s.cfg.Site.PostRoot, req.Slug)
	if err != nil {
		result.Error = err
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	indexPath := filepath.Join(postDir, "index.md")
	oldRaw, _ := os.ReadFile(indexPath)
	stat, statErr := os.Stat(indexPath)
	meta := s.loadMetadata()
	if statErr == nil {
		cached := parseTime(meta.Posts[req.Slug].CachedRemoteMtime)
		if !cached.IsZero() && stat.ModTime().After(cached) && !req.ConfirmOverwrite {
			result.Status = "conflict"
			result.Conflicts = []Conflict{{Slug: req.Slug, RemoteMtime: formatTime(stat.ModTime()), CachedRemoteMtime: formatTime(cached), Message: "远程文章比缓存更新。"}}
			s.appendAudit(req.Slug, "publish", result)
			return result
		}
	}
	backupID, _, backupErr := s.backup.Create(s.cfg.Site.ID, req.Slug, postDir)
	if backupErr != nil {
		result.Error = backupErr
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	result.BackupID = backupID
	raw, composeErr := s.content.Compose(req.Draft)
	if composeErr != nil {
		result.Error = apperror.Wrap("POST_COMPOSE_FAILED", "无法生成文章 Markdown。", composeErr, "检查 front matter。", false)
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	if writeErr := storage.AtomicWriteFile(indexPath, raw, 0644); writeErr != nil {
		result.Error = writeErr
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	result.UploadedFiles = append(result.UploadedFiles, "index.md")
	for rel, data := range req.Files {
		if !allowedAssetName(rel) {
			result.Error = apperror.New("ASSET_NAME_INVALID", "资源文件名不合法。", rel, "仅允许文章目录内的图片文件。", false)
			s.appendAudit(req.Slug, "publish", result)
			return result
		}
		target, safeErr := storage.SafeJoin(postDir, rel)
		if safeErr != nil {
			result.Error = safeErr
			s.appendAudit(req.Slug, "publish", result)
			return result
		}
		if err := storage.AtomicWriteFile(target, data, 0644); err != nil {
			result.Error = err
			s.appendAudit(req.Slug, "publish", result)
			return result
		}
		result.UploadedFiles = append(result.UploadedFiles, rel)
	}
	diffPath, _ := s.audit.WriteDiff(auditID, s.content.Diff(oldRaw, raw))
	result.DiffPath = diffPath
	result.BuildResult = s.runHugo(ctx)
	if !result.BuildResult.Success {
		result.Error = apperror.New("HUGO_BUILD_FAILED", "Hugo 构建失败。", result.BuildResult.Stderr, "检查文章内容、Hugo 配置和主题。", true)
		s.appendAudit(req.Slug, "publish", result)
		return result
	}
	newStat, _ := os.Stat(indexPath)
	state := meta.Posts[req.Slug]
	state.Dirty = false
	state.RemoteMtime = formatTime(newStat.ModTime())
	state.CachedRemoteMtime = state.RemoteMtime
	state.LastSyncedAt = formatTime(time.Now())
	state.LatestBackupID = backupID
	state.SyncStatus = SyncClean
	meta.Posts[req.Slug] = state
	_ = s.saveMetadata(meta)
	result.Status = "success"
	s.appendAudit(req.Slug, "publish", result)
	return result
}

func (s *Service) Rollback(ctx context.Context, slug string) Result {
	auditID := audit.NewID("rollback")
	result := Result{Target: "blog", Status: "failed", AuditID: auditID}
	backupID, backupDir, err := s.backup.Latest(s.cfg.Site.ID, slug)
	if err != nil {
		result.Error = err
		s.appendAudit(slug, "rollback", result)
		return result
	}
	postDir, postErr := storage.PostDir(s.cfg.Site.PostRoot, slug)
	if postErr != nil {
		result.Error = postErr
		s.appendAudit(slug, "rollback", result)
		return result
	}
	if copyErr := storage.CopyDir(backupDir, postDir, nil); copyErr != nil {
		result.Error = copyErr
		s.appendAudit(slug, "rollback", result)
		return result
	}
	result.BackupID = backupID
	result.BuildResult = s.runHugo(ctx)
	if !result.BuildResult.Success {
		result.Error = apperror.New("HUGO_BUILD_FAILED", "回滚后 Hugo 构建失败。", result.BuildResult.Stderr, "检查备份内容和 Hugo 配置。", true)
		s.appendAudit(slug, "rollback", result)
		return result
	}
	result.Status = "success"
	s.appendAudit(slug, "rollback", result)
	return result
}

func (s *Service) stateFromDir(slug string, meta metadataFile) (PostState, error) {
	postDir := filepath.Join(s.cfg.Site.PostRoot, slug)
	indexPath := filepath.Join(postDir, "index.md")
	stat, err := os.Stat(indexPath)
	if err != nil {
		return PostState{}, err
	}
	fm := content.FrontMatter{Title: slug}
	body := ""
	if stat.Size() <= content.MaxMarkdownBytes {
		if raw, err := os.ReadFile(indexPath); err == nil {
			if parsedFM, parsedBody, err := s.content.Parse(raw); err == nil {
				fm = parsedFM
				body = parsedBody
			}
		}
	}
	cached := meta.Posts[slug]
	state := PostState{
		Slug: slug, Title: fm.Title, Date: fm.Date, Draft: fm.Draft, Tags: fm.Tags, Categories: fm.Categories,
		RemotePath: indexPath, Dirty: cached.Dirty, RemoteMtime: formatTime(stat.ModTime()), CachedRemoteMtime: cached.CachedRemoteMtime,
		LastSyncedAt: cached.LastSyncedAt, LatestBackupID: cached.LatestBackupID, Large: s.content.IsLargeMarkdown(body),
	}
	if state.CachedRemoteMtime == "" {
		state.CachedRemoteMtime = state.RemoteMtime
		state.LastSyncedAt = formatTime(time.Now())
	}
	state.SyncStatus = computeStatus(state.Dirty, parseTime(state.CachedRemoteMtime), stat.ModTime())
	return state, nil
}

func (s *Service) loadMetadata() metadataFile {
	meta := metadataFile{Posts: map[string]PostState{}}
	data, err := os.ReadFile(filepath.Join(s.paths.Cache, "metadata.json"))
	if err != nil {
		return meta
	}
	_ = json.Unmarshal(data, &meta)
	if meta.Posts == nil {
		meta.Posts = map[string]PostState{}
	}
	return meta
}

func (s *Service) saveMetadata(meta metadataFile) *apperror.AppError {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return apperror.Wrap("CACHE_ENCODE_FAILED", "无法序列化缓存。", err, "重新同步文章列表。", false)
	}
	return storage.AtomicWriteFile(filepath.Join(s.paths.Cache, "metadata.json"), data, 0600)
}

func (s *Service) appendAudit(slug, op string, result Result) {
	entry := audit.Entry{AuditID: result.AuditID, Actor: "admin", SiteID: s.cfg.Site.ID, Slug: slug, Operation: op, Target: result.Target, Result: result.Status, BackupID: result.BackupID, DiffPath: result.DiffPath, BuildResult: result.BuildResult, ChannelResult: result.ChannelResult}
	if result.Error != nil {
		entry.ErrorCode = result.Error.Code
		entry.ErrorBrief = result.Error.Message
	}
	_ = s.audit.Append(entry)
}

func (s *Service) runHugo(ctx context.Context) CommandResult {
	args := []string{}
	if s.cfg.Site.BuildCommand != "hugo" {
		args = []string{"--minify"}
	}
	return s.runner.RunWithLabel(ctx, "publish", s.cfg.Site.BlogRoot, args...)
}

func computeStatus(dirty bool, cached time.Time, remoteTime time.Time) SyncStatus {
	if remoteTime.IsZero() {
		return SyncUnknown
	}
	remoteNewer := !cached.IsZero() && remoteTime.After(cached)
	if dirty && remoteNewer {
		return SyncConflict
	}
	if dirty {
		return SyncDirty
	}
	if remoteNewer {
		return SyncStale
	}
	return SyncClean
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed
	}
	parsed, _ = time.Parse(time.RFC3339, value)
	return parsed
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}

func listAssets(dir string) ([]content.Asset, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	assets := []content.Asset{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "index.md" {
			continue
		}
		info, err := entry.Info()
		if err == nil {
			assets = append(assets, content.Asset{Name: entry.Name(), Size: info.Size()})
		}
	}
	return assets, nil
}

func allowedAssetName(name string) bool {
	if strings.Contains(name, "..") || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return true
	default:
		return false
	}
}
