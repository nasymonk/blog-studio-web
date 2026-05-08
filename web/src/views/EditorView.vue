<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  SaveIcon, SendIcon, EyeIcon, SplitSquareHorizontalIcon, Loader2Icon,
  BoldIcon, ItalicIcon, LinkIcon, ImageIcon, CodeIcon, Heading1Icon,
  RotateCcwIcon, ListIcon
} from 'lucide-vue-next'
import { api } from '@/services/api'
import type { PostDraft } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import { useEditor } from '@/composables/useEditor'
import MarkdownPreview from '@/components/MarkdownPreview.vue'
import EditorOutline from '@/components/EditorOutline.vue'

const route = useRoute()
const router = useRouter()
const store = useStore()
const { t } = useI18n()
const { theme } = useTheme()
const notify = useNotify()

const slug = computed(() => decodeURIComponent(route.params.slug as string))
const loading = ref(true)
const publishing = ref(false)
const showPreview = ref(false)
const draft = ref<PostDraft | null>(null)

const showOutline = ref(true)

const editorContainer = ref<HTMLElement | null>(null)
const { body, dirty, saving, savedAt, wordCount, mount,
        execBold, execItalic, execLink, execCode, execHeading, insertText, goToLine } = useEditor(
  editorContainer,
  '',
  theme,
  async (newBody) => {
    if (!draft.value) return
    draft.value.body = newBody
    await api.saveDraft(draft.value)
    store.editor.savedAt = new Date()
    store.editor.dirty = false
  },
  async (file) => {
    const result = await api.uploadAsset(slug.value, file)
    return result.name
  }
)

watch(dirty, (val) => { store.editor.dirty = val })
watch(saving, (val) => { store.editor.saving = val })
watch(savedAt, (val) => { store.editor.savedAt = val })

async function loadPost() {
  loading.value = true
  try {
    draft.value = await api.post(slug.value)
  } catch (e: any) {
    notify.error(e, { onRetry: loadPost })
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadPost()
  if (editorContainer.value && draft.value) {
    mount(editorContainer.value)
  }
})

watch(editorContainer, (el) => {
  if (el && draft.value && !loading.value) mount(el)
})

async function saveDraft() {
  if (!draft.value) return
  draft.value.body = body.value
  try {
    await api.saveDraft(draft.value)
    store.editor.savedAt = new Date()
    store.editor.dirty = false
    notify.success(t.value.saved)
  } catch (e: any) {
    notify.error(e, { onRetry: saveDraft })
  }
}

async function publishBlog(confirmOverwrite = false) {
  if (!draft.value) return
  draft.value.body = body.value
  publishing.value = true
  try {
    const result = await api.publishBlog(draft.value, confirmOverwrite)
    if (result.status === 'conflict') {
      if (confirm(t.value.confirmOverwrite)) publishBlog(true)
      return
    }
    if (result.error) { notify.error(result.error); return }
    notify.success(t.value.publishOk)
    store.editor.dirty = false
    await api.posts().then(p => { store.posts = p })
  } catch (e: any) {
    notify.error(e, { onRetry: () => publishBlog(confirmOverwrite) })
  } finally {
    publishing.value = false
  }
}

async function previewPost() {
  if (!draft.value) return
  draft.value.body = body.value
  try {
    const result = await api.preview(draft.value)
    if (result.error) { notify.error(result.error); return }
    router.push(`/posts/${encodeURIComponent(slug.value)}/preview`)
  } catch (e: any) {
    notify.error(e, { onRetry: previewPost })
  }
}

async function rollback() {
  if (!draft.value || !confirm('确认回滚到上一个版本？此操作不可撤销。')) return
  try {
    await api.rollback(slug.value)
    notify.success(t.value.rollbackOk)
    await loadPost()
  } catch (e: any) {
    notify.error(e)
  }
}

function addTag(tag: string) {
  if (!draft.value || !tag || draft.value.frontMatter.tags.includes(tag)) return
  draft.value.frontMatter.tags.push(tag)
  store.editor.dirty = true
}
function removeTag(tag: string) {
  if (!draft.value) return
  draft.value.frontMatter.tags = draft.value.frontMatter.tags.filter(t => t !== tag)
  store.editor.dirty = true
}

function savedLabel(): string {
  if (saving.value) return t.value.saving
  if (store.editor.savedAt) {
    const secs = Math.round((Date.now() - store.editor.savedAt.getTime()) / 1000)
    if (secs < 5) return t.value.saved
    if (secs < 60) return `${t.value.autoSaved} ${secs}s 前`
    return `${t.value.autoSaved} ${Math.round(secs/60)}m 前`
  }
  if (dirty.value) return t.value.unsaved
  return t.value.saved
}

const readingTime = computed(() => Math.max(1, Math.round(wordCount.value / 300)))

function handleImageUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    try {
      const { name } = await api.uploadAsset(slug.value, file)
      insertText(`![](${name})`)
    } catch (e: any) {
      notify.error(e)
    }
  }
  input.click()

}

function onTagKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ',') {
    e.preventDefault()
    const input = e.target as HTMLInputElement
    addTag(input.value)
    input.value = ''
  }
}
</script>

