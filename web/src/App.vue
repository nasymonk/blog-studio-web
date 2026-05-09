<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { Toaster } from 'vue-sonner'
import {
  FileTextIcon, SettingsIcon, HomeIcon, ClockIcon, HeartPulseIcon,
  LogOutIcon, MoonIcon, SunIcon, Globe2Icon, RefreshCwIcon, Trash2Icon
} from 'lucide-vue-next'
import { useEventListener } from '@vueuse/core'
import { api } from '@/services/api'
import { createStore, provideStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import CommandPalette from '@/components/CommandPalette.vue'
import router from '@/router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

const store = createStore()
provideStore(store)

const route = useRoute()
const { t, lang, setLang } = useI18n()
const { theme, toggle: toggleTheme } = useTheme()
const isDark = computed(() => theme.value === 'dark')
const notify = useNotify()

const isAuthed = computed(() => store.session.authenticated)
const isLoginPage = computed(() => route.name === 'login')

const paletteOpen = ref(false)
useEventListener('keydown', (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    if (isAuthed.value) paletteOpen.value = true
  }
})

const pageTitle = computed(() => {
  const name = route.name as string
  if (name === 'posts') return t.value.posts
  if (name === 'editor') return store.editor.dirty ? `● ${t.value.edit}` : t.value.edit
  if (name === 'home') return t.value.home
  if (name === 'now') return t.value.now
  if (name === 'settings' || name?.startsWith('settings-')) return t.value.settings
  if (name === 'health') return t.value.health
  return 'Blog Studio'
})

const editorDirty = computed(() => route.name === 'editor' && store.editor.dirty)
const editorSaving = computed(() => route.name === 'editor' && store.editor.saving)

async function logout() {
  try {
    await api.logout()
  } catch {
    // ignore errors on logout
  }
  store.session = { authenticated: false, csrfToken: '' }
  router.push('/login')
}

async function syncPosts() {
  store.postsLoading = true
  try {
    store.posts = await api.posts()
  } catch (e: any) {
    notify.error(e)
  } finally {
    store.postsLoading = false
  }
}

router.beforeEach((to, from) => {
  if (to.name !== 'login' && !store.session.authenticated) {
    return { name: 'login' }
  }
  if (from.name === 'editor' && to.name !== 'editor' && store.editor.dirty) {
    if (!window.confirm('文章有未保存的更改，确认离开？')) return false
    store.editor.dirty = false
  }
})

