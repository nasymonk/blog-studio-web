<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AlertTriangle, BookOpen, CheckCircle2, ChevronDown, Clock, Eye, FileText, Globe2, HeartPulse, History, Home, Loader2, LogOut, Moon, Newspaper, RefreshCw, Rocket, Save, Send, Settings, Shield, Sun, XCircle } from 'lucide-vue-next'
import { markdown } from '@codemirror/lang-markdown'
import Codemirror from 'vue-codemirror6'
import { api, type AppError, type Config, type PostDraft, type PostState, type PreviewResult, type PublishResult, type SiteInfo } from './services/api'

type Lang = 'zh' | 'en'
type View = 'posts' | 'editor' | 'preview' | 'home' | 'now' | 'settings'
type Notice = { type: 'ok' | 'warn' | 'error'; title: string; detail?: string; suggestion?: string }

const dict = {
  zh: {
    product: 'Blog Studio',
    loginTitle: '博客后台',
    loginHint: '管理文章、预览、发布和站点设置。',
    password: '管理员密码',
    login: '登录',
    logout: '退出',
    posts: '文章',
    homeSettings: '主页',
    nowPage: 'Now',
    settings: '设置',
    preview: '预览',
    save: '保存',
    publishBlog: '发布到博客',
    publishWechat: '保存公众号草稿',
    sync: '同步',
    newPost: '新建文章',
    slugPlaceholder: '文章短链接，如 my-first-post',
    allPosts: '全部文章',
    drafts: '草稿',
    tags: '标签',
    updated: '同步状态',
    editor: '写作',
    previewTitle: '博客预览',
    homeTitle: '主页设置',
    nowTitle: 'Now 页面',
    settingsTitle: '系统设置',
    title: '文章标题',
    shortLink: '文章短链接',
    tagsLabel: '标签',
    categories: '分类',
    cover: '封面图片',
    draft: '保存为草稿',
    overwrite: '允许覆盖远端新版本',
    rollback: '回滚到上一版本',
    noPreview: '暂无预览',
    health: '健康检查',
    saveSettings: '保存设置',
    audit: '操作记录',
    basicInfo: '基本信息',
    writing: '写作与发布',
    wechatSettings: '公众号',
    security: '账户安全',
    advanced: '高级详情',
    siteName: '后台显示名称',
    adminPath: '后台访问路径',
    postsFolder: '文章保存位置',
    siteBuildMode: '生成网站方式',
    buildCompressed: '生成并压缩',
    buildStandard: '普通生成',
    previewRetention: '预览保留时间（分钟）',
    uploadLimit: '单个图片上限（MB）',
    wechatEnabled: '启用公众号草稿',
    wechatApp: '公众号应用编号',
    secretHint: '密钥不会在页面回显，请在服务器环境或密钥文件中保存。',
    currentPassword: '当前密码',
    newPassword: '新密码',
    changePassword: '修改密码',
    language: '语言',
    theme: '外观',
    light: '浅色',
    dark: '深色',
    saved: '已保存',
    loaded: '已加载',
    syncing: '正在同步文章...',
    synced: '同步完成',
    previewReady: '预览已生成',
    settingsSaved: '设置已保存并立即生效。',
    actionReady: '等待操作。',
    emptyPosts: '还没有文章。',
    openAdvanced: '查看高级详情',
    closeAdvanced: '收起高级详情',
    confirmOverwriteTitle: '确认覆盖远端新版本',
    confirmOverwriteBody: '远端文章可能已经被其他方式修改。勾选后发布会覆盖远端内容。',
    ok: '正常',
    warn: '注意',
    error: '错误',
    avatar: '头像',
    avatarHint: '上传后下次发布时生效',
    uploadAvatar: '更换头像',
    tagline: '个人简介',
    taglinePlaceholder: '首页显示的个签文字',
    avatarUploaded: '头像已上传，发布任意文章后生效。',
    nowHint: '直接编辑 Now 页面的 Markdown 源码（含 frontmatter）',
  },
  en: {
    product: 'Blog Studio',
    loginTitle: 'Blog Admin',
    loginHint: 'Manage posts, previews, publishing, and site settings.',
    password: 'Admin password',
    login: 'Sign in',
    logout: 'Sign out',
    posts: 'Posts',
    homeSettings: 'Homepage',
    nowPage: 'Now',
    settings: 'Settings',
    preview: 'Preview',
    save: 'Save',
    publishBlog: 'Publish to Blog',
    publishWechat: 'Save WeChat Draft',
    sync: 'Sync',
    newPost: 'New Post',
    slugPlaceholder: 'Readable link, such as my-first-post',
    allPosts: 'All Posts',
    drafts: 'Drafts',
    tags: 'Tags',
    updated: 'Sync Status',
    editor: 'Writing',
    previewTitle: 'Blog Preview',
    homeTitle: 'Homepage Settings',
    nowTitle: 'Now Page',
    settingsTitle: 'Settings',
    title: 'Post Title',
    shortLink: 'Readable Link',
    tagsLabel: 'Tags',
    categories: 'Categories',
    cover: 'Cover Image',
    draft: 'Keep as Draft',
    overwrite: 'Allow Overwriting Newer Remote Copy',
    rollback: 'Restore Previous Version',
    noPreview: 'No preview yet',
    health: 'Health Check',
    saveSettings: 'Save Settings',
    audit: 'Activity Log',
    basicInfo: 'Basic Info',
    writing: 'Writing & Publishing',
    wechatSettings: 'WeChat',
    security: 'Account Security',
    advanced: 'Advanced Details',
    siteName: 'Admin Display Name',
    adminPath: 'Admin URL Path',
    postsFolder: 'Posts Folder',
    siteBuildMode: 'Site Build Mode',
    buildCompressed: 'Build and Compress',
    buildStandard: 'Standard Build',
    previewRetention: 'Preview Retention (minutes)',
    uploadLimit: 'Image Size Limit (MB)',
    wechatEnabled: 'Enable WeChat Drafts',
    wechatApp: 'WeChat App ID',
    secretHint: 'Secrets are never shown here. Store them in server environment or the secrets file.',
    currentPassword: 'Current Password',
    newPassword: 'New Password',
    changePassword: 'Change Password',
    language: 'Language',
    theme: 'Theme',
    light: 'Light',
    dark: 'Dark',
    saved: 'Saved',
    loaded: 'Loaded',
    syncing: 'Syncing posts...',
    synced: 'Sync complete',
    previewReady: 'Preview generated',
    settingsSaved: 'Settings saved and applied.',
    actionReady: 'Ready.',
    emptyPosts: 'No posts yet.',
    openAdvanced: 'Show advanced details',
    closeAdvanced: 'Hide advanced details',
    confirmOverwriteTitle: 'Confirm overwrite',
    confirmOverwriteBody: 'The remote post may have been changed elsewhere. Publishing with this checked will overwrite it.',
    ok: 'OK',
    warn: 'Warning',
    error: 'Error',
    avatar: 'Avatar',
    avatarHint: 'Takes effect on next publish',
    uploadAvatar: 'Change Avatar',
    tagline: 'Tagline',
    taglinePlaceholder: 'Homepage tagline text',
    avatarUploaded: 'Avatar uploaded. Publish any post to apply.',
    nowHint: 'Edit the raw Markdown source of your Now page (including frontmatter)',
  }
} as const

