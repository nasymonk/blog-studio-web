package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/backup"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/hugobuild"
	"blog-studio-web/internal/metrics"
	"blog-studio-web/internal/preview"
	"blog-studio-web/internal/publish"
	"blog-studio-web/internal/storage"
	"blog-studio-web/internal/trash"
	"blog-studio-web/internal/wechat"
)

type APIResponse struct {
	OK    bool               `json:"ok"`
	Data  interface{}        `json:"data,omitempty"`
	Error *apperror.AppError `json:"error,omitempty"`
}

type Server struct {
	mu           sync.RWMutex
	paths        config.Paths
	cfgStore     *config.Store
	cfg          config.Config
	adminHash    string
	sessions     *auth.Store
	audit        *audit.Logger
	backupStore  *backup.Store
	trashStore   *trash.Store
	pub          *publish.Service
	prev         *preview.Service
	wec          *wechat.Service
	runner       *hugobuild.Runner
	logger       *slog.Logger
	loginLimiter *auth.LoginLimiter
	writeLimiter *writeRateLimiter
	staticPrefix string
}

func New(paths config.Paths, cfg config.Config, cfgStore *config.Store, adminHash string, sessions *auth.Store, logger *slog.Logger) *Server {
	return NewWithAudit(paths, cfg, cfgStore, adminHash, sessions, audit.NewLogger(paths), logger)
}

func NewWithAudit(paths config.Paths, cfg config.Config, cfgStore *config.Store, adminHash string, sessions *auth.Store, auditLogger *audit.Logger, logger *slog.Logger) *Server {
	backupStore := backup.NewStore(paths, 5)
	runner := hugobuild.NewRunner(logger)
	return &Server{
		paths:        paths,
		cfgStore:     cfgStore,
		cfg:          cfg,
		adminHash:    adminHash,
		sessions:     sessions,
		audit:        auditLogger,
		backupStore:  backupStore,
		trashStore:   trash.NewStore(paths),
		runner:       runner,
		logger:       logger,
		loginLimiter: auth.NewLoginLimiter(),
		writeLimiter: newWriteRateLimiter(10, 20),
		pub:          publish.NewService(paths, cfg, backupStore, auditLogger, runner),
		prev:         preview.NewService(paths, cfg, runner),
		wec:          wechat.NewService(paths, cfg, auditLogger),
	}
}

func (s *Server) PreviewService() *preview.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prev
}

func (s *Server) TrashStore() *trash.Store {
	return s.trashStore
}

// PublishScheduledPosts checks all posts for scheduledAt times that have passed
// and publishes them. Called by the background worker.
func (s *Server) PublishScheduledPosts() {
	pub := s.publisher()
	posts, err := pub.ListPosts()
	if err != nil {
		s.logger.Warn("scheduled publish: failed to list posts", "error", err)
		return
	}
	now := time.Now()
	for _, p := range posts {
		if !p.Draft {
			continue
		}
		draft, loadErr := pub.LoadPost(p.Slug)
		if loadErr != nil {
			continue
		}
		if draft.FrontMatter.ScheduledAt == "" {
			continue
		}
		scheduledTime, parseErr := time.Parse(time.RFC3339, draft.FrontMatter.ScheduledAt)
		if parseErr != nil {
			s.logger.Warn("scheduled publish: invalid scheduledAt", "slug", p.Slug, "scheduledAt", draft.FrontMatter.ScheduledAt)
			continue
		}
		if now.Before(scheduledTime) {
			continue
		}
		// Set draft to false and publish
		draft.FrontMatter.Draft = false
		draft.FrontMatter.ScheduledAt = ""
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		req := publish.BlogPublishRequest{
			Slug:             p.Slug,
			Draft:            draft,
			ConfirmOverwrite: true,
		}
		result := pub.PublishBlog(ctx, req)
		cancel()
		if result.Status == "success" {
			s.logger.Info("scheduled publish: published", "slug", p.Slug)
		} else {
			s.logger.Warn("scheduled publish: failed", "slug", p.Slug, "status", result.Status)
		}
	}
}


