# Blog Studio Web — V2 Optimization Design Spec

## Goal

Transform blog-studio-web into a polished, production-grade blog management tool with a focus on **editor experience** and **UI design quality**, while maintaining all-around improvements in testing and performance.

## Architecture

Progressive enhancement approach — each task improves the existing codebase without breaking changes. Six phases, ~38 tasks, each independently shippable.

**Tech Stack:** Go 1.25 (backend), Vue 3.5 + TypeScript + Vite 6 (frontend), CodeMirror 6 (editor), shadcn-vue + Tailwind v4 (UI), Vitest (tests), Playwright (E2E), Docker + nginx (deployment)

**Repository:** `/Users/xiang/Desktop/personal-server/blog-studio-web/`

---

## Phase 1: Editor Core (8 tasks)

The editor is the primary interaction surface. These tasks make it feel like a professional Markdown editor.

### 1.1 Split View

Side-by-side source editing + live preview.

- Left panel: CodeMirror source editor (existing)
- Right panel: live HTML preview (reuse `MarkdownPreview.vue`)
- Resizable divider (drag to adjust ratio)
- Three modes: source-only, preview-only, split (toggle via toolbar)
- Preview scrolls in sync with editor (scroll mapping)
- Store mode preference in localStorage

### 1.2 Outline Panel

Document structure navigation.

- Collapsible left sidebar (between sidebar nav and editor)
- Extracts headings from Lezer AST (H1-H6)
- Tree structure with indentation for nesting levels
- Click to jump to heading in editor
- Auto-highlights heading nearest to current scroll position
- Collapse/expand individual sections
- Toggle via toolbar button or keyboard shortcut

### 1.3 Table Editing

Markdown table support.

- Toolbar button inserts table template (3x3 default)
- In WYSIWYG mode: Tab navigates cells, Enter adds row
- Context menu: add/remove rows and columns
- Auto-align column widths in source mode
- Paste from spreadsheet (clipboard → markdown table)

### 1.4 Math/KaTeX Preview

LaTeX math rendering in preview pane.

- Inline math: `$...$` → rendered inline
- Block math: `$$...$$` → centered block rendering
- Use KaTeX for rendering (fast, server-side friendly)
- Error display for invalid LaTeX (red border + error message)
- Already have Lezer parsers for math; wire up KaTeX in preview

### 1.5 Code Block Syntax Themes

Customizable code highlighting.

- Theme selector in settings or editor preferences
- Preset themes: One Dark, GitHub Light, Catppuccin Mocha, Solarized
- Affects both editor code blocks and preview rendering
- Persist selection in localStorage
- Apply via CodeMirror Compartment for live switching

### 1.6 Auto-save UX

Clear save state communication.

- Status indicator in header: "Saving..." (pulse), "Saved at 14:32" (green), "Unsaved changes" (yellow dot)
- ⌘S for immediate save (bypasses debounce)
- Confirm dialog on navigation with unsaved changes (exists, polish wording)
- Auto-save debounce: 1.5s after last keystroke (existing)
- Visual feedback: subtle flash on successful save

### 1.7 Find and Replace

Enhanced search experience.

- ⌘F opens inline search bar at top of editor (not CodeMirror overlay)
- Input field with real-time match highlighting
- Match count display ("3 of 12")
- ↑↓ navigation between matches
- Replace: single match or replace all
- Toggle: case-sensitive, whole word, regex
- Escape to close

### 1.8 Editor Customization

Personal editor preferences.

- Font family: system monospace, JetBrains Mono, Fira Code (dropdown)
- Font size: 12-24px slider
- Line height: 1.2-2.0 slider
- Line numbers: on/off toggle
- Preset profiles: "Default", "Typewriter" (large font, centered), "Compact" (small font, dense)
- All settings persisted in localStorage

---

## Phase 2: UI Design System (12 tasks)

Make the UI feel cohesive, polished, and delightful.

### 2.1 Design Tokens

CSS custom properties as the foundation.

