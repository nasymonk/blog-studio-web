<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useEventListener } from '@vueuse/core'
import { SearchIcon } from 'lucide-vue-next'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
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

interface Cmd { label: string; description?: string; action: () => void }

const fixedCmds: Cmd[] = [
  { label: '新建文章', action: () => { close(); const s = prompt('Slug:'); if (s) router.push(`/posts/${encodeURIComponent(s)}`) } },
  { label: '切换主题', action: () => { close(); toggleTheme() } },
  { label: '切换语言', action: () => { close(); setLang(t.value === t.value ? 'en' : 'zh') } },
  { label: '跳到设置', action: () => go('/settings') },
  { label: '跳到审计', action: () => go('/settings/audit') },
  { label: '跳到健康检查', action: () => go('/health') },
  { label: '退出登录', action: () => { close(); store.session.authenticated = false; router.push('/login') } },
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

const items = computed<Cmd[]>(() => {
  const q = query.value.trim().toLowerCase()
  const cmds = q
    ? fixedCmds.filter(c => c.label.toLowerCase().includes(q))
    : fixedCmds
  return [...cmds, ...postCmds.value]
})

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
    <DialogContent class="overflow-hidden p-0 gap-0 max-w-[560px] top-[15vh] translate-y-0" @keydown="onKeydown">
      <div class="flex items-center gap-2 border-b border-border px-4 py-3">
        <SearchIcon class="h-4 w-4 text-muted-foreground shrink-0" />
        <Input
          ref="inputRef"
          v-model="query"
          class="border-0 shadow-none focus-visible:ring-0 text-sm h-auto p-0"
          placeholder="搜索命令或文章…"
          autocomplete="off"
        />
      </div>
      <ScrollArea class="max-h-[360px]">
        <div class="p-1.5">
          <button
            v-for="(item, i) in items"
            :key="i"
            class="flex w-full items-center justify-between rounded-md px-3 py-2.5 text-sm cursor-pointer transition-colors"
            :class="cursor === i ? 'bg-accent text-accent-foreground' : 'hover:bg-muted'"
            role="option"
            :aria-selected="cursor === i"
            @click="select(item)"
            @mouseenter="cursor = i"
          >
            <span class="flex-1 text-left">{{ item.label }}</span>
            <span v-if="item.description" class="text-xs" :class="cursor === i ? 'text-accent-foreground/70' : 'text-muted-foreground'">{{ item.description }}</span>
          </button>
          <div v-if="items.length === 0" class="py-8 text-center text-sm text-muted-foreground">无匹配结果</div>
        </div>
      </ScrollArea>
    </DialogContent>
  </Dialog>
</template>
