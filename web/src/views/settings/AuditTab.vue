<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { RefreshCwIcon, SearchIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { AuditEntry } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import DataTable from '@/components/DataTable.vue'
import type { Column } from '@/components/DataTable.vue'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'

const { t } = useI18n()
const notify = useNotify()
const entries = ref<AuditEntry[]>([])
const loading = ref(false)
const selected = ref<AuditEntry | null>(null)

const searchQuery = ref('')
const operationFilter = ref('')
const operations = ['', 'publish', 'save', 'delete', 'rollback']

const auditColumns: Column[] = [
  { key: 'timestamp', label: t.value.time, sortable: true },
  { key: 'slug', label: 'Slug', sortable: true },
  { key: 'operation', label: t.value.op, sortable: true },
  { key: 'result', label: t.value.result },
  { key: 'build', label: 'Build' },
]

function handleRowClick(row: Record<string, any>) {
  selected.value = row as AuditEntry
}

function opLabel(op: string) {
  if (!op) return t.value.opAll
  const map: Record<string, () => string> = {
    publish: () => t.value.opPublish,
    save: () => t.value.opSave,
    delete: () => t.value.opDelete,
    rollback: () => t.value.opRollback,
  }
  return map[op]?.() ?? op
}

function fmtTime(iso: string) {
  if (!iso) return ''
  try { return new Date(iso).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
  catch { return iso }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function load() {
  loading.value = true
  try {
    entries.value = await api.audit(100, operationFilter.value || undefined, searchQuery.value || undefined)
  }
  catch (e: any) { notify.error(e) }
  finally { loading.value = false }
}

function debouncedLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(load, 300)
}

watch(operationFilter, load)
watch(searchQuery, debouncedLoad)

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

    <!-- Search and filter controls -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <div class="relative flex-1">
        <SearchIcon class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          :placeholder="t.searchAudit"
          class="pl-8 h-8 text-sm"
        />
      </div>
      <div class="flex items-center gap-1 flex-wrap">
        <Button
          v-for="op in operations"
          :key="op"
          :variant="operationFilter === op ? 'default' : 'outline'"
          size="sm"
          class="h-7 text-xs rounded-full px-3"
          @click="operationFilter = op"
        >
          {{ opLabel(op) }}
        </Button>
      </div>
    </div>

    <Skeleton v-if="loading" class="h-[300px]" />
    <div v-else-if="entries.length === 0" class="text-center py-16 text-muted-foreground text-sm">{{ t.noAuditRecords }}</div>

    <DataTable
      v-else
      :columns="auditColumns"
      :data="entries"
      :empty-message="t.noAuditRecords"
      @row-click="handleRowClick"
    >
      <template #timestamp="{ value }">
        <span class="font-mono text-[11px] text-muted-foreground">{{ fmtTime(String(value)) }}</span>
      </template>
      <template #slug="{ value }">
        <span class="max-w-[160px] truncate font-mono text-sm">{{ value }}</span>
      </template>
      <template #operation="{ value }">
        <span class="text-sm">{{ value }}</span>
      </template>
      <template #result="{ value }">
        <Badge :variant="value === 'success' ? 'default' : 'destructive'" class="text-[10px]">{{ value }}</Badge>
      </template>
      <template #build="{ row }">
        <Badge v-if="row.buildResult" :variant="(row.buildResult as any).success ? 'default' : 'destructive'" class="text-[10px]">
          {{ (row.buildResult as any).exitCode }}
        </Badge>
      </template>
    </DataTable>

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