func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	api := http.NewServeMux()
	api.HandleFunc("POST /auth/login", s.withLoginLimit(s.login))
	api.HandleFunc("POST /auth/logout", s.withAuth(s.logout))
	api.HandleFunc("POST /auth/password", s.withWriteAuth(s.changePassword))
	api.HandleFunc("GET /session", s.session)
	api.HandleFunc("GET /posts", s.withAuth(s.listPosts))
	api.HandleFunc("GET /posts/", s.withAuth(s.postRouter))
	api.HandleFunc("PUT /posts/", s.withWriteAuth(s.postRouter))
	api.HandleFunc("POST /posts/", s.withWriteAuth(s.postRouter))
	api.HandleFunc("DELETE /posts/", s.withWriteAuth(s.postRouter))
	api.HandleFunc("POST /posts/bulk/trash", s.withWriteAuth(s.bulkTrash))
	api.HandleFunc("POST /posts/bulk/publish", s.withWriteAuth(s.bulkPublish))
	api.HandleFunc("POST /tags/rename", s.withWriteAuth(s.renameTag))
	api.HandleFunc("POST /tags/delete", s.withWriteAuth(s.deleteTag))
	api.HandleFunc("GET /site", s.withAuth(s.getSite))
	api.HandleFunc("PUT /site", s.withWriteAuth(s.putSite))
	api.HandleFunc("POST /site/avatar", s.withWriteAuth(s.uploadAvatar))
	api.HandleFunc("GET /pages/now", s.withAuth(s.getNowPage))
	api.HandleFunc("PUT /pages/now", s.withWriteAuth(s.putNowPage))
	api.HandleFunc("GET /health", s.healthPublic)
	api.HandleFunc("GET /health/full", s.withAuth(s.health))
	api.HandleFunc("GET /audit", s.withAuth(s.auditRecent))
	api.HandleFunc("GET /config", s.withAuth(s.getConfig))
	api.HandleFunc("PUT /config", s.withWriteAuth(s.putConfig))
	api.HandleFunc("GET /trash", s.withAuth(s.listTrash))
	api.HandleFunc("POST /trash/", s.withWriteAuth(s.trashRouter))
	api.HandleFunc("DELETE /trash/", s.withWriteAuth(s.trashRouter))
	api.Handle("GET /metrics", s.withAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.Handler().ServeHTTP(w, r)
	})))

	base := strings.TrimRight(s.cfg.BasePath, "/")
	mux.Handle(base+"/api/", http.StripPrefix(base+"/api", api))
	mux.Handle(base+"/preview/", http.StripPrefix(base+"/preview/", http.FileServer(http.Dir(filepath.Join(s.paths.Preview, "public")))))
	mux.Handle(base+"/", s.spaHandler(base))
	if base != "" {
		mux.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, base+"/", http.StatusMovedPermanently)
		})
	}
	chain := recoverer(s.logger)(
		requestID(
			accessLog(s.logger)(
				secureHeaders(
					withGzip(
						withMaxBytes(
							s.withWriteRateLimit(mux),
						),
					),
				),
			),
		),
	)
	return chain
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !auth.VerifyPassword(req.Password, s.snap().adminHash) {
		metrics.LoginAttempts.WithLabelValues("fail").Inc()
		writeError(w, http.StatusUnauthorized, apperror.New("LOGIN_FAILED", "密码错误。", "invalid admin password", "检查管理员密码。", false))
		return
	}
	metrics.LoginAttempts.WithLabelValues("ok").Inc()
	session := s.sessions.Create(w)
	writeOK(w, map[string]interface{}{"authenticated": true, "csrfToken": session.CSRF})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.sessions.Destroy(w, r)
	writeOK(w, map[string]bool{"authenticated": false})
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !auth.VerifyPassword(req.CurrentPassword, s.snap().adminHash) {
		writeError(w, http.StatusUnauthorized, apperror.New("PASSWORD_CURRENT_INVALID", "当前密码不正确。", "invalid current password", "重新输入当前管理员密码。", false))
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("PASSWORD_WEAK", "新密码不够安全。", err, "请使用至少 6 个字符的密码。", false))
		return
	}
	if err := os.WriteFile(s.paths.AdminHash, []byte(newHash+"\n"), 0600); err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("PASSWORD_WRITE_FAILED", "无法保存新密码。", err, "检查数据目录写入权限。", true))
		return
	}
	s.mu.Lock()
	s.adminHash = newHash
	s.mu.Unlock()
	writeOK(w, map[string]bool{"changed": true})
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	if session, ok := s.sessions.FromRequest(r); ok {
		writeOK(w, map[string]interface{}{"authenticated": true, "csrfToken": session.CSRF})
		return
	}
	writeOK(w, map[string]bool{"authenticated": false})
}

