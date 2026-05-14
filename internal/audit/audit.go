package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blog-studio-web/internal/apperror"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/sanitize"
	"github.com/google/uuid"
)

const MaxAuditLines = 5000

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
	BuildResult   any `json:"buildResult,omitempty"`
	ChannelResult any `json:"channelResult,omitempty"`
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

func (l *Logger) Rotate(maxLines int) *apperror.AppError {
	logPath := filepath.Join(l.paths.Logs, "audit.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperror.Wrap("AUDIT_ROTATE_READ", "无法读取审计日志。", err, "检查 /data/logs 权限。", true)
	}
	lines := splitLines(string(data))
	if len(lines) <= maxLines {
		return nil
	}
	kept := lines[len(lines)-maxLines:]
	content := strings.Join(kept, "\n") + "\n"
	tmp := logPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		return apperror.Wrap("AUDIT_ROTATE_WRITE", "无法写入轮转后日志。", err, "检查磁盘空间。", true)
	}
	if err := os.Rename(tmp, logPath); err != nil {
		_ = os.Remove(tmp)
		return apperror.Wrap("AUDIT_ROTATE_RENAME", "无法替换审计日志。", err, "检查文件权限。", true)
	}
	return nil
}

func (l *Logger) PruneDiffs() *apperror.AppError {
	logPath := filepath.Join(l.paths.Logs, "audit.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperror.Wrap("AUDIT_PRUNE_READ", "无法读取审计日志。", err, "检查权限。", true)
	}
	referenced := map[string]struct{}{}
	for _, line := range splitLines(string(data)) {
		var entry Entry
		if json.Unmarshal([]byte(line), &entry) == nil && entry.DiffPath != "" {
			referenced[filepath.Base(entry.DiffPath)] = struct{}{}
		}
	}
	entries, err := os.ReadDir(l.paths.Diffs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperror.Wrap("AUDIT_PRUNE_DIR", "无法读取 diffs 目录。", err, "检查权限。", true)
	}
	for _, entry := range entries {
		if _, keep := referenced[entry.Name()]; !keep {
			_ = os.Remove(filepath.Join(l.paths.Diffs, entry.Name()))
		}
	}
	return nil
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
