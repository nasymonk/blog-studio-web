<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, CheckCircleIcon, XCircleIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

const { t } = useI18n()
const notify = useNotify()

const loading = ref(false)
const health = ref<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null>(null)

async function load() {
  loading.value = true
  try { health.value = await api.health() }
  catch (e: any) { notify.error(e, { onRetry: load }) }
  finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div class="max-w-2xl space-y-5">
    <div class="flex items-center gap-2">
      <h2 class="font-serif text-lg font-semibold m-0">{{ t.health }}</h2>
      <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" aria-label="Refresh health status" :disabled="loading" @click="load">
        <RefreshCwIcon class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
      </Button>
      <Badge v-if="health" :variant="health.status === 'ok' ? 'default' : 'destructive'" class="text-[10px]" role="status">
        {{ health.status }}
      </Badge>
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

    <div v-else-if="!health" class="text-center py-16 text-muted-foreground animate-fade-up">
      <RefreshCwIcon class="h-8 w-8 mx-auto opacity-20 mb-3" />
      <p class="font-serif">{{ t.noHealthData }}</p>
      <Button variant="outline" class="rounded-full px-5 mt-3" @click="load">
        <RefreshCwIcon class="h-3.5 w-3.5 mr-1.5" />{{ t.refresh }}
      </Button>
    </div>

    <div v-else class="space-y-2 stagger">
      <div
        v-for="check in health.checks"
        :key="check.name"
        class="relative flex items-start gap-3 rounded border border-border/60 bg-card px-4 py-3 animate-fade-up"
      >
        <!-- Status bar -->
        <div class="absolute left-0 top-2 bottom-2 w-0.5 rounded-full" :class="check.status === 'ok' ? 'bg-ok' : 'bg-destructive'" />
        <div class="pl-2 space-y-1 min-w-0">
          <div class="flex items-center gap-2">
            <CheckCircleIcon v-if="check.status === 'ok'" class="h-4 w-4 text-ok shrink-0" />
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
