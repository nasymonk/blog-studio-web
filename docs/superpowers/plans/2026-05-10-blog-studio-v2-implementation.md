# Blog Studio Web — V2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform blog-studio-web into a polished, production-grade blog management tool with a professional editor experience and cohesive UI design system.

**Architecture:** Progressive enhancement — 6 phases, 38 tasks, each independently shippable. Phase 1 (Editor Core) and Phase 2 (UI Design System) are foundational. Phases 3-6 build on top.

**Tech Stack:** Go 1.25 (backend), Vue 3.5 + TypeScript + Vite 6 (frontend), CodeMirror 6 (editor), shadcn-vue + Tailwind v4 (UI), Vitest (tests), Playwright (E2E), KaTeX (math), Docker + nginx (deployment)

**Repository:** `/Users/xiang/Desktop/personal-server/blog-studio-web/`

---

## File Structure Map

### Frontend — New Files

| File | Responsibility |
|------|---------------|
| `web/src/styles/tokens.css` | Design tokens (CSS custom properties) |
| `web/src/styles/animations.css` | Animation keyframes and utilities |
| `web/src/styles/typography.css` | Typography scale and content styles |
| `web/src/components/SplitView.vue` | Split editor/preview panel |
| `web/src/components/FindReplace.vue` | Inline find & replace bar |
| `web/src/components/EditorToolbar.vue` | Enhanced grouped toolbar |
| `web/src/components/EditorStatusBar.vue` | Word count, reading time, cursor position |
| `web/src/components/VersionHistory.vue` | Version history side panel |
| `web/src/components/ToastContainer.vue` | Custom toast notification system |
| `web/src/components/Toast.vue` | Individual toast component |
| `web/src/components/EmptyState.vue` | Reusable empty state with illustration |
| `web/src/components/SkeletonCard.vue` | Card-shaped loading skeleton |
| `web/src/components/SkeletonLines.vue` | Line-shaped loading skeleton |
| `web/src/components/DataTable.vue` | Reusable sortable data table |
| `web/src/components/Breadcrumb.vue` | Breadcrumb navigation |
| `web/src/composables/useToast.ts` | Toast state management |
| `web/src/composables/useSplitView.ts` | Split view state & resize logic |
| `web/src/composables/useFindReplace.ts` | Find & replace logic |
| `web/src/composables/useEditorSettings.ts` | Editor customization (font, size, theme) |
| `web/src/composables/useVersionHistory.ts` | Version history state |
| `web/src/__tests__/SplitView.test.ts` | Split view tests |
| `web/src/__tests__/FindReplace.test.ts` | Find & replace tests |
| `web/src/__tests__/Toast.test.ts` | Toast system tests |
| `web/src/__tests__/DataTable.test.ts` | Data table tests |
| `web/src/__tests__/EditorView-v2.test.ts` | Enhanced editor tests |
| `web/e2e/smoke.spec.ts` | E2E smoke tests |
| `web/playwright.config.ts` | Playwright configuration |

### Frontend — Modified Files

| File | Changes |
|------|---------|
| `web/src/styles/tailwind.css` | Import tokens, animations, typography; dark mode refinements |
| `web/src/composables/useEditor.ts` | Split view, find/replace, editor themes, status bar integration |
| `web/src/views/EditorView.vue` | Split view, toolbar, status bar, focus mode, version history |
| `web/src/views/PostsView.vue` | Empty states, skeletons, mobile responsive, data table |
| `web/src/views/HealthView.vue` | Skeletons, empty states |
| `web/src/views/LoginView.vue` | Design tokens, animations, mobile responsive |
| `web/src/views/settings/SettingsLayout.vue` | Breadcrumbs, mobile tabs |
| `web/src/App.vue` | Design tokens, mobile sidebar, animations |
| `web/src/i18n/index.ts` | New translation keys |
| `web/src/components/MarkdownPreview.vue` | KaTeX rendering, code themes |
| `web/src/components/EditorOutline.vue` | Scroll sync, active heading highlight |
| `web/src/components/CommandPalette.vue` | Enhanced commands, fuzzy search |
| `web/package.json` | New dependencies (katex, fuse.js, playwright) |

### Backend — New Files

| File | Responsibility |
|------|---------------|
| `internal/httpapi/server_test.go` (expand) | Additional endpoint tests |
| `internal/storage/storage_test.go` | Storage package tests |
| `internal/publish/service_test.go` (expand) | Additional publish tests |
| `internal/httpapi/thumbnail.go` | Image thumbnail generation |

### Backend — Modified Files

| File | Changes |
|------|---------|
| `internal/httpapi/server.go` | Thumbnail endpoint, ETag headers |
| `internal/httpapi/middleware.go` | ETag middleware |
| `internal/publish/service.go` | Version history storage |
| `web/package.json` | Playwright dependency |

---

## Phase 1: Editor Core (8 tasks)

### Task 1.1: Design Tokens & CSS Foundation

**Files:**
- Create: `web/src/styles/tokens.css`
- Create: `web/src/styles/animations.css`
- Create: `web/src/styles/typography.css`
- Modify: `web/src/styles/tailwind.css`
- Modify: `web/src/App.vue`

- [ ] **Step 1: Create design tokens CSS**

```css
/* web/src/styles/tokens.css */
@layer base {
  :root {
    /* Colors — Light */
    --color-primary: oklch(0.55 0.15 250);
    --color-primary-foreground: oklch(0.98 0 0);
    --color-secondary: oklch(0.92 0.01 250);
    --color-secondary-foreground: oklch(0.20 0.01 250);
    --color-background: oklch(0.98 0.005 250);
    --color-foreground: oklch(0.15 0.01 250);
    --color-muted: oklch(0.95 0.005 250);
    --color-muted-foreground: oklch(0.50 0.01 250);
    --color-card: oklch(1 0 0);
    --color-card-foreground: oklch(0.15 0.01 250);
    --color-border: oklch(0.90 0.005 250);
    --color-ring: oklch(0.55 0.15 250);
    --color-destructive: oklch(0.55 0.2 25);
    --color-destructive-foreground: oklch(0.98 0 0);
    --color-success: oklch(0.60 0.18 145);
    --color-success-foreground: oklch(0.98 0 0);
    --color-warning: oklch(0.70 0.15 75);
    --color-warning-foreground: oklch(0.20 0 0);
    --color-info: oklch(0.60 0.12 250);
    --color-info-foreground: oklch(0.98 0 0);
    --color-overlay: oklch(0 0 0 / 0.5);
    --color-sidebar: oklch(0.96 0.005 250);
    --color-sidebar-foreground: oklch(0.15 0.01 250);
    --color-sidebar-accent: oklch(0.93 0.01 250);
    --color-sidebar-accent-foreground: oklch(0.15 0.01 250);

    /* Spacing */
    --space-0: 0px;
    --space-1: 4px;
    --space-2: 8px;
    --space-3: 12px;
    --space-4: 16px;
    --space-5: 20px;
    --space-6: 24px;
    --space-8: 32px;
    --space-10: 40px;
    --space-12: 48px;
    --space-16: 64px;

    /* Typography */
    --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
    --font-serif: 'Noto Serif SC', Georgia, 'Times New Roman', serif;
    --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;
    --font-display: 'Noto Serif SC', Georgia, serif;
    --text-xs: 0.75rem;
    --text-sm: 0.875rem;
    --text-base: 1rem;
    --text-lg: 1.125rem;
    --text-xl: 1.25rem;
    --text-2xl: 1.5rem;
    --text-3xl: 2rem;
    --text-4xl: 2.5rem;
    --leading-tight: 1.25;
    --leading-normal: 1.5;
    --leading-relaxed: 1.625;
    --tracking-tight: -0.025em;
    --tracking-normal: 0;
    --tracking-wide: 0.025em;

    /* Borders */
    --radius-sm: 4px;
    --radius-md: 8px;
    --radius-lg: 12px;
    --radius-xl: 16px;
    --radius-full: 9999px;

    /* Shadows */
    --shadow-sm: 0 1px 2px oklch(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px oklch(0 0 0 / 0.07), 0 2px 4px -2px oklch(0 0 0 / 0.05);
    --shadow-lg: 0 10px 15px -3px oklch(0 0 0 / 0.10), 0 4px 6px -4px oklch(0 0 0 / 0.05);
    --shadow-xl: 0 20px 25px -5px oklch(0 0 0 / 0.15), 0 8px 10px -6px oklch(0 0 0 / 0.05);

    /* Easing */
    --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
    --ease-smooth: cubic-bezier(0.16, 1, 0.3, 1);
    --ease-in: cubic-bezier(0.4, 0, 1, 1);
    --ease-out: cubic-bezier(0, 0, 0.2, 1);

    /* Duration */
    --duration-fast: 100ms;
    --duration-normal: 200ms;
    --duration-slow: 300ms;
    --duration-slower: 500ms;
  }

  .dark {
    --color-primary: oklch(0.65 0.15 250);
    --color-primary-foreground: oklch(0.10 0 0);
    --color-secondary: oklch(0.25 0.01 250);
    --color-secondary-foreground: oklch(0.90 0.01 250);
    --color-background: oklch(0.13 0.01 250);
    --color-foreground: oklch(0.93 0.005 250);
    --color-muted: oklch(0.22 0.01 250);
    --color-muted-foreground: oklch(0.60 0.01 250);
    --color-card: oklch(0.17 0.01 250);
    --color-card-foreground: oklch(0.93 0.005 250);
    --color-border: oklch(0.28 0.01 250);
    --color-ring: oklch(0.65 0.15 250);
    --color-destructive: oklch(0.60 0.2 25);
    --color-destructive-foreground: oklch(0.10 0 0);
    --color-success: oklch(0.65 0.18 145);
    --color-success-foreground: oklch(0.10 0 0);
    --color-warning: oklch(0.75 0.15 75);
    --color-warning-foreground: oklch(0.10 0 0);
    --color-info: oklch(0.65 0.12 250);
    --color-info-foreground: oklch(0.10 0 0);
    --color-overlay: oklch(0 0 0 / 0.7);
    --color-sidebar: oklch(0.15 0.01 250);
    --color-sidebar-foreground: oklch(0.93 0.005 250);
    --color-sidebar-accent: oklch(0.22 0.01 250);
    --color-sidebar-accent-foreground: oklch(0.93 0.005 250);
    --shadow-sm: 0 1px 2px oklch(0 0 0 / 0.2);
    --shadow-md: 0 4px 6px -1px oklch(0 0 0 / 0.3), 0 2px 4px -2px oklch(0 0 0 / 0.2);
    --shadow-lg: 0 10px 15px -3px oklch(0 0 0 / 0.4), 0 4px 6px -4px oklch(0 0 0 / 0.2);
    --shadow-xl: 0 20px 25px -5px oklch(0 0 0 / 0.5), 0 8px 10px -6px oklch(0 0 0 / 0.2);
  }
}
```

- [ ] **Step 2: Create animations CSS**

