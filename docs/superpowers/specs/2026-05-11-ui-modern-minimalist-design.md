# UI Modern Minimalist Redesign — Design Spec

## Goal

Restyle blog-studio-web with a modern minimalist aesthetic (Linear/Raycast inspired). Apply the existing design system tokens to actual UI components for visible, polished results.

## Design Direction

- **Style**: Modern minimalist — clean lines, subtle depth, generous whitespace
- **Accent**: Saturated primary color for interactive elements
- **Surfaces**: Flat with subtle hover states, minimal borders
- **Typography**: Tight, precise, well-spaced

## Scope

### 1. Sidebar (App.vue)

- Width: 220px (from 240px)
- Background: `bg-sidebar` with subtle border-right
- Brand area: Compact — icon + site name only, no "管理后台" subtitle
- Nav items: 13px text, `py-2` padding, hover shows left accent bar (2px, accent color)
- Active item: accent text color + left bar
- Group labels: smaller (10px), more muted
- Bottom actions: compact, icon + text
- Collapsed state: 56px, icon-only with tooltips

### 2. Post List (PostsView.vue)

- Cards: Remove border, use `hover:bg-muted/50` for interaction
- Each row: `py-3 px-4`, tighter spacing
- Title: 14px, medium weight
- Meta (date, tags): 12px, muted
- Tags: small pill (`text-[10px] px-1.5 py-0.5 rounded-full bg-muted`)
- Search bar: borderless, bottom-border style (`border-b border-transparent focus:border-ring`)
- Status filter: pill-style buttons
- Bulk action bar: floating, blur background

### 3. Editor (EditorView.vue)

- Toolbar: 36px height, transparent background, compact icon buttons (28px)
- Toolbar groups: subtle dividers
- Editor area: no border, blends with background
- Status bar: 24px height, very subtle
- Breadcrumbs: 12px, very muted
- Split view divider: 1px, accent color on hover/drag

### 4. Login Page (LoginView.vue)

- Keep current right-card layout
- Card: larger border-radius (16px), subtle shadow (`shadow-lg`)
- Input fields: 40px height, rounded (8px), focus ring with accent color
- Button: full-width, 40px height, accent background, rounded
- Brand area: larger emoji/text, better spacing
- Color: use gradient-brand for accent elements

### 5. Global

- Border-radius: 8px (cards), 6px (buttons), 4px (small elements)
- Shadows: only on hover/floating elements
- Muted colors: more subtle (increase opacity difference)
- Page transitions: keep existing fade-up

## Implementation Order

1. Global tokens refinement (colors, spacing)
2. Sidebar restyle
3. Post list restyle
4. Editor chrome restyle
5. Login page polish
6. Settings pages
7. Other views (Health, Trash, Tags)