func (s *Server) publisher() *publish.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pub
}

func (s *Server) previewSvc() *preview.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prev
}

func (s *Server) wechatSvc() *wechat.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wec
}

func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := s.publisher().ListPosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, posts)
}

func (s *Server) postRouter(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/posts/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "接口不存在。", r.URL.Path, "检查 API 路径。", false))
		return
	}
	slug := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		s.loadPost(w, r, slug)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		s.deletePost(w, r, slug)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "draft":
			s.saveDraft(w, r, slug)
		case "assets":
			s.uploadAsset(w, r, slug)
		case "preview":
			s.createPreview(w, r, slug)
		case "rollback":
			s.rollback(w, r, slug)
		default:
			writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "接口不存在。", r.URL.Path, "检查 API 路径。", false))
		}
		return
	}
	if len(parts) == 3 && parts[1] == "publish" {
		switch parts[2] {
		case "blog":
			s.publishBlog(w, r, slug)
		case "wechat-draft":
			s.publishWechatDraft(w, r, slug)
		default:
			writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "发布目标不存在。", parts[2], "选择 blog 或 wechat-draft。", false))
		}
		return
	}
	writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "接口不存在。", r.URL.Path, "检查 API 路径。", false))
}

func (s *Server) loadPost(w http.ResponseWriter, r *http.Request, slug string) {
	post, err := s.publisher().LoadPost(slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeOK(w, post)
}

func (s *Server) deletePost(w http.ResponseWriter, r *http.Request, slug string) {
	sn := s.snap()
	postDir, err := storage.SafeJoin(sn.cfg.Site.BlogRoot, "content", sn.cfg.Site.PostSection, slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if _, statErr := os.Stat(postDir); os.IsNotExist(statErr) {
		writeError(w, http.StatusNotFound, apperror.New("POST_NOT_FOUND", "文章不存在。", slug, "检查 slug。", false))
		return
	}
	trashID, trashErr := s.trashStore.Move(sn.cfg.Site.ID, slug, postDir)
	if trashErr != nil {
		writeError(w, http.StatusInternalServerError, trashErr)
		return
	}
	s.publisher().InvalidateCache()
	_ = s.audit.Append(audit.Entry{
		AuditID: audit.NewID("audit"), Actor: "admin", SiteID: sn.cfg.Site.ID,
		Slug: slug, Operation: "delete", Target: "trash", Result: "ok", BackupID: trashID,
	})
	writeOK(w, map[string]string{"trashId": trashID})
}

func (s *Server) bulkTrash(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slugs []string `json:"slugs"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sn := s.snap()
	type result struct {
		Slug    string `json:"slug"`
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	results := make([]result, 0, len(req.Slugs))
	for _, slug := range req.Slugs {
		postDir, err := storage.SafeJoin(sn.cfg.Site.BlogRoot, "content", sn.cfg.Site.PostSection, slug)
		if err != nil {
			results = append(results, result{Slug: slug, Success: false, Error: err.Error()})
			continue
		}
		if _, statErr := os.Stat(postDir); os.IsNotExist(statErr) {
			results = append(results, result{Slug: slug, Success: false, Error: "not found"})
			continue
		}
		trashID, trashErr := s.trashStore.Move(sn.cfg.Site.ID, slug, postDir)
		if trashErr != nil {
			results = append(results, result{Slug: slug, Success: false, Error: trashErr.Error()})
			continue
		}
		_ = s.audit.Append(audit.Entry{
			AuditID: audit.NewID("audit"), Actor: "admin", SiteID: sn.cfg.Site.ID,
			Slug: slug, Operation: "delete", Target: "trash", Result: "ok", BackupID: trashID,
		})
		results = append(results, result{Slug: slug, Success: true})
	}
	s.publisher().InvalidateCache()
	writeOK(w, results)
}

func (s *Server) bulkPublish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slugs []string `json:"slugs"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	pub := s.publisher()
	type result struct {
		Slug    string `json:"slug"`
		Status  string `json:"status"`
		Error   string `json:"error,omitempty"`
	}
	results := make([]result, 0, len(req.Slugs))
	for _, slug := range req.Slugs {
		draft, loadErr := pub.LoadPost(slug)
		if loadErr != nil {
			results = append(results, result{Slug: slug, Status: "failed", Error: "load failed"})
			continue
		}
		draft.FrontMatter.Draft = false
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		publishReq := publish.BlogPublishRequest{
			Slug:             slug,
			Draft:            draft,
			ConfirmOverwrite: true,
		}
		publishResult := pub.PublishBlog(ctx, publishReq)
		cancel()
		if publishResult.Status == "success" {
			results = append(results, result{Slug: slug, Status: "success"})
		} else {
			errMsg := ""
			if publishResult.Error != nil {
				errMsg = publishResult.Error.Message
			}
			results = append(results, result{Slug: slug, Status: publishResult.Status, Error: errMsg})
		}
	}
	writeOK(w, results)
}

func (s *Server) renameTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.OldName == "" || req.NewName == "" {
		writeError(w, http.StatusBadRequest, apperror.New("INVALID_INPUT", "标签名不能为空。", "", "提供新旧标签名。", false))
		return
	}
	pub := s.publisher()
	posts, err := pub.ListPosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	updated := 0
	for _, p := range posts {
		draft, loadErr := pub.LoadPost(p.Slug)
		if loadErr != nil {
			continue
		}
		changed := false
		for i, tag := range draft.FrontMatter.Tags {
			if tag == req.OldName {
				draft.FrontMatter.Tags[i] = req.NewName
				changed = true
			}
		}
		if !changed {
			continue
		}
		if saveErr := pub.SaveDraft(draft); saveErr != nil {
			continue
		}
		updated++
	}
	pub.InvalidateCache()
	writeOK(w, map[string]interface{}{"updated": updated})
}

func (s *Server) deleteTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, apperror.New("INVALID_INPUT", "标签名不能为空。", "", "提供要删除的标签名。", false))
		return
	}
	pub := s.publisher()
	posts, err := pub.ListPosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	updated := 0
	for _, p := range posts {
		draft, loadErr := pub.LoadPost(p.Slug)
		if loadErr != nil {
			continue
		}
		newTags := make([]string, 0, len(draft.FrontMatter.Tags))
		changed := false
		for _, tag := range draft.FrontMatter.Tags {
			if tag == req.Name {
				changed = true
			} else {
				newTags = append(newTags, tag)
			}
		}
		if !changed {
			continue
		}
		draft.FrontMatter.Tags = newTags
		if saveErr := pub.SaveDraft(draft); saveErr != nil {
			continue
		}
		updated++
	}
	pub.InvalidateCache()
	writeOK(w, map[string]interface{}{"updated": updated})
}

