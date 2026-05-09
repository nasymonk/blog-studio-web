<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, XIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { AuditEntry } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetClose } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'

const { t } = useI18n()
const notify = useNotify()
const entries = ref<AuditEntry[]>([])
const loading = ref(false)
const selected = ref<AuditEntry | null>(null)

function fmtTime(iso: string) {
  if (!iso) return ''
  try { return new Date(iso).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
  catch { return iso }
}

async function load() {
  loading.value = true
  try { entries.value = await api.audit(100) }
  catch (e: any) { notify.error(e) }
  finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2">
      <h2 class="font-serif text-lg font-semibold m-0">{{ t.auditLog }}</h2>
      <Button variant="ghost" size="icon" class="h-7 w-7" :disabled="loading" @click="load">
        <RefreshCwIcon class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
      </Button>
    </div>

    <Skeleton v-if="loading" class="h-[300px]" />
    <div v-else-if="entries.length === 0" class="text-center py-16 text-muted-foreground">暂无审计记录</div>

    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>{{ t.time }}</TableHead>
          <TableHead>Slug</TableHead>
          <TableHead>{{ t.op }}</TableHead>
          <TableHead>{{ t.result }}</TableHead>
          <TableHead>Build</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow
          v-for="e in entries"
          :key="e.auditId"
          class="cursor-pointer"
          @click="selected = e"
        >
          <TableCell class="text-sm">{{ fmtTime(e.timestamp) }}</TableCell>
          <TableCell class="max-w-[160px] truncate font-mono text-sm">{{ e.slug }}</TableCell>
          <TableCell class="text-sm">{{ e.operation }}</TableCell>
          <TableCell>
            <Badge :variant="e.result === 'success' ? 'default' : 'destructive'">{{ e.result }}</Badge>
          </TableCell>
          <TableCell>
            <Badge v-if="e.buildResult" :variant="e.buildResult.success ? 'default' : 'destructive'">
              {{ e.buildResult.exitCode }}
            </Badge>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <Sheet :open="!!selected" @update:open="(v: boolean) => { if (!v) selected = null }">
      <SheetContent>
        <SheetHeader>
          <SheetTitle class="font-serif">{{ selected?.slug }} · {{ selected?.operation }}</SheetTitle>
        </SheetHeader>
        <div v-if="selected" class="mt-4 space-y-4">
          <dl class="grid gap-3 text-sm" style="grid-template-columns:120px 1fr">
            <dt class="text-muted-foreground">时间</dt><dd>{{ fmtTime(selected.timestamp) }}</dd>
            <dt class="text-muted-foreground">审计 ID</dt><dd class="font-mono text-xs">{{ selected.auditId }}</dd>
            <dt class="text-muted-foreground">操作</dt><dd>{{ selected.operation }}</dd>
            <dt class="text-muted-foreground">结果</dt><dd><Badge :variant="selected.result === 'success' ? 'default' : 'destructive'">{{ selected.result }}</Badge></dd>
            <dt class="text-muted-foreground">备份 ID</dt><dd>{{ selected.backupId || '—' }}</dd>
            <dt class="text-muted-foreground">错误码</dt><dd>{{ selected.errorCode || '—' }}</dd>
            <dt class="text-muted-foreground">错误信息</dt><dd>{{ selected.errorBrief || '—' }}</dd>
          </dl>
          <div v-if="selected.buildResult">
            <p class="text-sm font-medium mb-1">构建日志</p>
            <pre class="text-xs font-mono bg-muted p-3 rounded-lg max-h-[300px] overflow-auto whitespace-pre-wrap">{{ selected.buildResult.stderr || selected.buildResult.stdout || '（空）' }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
