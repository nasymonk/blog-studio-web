<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { PlusIcon, SearchIcon, RefreshCwIcon, EditIcon, EyeIcon, AlertTriangleIcon, Trash2Icon, CheckSquareIcon, SquareIcon, SendIcon, DownloadIcon, UploadIcon } from 'lucide-vue-next'
import { useDebounceFn } from '@vueuse/core'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { api } from '@/services/api'
import type { PostState } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { filterPosts } from '@/utils/filterPosts'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import SkeletonCard from '@/components/SkeletonCard.vue'
import { Separator } from '@/components/ui/separator'
import EmptyState from '@/components/EmptyState.vue'

const router = useRouter()
const route = useRoute()
const store = useStore()
const { t } = useI18n()
const notify = useNotify()

const searchQuery = ref((route.query.q as string) || '')
const statusFilter = ref((route.query.status as string) || 'all')
const selectedTags = ref<string[]>((route.query.tags ? (route.query.tags as string).split(',') : []))
const sortBy = ref((route.query.sort as string) || 'date-desc')

const searchInputRef = ref<HTMLInputElement | null>(null)
const parentRef = ref<HTMLElement | null>(null)
const focusedIndex = ref(-1)
const selectedSlugs = ref<Set<string>>(new Set())

function onListKeydown(e: KeyboardEvent) {
  const active = document.activeElement as HTMLElement
  const isInSearch = active === searchInputRef.value
  if (isInSearch && e.key !== 'Escape') return

  if (e.key === '/' && !isInSearch) {
    e.preventDefault()
    searchInputRef.value?.focus()
    return
  }
  if (e.key === 'ArrowDown' || e.key === 'j') {
    e.preventDefault()
    focusedIndex.value = Math.min(focusedIndex.value + 1, filtered.value.length - 1)
    rowVirtualizer.value.scrollToIndex(focusedIndex.value, { align: 'auto' })
    focusCard(focusedIndex.value)
  } else if (e.key === 'ArrowUp' || e.key === 'k') {
    e.preventDefault()
    focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
    rowVirtualizer.value.scrollToIndex(focusedIndex.value, { align: 'auto' })
    focusCard(focusedIndex.value)
  } else if (e.key === 'Enter' && focusedIndex.value >= 0) {
    const slug = filtered.value[focusedIndex.value]?.slug
    if (slug) router.push(`/posts/${encodeURIComponent(slug)}`)
  } else if (e.key === 'Escape' && isInSearch) {
    searchInputRef.value?.blur()
  }
}

function focusCard(i: number) {
  const cards = document.querySelectorAll<HTMLElement>('.post-card')
  cards[i]?.focus()
}

onMounted(() => { document.addEventListener('keydown', onListKeydown) })
onBeforeUnmount(() => { document.removeEventListener('keydown', onListKeydown) })

const allTags = computed(() => {
  const set = new Set<string>()
  for (const p of store.posts) for (const tag of (p.tags || [])) set.add(tag)
  return [...set].sort()
})

const filtered = computed(() =>
  filterPosts(store.posts, searchQuery.value, statusFilter.value, selectedTags.value, sortBy.value)
)

const rowVirtualizer = useVirtualizer(computed(() => ({
  count: filtered.value.length,
  getScrollElement: () => parentRef.value,
  estimateSize: () => 92,
  overscan: 5,
})))

const virtualRows = computed(() => rowVirtualizer.value.getVirtualItems())

const syncUrl = useDebounceFn(() => {
  const q: Record<string, string> = {}
  if (searchQuery.value) q.q = searchQuery.value
  if (statusFilter.value !== 'all') q.status = statusFilter.value
  if (selectedTags.value.length) q.tags = selectedTags.value.join(',')
  if (sortBy.value !== 'date-desc') q.sort = sortBy.value
  router.replace({ query: q })
}, 300)

watch([searchQuery, statusFilter, selectedTags, sortBy], () => { syncUrl() }, { deep: true })

async function loadPosts() {
  store.postsLoading = true
  try { store.posts = await api.posts() }
  catch (e: any) { notify.error(e) }
  finally { store.postsLoading = false }
}

async function newPost() {
  const slug = prompt(t.value.newPostSlug)
  if (!slug) return
  router.push(`/posts/${encodeURIComponent(slug)}`)
}

