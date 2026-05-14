// Package testutil provides helpers for testing HTTP handlers.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// DoRequest is a convenience for making an HTTP request against the handler.
// It creates a GET request to the given path and attaches the optional cookie.
func DoRequest(t testing.TB, handler http.Handler, method, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// DoWriteRequest makes an authenticated write request with a CSRF token.
// It sets Content-Type: application/json and the X-CSRF-Token header.
// If body is nil, it sends an empty JSON object "{}".
func DoWriteRequest(t testing.TB, handler http.Handler, method, path string, cookie *http.Cookie, csrf string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader([]byte("{}"))
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// ParseAPIResponse unmarshals a JSON response body into a generic map.
// It is a convenience for tests that need to inspect response data.
func ParseAPIResponse(t testing.TB, body []byte) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}