func (s *Server) listTrash(w http.ResponseWriter, r *http.Request) {
	sn := s.snap()
	items, err := s.trashStore.List(sn.cfg.Site.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, items)
}

func (s *Server) trashRouter(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/trash/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "接口不存在。", r.URL.Path, "检查 API 路径。", false))
		return
	}
	id := parts[0]
	sn := s.snap()
	if r.Method == http.MethodDelete {
		if err := s.trashStore.Purge(sn.cfg.Site.ID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeOK(w, map[string]bool{"purged": true})
		return
	}
	if r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "restore" {
		postSection, safeErr := storage.SafeJoin(sn.cfg.Site.BlogRoot, "content", sn.cfg.Site.PostSection)
		if safeErr != nil {
			writeError(w, http.StatusBadRequest, safeErr)
			return
		}
		if err := s.trashStore.Restore(sn.cfg.Site.ID, id, postSection); err != nil {
			writeError(w, http.StatusConflict, err)
			return
		}
		s.publisher().InvalidateCache()
		_ = s.audit.Append(audit.Entry{
			AuditID: audit.NewID("audit"), Actor: "admin", SiteID: sn.cfg.Site.ID,
			Slug: id, Operation: "restore", Target: "trash", Result: "ok",
		})
		writeOK(w, map[string]bool{"restored": true})
		return
	}
	writeError(w, http.StatusNotFound, apperror.New("NOT_FOUND", "接口不存在。", r.URL.Path, "检查 API 路径。", false))
}

func (s *Server) saveDraft(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := decodeJSON(r, &draft); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	draft.Slug = slug
	if err := s.publisher().SaveDraft(draft); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeOK(w, map[string]bool{"saved": true})
}

