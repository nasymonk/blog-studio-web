<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Trash2Icon, UndoIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { TrashItem } from '@/services/api'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

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
  <div class="max-w-3xl space-y-4">
    <div>
      <h1 class="flex items-center gap-2 font-serif text-lg font-semibold"><Trash2Icon class="h-4 w-4" /> 回收站</h1>
      <p class="text-sm text-muted-foreground">已删除的文章在 30 天后自动清空</p>
    </div>

    <div v-if="loading" class="text-center py-16 text-muted-foreground">加载中…</div>

    <div v-else-if="items.length === 0" class="text-center py-16 text-muted-foreground">
      <Trash2Icon class="h-8 w-8 mx-auto opacity-30 mb-3" />
      <p>回收站为空</p>
    </div>

    <Card v-else>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>文章</TableHead>
              <TableHead>删除时间</TableHead>
              <TableHead>大小</TableHead>
              <TableHead class="text-right"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in items" :key="item.id">
              <TableCell><span class="font-mono text-sm">{{ item.slug }}</span></TableCell>
              <TableCell :title="item.deletedAt" class="text-sm text-muted-foreground">{{ relDate(item.deletedAt) }}</TableCell>
              <TableCell class="text-sm">{{ fmtSize(item.size) }}</TableCell>
              <TableCell class="text-right">
                <div class="flex gap-1 justify-end">
                  <Button variant="ghost" size="sm" @click="restore(item)">
                    <UndoIcon class="h-3.5 w-3.5 mr-1" /> 还原
                  </Button>
                  <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="purge(item)">
                    <Trash2Icon class="h-3.5 w-3.5 mr-1" /> 永久删除
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
