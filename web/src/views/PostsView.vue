<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { PlusIcon, SearchIcon, RefreshCwIcon, EditIcon, EyeIcon, AlertTriangleIcon, FileTextIcon, Trash2Icon } from 'lucide-vue-next'
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
import { Skeleton } from '@/components/ui/skeleton'

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
  estimateSize: () => 72,
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

onMounted(loadPosts)
</script>

<template>
  <div class="flex flex-col gap-5 max-w-4xl">
    <!-- Search bar: underline style -->
    <div class="flex items-end gap-3">
      <div class="relative flex-1 max-w-[360px]">
        <SearchIcon class="absolute left-0 bottom-3 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
        <Input
          ref="searchInputRef"
          v-model="searchQuery"
          class="border-0 border-b border-border rounded-none bg-transparent pl-6 pb-2 h-auto focus-visible:ring-0 focus-visible:border-accent transition-colors"
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
      <div class="flex gap-1 flex-wrap ml-auto">
        <Button
          v-for="s in ['all','drafts','published','conflicts','stale']" :key="s"
          variant="ghost" size="sm" class="text-xs h-7"
          :class="statusFilter === s ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
          @click="statusFilter = s">
          {{ t[s as keyof typeof t] || s }}
        </Button>
      </div>
      <Button size="sm" class="rounded-full px-4 h-8" @click="newPost">
        <PlusIcon class="h-3.5 w-3.5 mr-1" />{{ t.newPost }}
      </Button>
      <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" :disabled="store.postsLoading" @click="loadPosts">
        <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" />
      </Button>
    </div>

    <!-- Tag chips -->
    <div v-if="allTags.length" class="flex gap-1.5 flex-wrap">
      <button
        v-for="tag in allTags" :key="tag"
        class="font-deco text-[11px] px-2.5 py-0.5 rounded-full border transition-colors cursor-pointer"
        :class="selectedTags.includes(tag) ? 'bg-accent text-accent-foreground border-accent' : 'border-border text-muted-foreground hover:border-accent/40'"
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
      <Skeleton v-for="i in 5" :key="i" class="h-[72px] animate-fade-up" />
    </div>

    <!-- Empty -->
    <div v-else-if="filtered.length === 0" class="text-center py-20 text-muted-foreground space-y-3">
      <FileTextIcon class="h-10 w-10 mx-auto opacity-30" />
      <p class="font-serif font-medium">{{ searchQuery || statusFilter !== 'all' || selectedTags.length ? t.noResults : t.noPosts }}</p>
      <p v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length" class="text-sm">{{ t.createFirstPost }}</p>
      <Button v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length" class="rounded-full px-5" @click="newPost">
        <PlusIcon class="h-3.5 w-3.5 mr-1" />{{ t.newPost }}
      </Button>
    </div>

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
            class="post-card group relative flex items-stretch rounded cursor-pointer transition-all duration-200 hover:-translate-y-px hover:shadow-sm bg-card border border-border/60 focus-visible:outline-2 focus-visible:outline-accent focus-visible:outline-offset-1 h-full"
            tabindex="0"
            role="button"
            :aria-label="filtered[virtualRow.index]?.title"
            @click="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)"
            @focus="focusedIndex = virtualRow.index"
            @keydown.enter.prevent="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)"
          >
            <!-- Status color bar -->
            <div class="w-[3px] shrink-0 rounded-l" :class="statusColor(filtered[virtualRow.index]!)" />

            <div class="flex-1 flex items-center justify-between gap-3 px-4 py-3 min-w-0">
              <div class="min-w-0 space-y-1">
                <div class="flex items-center gap-2">
                  <span class="font-serif font-medium text-sm truncate">{{ filtered[virtualRow.index]!.title || filtered[virtualRow.index]!.slug }}</span>
                  <AlertTriangleIcon v-if="filtered[virtualRow.index]!.large" class="h-3 w-3 text-warn shrink-0" :title="t.large" />
                </div>
                <div class="flex items-center gap-2">
                  <span class="text-[11px] text-muted-foreground" :title="filtered[virtualRow.index]!.date">{{ relDate(filtered[virtualRow.index]!.date) }}</span>
                  <Badge :variant="statusBadge(filtered[virtualRow.index]!).variant" class="text-[10px] h-4 px-1.5">{{ statusBadge(filtered[virtualRow.index]!).label }}</Badge>
                  <span v-for="tag in (filtered[virtualRow.index]!.tags || []).slice(0, 3)" :key="tag" class="font-deco text-[11px] text-muted-foreground/70"># {{ tag }}</span>
                </div>
              </div>

              <!-- Actions (visible on focus/hover) -->
              <div class="flex gap-0.5 shrink-0 opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100"
                :class="{ 'opacity-100': focusedIndex === virtualRow.index }">
                <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :title="t.edit" @click.stop="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}`)">
                  <EditIcon class="h-3.5 w-3.5" />
                </Button>
                <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :title="t.preview" @click.stop="router.push(`/posts/${encodeURIComponent(filtered[virtualRow.index]!.slug)}/preview`)">
                  <EyeIcon class="h-3.5 w-3.5" />
                </Button>
                <Button variant="ghost" size="icon" class="h-7 w-7 text-destructive/60 hover:text-destructive" :title="t.moveToTrash" @click.stop="deletePost(filtered[virtualRow.index]!)">
                  <Trash2Icon class="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
