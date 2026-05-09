<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { Toaster, toast } from 'vue-sonner'
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

watch(() => store.banner, (b) => {
  if (!b) return
  const fn = b.type === 'error' ? toast.error : b.type === 'ok' ? toast.success : toast.warning
  fn(b.message, { duration: b.type === 'error' ? Infinity : 6000 })
  store.banner = null
})

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

const contentNav = [
  { to: '/posts', icon: FileTextIcon, label: () => t.value.posts, count: () => store.posts.length },
  { to: '/home', icon: HomeIcon, label: () => t.value.home },
  { to: '/now', icon: ClockIcon, label: () => t.value.now },
]

const systemNav = [
  { to: '/settings', icon: SettingsIcon, label: () => t.value.settings },
  { to: '/health', icon: HeartPulseIcon, label: () => t.value.health, dot: () => store.health && store.health.status !== 'ok' },
  { to: '/trash', icon: Trash2Icon, label: () => '回收站' },
]

async function logout() {
  try { await api.logout() } catch { /* ignore */ }
  store.session = { authenticated: false, csrfToken: '' }
  router.push('/login')
}

async function syncPosts() {
  store.postsLoading = true
  try { store.posts = await api.posts() }
  catch (e: any) { notify.error(e) }
  finally { store.postsLoading = false }
}

router.beforeEach((to, from) => {
  if (to.name !== 'login' && !store.session.authenticated) return { name: 'login' }
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
  } catch (e: any) { notify.error(e) }
})
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <Toaster position="top-right" rich-colors :expand="true" />
    <CommandPalette v-model:open="paletteOpen" />

    <a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:z-[9999] focus:top-2 focus:left-2 focus:px-3 focus:py-2 focus:bg-background focus:border focus:border-border focus:rounded">跳到主内容</a>

    <div v-if="isLoginPage || !isAuthed" class="min-h-svh">
      <RouterView />
    </div>

    <div v-else class="flex h-svh overflow-hidden">
      <!-- Sidebar -->
      <aside class="w-[240px] shrink-0 border-r border-border bg-sidebar flex flex-col" aria-label="主导航">
        <!-- Brand -->
        <div class="flex items-center gap-3 px-5 py-5">
          <div class="font-display text-3xl text-accent leading-none select-none">博</div>
          <div class="min-w-0">
            <div class="font-serif font-semibold text-sm truncate leading-tight">{{ store.config?.site.name || 'Blog Studio' }}</div>
            <div class="text-[10px] text-muted-foreground tracking-wider">管理后台</div>
          </div>
        </div>

        <Separator class="mx-4 w-auto" />

        <ScrollArea class="flex-1">
          <nav class="px-3 py-3 space-y-4" role="navigation" aria-label="主菜单">
            <!-- Content group -->
            <div>
              <div class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">内容</div>
              <div class="space-y-0.5">
                <Tooltip v-for="item in contentNav" :key="item.to">
                  <TooltipTrigger as-child>
                    <RouterLink
                      :to="item.to"
                      class="group relative flex items-center gap-2.5 rounded px-2.5 py-1.5 text-sm text-sidebar-foreground transition-all duration-200 hover:translate-x-0.5"
                      :class="route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50'"
                    >
                      <!-- Active indicator bar -->
                      <div v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                        class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                      <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                      <span class="flex-1">{{ item.label() }}</span>
                      <Badge v-if="item.count?.()" variant="secondary" class="h-5 min-w-[20px] justify-center px-1.5 text-[10px]">
                        {{ item.count() }}
                      </Badge>
                    </RouterLink>
                  </TooltipTrigger>
                  <TooltipContent side="right" :side-offset="8">{{ item.label() }}</TooltipContent>
                </Tooltip>
              </div>
            </div>

            <!-- System group -->
            <div>
              <div class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">系统</div>
              <div class="space-y-0.5">
                <Tooltip v-for="item in systemNav" :key="item.to">
                  <TooltipTrigger as-child>
                    <RouterLink
                      :to="item.to"
                      class="group relative flex items-center gap-2.5 rounded px-2.5 py-1.5 text-sm text-sidebar-foreground transition-all duration-200 hover:translate-x-0.5"
                      :class="route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50'"
                    >
                      <div v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                        class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                      <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                      <span class="flex-1">{{ item.label() }}</span>
                      <span v-if="item.dot?.()" class="h-2 w-2 rounded-full bg-destructive animate-pulse" role="status" aria-label="服务异常" />
                    </RouterLink>
                  </TooltipTrigger>
                  <TooltipContent side="right" :side-offset="8">{{ item.label() }}</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </nav>
        </ScrollArea>

        <Separator class="mx-4 w-auto" />

        <div class="px-3 py-3 space-y-0.5">
          <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5 text-muted-foreground" :disabled="store.postsLoading" @click="syncPosts">
            <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" />
            <span>{{ t.sync }}</span>
          </Button>
          <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5 text-muted-foreground" @click="logout">
            <LogOutIcon class="h-4 w-4" />
            <span>{{ t.logout }}</span>
          </Button>
        </div>
      </aside>

      <!-- Main stage -->
      <div class="flex-1 flex flex-col min-w-0">
        <header class="flex items-center h-12 px-6 border-b border-border shrink-0" role="banner">
          <h1 class="font-serif font-semibold text-base tracking-tight">{{ pageTitle }}</h1>
          <div class="flex items-center gap-2 ml-auto">
            <Badge v-if="editorSaving" variant="secondary" class="text-xs animate-pulse">{{ t.saving }}…</Badge>
            <Badge v-else-if="editorDirty" variant="destructive" class="text-xs">● {{ t.unsaved }}</Badge>
            <Badge v-else-if="store.editor.savedAt" variant="outline" class="text-xs text-ok border-ok/30">{{ t.saved }}</Badge>

            <Separator orientation="vertical" class="h-4 mx-1" />

            <Button variant="ghost" size="sm" class="h-7 px-2 text-xs gap-1 text-muted-foreground" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
              <Globe2Icon class="h-3.5 w-3.5" />{{ lang === 'zh' ? 'EN' : '中文' }}
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :title="t.theme" @click="toggleTheme">
              <MoonIcon v-if="!isDark" class="h-3.5 w-3.5" />
              <SunIcon v-else class="h-3.5 w-3.5" />
            </Button>
          </div>
        </header>

        <main id="main-content" class="flex-1 overflow-auto p-6" tabindex="-1">
          <transition name="page" mode="out-in">
            <RouterView :key="route.path" />
          </transition>
        </main>
      </div>
    </div>
  </TooltipProvider>
</template>

<style scoped>
.page-enter-active {
  animation: fade-up 0.25s ease-out;
}
.page-leave-active {
  animation: fade-in 0.12s ease-in reverse;
}
</style>
