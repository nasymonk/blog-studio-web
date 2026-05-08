<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { MoonIcon, SunIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
const router = useRouter()
const store = useStore()
const { t, lang, setLang } = useI18n()
const { theme, toggle } = useTheme()

const password = ref('')
const loading = ref(false)
const error = ref('')

async function login() {
  if (!password.value || loading.value) return
  error.value = ''
  loading.value = true
  try {
    const data = await api.login(password.value)
    api.setCSRF(data.csrfToken)
    store.session = { authenticated: true, csrfToken: data.csrfToken }
    router.push('/posts')
  } catch (e: any) {
    error.value = e?.message || t.value.errorLogin
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-toolbar">
        <button class="btn btn-ghost btn-icon btn-sm" :title="theme === 'dark' ? t.lightMode : t.darkMode" @click="toggle">
          <MoonIcon v-if="theme === 'light'" :size="16" />
          <SunIcon v-else :size="16" />
        </button>
        <button class="btn btn-ghost btn-sm" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
          {{ lang === 'zh' ? t.langEn : t.langZh }}
        </button>
      </div>
      <div class="login-logo">博</div>
      <div>
        <h1>{{ t.loginTitle }}</h1>
        <p>{{ t.loginDesc }}</p>
      </div>
      <form style="display:grid;gap:12px" @submit.prevent="login">
        <div class="field">
          <label class="field-label">{{ t.password }}</label>
          <input v-model="password" class="input" type="password" autocomplete="current-password" autofocus />
          <p v-if="error" class="field-error">{{ error }}</p>
        </div>
        <button class="btn btn-primary btn-lg w-full" type="submit" :disabled="loading || !password">
          <span v-if="loading" class="spin" style="display:inline-block;width:14px;height:14px;border:2px solid currentColor;border-top-color:transparent;border-radius:50%"></span>
          {{ loading ? t.loading : t.login }}
        </button>
      </form>
    </div>
  </div>
</template>
