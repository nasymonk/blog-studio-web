package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/sanitize"
	"github.com/google/uuid"
)

type Entry struct {
	AuditID       string      `json:"auditId"`
	Time          string      `json:"time"`
	Actor         string      `json:"actor"`
	SiteID        string      `json:"siteId"`
	Slug          string      `json:"slug"`
	Operation     string      `json:"operation"`
	Target        string      `json:"target"`
	Result        string      `json:"result"`
	BackupID      string      `json:"backupId,omitempty"`
	DiffPath      string      `json:"diffPath,omitempty"`
	BuildResult   interface{} `json:"buildResult,omitempty"`
	ChannelResult interface{} `json:"channelResult,omitempty"`
	ErrorCode     string      `json:"errorCode,omitempty"`
	ErrorBrief    string      `json:"errorBrief,omitempty"`
}

type Logger struct {
	paths config.Paths
}

func NewLogger(paths config.Paths) *Logger {
	return &Logger{paths: paths}
}

func NewID(prefix string) string {
	return prefix + "-" + uuid.NewString()
}

func (l *Logger) WriteDiff(auditID, diff string, secrets ...string) (string, *apperror.AppError) {
	if err := os.MkdirAll(l.paths.Diffs, 0700); err != nil {
		return "", apperror.Wrap("AUDIT_DIR_FAILED", "无法创建 diff 目录。", err, "检查 /data/logs 权限。", true)
	}
	diffPath := filepath.Join(l.paths.Diffs, auditID+".diff")
	if err := os.WriteFile(diffPath, []byte(sanitize.Text(diff, secrets...)), 0600); err != nil {
		return "", apperror.Wrap("AUDIT_DIFF_FAILED", "无法写入 diff。", err, "检查磁盘空间。", true)
	}
	return diffPath, nil
}

func (l *Logger) Append(entry Entry, secrets ...string) *apperror.AppError {
	if err := os.MkdirAll(l.paths.Logs, 0700); err != nil {
		return apperror.Wrap("AUDIT_DIR_FAILED", "无法创建审计日志目录。", err, "检查 /data/logs 权限。", true)
	}
	if entry.AuditID == "" {
		entry.AuditID = NewID("audit")
	}
	if entry.Time == "" {
		entry.Time = time.Now().Format(time.RFC3339Nano)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return apperror.Wrap("AUDIT_ENCODE_FAILED", "无法序列化审计日志。", err, "检查审计字段。", false)
	}
	line := sanitize.Text(string(data), secrets...) + "\n"
	file, err := os.OpenFile(filepath.Join(l.paths.Logs, "audit.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return apperror.Wrap("AUDIT_OPEN_FAILED", "无法打开审计日志。", err, "检查 /data/logs 权限。", true)
	}
	defer file.Close()
	if _, err := file.WriteString(line); err != nil {
		return apperror.Wrap("AUDIT_WRITE_FAILED", "无法写入审计日志。", err, "检查磁盘空间。", true)
	}
	return nil
}

func (l *Logger) Recent(limit int) ([]Entry, *apperror.AppError) {
	data, err := os.ReadFile(filepath.Join(l.paths.Logs, "audit.log"))
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, apperror.Wrap("AUDIT_READ_FAILED", "无法读取审计日志。", err, "检查 /data/logs 权限。", true)
	}
	lines := splitLines(string(data))
	out := []Entry{}
	for i := len(lines) - 1; i >= 0 && len(out) < limit; i-- {
		var entry Entry
		if json.Unmarshal([]byte(lines[i]), &entry) == nil {
			out = append(out, entry)
		}
	}
	return out, nil
}

func splitLines(text string) []string {
	raw := []string{}
	start := 0
	for i, ch := range text {
		if ch == '\n' {
			if i > start {
				raw = append(raw, text[start:i])
			}
			start = i + 1
		}
	}
	if start < len(text) {
		raw = append(raw, text[start:])
	}
	return raw
}