```css
:root {
  /* Colors (light) */
  --color-primary: oklch(0.55 0.15 250);
  --color-primary-foreground: oklch(0.98 0 0);
  --color-background: oklch(0.98 0.005 250);
  --color-foreground: oklch(0.15 0.01 250);
  --color-muted: oklch(0.95 0.005 250);
  --color-muted-foreground: oklch(0.50 0.01 250);
  --color-card: oklch(1 0 0);
  --color-border: oklch(0.90 0.005 250);
  --color-destructive: oklch(0.55 0.2 25);
  --color-success: oklch(0.60 0.18 145);
  --color-warning: oklch(0.70 0.15 75);

  /* Spacing */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;
  --space-12: 48px;
  --space-16: 64px;

  /* Typography */
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-serif: 'Noto Serif SC', Georgia, serif;
  --font-mono: 'JetBrains Mono', monospace;
  --text-xs: 0.75rem;
  --text-sm: 0.875rem;
  --text-base: 1rem;
  --text-lg: 1.125rem;
  --text-xl: 1.25rem;
  --text-2xl: 1.5rem;
  --text-3xl: 2rem;

  /* Borders */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-full: 9999px;

  /* Shadows */
  --shadow-sm: 0 1px 2px oklch(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px oklch(0 0 0 / 0.07);
  --shadow-lg: 0 10px 15px oklch(0 0 0 / 0.10);
  --shadow-xl: 0 20px 25px oklch(0 0 0 / 0.15);
}
```

- Dark mode variants via `[data-theme="dark"]` or `.dark` class
- All components reference tokens, not hardcoded values
- Semantic mapping: `--color-primary` → button bg, link color, active states

### 2.2 Color System

Beyond basic tokens.

- Semantic status colors: draft (amber), published (green), error (red), info (blue)
- Chart palette for metrics: 6-8 harmonious colors for data visualization
- Overlay colors: `oklch(0 0 0 / 0.5)` for modal backdrops
- Gradient accents: subtle gradient for brand elements (login page, sidebar header)
- WCAG AA contrast verification for all text/background combinations

### 2.3 Typography System

Content-focused type scale.

- Heading hierarchy: h1 (2rem/700), h2 (1.5rem/600), h3 (1.25rem/600), h4-h6 (1rem/600)
- Body: 16px base, 1.6 line-height, 400 weight
- Caption/metadata: 12px, muted color, 500 weight
- Code: monospace, 14px, background tint
- Chinese + English mixed: `word-break: break-word`, proper line-height for CJK
- Serif for content headings and body, sans-serif for UI chrome

### 2.4 Micro-interactions

Polish every interactive element.

- **Buttons**: `transform: scale(0.97)` on active, shadow change on hover, 150ms transition
- **Cards**: `translateY(-2px)` + shadow-lg on hover, 200ms transition
- **Toggles**: smooth spring animation (300ms, cubic-bezier(0.34, 1.56, 0.64, 1))
- **Focus rings**: 2px offset, accent color, `transition: box-shadow 150ms`
- **Inputs**: border-color transition on focus (200ms), placeholder fade
- **List items**: background fade on hover (100ms), slide-in on mount (200ms)
- **Links**: underline slides in from left on hover

### 2.5 Animation System

Consistent motion language.

- **Page transitions**: fade-up (enter 250ms ease-out) + fade-out (leave 150ms ease-in)
- **List animations**: staggered enter (50ms delay per item, max 10 items), exit fade
- **Modal**: scale(0.95→1) + opacity(0→1), 200ms, backdrop fade 150ms
- **Toast**: slide-in from right (300ms), auto-dismiss with progress bar
- **Easing**: `cubic-bezier(0.16, 1, 0.3, 1)` for spring-like feel
- **Reduced motion**: `@media (prefers-reduced-motion: reduce)` — disable all animations, instant transitions

### 2.6 Mobile Responsive

Touch-first on small screens.