const lang = ref<Lang>((localStorage.getItem('blog-studio-lang') as Lang) || 'zh')
const dark = ref(localStorage.getItem('blog-studio-theme') === 'dark')
const t = computed(() => dict[lang.value])
const authed = ref(false)
const password = ref('')
const loading = ref(false)
const busyText = ref('')
const view = ref<View>('posts')
const posts = ref<PostState[]>([])
const selected = ref<PostDraft | null>(null)
const config = ref<Config | null>(null)
const preview = ref<PreviewResult | null>(null)
const siteInfo = ref<SiteInfo>({ description: '', profileImage: '' })
const nowRaw = ref('')
const health = ref<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null>(null)
const audit = ref<unknown[]>([])
const notice = ref<Notice>({ type: 'ok', title: dict.zh.actionReady })
const confirmOverwrite = ref(false)
const showAdvanced = ref(false)
const newSlug = ref('')
const currentPassword = ref('')
const newAdminPassword = ref('')

const pageTitle = computed(() => {
  if (view.value === 'editor') return selected.value?.frontMatter.title || t.value.editor
  if (view.value === 'preview') return t.value.previewTitle
  if (view.value === 'home') return t.value.homeTitle
  if (view.value === 'now') return t.value.nowTitle
  if (view.value === 'settings') return t.value.settingsTitle
  return t.value.posts
})
const draftCount = computed(() => posts.value.filter((post) => post.draft).length)
const tagCount = computed(() => new Set(posts.value.flatMap((post) => post.tags || [])).size)
const uploadLimitMb = computed({
  get: () => Math.round((config.value?.maxUploadBytes || 0) / 1024 / 1024),
  set: (value) => { if (config.value) config.value.maxUploadBytes = Number(value || 1) * 1024 * 1024 }
})
const title = computed({
  get: () => selected.value?.frontMatter.title ?? '',
  set: (value) => { if (selected.value) selected.value.frontMatter.title = value }
})
const tags = computed({
  get: () => selected.value?.frontMatter.tags.join(', ') ?? '',
  set: (value) => { if (selected.value) selected.value.frontMatter.tags = splitList(value) }
})
const categories = computed({
  get: () => selected.value?.frontMatter.categories.join(', ') ?? '',
  set: (value) => { if (selected.value) selected.value.frontMatter.categories = splitList(value) }
})
const image = computed({
  get: () => selected.value?.frontMatter.image ?? '',
  set: (value) => { if (selected.value) selected.value.frontMatter.image = value }
})
const body = computed({
  get: () => selected.value?.body ?? '',
  set: (value) => { if (selected.value) selected.value.body = value }
})

