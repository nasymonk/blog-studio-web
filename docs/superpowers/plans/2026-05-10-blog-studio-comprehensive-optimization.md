# Blog Studio Web — Comprehensive Optimization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform blog-studio-web from a working prototype into a production-grade blog management platform with comprehensive testing, security hardening, performance optimization, accessible UI, and advanced features.

**Architecture:** Five independent tracks (Testing, Performance, Security, UI/UX, Features) organized into 8 phases. Each phase produces deployable software. Phases 1-4 are foundational (must be done in order). Phases 5-8 are independent and can be parallelized.

**Tech Stack:** Go 1.23 (backend), Vue 3.5 + TypeScript + Vite 6 (frontend), CodeMirror 6 (editor), shadcn-vue + Tailwind v4 (UI), Vitest + vue-test-utils (frontend tests), Go testing (backend tests), Docker + nginx (deployment)

**Repository:** `/Users/xiang/Desktop/personal-server/blog-studio-web/`

---

## Phase 1: Testing Foundation (Critical)

Every subsequent phase depends on tests to prevent regressions. Start here.

### Task 1.1: Config Package Tests

**Files:**
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write config validation tests**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalize_Defaults(t *testing.T) {
	c := Config{}
	c.normalize("/data")
	if c.BasePath != "/data" {
		t.Errorf("expected basePath /data, got %s", c.BasePath)
	}
	if c.Site.PostSection != "post" {
		t.Errorf("expected postSection 'post', got %s", c.Site.PostSection)
	}
	if c.Preview.TTLMinutes != 120 {
		t.Errorf("expected TTLMinutes 120, got %d", c.Preview.TTLMinutes)
	}
}

func TestValidate_RejectsPathTraversal(t *testing.T) {
	c := Config{
		Site: SiteConfig{
			BlogRoot:    "/blog",
			ContentRoot: "/blog/content",
			PostRoot:    "/blog/content/post",
			PublicRoot:  "/blog/public",
		},
	}
	// Set a path outside blog root
	c.Site.BlogRoot = "/other"
	if err := c.validate(); err == nil {
		t.Error("expected validation error for path outside blog root")
	}
}

func TestValidate_AcceptsValidPaths(t *testing.T) {
	c := Config{
		Site: SiteConfig{
			BlogRoot:    "/blog",
			ContentRoot: "/blog/content",
			PostRoot:    "/blog/content/post",
			PublicRoot:  "/blog/public",
		},
	}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := Config{
		Site: SiteConfig{
			Name:    "Test Blog",
			BlogRoot: "/blog",
			ContentRoot: "/blog/content",
			PostRoot:    "/blog/content/post",
			PostSection: "post",
			PublicRoot:  "/blog/public",
		},
		Preview: PreviewConfig{TTLMinutes: 60},
	}
	original.normalize("/data")

	if err := Save(path, &original); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path, "/data")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Site.Name != "Test Blog" {
		t.Errorf("expected name 'Test Blog', got %s", loaded.Site.Name)
	}
	if loaded.Preview.TTLMinutes != 60 {
		t.Errorf("expected TTLMinutes 60, got %d", loaded.Preview.TTLMinutes)
	}
}

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	c, err := Load("/nonexistent/config.json", "/data")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if c.BasePath != "/data" {
		t.Errorf("expected defaults applied, got basePath %s", c.BasePath)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/config/ -v -count=1`
Expected: Some tests fail because `normalize` and `validate` are not exported or have different signatures.

- [ ] **Step 3: Fix test compilation issues**

Check the actual function signatures in `internal/config/config.go` and adjust tests to match. The key functions are `normalize()` (private, called from `Load`) and `validate()` (private). Tests should go through `Load()` and `Save()` to test the full pipeline.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/config/ -v -count=1`
Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/xiang/Desktop/personal-server/blog-studio-web
git add internal/config/config_test.go
git commit -m "test: add config package tests (normalize, validate, load/save round-trip)"
```

---

### Task 1.2: Publish Package Tests

**Files:**
- Create: `internal/publish/service_test.go`

- [ ] **Step 1: Write publish service tests**

```go
// internal/publish/service_test.go
package publish

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nasymonk/blog-studio-web/internal/storage"
)

func setupTestService(t *testing.T) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	blogDir := filepath.Join(root, "blog")
	dataDir := filepath.Join(root, "data")
	os.MkdirAll(filepath.Join(blogDir, "content", "post"), 0o755)
	os.MkdirAll(dataDir, 0o755)

	paths := storage.Paths{
		BlogRoot:   blogDir,
		ContentDir: filepath.Join(blogDir, "content"),
		PostRoot:   filepath.Join(blogDir, "content", "post"),
		PublicDir:  filepath.Join(blogDir, "public"),
		DataDir:    dataDir,
	}
	svc := NewService(paths)
	return svc, blogDir
}