func (s *Server) uploadAsset(w http.ResponseWriter, r *http.Request, slug string) {
	data, name, appErr := parseUploadedFile(r, s.cfg.MaxUploadBytes)
	if appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	postDir, appErr := storage.PostDir(s.cfg.Site.PostRoot, slug)
	if appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	if appErr := validateUpload(name, data); appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	target, safeErr := storage.SafeJoin(postDir, name)
	if safeErr != nil {
		writeError(w, http.StatusBadRequest, safeErr)
		return
	}
	if err := storage.AtomicWriteFile(target, data, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, content.Asset{Name: name, Size: int64(len(data))})
}

func (s *Server) publishBlog(w http.ResponseWriter, r *http.Request, slug string) {
	var req publish.BlogPublishRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Slug = slug
	req.Draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	result := s.publisher().PublishBlog(ctx, req)
	writeOK(w, result)
}

func (s *Server) publishWechatDraft(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := decodeJSON(r, &draft); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
	defer cancel()
	writeOK(w, s.wechatSvc().PublishDraft(ctx, draft))
}

func (s *Server) createPreview(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := decodeJSON(r, &draft); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	prev := s.previewSvc()
	_ = prev.Cleanup()
	writeOK(w, prev.Create(ctx, draft))
}

func (s *Server) rollback(w http.ResponseWriter, r *http.Request, slug string) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	writeOK(w, s.publisher().Rollback(ctx, slug))
}

func (s *Server) getSite(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Join(s.cfg.Site.BlogRoot, "hugo.toml"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("SITE_READ_FAILED", "读取站点配置失败。", err, "检查博客目录。", true))
		return
	}
	toml := string(data)
	writeOK(w, map[string]string{
		"description":  readHugoParam(toml, "description"),
		"profileImage": readHugoParam(toml, "profileImage"),
	})
}

func (s *Server) putSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description  string `json:"description"`
		ProfileImage string `json:"profileImage"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	hugoPath := filepath.Join(s.cfg.Site.BlogRoot, "hugo.toml")
	data, err := os.ReadFile(hugoPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("SITE_READ_FAILED", "读取站点配置失败。", err, "检查博客目录。", true))
		return
	}
	toml := string(data)
	if req.Description != "" {
		toml = setHugoParam(toml, "description", req.Description)
	}
	if req.ProfileImage != "" {
		toml = setHugoParam(toml, "profileImage", req.ProfileImage)
	}
	if err := storage.AtomicWriteFile(hugoPath, []byte(toml), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("SITE_WRITE_FAILED", "保存站点配置失败。", err, "检查文件权限。", true))
		return
	}
	writeOK(w, map[string]string{"description": req.Description, "profileImage": req.ProfileImage})
}

func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	data, name, appErr := parseUploadedFile(r, s.cfg.MaxUploadBytes)
	if appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	if appErr := validateUpload(name, data); appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	ext := strings.ToLower(filepath.Ext(name))
	staticDir := filepath.Join(s.cfg.Site.BlogRoot, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("AVATAR_DIR_FAILED", "无法创建 static 目录。", err, "检查文件权限。", true))
		return
	}
	for _, e := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp"} {
		_ = os.Remove(filepath.Join(staticDir, "avatar"+e))
	}
	avatarPath := filepath.Join(staticDir, "avatar"+ext)
	if err := storage.AtomicWriteFile(avatarPath, data, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("AVATAR_WRITE_FAILED", "保存头像失败。", err, "检查文件权限。", true))
		return
	}
	imagePath := "/avatar" + ext
	hugoPath := filepath.Join(s.cfg.Site.BlogRoot, "hugo.toml")
	if tomlData, err := os.ReadFile(hugoPath); err == nil {
		_ = storage.AtomicWriteFile(hugoPath, []byte(setHugoParam(string(tomlData), "profileImage", imagePath)), 0644)
	}
	writeOK(w, map[string]string{"path": imagePath})
}

func (s *Server) getNowPage(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.cfg.Site.BlogRoot, "content", "now.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeOK(w, map[string]string{"raw": "---\ntitle: \"Now\"\nurl: \"/now/\"\nlayout: \"single\"\nShowToc: false\nShowReadingTime: false\n---\n\n## 此刻在做\n"})
			return
		}
		writeError(w, http.StatusInternalServerError, apperror.Wrap("NOW_READ_FAILED", "读取 Now 页面失败。", err, "检查博客目录。", true))
		return
	}
	writeOK(w, map[string]string{"raw": string(data)})
}

func (s *Server) putNowPage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Raw string `json:"raw"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	path := filepath.Join(s.cfg.Site.BlogRoot, "content", "now.md")
	if err := storage.AtomicWriteFile(path, []byte(req.Raw), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, apperror.Wrap("NOW_WRITE_FAILED", "保存 Now 页面失败。", err, "检查文件权限。", true))
		return
	}
	writeOK(w, map[string]bool{"saved": true})
}

