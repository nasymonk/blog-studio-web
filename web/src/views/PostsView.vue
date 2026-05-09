<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { PlusIcon, SearchIcon, RefreshCwIcon, EditIcon, EyeIcon, AlertTriangleIcon, FileTextIcon, Trash2Icon } from 'lucide-vue-next'
import { useDebounceFn } from '@vueuse/core'
import { api } from '@/services/api'
import type { PostState } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { filterPosts } from '@/utils/filterPosts'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
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
const page = ref(Number(route.query.page) || 1)
const PAGE_SIZE = 25

const searchInputRef = ref<HTMLInputElement | null>(null)
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
    focusedIndex.value = Math.min(focusedIndex.value + 1, paginated.value.length - 1)
    focusCard(focusedIndex.value)
  } else if (e.key === 'ArrowUp' || e.key === 'k') {
    e.preventDefault()
    focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
    focusCard(focusedIndex.value)
  } else if (e.key === 'Enter' && focusedIndex.value >= 0) {
    const slug = paginated.value[focusedIndex.value]?.slug
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

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / PAGE_SIZE)))
const paginated = computed(() => filtered.value.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE))

const syncUrl = useDebounceFn(() => {
  const q: Record<string, string> = {}
  if (searchQuery.value) q.q = searchQuery.value
  if (statusFilter.value !== 'all') q.status = statusFilter.value
  if (selectedTags.value.length) q.tags = selectedTags.value.join(',')
  if (sortBy.value !== 'date-desc') q.sort = sortBy.value
  if (page.value > 1) q.page = String(page.value)
  router.replace({ query: q })
}, 300)

watch([searchQuery, statusFilter, selectedTags, sortBy], () => { page.value = 1; syncUrl() }, { deep: true })
watch(page, syncUrl)

async function loadPosts() {
  store.postsLoading = true
  try {
    store.posts = await api.posts()
  } catch (e: any) {
    notify.error(e)
  } finally {
    store.postsLoading = false
  }
}

async function newPost() {
  const slug = prompt('新文章 slug（英文、连字符）：')
  if (!slug) return
  router.push(`/posts/${encodeURIComponent(slug)}`)
}

async function deletePost(post: PostState) {
  if (!confirm(`将《${post.title || post.slug}》移到回收站？可在回收站还原。`)) return
  try {
    await api.deletePost(post.slug)
    notify.success(`已删除《${post.title || post.slug}》`)
    store.posts = store.posts.filter(p => p.slug !== post.slug)
  } catch (e: any) {
    notify.error(e)
  }
}

const statusVariant: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  conflict: 'destructive',
  stale: 'secondary',
  draft: 'outline',
  clean: 'default',
}

function statusBadge(p: PostState): { variant: 'default' | 'secondary' | 'destructive' | 'outline'; label: string } {
  if (p.syncStatus === 'conflict') return { variant: 'destructive', label: t.value.conflicts }
  if (p.syncStatus === 'stale') return { variant: 'secondary', label: t.value.stale }
  if (p.draft) return { variant: 'outline', label: t.value.draft }
  if (p.syncStatus === 'clean') return { variant: 'default', label: t.value.published }
  return { variant: 'outline', label: p.syncStatus }
}

function relDate(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 30) return `${days} 天前`
  if (days < 365) return `${Math.floor(days / 30)} 月前`
  return `${Math.floor(days / 365)} 年前`
}

function toggleTag(tag: string) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) selectedTags.value.splice(idx, 1)
  else selectedTags.value.push(tag)
}