onMounted(async () => {
  applyTheme()
  try {
    const session = await api.session()
    authed.value = session.authenticated
    if (session.csrfToken) api.setCSRF(session.csrfToken)
    if (authed.value) await bootstrap()
  } catch (error) {
    setError(error)
  }
})

function setLang(value: Lang) {
  lang.value = value
  localStorage.setItem('blog-studio-lang', value)
  if (notice.value.title === dict.zh.actionReady || notice.value.title === dict.en.actionReady) {
    notice.value = { type: 'ok', title: t.value.actionReady }
  }
}

function toggleTheme() {
  dark.value = !dark.value
  localStorage.setItem('blog-studio-theme', dark.value ? 'dark' : 'light')
  applyTheme()
}

function applyTheme() {
  document.documentElement.dataset.theme = dark.value ? 'dark' : 'light'
}

async function withBusy<T>(label: string, task: () => Promise<T>) {
  loading.value = true
  busyText.value = label
  try {
    return await task()
  } finally {
    loading.value = false
    busyText.value = ''
  }
}

async function login() {
  await withBusy(t.value.login, async () => {
    try {
      const result = await api.login(password.value)
      api.setCSRF(result.csrfToken)
      authed.value = true
      password.value = ''
      await bootstrap()
    } catch (error) {
      setError(error)
    }
  })
}

async function logout() {
  await api.logout()
  authed.value = false
  selected.value = null
}

async function bootstrap() {
  config.value = await api.config()
  await Promise.all([syncPosts(false), runHealth(), loadSiteInfo()])
}