func (s *Server) healthPublic(w http.ResponseWriter, r *http.Request) {
	blogRoot := s.snap().cfg.Site.BlogRoot
	status := "ok"
	if _, err := os.Stat(blogRoot); err != nil {
		status = "error"
	}
	writeOK(w, map[string]string{"status": status})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()
	checks := []map[string]string{}
	add := func(name, status, message, detail, suggestion string) {
		checks = append(checks, map[string]string{"name": name, "status": status, "message": message, "technicalDetail": detail, "suggestion": suggestion})
	}
	if _, err := os.Stat(cfg.Site.BlogRoot); err != nil {
		add("blog-root", "error", "博客目录不可访问。", err.Error(), "检查 Docker 挂载 /blog。")
	} else {
		add("blog-root", "ok", "博客目录可访问。", cfg.Site.BlogRoot, "")
	}
	if _, err := os.Stat(cfg.Site.PostRoot); err != nil {
		add("post-root", "error", "文章目录不可访问。", err.Error(), "检查 content/post 路径。")
	} else {
		add("post-root", "ok", "文章目录可访问。", cfg.Site.PostRoot, "")
	}
	if err := os.MkdirAll(s.paths.DataRoot, 0700); err != nil {
		add("data-root", "error", "数据目录不可写。", err.Error(), "检查 /data 挂载权限。")
	} else {
		add("data-root", "ok", "数据目录可写。", s.paths.DataRoot, "")
	}
	// Hugo build test
	buildStart := time.Now()
	buildResult := s.runner.Run(context.Background(), cfg.Site.BlogRoot, "--renderToMemory")
	buildDuration := time.Since(buildStart)
	if !buildResult.Success {
		add("hugo-build", "error", "Hugo 构建测试失败。", buildResult.Stderr, "检查 hugo.toml 和主题是否正常。")
	} else {
		add("hugo-build", "ok", fmt.Sprintf("Hugo 构建测试通过（%dms）。", buildDuration.Milliseconds()), buildResult.Stdout, "")
	}
	// Disk space check
	var stat syscall.Statfs_t
	if err := syscall.Statfs(cfg.Site.BlogRoot, &stat); err != nil {
		add("disk-space", "error", "无法检查磁盘空间。", err.Error(), "检查文件系统挂载。")
	} else {
		freeGB := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
		if freeGB < 1.0 {
			add("disk-space", "warn", fmt.Sprintf("磁盘空间不足：%.1f GB 可用。", freeGB), fmt.Sprintf("%.2f GB free", freeGB), "清理磁盘空间或扩容。")
		} else {
			add("disk-space", "ok", fmt.Sprintf("磁盘空间充足：%.1f GB 可用。", freeGB), fmt.Sprintf("%.2f GB free", freeGB), "")
		}
	}
	status := "ok"
	for _, check := range checks {
		if check["status"] == "error" {
			status = "error"
			break
		}
		if check["status"] == "warn" {
			status = "warn"
		}
	}
	writeOK(w, map[string]interface{}{"status": status, "checks": checks})
}

func (s *Server) auditRecent(w http.ResponseWriter, r *http.Request) {
	items, err := s.audit.Recent(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, items)
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg
	cfg.Wechat.AppID = ""
	writeOK(w, cfg)
}

func (s *Server) putConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := decodeJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg.Wechat.AppID = ""
	if err := s.cfgStore.Save(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	saved, err := s.cfgStore.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.reloadConfig(saved)
	writeOK(w, saved)
}

func (s *Server) reloadConfig(cfg config.Config) {
	newPub := publish.NewService(s.paths, cfg, s.backupStore, s.audit, s.runner)
	newPrev := preview.NewService(s.paths, cfg, s.runner)
	newWec := wechat.NewService(s.paths, cfg, s.audit)
	s.mu.Lock()
	oldPub := s.pub
	s.cfg = cfg
	s.pub = newPub
	s.prev = newPrev
	s.wec = newWec
	s.mu.Unlock()
	// Invalidate the old service's cache so the new service starts fresh.
	oldPub.InvalidateCache()
}

