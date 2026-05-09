<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { AuditEntry } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
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
      <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :disabled="loading" @click="load">
        <RefreshCwIcon class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
      </Button>
    </div>

    <Skeleton v-if="loading" class="h-[300px]" />
    <div v-else-if="entries.length === 0" class="text-center py-16 text-muted-foreground text-sm">{{ t.noAuditRecords }}</div>

    <div v-else class="rounded border border-border/60 overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow class="hover:bg-transparent">
            <TableHead class="text-[10px] uppercase tracking-wider text-muted-foreground/60">{{ t.time }}</TableHead>
            <TableHead class="text-[10px] uppercase tracking-wider text-muted-foreground/60">Slug</TableHead>
            <TableHead class="text-[10px] uppercase tracking-wider text-muted-foreground/60">{{ t.op }}</TableHead>
            <TableHead class="text-[10px] uppercase tracking-wider text-muted-foreground/60">{{ t.result }}</TableHead>
            <TableHead class="text-[10px] uppercase tracking-wider text-muted-foreground/60">Build</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="e in entries"
            :key="e.auditId"
            class="cursor-pointer hover:bg-muted/30"
            @click="selected = e"
          >
            <TableCell class="font-mono text-[11px] text-muted-foreground">{{ fmtTime(e.timestamp) }}</TableCell>
            <TableCell class="max-w-[160px] truncate font-mono text-sm">{{ e.slug }}</TableCell>
            <TableCell class="text-sm">{{ e.operation }}</TableCell>
            <TableCell>
              <Badge :variant="e.result === 'success' ? 'default' : 'destructive'" class="text-[10px]">{{ e.result }}</Badge>
            </TableCell>
            <TableCell>
              <Badge v-if="e.buildResult" :variant="e.buildResult.success ? 'default' : 'destructive'" class="text-[10px]">
                {{ e.buildResult.exitCode }}
              </Badge>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <Sheet :open="!!selected" @update:open="(v: boolean) => { if (!v) selected = null }">
      <SheetContent>
        <SheetHeader>
          <SheetTitle class="font-serif">{{ selected?.slug }} · {{ selected?.operation }}</SheetTitle>
        </SheetHeader>
        <div v-if="selected" class="mt-4 space-y-4">
          <dl class="grid gap-3 text-sm" style="grid-template-columns:100px 1fr">
            <dt class="text-muted-foreground text-xs">{{ t.time }}</dt><dd class="font-mono text-[11px]">{{ fmtTime(selected.timestamp) }}</dd>
            <dt class="text-muted-foreground text-xs">{{ t.auditId }}</dt><dd class="font-mono text-[11px]">{{ selected.auditId }}</dd>
            <dt class="text-muted-foreground text-xs">{{ t.op }}</dt><dd>{{ selected.operation }}</dd>
            <dt class="text-muted-foreground text-xs">{{ t.result }}</dt><dd><Badge :variant="selected.result === 'success' ? 'default' : 'destructive'" class="text-[10px]">{{ selected.result }}</Badge></dd>
            <dt class="text-muted-foreground text-xs">{{ t.backupId }}</dt><dd>{{ selected.backupId || '—' }}</dd>
            <dt class="text-muted-foreground text-xs">{{ t.errorCode }}</dt><dd>{{ selected.errorCode || '—' }}</dd>
            <dt class="text-muted-foreground text-xs">{{ t.errorMessage }}</dt><dd>{{ selected.errorBrief || '—' }}</dd>
          </dl>
          <div v-if="selected.buildResult">
            <p class="text-xs font-medium mb-1.5">{{ t.buildLog }}</p>
            <pre class="text-[11px] font-mono bg-muted/50 p-3 rounded max-h-[300px] overflow-auto whitespace-pre-wrap border border-border/40">{{ selected.buildResult.stderr || selected.buildResult.stdout || t.emptyLog }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