func TestSaveDraft_CreatesCacheFile(t *testing.T) {
	svc, _ := setupTestService(t)
	draft := PostDraft{
		Slug: "test-post",
		FrontMatter: FrontMatter{
			Title: "Test Post",
			Date:  "2026-01-01",
			Tags:  []string{"test"},
		},
		Body: "Hello world",
	}
	if err := svc.SaveDraft(draft); err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}

	// Verify file exists in cache
	cachePath := filepath.Join(svc.paths.DataDir, "cache", "posts", "test-post", "index.md")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("expected draft file in cache directory")
	}
}

func TestLoadPost_FromCache(t *testing.T) {
	svc, _ := setupTestService(t)
	draft := PostDraft{
		Slug: "cached-post",
		FrontMatter: FrontMatter{
			Title: "Cached",
			Date:  "2026-01-01",
		},
		Body: "cached content",
	}
	svc.SaveDraft(draft)

	loaded, err := svc.LoadPost("cached-post")
	if err != nil {
		t.Fatalf("LoadPost failed: %v", err)
	}
	if loaded.Body != "cached content" {
		t.Errorf("expected body 'cached content', got '%s'", loaded.Body)
	}
}

func TestLoadPost_NotFound(t *testing.T) {
	svc, _ := setupTestService(t)
	_, err := svc.LoadPost("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent post")
	}
}

func TestPublishBlog_WritesToContentDir(t *testing.T) {
	svc, blogDir := setupTestService(t)
	draft := PostDraft{
		Slug: "pub-test",
		FrontMatter: FrontMatter{
			Title: "Published",
			Date:  "2026-01-01",
		},
		Body: "published content",
	}
	result, err := svc.PublishBlog(draft, false)
	if err != nil {
		t.Fatalf("PublishBlog failed: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result.Status)
	}

	// Verify file in content dir
	contentPath := filepath.Join(blogDir, "content", "post", "pub-test", "index.md")
	data, err := os.ReadFile(contentPath)
	if err != nil {
		t.Fatalf("published file not found: %v", err)
	}
	if len(data) == 0 {
		t.Error("published file is empty")
	}
}
```

- [ ] **Step 2: Run tests, fix compilation**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/publish/ -v -count=1`
Adjust struct names and field names to match actual `publish` package types.

- [ ] **Step 3: Verify all tests pass**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/publish/ -v -count=1`
Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/publish/service_test.go
git commit -m "test: add publish package tests (save draft, load post, publish blog)"
```

---

### Task 1.3: Trash Package Tests

**Files:**
- Create: `internal/trash/trash_test.go`

- [ ] **Step 1: Write trash service tests**

```go
// internal/trash/trash_test.go
package trash

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTrash(t *testing.T) (*Service, string, string) {
	t.Helper()
	root := t.TempDir()
	postDir := filepath.Join(root, "content", "post")
	trashDir := filepath.Join(root, "data", "trash")
	os.MkdirAll(postDir, 0o755)
	os.MkdirAll(trashDir, 0o755)

	// Create a test post
	postPath := filepath.Join(postDir, "test-post", "index.md")
	os.MkdirAll(filepath.Dir(postPath), 0o755)
	os.WriteFile(postPath, []byte("---\ntitle: Test\n---\nBody"), 0o644)

	svc := NewService(postDir, trashDir)
	return svc, postDir, trashDir
}

func TestMoveToTrash(t *testing.T) {
	svc, postDir, trashDir := setupTrash(t)
	if err := svc.MoveToTrash("test-post"); err != nil {
		t.Fatalf("MoveToTrash failed: %v", err)
	}
	// Original should be gone
	if _, err := os.Stat(filepath.Join(postDir, "test-post")); !os.IsNotExist(err) {
		t.Error("expected original post to be removed")
	}
	// Trash should have it
	entries, _ := os.ReadDir(trashDir)
	if len(entries) == 0 {
		t.Error("expected post in trash directory")
	}
}

func TestRestore(t *testing.T) {
	svc, postDir, _ := setupTrash(t)
	svc.MoveToTrash("test-post")

	entries, _ := os.ReadDir(filepath.Join(svc.trashDir))
	if len(entries) == 0 {
		t.Fatal("no entries in trash")
	}
	trashID := entries[0].Name()

	if err := svc.Restore(trashID); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(postDir, "test-post", "index.md")); os.IsNotExist(err) {
		t.Error("expected post to be restored")
	}
}

func TestList(t *testing.T) {
	svc, _, _ := setupTrash(t)
	svc.MoveToTrash("test-post")
	items, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 trash item, got %d", len(items))
	}
}
```

- [ ] **Step 2: Run tests, fix compilation**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/trash/ -v -count=1`
Adjust to match actual trash package API.

- [ ] **Step 3: Verify all tests pass**

- [ ] **Step 4: Commit**

```bash
git add internal/trash/trash_test.go
git commit -m "test: add trash package tests (move, restore, list)"
```

---

### Task 1.4: HTTP API Integration Tests

**Files:**
- Create: `internal/httpapi/server_test.go`

- [ ] **Step 1: Write HTTP handler tests**

```go
// internal/httpapi/server_test.go
package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	// Create a minimal server with test dependencies
	// Adjust constructor to match actual Server struct
	srv := &Server{
		// Fill with test defaults
	}
	return srv
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/studio/api/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", body["status"])
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	srv := newTestServer(t)
	body, _ := json.Marshal(map[string]string{"password": "wrong"})
	req := httptest.NewRequest("POST", "/studio/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPosts_Unauthorized(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/studio/api/posts", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSecureHeaders(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/studio/api/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("X-Frame-Options") == "" {
		t.Error("expected X-Frame-Options header")
	}
}
```

- [ ] **Step 2: Run tests, fix compilation**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test ./internal/httpapi/ -v -count=1`
Adjust `newTestServer` to match actual Server constructor. May need to mock dependencies.

- [ ] **Step 3: Verify all tests pass**

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server_test.go
git commit -m "test: add HTTP API integration tests (health, auth, security headers)"
```

---

### Task 1.5: CI Test Pipeline

**Files:**
- Modify: `.github/workflows/deploy.yml`
- Create: `.github/workflows/test.yml`

- [ ] **Step 1: Create test workflow**

```yaml
# .github/workflows/test.yml
name: Tests
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go vet ./...
      - run: go test -race -cover ./...
      - run: go build ./cmd/server

  frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: web
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
          cache-dependency-path: web/package-lock.json
      - run: npm ci
      - run: npm run lint
      - run: npm run build
      - run: npm run test
```

- [ ] **Step 2: Run locally to verify**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test -race -cover ./...`
Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run test`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add separate test workflow (backend + frontend)"
```

---

### Task 1.6: Frontend Component Tests — PostsView

**Files:**
- Create: `web/src/__tests__/PostsView.test.ts`

- [ ] **Step 1: Install test dependencies**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install -D @vue/test-utils`

- [ ] **Step 2: Write PostsView tests**

```typescript
// web/src/__tests__/PostsView.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import PostsView from '@/views/PostsView.vue'

vi.mock('@/services/api', () => ({
  api: {
    posts: vi.fn().mockResolvedValue([
      { slug: 'hello-world', title: 'Hello World', date: '2026-01-01', draft: false, tags: ['test'], categories: [], syncStatus: 'synced', remoteMtime: '', cachedRemoteMtime: '', latestBackupId: '', large: false },
      { slug: 'draft-post', title: 'Draft', date: '2026-01-02', draft: true, tags: [], categories: [], syncStatus: 'synced', remoteMtime: '', cachedRemoteMtime: '', latestBackupId: '', large: false },
    ]),
    trashPost: vi.fn().mockResolvedValue({}),
  },
}))

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/posts', component: PostsView }],
})

