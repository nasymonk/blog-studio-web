<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AlertTriangle, BookOpen, Eye, FileText, HeartPulse, History, Loader2, LogOut, MessageSquare, Newspaper, RefreshCw, Rocket, Save, Send, Settings } from 'lucide-vue-next'
import { markdown } from '@codemirror/lang-markdown'
import Codemirror from 'vue-codemirror6'
import { api, type AppError, type Config, type PostDraft, type PostState, type PreviewResult, type PublishResult } from './services/api'

const authed = ref(false)
const password = ref('')
const loading = ref(false)
const view = ref<'posts' | 'editor' | 'preview' | 'comments' | 'settings'>('posts')
const posts = ref<PostState[]>([])
const selected = ref<PostDraft | null>(null)
const config = ref<Config | null>(null)
const preview = ref<PreviewResult | null>(null)
const comments = ref<Array<{ id: string; url: string; nick: string; comment: string; created: string }>>([])
const commentAdminUrl = ref('/comment/admin/')
const health = ref<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null>(null)
const audit = ref<unknown[]>([])
const log = ref('等待操作。')
const confirmOverwrite = ref(false)
const newSlug = ref('')

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
  try {
    const session = await api.session()
    authed.value = session.authenticated
    if (session.csrfToken) api.setCSRF(session.csrfToken)
    if (authed.value) await bootstrap()
  } catch (error) {
    log.value = formatError(error)
  }
})