- **Breakpoints**: sm (640px), md (768px), lg (1024px), xl (1280px)
- **Sidebar**: slide-in drawer with backdrop overlay on < 768px, hamburger toggle
- **Editor**: full-width, toolbar pinned to bottom on mobile
- **Post list**: card layout → compact list on mobile
- **Settings**: tabs stack vertically on mobile
- **Touch targets**: minimum 44px × 44px for all interactive elements
- **Safe areas**: `env(safe-area-inset-*)` for notched devices

### 2.7 Dark Mode Refinements

Complete dark experience.

- Audit every component in dark mode, fix contrast issues
- Image containers: `brightness(0.9)` filter to prevent brightness shock
- Code blocks: dark syntax theme (One Dark or Catppuccin Mocha)
- Charts: dark-friendly palette (muted, low-saturation colors)
- Theme transition: `transition: background-color 200ms, color 200ms` on body
- Auto-detect: respect `prefers-color-scheme`, allow manual override
- Sidebar: slightly different bg from main content for depth

### 2.8 Empty States

Friendly, not frustrating.

- **No posts**: SVG illustration + "Write your first post" button + helpful tip
- **No search results**: illustration + "Try different keywords" + clear search button
- **No trash**: illustration + "Trash is empty"
- **First-time user**: subtle onboarding tooltips on key features
- **Error states**: distinct illustrations for 404 (page not found), 500 (server error), network error
- Consistent illustration style: line art, muted colors, friendly tone

### 2.9 Loading Skeletons

Content-shaped placeholders.

- **Post list**: rectangular card skeletons with shimmer, matching real card dimensions
- **Editor**: line-shaped skeletons (varying widths for paragraph feel)
- **Settings**: form-field skeletons (label + input shape)
- **Health**: check-item skeletons (icon + text + status)
- **Shimmer**: left-to-right gradient sweep, 1.5s ease-in-out infinite
- Replace all `<div>Loading...</div>` spinners with skeletons

### 2.10 Card & Surface Design

Consistent elevation system.

- **Level 0**: page background — flat, subtle texture or solid
- **Level 1**: cards, sidebars — 1px border + shadow-sm, radius-md
- **Level 2**: dropdowns, popovers — shadow-md, radius-md
- **Level 3**: modals, dialogs — shadow-xl + backdrop blur, radius-lg
- Consistent padding: cards (space-4), modals (space-6)
- Consistent border-radius: cards (radius-md), modals (radius-lg), buttons (radius-md)

### 2.11 Form & Input Design

Polished form experience.

- **Input**: 36px height, 1px border, radius-md, focus ring (2px accent), placeholder in muted color
- **Select**: custom dropdown with search filter, keyboard navigation (↑↓ Enter), animated open
- **Checkbox/Radio**: custom styled, smooth check animation (scale spring), accent color
- **Validation**: inline error with icon (AlertCircle) + red text below field, border turns red, subtle shake animation
- **Form layout**: label above input, 12px gap, responsive stacking on mobile
- **Textarea**: auto-resize to content, min 3 rows, max before scroll

### 2.12 Feedback & Notification System

Clear communication.

- **Toast**: 4 types (success/check, error/x, warning/triangle, info/circle), auto-dismiss 5s, action button, stack max 3
- **Alert/Banner**: inline dismissible, persistent until dismissed, semantic colors
- **Progress**: determinate bar (uploads, exports), indeterminate (builds, saves)
- **Badge**: status colors, pill shape, 2 sizes (sm/default)
- **Tooltip**: dark bg, white text, small font (12px), arrow, 300ms show delay, 100ms hide delay

---

## Phase 3: Editor Advanced (6 tasks)

### 3.1 Enhanced Markdown Toolbar

Rich grouped toolbar.

- **Text group**: Bold (⌘B), Italic (⌘I), Strikethrough, Inline Code, Link (⌘K)
- **Block group**: Heading dropdown (H1-H6), Blockquote, Code Block, Math Block
- **List group**: Ordered List, Unordered List, Task List
- **Insert group**: Table, Image, Horizontal Rule
- Groups separated by vertical dividers
- Active state: highlighted when cursor is in that format
- Tooltips with keyboard shortcuts on hover
- Responsive: groups collapse into overflow menu on narrow screens

