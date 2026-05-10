<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, CheckCircleIcon, XCircleIcon, TriangleAlertIcon, ActivityIcon, ClockIcon, EyeIcon, AlertCircleIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import EmptyState from '@/components/EmptyState.vue'

const { t } = useI18n()
const notify = useNotify()

const loading = ref(false)
const health = ref<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null>(null)

interface MetricsData {
  totalRequests: number
  errorRate: number
  avgBuildDuration: number
  activePreviews: number
}

const metricsLoading = ref(false)
const metrics = ref<MetricsData | null>(null)

function parsePrometheus(text: string): MetricsData {
  const lines = text.split('\n')
  let totalRequests = 0
  let errorRequests = 0
  let buildDurationSum = 0
  let buildDurationCount = 0
  let activePreviews = 0

  for (const line of lines) {
    // Skip comments and empty lines
    if (!line || line.startsWith('#')) continue

    // Parse: metric_name{labels} value
    const match = line.match(/^([^\s{]+)(?:\{([^}]*)\})?\s+(.+)$/)
    if (!match) continue

    const [, name, , valueStr] = match
    const value = parseFloat(valueStr)

    if (name === 'http_requests_total') {
      totalRequests += value
      // Check if status label indicates error (4xx or 5xx)
      const statusMatch = line.match(/status="(\d+)"/)
      if (statusMatch && parseInt(statusMatch[1]) >= 400) {
        errorRequests += value
      }
    } else if (name === 'hugo_build_duration_seconds_sum') {
      buildDurationSum = value
    } else if (name === 'hugo_build_duration_seconds_count') {
      buildDurationCount = value
    } else if (name === 'preview_active') {
      activePreviews = value
    }
  }

  return {
    totalRequests,
    errorRate: totalRequests > 0 ? (errorRequests / totalRequests) * 100 : 0,
    avgBuildDuration: buildDurationCount > 0 ? buildDurationSum / buildDurationCount : 0,
    activePreviews,
  }
}

async function loadMetrics() {
  metricsLoading.value = true
  try {
    const raw = await api.metrics()
    metrics.value = parsePrometheus(raw)
  } catch {
    // Metrics endpoint may not be available; silently ignore
    metrics.value = null
  } finally {
    metricsLoading.value = false
  }
}

async function load() {
  loading.value = true
  try { health.value = await api.health() }
  catch (e: any) { notify.error(e, { onRetry: load }) }
  finally { loading.value = false }
}

function loadAll() {
  load()
  loadMetrics()
}

onMounted(loadAll)
</script>

<template>
  <div class="max-w-3xl space-y-5">
    <div class="flex items-center gap-2">
      <h2 class="font-serif text-lg font-semibold m-0">{{ t.health }}</h2>
      <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" aria-label="Refresh health status" :disabled="loading || metricsLoading" @click="loadAll">
        <RefreshCwIcon class="h-3.5 w-3.5" :class="{ 'animate-spin': loading || metricsLoading }" />
      </Button>
      <Badge v-if="health" :variant="health.status === 'ok' ? 'default' : health.status === 'warn' ? 'outline' : 'destructive'" class="text-[10px]" role="status">
        {{ health.status }}
      </Badge>
    </div>

    <!-- Metrics Dashboard -->
    <div v-if="metrics" class="grid grid-cols-2 gap-3">
      <div class="rounded border border-border/60 bg-card p-4 animate-fade-up">
        <div class="flex items-center gap-2 text-muted-foreground mb-1">
          <ActivityIcon class="h-3.5 w-3.5" />
          <span class="text-xs">{{ t.totalRequests }}</span>
        </div>
        <p class="text-2xl font-semibold tabular-nums">{{ metrics.totalRequests.toLocaleString() }}</p>
      </div>
      <div class="rounded border border-border/60 bg-card p-4 animate-fade-up">
        <div class="flex items-center gap-2 text-muted-foreground mb-1">
          <AlertCircleIcon class="h-3.5 w-3.5" />
          <span class="text-xs">{{ t.errorRate }}</span>
        </div>
        <p class="text-2xl font-semibold tabular-nums" :class="metrics.errorRate > 5 ? 'text-destructive' : ''">{{ metrics.errorRate.toFixed(1) }}%</p>
      </div>
      <div class="rounded border border-border/60 bg-card p-4 animate-fade-up">
        <div class="flex items-center gap-2 text-muted-foreground mb-1">
          <ClockIcon class="h-3.5 w-3.5" />
          <span class="text-xs">{{ t.avgBuildDuration }}</span>
        </div>
        <p class="text-2xl font-semibold tabular-nums">{{ metrics.avgBuildDuration > 0 ? (metrics.avgBuildDuration * 1000).toFixed(0) + 'ms' : '—' }}</p>
      </div>
      <div class="rounded border border-border/60 bg-card p-4 animate-fade-up">
        <div class="flex items-center gap-2 text-muted-foreground mb-1">
          <EyeIcon class="h-3.5 w-3.5" />
          <span class="text-xs">{{ t.activePreviews }}</span>
        </div>
        <p class="text-2xl font-semibold tabular-nums">{{ metrics.activePreviews }}</p>
      </div>
    </div>
    <div v-else-if="metricsLoading" class="grid grid-cols-2 gap-3">
      <div v-for="i in 4" :key="i" class="rounded border border-border/60 bg-card p-4">
        <Skeleton class="h-3 w-20 mb-2" />
        <Skeleton class="h-7 w-16" />
      </div>
    </div>

    <div v-if="loading" class="space-y-2">
      <div v-for="i in 3" :key="i" class="relative flex items-start gap-3 rounded border border-border/60 bg-card px-4 py-3 animate-fade-up">
        <div class="absolute left-0 top-2 bottom-2 w-0.5 rounded-full bg-primary/10" />
        <div class="pl-2 space-y-2 flex-1">
          <Skeleton class="h-4 w-28" />
          <Skeleton class="h-3 w-full max-w-xs" />
        </div>
      </div>
    </div>

    <EmptyState
      v-else-if="!health"
      :title="t.noHealthData"
      :description="t.healthLoadFailed"
      :action-label="t.refresh"
      @action="load"
    />

    <div v-else class="space-y-2 stagger">
      <div
        v-for="check in health.checks"
        :key="check.name"
        class="relative flex items-start gap-3 rounded border border-border/60 bg-card px-4 py-3 animate-fade-up"
      >
        <!-- Status bar -->
        <div class="absolute left-0 top-2 bottom-2 w-0.5 rounded-full" :class="check.status === 'ok' ? 'bg-ok' : check.status === 'warn' ? 'bg-yellow-500' : 'bg-destructive'" />
        <div class="pl-2 space-y-1 min-w-0">
          <div class="flex items-center gap-2">
            <CheckCircleIcon v-if="check.status === 'ok'" class="h-4 w-4 text-ok shrink-0" />
            <TriangleAlertIcon v-else-if="check.status === 'warn'" class="h-4 w-4 text-yellow-500 shrink-0" />
            <XCircleIcon v-else class="h-4 w-4 text-destructive shrink-0" />
            <span class="text-sm font-medium">{{ check.name }}</span>
          </div>
          <p class="text-sm text-muted-foreground">{{ check.message }}</p>
          <p v-if="check.technicalDetail" class="text-[11px] font-mono text-muted-foreground/60">{{ check.technicalDetail }}</p>
          <p v-if="check.suggestion && check.status !== 'ok'" class="text-xs text-warn">{{ t.suggestion }}{{ check.suggestion }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