```css
/* web/src/styles/animations.css */
@layer utilities {
  /* Page transitions */
  .page-enter-active {
    animation: fade-up var(--duration-slow) var(--ease-smooth) both;
  }
  .page-leave-active {
    animation: fade-out var(--duration-fast) var(--ease-in) both;
  }

  /* List stagger */
  .list-enter-active {
    animation: slide-in var(--duration-normal) var(--ease-smooth) both;
  }
  .list-leave-active {
    animation: fade-out var(--duration-fast) var(--ease-in) both;
  }

  /* Modal */
  .modal-enter-active {
    animation: modal-in var(--duration-normal) var(--ease-smooth) both;
  }
  .modal-leave-active {
    animation: modal-out var(--duration-fast) var(--ease-in) both;
  }

  /* Toast */
  .toast-enter-active {
    animation: slide-in-right var(--duration-slow) var(--ease-spring) both;
  }
  .toast-leave-active {
    animation: slide-out-right var(--duration-normal) var(--ease-in) both;
  }

  /* Shimmer for skeletons */
  .shimmer {
    background: linear-gradient(
      90deg,
      var(--color-muted) 25%,
      var(--color-muted-foreground) / 0.05 50%,
      var(--color-muted) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s ease-in-out infinite;
  }

  /* Pulse for saving indicator */
  .pulse-slow {
    animation: pulse-slow 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  /* Shake for validation errors */
  .shake {
    animation: shake 0.3s var(--ease-spring);
  }
}

@keyframes fade-up {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fade-out {
  from { opacity: 1; }
  to { opacity: 0; }
}

@keyframes slide-in {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes slide-in-right {
  from { opacity: 0; transform: translateX(100%); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes slide-out-right {
  from { opacity: 1; transform: translateX(0); }
  to { opacity: 0; transform: translateX(100%); }
}

@keyframes modal-in {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes modal-out {
  from { opacity: 1; transform: scale(1); }
  to { opacity: 0; transform: scale(0.95); }
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

@keyframes pulse-slow {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-4px); }
  40% { transform: translateX(4px); }
  60% { transform: translateX(-2px); }
  80% { transform: translateX(2px); }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .page-enter-active,
  .page-leave-active,
  .list-enter-active,
  .list-leave-active,
  .modal-enter-active,
  .modal-leave-active,
  .toast-enter-active,
  .toast-leave-active {
    animation-duration: 0ms !important;
  }
  .shimmer {
    animation: none;
  }
}
```

- [ ] **Step 3: Create typography CSS**

```css
/* web/src/styles/typography.css */
@layer base {
  /* Heading hierarchy */
  h1, .h1 {
    font-family: var(--font-serif);
    font-size: var(--text-3xl);
    font-weight: 700;
    line-height: var(--leading-tight);
    letter-spacing: var(--tracking-tight);
  }
  h2, .h2 {
    font-family: var(--font-serif);
    font-size: var(--text-2xl);
    font-weight: 600;
    line-height: var(--leading-tight);
    letter-spacing: var(--tracking-tight);
  }
  h3, .h3 {
    font-family: var(--font-serif);
    font-size: var(--text-xl);
    font-weight: 600;
    line-height: var(--leading-normal);
  }
  h4, .h4 {
    font-family: var(--font-sans);
    font-size: var(--text-base);
    font-weight: 600;
    line-height: var(--leading-normal);
  }

  /* Body text */
  body {
    font-family: var(--font-sans);
    font-size: var(--text-base);
    line-height: var(--leading-relaxed);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  /* Code */
  code, .font-mono {
    font-family: var(--font-mono);
    font-size: 0.9em;
  }

  /* CJK mixed text */
  .cjk-text {
    word-break: break-word;
    line-break: anywhere;
  }

  /* Caption / metadata */
  .text-caption {
    font-size: var(--text-xs);
    font-weight: 500;
    color: var(--color-muted-foreground);
    letter-spacing: var(--tracking-wide);
  }
}
```

- [ ] **Step 4: Import new CSS files in tailwind.css**

Add at the top of `web/src/styles/tailwind.css`:
```css
@import './tokens.css';
@import './animations.css';
@import './typography.css';
```

- [ ] **Step 5: Update App.vue to use CSS variables**

Replace hardcoded color references with `var(--color-*)` where applicable. The existing Tailwind classes should work with the new tokens since they map to the same semantic names.

- [ ] **Step 6: Run build to verify**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Expected: SUCCESS

- [ ] **Step 7: Commit**

```bash
git add web/src/styles/tokens.css web/src/styles/animations.css web/src/styles/typography.css web/src/styles/tailwind.css web/src/App.vue
git commit -m "feat: add design tokens, animation system, and typography foundation"
```

---

### Task 1.2: Split View

**Files:**
- Create: `web/src/components/SplitView.vue`
- Create: `web/src/composables/useSplitView.ts`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Create useSplitView composable**

```typescript
// web/src/composables/useSplitView.ts
import { ref, onMounted, onBeforeUnmount } from 'vue'

export type SplitMode = 'source' | 'preview' | 'split'

export function useSplitView() {
  const mode = ref<SplitMode>(
    (localStorage.getItem('editor:splitMode') as SplitMode) || 'source'
  )
  const splitRatio = ref(0.5)
  const isDragging = ref(false)

  function setMode(newMode: SplitMode) {
    mode.value = newMode
    localStorage.setItem('editor:splitMode', newMode)
  }

  function toggleMode() {
    const modes: SplitMode[] = ['source', 'split', 'preview']
    const idx = modes.indexOf(mode.value)
    setMode(modes[(idx + 1) % modes.length])
  }

  function onDragStart(e: MouseEvent) {
    isDragging.value = true
    const startX = e.clientX
    const startRatio = splitRatio.value
    const container = (e.target as HTMLElement).parentElement!
    const containerWidth = container.getBoundingClientRect().width

    function onMove(e: MouseEvent) {
      const delta = e.clientX - startX
      splitRatio.value = Math.max(0.2, Math.min(0.8, startRatio + delta / containerWidth))
    }

    function onUp() {
      isDragging.value = false
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  return { mode, splitRatio, isDragging, setMode, toggleMode, onDragStart }
}
```

- [ ] **Step 2: Create SplitView component**

```vue
<!-- web/src/components/SplitView.vue -->
<script setup lang="ts">
import type { SplitMode } from '@/composables/useSplitView'

const props = defineProps<{
  mode: SplitMode
  splitRatio: number
  isDragging: boolean
}>()

const emit = defineEmits<{
  'update:splitRatio': [value: number]
  dragStart: [e: MouseEvent]
}>()
</script>

<template>
  <div class="flex h-full overflow-hidden" :class="{ 'select-none': isDragging }">
    <!-- Source panel -->
    <div
      v-show="mode !== 'preview'"
      class="h-full overflow-auto"
      :style="{ width: mode === 'split' ? `${splitRatio * 100}%` : '100%' }"
    >
      <slot name="source" />
    </div>

    <!-- Divider -->
    <div
      v-if="mode === 'split'"
      class="w-1 shrink-0 cursor-col-resize bg-border hover:bg-accent transition-colors"
      :class="{ 'bg-accent': isDragging }"
      @mousedown="emit('dragStart', $event)"
    />

    <!-- Preview panel -->
    <div
      v-show="mode !== 'source'"
      class="h-full overflow-auto bg-background"
      :style="{ width: mode === 'split' ? `${(1 - splitRatio) * 100}%` : '100%' }"
    >
      <slot name="preview" />
    </div>
  </div>
</template>
```

- [ ] **Step 3: Add i18n keys**

In `web/src/i18n/index.ts`, add to both `zh` and `en`:
```typescript
// zh
splitView: '分屏视图',
sourceOnly: '仅源码',
previewOnly: '仅预览',
splitMode: '分屏',
// en
splitView: 'Split View',
sourceOnly: 'Source Only',
previewOnly: 'Preview Only',
splitMode: 'Split',
```

- [ ] **Step 4: Integrate SplitView in EditorView**

In `web/src/views/EditorView.vue`:
- Import `SplitView` and `useSplitView`
- Replace the direct editor container with `<SplitView>` wrapping both the editor and `<MarkdownPreview>`
- Add split mode toggle button to toolbar
- Wire up `MarkdownPreview` to receive `body` content

- [ ] **Step 5: Run build**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Expected: SUCCESS

- [ ] **Step 6: Commit**

```bash
git add web/src/components/SplitView.vue web/src/composables/useSplitView.ts web/src/views/EditorView.vue web/src/i18n/index.ts
git commit -m "feat: add split view with resizable divider and mode toggle"
```

---

### Task 1.3: Outline Panel Enhancement

**Files:**
- Modify: `web/src/components/EditorOutline.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Enhance EditorOutline with scroll sync**

Read `web/src/components/EditorOutline.vue` first. Rewrite to:
- Accept `headings` prop (array of `{ level: number, text: string, line: number }`)
- Accept `activeLine` prop for scroll sync highlighting
- Emit `jump` event with line number
- Tree structure with indentation based on heading level
- Active heading highlighted with accent color
- Smooth scroll to heading on click

```vue
<!-- web/src/components/EditorOutline.vue (enhanced) -->
<script setup lang="ts">
export interface Heading {
  level: number
  text: string
  line: number
}

defineProps<{
  headings: Heading[]
  activeLine: number
}>()

const emit = defineEmits<{
  jump: [line: number]
}>()

function isActive(heading: Heading, index: number, headings: Heading[]): boolean {
  // Find the heading that contains the active line
  for (let i = index; i < headings.length - 1; i++) {
    if (headings[i + 1].line > heading.line) return true
  }
  return index === headings.length - 1
}
</script>

<template>
  <nav class="w-48 shrink-0 border-r border-border bg-sidebar overflow-y-auto py-3" aria-label="文档大纲">
    <div class="px-3 mb-2 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">
      {{ headings.length === 0 ? '无标题' : '大纲' }}
    </div>
    <div v-if="headings.length === 0" class="px-3 text-xs text-muted-foreground/40">
      使用 # 创建标题
    </div>
    <button
      v-for="(heading, i) in headings"
      :key="i"
      class="w-full text-left px-3 py-1 text-xs truncate transition-colors hover:bg-sidebar-accent/50"
      :class="[
        heading.level === 1 && 'font-semibold',
        heading.level === 2 && 'pl-5',
        heading.level >= 3 && 'pl-8 text-muted-foreground',
        activeLine >= heading.line && (i === headings.length - 1 || headings[i + 1].line > activeLine)
          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
          : 'text-sidebar-foreground',
      ]"
      @click="emit('jump', heading.line)"
    >
      {{ heading.text }}
    </button>
  </nav>
</template>
```

- [ ] **Step 2: Extract headings from editor in useEditor.ts**

Add to `web/src/composables/useEditor.ts`:
```typescript
const headings = ref<Array<{ level: number, text: string, line: number }>>([])
const activeLine = ref(0)

