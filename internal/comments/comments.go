package comments

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"blog-studio-web/internal/apperror"
)

type Summary struct {
	URL      string `json:"url"`
	Count    int    `json:"count"`
	LatestAt string `json:"latestAt"`
}

type Recent struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Nick    string `json:"nick"`
	Comment string `json:"comment"`
	Created string `json:"created"`
}

type Service struct {
	dataPath string
	adminURL string
}

func NewService(dataPath, adminURL string) *Service {
	return &Service{dataPath: dataPath, adminURL: adminURL}
}

func (s *Service) AdminURL() string {
	return s.adminURL
}

func (s *Service) Summary() ([]Summary, *apperror.AppError) {
	rows, err := s.readRows()
	if err != nil {
		return nil, err
	}
	byURL := map[string]Summary{}
	for _, row := range rows {
		urlValue := stringField(row, "url")
		if urlValue == "" {
			continue
		}
		item := byURL[urlValue]
		item.URL = urlValue
		item.Count++
		created := createdAt(row)
		if created > item.LatestAt {
			item.LatestAt = created
		}
		byURL[urlValue] = item
	}
	out := []Summary{}
	for _, item := range byURL {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LatestAt > out[j].LatestAt })
	return out, nil
}

func (s *Service) Recent(limit int) ([]Recent, *apperror.AppError) {
	rows, err := s.readRows()
	if err != nil {
		return nil, err
	}
	out := []Recent{}
	for _, row := range rows {
		out = append(out, Recent{ID: stringField(row, "_id"), URL: stringField(row, "url"), Nick: stringField(row, "nick"), Comment: stringField(row, "comment"), Created: createdAt(row)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created > out[j].Created })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Service) readRows() ([]map[string]interface{}, *apperror.AppError) {
	if s.dataPath == "" {
		return []map[string]interface{}{}, nil
	}
	data, err := os.ReadFile(s.dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, apperror.Wrap("COMMENTS_READ_FAILED", "无法读取 Twikoo 数据。", err, "检查 Twikoo 数据挂载路径；Blog Studio 不会写入该文件。", true)
	}
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, apperror.Wrap("COMMENTS_PARSE_FAILED", "无法解析 Twikoo 数据。", err, "检查 Twikoo 数据格式是否变化。", false)
	}
	rows := extractCommentRows(raw)
	seen := map[string]bool{}
	out := []map[string]interface{}{}
	for _, row := range rows {
		key := stringField(row, "_id")
		if key == "" {
			key = stringField(row, "url") + "|" + stringField(row, "comment") + "|" + createdAt(row)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, row)
	}
	return out, nil
}

func extractCommentRows(value interface{}) []map[string]interface{} {
	out := []map[string]interface{}{}
	switch typed := value.(type) {
	case []interface{}:
		for _, item := range typed {
			if row, ok := item.(map[string]interface{}); ok {
				out = append(out, row)
			}
		}
	case map[string]interface{}:
		for key, item := range typed {
			if key == "comment" || key == "comments" || key == "Comment" {
				out = append(out, extractCommentRows(item)...)
			}
			if arr, ok := item.([]interface{}); ok {
				for _, entry := range arr {
					if row, ok := entry.(map[string]interface{}); ok && stringField(row, "comment") != "" {
						out = append(out, row)
					}
				}
			}
		}
	}
	return out
}

func stringField(row map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key]; ok {
			switch typed := value.(type) {
			case string:
				return typed
			case float64:
				return time.UnixMilli(int64(typed)).Format(time.RFC3339)
			}
		}
	}
	return ""
}

func createdAt(row map[string]interface{}) string {
	for _, key := range []string{"created", "createdAt", "insertedAt", "created_at"} {
		if value := stringField(row, key); value != "" {
			return value
		}
	}
	return ""
}
