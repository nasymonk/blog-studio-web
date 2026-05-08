<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, CheckCircleIcon, XCircleIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'

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
  <div class="page-content" style="max-width:720px">
    <div class="flex items-center gap-2 mb-4">
      <h2 style="margin:0">{{ t.health }}</h2>
      <button class="btn btn-ghost btn-icon btn-sm" :disabled="loading" :title="t.loading" @click="load">
        <RefreshCwIcon :size="14" :class="{ spin: loading }" />
      </button>
      <span v-if="health" class="badge" :class="health.status === 'ok' ? 'badge-ok' : 'badge-error'">
        {{ health.status }}
      </span>
    </div>

    <div class="health-grid">
      <template v-if="loading">
        <div v-for="i in 3" :key="i" class="skeleton" style="height:72px"></div>
      </template>
      <template v-else-if="health">
        <div v-for="check in health.checks" :key="check.name" class="health-item" :class="check.status">
          <div class="flex items-center gap-2">
            <CheckCircleIcon v-if="check.status === 'ok'" :size="15" style="color:var(--ok)" />
            <XCircleIcon v-else :size="15" style="color:var(--error)" />
            <span class="health-name">{{ check.name }}</span>
          </div>
          <span class="health-detail">{{ check.message }}</span>
          <span v-if="check.technicalDetail" class="health-detail" style="opacity:0.7;font-family:monospace">{{ check.technicalDetail }}</span>
          <span v-if="check.suggestion && check.status !== 'ok'" class="health-suggestion">建议：{{ check.suggestion }}</span>
        </div>
      </template>
    </div>
  </div>
</template>
