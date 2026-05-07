package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/backup"
	"blog-studio-web/internal/comments"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/preview"
	"blog-studio-web/internal/publish"
	"blog-studio-web/internal/storage"
	"blog-studio-web/internal/wechat"
)

type APIResponse struct {
	OK    bool               `json:"ok"`
	Data  interface{}        `json:"data,omitempty"`
	Error *apperror.AppError `json:"error,omitempty"`
}

type Server struct {
	paths        config.Paths
	cfgStore     *config.Store
	cfg          config.Config
	adminHash    string
	sessions     *auth.Store
	audit        *audit.Logger
	publisher    *publish.Service
	preview      *preview.Service
	wechat       *wechat.Service
	comments     *comments.Service
	staticPrefix string
}

func New(paths config.Paths, cfg config.Config, cfgStore *config.Store, adminHash string, sessions *auth.Store) *Server {
	auditLogger := audit.NewLogger(paths)
	backupStore := backup.NewStore(paths, 5)
	return &Server{
		paths:     paths,
		cfgStore:  cfgStore,
		cfg:       cfg,
		adminHash: adminHash,
		sessions:  sessions,
		audit:     auditLogger,
		publisher: publish.NewService(paths, cfg, backupStore, auditLogger),
		preview:   preview.NewService(paths, cfg),
		wechat:    wechat.NewService(paths, cfg, auditLogger),
		comments:  comments.NewService(commentDataPath(cfg), cfg.Comments.AdminURL),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	api := http.NewServeMux()
	api.HandleFunc("POST /auth/login", s.login)
	api.HandleFunc("POST /auth/logout", s.withAuth(s.logout))
	api.HandleFunc("GET /session", s.session)
	api.HandleFunc("GET /posts", s.withAuth(s.listPosts))
	api.HandleFunc("GET /posts/", s.withAuth(s.postRouter))
	api.HandleFunc("PUT /posts/", s.withWriteAuth(s.postRouter))
	api.HandleFunc("POST /posts/", s.withWriteAuth(s.postRouter))
	api.HandleFunc("GET /comments/summary", s.withAuth(s.commentSummary))
	api.HandleFunc("GET /comments/recent", s.withAuth(s.commentRecent))
	api.HandleFunc("GET /health", s.withAuth(s.health))
	api.HandleFunc("GET /audit", s.withAuth(s.auditRecent))
	api.HandleFunc("GET /config", s.withAuth(s.getConfig))
	api.HandleFunc("PUT /config", s.withWriteAuth(s.putConfig))

	base := strings.TrimRight(s.cfg.BasePath, "/")
	mux.Handle(base+"/api/", http.StripPrefix(base+"/api", api))
	mux.Handle(base+"/preview/", http.StripPrefix(base+"/preview/", http.FileServer(http.Dir(filepath.Join(s.paths.Preview, "public")))))
	mux.Handle(base+"/", s.spaHandler(base))
	if base != "" {
		mux.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, base+"/", http.StatusMovedPermanently)
		})
	}
	return secureHeaders(mux)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交登录表单。", false))
		return
	}
	if !auth.VerifyPassword(req.Password, s.adminHash) {
		writeError(w, http.StatusUnauthorized, apperror.New("LOGIN_FAILED", "密码错误。", "invalid admin password", "检查管理员密码。", false))
		return
	}
	session := s.sessions.Create(w)
	writeOK(w, map[string]interface{}{"authenticated": true, "csrfToken": session.CSRF})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.sessions.Destroy(w, r)
	writeOK(w, map[string]bool{"authenticated": false})
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	if session, ok := s.sessions.FromRequest(r); ok {
		writeOK(w, map[string]interface{}{"authenticated": true, "csrfToken": session.CSRF})
		return
	}
	writeOK(w, map[string]bool{"authenticated": false})
}

func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := s.publisher.ListPosts()
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
	post, err := s.publisher.LoadPost(slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeOK(w, post)
}

func (s *Server) saveDraft(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交草稿。", false))
		return
	}
	draft.Slug = slug
	if err := s.publisher.SaveDraft(draft); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeOK(w, map[string]bool{"saved": true})
}