func (s *Server) spaHandler(base string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, base+"/")
		if rel == "" {
			rel = "index.html"
		}
		target := filepath.Join(s.paths.Static, filepath.Clean(rel))
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		http.ServeFile(w, r, filepath.Join(s.paths.Static, "index.html"))
	})
}

func (s *Server) withLoginLimit(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := auth.RealIP(r)
		if !s.loginLimiter.Allow(ip) {
			metrics.LoginAttempts.WithLabelValues("limited").Inc()
			writeError(w, http.StatusTooManyRequests, apperror.New("LOGIN_RATE_LIMITED", "登录尝试过于频繁。", "rate limit exceeded for "+ip, "请等待 15 分钟后重试。", true))
			return
		}
		fn(w, r)
	}
}

func (s *Server) withAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.sessions.FromRequest(r); !ok {
			writeError(w, http.StatusUnauthorized, apperror.New("UNAUTHORIZED", "请先登录。", "missing or invalid session", "登录后重试。", false))
			return
		}
		fn(w, r)
	}
}

func (s *Server) withWriteAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := s.sessions.FromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, apperror.New("UNAUTHORIZED", "请先登录。", "missing or invalid session", "登录后重试。", false))
			return
		}
		if r.Header.Get("X-CSRF-Token") != session.CSRF {
			writeError(w, http.StatusForbidden, apperror.New("CSRF_INVALID", "请求校验失败。", "invalid csrf token", "刷新页面后重试。", false))
			return
		}
		fn(w, r)
	}
}

func writeOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(APIResponse{OK: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err *apperror.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{OK: false, Error: err})
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; frame-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'self'")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func allowedUpload(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".pdf":
		return true
	default:
		return false
	}
}

// allowedMIMETypes maps file extensions to their expected MIME type prefixes.
var allowedMIMETypes = map[string][]string{
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".png":  {"image/png"},
	".gif":  {"image/gif"},
	".webp": {"image/webp"},
	".svg":  {"image/svg+xml", "text/xml", "application/xml"},
	".pdf":  {"application/pdf"},
}

// validateUpload checks file extension against an allowlist and verifies the
// MIME type by reading the first 512 bytes of data.
func validateUpload(name string, data []byte) *apperror.AppError {
	ext := strings.ToLower(filepath.Ext(name))
	if !allowedUpload(name) {
		return apperror.New("UPLOAD_TYPE_INVALID", "不支持的文件类型。", name, "仅上传 jpg、png、gif、webp、svg、pdf。", false)
	}
	// Detect MIME type from file content.
	buf := data
	if len(buf) > 512 {
		buf = buf[:512]
	}
	detected := http.DetectContentType(buf)
	expected, ok := allowedMIMETypes[ext]
	if !ok {
		return apperror.New("UPLOAD_TYPE_INVALID", "不支持的文件类型。", name, "仅上传 jpg、png、gif、webp、svg、pdf。", false)
	}
	mimeOK := false
	for _, m := range expected {
		if strings.HasPrefix(detected, m) {
			mimeOK = true
			break
		}
	}
	if !mimeOK {
		return apperror.New("UPLOAD_MIME_MISMATCH", "文件内容与扩展名不匹配。", "detected="+detected+", ext="+ext, "请上传真实的图片或 PDF 文件。", false)
	}
	return nil
}

func readHugoParam(toml, key string) string {
	re := regexp.MustCompile(`(?m)^\s+` + regexp.QuoteMeta(key) + `\s*=\s*"([^"]*)"`)
	m := re.FindStringSubmatch(toml)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func setHugoParam(toml, key, value string) string {
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	re := regexp.MustCompile(`(?m)^(\s+)` + regexp.QuoteMeta(key) + `(\s*=\s*)"[^"]*"`)
	if re.MatchString(toml) {
		return re.ReplaceAllString(toml, "${1}"+key+"${2}\""+escaped+"\"")
	}
	paramsRe := regexp.MustCompile(`(?m)^\[params\]`)
	return paramsRe.ReplaceAllString(toml, "[params]\n  "+key+` = "`+escaped+`"`)
}