onMounted(async () => {
  try {
    const session = await api.session()
    store.session = { authenticated: session.authenticated, csrfToken: session.csrfToken ?? '' }
    if (session.csrfToken) api.setCSRF(session.csrfToken)
    if (session.authenticated) {
      const [cfg, posts] = await Promise.all([api.config(), api.posts()])
      store.config = cfg
      store.posts = posts
      try { store.health = await api.health() } catch { /* non-critical */ }
    }
  } catch (e: any) {
    notify.error(e)
  }
})
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <Toaster position="top-right" rich-colors :expand="true" />
    <CommandPalette v-model:open="paletteOpen" />

    <!-- Skip navigation link for keyboard users -->
    <a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:z-[9999] focus:top-2 focus:left-2 focus:px-3 focus:py-2 focus:bg-background focus:border focus:border-border focus:rounded-md">跳到主内容</a>

    <div v-if="isLoginPage || !isAuthed" class="min-h-svh">
      <RouterView />
    </div>

    <div v-else class="flex min-h-svh">
      <!-- Sidebar -->
      <aside class="w-[220px] shrink-0 border-r border-border bg-sidebar flex flex-col" aria-label="主导航">
        <div class="flex items-center gap-2.5 px-4 py-4" aria-label="Blog Studio">
          <svg class="h-8 w-8 shrink-0" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <rect width="32" height="32" rx="8" fill="var(--accent)" />
            <text x="16" y="22" text-anchor="middle" fill="white" font-size="18" font-weight="700" font-family="serif">B</text>
          </svg>
          <span class="font-serif font-semibold text-sm truncate">{{ store.config?.site.name || 'Blog Studio' }}</span>
        </div>

        <ScrollArea class="flex-1">
          <nav class="px-2 space-y-0.5" role="navigation" aria-label="主菜单">
            <Tooltip v-for="item in [
              { to: '/posts', icon: FileTextIcon, label: t.posts, count: store.posts.length },
              { to: '/home', icon: HomeIcon, label: t.home },
              { to: '/now', icon: ClockIcon, label: t.now },
              { to: '/settings', icon: SettingsIcon, label: t.settings },
              { to: '/health', icon: HeartPulseIcon, label: t.health, dot: store.health && store.health.status !== 'ok' },
              { to: '/trash', icon: Trash2Icon, label: '回收站' },
            ]" :key="item.to">
              <TooltipTrigger as-child>
                <RouterLink
                  :to="item.to"
                  class="flex items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm text-sidebar-foreground hover:bg-sidebar-accent transition-colors"
                  active-class="bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" aria-hidden="true" />
                  <span class="flex-1">{{ item.label }}</span>
                  <Badge v-if="item.count" variant="secondary" class="h-5 min-w-[20px] justify-center px-1.5 text-[10px]">
                    {{ item.count }}
                  </Badge>
                  <span v-if="item.dot" class="h-2 w-2 rounded-full bg-destructive" role="status" aria-label="服务异常" />
                </RouterLink>
              </TooltipTrigger>
              <TooltipContent side="right" :side-offset="8">{{ item.label }}</TooltipContent>
            </Tooltip>
          </nav>
        </ScrollArea>

        <div class="px-2 py-2 border-t border-sidebar-border space-y-0.5">
          <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5" :disabled="store.postsLoading" :aria-label="t.sync" @click="syncPosts">
            <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" aria-hidden="true" />
            <span>{{ t.sync }}</span>
          </Button>
          <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5" :aria-label="t.logout" @click="logout">
            <LogOutIcon class="h-4 w-4" aria-hidden="true" />
            <span>{{ t.logout }}</span>
          </Button>
        </div>
      </aside>

      <!-- Main stage -->
      <div class="flex-1 flex flex-col min-w-0">
        <header class="flex items-center h-12 px-4 border-b border-border shrink-0" role="banner">
          <h1 class="font-serif font-semibold text-base">{{ pageTitle }}</h1>
          <div class="flex items-center gap-2 ml-auto">
            <Badge v-if="editorSaving" variant="secondary" class="text-xs animate-pulse">{{ t.saving }}…</Badge>
            <Badge v-else-if="editorDirty" variant="destructive" class="text-xs">● {{ t.unsaved }}</Badge>
            <Badge v-else-if="store.editor.savedAt" variant="outline" class="text-xs text-ok">{{ t.saved }}</Badge>

            <Button variant="ghost" size="sm" class="h-7 px-2 text-xs gap-1" :aria-label="`切换语言为 ${lang === 'zh' ? 'English' : '中文'}`" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
              <Globe2Icon class="h-3.5 w-3.5" aria-hidden="true" />{{ lang === 'zh' ? 'EN' : '中文' }}
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7" :aria-label="isDark ? t.lightMode : t.darkMode" :title="t.theme" @click="toggleTheme">
              <MoonIcon v-if="!isDark" class="h-3.5 w-3.5" aria-hidden="true" />
              <SunIcon v-else class="h-3.5 w-3.5" aria-hidden="true" />
            </Button>
          </div>
        </header>

        <div v-if="store.banner" class="flex items-center gap-2 px-4 py-2 bg-warn-bg text-warn text-sm" role="alert">
          <span class="flex-1">{{ store.banner.message }}</span>
          <Button variant="ghost" size="icon" class="h-6 w-6" aria-label="关闭提示" @click="store.banner = null">
            <span class="text-xs">✕</span>
          </Button>
        </div>

        <main id="main-content" class="flex-1 overflow-auto p-6" tabindex="-1">
          <RouterView />
        </main>
      </div>
    </div>
  </TooltipProvider>
</template>