async function login() {
  loading.value = true
  try {
    const result = await api.login(password.value)
    api.setCSRF(result.csrfToken)
    authed.value = true
    password.value = ''
    await bootstrap()
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function logout() {
  await api.logout()
  authed.value = false
  selected.value = null
}

async function bootstrap() {
  config.value = await api.config()
  await Promise.all([syncPosts(), runHealth(), loadComments()])
}

async function syncPosts() {
  loading.value = true
  log.value = '正在同步文章列表...'
  try {
    posts.value = await api.posts()
    view.value = 'posts'
    log.value = `同步完成，共 ${posts.value.length} 篇文章。`
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function openPost(slug: string) {
  loading.value = true
  try {
    selected.value = await api.post(slug)
    view.value = 'editor'
    log.value = `已加载 ${slug}。`
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

function createPost() {
  const slug = newSlug.value.trim() || `new-post-${Date.now()}`
  selected.value = {
    slug,
    frontMatter: { title: '未命名文章', draft: true, tags: [], categories: [] },
    body: '## 背景\n\n\n## 正文\n\n\n## 总结\n',
    assets: []
  }
  newSlug.value = ''
  view.value = 'editor'
}

async function saveDraft() {
  if (!selected.value) return
  try {
    await api.saveDraft(selected.value)
    log.value = '草稿已保存。'
  } catch (error) {
    log.value = formatError(error)
  }
}

async function publishBlog() {
  if (!selected.value) return
  loading.value = true
  try {
    const result = await api.publishBlog(selected.value, confirmOverwrite.value)
    showResult(result)
    if (result.status === 'success') await syncPosts()
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function publishWechat() {
  if (!selected.value) return
  loading.value = true
  try {
    showResult(await api.publishWechat(selected.value))
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function createPreview() {
  if (!selected.value) return
  loading.value = true
  try {
    preview.value = await api.preview(selected.value)
    view.value = 'preview'
    log.value = preview.value.error ? formatError(preview.value.error) : '预览已生成。'
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function rollback() {
  if (!selected.value) return
  loading.value = true
  try {
    showResult(await api.rollback(selected.value.slug))
    await syncPosts()
  } catch (error) {
    log.value = formatError(error)
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  try {
    const result = await api.commentsRecent()
    comments.value = result.items
    commentAdminUrl.value = result.adminUrl
  } catch {
    comments.value = []
  }
}

async function runHealth() {
  try {
    health.value = await api.health()
  } catch (error) {
    log.value = formatError(error)
  }
}

async function loadAudit() {
  audit.value = await api.audit()
}

async function saveConfig() {
  if (!config.value) return
  try {
    config.value = await api.saveConfig(config.value)
    log.value = '配置已保存，重启服务后所有运行时组件会使用新配置。'
  } catch (error) {
    log.value = formatError(error)
  }
}

function showResult(result: PublishResult) {
  if (result.error) {
    log.value = formatError(result.error)
    return
  }
  log.value = [`${result.target} ${result.status}`, `audit: ${result.auditId}`, `backup: ${result.backupId || '-'}`].join('\n')
}

function splitList(value: string) {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}

function formatError(error: unknown) {
  const app = error as AppError
  if (app?.message) return [app.message, app.technicalDetail, app.suggestion].filter(Boolean).join('\n')
  if (error instanceof Error) return error.message
  return String(error)
}
</script>

<template>
  <main v-if="!authed" class="login-page">
    <form class="login-card" @submit.prevent="login">
      <div class="logo">B</div>
      <h1>Blog Studio</h1>
      <p>登录后管理 Hugo 博客、预览、发布和评论入口。</p>
      <input v-model="password" type="password" placeholder="管理员密码" autocomplete="current-password" />
      <button class="primary" :disabled="loading || password.length === 0">
        <Loader2 v-if="loading" class="spin" :size="16" />登录
      </button>
      <pre v-if="log !== '等待操作。'">{{ log }}</pre>
    </form>
  </main>

  <main v-else class="app-shell">
    <aside class="side">
      <div class="brand"><div class="logo">B</div><strong>{{ config?.site.name || 'Blog Studio' }}</strong></div>
      <nav>
        <button :class="{ active: view === 'posts' }" @click="view = 'posts'"><FileText :size="18" />文章 <span>{{ posts.length }}</span></button>
        <button :class="{ active: view === 'comments' }" @click="view = 'comments'; loadComments()"><MessageSquare :size="18" />评论</button>
        <button :class="{ active: view === 'settings' }" @click="view = 'settings'; loadAudit()"><Settings :size="18" />设置</button>
      </nav>
      <div class="side-actions">
        <button @click="syncPosts" :disabled="loading"><RefreshCw :class="{ spin: loading }" :size="17" />同步</button>
        <button @click="logout"><LogOut :size="17" />退出</button>
      </div>
    </aside>

    <section class="stage">
      <header class="topbar">
        <div>
          <h1>{{ view === 'editor' ? title : view === 'preview' ? '博客预览' : view === 'comments' ? '评论中心' : view === 'settings' ? '系统设置' : '文章管理' }}</h1>
          <p>{{ config?.site.postRoot || '/blog/content/post' }}</p>
        </div>
        <div class="actions">
          <button v-if="view === 'editor'" @click="createPreview"><Eye :size="16" />预览</button>
          <button v-if="view === 'editor'" @click="saveDraft"><Save :size="16" />保存</button>
          <button v-if="view === 'editor'" @click="publishWechat"><Newspaper :size="16" />公众号草稿</button>
          <button v-if="view === 'editor'" class="primary" @click="publishBlog"><Send :size="16" />发布博客</button>
        </div>
      </header>

      <section v-if="health && health.status !== 'ok'" class="health">
        <AlertTriangle :size="18" />
        <span>{{ health.checks.find((item) => item.status !== 'ok')?.message }}</span>
      </section>

      <section v-if="view === 'posts'" class="posts-page">
        <div class="stats">
          <div><span>全部文章</span><strong>{{ posts.length }}</strong></div>
          <div><span>草稿</span><strong>{{ posts.filter((post) => post.draft).length }}</strong></div>
          <div><span>标签</span><strong>{{ new Set(posts.flatMap((post) => post.tags)).size }}</strong></div>
        </div>
        <div class="new-row">
          <input v-model="newSlug" placeholder="new-post-slug" />
          <button @click="createPost">新建文章</button>
        </div>
        <button v-for="post in posts" :key="post.slug" class="post-card" @click="openPost(post.slug)">
          <span><strong>{{ post.title || post.slug }}</strong><em>{{ post.date || '未设置日期' }} · {{ post.draft ? '草稿' : '已发布' }}</em></span>
          <i>{{ post.syncStatus }}</i>
        </button>
      </section>

      <section v-else-if="view === 'editor' && selected" class="editor-page">
        <input v-model="title" class="title-input" placeholder="文章标题" />
        <div class="meta-grid">
          <label>Slug<input v-model="selected.slug" disabled /></label>
          <label>Tags<input v-model="tags" placeholder="Hugo, 笔记" /></label>
          <label>Categories<input v-model="categories" placeholder="Blog" /></label>
          <label>Cover<input v-model="image" placeholder="cover.jpg" /></label>
          <label class="check"><input v-model="selected.frontMatter.draft" type="checkbox" />草稿</label>
          <label class="check"><input v-model="confirmOverwrite" type="checkbox" />冲突覆盖</label>
        </div>
        <Codemirror v-model="body" :extensions="[markdown()]" class="editor" />
        <div class="danger-row">
          <button @click="rollback"><History :size="16" />回滚到上一版本</button>
        </div>
      </section>

      <section v-else-if="view === 'preview'" class="preview-page">
        <iframe v-if="preview?.url && !preview.error" :src="preview.url"></iframe>
        <pre v-else>{{ preview?.error?.message || '暂无预览' }}</pre>
      </section>

      <section v-else-if="view === 'comments'" class="comments-page">
        <a class="primary link-button" :href="commentAdminUrl" target="_blank">打开 Twikoo 管理面板</a>
        <article v-for="item in comments" :key="item.id || item.created" class="comment-card">
          <strong>{{ item.nick || '匿名' }}</strong>
          <span>{{ item.url }} · {{ item.created }}</span>
          <p>{{ item.comment }}</p>
        </article>
      </section>

      <section v-else-if="view === 'settings' && config" class="settings-page">
        <div class="settings-grid">
          <label>站点名称<input v-model="config.site.name" /></label>
          <label>Blog Root<input v-model="config.site.blogRoot" /></label>
          <label>Post Root<input v-model="config.site.postRoot" /></label>
          <label>Base Path<input v-model="config.basePath" /></label>
          <label>Twikoo Admin URL<input v-model="config.comments.adminUrl" /></label>
          <label>Preview TTL<input v-model.number="config.preview.ttlMinutes" type="number" /></label>
        </div>
        <button class="primary" @click="saveConfig">保存配置</button>
        <button @click="runHealth"><HeartPulse :size="16" />健康检查</button>
        <section class="audit-list"><pre>{{ JSON.stringify(audit, null, 2) }}</pre></section>
      </section>

      <footer><pre>{{ log }}</pre></footer>
    </section>
  </main>
</template>
