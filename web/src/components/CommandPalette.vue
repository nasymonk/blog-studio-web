<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useEventListener } from '@vueuse/core'
import { SearchIcon } from 'lucide-vue-next'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ 'update:open': [val: boolean] }>()

const router = useRouter()
const store = useStore()
const { t, setLang } = useI18n()
const { toggle: toggleTheme } = useTheme()

const query = ref('')
const cursor = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

interface Cmd { label: string; description?: string; shortcut?: string; action: () => void }

const fixedCmds: Cmd[] = [
  { label: '新建文章', shortcut: '', action: () => { close(); const s = prompt('Slug:'); if (s) router.push(`/posts/${encodeURIComponent(s)}`) } },
  { label: '切换主题', shortcut: '', action: () => { close(); toggleTheme() } },
  { label: '切换语言', shortcut: '', action: () => { close(); setLang(t.value === t.value ? 'en' : 'zh') } },
  { label: '跳到设置', shortcut: '', action: () => go('/settings') },
  { label: '跳到审计', shortcut: '', action: () => go('/settings/audit') },
  { label: '跳到健康检查', shortcut: '', action: () => go('/health') },
  { label: '退出登录', shortcut: '', action: () => { close(); store.session.authenticated = false; router.push('/login') } },
]

const postCmds = computed<Cmd[]>(() =>
  store.posts
    .filter(p => {
      const q = query.value.trim().toLowerCase()
      if (!q) return true
      return p.title.toLowerCase().includes(q) || p.slug.includes(q)
    })
    .slice(0, 8)
    .map(p => ({ label: p.title, description: p.slug, action: () => go(`/posts/${p.slug}`) }))
)

const filteredCmds = computed(() => {
  const q = query.value.trim().toLowerCase()
  return q ? fixedCmds.filter(c => c.label.toLowerCase().includes(q)) : fixedCmds
})

const items = computed<Cmd[]>(() => [...filteredCmds.value, ...postCmds.value])

const hasCommands = computed(() => filteredCmds.value.length > 0)
const hasPosts = computed(() => postCmds.value.length > 0)

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
          placeholder="搜索命令或文章…"
          autocomplete="off"
        />
        <kbd class="hidden sm:inline-flex h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 text-[10px] text-muted-foreground/60 font-mono">ESC</kbd>
      </div>

      <ScrollArea class="max-h-[360px]">
        <div class="py-2">
          <!-- Commands group -->
          <template v-if="hasCommands">
            <div class="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/50">命令</div>
            <button
              v-for="(item, i) in filteredCmds"
              :key="'cmd-' + i"
              class="relative flex w-full items-center gap-3 px-4 py-2 text-sm cursor-pointer transition-colors"
              :class="cursor === i ? 'bg-accent/8' : 'hover:bg-muted/50'"
              role="option"
              :aria-selected="cursor === i"
              @click="select(item)"
              @mouseenter="cursor = i"
            >
              <!-- Active indicator -->
              <div v-if="cursor === i" class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-accent rounded-full" />
              <span class="flex-1 text-left font-sans">{{ item.label }}</span>
              <kbd v-if="item.shortcut" class="text-[10px] font-mono text-muted-foreground/40 bg-muted px-1.5 py-0.5 rounded border border-border/50">{{ item.shortcut }}</kbd>
            </button>
          </template>

          <!-- Posts group -->
          <template v-if="hasPosts">
            <div class="px-4 py-1.5 mt-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/50">文章</div>
            <button
              v-for="(item, j) in postCmds"
              :key="'post-' + j"
              class="relative flex w-full items-center gap-3 px-4 py-2 text-sm cursor-pointer transition-colors"
              :class="cursor === filteredCmds.length + j ? 'bg-accent/8' : 'hover:bg-muted/50'"
              role="option"
              :aria-selected="cursor === filteredCmds.length + j"
              @click="select(item)"
              @mouseenter="cursor = filteredCmds.length + j"
            >
              <div v-if="cursor === filteredCmds.length + j" class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-accent rounded-full" />
              <span class="flex-1 text-left font-serif truncate">{{ item.label }}</span>
              <span v-if="item.description" class="text-[11px] text-muted-foreground/40 font-mono truncate max-w-[120px]">{{ item.description }}</span>
            </button>
          </template>

          <!-- Empty -->
          <div v-if="items.length === 0" class="py-10 text-center text-sm text-muted-foreground/50">无匹配结果</div>
        </div>
      </ScrollArea>
    </DialogContent>
  </Dialog>
</template>
