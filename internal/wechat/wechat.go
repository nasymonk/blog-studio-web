package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/audit"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/content"
	"blog-studio-web/internal/publish"
)

type Service struct {
	cfg     config.Config
	paths   config.Paths
	audit   *audit.Logger
	content *content.Service
	client  *http.Client
}

type DraftResult struct {
	MediaID string `json:"mediaId,omitempty"`
	Message string `json:"message"`
}

func NewService(paths config.Paths, cfg config.Config, auditLogger *audit.Logger) *Service {
	return &Service{paths: paths, cfg: cfg, audit: auditLogger, content: content.NewService(), client: &http.Client{Timeout: 20 * time.Second}}
}

func (s *Service) PublishDraft(ctx context.Context, draft content.PostDraft) publish.Result {
	auditID := audit.NewID("wechat")
	result := publish.Result{Target: "wechat_draft", Status: "failed", AuditID: auditID}
	appID := os.Getenv("WECHAT_APP_ID")
	appSecret := os.Getenv("WECHAT_APP_SECRET")
	if appID == "" || appSecret == "" {
		result.Error = apperror.New("WECHAT_CREDENTIALS_MISSING", "缺少微信公众号凭据。", "WECHAT_APP_ID or WECHAT_APP_SECRET is empty", "在服务环境变量中配置公众号凭据后重试。", false)
		s.appendAudit(draft.Slug, result, appSecret)
		return result
	}
	token, err := s.accessToken(ctx, appID, appSecret)
	if err != nil {
		result.Error = err
		s.appendAudit(draft.Slug, result, appSecret)
		return result
	}
	thumbID, thumbErr := s.coverMediaID(ctx, token, draft)
	if thumbErr != nil {
		result.Error = thumbErr
		s.appendAudit(draft.Slug, result, appSecret)
		return result
	}
	html := markdownToWechatHTML(draft.Body)
	mediaID, addErr := s.addDraft(ctx, token, draft, html, thumbID)
	if addErr != nil {
		result.Error = addErr
		s.appendAudit(draft.Slug, result, appSecret)
		return result
	}
	result.Status = "success"
	result.ChannelResult = DraftResult{MediaID: mediaID, Message: "已保存到公众号草稿箱，请在微信后台预览并确认发布。"}
	s.appendAudit(draft.Slug, result, appSecret)
	return result
}

func (s *Service) accessToken(ctx context.Context, appID, appSecret string) (string, *apperror.AppError) {
	endpoint := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + url.QueryEscape(appID) + "&secret=" + url.QueryEscape(appSecret)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", apperror.New("WECHAT_TOKEN_FAILED", "无法获取微信 access token。", fmt.Sprintf("访问微信 API (%s) 失败, AppSecret 已从错误详情中移除", req.URL.Host), "检查 AppID/AppSecret、IP 白名单和网络。", true)
	}
	defer resp.Body.Close()
	var payload struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload.AccessToken == "" {
		return "", apperror.New("WECHAT_TOKEN_REJECTED", "微信拒绝 access token 请求。", fmt.Sprintf("%d %s", payload.ErrCode, payload.ErrMsg), "检查公众号权限、IP 白名单和凭据。", false)
	}
	return payload.AccessToken, nil
}

func (s *Service) coverMediaID(ctx context.Context, token string, draft content.PostDraft) (string, *apperror.AppError) {
	if draft.FrontMatter.Image == "" {
		return "", apperror.New("WECHAT_COVER_MISSING", "公众号草稿需要封面图。", "front matter image is empty", "在文章 front matter 添加 image，并将图片放入 Page Bundle。", false)
	}
	// Validate image path — reject traversal
	image := filepath.Clean(draft.FrontMatter.Image)
	if filepath.IsAbs(image) || strings.HasPrefix(image, "..") {
		return "", apperror.New("INVALID_IMAGE_PATH", "封面图片路径无效", "image path contains path traversal", "请使用文章目录内的图片作为封面。", false)
	}
	postDir := filepath.Join(s.cfg.Site.PostRoot, draft.Slug)
	coverPath := filepath.Join(postDir, image)
	// Verify coverPath stays within postDir
	if !strings.HasPrefix(coverPath, filepath.Clean(postDir)+string(filepath.Separator)) && coverPath != filepath.Clean(postDir) {
		return "", apperror.New("PATH_TRAVERSAL", "封面图片路径越界", "resolved path escapes post directory", "请使用文章目录内的图片作为封面。", false)
	}
	return s.uploadPermanentImage(ctx, token, coverPath)
}