onMounted(loadPosts)
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- Toolbar -->
    <div class="flex items-center gap-2 flex-wrap">
      <div class="relative flex-1 max-w-[340px]">
        <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
        <Input ref="searchInputRef" v-model="searchQuery" class="pl-9" :placeholder="t.search" />
      </div>
      <Select v-model="sortBy">
        <SelectTrigger class="w-auto min-w-[130px]"><SelectValue /></SelectTrigger>
        <SelectContent>
          <SelectItem value="date-desc">{{ t.sortDate }} ↓</SelectItem>
          <SelectItem value="date-asc">{{ t.sortDate }} ↑</SelectItem>
          <SelectItem value="title">{{ t.sortTitle }}</SelectItem>
          <SelectItem value="status">{{ t.sortStatus }}</SelectItem>
        </SelectContent>
      </Select>
      <div class="flex gap-1 flex-wrap">
        <Button
          v-for="s in ['all','drafts','published','conflicts','stale']" :key="s"
          variant="ghost" size="sm"
          :class="{ 'bg-primary text-primary-foreground hover:bg-primary/90': statusFilter === s }"
          @click="statusFilter = s">
          {{ t[s as keyof typeof t] || s }}
        </Button>
      </div>
      <div class="ml-auto flex gap-2">
        <Button @click="newPost"><PlusIcon class="h-4 w-4 mr-1" />{{ t.newPost }}</Button>
        <Button variant="ghost" size="icon" class="h-9 w-9" :title="t.loading" :disabled="store.postsLoading" @click="loadPosts">
          <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" />
        </Button>
      </div>
    </div>

    <!-- Tag chips -->
    <div v-if="allTags.length" class="flex gap-1.5 flex-wrap">
      <Button
        v-for="tag in allTags" :key="tag" variant="ghost" size="sm"
        :class="{ 'bg-primary text-primary-foreground hover:bg-primary/90': selectedTags.includes(tag) }"
        @click="toggleTag(tag)">
        # {{ tag }}
      </Button>
    </div>

    <!-- Stats row -->
    <div class="flex gap-2 text-sm text-muted-foreground">
      <span>共 {{ store.posts.length }} 篇</span>
      <span class="text-border">·</span>
      <span>草稿 {{ store.posts.filter(p => p.draft).length }}</span>
      <span class="text-border">·</span>
      <span>冲突 {{ store.posts.filter(p => p.syncStatus === 'conflict').length }}</span>
      <template v-if="filtered.length !== store.posts.length">
        <span class="text-border">·</span>
        <span>筛选后 {{ filtered.length }} 篇</span>
      </template>
    </div>

    <!-- Loading skeletons -->
    <template v-if="store.postsLoading">
      <Skeleton v-for="i in 5" :key="i" class="h-[88px] rounded-xl" />
    </template>

    <!-- Empty -->
    <div v-else-if="paginated.length === 0" class="text-center py-16 text-muted-foreground space-y-2">
      <FileTextIcon class="h-10 w-10 mx-auto opacity-40" />
      <p class="font-medium">{{ searchQuery || statusFilter !== 'all' || selectedTags.length ? t.noResults : t.noPosts }}</p>
      <p v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length" class="text-sm">点击「新建文章」创建你的第一篇文章</p>
      <Button v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length" @click="newPost">
        <PlusIcon class="h-3.5 w-3.5 mr-1" />{{ t.newPost }}
      </Button>
    </div>

    <!-- Post list -->
    <template v-else>
      <Card
        v-for="(post, i) in paginated"
        :key="post.slug"
        class="post-card cursor-pointer transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        tabindex="0"
        role="button"
        :aria-label="post.title"
        @click="router.push(`/posts/${encodeURIComponent(post.slug)}`)"
        @focus="focusedIndex = i"
        @keydown.enter.prevent="router.push(`/posts/${encodeURIComponent(post.slug)}`)"
      >
        <CardContent class="py-3 px-4 flex items-center justify-between gap-3">
          <div class="flex-1 min-w-0 space-y-1">
            <div class="flex items-center gap-2">
              <span class="font-medium truncate">{{ post.title || post.slug }}</span>
              <AlertTriangleIcon v-if="post.large" class="h-3.5 w-3.5 text-warn shrink-0" :title="t.large" />
            </div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <span :title="post.date">{{ relDate(post.date) }}</span>
              <Badge :variant="statusBadge(post).variant" class="text-xs">{{ statusBadge(post).label }}</Badge>
              <span v-for="tag in (post.tags || []).slice(0, 4)" :key="tag" class="text-xs text-muted-foreground"># {{ tag }}</span>
            </div>
          </div>
          <div class="flex gap-0.5 shrink-0 opacity-0 group-hover:opacity-100 focus-within:opacity-100" :class="{ 'opacity-100': focusedIndex === i }">
            <Button variant="ghost" size="icon" class="h-7 w-7" :title="t.edit" @click.stop="router.push(`/posts/${encodeURIComponent(post.slug)}`)">
              <EditIcon class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7" :title="t.preview" @click.stop="router.push(`/posts/${encodeURIComponent(post.slug)}/preview`)">
              <EyeIcon class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-destructive hover:text-destructive" title="移到回收站" @click.stop="deletePost(post)">
              <Trash2Icon class="h-3.5 w-3.5" />
            </Button>
          </div>
        </CardContent>
      </Card>
    </template>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between">
      <Button variant="ghost" size="sm" :disabled="page === 1" @click="page--">‹ 上一页</Button>
      <span class="text-sm text-muted-foreground">{{ page }} / {{ totalPages }}</span>
      <Button variant="ghost" size="sm" :disabled="page === totalPages" @click="page++">下一页 ›</Button>
    </div>
  </div>
</template>