function extractHeadings(doc: string) {
  const result: Array<{ level: number, text: string, line: number }> = []
  const lines = doc.split('\n')
  for (let i = 0; i < lines.length; i++) {
    const match = lines[i].match(/^(#{1,6})\s+(.+)$/)
    if (match) {
      result.push({ level: match[1].length, text: match[2].trim(), line: i + 1 })
    }
  }
  headings.value = result
}

function updateActiveLine() {
  if (!view) return
  const pos = view.state.selection.main.head
  activeLine.value = view.state.doc.lineAt(pos).number
}

// In the updateListener, add:
// updateActiveLine()
// if (update.docChanged) extractHeadings(update.state.doc.toString())
```

Export `headings` and `activeLine` from useEditor.

- [ ] **Step 3: Wire up in EditorView**

Pass `headings`, `activeLine` to `EditorOutline`, handle `jump` by calling `goToLine`.

- [ ] **Step 4: Commit**

```bash
git add web/src/components/EditorOutline.vue web/src/composables/useEditor.ts web/src/views/EditorView.vue
git commit -m "feat: enhance outline panel with scroll sync and heading extraction"
```

---

### Task 1.4: Math/KaTeX Preview

**Files:**
- Modify: `web/src/components/MarkdownPreview.vue`
- Modify: `web/package.json`

- [ ] **Step 1: Install KaTeX**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install katex`

- [ ] **Step 2: Add KaTeX rendering to MarkdownPreview**

Read `web/src/components/MarkdownPreview.vue`. Add KaTeX rendering:

```typescript
import katex from 'katex'
import 'katex/dist/katex.min.css'

// Custom renderer for math blocks in marked
const renderer = {
  // ... existing renderer methods
  codespan(code: string) {
    // Check if it's inline math
    if (code.startsWith('$') && code.endsWith('$')) {
      try {
        return katex.renderToString(code.slice(1, -1), { throwOnError: false, displayMode: false })
      } catch { return `<code>${code}</code>` }
    }
    return `<code>${code}</code>`
  },
}

// For block math ($$...$$), add a custom extension to marked:
// Process $$ blocks before rendering
function preprocessMath(text: string): string {
  // Block math: $$...$$
  text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_, math) => {
    try {
      return `<div class="math-block">${katex.renderToString(math.trim(), { throwOnError: false, displayMode: true })}</div>`
    } catch (e: any) {
      return `<div class="math-block math-error">${math}</div>`
    }
  })
  return text
}
```

- [ ] **Step 3: Add math styles to tailwind.css**

```css
/* web/src/styles/tailwind.css — add to .markdown-body */
.math-block {
  @apply my-4 overflow-x-auto text-center;
}
.math-error {
  @apply text-destructive border border-destructive/20 rounded p-2;
}
```

- [ ] **Step 4: Run build**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add web/src/components/MarkdownPreview.vue web/src/styles/tailwind.css web/package.json web/package-lock.json
git commit -m "feat: add KaTeX math rendering in preview pane"
```

---

### Task 1.5: Code Block Syntax Themes

**Files:**
- Create: `web/src/composables/useEditorSettings.ts`
- Modify: `web/src/composables/useEditor.ts`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Create useEditorSettings composable**

```typescript
// web/src/composables/useEditorSettings.ts
import { ref, watch } from 'vue'

export type CodeTheme = 'oneDark' | 'githubLight' | 'catppuccin' | 'solarized'

export interface EditorSettings {
  fontFamily: string
  fontSize: number
  lineHeight: number
  lineNumbers: boolean
  codeTheme: CodeTheme
}

const defaults: EditorSettings = {
  fontFamily: 'system',
  fontSize: 14,
  lineHeight: 1.6,
  lineNumbers: true,
  codeTheme: 'oneDark',
}

export function useEditorSettings() {
  const stored = localStorage.getItem('editor:settings')
  const settings = ref<EditorSettings>(stored ? { ...defaults, ...JSON.parse(stored) } : defaults)

  watch(settings, (val) => {
    localStorage.setItem('editor:settings', JSON.stringify(val))
  }, { deep: true })

  function applyPreset(preset: 'default' | 'typewriter' | 'compact') {
    switch (preset) {
      case 'typewriter':
        settings.value = { ...settings.value, fontFamily: 'serif', fontSize: 18, lineHeight: 2.0 }
        break
      case 'compact':
        settings.value = { ...settings.value, fontFamily: 'system', fontSize: 13, lineHeight: 1.4 }
        break
      default:
        settings.value = { ...settings.value, ...defaults }
    }
  }

  return { settings, applyPreset }
}
```

- [ ] **Step 2: Apply code theme in useEditor.ts**

Import code themes and use Compartment for live switching:

```typescript
import { oneDark } from '@codemirror/theme-one-dark'

// In mount(), add a codeThemeCompartment
const codeThemeCompartment = new Compartment()

// In the extensions array:
codeThemeCompartment.of(settings.value.codeTheme === 'oneDark' ? oneDark : []),

// Watch for theme changes:
watch(() => settings.value.codeTheme, (theme) => {
  if (view) {
    view.dispatch({
      effects: codeThemeCompartment.reconfigure(theme === 'oneDark' ? oneDark : []),
    })
  }
})
```

- [ ] **Step 3: Add i18n keys for editor settings**

```typescript
// zh
editorSettings: '编辑器设置',
fontFamily: '字体',
fontSize: '字号',
lineHeight: '行高',
lineNumbers: '行号',
codeTheme: '代码主题',
preset: '预设',
// en
editorSettings: 'Editor Settings',
fontFamily: 'Font Family',
fontSize: 'Font Size',
lineHeight: 'Line Height',
lineNumbers: 'Line Numbers',
codeTheme: 'Code Theme',
preset: 'Preset',
```

- [ ] **Step 4: Add settings panel to EditorView**

Add an editor settings dropdown/popover accessible from toolbar with controls for font, size, theme, and presets.

- [ ] **Step 5: Commit**

```bash
git add web/src/composables/useEditorSettings.ts web/src/composables/useEditor.ts web/src/views/EditorView.vue web/src/i18n/index.ts
git commit -m "feat: add editor customization (font, size, code theme, presets)"
```

---

### Task 1.6: Auto-save UX Polish

**Files:**
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Add save status composable**

Add to `web/src/composables/useEditor.ts`:
```typescript
const saveStatus = ref<'idle' | 'saving' | 'saved' | 'unsaved'>('idle')
const lastSavedTime = ref<string>('')

function updateSaveStatus() {
  if (saving.value) {
    saveStatus.value = 'saving'
  } else if (dirty.value) {
    saveStatus.value = 'unsaved'
  } else if (savedAt.value) {
    saveStatus.value = 'saved'
    lastSavedTime.value = savedAt.value.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
}

// Watch dirty and saving to update status
watch([dirty, saving, savedAt], updateSaveStatus)
```

- [ ] **Step 2: Create save status indicator in EditorView**

Replace the existing save badge in the header with a more polished indicator:

```vue
<div class="flex items-center gap-2 text-xs">
  <template v-if="saveStatus === 'saving'">
    <Loader2Icon class="h-3 w-3 animate-spin text-muted-foreground" />
    <span class="text-muted-foreground">{{ t.saving }}</span>
  </template>
  <template v-else-if="saveStatus === 'unsaved'">
    <div class="h-2 w-2 rounded-full bg-warning animate-pulse-slow" />
    <span class="text-warning">{{ t.unsaved }}</span>
  </template>
  <template v-else-if="saveStatus === 'saved'">
    <div class="h-2 w-2 rounded-full bg-success" />
    <span class="text-muted-foreground">{{ t.saved }} {{ lastSavedTime }}</span>
  </template>
</div>
```

- [ ] **Step 3: Add subtle save flash animation**

When save completes successfully, briefly flash the save indicator green:

```css
.save-flash {
  animation: save-flash 0.6s ease-out;
}
@keyframes save-flash {
  0% { background-color: var(--color-success); opacity: 0.3; }
  100% { background-color: transparent; opacity: 1; }
}
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/EditorView.vue web/src/composables/useEditor.ts web/src/i18n/index.ts
git commit -m "feat: polished auto-save UX with status indicator and flash animation"
```

---

### Task 1.7: Find and Replace

**Files:**
- Create: `web/src/components/FindReplace.vue`
- Create: `web/src/composables/useFindReplace.ts`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Create useFindReplace composable**

```typescript
// web/src/composables/useFindReplace.ts
import { ref, computed } from 'vue'
import type { EditorView } from '@codemirror/view'
import { SearchQuery, setSearchQuery, findNext, findPrevious, replaceNext,replaceAll } from '@codemirror/search'

export function useFindReplace(view: () => EditorView | null) {
  const open = ref(false)
  const searchText = ref('')
  const replaceText = ref('')
  const caseSensitive = ref(false)
  const wholeWord = ref(false)
  const useRegex = ref(false)
  const matchCount = ref(0)
  const currentMatch = ref(0)

  function openFind() {
    open.value = true
    const v = view()
    if (v) {
      const sel = v.state.sliceDoc(
        v.state.selection.main.from,
        v.state.selection.main.to
      )
      if (sel) searchText.value = sel
    }
  }

  function close() {
    open.value = false
    searchText.value = ''
    replaceText.value = ''
  }

  function updateQuery() {
    const v = view()
    if (!v) return
    const query = new SearchQuery({
      search: searchText.value,
      caseSensitive: caseSensitive.value,
      wholeWord: wholeWord.value,
      regexp: useRegex.value,
    })
    v.dispatch({ effects: setSearchQuery.of(query) })
  }

  function findNextMatch() {
    const v = view()
    if (v) findNext(v)
  }

  function findPrevMatch() {
    const v = view()
    if (v) findPrevious(v)
  }

  function replaceMatch() {
    const v = view()
    if (v) replaceNext(v)
  }

  function replaceAllMatches() {
    const v = view()
    if (v) replaceAll(v)
  }

  return {
    open, searchText, replaceText, caseSensitive, wholeWord, useRegex,
    matchCount, currentMatch,
    openFind, close, updateQuery, findNextMatch, findPrevMatch, replaceMatch, replaceAllMatches,
  }
}
```

- [ ] **Step 2: Create FindReplace component**

```vue
<!-- web/src/components/FindReplace.vue -->
<script setup lang="ts">
import { XIcon, ChevronUpIcon, ChevronDownIcon, ReplaceIcon } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

defineProps<{
  searchText: string
  replaceText: string
  caseSensitive: boolean
  wholeWord: boolean
  useRegex: boolean
}>()

const emit = defineEmits<{
  'update:searchText': [value: string]
  'update:replaceText': [value: string]
  'update:caseSensitive': [value: boolean]
  'update:wholeWord': [value: boolean]
  'update:useRegex': [value: boolean]
  findNext: []
  findPrev: []
  replace: []
  replaceAll: []
  close: []
}>()
</script>

<template>
  <div class="flex items-center gap-2 px-3 py-2 bg-muted/50 border-b border-border animate-fade-up">
    <div class="flex items-center gap-1 flex-1">
      <Input
        :value="searchText"
        @update:model-value="emit('update:searchText', $event)"
        placeholder="搜索..."
        class="h-7 text-sm"
        @keydown.enter.prevent="emit('findNext')"
        @keydown.escape="emit('close')"
      />
      <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': caseSensitive }" @click="emit('update:caseSensitive', !caseSensitive)">
        Aa
      </Button>
      <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': wholeWord }" @click="emit('update:wholeWord', !wholeWord)">
        Ab
      </Button>
      <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': useRegex }" @click="emit('update:useRegex', !useRegex)">
        .*
      </Button>
    </div>
    <div class="flex items-center gap-1">
      <Button variant="ghost" size="sm" class="h-7 w-7 p-0" @click="emit('findPrev')">
        <ChevronUpIcon class="h-3.5 w-3.5" />
      </Button>
      <Button variant="ghost" size="sm" class="h-7 w-7 p-0" @click="emit('findNext')">
        <ChevronDownIcon class="h-3.5 w-3.5" />
      </Button>
    </div>
    <div class="flex items-center gap-1">
      <Input
        :value="replaceText"
        @update:model-value="emit('update:replaceText', $event)"
        placeholder="替换..."
        class="h-7 text-sm w-32"
      />
      <Button variant="ghost" size="sm" class="h-7 text-xs" @click="emit('replace')">替换</Button>
      <Button variant="ghost" size="sm" class="h-7 text-xs" @click="emit('replaceAll')">全部</Button>
    </div>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" @click="emit('close')">
      <XIcon class="h-3.5 w-3.5" />
    </Button>
  </div>
</template>
```

- [ ] **Step 3: Wire up in useEditor and EditorView**

Add `findReplace` to useEditor return. Add keyboard shortcut `Mod-f` in useEditor keymap:

```typescript
{ key: 'Mod-f', run: () => { findReplace.openFind(); return true } },
```

In EditorView, import FindReplace and render it above the editor when `findReplace.open` is true.

- [ ] **Step 4: Run build**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add web/src/components/FindReplace.vue web/src/composables/useFindReplace.ts web/src/composables/useEditor.ts web/src/views/EditorView.vue
git commit -m "feat: add inline find and replace bar with regex support"
```

---

### Task 1.8: Editor Status Bar

**Files:**
- Create: `web/src/components/EditorStatusBar.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Create EditorStatusBar component**

```vue
<!-- web/src/components/EditorStatusBar.vue -->
<script setup lang="ts">
defineProps<{
  wordCount: number
  charCount: number
  lineCount: number
  cursorLine: number
  cursorCol: number
  readingTime: string
}>()
</script>

<template>
  <div class="flex items-center gap-4 px-4 py-1.5 text-[11px] text-muted-foreground border-t border-border bg-muted/30">
    <span>{{ wordCount }} 词</span>
    <span>{{ charCount }} 字符</span>
    <span>{{ lineCount }} 行</span>
    <span class="ml-auto">行 {{ cursorLine }}, 列 {{ cursorCol }}</span>
    <span>阅读 ~{{ readingTime }}</span>
  </div>
</template>
```

- [ ] **Step 2: Track cursor position in useEditor**

Add to useEditor:
```typescript
const cursorLine = ref(1)
const cursorCol = ref(1)
const charCount = ref(0)
const lineCount = ref(0)

function updateCursorInfo() {
  if (!view) return
  const pos = view.state.selection.main.head
  const line = view.state.doc.lineAt(pos)
  cursorLine.value = line.number
  cursorCol.value = pos - line.from + 1
  charCount.value = view.state.doc.length
  lineCount.value = view.state.doc.lines
}

const readingTime = computed(() => {
  const engWords = wordCount.value
  const minutes = Math.ceil(engWords / 200)
  return minutes < 1 ? '< 1 分钟' : `${minutes} 分钟`
})
```

- [ ] **Step 3: Add to EditorView**

Import EditorStatusBar, place below the editor container.

- [ ] **Step 4: Commit**

```bash
git add web/src/components/EditorStatusBar.vue web/src/composables/useEditor.ts web/src/views/EditorView.vue
git commit -m "feat: add editor status bar with word count, cursor position, reading time"
```

---

## Phase 2: UI Design System (remaining tasks build on tokens from 1.1)

### Task 2.1: Micro-interactions

**Files:**
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add micro-interaction styles**

```css
/* web/src/styles/tailwind.css — add @layer utilities */

/* Button press */
.btn-press {
  transition: transform var(--duration-fast) var(--ease-spring), box-shadow var(--duration-normal);
}
.btn-press:active {
  transform: scale(0.97);
}

/* Card hover */
.card-hover {
  transition: transform var(--duration-normal), box-shadow var(--duration-normal);
}
.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

/* Focus ring */
.focus-ring {
  transition: box-shadow var(--duration-fast);
}
.focus-ring:focus-visible {
  outline: 2px solid var(--color-ring);
  outline-offset: 2px;
}

/* Link underline slide */
.link-slide {
  position: relative;
}
.link-slide::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 1px;
  background: currentColor;
  transition: width var(--duration-normal) var(--ease-smooth);
}
.link-slide:hover::after {
  width: 100%;
}

/* Toggle spring */
.toggle-spring {
  transition: background-color var(--duration-slow) var(--ease-spring);
}

/* Input focus */
.input-focus {
  transition: border-color var(--duration-normal), box-shadow var(--duration-normal);
}
.input-focus:focus {
  border-color: var(--color-ring);
  box-shadow: 0 0 0 2px oklch(from var(--color-ring) l c h / 0.2);
}
```

- [ ] **Step 2: Apply to existing components**

Apply `card-hover` to post cards in PostsView, `btn-press` to buttons, `input-focus` to inputs, `focus-ring` to all interactive elements.

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css web/src/views/PostsView.vue
git commit -m "feat: add micro-interactions (button press, card hover, focus rings, link slides)"
```

---

### Task 2.2: Mobile Responsive Layout

**Files:**
- Modify: `web/src/App.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/views/PostsView.vue`

- [ ] **Step 1: Add mobile sidebar drawer**

In `web/src/App.vue`:
- Add a hamburger button visible on `< md` (768px)
- Sidebar: fixed overlay on mobile with backdrop, slide-in from left
- Add `md:hidden` / `hidden md:flex` classes for responsive visibility

```vue
<!-- Mobile hamburger -->
<Button variant="ghost" size="sm" class="md:hidden h-8 w-8 p-0" @click="sidebarOpen = true">
  <MenuIcon class="h-4 w-4" />
</Button>

<!-- Sidebar: hidden on mobile, shown on desktop -->
<!-- Mobile: fixed overlay -->
<aside
  v-if="sidebarOpen"
  class="fixed inset-0 z-50 md:hidden"
>
  <div class="absolute inset-0 bg-black/50" @click="sidebarOpen = false" />
  <aside class="relative w-[280px] h-full bg-sidebar border-r ...">
    <!-- existing sidebar content -->
  </aside>
</aside>

<!-- Desktop sidebar -->
<aside class="hidden md:flex w-[240px] shrink-0 border-r border-border bg-sidebar flex-col">
  <!-- existing sidebar content -->
</aside>
```

- [ ] **Step 2: Make EditorView mobile-friendly**

- Toolbar: `flex-wrap` on mobile, or pin to bottom as a compact bar
- Editor: full width on mobile (hide outline panel)
- Meta section: collapsible, hidden by default on mobile

- [ ] **Step 3: Make PostsView mobile-friendly**

- Cards: full width on mobile
- Search/filter bar: stack vertically on mobile

- [ ] **Step 4: Commit**

```bash
git add web/src/App.vue web/src/views/EditorView.vue web/src/views/PostsView.vue
git commit -m "feat: mobile responsive layout (sidebar drawer, stacked toolbar, touch targets)"
```

---

### Task 2.3: Dark Mode Refinements

**Files:**
- Modify: `web/src/styles/tailwind.css`
- Modify: `web/src/composables/useTheme.ts`
- Modify: `web/src/components/MarkdownPreview.vue`

- [ ] **Step 1: Add theme transition**

```css
/* web/src/styles/tailwind.css */
body {
  transition: background-color var(--duration-normal), color var(--duration-normal);
}
```

- [ ] **Step 2: Fix dark mode contrast issues**

Audit all views in dark mode. Add to tailwind.css:
```css
.dark img {
  filter: brightness(0.9);
}
.dark .markdown-body {
  color: var(--color-foreground);
}
```

- [ ] **Step 3: Respect prefers-color-scheme**

In `web/src/composables/useTheme.ts`, default to system preference:
```typescript
const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
const stored = localStorage.getItem('theme')
const theme = ref<'light' | 'dark'>(stored ? stored as any : systemDark ? 'dark' : 'light')
```

- [ ] **Step 4: Commit**

```bash
git add web/src/styles/tailwind.css web/src/composables/useTheme.ts web/src/components/MarkdownPreview.vue
git commit -m "feat: dark mode refinements (contrast, transitions, system preference)"
```

---

### Task 2.4: Empty States

**Files:**
- Create: `web/src/components/EmptyState.vue`
- Modify: `web/src/views/PostsView.vue`
- Modify: `web/src/views/TrashView.vue`
- Modify: `web/src/views/HealthView.vue`

- [ ] **Step 1: Create EmptyState component**

```vue
<!-- web/src/components/EmptyState.vue -->
<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  description?: string
  action?: { label: string, handler: () => void }
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center animate-fade-up">
    <div class="mb-4 text-muted-foreground/30">
      <slot name="icon">
        <svg class="h-16 w-16" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="8" y="12" width="48" height="40" rx="4" />
          <line x1="8" y1="24" x2="56" y2="24" />
          <line x1="20" y1="32" x2="44" y2="32" />
          <line x1="20" y1="40" x2="36" y2="40" />
        </svg>
      </slot>
    </div>
    <h3 class="font-serif text-lg font-semibold text-foreground/80 mb-1">{{ title }}</h3>
    <p v-if="description" class="text-sm text-muted-foreground max-w-sm mb-4">{{ description }}</p>
    <button
      v-if="action"
      class="px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
      @click="action.handler"
    >
      {{ action.label }}
    </button>
  </div>
</template>
```

- [ ] **Step 2: Add empty states to views**

- PostsView: show EmptyState when no posts with "Write your first post" CTA
- TrashView: show EmptyState when trash is empty
- HealthView: show EmptyState if health check fails to load

- [ ] **Step 3: Add i18n keys**

```typescript
// zh
noPosts: '还没有文章',
noPostsDesc: '创建你的第一篇文章开始写作吧',
noSearchResults: '没有找到匹配的文章',
noSearchResultsDesc: '试试不同的关键词',
trashEmpty: '回收站是空的',
trashEmptyDesc: '删除的文章会出现在这里',
// en
noPosts: 'No posts yet',
noPostsDesc: 'Create your first post to start writing',
noSearchResults: 'No matching posts found',
noSearchResultsDesc: 'Try different keywords',
trashEmpty: 'Trash is empty',
trashEmptyDesc: 'Deleted posts will appear here',
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/EmptyState.vue web/src/views/PostsView.vue web/src/views/TrashView.vue web/src/views/HealthView.vue web/src/i18n/index.ts
git commit -m "feat: add empty states with illustrations and CTAs"
```

---

### Task 2.5: Loading Skeletons

**Files:**
- Create: `web/src/components/SkeletonCard.vue`
- Create: `web/src/components/SkeletonLines.vue`
- Modify: `web/src/views/PostsView.vue`
- Modify: `web/src/views/EditorView.vue`

- [ ] **Step 1: Create skeleton components**

```vue
<!-- web/src/components/SkeletonCard.vue -->
<template>
  <div class="rounded-lg border border-border p-4 space-y-3">
    <div class="h-4 w-3/4 rounded shimmer" />
    <div class="h-3 w-1/2 rounded shimmer" />
    <div class="flex gap-2 mt-2">
      <div class="h-5 w-16 rounded-full shimmer" />
      <div class="h-5 w-12 rounded-full shimmer" />
    </div>
  </div>
</template>

<!-- web/src/components/SkeletonLines.vue -->
<script setup lang="ts">
defineProps<{ lines?: number }>()
</script>
<template>
  <div class="space-y-2 p-4">
    <div v-for="i in (lines ?? 5)" :key="i"
      class="h-3 rounded shimmer"
      :style="{ width: `${60 + Math.random() * 30}%` }"
    />
  </div>
</template>
```

- [ ] **Step 2: Replace loading spinners**

In PostsView, show `SkeletonCard` list when `postsLoading` is true.
In EditorView, show `SkeletonLines` when `loading` is true.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/SkeletonCard.vue web/src/components/SkeletonLines.vue web/src/views/PostsView.vue web/src/views/EditorView.vue
git commit -m "feat: add loading skeletons for posts list and editor"
```

---

### Task 2.6: Toast Notification System

**Files:**
- Create: `web/src/components/Toast.vue`
- Create: `web/src/components/ToastContainer.vue`
- Create: `web/src/composables/useToast.ts`
- Modify: `web/src/App.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Create useToast composable**

```typescript
// web/src/composables/useToast.ts
import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  description?: string
  action?: { label: string, handler: () => void }
  duration: number
}

const toasts = ref<Toast[]>([])

export function useToast() {
  function add(toast: Omit<Toast, 'id'>) {
    const id = Math.random().toString(36).slice(2)
    const duration = toast.type === 'error' ? Infinity : (toast.duration ?? 5000)
    toasts.value.push({ ...toast, id, duration })
    if (toasts.value.length > 3) toasts.value.shift()
    if (duration !== Infinity) {
      setTimeout(() => remove(id), duration)
    }
  }

  function remove(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {
    toasts,
    success: (title: string, desc?: string) => add({ type: 'success', title, description: desc, duration: 5000 }),
    error: (title: string, desc?: string) => add({ type: 'error', title, description: desc, duration: Infinity }),
    warning: (title: string, desc?: string) => add({ type: 'warning', title, description: desc, duration: 6000 }),
    info: (title: string, desc?: string) => add({ type: 'info', title, description: desc, duration: 5000 }),
    remove,
  }
}
```

- [ ] **Step 2: Create Toast component**

```vue
<!-- web/src/components/Toast.vue -->
<script setup lang="ts">
import { CheckCircleIcon, XCircleIcon, AlertTriangleIcon, InfoIcon, XIcon } from 'lucide-vue-next'
import type { Toast } from '@/composables/useToast'

defineProps<{ toast: Toast }>()
defineEmits<{ dismiss: [id: string] }>()

const icons = {
  success: CheckCircleIcon,
  error: XCircleIcon,
  warning: AlertTriangleIcon,
  info: InfoIcon,
}

const colors = {
  success: 'border-success/30 bg-success/10',
  error: 'border-destructive/30 bg-destructive/10',
  warning: 'border-warning/30 bg-warning/10',
  info: 'border-info/30 bg-info/10',
}
</script>

<template>
  <div
    class="flex items-start gap-3 p-4 rounded-lg border shadow-md max-w-sm toast-enter-active"
    :class="colors[toast.type]"
  >
    <component :is="icons[toast.type]" class="h-5 w-5 shrink-0 mt-0.5" />
    <div class="flex-1 min-w-0">
      <div class="text-sm font-medium">{{ toast.title }}</div>
      <div v-if="toast.description" class="text-xs text-muted-foreground mt-0.5">{{ toast.description }}</div>
      <button
        v-if="toast.action"
        class="text-xs font-medium text-primary mt-1 hover:underline"
        @click="toast.action.handler"
      >
        {{ toast.action.label }}
      </button>
    </div>
    <button class="shrink-0 opacity-50 hover:opacity-100" @click="$emit('dismiss', toast.id)">
      <XIcon class="h-4 w-4" />
    </button>
  </div>
</template>
```

- [ ] **Step 3: Create ToastContainer**

```vue
<!-- web/src/components/ToastContainer.vue -->
<script setup lang="ts">
import { useToast } from '@/composables/useToast'
import ToastComponent from './Toast.vue'

const { toasts, remove } = useToast()
</script>

<template>
  <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2">
    <TransitionGroup name="toast">
      <ToastComponent
        v-for="toast in toasts"
        :key="toast.id"
        :toast="toast"
        @dismiss="remove"
      />
    </TransitionGroup>
  </div>
</template>
```

- [ ] **Step 4: Replace vue-sonner in App.vue**

Remove `<Toaster>` from vue-sonner, add `<ToastContainer />`. Update all `toast.success/error/warning` calls to use the new `useToast()` composable.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/Toast.vue web/src/components/ToastContainer.vue web/src/composables/useToast.ts web/src/App.vue web/src/i18n/index.ts
git commit -m "feat: custom toast notification system (4 types, actions, animations)"
```

---

### Task 2.7: Modal & Dialog Polish

**Files:**
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add dialog animation styles**

```css
/* Override shadcn dialog animations */
[data-state="open"] > .dialog-content {
  animation: modal-in var(--duration-normal) var(--ease-smooth);
}
[data-state="closed"] > .dialog-content {
  animation: modal-out var(--duration-fast) var(--ease-in);
}

/* Backdrop blur */
.dialog-overlay {
  backdrop-filter: blur(4px);
  background: var(--color-overlay);
}

/* Full-screen on mobile */
@media (max-width: 639px) {
  .dialog-content {
    width: 100vw !important;
    height: 100vh !important;
    max-width: none !important;
    max-height: none !important;
    border-radius: 0 !important;
  }
}

/* Destructive confirmation */
.dialog-destructive {
  border-color: var(--color-destructive) / 0.3;
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/styles/tailwind.css
git commit -m "feat: modal/dialog polish (backdrop blur, animations, mobile fullscreen)"
```

---

### Task 2.8: Form & Input Polish

**Files:**
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add form validation styles**

```css
/* Validation error */
.input-error {
  border-color: var(--color-destructive);
  animation: shake 0.3s var(--ease-spring);
}
.input-error:focus {
  box-shadow: 0 0 0 2px oklch(from var(--color-destructive) l c h / 0.2);
}

/* Validation success */
.input-success {
  border-color: var(--color-success);
}

/* Error message */
.field-error {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  margin-top: var(--space-1);
  font-size: var(--text-xs);
  color: var(--color-destructive);
}

/* Auto-resize textarea */
textarea.auto-resize {
  field-sizing: content;
  min-height: 3lh;
  max-height: 10lh;
}
```

- [ ] **Step 2: Apply to LoginView and Settings forms**

Add validation error styles to login form and settings forms. Add shake animation on validation failure.

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css web/src/views/LoginView.vue web/src/views/settings/SecurityTab.vue
git commit -m "feat: form validation UX (error shake, success border, inline messages)"
```

---

### Task 2.9: Breadcrumb Navigation

**Files:**
- Create: `web/src/components/Breadcrumb.vue`
- Modify: `web/src/views/settings/SettingsLayout.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/App.vue`

- [ ] **Step 1: Create Breadcrumb component**

```vue
<!-- web/src/components/Breadcrumb.vue -->
<script setup lang="ts">
import { ChevronRightIcon } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

export interface BreadcrumbItem {
  label: string
  to?: string
}

defineProps<{ items: BreadcrumbItem[] }>()
</script>

<template>
  <nav class="flex items-center gap-1 text-sm text-muted-foreground mb-4" aria-label="面包屑导航">
    <template v-for="(item, i) in items" :key="i">
      <ChevronRightIcon v-if="i > 0" class="h-3 w-3" />
      <RouterLink
        v-if="item.to"
        :to="item.to"
        class="hover:text-foreground transition-colors"
      >
        {{ item.label }}
      </RouterLink>
      <span v-else class="text-foreground font-medium">{{ item.label }}</span>
    </template>
  </nav>
</template>
```

- [ ] **Step 2: Add to SettingsLayout**

```vue
<Breadcrumb :items="[
  { label: t.settings, to: '/settings' },
  { label: currentTabLabel },
]" />
```

- [ ] **Step 3: Add to EditorView (post title as breadcrumb)**

```vue
<Breadcrumb :items="[
  { label: t.posts, to: '/posts' },
  { label: draft?.title || slug },
]" />
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/Breadcrumb.vue web/src/views/settings/SettingsLayout.vue web/src/views/EditorView.vue
git commit -m "feat: breadcrumb navigation for settings and editor"
```

---

## Phase 3: Editor Advanced (6 tasks)

### Task 3.1: Enhanced Markdown Toolbar

**Files:**
- Create: `web/src/components/EditorToolbar.vue`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Create EditorToolbar component**

```vue
<!-- web/src/components/EditorToolbar.vue -->
<script setup lang="ts">
import {
  BoldIcon, ItalicIcon, StrikethroughIcon, CodeIcon, LinkIcon,
  Heading1Icon, Heading2Icon, Heading3Icon, QuoteIcon, ListIcon,
  ListOrderedIcon, ListChecksIcon, ImageIcon, TableIcon, MinusIcon,
  SplitSquareHorizontalIcon, ColumnsIcon, EyeIcon, PenLineIcon,
  SettingsIcon, KeyboardIcon, GalleryHorizontalEndIcon
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import type { SplitMode } from '@/composables/useSplitView'

defineProps<{
  splitMode: SplitMode
  isWysiwyg: boolean
}>()

const emit = defineEmits<{
  bold: []
  italic: []
  strikethrough: []
  code: []
  link: []
  heading: [level: number]
  blockquote: []
  orderedList: []
  unorderedList: []
  taskList: []
  image: []
  table: []
  hr: []
  splitMode: [mode: SplitMode]
  toggleWysiwyg: []
  openGallery: []
  openSettings: []
  openKeybindings: []
}>()

interface ToolbarGroup {
  items: Array<{
    icon: any
    label: string
    shortcut?: string
    action: () => void
    active?: boolean
  }>
}
</script>

<template>
  <div class="flex items-center gap-0.5 px-2 py-1.5 border-b border-border bg-muted/30 flex-wrap">
    <!-- Text group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="粗体 (⌘B)" @click="emit('bold')">
      <BoldIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="斜体 (⌘I)" @click="emit('italic')">
      <ItalicIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="删除线" @click="emit('strikethrough')">
      <StrikethroughIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="行内代码" @click="emit('code')">
      <CodeIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="链接 (⌘K)" @click="emit('link')">
      <LinkIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- Block group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="标题 1" @click="emit('heading', 1)">
      <Heading1Icon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="标题 2" @click="emit('heading', 2)">
      <Heading2Icon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="引用" @click="emit('blockquote')">
      <QuoteIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- List group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="无序列表" @click="emit('unorderedList')">
      <ListIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="有序列表" @click="emit('orderedList')">
      <ListOrderedIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="任务列表" @click="emit('taskList')">
      <ListChecksIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- Insert group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="插入图片" @click="emit('image')">
      <ImageIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="插入表格" @click="emit('table')">
      <TableIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="分割线" @click="emit('hr')">
      <MinusIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- View group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': splitMode === 'source' }" title="仅源码" @click="emit('splitMode', 'source')">
      <PenLineIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': splitMode === 'split' }" title="分屏" @click="emit('splitMode', 'split')">
      <ColumnsIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="{ 'bg-accent': splitMode === 'preview' }" title="仅预览" @click="emit('splitMode', 'preview')">
      <EyeIcon class="h-3.5 w-3.5" />
    </Button>

    <div class="flex-1" />

    <!-- Right-side actions -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="图片库" @click="emit('openGallery')">
      <GalleryHorizontalEndIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="快捷键" @click="emit('openKeybindings')">
      <KeyboardIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" title="编辑器设置" @click="emit('openSettings')">
      <SettingsIcon class="h-3.5 w-3.5" />
    </Button>
  </div>
</template>
```

- [ ] **Step 2: Add execStrikethrough, execBlockquote, execOrderedList, execUnorderedList, execTaskList to useEditor**

```typescript
function execStrikethrough() { if (view) wrapSelection(view, '~~', '~~') }
function execBlockquote() {
  if (!view) return
  const { from } = view.state.selection.main
  const line = view.state.doc.lineAt(from)
  const prefix = line.text.startsWith('> ') ? '' : '> '
  view.dispatch(view.state.update({
    changes: { from: line.from, to: line.to, insert: prefix + line.text.replace(/^>\s*/, '') }
  }))
}
function execOrderedList() { if (view) insertLinePrefix(view, '1. ') }
function execUnorderedList() { if (view) insertLinePrefix(view, '- ') }
function execTaskList() { if (view) insertLinePrefix(view, '- [ ] ') }

function insertLinePrefix(view: EditorView, prefix: string) {
  const { from } = view.state.selection.main
  const line = view.state.doc.lineAt(from)
  view.dispatch(view.state.update({
    changes: { from: line.from, to: line.to, insert: prefix + line.text }
  }))
}
```

- [ ] **Step 3: Replace existing toolbar in EditorView**

Replace the current inline toolbar with `<EditorToolbar>`, wiring up all events.

- [ ] **Step 4: Commit**

```bash
git add web/src/components/EditorToolbar.vue web/src/composables/useEditor.ts web/src/views/EditorView.vue
git commit -m "feat: enhanced grouped markdown toolbar with split view controls"
```

---

### Task 3.2: Command Palette Enhancement

**Files:**
- Modify: `web/src/components/CommandPalette.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Read existing CommandPalette.vue**

Understand the current implementation, then enhance:
- Add "Recent Posts" section (from `store.posts`)
- Add "Quick Actions" section: New Post, Toggle Theme, Go to Settings, Health Check
- Add fuzzy search using simple includes (no library needed for this scale)
- Add icons for each command type
- Keyboard navigation with ↑↓

- [ ] **Step 2: Add i18n keys for command palette**

```typescript
// zh
quickActions: '快速操作',
recentPosts: '最近文章',
newPost: '新建文章',
goToSettings: '前往设置',
toggleTheme: '切换主题',
// en
quickActions: 'Quick Actions',
recentPosts: 'Recent Posts',
newPost: 'New Post',
goToSettings: 'Go to Settings',
toggleTheme: 'Toggle Theme',
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/CommandPalette.vue web/src/i18n/index.ts
git commit -m "feat: enhanced command palette (recent posts, quick actions, fuzzy search)"
```

---

### Task 3.3: Table Editing

**Files:**
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Add table insertion function**

```typescript
function execTable() {
  if (!view) return
  const table = '\n| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n| | | |\n| | | |\n'
  const pos = view.state.selection.main.head
  view.dispatch(view.state.update({
    changes: { from: pos, insert: table },
    selection: { anchor: pos + table.length }
  }))
}

function execHr() {
  if (!view) return
  const pos = view.state.selection.main.head
  const line = view.state.doc.lineAt(pos)
  const insert = '\n---\n'
  view.dispatch(view.state.update({
    changes: { from: line.to, insert },
    selection: { anchor: line.to + insert.length }
  }))
}
```

- [ ] **Step 2: Add Tab navigation in tables**

In the keymap of useEditor, add:
```typescript
{
  key: 'Tab',
  run: (v) => {
    const pos = v.state.selection.main.head
    const line = v.state.doc.lineAt(pos)
    if (line.text.includes('|')) {
      // Find next cell
      const nextPipe = line.text.indexOf('|', pos - line.from + 1)
      if (nextPipe !== -1) {
        v.dispatch({ selection: { anchor: line.from + nextPipe + 1 } })
        return true
      }
    }
    return false // fall through to default tab
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/composables/useEditor.ts
git commit -m "feat: table editing (insert template, Tab navigation between cells)"
```

---

### Task 3.4: Version History

**Files:**
- Create: `web/src/components/VersionHistory.vue`
- Create: `web/src/composables/useVersionHistory.ts`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Create useVersionHistory composable**

```typescript
// web/src/composables/useVersionHistory.ts
import { ref } from 'vue'

export interface Version {
  id: string
  timestamp: Date
  body: string
  wordCount: number
}

export function useVersionHistory(maxVersions = 10) {
  const versions = ref<Version[]>([])
  const isOpen = ref(false)

  function saveVersion(body: string, wordCount: number) {
    const last = versions.value[0]
    // Don't save if body hasn't changed
    if (last && last.body === body) return

    versions.value.unshift({
      id: Date.now().toString(36),
      timestamp: new Date(),
      body,
      wordCount,
    })

    // Keep only maxVersions
    if (versions.value.length > maxVersions) {
      versions.value = versions.value.slice(0, maxVersions)
    }

    // Persist to localStorage
    try {
      const key = `editor:history:${window.location.pathname}`
      localStorage.setItem(key, JSON.stringify(versions.value.map(v => ({
        ...v,
        timestamp: v.timestamp.toISOString(),
      }))))
    } catch { /* storage full, ignore */ }
  }

  function loadHistory(slug: string) {
    try {
      const key = `editor:history:/posts/${slug}`
      const stored = localStorage.getItem(key)
      if (stored) {
        versions.value = JSON.parse(stored).map((v: any) => ({
          ...v,
          timestamp: new Date(v.timestamp),
        }))
      }
    } catch { /* ignore */ }
  }

  function toggle() { isOpen.value = !isOpen.value }

  return { versions, isOpen, saveVersion, loadHistory, toggle }
}
```

- [ ] **Step 2: Create VersionHistory component**

```vue
<!-- web/src/components/VersionHistory.vue -->
<script setup lang="ts">
import { HistoryIcon, RotateCcwIcon, XIcon } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import type { Version } from '@/composables/useVersionHistory'

defineProps<{
  versions: Version[]
  currentBody: string
}>()

const emit = defineEmits<{
  restore: [body: string]
  close: []
}>()

function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return '刚刚'
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return date.toLocaleDateString()
}
</script>

<template>
  <div class="w-72 border-l border-border bg-sidebar overflow-y-auto">
    <div class="flex items-center justify-between px-3 py-2 border-b border-border">
      <div class="flex items-center gap-2 text-sm font-medium">
        <HistoryIcon class="h-4 w-4" />
        版本历史
      </div>
      <Button variant="ghost" size="sm" class="h-6 w-6 p-0" @click="emit('close')">
        <XIcon class="h-3.5 w-3.5" />
      </Button>
    </div>
    <div v-if="versions.length === 0" class="p-4 text-center text-xs text-muted-foreground">
      暂无历史版本
    </div>
    <div v-else class="divide-y divide-border">
      <div
        v-for="version in versions"
        :key="version.id"
        class="px-3 py-2 hover:bg-sidebar-accent/50 cursor-pointer group"
      >
        <div class="flex items-center justify-between">
          <span class="text-xs text-muted-foreground">{{ timeAgo(version.timestamp) }}</span>
          <Button
            variant="ghost"
            size="sm"
            class="h-6 text-xs opacity-0 group-hover:opacity-100 transition-opacity"
            @click="emit('restore', version.body)"
          >
            <RotateCcwIcon class="h-3 w-3 mr-1" />
            恢复
          </Button>
        </div>
        <div class="text-xs text-muted-foreground/60 mt-0.5">
          {{ version.wordCount }} 词
        </div>
        <div class="text-xs text-muted-foreground/40 mt-1 line-clamp-2 font-mono">
          {{ version.body.slice(0, 100) }}...
        </div>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 3: Wire up in EditorView**

- Import and add VersionHistory panel (toggleable via toolbar button)
- On each auto-save, call `saveVersion(body, wordCount)`
- On restore, set `body` to version body and mark dirty

- [ ] **Step 4: Commit**

```bash
git add web/src/components/VersionHistory.vue web/src/composables/useVersionHistory.ts web/src/views/EditorView.vue web/src/i18n/index.ts
git commit -m "feat: version history with localStorage persistence and restore"
```

---

### Task 3.5: Focus Mode

**Files:**
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/src/composables/useEditor.ts`

- [ ] **Step 1: Add focus mode state**

```typescript
// In useEditor.ts
const focusMode = ref(false)
function toggleFocusMode() {
  focusMode.value = !focusMode.value
}
```

- [ ] **Step 2: Apply focus mode styles in EditorView**

When `focusMode` is true:
- Hide sidebar (via App.vue or CSS)
- Hide toolbar
- Hide header
- Center editor content (max-width: 700px, margin: auto)
- Add subtle vignette overlay

```vue
<template>
  <div v-if="focusMode" class="fixed inset-0 z-40 bg-background flex flex-col items-center">
    <div class="w-full max-w-[700px] h-full flex flex-col">
      <div class="flex-1 overflow-auto" ref="editorContainer" />
    </div>
    <button
      class="absolute top-4 right-4 text-muted-foreground hover:text-foreground"
      @click="toggleFocusMode"
    >
      按 Esc 退出专注模式
    </button>
  </div>
</template>
```

- [ ] **Step 3: Add Escape key handler**

```typescript
{ key: 'Escape', run: () => { if (focusMode.value) { toggleFocusMode(); return true } return false } }
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/EditorView.vue web/src/composables/useEditor.ts
git commit -m "feat: focus mode (distraction-free writing with centered content)"
```

---

### Task 3.6: Data Table Component

**Files:**
- Create: `web/src/components/DataTable.vue`
- Modify: `web/src/views/settings/AuditTab.vue`
- Modify: `web/src/views/TrashView.vue`

- [ ] **Step 1: Create DataTable component**

```vue
<!-- web/src/components/DataTable.vue -->
<script setup lang="ts" generic="T extends Record<string, any>">
import { ref, computed } from 'vue'
import { ChevronUpIcon, ChevronDownIcon } from 'lucide-vue-next'

export interface Column<T> {
  key: keyof T & string
  label: string
  sortable?: boolean
  width?: string
  render?: (value: T[keyof T], row: T) => string
}

const props = defineProps<{
  columns: Column<T>[]
  data: T[]
  emptyMessage?: string
}>()

const sortKey = ref<string>('')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(key: string) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

const sorted = computed(() => {
  if (!sortKey.value) return props.data
  return [...props.data].sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    const cmp = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
})
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <table class="w-full text-sm">
      <thead>
        <tr class="bg-muted/50 border-b border-border">
          <th
            v-for="col in columns"
            :key="col.key"
            class="text-left px-3 py-2 font-medium text-muted-foreground"
            :class="{ 'cursor-pointer hover:text-foreground': col.sortable }"
            :style="{ width: col.width }"
            @click="col.sortable && toggleSort(col.key)"
          >
            <div class="flex items-center gap-1">
              {{ col.label }}
              <template v-if="col.sortable && sortKey === col.key">
                <ChevronUpIcon v-if="sortDir === 'asc'" class="h-3 w-3" />
                <ChevronDownIcon v-else class="h-3 w-3" />
              </template>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="data.length === 0">
          <td :colspan="columns.length" class="text-center py-8 text-muted-foreground">
            {{ emptyMessage ?? '暂无数据' }}
          </td>
        </tr>
        <tr
          v-for="(row, i) in sorted"
          :key="i"
          class="border-b border-border last:border-0 hover:bg-muted/30 transition-colors"
        >
          <td v-for="col in columns" :key="col.key" class="px-3 py-2">
            <slot :name="col.key" :row="row" :value="row[col.key]">
              {{ col.render ? col.render(row[col.key], row) : row[col.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
```

- [ ] **Step 2: Apply to AuditTab and TrashView**

Replace existing table markup with `<DataTable>`, defining columns and using slots for custom cell rendering.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/DataTable.vue web/src/views/settings/AuditTab.vue web/src/views/TrashView.vue
git commit -m "feat: reusable DataTable component with sorting"
```

---

## Phase 4: Backend & Testing (4 tasks)

### Task 4.1: Image Thumbnails

**Files:**
- Create: `internal/httpapi/thumbnail.go`
- Modify: `internal/httpapi/server.go`

- [ ] **Step 1: Create thumbnail generator**

```go
// internal/httpapi/thumbnail.go
package httpapi

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

const thumbnailMaxSize = 300

func generateThumbnail(srcPath string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(srcPath))
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	default:
		return "", nil // not an image
	}
	if err != nil {
		return "", err
	}

	// Calculate thumbnail size
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= thumbnailMaxSize && h <= thumbnailMaxSize {
		return "", nil // already small enough
	}

	ratio := float64(thumbnailMaxSize) / float64(max(w, h))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	// Resize
	thumb := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.ApproxBiLinear.Scale(thumb, thumb.Bounds(), img, bounds, draw.Over, nil)

	// Save
	thumbPath := strings.TrimSuffix(srcPath, ext) + "_thumb" + ext
	var buf bytes.Buffer
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80})
	case ".png":
		err = png.Encode(&buf, thumb)
	}
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(thumbPath, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return thumbPath, nil
}
```

- [ ] **Step 2: Generate thumbnail on upload**

In the existing upload handler in `server.go`, after saving the file, call `generateThumbnail`.

- [ ] **Step 3: Add /api/posts/:slug/assets/:name/thumbnail endpoint**

Serve the `_thumb` version of the file if it exists.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/thumbnail.go internal/httpapi/server.go
git commit -m "feat: server-side image thumbnail generation on upload"
```

---

### Task 4.2: httpapi Test Coverage

**Files:**
- Modify: `internal/httpapi/server_test.go`

- [ ] **Step 1: Read existing server_test.go**

Understand the test patterns, then add tests for:
- All CRUD endpoints (save draft, publish, rollback, delete)
- Auth flows (login failure, session check, password change)
- Error responses (400 bad request, 404 not found)
- File upload (valid file, invalid type)
- Middleware (rate limiting, max body size)

- [ ] **Step 2: Write tests**

```go
// Examples of additional tests to add:

func TestSaveDraft_RequiresAuth(t *testing.T) {
	// No auth cookie → 401
}

func TestSaveDraft_ValidDraft(t *testing.T) {
	// Valid draft → 200, file created
}

func TestPublishBlog_ConflictDetection(t *testing.T) {
	// Remote mtime mismatch → conflict error
}

func TestDeletePost_MovesToTrash(t *testing.T) {
	// Delete → 200, file in trash
}

func TestUploadAsset_ValidImage(t *testing.T) {
	// Upload PNG → 200, file saved
}

func TestUploadAsset_OversizedFile(t *testing.T) {
	// File > maxUploadBytes → 413
}

func TestRateLimit_WriteEndpoints(t *testing.T) {
	// 20 rapid POSTs → some get 429
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web && go test -cover ./internal/httpapi/`
Expected: PASS, coverage > 60%

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/server_test.go
git commit -m "test: increase httpapi test coverage to 60%+"
```

---

### Task 4.3: Storage & Publish Test Coverage

**Files:**
- Create: `internal/storage/storage_test.go`
- Modify: `internal/publish/service_test.go`

- [ ] **Step 1: Write storage tests**

```go
// internal/storage/storage_test.go
package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile_Success(t *testing.T) { /* ... */ }
func TestReadFile_NotFound(t *testing.T) { /* ... */ }
func TestWriteFile_CreatesDirectory(t *testing.T) { /* ... */ }
func TestWriteFile_PathTraversal(t *testing.T) { /* ... */ }
func TestDeleteFile_Success(t *testing.T) { /* ... */ }
func TestConcurrentWrite(t *testing.T) { /* ... */ }
```

- [ ] **Step 2: Expand publish tests**

Add tests for: conflict detection, cache invalidation, large file handling.

- [ ] **Step 3: Run tests**

Run: `go test -cover ./internal/storage/ ./internal/publish/`
Expected: storage > 60%, publish > 70%

- [ ] **Step 4: Commit**

```bash
git add internal/storage/storage_test.go internal/publish/service_test.go
git commit -m "test: increase storage and publish test coverage"
```

---

### Task 4.4: E2E Test Setup

**Files:**
- Create: `web/playwright.config.ts`
- Create: `web/e2e/smoke.spec.ts`
- Modify: `web/package.json`

- [ ] **Step 1: Install Playwright**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install -D @playwright/test`
Run: `npx playwright install chromium`

- [ ] **Step 2: Create Playwright config**

```typescript
// web/playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: 'http://localhost:8080/studio/',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
})
```

- [ ] **Step 3: Write smoke tests**

```typescript
// web/e2e/smoke.spec.ts
import { test, expect } from '@playwright/test'

test('login and navigate to posts', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="password"]', 'testpassword')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL(/\/posts/)
})

test('create new post', async ({ page }) => {
  // Login first...
  // Click "New Post" or navigate to a new slug
  // Verify editor loads
})

test('health page loads', async ({ page }) => {
  // Login, navigate to /health
  // Verify health checks display
})
```

- [ ] **Step 4: Add to package.json scripts**

```json
"test:e2e": "playwright test"
```

- [ ] **Step 5: Commit**

```bash
git add web/playwright.config.ts web/e2e/smoke.spec.ts web/package.json web/package-lock.json
git commit -m "test: add Playwright E2E test setup with smoke tests"
```

---

## Phase 5: Performance (2 tasks)

### Task 5.1: Bundle Optimization

**Files:**
- Modify: `web/src/router/index.ts`
- Modify: `web/src/views/EditorView.vue`
- Modify: `web/package.json`

- [ ] **Step 1: Verify route-based code splitting**

The router already uses `() => import(...)` for all routes. Verify with `npx vite-bundle-analyzer` that heavy components are split.

- [ ] **Step 2: Tree-shake lucide icons**

All lucide imports should be specific (not barrel imports). Check all files for `import { ... } from 'lucide-vue-next'` — these are already tree-shakeable. Verify no `import * from 'lucide-vue-next'` exists.

- [ ] **Step 3: Lazy-load CodeMirror language modes**

The `@codemirror/language-data` package includes all language modes. Verify they're loaded lazily (they should be by default with the `codeLanguages` option).

- [ ] **Step 4: Analyze and verify bundle size**

Run: `cd web && npm run build`
Check the build output for chunk sizes. Target: < 200KB initial JS gzipped.

- [ ] **Step 5: Commit**

```bash
git add web/src/router/index.ts web/package.json
git commit -m "perf: bundle optimization (verify code splitting, tree shaking, lazy loading)"
```

---

### Task 5.2: API Response Caching

**Files:**
- Modify: `web/src/services/api.ts`
- Modify: `internal/httpapi/middleware.go`

- [ ] **Step 1: Add ETag middleware**

```go
// internal/httpapi/middleware.go — add ETag support
func withETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response to capture body hash
		rec := &etagRecorder{ResponseWriter: w, buf: &bytes.Buffer{}}
		next.ServeHTTP(rec, r)

		// Generate ETag from body
		hash := md5.Sum(rec.buf.Bytes())
		etag := fmt.Sprintf(`"%x"`, hash)

		// Check If-None-Match
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")
		w.Write(rec.buf.Bytes())
	})
}
```

- [ ] **Step 2: Add SWR cache for post list in api.ts**

```typescript
// web/src/services/api.ts
let postListCache: { data: PostState[], timestamp: number } | null = null
const POST_LIST_TTL = 30000 // 30s

async function posts(): Promise<PostState[]> {
  // Return cached data immediately if fresh
  if (postListCache && Date.now() - postListCache.timestamp < POST_LIST_TTL) {
    // Revalidate in background
    request<PostState[]>('/posts').then(data => {
      postListCache = { data, timestamp: Date.now() }
    }).catch(() => {})
    return postListCache.data
  }

  const data = await request<PostState[]>('/posts')
  postListCache = { data, timestamp: Date.now() }
  return data
}

function invalidatePostCache() {
  postListCache = null
}
```

- [ ] **Step 3: Invalidate cache on mutations**

Call `invalidatePostCache()` after saveDraft, publish, delete, bulk operations.

- [ ] **Step 4: Commit**

```bash
git add web/src/services/api.ts internal/httpapi/middleware.go
git commit -m "feat: API response caching (ETag, SWR for post list, cache invalidation)"
```

---

## Phase 2 (continued): Additional UI Tasks

### Task 2.9: Color System

**Files:**
- Modify: `web/src/styles/tokens.css`
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add semantic status colors**

```css
/* Add to tokens.css :root */
--color-status-draft: oklch(0.75 0.15 75);
--color-status-draft-foreground: oklch(0.20 0 0);
--color-status-published: oklch(0.60 0.18 145);
--color-status-published-foreground: oklch(0.98 0 0);
--color-status-error: oklch(0.55 0.2 25);
--color-status-error-foreground: oklch(0.98 0 0);
--color-status-info: oklch(0.60 0.12 250);
--color-status-info-foreground: oklch(0.98 0 0);

/* Chart palette for metrics */
--chart-1: oklch(0.60 0.15 250);
--chart-2: oklch(0.65 0.18 145);
--chart-3: oklch(0.70 0.15 75);
--chart-4: oklch(0.55 0.2 25);
--chart-5: oklch(0.60 0.12 300);
--chart-6: oklch(0.50 0.10 200);
```

- [ ] **Step 2: Add gradient accent for brand elements**

```css
.gradient-brand {
  background: linear-gradient(135deg, var(--color-primary), oklch(from var(--color-primary) calc(l + 0.1) c h));
}
```

- [ ] **Step 3: Apply to LoginView sidebar header and other brand elements**

- [ ] **Step 4: Commit**

```bash
git add web/src/styles/tokens.css web/src/styles/tailwind.css web/src/views/LoginView.vue
git commit -m "feat: color system (semantic status, chart palette, gradient accents)"
```

---

### Task 2.10: Card & Surface Design

**Files:**
- Modify: `web/src/styles/tailwind.css`
- Modify: `web/src/views/PostsView.vue`

- [ ] **Step 1: Add elevation utility classes**

```css
/* Surface levels */
.surface-0 { background: var(--color-background); }
.surface-1 {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
}
.surface-2 {
  background: var(--color-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
}
.surface-3 {
  background: var(--color-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
}

/* Consistent card padding */
.card-padding { padding: var(--space-4); }
.modal-padding { padding: var(--space-6); }
```

- [ ] **Step 2: Apply to post cards, dropdowns, and modals**

Replace ad-hoc border/shadow with surface classes in PostsView, dropdown menus, and dialogs.

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css web/src/views/PostsView.vue
git commit -m "feat: consistent card and surface elevation system"
```

---

### Task 2.11: Feedback & Notification Polish

**Files:**
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add progress bar styles**

```css
.progress-bar {
  height: 4px;
  border-radius: var(--radius-full);
  background: var(--color-muted);
  overflow: hidden;
}
.progress-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  transition: width var(--duration-normal);
}
.progress-bar-indeterminate .progress-bar-fill {
  width: 30%;
  animation: progress-indeterminate 1.5s ease-in-out infinite;
}
@keyframes progress-indeterminate {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(400%); }
}

/* Badge variants */
.badge-success { background: oklch(from var(--color-success) l c h / 0.15); color: var(--color-success); }
.badge-warning { background: oklch(from var(--color-warning) l c h / 0.15); color: var(--color-warning); }
.badge-error { background: oklch(from var(--color-destructive) l c h / 0.15); color: var(--color-destructive); }
.badge-info { background: oklch(from var(--color-info) l c h / 0.15); color: var(--color-info); }
```

- [ ] **Step 2: Add tooltip animation**

```css
.tooltip-content {
  animation: fade-up var(--duration-fast) var(--ease-smooth);
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css
git commit -m "feat: feedback system polish (progress bars, badges, tooltip animations)"
```

---

### Task 2.12: Typography Polish for Content

**Files:**
- Modify: `web/src/styles/tailwind.css`
- Modify: `web/src/components/MarkdownPreview.vue`

- [ ] **Step 1: Add markdown content typography**

```css
.markdown-body {
  font-family: var(--font-serif);
  font-size: var(--text-base);
  line-height: var(--leading-relaxed);
  color: var(--color-foreground);
  word-break: break-word;
}
.markdown-body h1 { font-size: var(--text-3xl); font-weight: 700; margin: 1.5em 0 0.5em; }
.markdown-body h2 { font-size: var(--text-2xl); font-weight: 600; margin: 1.2em 0 0.5em; }
.markdown-body h3 { font-size: var(--text-xl); font-weight: 600; margin: 1em 0 0.4em; }
.markdown-body p { margin: 0.8em 0; }
.markdown-body code {
  font-family: var(--font-mono);
  font-size: 0.9em;
  background: var(--color-muted);
  padding: 0.15em 0.4em;
  border-radius: var(--radius-sm);
}
.markdown-body pre {
  background: var(--color-muted);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  overflow-x: auto;
}
.markdown-body pre code {
  background: none;
  padding: 0;
}
.markdown-body blockquote {
  border-left: 3px solid var(--color-border);
  padding-left: var(--space-4);
  color: var(--color-muted-foreground);
  font-style: italic;
}
.markdown-body img {
  max-width: 100%;
  border-radius: var(--radius-md);
}
.markdown-body table {
  border-collapse: collapse;
  width: 100%;
}
.markdown-body th, .markdown-body td {
  border: 1px solid var(--color-border);
  padding: var(--space-2) var(--space-3);
  text-align: left;
}
.markdown-body th {
  background: var(--color-muted);
  font-weight: 600;
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/styles/tailwind.css web/src/components/MarkdownPreview.vue
git commit -m "feat: polished markdown content typography (headings, code, tables, blockquotes)"
```

---

### Task 2.13: Touch Targets & Safe Areas

**Files:**
- Modify: `web/src/styles/tailwind.css`

- [ ] **Step 1: Add touch target utilities**

```css
/* Ensure minimum 44px touch targets on mobile */
@media (pointer: coarse) {
  button, a, [role="button"], input[type="checkbox"], input[type="radio"] {
    min-height: 44px;
    min-width: 44px;
  }
}

/* Safe area insets for notched devices */
.safe-area-top { padding-top: env(safe-area-inset-top); }
.safe-area-bottom { padding-bottom: env(safe-area-inset-bottom); }
.safe-area-left { padding-left: env(safe-area-inset-left); }
.safe-area-right { padding-right: env(safe-area-inset-right); }
```

- [ ] **Step 2: Apply safe areas to main layout**

In `App.vue`, add `safe-area-top` to header, `safe-area-bottom` to main content area.

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css web/src/App.vue
git commit -m "feat: touch target minimums and safe area insets for mobile"
```

---

### Task 5.3: Image Lazy Loading in Preview

**Files:**
- Modify: `web/src/components/MarkdownPreview.vue`

- [ ] **Step 1: Add lazy loading to rendered images**

In the marked renderer for images, add `loading="lazy"` and `decoding="async"`:

```typescript
image(href: string, title: string, text: string) {
  return `<img src="${href}" alt="${text}" loading="lazy" decoding="async" class="max-w-full rounded" ${title ? `title="${title}"` : ''} />`
}
```

- [ ] **Step 2: Add progressive loading placeholder**

```css
.markdown-body img {
  background: var(--color-muted);
  transition: opacity var(--duration-normal);
}
.markdown-body img[loading="lazy"] {
  opacity: 0;
}
.markdown-body img[loading="lazy"].loaded {
  opacity: 1;
}
```

Add an `onload` handler in the image renderer: `onload="this.classList.add('loaded')"`

- [ ] **Step 3: Commit**

```bash
git add web/src/components/MarkdownPreview.vue
git commit -m "feat: image lazy loading with progressive fade-in in preview"
```

---

## Phase 4 (continued): Additional Tasks

### Task 4.4: Navigation Improvements

**Files:**
- Modify: `web/src/App.vue`
- Modify: `web/src/i18n/index.ts`

- [ ] **Step 1: Add sidebar collapse state**

```typescript
// In App.vue
const sidebarCollapsed = ref(localStorage.getItem('sidebar:collapsed') === 'true')
watch(sidebarCollapsed, (val) => localStorage.setItem('sidebar:collapsed', String(val)))
```

- [ ] **Step 2: Add keyboard shortcut for sidebar toggle**

```typescript
useEventListener('keydown', (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === '\\') {
    e.preventDefault()
    sidebarCollapsed.value = !sidebarCollapsed.value
  }
})
```

- [ ] **Step 3: Apply collapsed state to sidebar**

When collapsed, hide labels and show only icons. Sidebar width: 240px → 56px.

- [ ] **Step 4: Commit**

```bash
git add web/src/App.vue web/src/i18n/index.ts
git commit -m "feat: collapsible sidebar with keyboard shortcut (⌘\\)"
```

---

### Task 4.5: Frontend Component Tests

**Files:**
- Create: `web/src/__tests__/EditorView-v2.test.ts`
- Create: `web/src/__tests__/SplitView.test.ts`
- Create: `web/src/__tests__/Toast.test.ts`

- [ ] **Step 1: Write SplitView tests**

```typescript
// web/src/__tests__/SplitView.test.ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SplitView from '@/components/SplitView.vue'

describe('SplitView', () => {
  it('renders source panel in source mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'source', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Source')
  })

  it('renders both panels in split mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Source')
    expect(wrapper.text()).toContain('Preview')
  })

  it('renders preview panel in preview mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'preview', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Preview')
  })
})
```

- [ ] **Step 2: Write Toast tests**

```typescript
// web/src/__tests__/Toast.test.ts
import { describe, it, expect, vi } from 'vitest'
import { useToast } from '@/composables/useToast'