async function deletePost(post: PostState) {
  if (!confirm(t.value.deleteConfirm(post.title || post.slug))) return
  try {
    await api.deletePost(post.slug)
    notify.success(t.value.deleteOk(post.title || post.slug))
    store.posts = store.posts.filter(p => p.slug !== post.slug)
  } catch (e: any) { notify.error(e) }
}

type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline'

function statusBadge(p: PostState): { variant: BadgeVariant; label: string } {
  if (p.syncStatus === 'conflict') return { variant: 'destructive', label: t.value.conflicts }
  if (p.syncStatus === 'stale') return { variant: 'secondary', label: t.value.stale }
  if (p.draft) return { variant: 'outline', label: t.value.draft }
  if (p.syncStatus === 'clean') return { variant: 'default', label: t.value.published }
  return { variant: 'outline', label: p.syncStatus }
}

function statusColor(p: PostState): string {
  if (p.syncStatus === 'conflict') return 'bg-destructive'
  if (p.syncStatus === 'stale') return 'bg-warn'
  if (p.draft) return 'bg-muted-foreground/40'
  return 'bg-accent'
}

function relDate(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return t.value.today
  if (days === 1) return t.value.yesterday
  if (days < 30) return t.value.daysAgo(days)
  if (days < 365) return t.value.monthsAgo(Math.floor(days / 30))
  return t.value.yearsAgo(Math.floor(days / 365))
}

function toggleTag(tag: string) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) selectedTags.value.splice(idx, 1)
  else selectedTags.value.push(tag)
}

function toggleSelect(slug: string) {
  const next = new Set(selectedSlugs.value)
  if (next.has(slug)) next.delete(slug)
  else next.add(slug)
  selectedSlugs.value = next
}

function selectAll() {
  selectedSlugs.value = new Set(filtered.value.map(p => p.slug))
}

function deselectAll() {
  selectedSlugs.value = new Set()
}

async function bulkTrash() {
  const slugs = [...selectedSlugs.value]
  if (!slugs.length || !confirm(t.value.trashSelected + ` (${slugs.length})?`)) return
  try {
    const results = await api.bulkTrash(slugs)
    const ok = results.filter(r => r.success).length
    notify.success(`${ok}/${slugs.length} deleted`)
    selectedSlugs.value = new Set()
    store.posts = store.posts.filter(p => !selectedSlugs.value.has(p.slug) || !slugs.includes(p.slug))
    await loadPosts()
  } catch (e: any) { notify.error(e) }
}

async function bulkPublish() {
  const slugs = [...selectedSlugs.value]
  if (!slugs.length) return
  try {
    const results = await api.bulkPublish(slugs)
    const ok = results.filter(r => r.status === 'success').length
    notify.success(`${ok}/${slugs.length} published`)
    selectedSlugs.value = new Set()
    await loadPosts()
  } catch (e: any) { notify.error(e) }
}

async function exportAll() {
  try {
    await api.exportAll()
    notify.success(t.value.exportOk)
  } catch (e: any) { notify.error(e) }
}

function importZip() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.zip'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    try {
      const result = await api.importZip(file)
      notify.success(t.value.importOk(result.imported))
      await loadPosts()
    } catch (e: any) { notify.error(e) }
  }
  input.click()
}

onMounted(loadPosts)
</script>