describe('PostsView', () => {
  beforeEach(() => {
    router.push('/posts')
  })

  it('renders post list', async () => {
    const wrapper = mount(PostsView, {
      global: { plugins: [router] },
    })
    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Hello World')
    })
  })

  it('shows draft badge for draft posts', async () => {
    const wrapper = mount(PostsView, {
      global: { plugins: [router] },
    })
    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Draft')
    })
  })
})
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npx vitest run src/__tests__/PostsView.test.ts`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add web/src/__tests__/PostsView.test.ts web/package.json web/package-lock.json
git commit -m "test: add PostsView component tests"
```

---

### Task 1.7: Frontend Composable Tests — useEditor

**Files:**
- Create: `web/src/__tests__/useEditor.test.ts`

- [ ] **Step 1: Write useEditor tests**

```typescript
// web/src/__tests__/useEditor.test.ts
import { describe, it, expect, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { countWords } from '@/utils/words'

describe('countWords', () => {
  it('counts English words', () => {
    expect(countWords('hello world')).toBe(2)
  })
  it('counts Chinese characters', () => {
    expect(countWords('你好世界')).toBe(4)
  })
  it('handles mixed content', () => {
    expect(countWords('hello 你好 world')).toBe(3)
  })
  it('handles empty string', () => {
    expect(countWords('')).toBe(0)
  })
})

describe('useEditor composable', () => {
  it('exports expected functions', async () => {
    // Dynamically import to check exports
    const mod = await import('@/composables/useEditor')
    expect(mod.useEditor).toBeDefined()
    expect(typeof mod.useEditor).toBe('function')
  })
})
```

- [ ] **Step 2: Run tests**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npx vitest run src/__tests__/useEditor.test.ts`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add web/src/__tests__/useEditor.test.ts
git commit -m "test: add useEditor composable tests"
```

