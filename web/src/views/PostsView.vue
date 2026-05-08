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

function statusBadge(p: PostState): { cls: string; label: string } {
  if (p.syncStatus === 'conflict') return { cls: 'badge-error', label: t.value.conflicts }
  if (p.syncStatus === 'stale') return { cls: 'badge-warn', label: t.value.stale }
  if (p.draft) return { cls: 'badge-neutral', label: t.value.draft }
  if (p.syncStatus === 'clean') return { cls: 'badge-ok', label: t.value.published }
  return { cls: 'badge-neutral', label: p.syncStatus }
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
  <div class="page-content" style="display:flex;flex-direction:column;gap:16px">
    <!-- Toolbar -->
    <div class="flex items-center gap-2">
      <div style="position:relative;flex:1;max-width:340px">
        <SearchIcon :size="15" style="position:absolute;left:10px;top:50%;transform:translateY(-50%);color:var(--tertiary);pointer-events:none" />
        <input ref="searchInputRef" v-model="searchQuery" class="input" style="padding-left:32px" :placeholder="t.search" />
      </div>
      <select v-model="sortBy" class="input" style="width:auto;min-width:130px">
        <option value="date-desc">{{ t.sortDate }} ↓</option>
        <option value="date-asc">{{ t.sortDate }} ↑</option>
        <option value="title">{{ t.sortTitle }}</option>
        <option value="status">{{ t.sortStatus }}</option>
      </select>
      <div style="display:flex;gap:4px;flex-wrap:wrap">
        <button
v-for="s in ['all','drafts','published','conflicts','stale']" :key="s"
          class="btn btn-sm" :class="{ 'btn-primary': statusFilter === s }"
          @click="statusFilter = s">
          {{ t[s as keyof typeof t] || s }}
        </button>
      </div>
      <div style="margin-left:auto;display:flex;gap:8px">
        <button class="btn btn-primary" @click="newPost">
          <PlusIcon :size="15" />{{ t.newPost }}
        </button>
        <button class="btn btn-ghost btn-icon" :title="t.loading" :disabled="store.postsLoading" @click="loadPosts">
          <RefreshCwIcon :size="15" :class="{ spin: store.postsLoading }" />
        </button>
      </div>
    </div>

    <!-- Tag chips -->
    <div v-if="allTags.length" class="flex gap-2 flex-wrap">
      <button
v-for="tag in allTags" :key="tag" class="btn btn-sm"
        :class="{ 'btn-primary': selectedTags.includes(tag) }"
        @click="toggleTag(tag)">
        # {{ tag }}
      </button>
    </div>

    <!-- Stats row -->
    <div class="flex gap-2 text-sm text-muted">
      <span>共 {{ store.posts.length }} 篇</span>
      <span class="sep">·</span>
      <span>草稿 {{ store.posts.filter(p => p.draft).length }}</span>
      <span class="sep">·</span>
      <span>冲突 {{ store.posts.filter(p => p.syncStatus === 'conflict').length }}</span>
      <span v-if="filtered.length !== store.posts.length" class="sep">·</span>
      <span v-if="filtered.length !== store.posts.length">筛选后 {{ filtered.length }} 篇</span>
    </div>

    <!-- Loading skeletons -->
    <template v-if="store.postsLoading">
      <div v-for="i in 5" :key="i" class="skeleton" style="height:88px;border-radius:12px"></div>
    </template>

    <!-- Empty -->
    <div v-else-if="paginated.length === 0" class="empty-state">
      <FileTextIcon :size="40" />
      <strong>{{ searchQuery || statusFilter !== 'all' || selectedTags.length ? t.noResults : t.noPosts }}</strong>
      <p v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length">点击「新建文章」创建你的第一篇文章</p>
      <button v-if="!searchQuery && statusFilter === 'all' && !selectedTags.length" class="btn btn-primary" @click="newPost">
        <PlusIcon :size="14" />{{ t.newPost }}
      </button>
    </div>

    <!-- Post list -->
    <template v-else>
      <div v-for="(post, i) in paginated" :key="post.slug" class="post-card" tabindex="0" role="button" :aria-label="post.title" @click="router.push(`/posts/${encodeURIComponent(post.slug)}`)" @focus="focusedIndex = i" @keydown.enter.prevent="router.push(`/posts/${encodeURIComponent(post.slug)}`)">
        <div class="post-card-body">
          <div class="flex items-center gap-2">
            <span class="post-card-title">{{ post.title || post.slug }}</span>
            <AlertTriangleIcon v-if="post.large" :size="13" style="color:var(--warn);flex-shrink:0" :title="t.large" />
          </div>
          <div class="post-card-meta">
            <span class="post-card-date" :title="post.date">{{ relDate(post.date) }}</span>
            <span class="badge" :class="statusBadge(post).cls">{{ statusBadge(post).label }}</span>
            <div class="post-card-tags">
              <span v-for="tag in (post.tags || []).slice(0, 4)" :key="tag" class="tag-chip"># {{ tag }}</span>
            </div>
          </div>
        </div>
        <div class="post-actions">
          <button class="btn btn-ghost btn-icon btn-sm" :title="t.edit" @click.stop="router.push(`/posts/${encodeURIComponent(post.slug)}`)">
            <EditIcon :size="14" />
          </button>
          <button class="btn btn-ghost btn-icon btn-sm" :title="t.preview" @click.stop="router.push(`/posts/${encodeURIComponent(post.slug)}/preview`)">
            <EyeIcon :size="14" />
          </button>
          <button class="btn btn-ghost btn-icon btn-sm btn-danger-text" title="移到回收站" @click.stop="deletePost(post)">
            <Trash2Icon :size="14" />
          </button>
        </div>
      </div>
    </template>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center gap-2 justify-between">
      <button class="btn btn-sm" :disabled="page === 1" @click="page--">‹ 上一页</button>
      <span class="text-sm text-muted">{{ page }} / {{ totalPages }}</span>
      <button class="btn btn-sm" :disabled="page === totalPages" @click="page++">下一页 ›</button>
    </div>
  </div>
</template>
