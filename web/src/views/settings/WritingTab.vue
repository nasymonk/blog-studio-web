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
  saving.value = true
  try {
    const saved = await api.saveConfig(cfg.value)
    store.config = saved; cfg.value = saved; dirty.value = false
    notify.success(t.value.saveSettings)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}
</script>

<template>
  <div v-if="cfg" style="max-width:680px;display:grid;gap:16px">
    <div class="settings-card">
      <h2>{{ t.writingSettings }}</h2>
      <div class="settings-grid">
        <div class="field" style="grid-column:1/-1">
          <label class="field-label">{{ t.blogRoot }}</label>
          <input v-model="cfg.site.blogRoot" class="input" @input="dirty=true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.postSection }}</label>
          <input v-model="cfg.site.postSection" class="input" @input="dirty=true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.buildCommand }}</label>
          <select v-model="cfg.site.buildCommand" class="input" @change="dirty=true">
            <option value="hugo --minify">hugo --minify</option>
            <option value="hugo">hugo</option>
          </select>
        </div>
      </div>
    </div>
    <div class="settings-actions">
      <button class="btn btn-primary" :disabled="saving || !dirty" @click="save">
        <SaveIcon :size="14" />{{ saving ? t.saving : t.saveSettings }}
      </button>
    </div>
  </div>
  <div v-else class="skeleton" style="height:200px"></div>
</template>