---

## Phase 2: Security Hardening

### Task 2.1: Vue Router Navigation Guard

**Files:**
- Modify: `web/src/App.vue`
- Modify: `web/src/router/index.ts`

- [ ] **Step 1: Add auth check function to store**

In `web/src/store/index.ts`, ensure there's an `isAuthenticated` computed or method. The store already has `session.authenticated`.

- [ ] **Step 2: Add router guard**

In `web/src/App.vue`, the guard already exists (check current implementation). Verify it checks `store.session.authenticated` and redirects to `/login` for protected routes. If missing, add:

```typescript
router.beforeEach(async (to) => {
  if (!store.session.authenticated && to.name !== 'login') {
    // Try to restore session
    try {
      const session = await api.session()
      if (session.authenticated) {
        store.session = session
        return // allow navigation
      }
    } catch {}
    return { name: 'login' }
  }
})
```

- [ ] **Step 3: Test manually**

Navigate to `/posts` without logging in. Should redirect to `/login`.

- [ ] **Step 4: Commit**

```bash
git add web/src/App.vue
git commit -m "fix: add router navigation guard for authenticated routes"
```

---

### Task 2.2: Fix 401 Redirect to Use Router

**Files:**
- Modify: `web/src/services/api.ts`

- [ ] **Step 1: Replace window.location.href with router push**

The current 401 handler at line 116 uses `window.location.href = '/studio/#/login'`. Replace with a callback pattern:

```typescript
// At the top of api.ts, export a setter for the redirect handler
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler
}

// In the request function, replace the 401 block:
if (response.status === 401 && path !== '/auth/login' && path !== '/session') {
  onUnauthorized?.()
  return new Promise(() => {})
}
```

- [ ] **Step 2: Wire up in App.vue**

```typescript
import { setUnauthorizedHandler } from '@/services/api'

// In setup:
setUnauthorizedHandler(() => {
  store.session.authenticated = false
  router.push('/login')
})
```

- [ ] **Step 3: Test**

Log in, then clear cookies manually. Navigate to an API-backed page. Should redirect to `/login` without full page reload.

- [ ] **Step 4: Commit**

```bash
git add web/src/services/api.ts web/src/App.vue
git commit -m "fix: use router navigation for 401 redirect instead of full page reload"
```

---

### Task 2.3: Request Body Size Limit Middleware

**Files:**
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add maxBytes middleware**

```go
func withMaxBytes(next http.Handler, maxBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Apply to write endpoints**

Wrap the write router with `withMaxBytes(handler, 10<<20)` (10MB) for file uploads, and `withMaxBytes(handler, 1<<20)` (1MB) for JSON endpoints.

- [ ] **Step 3: Test**

Send a POST with a body exceeding the limit. Verify 413 response.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go
git commit -m "security: add request body size limit middleware"
```

---

### Task 2.4: Rate Limiting on Write Endpoints

**Files:**
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add general rate limiter**

The auth package already has `NewLoginRateLimiter`. Create a similar limiter for write endpoints:

```go
func withWriteRateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 req/s burst
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, `{"ok":false,"error":{"code":"RATE_LIMITED","message":"Too many requests"}}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Apply to write routes**

- [ ] **Step 3: Test**

Send rapid requests. Verify 429 after burst.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go
git commit -m "security: add rate limiting on write API endpoints"
```

---

### Task 2.5: Upload File Validation

**Files:**
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add file type validation**

In the asset upload handler, validate file extension against an allowlist:

```go
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".svg": true, ".pdf": true,
}
```

- [ ] **Step 2: Validate MIME type**

Read first 512 bytes and use `http.DetectContentType` to verify the file type matches the extension.

- [ ] **Step 3: Test**

Upload a `.txt` file renamed to `.jpg`. Verify rejection.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go
git commit -m "security: validate uploaded file type (extension + MIME)"
```

---

## Phase 3: Performance Optimization

### Task 3.1: Hugo Build Request Deduplication

**Files:**
- Modify: `internal/hugobuild/runner.go`

- [ ] **Step 1: Add singleflight for builds**

```go
import "golang.org/x/sync/singleflight"

var buildGroup singleflight.Group

func (r *Runner) BuildWithDedup(args ...string) (Result, error) {
	key := strings.Join(args, " ")
	v, err, _ := buildGroup.Do(key, func() (interface{}, error) {
		return r.Build(args...)
	})
	if err != nil {
		return Result{}, err
	}
	return v.(Result), nil
}
```

- [ ] **Step 2: Wire up in preview and publish handlers**

- [ ] **Step 3: Test**

Send concurrent preview requests for the same post. Verify only one Hugo build runs.

- [ ] **Step 4: Commit**

```bash
git add internal/hugobuild/runner.go
git commit -m "perf: deduplicate concurrent Hugo builds via singleflight"
```

