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
      <h2>{{ t.wechatSettings }}</h2>
      <p>微信公众号草稿发布需要在服务器环境变量中配置 WECHAT_APP_ID / WECHAT_APP_SECRET。</p>
      <div class="field">
        <label class="check-field">
          <input v-model="cfg.wechat.enabled" type="checkbox" @change="dirty=true" />
          <span class="field-label" style="margin:0">{{ t.wechatEnabled }}</span>
        </label>
      </div>
      <div v-if="cfg.wechat.enabled" class="field">
        <label class="field-label">{{ t.wechatAppId }}</label>
        <input v-model="cfg.wechat.appId" class="input" placeholder="wx…" @input="dirty=true" />
        <p class="field-hint">AppSecret 通过环境变量配置，不在此显示。</p>
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
