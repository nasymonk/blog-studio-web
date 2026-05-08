<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { SaveIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { Config } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'

const store = useStore()
const { t } = useI18n()
const notify = useNotify()

const cfg = ref<Config | null>(null)
const dirty = ref(false)
const saving = ref(false)

onMounted(async () => {
  try {
    cfg.value = store.config ?? await api.config()
    store.config = cfg.value
  } catch (e: any) { notify.error(e) }
})

async function save() {
  if (!cfg.value) return
  const err = validate()
  if (err) { notify.warn(err); return }
  saving.value = true
  try {
    const saved = await api.saveConfig(cfg.value)
    store.config = saved
    cfg.value = saved
    dirty.value = false
    notify.success(t.value.saveSettings)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}

function validate(): string {
  if (!cfg.value) return ''
  if (!cfg.value.basePath.startsWith('/')) return 'Base Path 必须以 / 开头'
  if (cfg.value.preview.ttlMinutes < 10) return '预览有效期不能少于 10 分钟'
  return ''
}
</script>

<template>
  <div v-if="cfg" style="max-width:680px;display:grid;gap:16px">
    <div class="settings-card">
      <h2>{{ t.generalSettings }}</h2>
      <div class="settings-grid">
        <div class="field">
          <label class="field-label">{{ t.siteId }}</label>
          <input v-model="cfg.site.id" class="input" @input="dirty=true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.siteName }}</label>
          <input v-model="cfg.site.name" class="input" @input="dirty=true" />
        </div>
        <div class="field" style="grid-column:1/-1">
          <label class="field-label">{{ t.basePath }}</label>
          <input v-model="cfg.basePath" class="input" @input="dirty=true" />
          <p class="field-hint">例：/studio</p>
        </div>
      </div>
    </div>

    <div class="settings-card">
      <h2>预览</h2>
      <div class="settings-grid">
        <div class="field">
          <label class="field-label">{{ t.previewTTL }}</label>
          <input v-model.number="cfg.preview.ttlMinutes" class="input" type="number" min="10" @input="dirty=true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.maxUploadMb }}</label>
          <input
class="input" type="number" min="1" :value="Math.round(cfg.maxUploadBytes / 1024 / 1024)"
            @input="e => { cfg!.maxUploadBytes = Number((e.target as HTMLInputElement).value) * 1024 * 1024; dirty=true }" />
        </div>
      </div>
    </div>

    <div class="settings-actions">
      <button class="btn btn-primary" :disabled="saving || !dirty" @click="save">
        <SaveIcon :size="14" />{{ saving ? t.saving : t.saveSettings }}
      </button>
    </div>
  </div>
  <div v-else class="skeleton" style="height:300px"></div>
</template>
