# UI Modern Minimalist Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle blog-studio-web with a modern minimalist aesthetic — apply design tokens to actual components for visible, polished results.

**Architecture:** Progressive restyle — each task modifies 1-3 files, produces a visually improved app. All changes use existing Tailwind + design tokens, no new dependencies.

**Tech Stack:** Vue 3.5, Tailwind v4, shadcn-vue, existing design tokens (tokens.css, tailwind.css)

**Repository:** `/Users/xiang/Desktop/personal-server/blog-studio-web/`

---

## Task 1: Color & Token Refinement

**Files:**
- Modify: `web/src/styles/tailwind.css`

**Goal:** Make accent color more saturated, improve muted contrast, refine shadows.

- [ ] **Step 1: Update accent color to be more vibrant**

Read `web/src/styles/tailwind.css`. Find the `:root` block. Change:
```css
--accent: #4f6f82;
```
to:
```css
--accent: #3b82c4;
```

This shifts from muted blue-grey to a cleaner, more saturated blue.

- [ ] **Step 2: Update dark mode accent**

In the `.dark` block, change:
```css
--accent: #9dc2d6;
```
to:
```css
--accent: #6ba3d6;
```

- [ ] **Step 3: Refine muted colors for better contrast**

Light mode — change `--muted-foreground` from whatever it currently is to:
```css
--muted-foreground: #6b7280;
```

Dark mode — change `--muted-foreground` to:
```css
--muted-foreground: #9ca3af;
```

- [ ] **Step 4: Remove paper texture from body**

Find the `body` rule with `background-image` (SVG dot pattern). Comment it out or remove it. The solid background color is cleaner for minimalist style.

- [ ] **Step 5: Verify build**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`
Expected: SUCCESS

- [ ] **Step 6: Commit**

```bash
git add web/src/styles/tailwind.css
git commit -m "style: refine accent color, improve contrast, remove paper texture"
```

---

## Task 2: Sidebar Restyle

**Files:**
- Modify: `web/src/App.vue`

**Goal:** More compact sidebar (220px), cleaner nav items with accent hover bar, remove subtitle.

- [ ] **Step 1: Read App.vue**

Read the file. Find the sidebar `<aside>` element.

- [ ] **Step 2: Change sidebar width**

Find `w-[240px]` on the desktop sidebar. Change to `w-[220px]`.

- [ ] **Step 3: Simplify brand area**

Find the brand section with "博" + site name + "管理后台". Remove the "管理后台" subtitle line. Keep "博" + site name only.

Change the brand container from `gap-3` to `gap-2`.

- [ ] **Step 4: Restyle nav items**

Find the nav link template. Current classes:
```
flex items-center gap-2.5 rounded px-2.5 py-1.5 text-sm
```

Change to:
```
flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px]
```

Remove `hover:translate-x-0.5` (too playful for minimalist).

Change active indicator from `w-0.5` to `w-[2px]` for slightly more visible accent bar.

- [ ] **Step 5: Make group labels smaller**

Find `text-[10px]` on group labels. Keep as-is but reduce opacity from `text-muted-foreground/60` to `text-muted-foreground/50`.

- [ ] **Step 6: Compact bottom actions**

Change sync/logout buttons from `h-8` to `h-7`.

- [ ] **Step 7: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/App.vue
git commit -m "style: restyle sidebar — compact, cleaner nav, accent hover bar"
```

---

## Task 3: Post List Restyle

**Files:**
- Modify: `web/src/views/PostsView.vue`

**Goal:** Remove card borders, tighter spacing, cleaner meta, borderless search.

- [ ] **Step 1: Read PostsView.vue**

Read the file. Find the post card template.

- [ ] **Step 2: Remove card borders**

Find `border border-border/60` on post cards. Change to just `border border-transparent` (invisible border for layout stability, no visual border).

Add `hover:bg-muted/40` to the card classes for subtle hover feedback.

- [ ] **Step 3: Tighten card padding**

Find the card's internal padding. Change from current value to `py-3 px-4`.

- [ ] **Step 4: Clean up title**

Find `font-serif font-medium text-sm`. Change to `text-sm font-medium` (remove serif, use system sans for cleaner look).

- [ ] **Step 5: Make tags smaller**

Find tag elements with `font-deco text-[11px]`. Change to `text-[10px] px-1.5 py-0.5 rounded-full bg-muted`.

- [ ] **Step 6: Clean up search bar**

The search bar already uses underline style. Add `transition-colors` for smooth focus effect:
```
border-0 border-b border-border rounded-none bg-transparent pl-6 pb-2 h-auto focus-visible:ring-0 focus-visible:border-accent transition-colors
```

- [ ] **Step 7: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/views/PostsView.vue
git commit -m "style: restyle post list — borderless cards, tighter spacing, clean meta"
```

---

## Task 4: Editor Chrome Restyle

**Files:**
- Modify: `web/src/views/EditorView.vue`

**Goal:** Transparent toolbar, no border on editor area, subtle status bar.

- [ ] **Step 1: Read EditorView.vue**

Read the file. Find the toolbar and editor body sections.

- [ ] **Step 2: Make toolbar transparent**

Find the toolbar container. If it has a background class, remove it. If it has a border, change to `border-b border-border/40`.

- [ ] **Step 3: Remove editor border**

Find `rounded border border-border/60` on the editor body. Change to `rounded-lg` (keep rounded, remove border).

- [ ] **Step 4: Subtle status bar**

Find the status bar container. Change from `py-2` to `py-1.5`. Add `border-t border-border/30` for very subtle top separator.

- [ ] **Step 5: Clean up title input**

Find the title input with `font-serif`. Change to just use system font (remove `font-serif`).

- [ ] **Step 6: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/views/EditorView.vue
git commit -m "style: restyle editor — transparent toolbar, borderless editor, subtle status bar"
```