---

### Task 3.2: Post List Caching

**Files:**
- Modify: `internal/publish/service.go`

- [ ] **Step 1: Add in-memory cache with TTL**

```go
type postListCache struct {
	entries []PostState
	expires time.Time
}

var listCache postListCache
var listCacheMu sync.RWMutex

func (s *Service) ListPosts() ([]PostState, error) {
	listCacheMu.RLock()
	if time.Now().Before(listCache.expires) {
		defer listCacheMu.RUnlock()
		return listCache.entries, nil
	}
	listCacheMu.RUnlock()

	// ... existing list logic ...

	listCacheMu.Lock()
	listCache.entries = result
	listCache.expires = time.Now().Add(30 * time.Second)
	listCacheMu.Unlock()

	return result, nil
}

func (s *Service) InvalidateListCache() {
	listCacheMu.Lock()
	listCache.expires = time.Time{}
	listCacheMu.Unlock()
}
```

- [ ] **Step 2: Invalidate on publish/save/trash**

Call `InvalidateListCache()` in `SaveDraft`, `PublishBlog`, and trash operations.

- [ ] **Step 3: Test**

- [ ] **Step 4: Commit**

```bash
git add internal/publish/service.go
git commit -m "perf: cache post list for 30s, invalidate on mutations"
```

---

### Task 3.3: Frontend — Virtual Scrolling for Post List

**Files:**
- Modify: `web/src/views/PostsView.vue`

- [ ] **Step 1: Install @tanstack/vue-virtual**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install @tanstack/vue-virtual`

- [ ] **Step 2: Implement virtual list**

Replace the `v-for` list with `useVirtualizer` from `@tanstack/vue-virtual`. Keep existing card styling.

- [ ] **Step 3: Test with large dataset**

Mock 1000 posts and verify smooth scrolling.

- [ ] **Step 4: Commit**

```bash
git add web/src/views/PostsView.vue web/package.json web/package-lock.json
git commit -m "perf: virtual scrolling for post list (handles 1000+ posts)"
```

---

### Task 3.4: Frontend — CodeMirror Lazy Extension Loading

**Files:**
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Lazy load search and WYSIWYG**

```typescript
// In mount(), load extensions dynamically
const searchExt = await import('@codemirror/search')
const wysiwygExt = await import('./useWysiwyg')
```

- [ ] **Step 2: Split editor chunk**

In `vite.config.ts`, add CodeMirror search to the vendor chunk:

```typescript
manualChunks: {
  'vendor-codemirror': [...existing, '@codemirror/search'],
}
```

- [ ] **Step 3: Measure bundle size change**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Compare `dist/assets/` sizes before and after.

- [ ] **Step 4: Commit**

```bash
git add web/src/composables/useEditor.ts web/vite.config.ts
git commit -m "perf: lazy load CodeMirror extensions, optimize chunk splitting"
```

---

### Task 3.5: Backend — Response Compression

**Files:**
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add gzip middleware**

```go
import "compress/gzip"

func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
```

- [ ] **Step 2: Apply to JSON API responses**

- [ ] **Step 3: Test**

Request with `Accept-Encoding: gzip`. Verify compressed response.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go
git commit -m "perf: add gzip compression for API responses"
```

---

## Phase 4: UI/UX Improvements

### Task 4.1: Error Boundary Component

**Files:**
- Create: `web/src/components/ErrorBoundary.vue`
- Modify: `web/src/App.vue`

- [ ] **Step 1: Create ErrorBoundary component**

```vue
<!-- web/src/components/ErrorBoundary.vue -->
<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import { AlertCircleIcon, RotateCcwIcon } from 'lucide-vue-next'

const error = ref<Error | null>(null)

onErrorCaptured((err) => {
  error.value = err
  return false // stop propagation
})

function retry() {
  error.value = null
}
</script>

<template>
  <slot v-if="!error" />
  <div v-else class="flex flex-col items-center justify-center gap-4 py-20 animate-fade-up">
    <AlertCircleIcon class="h-12 w-12 text-destructive/40" />
    <p class="text-muted-foreground font-serif">Something went wrong</p>
    <p class="text-xs text-muted-foreground/60 max-w-md text-center">{{ error.message }}</p>
    <button class="px-4 py-2 rounded-full bg-primary text-primary-foreground text-sm" @click="retry">
      <RotateCcwIcon class="h-3 w-3 mr-1 inline" /> Try Again
    </button>
  </div>
</template>
```

- [ ] **Step 2: Wrap router-view in App.vue**

```vue
<ErrorBoundary>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</ErrorBoundary>
```

- [ ] **Step 3: Test**

Throw an error in a component. Verify the error boundary catches it and shows the fallback UI.

- [ ] **Step 4: Commit**

```bash
git add web/src/components/ErrorBoundary.vue web/src/App.vue
git commit -m "feat: add error boundary component to catch view crashes"
```