describe('useToast', () => {
  it('adds success toast', () => {
    const { toasts, success } = useToast()
    success('Saved!', 'Your changes were saved')
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].type).toBe('success')
    expect(toasts.value[0].title).toBe('Saved!')
  })

  it('adds error toast with infinite duration', () => {
    const { toasts, error } = useToast()
    error('Failed', 'Something went wrong')
    expect(toasts.value[0].duration).toBe(Infinity)
  })

  it('removes toast by id', () => {
    const { toasts, success, remove } = useToast()
    success('Test')
    const id = toasts.value[0].id
    remove(id)
    expect(toasts.value).toHaveLength(0)
  })

  it('limits to 3 toasts', () => {
    const { toasts, info } = useToast()
    info('1'); info('2'); info('3'); info('4')
    expect(toasts.value).toHaveLength(3)
  })
})
```

- [ ] **Step 3: Write EditorView v2 tests**

```typescript
// web/src/__tests__/EditorView-v2.test.ts
// Test split view mode switching, toolbar actions, focus mode toggle
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npx vitest run`
Expected: All pass

- [ ] **Step 5: Commit**

```bash
git add web/src/__tests__/SplitView.test.ts web/src/__tests__/Toast.test.ts web/src/__tests__/EditorView-v2.test.ts
git commit -m "test: add SplitView, Toast, and EditorView v2 component tests"
```

---

### Task 4.6: E2E Test Setup (Playwright)

**Files:**
- Create: `web/playwright.config.ts`
- Create: `web/e2e/smoke.spec.ts`
- Modify: `web/package.json`

- [ ] **Step 1: Install Playwright**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm install -D @playwright/test`
Run: `npx playwright install chromium`