### 3.2 Command Palette Enhancement

Full command center.

- ⌘K opens palette (existing, enhance)
- **Sections**: Recent Posts, Quick Actions, Settings
- **Commands**: New Post, Go to Settings, Toggle Theme, Toggle Dark Mode, Health Check
- **Search**: fuzzy match on post titles and commands
- **Keyboard**: ↑↓ navigate, Enter execute, Escape close, Tab switch section
- **Visual**: highlighted selected item, icons for each command type

### 3.3 Image Management

Enhanced asset handling.

- Gallery dialog accessible from toolbar button
- Grid layout with image thumbnails (3-4 per row)
- Non-image files shown with file-type icons
- Click to insert markdown reference `![alt](filename)`
- Upload new images directly from gallery
- Alt-text editor: inline input when inserting
- Copy markdown reference button per asset

### 3.4 Version History

Post-level revision tracking.

- Auto-save creates snapshots (keep last 10 per post)
- Side panel (right drawer) showing version list
- Each version: timestamp, word count delta, preview snippet
- Click to preview a version (read-only overlay)
- "Restore" button to revert to a version
- Simple diff view: additions in green, deletions in red
- Storage: JSON files in `{dataRoot}/history/{slug}/`

### 3.5 Word Count & Reading Stats

Editor footer status bar.

- Real-time word count (Chinese: 1 char = 1 word, English: space-delimited)
- Estimated reading time (English: 200 wpm, Chinese: 400 chars/min)
- Character count (with/without spaces)
- Current line:column position
- Total line count
- All stats update on every keystroke (debounced 100ms for performance)

### 3.6 Focus Mode

Distraction-free writing.

- Hides sidebar, toolbar, header — only editor content visible
- Centered narrow content area (max 700px width)
- Typewriter scrolling: current line stays vertically centered
- Subtle vignette or ambient background
- Escape to exit, or click outside content area
- Toggle via toolbar button or keyboard shortcut

---

## Phase 4: UI Components (5 tasks)

### 4.1 Toast Notification System

Replace vue-sonner with polished custom implementation.

- 4 types: success (green, check icon), error (red, x icon), warning (amber, triangle), info (blue, circle)
- Auto-dismiss with visible progress bar (5s default, ∞ for errors)
- Action button (e.g., "Undo" after delete, "Retry" on error)
- Stack management: max 3 visible, older ones collapse
- Enter animation: slide-in from right + fade
- Exit animation: slide-out right + fade
- Position: top-right (desktop), top-center (mobile)

### 4.2 Modal & Dialog Polish

Consistent dialog system.

- Backdrop: blur(4px) + dark overlay, click to dismiss (optional)
- Enter: scale(0.95→1) + opacity(0→1), 200ms
- Exit: reverse of enter, 150ms
- Focus trap: Tab cycles within dialog (radix handles this)
- Destructive confirmations: red accent, "Are you sure?" pattern
- Full-screen on mobile (< 640px)
- Consistent padding: 24px body, 16px header/footer

### 4.3 Form Validation UX

Inline validation patterns.

- Validate on blur (not on keystroke to avoid annoying users)
- Error: red border + shake animation (300ms) + error text with icon below field
- Success: green border + checkmark icon (optional, for registration-style forms)
- Form-level: scroll to first error on submit attempt
- Disabled state: submit button disabled until form is valid (optional)
- Error messages: specific and actionable ("Password must be 8+ characters", not "Invalid input")

### 4.4 Navigation Improvements

Better wayfinding.

- Breadcrumbs on nested pages: Settings > Security, Posts > my-post-slug
- Active state animation: smooth background transition on nav items
- Keyboard navigation: Tab through sidebar items, Enter to navigate
- Quick switcher: ⌘P to switch between recent pages
- Sidebar collapse: remembers state in localStorage

