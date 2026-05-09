<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { MoonIcon, SunIcon, ArrowRightIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import AnimatedCharacters from '@/components/AnimatedCharacters.vue'

const router = useRouter()
const store = useStore()
const { t, lang, setLang } = useI18n()
const { theme, toggle } = useTheme()

const password = ref('')
const loading = ref(false)
const error = ref('')
const btnHover = ref(false)
const isTyping = ref(false)
let typingTimeout: ReturnType<typeof setTimeout> | null = null

function onPasswordInput() {
  isTyping.value = true
  if (typingTimeout) clearTimeout(typingTimeout)
  typingTimeout = setTimeout(() => { isTyping.value = false }, 600)
}

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
  <div class="grid min-h-svh lg:grid-cols-2">
    <!-- Left panel: brand / personality -->
    <div class="login-left-panel relative hidden lg:flex flex-col items-center justify-center overflow-hidden">
      <!-- Decorative blurred circles -->
      <div class="absolute top-1/4 -left-20 h-80 w-80 rounded-full bg-accent/10 blur-3xl" />
      <div class="absolute bottom-1/4 -right-16 h-64 w-64 rounded-full bg-accent/8 blur-3xl" />
      <!-- Grid texture overlay -->
      <div class="absolute inset-0 opacity-[0.03]"
        style="background-image: linear-gradient(rgba(30,28,24,1) 1px, transparent 1px), linear-gradient(90deg, rgba(30,28,24,1) 1px, transparent 1px); background-size: 40px 40px;" />

      <!-- Center content -->
      <div class="relative z-10 flex flex-col items-center justify-center">
        <AnimatedCharacters
          :is-typing="isTyping"
          :show-password="false"
          :password-length="password.length"
        />
        <p class="font-deco text-lg tracking-widest mt-4" style="color: #8a7e6e;">
          写作是一种修行
        </p>
      </div>

      <!-- Bottom brand -->
      <div class="absolute bottom-8 text-xs tracking-[0.2em] uppercase" style="color: rgba(107,100,87,0.3);">
        Blog Studio
      </div>
    </div>

    <!-- Right panel: form -->
    <div class="relative flex flex-col items-center justify-center p-6 bg-background">
      <!-- Mobile brand (hidden on lg+) -->
      <div class="lg:hidden mb-8 text-center">
        <div class="font-display text-6xl text-accent animate-ink-breathe select-none">博</div>
        <p class="font-deco text-sm mt-2 text-muted-foreground tracking-widest">写作是一种修行</p>
      </div>

      <div class="w-full max-w-[400px] space-y-8">
        <div class="space-y-2">
          <h1 class="font-serif text-2xl font-semibold tracking-tight">{{ t.loginTitle }}</h1>
          <p class="text-sm text-muted-foreground">{{ t.loginDesc }}</p>
        </div>

        <form class="space-y-6" @submit.prevent="login">
          <div class="space-y-2">
            <Label for="password" class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.password }}</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              autofocus
              class="h-12 border-0 border-b border-border rounded-none bg-transparent px-0 focus-visible:ring-0 focus-visible:border-accent transition-colors"
              @input="onPasswordInput"
            />
            <p v-if="error" class="text-xs text-destructive mt-1">{{ error }}</p>
          </div>

          <!-- Interactive hover button -->
          <button
            type="submit"
            class="relative w-full h-12 rounded-full bg-accent text-accent-foreground font-medium text-sm overflow-hidden transition-all duration-300 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="loading || !password"
            @mouseenter="btnHover = true"
            @mouseleave="btnHover = false"
          >
            <span class="absolute inset-0 flex items-center justify-center gap-2 transition-all duration-300"
              :class="btnHover ? 'translate-x-4 opacity-0' : 'translate-x-0 opacity-100'">
              <span v-if="loading" class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              {{ loading ? t.loading : t.login }}
            </span>
            <span class="absolute inset-0 flex items-center justify-center gap-2 transition-all duration-300"
              :class="btnHover ? 'translate-x-0 opacity-100' : '-translate-x-4 opacity-0'">
              {{ loading ? t.loading : t.login }}
              <ArrowRightIcon class="h-4 w-4" />
            </span>
          </button>
        </form>
      </div>

      <!-- Bottom-right controls -->
      <div class="absolute bottom-6 right-6 flex items-center gap-2">
        <button
          class="h-8 w-8 rounded-full flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          :title="theme === 'dark' ? t.lightMode : t.darkMode"
          @click="toggle"
        >
          <MoonIcon v-if="theme === 'light'" class="h-4 w-4" />
          <SunIcon v-else class="h-4 w-4" />
        </button>
        <button
          class="h-8 px-3 rounded-full text-xs text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          @click="setLang(lang === 'zh' ? 'en' : 'zh')"
        >
          {{ lang === 'zh' ? t.langEn : t.langZh }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-left-panel {
  background: linear-gradient(135deg, #f5f0e6 0%, #ece5d5 50%, #e0d8c8 100%);
}
:global(.dark) .login-left-panel {
  background: linear-gradient(135deg, #1e2023 0%, #24262a 50%, #1a1b1e 100%);
}
:global(.dark) .login-left-panel .animate-ink-breathe {
  color: var(--accent);
  filter: drop-shadow(0 4px 24px rgba(157,194,214,0.15));
}
</style>
