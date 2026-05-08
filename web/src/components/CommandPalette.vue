<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useEventListener } from '@vueuse/core'
import { SearchIcon } from 'lucide-vue-next'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'

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

useEventListener('keydown', (e: KeyboardEvent) => {
  if (props.open && e.key === 'Escape') close()
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="palette-backdrop" @click.self="close">
      <div class="palette-dialog" role="dialog" aria-modal="true" aria-label="命令面板" @keydown="onKeydown">
        <div class="palette-search">
          <SearchIcon :size="16" class="palette-icon" />
          <input
            ref="inputRef"
            v-model="query"
            class="palette-input"
            placeholder="搜索命令或文章…"
            autocomplete="off"
          />
        </div>
        <ul class="palette-list" role="listbox">
          <li
            v-for="(item, i) in items"
            :key="i"
            class="palette-item"
            :class="{ active: cursor === i }"
            role="option"
            :aria-selected="cursor === i"
            @click="select(item)"
            @mouseenter="cursor = i"
          >
            <span class="palette-label">{{ item.label }}</span>
            <span v-if="item.description" class="palette-desc">{{ item.description }}</span>
          </li>
          <li v-if="items.length === 0" class="palette-empty">无匹配结果</li>
        </ul>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.palette-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 15vh;
}
.palette-dialog {
  background: var(--bg-elevated, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.25);
  width: min(560px, 90vw);
  overflow: hidden;
}
.palette-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}
.palette-icon { opacity: 0.4; flex-shrink: 0; }
.palette-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 15px;
  color: var(--text-primary, #111);
}
.palette-list {
  list-style: none;
  margin: 0;
  padding: 6px;
  max-height: 360px;
  overflow-y: auto;
}
.palette-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  gap: 12px;
}
.palette-item.active { background: var(--accent-color, #4f46e5); color: #fff; }
.palette-label { flex: 1; }
.palette-desc { font-size: 12px; opacity: 0.55; }
.palette-item.active .palette-desc { opacity: 0.75; color: #fff; }
.palette-empty { padding: 20px; text-align: center; color: var(--text-muted, #9ca3af); font-size: 14px; }
</style>