func (s *Service) uploadPermanentImage(ctx context.Context, token, pathName string) (string, *apperror.AppError) {
	file, err := os.Open(pathName)
	if err != nil {
		return "", apperror.Wrap("WECHAT_COVER_READ_FAILED", "无法读取公众号封面图。", err, "检查封面图是否存在。", false)
	}
	defer file.Close()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("media", filepath.Base(pathName))
	if err != nil {
		return "", apperror.Wrap("WECHAT_UPLOAD_FORM_FAILED", "无法创建微信上传表单。", err, "重试上传。", true)
	}
	_, _ = io.Copy(part, file)
	_ = writer.Close()
	endpoint := "https://api.weixin.qq.com/cgi-bin/material/add_material?type=image&access_token=" + url.QueryEscape(token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := s.client.Do(req)
	if err != nil {
		return "", apperror.Wrap("WECHAT_COVER_UPLOAD_FAILED", "上传公众号封面图失败。", err, "检查网络和素材权限。", true)
	}
	defer resp.Body.Close()
	var payload struct {
		MediaID string `json:"media_id"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload.MediaID == "" {
		return "", apperror.New("WECHAT_COVER_UPLOAD_REJECTED", "微信拒绝封面图上传。", fmt.Sprintf("%d %s", payload.ErrCode, payload.ErrMsg), "检查图片格式、大小和公众号素材权限。", false)
	}
	return payload.MediaID, nil
}

func (s *Service) addDraft(ctx context.Context, token string, draft content.PostDraft, html string, thumbID string) (string, *apperror.AppError) {
	payload := map[string]any{
		"articles": []map[string]any{{
			"title":                 draft.FrontMatter.Title,
			"author":                "",
			"digest":                draft.FrontMatter.Description,
			"content":               html,
			"thumb_media_id":        thumbID,
			"need_open_comment":     0,
			"only_fans_can_comment": 0,
		}},
	}
	data, _ := json.Marshal(payload)
	endpoint := "https://api.weixin.qq.com/cgi-bin/draft/add?access_token=" + url.QueryEscape(token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", apperror.Wrap("WECHAT_DRAFT_FAILED", "保存公众号草稿失败。", err, "检查网络和草稿箱权限。", true)
	}
	defer resp.Body.Close()
	var out struct {
		MediaID string `json:"media_id"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.MediaID == "" {
		return "", apperror.New("WECHAT_DRAFT_REJECTED", "微信拒绝保存草稿。", fmt.Sprintf("%d %s", out.ErrCode, out.ErrMsg), "检查公众号权限、HTML 内容和封面素材。", false)
	}
	return out.MediaID, nil
}

func markdownToWechatHTML(body string) string {
	lines := strings.Split(body, "\n")
	out := strings.Builder{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### "):
			out.WriteString("<h3>" + htmlEscape(strings.TrimPrefix(trimmed, "### ")) + "</h3>")
		case strings.HasPrefix(trimmed, "## "):
			out.WriteString("<h2>" + htmlEscape(strings.TrimPrefix(trimmed, "## ")) + "</h2>")
		case trimmed == "":
			out.WriteString("<p><br></p>")
		default:
			out.WriteString("<p>" + htmlEscape(trimmed) + "</p>")
		}
	}
	return out.String()
}

func htmlEscape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	return value
}

func (s *Service) appendAudit(slug string, result publish.Result, secrets ...string) {
	entry := audit.Entry{AuditID: result.AuditID, Actor: "admin", SiteID: s.cfg.Site.ID, Slug: slug, Operation: "publish", Target: result.Target, Result: result.Status, ChannelResult: result.ChannelResult}
	if result.Error != nil {
		entry.ErrorCode = result.Error.Code
		entry.ErrorBrief = result.Error.Message
	}
	_ = s.audit.Append(entry, secrets...)
}