---

## Task 5: Login Page Polish

**Files:**
- Modify: `web/src/views/LoginView.vue`

**Goal:** Better card shadow, input styling, keep right-card layout.

- [ ] **Step 1: Read LoginView.vue**

Read the file. Find the right panel form area.

- [ ] **Step 2: Improve card feel**

Find the form container (`max-w-[400px]`). Add `p-8` padding and `rounded-2xl bg-card shadow-lg` to create a card effect. Or if it's already in a card, increase `border-radius` to `rounded-2xl` (16px).

- [ ] **Step 3: Improve input styling**

The password input uses underline style (`border-0 border-b`). Keep this but add `transition-colors` for smooth focus:
```
h-12 border-0 border-b border-border rounded-none bg-transparent px-0 focus-visible:ring-0 focus-visible:border-accent transition-colors
```

- [ ] **Step 4: Improve submit button**

Find the submit button. It's already `rounded-full bg-accent`. Change height from `h-12` to `h-11` (slightly more compact). Ensure it uses `font-medium tracking-wide`.

- [ ] **Step 5: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/views/LoginView.vue
git commit -m "style: polish login page — better card, input transitions, compact button"
```

---

## Task 6: Settings Pages Restyle

**Files:**
- Modify: `web/src/views/settings/SettingsLayout.vue`
- Modify: `web/src/views/settings/GeneralTab.vue`
- Modify: `web/src/views/settings/SecurityTab.vue`
- Modify: `web/src/views/settings/WritingTab.vue`
- Modify: `web/src/views/settings/AuditTab.vue`

**Goal:** Consistent minimalist styling across settings.

- [ ] **Step 1: Read SettingsLayout.vue**

Find the tab navigation. Clean up spacing and styling:
- Tab buttons: `text-sm` instead of any larger size
- Active tab: accent color text + bottom accent border
- Tab content area: `pt-4` spacing

- [ ] **Step 2: Read and clean GeneralTab.vue**

- Section headers: `text-sm font-medium` (smaller, cleaner)
- Form labels: `text-xs uppercase tracking-wider text-muted-foreground`
- Input fields: consistent `h-9` height
- Spacing between sections: `space-y-6`

- [ ] **Step 3: Apply same patterns to SecurityTab, WritingTab, AuditTab**

For each file:
- Consistent section headers
- Consistent label styles
- Consistent input heights
- Consistent spacing

- [ ] **Step 4: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/views/settings/
git commit -m "style: restyle settings pages — consistent minimalist form styling"
```

---

## Task 7: Other Views Restyle

**Files:**
- Modify: `web/src/views/HealthView.vue`
- Modify: `web/src/views/TrashView.vue`
- Modify: `web/src/views/TagsView.vue`
- Modify: `web/src/views/HomeView.vue`
- Modify: `web/src/views/NowView.vue`

**Goal:** Apply consistent minimalist styling to remaining views.

- [ ] **Step 1: Read HealthView.vue**

- Health check cards: remove borders, use `bg-muted/30 rounded-lg p-4` for each check
- Status badges: use semantic badge classes (`badge-success`, `badge-error`)
- Spacing: `space-y-3` between checks

- [ ] **Step 2: Read TrashView.vue**

- Table/list: consistent with PostsView style (borderless rows, hover highlight)
- Action buttons: ghost variant, compact

- [ ] **Step 3: Read TagsView.vue**

- Tag list: clean rows with hover highlight
- Tag pills: `text-xs px-2 py-0.5 rounded-full bg-muted`

- [ ] **Step 4: Read HomeView.vue and NowView.vue**

- Editor areas: consistent with EditorView style
- Minimal chrome, clean typography

- [ ] **Step 5: Verify and commit**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

```bash
git add web/src/views/HealthView.vue web/src/views/TrashView.vue web/src/views/TagsView.vue web/src/views/HomeView.vue web/src/views/NowView.vue
git commit -m "style: restyle remaining views — consistent minimalist design"
```

---

## Task 8: Global Polish & Shared Utilities

**Files:**
- Modify: `web/src/styles/tailwind.css`

**Goal:** Extract repeated patterns into shared utility classes.

- [ ] **Step 1: Add shared utility classes**

Add to `@layer utilities` in tailwind.css:

```css
/* Underline input — used in search, title, login */
.input-underline {
  @apply border-0 border-b border-border rounded-none bg-transparent px-0 focus-visible:ring-0 focus-visible:border-accent transition-colors;
}

/* Section label — used in editor meta, settings, tag groups */
.section-label {
  @apply text-[10px] uppercase tracking-wider text-muted-foreground/70 font-medium;
}

/* Compact ghost icon button */
.btn-icon-ghost {
  @apply h-7 w-7 p-0;
}

/* Tag pill */
.tag-pill {
  @apply text-[10px] px-1.5 py-0.5 rounded-full bg-muted;
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/xiang/Desktop/personal-server/blog-studio-web/web && npm run build`

- [ ] **Step 3: Commit**

```bash
git add web/src/styles/tailwind.css
git commit -m "style: add shared utility classes for consistent patterns"
```

---

## Verification

After all tasks:

```bash
cd /Users/xiang/Desktop/personal-server/blog-studio-web/web
npm run build    # No errors
npm run lint     # No new errors
npm run test     # All tests pass
```

Visual checks:
- Sidebar: 220px, clean nav, accent hover bar
- Post list: borderless cards, hover highlight, compact meta
- Editor: transparent toolbar, no border, subtle status bar
- Login: polished card, smooth transitions
- Settings: consistent form styling
- All views: cohesive minimalist aesthetic
