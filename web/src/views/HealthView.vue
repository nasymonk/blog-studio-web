<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, CheckCircleIcon, XCircleIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const { t } = useI18n()
const notify = useNotify()

const loading = ref(false)
const health = ref<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null>(null)

async function load() {
  loading.value = true
  try {
    health.value = await api.health()
  } catch (e: any) {
    notify.error(e, { onRetry: load })
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="max-w-3xl space-y-4">
    <div class="flex items-center gap-2">
      <h2 class="font-serif text-lg font-semibold m-0">{{ t.health }}</h2>
      <Button variant="ghost" size="icon" class="h-7 w-7" :disabled="loading" :title="t.loading" @click="load">
        <RefreshCwIcon class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
      </Button>
      <Badge v-if="health" :variant="health.status === 'ok' ? 'default' : 'destructive'">
        {{ health.status }}
      </Badge>
    </div>

    <div class="grid gap-3">
      <template v-if="loading">
        <Skeleton v-for="i in 3" :key="i" class="h-[72px] rounded-lg" />
      </template>
      <template v-else-if="health">
        <Card v-for="check in health.checks" :key="check.name">
          <CardContent class="py-3 px-4 space-y-1">
            <div class="flex items-center gap-2">
              <CheckCircleIcon v-if="check.status === 'ok'" class="h-4 w-4 text-ok" />
              <XCircleIcon v-else class="h-4 w-4 text-destructive" />
              <span class="text-sm font-medium">{{ check.name }}</span>
            </div>
            <p class="text-sm text-muted-foreground">{{ check.message }}</p>
            <p v-if="check.technicalDetail" class="text-xs font-mono text-muted-foreground/70">{{ check.technicalDetail }}</p>
            <p v-if="check.suggestion && check.status !== 'ok'" class="text-xs text-warn">建议：{{ check.suggestion }}</p>
          </CardContent>
        </Card>
      </template>
    </div>
  </div>
</template>
