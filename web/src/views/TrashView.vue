<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Trash2Icon, UndoIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { TrashItem } from '@/services/api'
import { useNotify } from '@/composables/useNotify'

const notify = useNotify()
const items = ref<TrashItem[]>([])
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    items.value = await api.trash()
  } catch (e: any) {
    notify.error(e)
  } finally {
    loading.value = false
  }
}

async function restore(item: TrashItem) {
  try {
    await api.restoreTrash(item.id)
    notify.success(`已还原《${item.slug}》`)
    await load()
  } catch (e: any) {
    notify.error(e)
  }
}

async function purge(item: TrashItem) {
  if (!confirm(`永久删除《${item.slug}》？此操作不可撤销。`)) return
  try {
    await api.purgeTrash(item.id)
    notify.success(`已永久删除《${item.slug}》`)
    await load()
  } catch (e: any) {
    notify.error(e)
  }
}

function relDate(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 30) return `${days} 天前`
  if (days < 365) return `${Math.floor(days / 30)} 个月前`
  return `${Math.floor(days / 365)} 年前`
}

function fmtSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

onMounted(load)
</script>

<template>
  <div class="view-root">
    <div class="view-header">
      <h1 class="view-title"><Trash2Icon :size="18" /> 回收站</h1>
      <p class="view-subtitle">已删除的文章在 30 天后自动清空</p>
    </div>

    <div v-if="loading" class="empty-state">加载中…</div>

    <div v-else-if="items.length === 0" class="empty-state">
      <Trash2Icon :size="32" style="opacity:0.3;margin-bottom:12px" />
      <p>回收站为空</p>
    </div>

    <table v-else class="trash-table">
      <thead>
        <tr>
          <th>文章</th>
          <th>删除时间</th>
          <th>大小</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <td>
            <span class="trash-slug">{{ item.slug }}</span>
          </td>
          <td :title="item.deletedAt">{{ relDate(item.deletedAt) }}</td>
          <td>{{ fmtSize(item.size) }}</td>
          <td class="trash-actions">
            <button class="btn btn-ghost btn-sm" @click="restore(item)">
              <UndoIcon :size="13" /> 还原
            </button>
            <button class="btn btn-ghost btn-sm btn-danger-text" @click="purge(item)">
              <Trash2Icon :size="13" /> 永久删除
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.view-root { padding: 32px; max-width: 900px; }
.view-header { margin-bottom: 24px; }
.view-title { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 700; margin: 0 0 4px; }
.view-subtitle { color: var(--text-muted, #9ca3af); font-size: 13px; margin: 0; }
.empty-state { text-align: center; padding: 60px 0; color: var(--text-muted, #9ca3af); }
.trash-table { width: 100%; border-collapse: collapse; font-size: 14px; }
.trash-table th { text-align: left; padding: 8px 12px; font-size: 12px; font-weight: 600; color: var(--text-muted, #9ca3af); border-bottom: 1px solid var(--border-color, #e5e7eb); }
.trash-table td { padding: 12px; border-bottom: 1px solid var(--border-light, #f3f4f6); }
.trash-slug { font-family: monospace; font-size: 13px; }
.trash-actions { display: flex; gap: 6px; justify-content: flex-end; }
</style>
