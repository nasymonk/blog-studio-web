package apperror

import (
	"fmt"
	"testing"
)

func TestNew_SetsAllFields(t *testing.T) {
	err := New("NOT_FOUND", "资源未找到", "slug article-xyz not in index", "请检查文章路径是否正确", false)

	if err.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want %q", err.Code, "NOT_FOUND")
	}
	if err.Message != "资源未找到" {
		t.Errorf("Message = %q, want %q", err.Message, "资源未找到")
	}
	if err.TechnicalDetail != "slug article-xyz not in index" {
		t.Errorf("TechnicalDetail = %q, want %q", err.TechnicalDetail, "slug article-xyz not in index")
	}
	if err.Suggestion != "请检查文章路径是否正确" {
		t.Errorf("Suggestion = %q, want %q", err.Suggestion, "请检查文章路径是否正确")
	}
	if err.Retryable != false {
		t.Errorf("Retryable = %t, want false", err.Retryable)
	}
}

func TestNew_RetryableTrue(t *testing.T) {
	err := New("TIMEOUT", "请求超时", "connection reset by peer", "请稍后重试", true)

	if !err.Retryable {
		t.Error("Retryable should be true")
	}
}

func TestError_Format(t *testing.T) {
	err := New("NOT_FOUND", "资源未找到", "", "", false)
	got := err.Error()
	want := "NOT_FOUND: 资源未找到"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_NilReceiver(t *testing.T) {
	var err *AppError
	got := err.Error()
	if got != "" {
		t.Errorf("Error() on nil receiver = %q, want empty string", got)
	}
}

func TestError_NilPointer(t *testing.T) {
	err := (*AppError)(nil)
	got := err.Error()
	if got != "" {
		t.Errorf("Error() on nil pointer = %q, want empty string", got)
	}
}

func TestWrap_NonNilError(t *testing.T) {
	original := fmt.Errorf("underlying db error")
	err := Wrap("DB_FAIL", "数据库操作失败", original, "检查数据库连接", true)

	if err.Code != "DB_FAIL" {
		t.Errorf("Code = %q, want %q", err.Code, "DB_FAIL")
	}
	if err.Message != "数据库操作失败" {
		t.Errorf("Message = %q, want %q", err.Message, "数据库操作失败")
	}
	if err.TechnicalDetail != "underlying db error" {
		t.Errorf("TechnicalDetail = %q, want %q", err.TechnicalDetail, "underlying db error")
	}
	if err.Suggestion != "检查数据库连接" {
		t.Errorf("Suggestion = %q, want %q", err.Suggestion, "检查数据库连接")
	}
	if !err.Retryable {
		t.Error("Retryable should be true")
	}
}

func TestWrap_NilError(t *testing.T) {
	err := Wrap("OK", "操作成功", nil, "", false)

	if err.TechnicalDetail != "" {
		t.Errorf("TechnicalDetail with nil error = %q, want empty string", err.TechnicalDetail)
	}
	if err.Error() != "OK: 操作成功" {
		t.Errorf("Error() = %q, want %q", err.Error(), "OK: 操作成功")
	}
}

func TestWrap_EmptyCode(t *testing.T) {
	err := Wrap("", "no code", nil, "", false)
	if err.Code != "" {
		t.Errorf("Code = %q, want empty string", err.Code)
	}
	if err.Error() != ": no code" {
		t.Errorf("Error() = %q, want %q", err.Error(), ": no code")
	}
}

func TestNew_SupportsEmptyDetailAndSuggestion(t *testing.T) {
	err := New("HELLO", "world", "", "", false)
	if err.TechnicalDetail != "" {
		t.Errorf("TechnicalDetail = %q, want empty", err.TechnicalDetail)
	}
	if err.Suggestion != "" {
		t.Errorf("Suggestion = %q, want empty", err.Suggestion)
	}
}

func TestNew_AndNewAreIndependent(t *testing.T) {
	a := New("A", "first", "detail a", "suggestion a", false)
	b := New("B", "second", "detail b", "suggestion b", true)

	if a.Code == b.Code {
		t.Error("two New() calls should produce independent AppError values")
	}
	if a.Error() != "A: first" {
		t.Errorf("a.Error() = %q", a.Error())
	}
	if b.Error() != "B: second" {
		t.Errorf("b.Error() = %q", b.Error())
	}
}
