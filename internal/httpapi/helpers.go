package httpapi

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/preview"
	"blog-studio-web/internal/publish"
	"blog-studio-web/internal/wechat"
)

type serverSnapshot struct {
	adminHash string
	cfg       config.Config
	pub       *publish.Service
	prev      *preview.Service
	wec       *wechat.Service
}

func (s *Server) snap() serverSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return serverSnapshot{adminHash: s.adminHash, cfg: s.cfg, pub: s.pub, prev: s.prev, wec: s.wec}
}

func decodeJSON[T any](r *http.Request, target *T) *apperror.AppError {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return apperror.Wrap("BAD_REQUEST", "请求格式无效。", err, "检查请求体格式并重新提交。", false)
	}
	return nil
}

func parseUploadedFile(r *http.Request, maxBytes int64) ([]byte, string, *apperror.AppError) {
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		return nil, "", apperror.Wrap("UPLOAD_INVALID", "上传请求无效。", err, "检查文件大小和表单字段。", false)
	}
	var file multipart.File
	var header *multipart.FileHeader
	var err error
	if file, header, err = r.FormFile("file"); err != nil {
		return nil, "", apperror.Wrap("UPLOAD_FILE_MISSING", "缺少上传文件。", err, "使用 file 字段上传图片。", false)
	}
	defer file.Close()
	if header.Size > maxBytes {
		return nil, "", apperror.New("UPLOAD_TOO_LARGE", "上传文件过大。", strconv.FormatInt(header.Size, 10), "压缩图片或调整上传上限。", false)
	}
	data, _ := io.ReadAll(io.LimitReader(file, maxBytes+1))
	return data, filepath.Base(header.Filename), nil
}
