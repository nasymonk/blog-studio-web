<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Trash2Icon, UndoIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { TrashItem } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import EmptyState from '@/components/EmptyState.vue'

const { t } = useI18n()
const notify = useNotify()
const items = ref<TrashItem[]>([])
const loading = ref(true)

async function load() {
  loading.value = true
  try { items.value = await api.trash() }
  catch (e: any) { notify.error(e) }
  finally { loading.value = false }
}

async function restoreItem(item: Record<string, any>) {
  return restore(item as TrashItem)
}

async function purgeItem(item: Record<string, any>) {
  return purge(item as TrashItem)
}

async function restore(item: TrashItem) {
  try {
    await api.restoreTrash(item.id)
    notify.success(t.value.restoreOk(item.slug))
    await load()
  } catch (e: any) { notify.error(e) }
}

async function purge(item: TrashItem) {
  if (!confirm(t.value.purgeConfirm(item.slug))) return
  try {
    await api.purgeTrash(item.id)
    notify.success(t.value.purgeOk(item.slug))
    await load()
  } catch (e: any) { notify.error(e) }
}

function relDate(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return t.value.today
  if (days === 1) return t.value.yesterday
  if (days < 30) return t.value.daysAgo(days)
  if (days < 365) return t.value.monthsAgoLong(Math.floor(days / 30))
  return t.value.yearsAgo(Math.floor(days / 365))
}

function fmtSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

onMounted(load)
</script>

<template>
  <div class="max-w-3xl space-y-5">
    <div>
      <h1 class="flex items-center gap-2 font-serif text-lg font-semibold"><Trash2Icon class="h-4 w-4" /> {{ t.trash }}</h1>
      <p class="text-xs text-muted-foreground mt-0.5">{{ loading ? t.loading : `${items.length} ${t.trashSubtitle}` }}</p>
    </div>

    <div v-if="loading" class="stagger space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-[44px] rounded-lg animate-fade-up" />
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      :title="t.trashEmpty"
      :description="t.trashEmptyDesc"
    />

    <div v-else class="space-y-3 stagger">
      <div
        v-for="item in items" :key="item.id"
        class="group flex items-center justify-between gap-3 rounded-lg bg-muted/30 px-4 py-3 hover:bg-muted/50 transition-colors animate-fade-up"
      >
        <div class="flex items-center gap-3 min-w-0">
          <span class="font-mono text-sm truncate">{{ item.slug }}</span>
          <span class="text-xs text-muted-foreground">{{ relDate(String(item.deletedAt)) }}</span>
          <span class="text-xs text-muted-foreground">{{ fmtSize(Number(item.size)) }}</span>
        </div>
        <div class="flex gap-1 shrink-0 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity" @click.stop>
          <Button variant="ghost" size="sm" class="text-xs h-7 text-muted-foreground" @click="restoreItem(item)">
            <UndoIcon class="h-3 w-3 mr-1" /> {{ t.restore }}
          </Button>
          <Button variant="ghost" size="sm" class="text-xs h-7 text-destructive/60 hover:text-destructive" @click="purgeItem(item)">
            <Trash2Icon class="h-3 w-3 mr-1" /> {{ t.purge }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