---

### Task 4.2: Accessibility — ARIA Labels

**Files:**
- Modify: `web/src/views/LoginView.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/views/HealthView.vue`

- [ ] **Step 1: Add aria-labels to LoginView**

```vue
<!-- Theme toggle -->
<button aria-label="Toggle theme" @click="toggleTheme">
<!-- Language toggle -->
<button aria-label="Switch language" @click="toggleLang">
<!-- Error message -->
<p v-if="error" role="alert" aria-live="assertive">{{ error }}</p>
```

- [ ] **Step 2: Add aria-labels to EditorView toolbar**

```vue
<Button variant="ghost" size="icon" aria-label="Bold" title="Bold (⌘B)">
<Button variant="ghost" size="icon" aria-label="Italic" title="Italic (⌘I)">
<!-- ... for all toolbar buttons -->
```

- [ ] **Step 3: Add aria-expanded to meta toggle**

```vue
<button :aria-expanded="metaExpanded" aria-controls="meta-panel">
```

- [ ] **Step 4: Add role="status" to HealthView**

```vue
<div role="status" aria-label="Server health status">
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/LoginView.vue web/src/views/EditorView.vue web/src/views/HealthView.vue
git commit -m "a11y: add ARIA labels to interactive elements across views"
```

---

### Task 4.3: i18n — Fix Hardcoded Strings

**Files:**
- Modify: `web/src/i18n/index.ts`
- Modify: `web/src/views/LoginView.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/views/HealthView.vue`

- [ ] **Step 1: Add missing i18n keys**

In `web/src/i18n/index.ts`, add to both `zh` and `en` dicts:

```typescript
// zh
loginSubtitle: '写作是一种修行',
editMeta: '点击编辑元信息…',
collapse: '收起',
postLoadFailed: '文章加载失败',
// en
loginSubtitle: 'Writing is a practice',
editMeta: 'Click to edit metadata…',
collapse: 'Collapse',
postLoadFailed: 'Failed to load post',
```

- [ ] **Step 2: Replace hardcoded strings in views**

Replace all Chinese hardcoded strings with `t.value.keyName`.

- [ ] **Step 3: Test**

Switch language to English. Verify no Chinese strings remain in the UI.

- [ ] **Step 4: Commit**

```bash
git add web/src/i18n/index.ts web/src/views/LoginView.vue web/src/views/EditorView.vue web/src/views/HealthView.vue
git commit -m "i18n: replace hardcoded Chinese strings with translation keys"
```

---

### Task 4.4: Skip-to-Content Link

**Files:**
- Modify: `web/src/App.vue`

- [ ] **Step 1: Add skip link**

```vue
<template>
  <a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:p-2 focus:bg-background">
    Skip to content
  </a>
  <!-- ... existing layout ... -->
  <main id="main-content" class="...">
    <!-- router-view -->
  </main>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/App.vue
git commit -m "a11y: add skip-to-content link for keyboard navigation"
```

---

## Phase 5: Editor Enhancements

### Task 5.1: Syntax Highlighting in Code Blocks

**Files:**
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Add @codemirror/language-data**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install @codemirror/language-data`

- [ ] **Step 2: Configure markdown with codeLanguages**

```typescript
import { languages } from '@codemirror/language-data'

markdown({
  extensions: [GFM, mathExtension()],
  codeLanguages: languages,
})
```

- [ ] **Step 3: Test**

Write a fenced code block with ` ```python `. Verify syntax highlighting appears.

- [ ] **Step 4: Commit**

```bash
git add web/src/composables/useEditor.ts web/package.json web/package-lock.json
git commit -m "feat: syntax highlighting in fenced code blocks via language-data"
```

---

### Task 5.2: Editor Keybinding Cheatsheet

**Files:**
- Create: `web/src/components/KeybindingHelp.vue`
- Modify: `web/src/views/EditorView.vue`

- [ ] **Step 1: Create help dialog component**

```vue
<!-- web/src/components/KeybindingHelp.vue -->
<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

defineProps<{ open: boolean }>()
defineEmits<{ 'update:open': [value: boolean] }>()

const shortcuts = [
  { key: '⌘S', action: 'Save draft' },
  { key: '⌘B', action: 'Bold' },
  { key: '⌘I', action: 'Italic' },
  { key: '⌘K', action: 'Insert link' },
  { key: '⌘F', action: 'Search' },
  { key: 'Tab', action: 'Indent' },
  { key: 'Shift+Tab', action: 'Outdent' },
]
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Keyboard Shortcuts</DialogTitle>
      </DialogHeader>
      <div class="space-y-2">
        <div v-for="s in shortcuts" :key="s.key" class="flex justify-between text-sm">
          <span class="text-muted-foreground">{{ s.action }}</span>
          <kbd class="px-2 py-0.5 rounded bg-muted text-xs font-mono">{{ s.key }}</kbd>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
```