### 4.5 Data Table Component

Reusable table for structured data.

- Sortable columns (click header to sort, click again to reverse)
- Column resize (drag column borders)
- Row selection (checkboxes, bulk actions)
- Pagination or virtual scroll (configurable)
- Empty state per table (customizable message)
- Used for: audit logs, trash items, tag management, post stats

---

## Phase 5: Testing (4 tasks)

### 5.1 httpapi Test Coverage

Increase from 18.7% → 60%+.

- Test all CRUD endpoints (posts, config, site, pages)
- Test auth flows: login success/failure, logout, session, password change
- Test error responses: 400 (bad request), 404 (not found), 409 (conflict), 500 (internal)
- Test middleware: rate limiting triggers, gzip compression, max body size
- Test file upload: valid file, invalid type, oversized
- Use httptest.ResponseRecorder for unit-style tests

### 5.2 Storage & Publish Tests

Fill coverage gaps.

- Storage: file read/write/delete, directory creation, path traversal rejection, concurrent access
- Publish: conflict detection, rollback, cache invalidation, large file handling
- Backup: create/restore/list, retention policy, corruption handling
- Target: storage 60%+, publish 70%+, backup 70%+

### 5.3 Frontend Component Tests

Key component coverage.

- EditorView: mount with post, save flow, mode toggle, toolbar actions
- PostsView: search filtering, status filtering, bulk operations, virtual scroll
- TagsView: tag list, rename, delete
- SettingsView: tab switching, form submission, validation
- LoginView: form submission, error handling, theme toggle
- Target: 80%+ branch coverage on critical paths

### 5.4 E2E Test Setup

Playwright foundation.

- Install Playwright, configure for blog-studio-web
- Test user flows:
  - Login → create post → edit → save → publish → verify
  - Login → search posts → filter → bulk delete
  - Login → settings → change password
  - Login → health check → verify metrics
- CI integration: run E2E on PR (headless Chromium)
- Screenshot on failure for debugging

---

## Phase 6: Performance (3 tasks)

### 6.1 Bundle Optimization

Reduce initial load time.

- Analyze bundle with `rollup-plugin-visualizer`
- Route-based code splitting: lazy load EditorView, SettingsView, HealthView
- Tree-shake unused lucide-vue-next icons (import specific icons, not barrel)
- Optimize CodeMirror: lazy-load language modes on demand
- Target: < 200KB initial JS gzipped (currently ~300KB+)

### 6.2 Image Optimization

Faster asset loading.

- Server-side thumbnail generation on upload (300px max dimension)
- Lazy loading for images in preview (`loading="lazy"`)
- WebP conversion for uploaded JPEG/PNG (Go `golang.org/x/image`)
- Progressive loading: show blurred placeholder → full image
- Serve thumbnails in gallery/list views, full images on click

### 6.3 API Response Caching

Smarter client-side cache.

- Post list: SWR pattern — show cached data immediately, revalidate in background
- Post content: ETag/If-None-Match headers, 304 responses for unchanged content
- Optimistic updates: update UI immediately on save/publish, rollback on error
- Cache invalidation: clear relevant caches on mutations (save, publish, delete)
- Config/health: cache for 60s, refresh on demand

---

## Verification Checklist

After completing all phases:

```bash
# Backend
cd blog-studio-web
go vet ./...
go test -race -cover ./...    # All pass, coverage improved
go build ./cmd/server

# Frontend
cd web
npm run lint                   # No errors
npm run build                  # Success
npx vitest run                 # All pass
npx playwright test            # E2E pass (if setup complete)

# Visual
# - Split view works with live preview
# - All UI animations smooth (60fps)
# - Dark mode complete and polished
# - Mobile responsive on iPhone SE / iPad
# - Empty states display correctly
# - Loading skeletons appear during data fetch
# - Toast notifications work for all types
# - Form validation shows inline errors
```