<template>
  <div class="flex flex-col gap-5 max-w-4xl">
    <!-- Search bar: underline style -->
    <div class="flex flex-col md:flex-row md:items-end gap-3">
      <div class="flex items-end gap-3 flex-1 min-w-0">
        <div class="relative flex-1 max-w-[360px]">
          <label for="posts-search" class="sr-only">{{ t.search }}</label>
          <SearchIcon class="absolute left-0 bottom-3 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
          <Input
            id="posts-search"
            ref="searchInputRef"
            v-model="searchQuery"
            class="border-0 border-b border-border rounded-none bg-transparent pl-6 pb-2 h-auto focus-visible:ring-2 focus-visible:ring-accent transition-colors"
            :placeholder="t.search"
          />
        </div>
        <Select v-model="sortBy">
          <SelectTrigger class="w-auto min-w-[120px] h-8 text-xs border-0 border-b border-border rounded-none bg-transparent"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="date-desc">{{ t.sortDate }} ↓</SelectItem>
            <SelectItem value="date-asc">{{ t.sortDate }} ↑</SelectItem>
            <SelectItem value="title">{{ t.sortTitle }}</SelectItem>
            <SelectItem value="status">{{ t.sortStatus }}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex items-center gap-1 flex-wrap">
        <Button
          v-for="s in ['all','drafts','published','conflicts','stale']" :key="s"
          variant="ghost" size="sm" class="text-xs h-8 md:h-7"
          :class="statusFilter === s ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
          @click="statusFilter = s">
          {{ t[s as keyof typeof t] || s }}
        </Button>
        <Separator orientation="vertical" class="h-5 mx-1 hidden md:block" />
        <Button size="sm" class="rounded-full px-4 h-8" @click="newPost">
          <PlusIcon class="h-3.5 w-3.5 mr-1" />{{ t.newPost }}
        </Button>
        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" aria-label="Refresh posts" :disabled="store.postsLoading" @click="loadPosts">
          <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" aria-hidden="true" />
        </Button>
        <Separator orientation="vertical" class="h-5 hidden md:block" />
        <Button variant="ghost" size="sm" class="text-xs h-8 text-muted-foreground hidden md:inline-flex" :title="t.exportAll" @click="exportAll">
          <DownloadIcon class="h-3.5 w-3.5 mr-1" />{{ t.exportAll }}
        </Button>
        <Button variant="ghost" size="sm" class="text-xs h-8 text-muted-foreground hidden md:inline-flex" :title="t.importZip" @click="importZip">
          <UploadIcon class="h-3.5 w-3.5 mr-1" />{{ t.importZip }}
        </Button>
      </div>
    </div>

    <!-- Tag chips -->
    <div v-if="allTags.length" class="flex gap-1.5 flex-wrap">
      <button
        v-for="tag in allTags" :key="tag"
        class="text-[10px] px-1.5 py-0.5 rounded-full bg-muted cursor-pointer transition-colors"
        :class="selectedTags.includes(tag) ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-muted/60'"
        @click="toggleTag(tag)">
        # {{ tag }}
      </button>
    </div>

    <!-- Stats -->
    <div class="flex gap-3 text-xs text-muted-foreground">
      <span>{{ t.totalPosts(store.posts.length) }}</span>
      <span class="text-border">·</span>
      <span>{{ t.draftCount(store.posts.filter(p => p.draft).length) }}</span>
      <span class="text-border">·</span>
      <span>{{ t.conflictCount(store.posts.filter(p => p.syncStatus === 'conflict').length) }}</span>
      <template v-if="filtered.length !== store.posts.length">
        <span class="text-border">·</span>
        <span>{{ t.filteredCount(filtered.length) }}</span>
      </template>
    </div>

    <!-- Loading -->
    <div v-if="store.postsLoading" class="stagger space-y-3">
      <SkeletonCard v-for="i in 5" :key="i" class="animate-fade-up" />
    </div>

    <!-- Empty -->
    <EmptyState
      v-else-if="filtered.length === 0"
      :title="searchQuery || statusFilter !== 'all' || selectedTags.length ? t.noSearchResults : t.noPosts"
      :description="searchQuery || statusFilter !== 'all' || selectedTags.length ? t.noSearchResultsDesc : t.noPostsDesc"
      :action-label="searchQuery || statusFilter !== 'all' || selectedTags.length ? undefined : t.newPost"
      @action="newPost"
    />

    <!-- Post list (virtual scrolling) -->
    <div
      v-else
      ref="parentRef"
      class="stagger overflow-auto"
      :style="{ height: 'calc(100vh - 280px)', minHeight: '400px' }"
    >
      <div
        class="relative w-full"
        :style="{ height: `${rowVirtualizer.getTotalSize()}px` }"
      >
        <div
          v-for="virtualRow in virtualRows"
          :key="String(virtualRow.key)"
          class="absolute left-0 right-0"
          :style="{
            top: `${virtualRow.start}px`,
            height: `${virtualRow.size}px`,
            paddingTop: '4px',
            paddingBottom: '4px',
          }"
        >
          <div
            class="post-card card-hover group relative flex items-stretch rounded-lg cursor-pointer bg-card border border-border/40 shadow-sm hover:border-border/80 hover:bg-muted/20 focus-visible:outline-2 focus-visible:outline-accent focus-visible:outline-offset-1 h-full"
            tabindex="0"
            role="button"
            :aria-label="filtered[virtualRow.index]?.title"
            @click="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)"
            @focus="focusedIndex = virtualRow.index"
            @keydown.enter.prevent="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)"
          >
            <!-- Selection checkbox -->
            <button
              class="flex items-center justify-center w-10 md:w-8 shrink-0 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
              :aria-label="selectedSlugs.has(filtered[virtualRow.index]!.slug) ? 'Deselect post' : 'Select post'"
              @click.stop="toggleSelect(filtered[virtualRow.index]!.slug)"
            >
              <CheckSquareIcon v-if="selectedSlugs.has(filtered[virtualRow.index]!.slug)" class="h-4 w-4 text-accent" />
              <SquareIcon v-else class="h-4 w-4" />
            </button>

            <!-- Status color bar -->
            <div class="w-1.5 shrink-0 rounded-l-lg" :class="[statusColor(filtered[virtualRow.index]!), filtered[virtualRow.index]!.draft ? 'opacity-40' : 'opacity-100']" />

            <div class="flex-1 flex flex-col justify-center gap-1.5 py-3 px-4 min-w-0">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 space-y-1">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-semibold truncate">{{ filtered[virtualRow.index]!.title || filtered[virtualRow.index]!.slug }}</span>
                    <AlertTriangleIcon v-if="filtered[virtualRow.index]!.large" class="h-3 w-3 text-warn shrink-0" :title="t.large" />
                  </div>
                  <div v-if="filtered[virtualRow.index]!.description" class="text-xs text-muted-foreground line-clamp-1">
                    {{ filtered[virtualRow.index]!.description }}
                  </div>
                </div>

                <!-- Actions (always visible, subtle) -->
                <div class="flex gap-1 shrink-0 text-muted-foreground/40">
                  <Button variant="ghost" size="icon" class="h-8 w-8 md:h-7 md:w-7 hover:text-foreground hover:bg-muted/50" :aria-label="t.edit" :title="t.edit" @click.stop="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)">
                    <EditIcon class="h-3.5 w-3.5" aria-hidden="true" />
                  </Button>
                  <Button variant="ghost" size="icon" class="h-8 w-8 md:h-7 md:w-7 hover:text-foreground hover:bg-muted/50" :aria-label="t.preview" :title="t.preview" @click.stop="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}/preview`)">
                    <EyeIcon class="h-3.5 w-3.5" aria-hidden="true" />
                  </Button>
                  <Button variant="ghost" size="icon" class="h-8 w-8 md:h-7 md:w-7 hover:text-destructive hover:bg-destructive/10" :aria-label="t.moveToTrash" :title="t.moveToTrash" @click.stop="deletePost(filtered[virtualRow.index]!)">
                    <Trash2Icon class="h-3.5 w-3.5" aria-hidden="true" />
                  </Button>
                </div>
              </div>

              <div class="flex items-center gap-2 mt-0.5">
                <span class="text-[11px] text-muted-foreground" :title="filtered[virtualRow.index]!.date">{{ relDate(filtered[virtualRow.index]!.date) }}</span>
                <Badge :variant="statusBadge(filtered[virtualRow.index]!).variant" class="text-[10px] h-4 px-1.5 font-normal">{{ statusBadge(filtered[virtualRow.index]!).label }}</Badge>
                <span v-for="tag in (filtered[virtualRow.index]!.tags || []).slice(0, 3)" :key="tag" class="text-[10px] px-1.5 py-0.5 rounded-full bg-muted text-muted-foreground/70"># {{ tag }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bulk action bar -->
    <div
      v-if="selectedSlugs.size > 0"
      class="fixed bottom-4 left-4 right-4 md:bottom-6 md:left-1/2 md:right-auto md:-translate-x-1/2 z-50 flex items-center gap-2 md:gap-3 px-4 md:px-5 py-3 bg-card border border-border rounded-xl shadow-lg animate-fade-up"
    >
      <span class="text-sm text-muted-foreground">{{ t.selectedCount(selectedSlugs.size) }}</span>
      <Separator orientation="vertical" class="h-5" />
      <Button variant="ghost" size="sm" class="text-xs h-7" @click="selectAll">
        {{ t.selectAll }}
      </Button>
      <Button variant="ghost" size="sm" class="text-xs h-7" @click="deselectAll">
        {{ t.deselectAll }}
      </Button>
      <Separator orientation="vertical" class="h-5" />
      <Button variant="outline" size="sm" class="text-xs h-7 text-destructive" @click="bulkTrash">
        <Trash2Icon class="h-3 w-3 mr-1" />{{ t.trashSelected }}
      </Button>
      <Button size="sm" class="text-xs h-7 rounded-full" @click="bulkPublish">
        <SendIcon class="h-3 w-3 mr-1" />{{ t.publishSelected }}
      </Button>
    </div>
  </div>
</template>