- [ ] **Step 2: Add ? keybinding to open help**

In `useEditor.ts`, add to keymap:

```typescript
{ key: '?', run: () => { showKeybindingHelp.value = true; return true } },
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/KeybindingHelp.vue web/src/views/EditorView.vue web/src/composables/useEditor.ts
git commit -m "feat: keyboard shortcut help dialog (? key)"
```

---

### Task 5.3: Drag-and-Drop Image Upload

**Files:**
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Add drop handler**

```typescript
const dropHandler = EditorView.domEventHandlers({
  drop(event, editorView) {
    if (!onPasteImage) return false
    const files = event.dataTransfer?.files
    if (!files?.length) return false
    const file = files[0]
    if (!file.type.startsWith('image/')) return false
    event.preventDefault()
    onPasteImage(file).then((filename) => {
      const insert = `![](${filename})`
      const pos = editorView.state.selection.main.head
      editorView.dispatch(editorView.state.update({
        changes: { from: pos, insert },
        selection: { anchor: pos + insert.length },
      }))
    })
    return true
  },
})
```

- [ ] **Step 2: Add to extensions in mount()**

- [ ] **Step 3: Test**

Drag an image file onto the editor. Verify it uploads and inserts markdown.

- [ ] **Step 4: Commit**

```bash
git add web/src/composables/useEditor.ts
git commit -m "feat: drag-and-drop image upload in editor"
```

---

## Phase 6: Content Management Features

### Task 6.1: Scheduled Publishing

**Files:**
- Modify: `internal/publish/service.go`
- Modify: `internal/httpapi/server.go`
- Modify: `web/src/views/EditorView.vue`

- [ ] **Step 1: Add scheduledAt field to FrontMatter**

```go
type FrontMatter struct {
	// ... existing fields
	ScheduledAt string `yaml:"scheduledAt,omitempty" json:"scheduledAt,omitempty"`
}
```

- [ ] **Step 2: Add background scheduler goroutine**

In `cmd/server/main.go`, add a goroutine that checks every minute for posts with `scheduledAt` in the past that haven't been published yet.

- [ ] **Step 3: Add UI for scheduling**

In EditorView, add a datetime input in the meta section for `scheduledAt`.

- [ ] **Step 4: Test**

Create a post with `scheduledAt` set to 1 minute in the future. Wait. Verify it publishes automatically.

- [ ] **Step 5: Commit**

```bash
git add internal/publish/service.go internal/httpapi/server.go web/src/views/EditorView.vue cmd/server/main.go
git commit -m "feat: scheduled publishing (auto-publish at specified time)"
```

---

### Task 6.2: Bulk Operations

**Files:**
- Modify: `web/src/views/PostsView.vue`
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add bulk endpoints**

```go
// POST /api/posts/bulk/trash
// POST /api/posts/bulk/publish
// POST /api/posts/bulk/tag
```

- [ ] **Step 2: Add selection UI**

In PostsView, add checkboxes to each post card and a floating action bar for bulk operations.

- [ ] **Step 3: Test**

Select 3 posts, click "Move to Trash". Verify all 3 are trashed.

- [ ] **Step 4: Commit**

```bash
git add web/src/views/PostsView.vue internal/httpapi/server.go
git commit -m "feat: bulk operations (trash, publish, tag multiple posts)"
```

---

### Task 6.3: Tag Management

**Files:**
- Create: `web/src/views/TagsView.vue`
- Modify: `web/src/router/index.ts`

- [ ] **Step 1: Create TagsView**

A dedicated view for managing all tags: rename, merge, delete. Shows tag usage count.

- [ ] **Step 2: Add route**

```typescript
{ name: 'tags', path: '/tags', component: () => import('@/views/TagsView.vue') }
```

- [ ] **Step 3: Add sidebar link**

- [ ] **Step 4: Commit**

```bash
git add web/src/views/TagsView.vue web/src/router/index.ts
git commit -m "feat: tag management view (rename, merge, delete tags)"
```

---

## Phase 7: Monitoring & Observability

### Task 7.1: Enhanced Health Checks

**Files:**
- Modify: `internal/httpapi/server.go`
- Modify: `web/src/views/HealthView.vue`

- [ ] **Step 1: Add Hugo build test to health check**

```go
func (s *Server) checkHugoBuild() healthCheck {
	start := time.Now()
	_, err := s.hugoRunner.Build("--renderToMemory")
	elapsed := time.Since(start)
	status := "ok"
	msg := fmt.Sprintf("Hugo build completed in %s", elapsed)
	if err != nil {
		status = "error"
		msg = fmt.Sprintf("Hugo build failed: %v", err)
	}
	return healthCheck{Name: "hugo-build", Status: status, Message: msg}
}
```