func (s *Server) uploadAsset(w http.ResponseWriter, r *http.Request, slug string) {
	if err := r.ParseMultipartForm(s.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("UPLOAD_INVALID", "上传请求无效。", err, "检查文件大小和表单字段。", false))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("UPLOAD_FILE_MISSING", "缺少上传文件。", err, "使用 file 字段上传图片。", false))
		return
	}
	defer file.Close()
	if header.Size > s.cfg.MaxUploadBytes {
		writeError(w, http.StatusRequestEntityTooLarge, apperror.New("UPLOAD_TOO_LARGE", "上传文件过大。", strconv.FormatInt(header.Size, 10), "压缩图片或调整上传上限。", false))
		return
	}
	postDir, appErr := storage.PostDir(s.cfg.Site.PostRoot, slug)
	if appErr != nil {
		writeError(w, http.StatusBadRequest, appErr)
		return
	}
	if !allowedUpload(header.Filename) {
		writeError(w, http.StatusBadRequest, apperror.New("UPLOAD_TYPE_INVALID", "不支持的文件类型。", header.Filename, "仅上传 jpg、png、gif、webp、svg。", false))
		return
	}
	data, _ := io.ReadAll(io.LimitReader(file, s.cfg.MaxUploadBytes+1))
	target, safeErr := storage.SafeJoin(postDir, filepath.Base(header.Filename))
	if safeErr != nil {
		writeError(w, http.StatusBadRequest, safeErr)
		return
	}
	if err := storage.AtomicWriteFile(target, data, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, content.Asset{Name: filepath.Base(header.Filename), Size: int64(len(data))})
}

func (s *Server) publishBlog(w http.ResponseWriter, r *http.Request, slug string) {
	var req publish.BlogPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交发布请求。", false))
		return
	}
	req.Slug = slug
	req.Draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	result := s.publisher.PublishBlog(ctx, req)
	writeOK(w, result)
}

func (s *Server) publishWechatDraft(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交公众号发布请求。", false))
		return
	}
	draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
	defer cancel()
	writeOK(w, s.wechat.PublishDraft(ctx, draft))
}

func (s *Server) createPreview(w http.ResponseWriter, r *http.Request, slug string) {
	var draft content.PostDraft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交预览请求。", false))
		return
	}
	draft.Slug = slug
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	_ = s.preview.Cleanup()
	writeOK(w, s.preview.Create(ctx, draft))
}

func (s *Server) rollback(w http.ResponseWriter, r *http.Request, slug string) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	writeOK(w, s.publisher.Rollback(ctx, slug))
}

func (s *Server) commentSummary(w http.ResponseWriter, r *http.Request) {
	data, err := s.comments.Summary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, map[string]interface{}{"adminUrl": s.comments.AdminURL(), "items": data})
}

func (s *Server) commentRecent(w http.ResponseWriter, r *http.Request) {
	data, err := s.comments.Recent(20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, map[string]interface{}{"adminUrl": s.comments.AdminURL(), "items": data})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	checks := []map[string]string{}
	add := func(name, status, message, detail, suggestion string) {
		checks = append(checks, map[string]string{"name": name, "status": status, "message": message, "technicalDetail": detail, "suggestion": suggestion})
	}
	if _, err := os.Stat(s.cfg.Site.BlogRoot); err != nil {
		add("blog-root", "error", "博客目录不可访问。", err.Error(), "检查 Docker 挂载 /blog。")
	} else {
		add("blog-root", "ok", "博客目录可访问。", s.cfg.Site.BlogRoot, "")
	}
	if _, err := os.Stat(s.cfg.Site.PostRoot); err != nil {
		add("post-root", "error", "文章目录不可访问。", err.Error(), "检查 content/post 路径。")
	} else {
		add("post-root", "ok", "文章目录可访问。", s.cfg.Site.PostRoot, "")
	}
	if err := os.MkdirAll(s.paths.DataRoot, 0700); err != nil {
		add("data-root", "error", "数据目录不可写。", err.Error(), "检查 /data 挂载权限。")
	} else {
		add("data-root", "ok", "数据目录可写。", s.paths.DataRoot, "")
	}
	status := "ok"
	for _, check := range checks {
		if check["status"] == "error" {
			status = "error"
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
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "重新提交配置。", false))
		return
	}
	cfg.Wechat.AppID = ""
	if err := s.cfgStore.Save(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.cfg = cfg
	writeOK(w, cfg)
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
		next.ServeHTTP(w, r)
	})
}

func allowedUpload(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return true
	default:
		return false
	}
}

func commentDataPath(cfg config.Config) string {
	if cfg.Comments.DataPath != "" {
		return cfg.Comments.DataPath
	}
	if value := os.Getenv("TWIKOO_DATA_PATH"); value != "" {
		return value
	}
	return ""
}

var _ = errors.New
