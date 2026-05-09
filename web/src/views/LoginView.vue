<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { MoonIcon, SunIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader } from '@/components/ui/card'

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
  <div class="flex min-h-svh items-center justify-center bg-background p-4">
    <Card class="w-full max-w-sm relative">
      <div class="absolute right-3 top-3 flex gap-1">
        <Button variant="ghost" size="icon" class="h-8 w-8" :title="theme === 'dark' ? t.lightMode : t.darkMode" @click="toggle">
          <MoonIcon v-if="theme === 'light'" class="h-4 w-4" />
          <SunIcon v-else class="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="sm" class="h-8 px-2 text-xs" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
          {{ lang === 'zh' ? t.langEn : t.langZh }}
        </Button>
      </div>
      <CardHeader class="items-center text-center pt-8">
        <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-xl bg-accent text-accent-foreground font-serif text-2xl font-bold">博</div>
        <h1 class="font-serif text-xl font-semibold mt-3">{{ t.loginTitle }}</h1>
        <p class="text-sm text-muted-foreground">{{ t.loginDesc }}</p>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4" @submit.prevent="login">
          <div class="grid gap-1.5">
            <Label for="password">{{ t.password }}</Label>
            <Input id="password" v-model="password" type="password" autocomplete="current-password" autofocus />
            <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
          </div>
          <Button type="submit" class="w-full" :disabled="loading || !password">
            <span v-if="loading" class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
            {{ loading ? t.loading : t.login }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