<template>
  <div class="editor-page">
    <!-- Loading state -->
    <div v-if="loading" class="editor-loading">
      <Loader2Icon :size="28" class="spin" style="color:var(--tertiary)" />
    </div>

    <template v-else-if="draft">
      <!-- Title input -->
      <input v-model="draft.frontMatter.title" class="title-input" type="text" :placeholder="t.title" @input="store.editor.dirty = true" />

      <!-- Meta strip -->
      <div class="editor-meta">
        <div class="field">
          <label class="field-label">{{ t.date }}</label>
          <input v-model="draft.frontMatter.date" class="input" type="date" @change="store.editor.dirty = true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.description }}</label>
          <input v-model="draft.frontMatter.description" class="input" @input="store.editor.dirty = true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.image }}</label>
          <input v-model="draft.frontMatter.image" class="input" @input="store.editor.dirty = true" />
        </div>
        <div class="field">
          <label class="field-label">{{ t.tags }}</label>
          <div class="tag-input-box">
            <span v-for="tag in draft.frontMatter.tags" :key="tag" class="tag-chip tag-chip-removable" @click="removeTag(tag)">
              # {{ tag }} ×
            </span>
            <input class="tag-input-field" placeholder="添加…" @keydown="onTagKeydown" />
          </div>
        </div>
        <div class="field">
          <label class="field-label">Slug</label>
          <input class="input" :value="draft.slug" disabled style="color:var(--tertiary)" />
        </div>
        <div class="field">
          <label class="field-label">&nbsp;</label>
          <div class="flex items-center gap-3" style="min-height:38px">
            <label class="check-field">
              <input v-model="draft.frontMatter.draft" type="checkbox" @change="store.editor.dirty = true" />
              <span class="field-label" style="margin:0">{{ t.draft }}</span>
            </label>
            <label class="check-field">
              <input v-model="draft.frontMatter.math" type="checkbox" @change="store.editor.dirty = true" />
              <span class="field-label" style="margin:0">{{ t.math }}</span>
            </label>
          </div>
        </div>
      </div>

      <!-- Markdown toolbar -->
      <div class="editor-toolbar">
        <button class="btn btn-ghost btn-icon btn-sm" :title="'加粗 (⌘B)'" @click="execBold"><BoldIcon :size="14" /></button>
        <button class="btn btn-ghost btn-icon btn-sm" :title="'斜体 (⌘I)'" @click="execItalic"><ItalicIcon :size="14" /></button>
        <button class="btn btn-ghost btn-icon btn-sm" :title="'链接 (⌘K)'" @click="execLink"><LinkIcon :size="14" /></button>
        <button class="btn btn-ghost btn-icon btn-sm" :title="'插入图片'" @click="handleImageUpload"><ImageIcon :size="14" /></button>
        <button class="btn btn-ghost btn-icon btn-sm" :title="'行内代码'" @click="execCode"><CodeIcon :size="14" /></button>
        <button class="btn btn-ghost btn-icon btn-sm" :title="'标题'" @click="execHeading"><Heading1Icon :size="14" /></button>
        <span class="toolbar-sep"></span>
        <button
class="btn btn-ghost btn-icon btn-sm" :title="t.splitPreview" :style="{ background: showPreview ? 'var(--hover)' : undefined }"
          @click="showPreview = !showPreview">
          <SplitSquareHorizontalIcon :size="14" />
        </button>
        <button
          class="btn btn-ghost btn-icon btn-sm"
          title="大纲"
          :style="{ background: showOutline ? 'var(--hover)' : undefined }"
          @click="showOutline = !showOutline"
        >
          <ListIcon :size="14" />
        </button>
        <span class="toolbar-spacer"></span>
        <button class="btn btn-ghost btn-sm btn-danger-text" :title="t.rollback" @click="rollback">
          <RotateCcwIcon :size="13" />{{ t.rollback }}
        </button>
      </div>

      <!-- Editor + preview body -->
      <div class="editor-body">
        <div class="editor-pane">
          <div ref="editorContainer" class="cm-wrap"></div>
        </div>
        <div v-if="showPreview" class="editor-pane preview-pane">
          <MarkdownPreview :source="body" />
        </div>
        <EditorOutline v-if="showOutline" :body="body" :on-jump="goToLine" />
      </div>

      <!-- Status bar -->
      <div class="editor-statusbar">
        <span class="statusbar-stat">{{ wordCount }} {{ t.wordCount }}</span>
        <span class="sep">·</span>
        <span class="statusbar-stat">{{ readingTime }} {{ t.readingTime }}</span>
        <span class="sep">·</span>
        <span class="statusbar-save" :class="{ 'save-dirty': dirty, 'save-ok': !dirty }">{{ savedLabel() }}</span>
        <span style="margin-left:auto"></span>
        <button class="btn btn-sm" :disabled="!dirty" @click="saveDraft">
          <SaveIcon :size="13" />{{ t.saveDraft }}
        </button>
        <button class="btn btn-ghost btn-sm" :title="t.preview" @click="previewPost">
          <EyeIcon :size="13" />
        </button>
        <button class="btn btn-primary btn-sm" :disabled="publishing" @click="publishBlog()">
          <Loader2Icon v-if="publishing" :size="13" class="spin" />
          <SendIcon v-else :size="13" />{{ t.publish }}
        </button>
      </div>
    </template>

    <!-- Error / not found state -->
    <div v-else class="empty-state" style="margin:48px auto">
      <p>文章加载失败</p>
      <button class="btn btn-primary" @click="loadPost">{{ t.retry }}</button>
    </div>
  </div>
</template>