- [ ] **Step 2: Add disk space check**

```go
func (s *Server) checkDiskSpace() healthCheck {
	var stat syscall.Statfs_t
	syscall.Statfs(s.config.Site.BlogRoot, &stat)
	freeGB := float64(stat.Bavail*uint64(stat.Bsize)) / (1 << 30)
	status := "ok"
	if freeGB < 1 {
		status = "warn"
	}
	return healthCheck{Name: "disk-space", Status: status, Message: fmt.Sprintf("%.1f GB free", freeGB)}
}
```

- [ ] **Step 3: Display in HealthView**

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go web/src/views/HealthView.vue
git commit -m "feat: enhanced health checks (hugo build test, disk space)"
```

---

### Task 7.2: Prometheus Metrics Dashboard

**Files:**
- Modify: `web/src/views/HealthView.vue`

- [ ] **Step 1: Fetch /api/metrics**

Parse Prometheus text format and display key metrics: request count, error rate, Hugo build duration, active sessions.

- [ ] **Step 2: Create simple chart components**

Use CSS bar charts or sparklines (no chart library needed for this scale).

- [ ] **Step 3: Commit**

```bash
git add web/src/views/HealthView.vue
git commit -m "feat: metrics dashboard in health view (request count, error rate, build duration)"
```

---

### Task 7.3: Audit Log Search

**Files:**
- Modify: `web/src/views/settings/AuditTab.vue`
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Add search/filter params to audit endpoint**

```go
// GET /api/audit?limit=50&operation=publish&search=keyword
```

- [ ] **Step 2: Add search input to AuditTab**

- [ ] **Step 3: Commit**

```bash
git add web/src/views/settings/AuditTab.vue internal/httpapi/server.go
git commit -m "feat: audit log search and filter by operation type"
```

---

## Phase 8: Advanced Features

### Task 8.1: Post Statistics

**Files:**
- Modify: `web/src/views/EditorView.vue`
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Track view counts**

Add a simple JSON counter file at `/data/stats/<slug>.json`:

```json
{"views": 42, "lastViewed": "2026-05-10T12:00:00Z"}
```

- [ ] **Step 2: Add stats API endpoint**

```go
// GET /api/posts/:slug/stats
```

- [ ] **Step 3: Display in editor status bar**

Show view count and reading time alongside word count.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go web/src/views/EditorView.vue
git commit -m "feat: post view statistics (count, last viewed)"
```

---

### Task 8.2: Image Gallery

**Files:**
- Create: `web/src/components/ImageGallery.vue`
- Modify: `web/src/views/EditorView.vue`

- [ ] **Step 1: Create gallery component**

Shows all images uploaded to the current post with preview, delete, and copy-markdown-link actions.

- [ ] **Step 2: Add gallery button to toolbar**

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ImageGallery.vue web/src/views/EditorView.vue
git commit -m "feat: image gallery for managing post assets"
```

---

### Task 8.3: Export/Import

**Files:**
- Modify: `internal/httpapi/server.go`
- Modify: `web/src/views/PostsView.vue`

- [ ] **Step 1: Add export endpoint**

```go
// GET /api/posts/export?format=zip
// Returns a ZIP of all posts as markdown files
```

- [ ] **Step 2: Add import endpoint**

```go
// POST /api/posts/import (multipart form with ZIP)
```

- [ ] **Step 3: Add UI buttons**

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server.go web/src/views/PostsView.vue
git commit -m "feat: export all posts as ZIP, import from ZIP"
```

---

## Verification Checklist

After completing all phases, run the full verification:

```bash
# Backend
cd /Users/xiang/Desktop/personal-server/blog-studio-web
go vet ./...
go test -race -cover ./...
go build ./cmd/server

# Frontend
cd web
npm run lint
npm run build
npm run test
npm run test:coverage  # verify >60% coverage

# Docker
docker compose build
docker compose up -d
curl -f http://localhost:8080/studio/api/health
```

---

## Execution Order Summary

| Phase | Track | Tasks | Effort |
|-------|-------|-------|--------|
| 1 | Testing | 1.1-1.7 | 7 tasks, ~2 days |
| 2 | Security | 2.1-2.5 | 5 tasks, ~1 day |
| 3 | Performance | 3.1-3.5 | 5 tasks, ~1 day |
| 4 | UI/UX | 4.1-4.4 | 4 tasks, ~1 day |
| 5 | Editor | 5.1-5.3 | 3 tasks, ~0.5 day |
| 6 | Features | 6.1-6.3 | 3 tasks, ~1 day |
| 7 | Monitoring | 7.1-7.3 | 3 tasks, ~0.5 day |
| 8 | Advanced | 8.1-8.3 | 3 tasks, ~1 day |

**Total: 33 tasks, ~8 days of focused work**

Phases 5-8 are independent and can be parallelized with multiple agents.
