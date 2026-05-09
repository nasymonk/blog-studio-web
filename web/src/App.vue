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
  <Toaster position="top-right" rich-colors :expand="true" />
  <CommandPalette v-model:open="paletteOpen" />

  <!-- Skip navigation link for keyboard users -->
  <a href="#main-content" class="skip-link">跳到主内容</a>

  <div v-if="isLoginPage || !isAuthed" class="login-outer">
    <RouterView />
  </div>

  <div v-else class="app-shell">
    <aside class="sidebar" aria-label="主导航">
      <div class="sidebar-brand" aria-label="Blog Studio">
        <svg class="brand-mark" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
          <rect width="32" height="32" rx="8" fill="var(--accent)" />
          <text x="16" y="22" text-anchor="middle" fill="white" font-size="18" font-weight="700" font-family="serif">B</text>
        </svg>
        <span class="brand-name">{{ store.config?.site.name || 'Blog Studio' }}</span>
      </div>

      <nav class="sidebar-nav" role="navigation" aria-label="主菜单">
        <RouterLink to="/posts" class="nav-btn" active-class="nav-btn-active" :aria-label="t.posts">
          <FileTextIcon :size="18" aria-hidden="true" />
          <span>{{ t.posts }}</span>
          <b v-if="store.posts.length" class="nav-count" :aria-label="`${store.posts.length} 篇文章`">{{ store.posts.length }}</b>
        </RouterLink>
        <RouterLink to="/home" class="nav-btn" active-class="nav-btn-active" :aria-label="t.home">
          <HomeIcon :size="18" aria-hidden="true" />
          <span>{{ t.home }}</span>
        </RouterLink>
        <RouterLink to="/now" class="nav-btn" active-class="nav-btn-active" :aria-label="t.now">
          <ClockIcon :size="18" aria-hidden="true" />
          <span>{{ t.now }}</span>
        </RouterLink>
        <RouterLink to="/settings" class="nav-btn" active-class="nav-btn-active" :aria-label="t.settings">
          <SettingsIcon :size="18" aria-hidden="true" />
          <span>{{ t.settings }}</span>
        </RouterLink>
        <RouterLink to="/health" class="nav-btn" active-class="nav-btn-active" :aria-label="t.health">
          <HeartPulseIcon :size="18" aria-hidden="true" />
          <span>{{ t.health }}</span>
          <span
            v-if="store.health && store.health.status !== 'ok'"
            class="nav-dot nav-dot-error"
            role="status"
            aria-label="服务异常"
          ></span>
        </RouterLink>
        <RouterLink to="/trash" class="nav-btn" active-class="nav-btn-active" aria-label="回收站">
          <Trash2Icon :size="18" aria-hidden="true" />
          <span>回收站</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <button class="nav-btn" :disabled="store.postsLoading" :aria-label="t.sync" @click="syncPosts">
          <RefreshCwIcon :size="16" :class="{ spin: store.postsLoading }" aria-hidden="true" />
          <span>{{ t.sync }}</span>
        </button>
        <button class="nav-btn" :aria-label="t.logout" @click="logout">
          <LogOutIcon :size="16" aria-hidden="true" />
          <span>{{ t.logout }}</span>
        </button>
      </div>
    </aside>

    <div class="stage">
      <header class="topbar" role="banner">
        <h1 class="topbar-title">{{ pageTitle }}</h1>
        <div class="topbar-actions">
          <span v-if="editorSaving" class="save-label save-label-saving" role="status">{{ t.saving }}…</span>
          <span v-else-if="editorDirty" class="save-label save-label-dirty" role="status">● {{ t.unsaved }}</span>
          <span v-else-if="store.editor.savedAt" class="save-label save-label-ok" role="status">{{ t.saved }}</span>

          <button class="btn btn-ghost btn-sm" :aria-label="`切换语言为 ${lang === 'zh' ? 'English' : '中文'}`" @click="setLang(lang === 'zh' ? 'en' : 'zh')">
            <Globe2Icon :size="15" aria-hidden="true" />{{ lang === 'zh' ? 'EN' : '中文' }}
          </button>
          <button class="btn btn-ghost btn-icon btn-sm" :aria-label="isDark ? t.lightMode : t.darkMode" :title="t.theme" @click="toggleTheme">
            <MoonIcon v-if="!isDark" :size="15" aria-hidden="true" />
            <SunIcon v-else :size="15" aria-hidden="true" />
          </button>
        </div>
      </header>

      <div v-if="store.banner" class="status-banner status-banner-warn" role="alert">
        {{ store.banner.message }}
        <button class="btn btn-ghost btn-sm" aria-label="关闭提示" @click="store.banner = null">✕</button>
      </div>

      <main id="main-content" class="page-content" tabindex="-1">
        <RouterView />
      </main>
    </div>
  </div>
</template>
