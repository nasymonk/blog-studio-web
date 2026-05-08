<script setup lang="ts">
import { ref } from 'vue'
import { LockIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'

const { t } = useI18n()
const notify = useNotify()
const currentPassword = ref('')
const newPassword = ref('')
const saving = ref(false)
const error = ref('')

async function changePassword() {
  error.value = ''
  if (newPassword.value.length < 6) { error.value = '密码至少 6 位'; return }
  saving.value = true
  try {
    await api.changePassword(currentPassword.value, newPassword.value)
    notify.success('密码已修改')
    currentPassword.value = ''
    newPassword.value = ''
  } catch (e: any) {
    error.value = e?.message || '修改失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div style="max-width:480px;display:grid;gap:16px">
    <div class="settings-card">
      <h2><LockIcon :size="18" style="vertical-align:middle" /> {{ t.securitySettings }}</h2>
      <p>修改管理员密码。密码至少 6 位。</p>
      <div class="field">
        <label class="field-label">{{ t.currentPassword }}</label>
        <input v-model="currentPassword" class="input" type="password" autocomplete="current-password" />
      </div>
      <div class="field">
        <label class="field-label">{{ t.newPassword }}</label>
        <input v-model="newPassword" class="input" type="password" autocomplete="new-password" />
        <p v-if="error" class="field-error">{{ error }}</p>
      </div>
      <div class="settings-actions">
        <button class="btn btn-primary" :disabled="saving || !currentPassword || !newPassword" @click="changePassword">
          {{ saving ? t.saving : t.changePassword }}
        </button>
      </div>
    </div>
  </div>
</template>