async function syncPosts(forceView = true) {
  await withBusy(t.value.syncing, async () => {
    try {
      posts.value = await api.posts()
      if (forceView) view.value = 'posts'
      setNotice('ok', `${t.value.synced} · ${posts.value.length}`)
    } catch (error) {
      setError(error)
    }
  })
}

async function openPost(slug: string) {
  await withBusy(t.value.loaded, async () => {
    try {
      selected.value = await api.post(slug)
      view.value = 'editor'
      setNotice('ok', `${t.value.loaded}: ${slug}`)
    } catch (error) {
      setError(error)
    }
  })
}

function createPost() {
  const slug = newSlug.value.trim() || `new-post-${Date.now()}`
  selected.value = {
    slug,
    frontMatter: { title: lang.value === 'zh' ? '未命名文章' : 'Untitled Post', draft: true, tags: [], categories: [] },
    body: lang.value === 'zh' ? '## 背景\n\n\n## 正文\n\n\n## 总结\n' : '## Context\n\n\n## Body\n\n\n## Notes\n',
    assets: []
  }
  newSlug.value = ''
  view.value = 'editor'
}

async function saveDraft() {
  if (!selected.value) return
  try {
    await api.saveDraft(selected.value)
    setNotice('ok', t.value.saved)
  } catch (error) {
    setError(error)
  }
}

async function publishBlog() {
  if (!selected.value) return
  if (confirmOverwrite.value && !window.confirm(`${t.value.confirmOverwriteTitle}\n\n${t.value.confirmOverwriteBody}`)) return
  await withBusy(t.value.publishBlog, async () => {
    try {
      const result = await api.publishBlog(selected.value!, confirmOverwrite.value)
      showResult(result)
      if (result.status === 'success') await syncPosts(false)
    } catch (error) {
      setError(error)
    }
  })
}

async function publishWechat() {
  if (!selected.value) return
  await withBusy(t.value.publishWechat, async () => {
    try {
      showResult(await api.publishWechat(selected.value!))
    } catch (error) {
      setError(error)
    }
  })
}

async function createPreview() {
  if (!selected.value) return
  await withBusy(t.value.preview, async () => {
    try {
      preview.value = await api.preview(selected.value!)
      view.value = 'preview'
      preview.value.error ? setError(preview.value.error) : setNotice('ok', t.value.previewReady, preview.value.url)
    } catch (error) {
      setError(error)
    }
  })
}

async function rollback() {
  if (!selected.value) return
  await withBusy(t.value.rollback, async () => {
    try {
      showResult(await api.rollback(selected.value!.slug))
      await syncPosts(false)
    } catch (error) {
      setError(error)
    }
  })
}

async function loadSiteInfo() {
  try {
    siteInfo.value = await api.getSite()
  } catch {
    // non-critical
  }
}

async function saveSiteInfo() {
  try {
    await api.saveSite(siteInfo.value)
    setNotice('ok', t.value.saved)
  } catch (error) {
    setError(error)
  }
}

async function handleAvatarUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await withBusy(t.value.uploadAvatar, async () => {
    try {
      const result = await api.uploadAvatar(file)
      siteInfo.value.profileImage = result.path
      setNotice('ok', t.value.avatarUploaded)
    } catch (error) {
      setError(error)
    }
  })
  input.value = ''
}

async function loadNowPage() {
  try {
    const result = await api.getNowPage()
    nowRaw.value = result.raw
  } catch (error) {
    setError(error)
  }
}

async function saveNowPage() {
  await withBusy(t.value.save, async () => {
    try {
      await api.saveNowPage(nowRaw.value)
      setNotice('ok', t.value.saved)
    } catch (error) {
      setError(error)
    }
  })
}

async function runHealth() {
  try {
    health.value = await api.health()
  } catch (error) {
    setError(error)
  }
}

async function loadAudit() {
  audit.value = await api.audit()
}

async function saveConfig() {
  if (!config.value) return
  try {
    config.value = await api.saveConfig(config.value)
    setNotice('ok', t.value.settingsSaved)
    await runHealth()
  } catch (error) {
    setError(error)
  }
}