- [ ] **Step 2: Create Playwright config**

```typescript
// web/playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 1,
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8080/studio/',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  webServer: {
    command: 'cd .. && BLOG_STUDIO_ADMIN_PASSWORD_HASH=$2a$10$test BLOG_STUDIO_SESSION_SECRET=testsecret32charslong!! BLOG_STUDIO_BLOG_ROOT=./testdata/hugo BLOG_STUDIO_DATA_ROOT=./tmp/data BLOG_STUDIO_STATIC_DIR=./web/dist go run ./cmd/server',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 30000,
  },
})
```

- [ ] **Step 3: Write smoke tests**

```typescript
// web/e2e/smoke.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Smoke Tests', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[type="password"]')).toBeVisible()
  })

  test('login with valid password redirects to posts', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="password"]', 'testpassword')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL(/\/posts/)
  })

  test('login with invalid password shows error', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')
    await expect(page.locator('[role="alert"]')).toBeVisible()
  })

  test('posts page shows post list', async ({ page }) => {
    // Login first...
    await page.goto('/login')
    await page.fill('input[type="password"]', 'testpassword')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL(/\/posts/)
    // Should show either posts or empty state
    await expect(page.locator('text=还没有文章, text=文章')).toBeVisible()
  })
})
```

- [ ] **Step 4: Add script to package.json**

```json
"test:e2e": "playwright test"
```

- [ ] **Step 5: Commit**

```bash
git add web/playwright.config.ts web/e2e/smoke.spec.ts web/package.json web/package-lock.json
git commit -m "test: add Playwright E2E test setup with smoke tests"
```

---

## Verification Checklist

After completing all tasks:

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
npx vitest run
npx playwright test  # if Playwright is installed

# Visual verification
# - Split view: source, preview, split modes work
# - Outline panel: headings extracted, click to jump
# - Math: $...$ renders inline, $$...$$ renders block
# - Find/Replace: ⌘F opens, search works, replace works
# - Status bar: word count, cursor position, reading time
# - Focus mode: hides chrome, centered content, Esc exits
# - Mobile: sidebar drawer, stacked layout, 44px targets
# - Dark mode: all components look correct
# - Empty states: display when no data
# - Skeletons: shimmer during loading
# - Toasts: 4 types display correctly
# - Version history: save, list, restore
# - Breadcrumbs: settings and editor pages
```
