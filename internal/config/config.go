package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"blog-studio-web/internal/apperror"
)

type Paths struct {
	BlogRoot string
	DataRoot string
	Config   string
	Cache    string
	Backups  string
	Logs     string
	Diffs    string
	Preview  string
	Static   string
}

type Config struct {
	BasePath       string         `json:"basePath"`
	Site           SiteConfig     `json:"site"`
	Preview        PreviewConfig  `json:"preview"`
	Wechat         WechatConfig   `json:"wechat"`
	Comments       CommentsConfig `json:"comments"`
	MaxUploadBytes int64          `json:"maxUploadBytes"`
}

type SiteConfig struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Theme        string `json:"theme"`
	BlogRoot     string `json:"blogRoot"`
	ContentRoot  string `json:"contentRoot"`
	PostSection  string `json:"postSection"`
	PostRoot     string `json:"postRoot"`
	BuildCommand string `json:"buildCommand"`
	PublicRoot   string `json:"publicRoot"`
}

type PreviewConfig struct {
	TTLMinutes int `json:"ttlMinutes"`
}

type WechatConfig struct {
	Enabled bool   `json:"enabled"`
	AppID   string `json:"appId,omitempty"`
}

type CommentsConfig struct {
	Enabled   bool   `json:"enabled"`
	TwikooURL string `json:"twikooUrl"`
	AdminURL  string `json:"adminUrl"`
	DataPath  string `json:"dataPath,omitempty"`
}

func DefaultPaths() Paths {
	blogRoot := env("BLOG_STUDIO_BLOG_ROOT", "/blog")
	dataRoot := env("BLOG_STUDIO_DATA_ROOT", "/data")
	return Paths{
		BlogRoot: blogRoot,
		DataRoot: dataRoot,
		Config:   filepath.Join(dataRoot, "config.json"),
		Cache:    filepath.Join(dataRoot, "cache"),
		Backups:  filepath.Join(dataRoot, "backups"),
		Logs:     filepath.Join(dataRoot, "logs"),
		Diffs:    filepath.Join(dataRoot, "logs", "diffs"),
		Preview:  filepath.Join(dataRoot, "preview"),
		Static:   env("BLOG_STUDIO_STATIC_DIR", "./web/dist"),
	}
}

func DefaultConfig(paths Paths) Config {
	basePath := env("BASE_PATH", "/studio")
	return Config{
		BasePath: strings.TrimRight(basePath, "/"),
		Site: SiteConfig{
			ID:           "default",
			Name:         env("BLOG_STUDIO_SITE_NAME", "My Hugo Blog"),
			Theme:        env("BLOG_STUDIO_THEME", "PaperMod"),
			BlogRoot:     paths.BlogRoot,
			ContentRoot:  filepath.Join(paths.BlogRoot, "content"),
			PostSection:  "post",
			PostRoot:     filepath.Join(paths.BlogRoot, "content", "post"),
			BuildCommand: "hugo --minify",
			PublicRoot:   filepath.Join(paths.BlogRoot, "public"),
		},
		Preview:        PreviewConfig{TTLMinutes: envInt("BLOG_STUDIO_PREVIEW_TTL_MINUTES", 120)},
		Wechat:         WechatConfig{Enabled: os.Getenv("WECHAT_APP_ID") != ""},
		Comments:       CommentsConfig{Enabled: true, TwikooURL: "/comment/", AdminURL: "/comment/admin/"},
		MaxUploadBytes: int64(envInt("BLOG_STUDIO_MAX_UPLOAD_MB", 10)) * 1024 * 1024,
	}
}

type Store struct {
	paths Paths
}

func NewStore(paths Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) Load() (Config, *apperror.AppError) {
	data, err := os.ReadFile(s.paths.Config)
	if errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig(s.paths)
		return cfg, nil
	}
	if err != nil {
		return Config{}, apperror.Wrap("CONFIG_READ_FAILED", "无法读取配置文件。", err, "检查 /data/config.json 权限。", true)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, apperror.Wrap("CONFIG_PARSE_FAILED", "配置文件格式无效。", err, "修复 /data/config.json 或删除后重新生成。", false)
	}
	normalize(&cfg, s.paths)
	return cfg, nil
}

func (s *Store) Save(cfg Config) *apperror.AppError {
	normalize(&cfg, s.paths)
	if err := validate(cfg, s.paths); err != nil {
		return apperror.Wrap("CONFIG_INVALID", "配置不合法。", err, "检查博客路径、文章路径和 base path。", false)
	}
	if err := os.MkdirAll(filepath.Dir(s.paths.Config), 0700); err != nil {
		return apperror.Wrap("CONFIG_DIR_FAILED", "无法创建配置目录。", err, "检查 /data 目录权限。", true)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return apperror.Wrap("CONFIG_ENCODE_FAILED", "无法序列化配置。", err, "检查配置字段。", false)
	}
	if err := os.WriteFile(s.paths.Config, data, 0600); err != nil {
		return apperror.Wrap("CONFIG_WRITE_FAILED", "无法写入配置文件。", err, "检查 /data 目录权限。", true)
	}
	return nil
}

func normalize(cfg *Config, paths Paths) {
	if cfg.BasePath == "" {
		cfg.BasePath = "/studio"
	}
	cfg.BasePath = strings.TrimRight(cfg.BasePath, "/")
	if cfg.Site.ID == "" {
		cfg.Site.ID = "default"
	}
	if cfg.Site.BlogRoot == "" {
		cfg.Site.BlogRoot = paths.BlogRoot
	}
	if cfg.Site.ContentRoot == "" {
		cfg.Site.ContentRoot = filepath.Join(cfg.Site.BlogRoot, "content")
	}
	if cfg.Site.PostSection == "" {
		cfg.Site.PostSection = "post"
	}
	if cfg.Site.PostRoot == "" {
		cfg.Site.PostRoot = filepath.Join(cfg.Site.ContentRoot, cfg.Site.PostSection)
	}
	if cfg.Site.PublicRoot == "" {
		cfg.Site.PublicRoot = filepath.Join(cfg.Site.BlogRoot, "public")
	}
	if cfg.Site.BuildCommand == "" {
		cfg.Site.BuildCommand = "hugo --minify"
	}
	if cfg.Preview.TTLMinutes <= 0 {
		cfg.Preview.TTLMinutes = 120
	}
	if cfg.MaxUploadBytes <= 0 {
		cfg.MaxUploadBytes = 10 * 1024 * 1024
	}
}

func validate(cfg Config, paths Paths) error {
	if !strings.HasPrefix(cfg.BasePath, "/") {
		return errors.New("basePath must start with /")
	}
	for _, p := range []string{cfg.Site.BlogRoot, cfg.Site.ContentRoot, cfg.Site.PostRoot, cfg.Site.PublicRoot} {
		if !inside(paths.BlogRoot, p) {
			return errors.New("site paths must stay inside blog root")
		}
	}
	if cfg.Site.BuildCommand != "hugo --minify" && cfg.Site.BuildCommand != "hugo" {
		return errors.New("buildCommand only supports hugo or hugo --minify in V1")
	}
	return nil
}

func inside(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) || targetAbs == rootAbs
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