async function changePassword() {
  try {
    await api.changePassword(currentPassword.value, newAdminPassword.value)
    currentPassword.value = ''
    newAdminPassword.value = ''
    setNotice('ok', t.value.saved)
  } catch (error) {
    setError(error)
  }
}

function showResult(result: PublishResult) {
  if (result.error) {
    setError(result.error)
    return
  }
  setNotice('ok', `${result.target} ${result.status}`, [`audit: ${result.auditId}`, `backup: ${result.backupId || '-'}`].join('\n'))
}

function splitList(value: string) {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}

function setNotice(type: Notice['type'], title: string, detail = '', suggestion = '') {
  notice.value = { type, title, detail, suggestion }
}

function setError(error: unknown) {
  const app = error as AppError
  if (app?.message) {
    setNotice('error', app.message, app.technicalDetail, app.suggestion)
    return
  }
  setNotice('error', error instanceof Error ? error.message : String(error))
}
</script>

<template>
  <main v-if="!authed" class="login-page">
    <form class="login-card" @submit.prevent="login">
      <div class="login-toolbar">
        <button type="button" class="ghost" @click="setLang(lang === 'zh' ? 'en' : 'zh')"><Globe2 :size="16" />{{ lang === 'zh' ? 'EN' : '中文' }}</button>
        <button type="button" class="icon-button" @click="toggleTheme" :aria-label="t.theme"><Moon v-if="!dark" :size="16" /><Sun v-else :size="16" /></button>
      </div>
      <div class="logo">B</div>
      <h1>{{ t.loginTitle }}</h1>
      <p>{{ t.loginHint }}</p>
      <input v-model="password" type="password" :placeholder="t.password" autocomplete="current-password" />
      <button class="primary" :disabled="loading || password.length === 0">
        <Loader2 v-if="loading" class="spin" :size="16" />{{ t.login }}
      </button>
      <article v-if="notice.type === 'error'" class="notice error"><XCircle :size="17" /><span>{{ notice.title }}</span></article>
    </form>
  </main>

  <main v-else class="app-shell">
    <aside class="side">
      <div class="brand">
        <div class="logo">B</div>
        <strong>{{ config?.site.name || t.product }}</strong>
        <span>{{ config?.site.theme || 'PaperMod' }}</span>
      </div>
      <nav>
        <button :class="{ active: view === 'posts' || view === 'editor' || view === 'preview' }" @click="view = 'posts'"><FileText :size="18" />{{ t.posts }}<b>{{ posts.length }}</b></button>
        <button :class="{ active: view === 'home' }" @click="view = 'home'"><Home :size="18" />{{ t.homeSettings }}</button>
        <button :class="{ active: view === 'now' }" @click="view = 'now'; loadNowPage()"><Clock :size="18" />{{ t.nowPage }}</button>
        <button :class="{ active: view === 'settings' }" @click="view = 'settings'; loadAudit()"><Settings :size="18" />{{ t.settings }}</button>
      </nav>
      <div class="side-actions">
        <button @click="syncPosts()" :disabled="loading"><RefreshCw :class="{ spin: loading }" :size="17" />{{ t.sync }}</button>
        <button @click="logout"><LogOut :size="17" />{{ t.logout }}</button>
      </div>
    </aside>

    <section class="stage">
      <header class="topbar">
        <div>
          <p class="eyebrow">{{ config?.site.name || t.product }}</p>
          <h1>{{ pageTitle }}</h1>
        </div>
        <div class="actions">
          <button class="ghost" @click="setLang(lang === 'zh' ? 'en' : 'zh')"><Globe2 :size="16" />{{ lang === 'zh' ? 'EN' : '中文' }}</button>
          <button class="icon-button" @click="toggleTheme"><Moon v-if="!dark" :size="16" /><Sun v-else :size="16" /></button>
          <button v-if="view === 'editor'" @click="createPreview"><Eye :size="16" />{{ t.preview }}</button>
          <button v-if="view === 'editor'" @click="saveDraft"><Save :size="16" />{{ t.save }}</button>
          <button v-if="view === 'editor'" @click="publishWechat"><Newspaper :size="16" />{{ t.publishWechat }}</button>
          <button v-if="view === 'editor'" class="primary" @click="publishBlog"><Send :size="16" />{{ t.publishBlog }}</button>
          <button v-if="view === 'now'" class="primary" @click="saveNowPage"><Save :size="16" />{{ t.save }}</button>
        </div>
      </header>

      <section class="status-strip">
        <article :class="['notice', notice.type]">
          <CheckCircle2 v-if="notice.type === 'ok'" :size="18" />
          <AlertTriangle v-else-if="notice.type === 'warn'" :size="18" />
          <XCircle v-else :size="18" />
          <span>{{ loading ? busyText : notice.title }}</span>
          <Loader2 v-if="loading" class="spin" :size="16" />
        </article>
        <article v-if="health && health.status !== 'ok'" class="notice warn">
          <HeartPulse :size="18" />
          <span>{{ health.checks.find((item) => item.status !== 'ok')?.message }}</span>
        </article>
      </section>

      <section v-if="view === 'posts'" class="posts-page">
        <div class="stats">
          <article><span>{{ t.allPosts }}</span><strong>{{ posts.length }}</strong></article>
          <article><span>{{ t.drafts }}</span><strong>{{ draftCount }}</strong></article>
          <article><span>{{ t.tags }}</span><strong>{{ tagCount }}</strong></article>
        </div>
        <div class="new-row">
          <input v-model="newSlug" :placeholder="t.slugPlaceholder" />
          <button class="primary" @click="createPost"><Rocket :size="16" />{{ t.newPost }}</button>
        </div>
        <p v-if="posts.length === 0" class="empty">{{ t.emptyPosts }}</p>
        <button v-for="post in posts" :key="post.slug" class="post-card" @click="openPost(post.slug)">
          <span><strong>{{ post.title || post.slug }}</strong><em>{{ post.date || '-' }} · {{ post.draft ? t.draft : t.publishBlog }}</em></span>
          <i>{{ post.syncStatus }}</i>
        </button>
      </section>

      <section v-else-if="view === 'editor' && selected" class="editor-page">
        <input v-model="title" class="title-input" :placeholder="t.title" />
        <div class="meta-grid">
          <label>{{ t.shortLink }}<input v-model="selected.slug" disabled /></label>
          <label>{{ t.tagsLabel }}<input v-model="tags" placeholder="Hugo, AI" /></label>
          <label>{{ t.categories }}<input v-model="categories" placeholder="Blog" /></label>
          <label>{{ t.cover }}<input v-model="image" placeholder="cover.jpg" /></label>
          <label class="check"><input v-model="selected.frontMatter.draft" type="checkbox" />{{ t.draft }}</label>
          <label class="check"><input v-model="confirmOverwrite" type="checkbox" />{{ t.overwrite }}</label>
        </div>
        <p v-if="confirmOverwrite" class="inline-warning">{{ t.confirmOverwriteBody }}</p>
        <Codemirror v-model="body" :extensions="[markdown()]" class="editor" />
        <div class="danger-row">
          <button @click="rollback"><History :size="16" />{{ t.rollback }}</button>
        </div>
      </section>

      <section v-else-if="view === 'preview'" class="preview-page">
        <iframe v-if="preview?.url && !preview.error" :src="preview.url"></iframe>
        <pre v-else>{{ preview?.error?.message || t.noPreview }}</pre>
      </section>

      <section v-else-if="view === 'home'" class="home-settings-page">
        <div class="settings-layout">
          <article class="settings-card">
            <h2>{{ t.avatar }}</h2>
            <div class="avatar-area">
              <img v-if="siteInfo.profileImage" :src="siteInfo.profileImage" class="avatar-preview" alt="avatar" />
              <div v-else class="avatar-emoji-placeholder">🌿</div>
              <label>
                <input type="file" accept="image/jpeg,image/png,image/gif,image/webp" @change="handleAvatarUpload" style="display:none" />
                <span class="button" style="display:inline-flex;align-items:center;gap:8px;cursor:pointer;">{{ t.uploadAvatar }}</span>
              </label>
              <p class="hint">{{ t.avatarHint }}</p>
            </div>
          </article>

          <article class="settings-card">
            <h2>{{ t.tagline }}</h2>
            <textarea v-model="siteInfo.description" :placeholder="t.taglinePlaceholder" rows="3" class="tagline-input"></textarea>
            <button class="primary" @click="saveSiteInfo"><Save :size="16" />{{ t.save }}</button>
            <p class="hint">{{ t.avatarHint }}</p>
          </article>
        </div>
      </section>

      <section v-else-if="view === 'now'" class="now-page">
        <p class="now-hint">{{ t.nowHint }}</p>
        <Codemirror v-model="nowRaw" :extensions="[markdown()]" class="editor now-editor" />
      </section>

      <section v-else-if="view === 'settings' && config" class="settings-page">
        <div class="settings-layout">
          <article class="settings-card">
            <h2><BookOpen :size="18" />{{ t.basicInfo }}</h2>
            <label>{{ t.siteName }}<input v-model="config.site.name" /></label>
            <label>{{ t.adminPath }}<input v-model="config.basePath" /></label>
          </article>

          <article class="settings-card">
            <h2><Rocket :size="18" />{{ t.writing }}</h2>
            <label>{{ t.postsFolder }}<input v-model="config.site.postRoot" /></label>
            <label>{{ t.siteBuildMode }}
              <select v-model="config.site.buildCommand">
                <option value="hugo --minify">{{ t.buildCompressed }}</option>
                <option value="hugo">{{ t.buildStandard }}</option>
              </select>
            </label>
            <label>{{ t.previewRetention }}<input v-model.number="config.preview.ttlMinutes" type="number" min="10" /></label>
            <label>{{ t.uploadLimit }}<input v-model.number="uploadLimitMb" type="number" min="1" /></label>
          </article>

          <article class="settings-card">
            <h2><Shield :size="18" />{{ t.wechatSettings }}</h2>
            <label class="switch"><input v-model="config.wechat.enabled" type="checkbox" />{{ t.wechatEnabled }}</label>
            <label>{{ t.wechatApp }}<input v-model="config.wechat.appId" /></label>
            <p>{{ t.secretHint }}</p>
          </article>

          <article class="settings-card">
            <h2><Shield :size="18" />{{ t.security }}</h2>
            <label>{{ t.currentPassword }}<input v-model="currentPassword" type="password" autocomplete="current-password" /></label>
            <label>{{ t.newPassword }}<input v-model="newAdminPassword" type="password" autocomplete="new-password" /></label>
            <button :disabled="!currentPassword || newAdminPassword.length < 6" @click="changePassword">{{ t.changePassword }}</button>
          </article>
        </div>

        <div class="settings-actions">
          <button class="primary" @click="saveConfig"><Save :size="16" />{{ t.saveSettings }}</button>
          <button @click="runHealth"><HeartPulse :size="16" />{{ t.health }}</button>
          <button @click="showAdvanced = !showAdvanced"><ChevronDown :class="{ open: showAdvanced }" :size="16" />{{ showAdvanced ? t.closeAdvanced : t.openAdvanced }}</button>
        </div>
        <section v-if="showAdvanced" class="advanced-panel">
          <h2>{{ t.advanced }}</h2>
          <pre>{{ JSON.stringify({ config, health, audit }, null, 2) }}</pre>
        </section>
      </section>
    </section>
  </main>
</template>
