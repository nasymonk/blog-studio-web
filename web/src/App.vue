<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { useToast } from '@/composables/useToast'
import ToastContainer from '@/components/ToastContainer.vue'
import {
  FileTextIcon, SettingsIcon, HomeIcon, ClockIcon, HeartPulseIcon,
  LogOutIcon, MoonIcon, SunIcon, Globe2Icon, RefreshCwIcon, Trash2Icon, TagsIcon,
  MenuIcon
} from 'lucide-vue-next'
import { useEventListener } from '@vueuse/core'
import { api, setUnauthorizedHandler } from '@/services/api'
import { createStore, provideStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import CommandPalette from '@/components/CommandPalette.vue'
import ErrorBoundary from '@/components/ErrorBoundary.vue'
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
const toast = useToast()

watch(() => store.banner, (b) => {
  if (!b) return
  const fn = b.type === 'error' ? toast.error : b.type === 'ok' ? toast.success : toast.warning
  fn(b.message)
  store.banner = null
})

const isAuthed = computed(() => store.session.authenticated)
const isLoginPage = computed(() => route.name === 'login')

const paletteOpen = ref(false)
const sidebarOpen = ref(false)
const sidebarCollapsed = ref(localStorage.getItem('sidebar:collapsed') === 'true')
watch(sidebarCollapsed, (val) => localStorage.setItem('sidebar:collapsed', String(val)))

useEventListener('keydown', (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    if (isAuthed.value) paletteOpen.value = true
  }
  if ((e.metaKey || e.ctrlKey) && e.key === '\\') {
    e.preventDefault()
    sidebarCollapsed.value = !sidebarCollapsed.value
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
  if (name === 'tags') return t.value.tagManagement
  return 'Blog Studio'
})

const editorDirty = computed(() => route.name === 'editor' && store.editor.dirty)
const editorSaving = computed(() => route.name === 'editor' && store.editor.saving)

setUnauthorizedHandler(() => {
  store.session.authenticated = false
  router.push('/login')
})

const contentNav = [
  { to: '/posts', icon: FileTextIcon, label: () => t.value.posts, count: () => store.posts.length },
  { to: '/tags', icon: TagsIcon, label: () => t.value.tags },
  { to: '/home', icon: HomeIcon, label: () => t.value.home },
  { to: '/now', icon: ClockIcon, label: () => t.value.now },
]

const systemNav = [
  { to: '/settings', icon: SettingsIcon, label: () => t.value.settings },
  { to: '/health', icon: HeartPulseIcon, label: () => t.value.health, dot: () => store.health && store.health.status !== 'ok' },
  { to: '/trash', icon: Trash2Icon, label: () => t.value.trash },
]

async function logout() {
  try { await api.logout() } catch { /* ignore */ }
  store.session = { authenticated: false, csrfToken: '' }
  router.push('/login')
}

async function syncPosts() {
  store.postsLoading = true
  try { store.posts = await api.posts() }
  catch (e: unknown) { notify.error(e) }
  finally { store.postsLoading = false }
}

watch(() => route.path, () => { sidebarOpen.value = false })

router.beforeEach((to, from) => {
  if (to.name !== 'login' && !store.session.authenticated) return { name: 'login' }
  if (from.name === 'editor' && to.name !== 'editor' && store.editor.dirty) {
    if (!window.confirm(t.value.confirmLeave)) return false
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
  } catch (e: unknown) { notify.error(e) }
})
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <ToastContainer />
    <CommandPalette v-model:open="paletteOpen" />

    <a href="#main-content" :aria-label="t.skipToContent" class="sr-only focus:not-sr-only focus:absolute focus:z-[9999] focus:top-2 focus:left-2 focus:px-3 focus:py-2 focus:bg-background focus:border focus:border-border focus:rounded">{{ t.skipToContent }}</a>

    <div v-if="isLoginPage || !isAuthed" class="min-h-svh">
      <ErrorBoundary>
        <RouterView />
      </ErrorBoundary>
    </div>

    <div v-else class="flex h-svh overflow-hidden">
      <!-- Mobile sidebar drawer -->
      <div v-if="sidebarOpen" class="fixed inset-0 z-50 md:hidden">
        <div class="absolute inset-0 bg-black/50" @click="sidebarOpen = false" />
        <aside class="relative w-[280px] h-full bg-sidebar border-r border-border flex flex-col" aria-label="主导航">
          <!-- Brand -->
          <div class="flex items-center gap-3 px-5 py-5">
            <div class="font-display text-3xl text-accent leading-none select-none">博</div>
            <div class="min-w-0">
              <div class="font-serif font-semibold text-sm truncate leading-tight">{{ store.config?.site.name || 'Blog Studio' }}</div>
              <div class="text-[10px] text-muted-foreground tracking-wider">{{ t.adminSubtitle }}</div>
            </div>
          </div>

          <Separator class="mx-4 w-auto" />

          <ScrollArea class="flex-1">
            <nav class="px-3 py-3 space-y-4" role="navigation" aria-label="主菜单">
              <!-- Content group -->
              <div>
                <div class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">{{ t.contentGroup }}</div>
                <div class="space-y-0.5">
                  <RouterLink
                    v-for="item in contentNav" :key="item.to"
                    :to="item.to"
                    class="group relative flex items-center gap-2.5 rounded px-2.5 py-2.5 text-sm text-sidebar-foreground transition-all duration-200"
                    :class="route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50'"
                  >
                    <div
                      v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                      class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                    <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                    <span class="flex-1">{{ item.label() }}</span>
                    <Badge v-if="item.count?.()" variant="secondary" class="h-5 min-w-[20px] justify-center px-1.5 text-[10px]">
                      {{ item.count() }}
                    </Badge>
                  </RouterLink>
                </div>
              </div>

              <!-- System group -->
              <div>
                <div class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">{{ t.systemGroup }}</div>
                <div class="space-y-0.5">
                  <RouterLink
                    v-for="item in systemNav" :key="item.to"
                    :to="item.to"
                    class="group relative flex items-center gap-2.5 rounded px-2.5 py-2.5 text-sm text-sidebar-foreground transition-all duration-200"
                    :class="route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50'"
                  >
                    <div
                      v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                      class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                    <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                    <span class="flex-1">{{ item.label() }}</span>
                    <span v-if="item.dot?.()" class="h-2 w-2 rounded-full bg-destructive animate-pulse" role="status" aria-label="服务异常" />
                  </RouterLink>
                </div>
              </div>
            </nav>
          </ScrollArea>

          <Separator class="mx-4 w-auto" />

          <div class="px-3 py-3 space-y-0.5">
            <Button variant="ghost" class="w-full justify-start gap-2.5 h-10 text-sm px-2.5 text-muted-foreground" :aria-label="t.sync" :disabled="store.postsLoading" @click="syncPosts">
              <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" />
              <span>{{ t.sync }}</span>
            </Button>
            <Button variant="ghost" class="w-full justify-start gap-2.5 h-10 text-sm px-2.5 text-muted-foreground" :aria-label="t.logout" @click="logout">
              <LogOutIcon class="h-4 w-4" />
              <span>{{ t.logout }}</span>
            </Button>
          </div>
        </aside>
      </div>

      <!-- Desktop sidebar -->
      <aside
        class="hidden md:flex shrink-0 border-r border-border bg-sidebar flex-col transition-all duration-200"
        :class="sidebarCollapsed ? 'w-[56px]' : 'w-[240px]'"
        aria-label="主导航"
      >
        <!-- Brand -->
        <div class="flex items-center gap-3 px-5 py-5">
          <div class="font-display text-3xl text-accent leading-none select-none">博</div>
          <div v-if="!sidebarCollapsed" class="min-w-0">
            <div class="font-serif font-semibold text-sm truncate leading-tight">{{ store.config?.site.name || 'Blog Studio' }}</div>
            <div class="text-[10px] text-muted-foreground tracking-wider">{{ t.adminSubtitle }}</div>
          </div>
        </div>

        <Separator class="mx-4 w-auto" />

        <ScrollArea class="flex-1">
          <nav class="px-3 py-3 space-y-4" role="navigation" aria-label="主菜单">
            <!-- Content group -->
            <div>
              <div v-if="!sidebarCollapsed" class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">{{ t.contentGroup }}</div>
              <div class="space-y-0.5">
                <Tooltip v-for="item in contentNav" :key="item.to">
                  <TooltipTrigger as-child>
                    <RouterLink
                      :to="item.to"
                      class="group relative flex items-center gap-2.5 rounded px-2.5 py-1.5 text-sm text-sidebar-foreground transition-all duration-200 hover:translate-x-0.5"
                      :class="[
                        route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50',
                        sidebarCollapsed && 'justify-center px-0',
                      ]"
                    >
                      <!-- Active indicator bar -->
                      <div
                        v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                        class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                      <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                      <span v-if="!sidebarCollapsed" class="flex-1">{{ item.label() }}</span>
                      <Badge v-if="!sidebarCollapsed && item.count?.()" variant="secondary" class="h-5 min-w-[20px] justify-center px-1.5 text-[10px]">
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
              <div v-if="!sidebarCollapsed" class="px-2.5 mb-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/60">{{ t.systemGroup }}</div>
              <div class="space-y-0.5">
                <Tooltip v-for="item in systemNav" :key="item.to">
                  <TooltipTrigger as-child>
                    <RouterLink
                      :to="item.to"
                      class="group relative flex items-center gap-2.5 rounded px-2.5 py-1.5 text-sm text-sidebar-foreground transition-all duration-200 hover:translate-x-0.5"
                      :class="[
                        route.path === item.to || route.path.startsWith(item.to + '/') ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium' : 'hover:bg-sidebar-accent/50',
                        sidebarCollapsed && 'justify-center px-0',
                      ]"
                    >
                      <div
                        v-if="route.path === item.to || route.path.startsWith(item.to + '/')"
                        class="absolute left-0 top-1 bottom-1 w-0.5 bg-accent rounded-full origin-top animate-fade-in" />
                      <component :is="item.icon" class="h-4 w-4 shrink-0 opacity-70" aria-hidden="true" />
                      <span v-if="!sidebarCollapsed" class="flex-1">{{ item.label() }}</span>
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
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5 text-muted-foreground" :class="sidebarCollapsed && 'justify-center px-0'" :aria-label="t.sync" :disabled="store.postsLoading" @click="syncPosts">
                <RefreshCwIcon class="h-4 w-4" :class="{ 'animate-spin': store.postsLoading }" />
                <span v-if="!sidebarCollapsed">{{ t.sync }}</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent v-if="sidebarCollapsed" side="right" :side-offset="8">{{ t.sync }}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" class="w-full justify-start gap-2.5 h-8 text-sm px-2.5 text-muted-foreground" :class="sidebarCollapsed && 'justify-center px-0'" :aria-label="t.logout" @click="logout">
                <LogOutIcon class="h-4 w-4" />
                <span v-if="!sidebarCollapsed">{{ t.logout }}</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent v-if="sidebarCollapsed" side="right" :side-offset="8">{{ t.logout }}</TooltipContent>
          </Tooltip>
        </div>
      </aside>

      <!-- Main stage -->
      <div class="flex-1 flex flex-col min-w-0">
        <header class="flex items-center h-12 px-3 md:px-6 border-b border-border shrink-0 gap-2" role="banner">
          <Button variant="ghost" size="sm" class="md:hidden h-8 w-8 p-0 shrink-0" aria-label="Open menu" @click="sidebarOpen = true">
            <MenuIcon class="h-4 w-4" />
          </Button>
          <h1 class="font-serif font-semibold text-base tracking-tight truncate">{{ pageTitle }}</h1>
          <div class="flex items-center gap-1 md:gap-2 ml-auto shrink-0">
            <Badge v-if="editorSaving" variant="secondary" class="text-xs animate-pulse hidden sm:inline-flex">{{ t.saving }}…</Badge>
            <Badge v-else-if="editorDirty" variant="destructive" class="text-xs hidden sm:inline-flex">● {{ t.unsaved }}</Badge>
            <Badge v-else-if="store.editor.savedAt" variant="outline" class="text-xs text-ok border-ok/30 hidden sm:inline-flex">{{ t.saved }}</Badge>

            <Separator orientation="vertical" class="h-4 mx-1 hidden sm:block" />

            <Button variant="ghost" size="sm" class="h-8 px-2 text-xs gap-1 text-muted-foreground" :aria-label="lang === 'zh' ? 'Switch to English' : '切换到中文'" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
              <Globe2Icon class="h-3.5 w-3.5" />{{ lang === 'zh' ? 'EN' : '中文' }}
            </Button>
            <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground" :title="t.theme" @click="toggleTheme">
              <MoonIcon v-if="!isDark" class="h-3.5 w-3.5" />
              <SunIcon v-else class="h-3.5 w-3.5" />
            </Button>
          </div>
        </header>

        <main id="main-content" class="flex-1 overflow-auto p-3 md:p-6" tabindex="-1">
          <ErrorBoundary>
            <transition name="page" mode="out-in">
              <RouterView :key="route.path" />
            </transition>
          </ErrorBoundary>
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
