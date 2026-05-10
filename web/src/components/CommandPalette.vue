<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useEventListener } from '@vueuse/core'
import {
  SearchIcon,
  FilePlusIcon,
  PaletteIcon,
  SettingsIcon,
  HeartPulseIcon,
  LanguagesIcon,
  ShieldIcon,
  LogOutIcon,
  FileTextIcon,
} from 'lucide-vue-next'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ 'update:open': [val: boolean] }>()

const router = useRouter()
const store = useStore()
const { t, lang, setLang } = useI18n()
const { toggle: toggleTheme } = useTheme()

const query = ref('')
const cursor = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

interface Cmd {
  label: string
  description?: string
  shortcut?: string
  icon: typeof SearchIcon
  action: () => void
}

/** Simple subsequence fuzzy match — returns true if query chars appear in order within text */
function fuzzyMatch(query: string, text: string): boolean {
  if (!query) return true
  const q = query.toLowerCase()
  const t = text.toLowerCase()
  let qi = 0
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) qi++
  }
  return qi === q.length
}

const quickActions = computed<Cmd[]>(() => [
  {
    label: t.value.newPost,
    icon: FilePlusIcon,
    action: () => {
      close()
      const s = prompt(t.value.newPostSlug)
      if (s) router.push(`/posts/${encodeURIComponent(s)}`)
    },
  },
  {
    label: t.value.theme,
    icon: PaletteIcon,
    action: () => {
      close()
      toggleTheme()
    },
  },
  {
    label: t.value.switchLanguage,
    icon: LanguagesIcon,
    action: () => {
      close()
      setLang(lang.value === 'zh' ? 'en' : 'zh')
    },
  },
  {
    label: t.value.settings,
    icon: SettingsIcon,
    action: () => go('/settings'),
  },
  {
    label: t.value.auditLog,
    icon: ShieldIcon,
    action: () => go('/settings/audit'),
  },
  {
    label: t.value.health,
    icon: HeartPulseIcon,
    action: () => go('/health'),
  },
  {
    label: t.value.logout,
    icon: LogOutIcon,
    action: () => {
      close()
      store.session.authenticated = false
      router.push('/login')
    },
  },
])

const filteredActions = computed(() => {
  const q = query.value.trim()
  if (!q) return quickActions.value
  return quickActions.value.filter(c => fuzzyMatch(q, c.label))
})

const recentPosts = computed<Cmd[]>(() => {
  const q = query.value.trim()
  const sorted = [...store.posts].sort((a, b) => b.date.localeCompare(a.date))
  const filtered = q
    ? sorted.filter(p => fuzzyMatch(q, p.title) || fuzzyMatch(q, p.slug))
    : sorted
  return filtered.slice(0, 8).map(p => ({
    label: p.title,
    description: p.slug,
    icon: FileTextIcon,
    action: () => go(`/posts/${p.slug}`),
  }))
})

const items = computed<Cmd[]>(() => [...filteredActions.value, ...recentPosts.value])

const hasActions = computed(() => filteredActions.value.length > 0)
const hasPosts = computed(() => recentPosts.value.length > 0)
const actionCount = computed(() => filteredActions.value.length)

watch(() => props.open, (v) => {
  if (v) {
    query.value = ''
    cursor.value = 0
    nextTick(() => inputRef.value?.focus())
  }
})

watch(query, () => { cursor.value = 0 })

function go(path: string) {
  close()
  router.push(path)
}

function close() {
  emit('update:open', false)
}

function select(item: Cmd) {
  item.action()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') { e.preventDefault(); cursor.value = Math.min(cursor.value + 1, items.value.length - 1) }
  else if (e.key === 'ArrowUp') { e.preventDefault(); cursor.value = Math.max(cursor.value - 1, 0) }
  else if (e.key === 'Enter') { e.preventDefault(); if (items.value[cursor.value]) select(items.value[cursor.value]) }
  else if (e.key === 'Escape') close()
}
</script>

<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="overflow-hidden p-0 gap-0 max-w-[520px] top-[18vh] translate-y-0 rounded-lg" @keydown="onKeydown">
      <!-- Search -->
      <div class="flex items-center gap-3 px-4 py-3.5 border-b border-border">
        <SearchIcon class="h-4 w-4 text-muted-foreground/50 shrink-0" />
        <input
          ref="inputRef"
          v-model="query"
          class="flex-1 bg-transparent outline-none text-sm font-serif placeholder:text-muted-foreground/40"
          :placeholder="t.commandPalettePlaceholder"
          autocomplete="off"
        />
        <kbd class="hidden sm:inline-flex h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 text-[10px] text-muted-foreground/60 font-mono">ESC</kbd>
      </div>

      <ScrollArea class="max-h-[360px]">
        <div class="py-2">
          <!-- Quick Actions group -->
          <template v-if="hasActions">
            <div class="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/50">{{ t.quickActions }}</div>
            <button
              v-for="(item, i) in filteredActions"
              :key="'action-' + i"
              class="relative flex w-full items-center gap-3 px-4 py-2 text-sm cursor-pointer transition-colors"
              :class="cursor === i ? 'bg-accent/8' : 'hover:bg-muted/50'"
              role="option"
              :aria-selected="cursor === i"
              @click="select(item)"
              @mouseenter="cursor = i"
            >
              <!-- Active indicator -->
              <div v-if="cursor === i" class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-accent rounded-full" />
              <component :is="item.icon" class="h-4 w-4 text-muted-foreground/60 shrink-0" />
              <span class="flex-1 text-left font-sans">{{ item.label }}</span>
              <kbd v-if="item.shortcut" class="text-[10px] font-mono text-muted-foreground/40 bg-muted px-1.5 py-0.5 rounded border border-border/50">{{ item.shortcut }}</kbd>
            </button>
          </template>

          <!-- Recent Posts group -->
          <template v-if="hasPosts">
            <div class="px-4 py-1.5 mt-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/50">{{ t.recentPosts }}</div>
            <button
              v-for="(item, j) in recentPosts"
              :key="'post-' + j"
              class="relative flex w-full items-center gap-3 px-4 py-2 text-sm cursor-pointer transition-colors"
              :class="cursor === actionCount + j ? 'bg-accent/8' : 'hover:bg-muted/50'"
              role="option"
              :aria-selected="cursor === actionCount + j"
              @click="select(item)"
              @mouseenter="cursor = actionCount + j"
            >
              <div v-if="cursor === actionCount + j" class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-accent rounded-full" />
              <component :is="item.icon" class="h-4 w-4 text-muted-foreground/60 shrink-0" />
              <span class="flex-1 text-left font-serif truncate">{{ item.label }}</span>
              <span v-if="item.description" class="text-[11px] text-muted-foreground/40 font-mono truncate max-w-[120px]">{{ item.description }}</span>
            </button>
          </template>

          <!-- Empty -->
          <div v-if="items.length === 0" class="py-10 text-center space-y-2 animate-fade-in">
            <SearchIcon class="h-6 w-6 mx-auto text-muted-foreground/25" />
            <p class="text-sm text-muted-foreground/50">{{ t.noMatchResults }}</p>
          </div>
        </div>
      </ScrollArea>
    </DialogContent>
  </Dialog>
</template>
