<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  SaveIcon, SendIcon, EyeIcon, Loader2Icon, AlertCircleIcon,
  BoldIcon, ItalicIcon, LinkIcon, ImageIcon, CodeIcon, Heading1Icon,
  RotateCcwIcon, ListIcon, ChevronDownIcon, ChevronUpIcon, PenLineIcon
} from 'lucide-vue-next'
import { api } from '@/services/api'
import type { PostDraft } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import { useEditor } from '@/composables/useEditor'
import EditorOutline from '@/components/EditorOutline.vue'
import KeybindingHelp from '@/components/KeybindingHelp.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'

const route = useRoute()
const router = useRouter()
const store = useStore()
const { t } = useI18n()
const { theme } = useTheme()
const notify = useNotify()

const slug = computed(() => decodeURIComponent(route.params.slug as string))
const loading = ref(true)
const publishing = ref(false)
const draft = ref<PostDraft | null>(null)
const showOutline = ref(true)
const metaExpanded = ref(false)
const keybindingHelpOpen = ref(false)

const editorContainer = ref<HTMLElement | null>(null)
const { body, dirty, saving, savedAt, wordCount, mode, mount,
        execBold, execItalic, execLink, execCode, execHeading, insertText, goToLine, toggleMode } = useEditor(
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
    body.value = draft.value.body
  }
  catch (e: any) {
    if (e.status === 404) {
      draft.value = {
        slug: slug.value,
        body: '',
        frontMatter: {
          title: '',
          date: new Date().toISOString().slice(0, 10),
          draft: true,
          tags: [],
          categories: [],
        },
        assets: [],
      }
      body.value = ''
    } else {
      notify.error(e, { onRetry: loadPost })
    }
  }
  finally { loading.value = false }
}

onMounted(async () => {
  await loadPost()
  if (editorContainer.value) mount(editorContainer.value)
})

watch(editorContainer, (el) => {
  if (el && !loading.value) mount(el)
})

async function saveDraft() {
  if (!draft.value) return
  draft.value.body = body.value
  try {
    await api.saveDraft(draft.value)
    store.editor.savedAt = new Date()
    store.editor.dirty = false
    notify.success(t.value.saved)
  } catch (e: any) { notify.error(e, { onRetry: saveDraft }) }
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
  } catch (e: any) { notify.error(e, { onRetry: () => publishBlog(confirmOverwrite) }) }
  finally { publishing.value = false }
}

function previewPost() {
  if (!draft.value) return
  draft.value.body = body.value
  router.push(`/posts/${encodeURIComponent(slug.value)}/preview`)
}

async function rollback() {
  if (!draft.value || !confirm(t.value.rollbackConfirm)) return
  try {
    await api.rollback(slug.value)
    notify.success(t.value.rollbackOk)
    await loadPost()
  } catch (e: any) { notify.error(e) }
}

function addTag(tag: string) {
  if (!draft.value || !tag || draft.value.frontMatter.tags.includes(tag)) return
  draft.value.frontMatter.tags.push(tag)
  store.editor.dirty = true
}
function removeTag(tag: string) {
  if (!draft.value) return
  draft.value.frontMatter.tags = draft.value.frontMatter.tags.filter(tg => tg !== tag)
  store.editor.dirty = true
}

function savedLabel(): string {
  if (saving.value) return t.value.saving
  if (store.editor.savedAt) {
    const secs = Math.round((Date.now() - store.editor.savedAt.getTime()) / 1000)
    if (secs < 5) return t.value.saved
    if (secs < 60) return `${t.value.autoSaved} ${t.value.agoSeconds(secs)}`
    return `${t.value.autoSaved} ${t.value.agoMinutes(Math.round(secs/60))}`
  }
  if (dirty.value) return t.value.unsaved
  return t.value.saved
}

const readingTime = computed(() => Math.max(1, Math.round(wordCount.value / 300)))

const metaSummary = computed(() => {
  if (!draft.value) return ''
  const parts: string[] = []
  if (draft.value.frontMatter.date) parts.push(draft.value.frontMatter.date)
  if (draft.value.frontMatter.tags.length) parts.push(t.value.tagCount(draft.value.frontMatter.tags.length))
  if (draft.value.frontMatter.draft) parts.push(t.value.draft)
  if (draft.value.frontMatter.math) parts.push(t.value.math)
  return parts.join(' · ')
})

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
    } catch (e: any) { notify.error(e) }
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
  <div class="flex flex-col h-full max-w-5xl" @keydown="e => { if (e.key === '?' && !e.ctrlKey && !e.metaKey && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) { e.preventDefault(); keybindingHelpOpen = true } }">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2Icon class="h-6 w-6 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="draft">
      <!-- Title -->
      <input
        v-model="draft.frontMatter.title"
        class="w-full bg-transparent border-0 outline-none font-serif text-2xl font-semibold placeholder:text-muted-foreground/40 py-3 px-0 focus:ring-0"
        :placeholder="t.title"
        @input="store.editor.dirty = true"
      />

      <!-- Meta strip: collapsed summary / expanded form -->
      <div class="border-y border-border/60">
        <!-- Collapsed view -->
        <button
          v-if="!metaExpanded"
          class="flex items-center gap-2 w-full py-2 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer text-left"
          @click="metaExpanded = true"
        >
          <span v-if="metaSummary">{{ metaSummary }}</span>
          <span v-else class="opacity-50">{{ t.editMeta }}</span>
          <ChevronDownIcon class="h-3 w-3 ml-auto opacity-40" />
        </button>

        <!-- Expanded view -->
        <div v-else class="py-3 space-y-3 animate-fade-in">
          <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            <div class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.date }}</Label>
              <Input v-model="draft.frontMatter.date" type="date" class="h-8 text-sm" @change="store.editor.dirty = true" />
            </div>
            <div class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.description }}</Label>
              <Input v-model="draft.frontMatter.description" class="h-8 text-sm" @input="store.editor.dirty = true" />
            </div>
            <div class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.image }}</Label>
              <Input v-model="draft.frontMatter.image" class="h-8 text-sm" @input="store.editor.dirty = true" />
            </div>
            <div class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.tags }}</Label>
              <div class="flex flex-wrap items-center gap-1 min-h-[32px] rounded border border-input bg-transparent px-2 py-1 text-sm">
                <Badge v-for="tag in draft.frontMatter.tags" :key="tag" variant="secondary" class="font-deco text-[11px] cursor-pointer gap-0.5" @click="removeTag(tag)">
                  # {{ tag }} ×
                </Badge>
                <input class="flex-1 min-w-[50px] bg-transparent outline-none text-sm placeholder:text-muted-foreground" :placeholder="t.addTag" @keydown="onTagKeydown" />
              </div>
            </div>
            <div class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">Slug</Label>
              <Input :model-value="draft.slug" disabled class="h-8 text-sm text-muted-foreground" />
            </div>
            <div class="flex items-end gap-3 pb-0.5">
              <label class="flex items-center gap-1.5 text-xs cursor-pointer text-muted-foreground">
                <Checkbox :checked="draft.frontMatter.draft" @update:checked="(v: boolean) => { draft!.frontMatter.draft = v; store.editor.dirty = true }" />
                <span>{{ t.draft }}</span>
              </label>
              <label class="flex items-center gap-1.5 text-xs cursor-pointer text-muted-foreground">
                <Checkbox :checked="draft.frontMatter.math" @update:checked="(v: boolean) => { draft!.frontMatter.math = v; store.editor.dirty = true }" />
                <span>{{ t.math }}</span>
              </label>
            </div>
          </div>
          <button class="text-[10px] text-muted-foreground/50 hover:text-muted-foreground transition-colors flex items-center gap-1 cursor-pointer" @click="metaExpanded = false">
            <ChevronUpIcon class="h-3 w-3" /> {{ t.collapse }}
          </button>
        </div>
      </div>

      <!-- Toolbar -->
      <div class="flex items-center gap-1 py-2">
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Bold (⌘B)" title="加粗 (⌘B)" @click="execBold"><BoldIcon class="h-3.5 w-3.5" /></Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Italic (⌘I)" title="斜体 (⌘I)" @click="execItalic"><ItalicIcon class="h-3.5 w-3.5" /></Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Link (⌘K)" title="链接 (⌘K)" @click="execLink"><LinkIcon class="h-3.5 w-3.5" /></Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Insert image" title="插入图片" @click="handleImageUpload"><ImageIcon class="h-3.5 w-3.5" /></Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Inline code" title="行内代码" @click="execCode"><CodeIcon class="h-3.5 w-3.5" /></Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-foreground" aria-label="Heading" title="标题" @click="execHeading"><Heading1Icon class="h-3.5 w-3.5" /></Button>
        <Separator orientation="vertical" class="mx-1 h-4" />
        <Button variant="ghost" size="icon" class="h-7 w-7" :class="mode === 'source' ? 'bg-accent/10 text-accent' : 'text-muted-foreground'" :aria-label="mode === 'wysiwyg' ? 'Switch to source mode' : 'Switch to WYSIWYG'" :title="mode === 'wysiwyg' ? '切换源码模式' : '切换所见即所得'" @click="toggleMode">
          <CodeIcon v-if="mode === 'wysiwyg'" class="h-3.5 w-3.5" />
          <PenLineIcon v-else class="h-3.5 w-3.5" />
        </Button>
        <Button variant="ghost" size="icon" class="h-7 w-7" :class="showOutline ? 'bg-accent/10 text-accent' : 'text-muted-foreground'" aria-label="Toggle outline" title="大纲" @click="showOutline = !showOutline">
          <ListIcon class="h-3.5 w-3.5" />
        </Button>
        <div class="flex-1" />
        <Separator orientation="vertical" class="mx-1 h-4" />
        <Button variant="ghost" size="sm" class="text-xs h-7 text-destructive/60 hover:text-destructive" :aria-label="t.rollback" :title="t.rollback" @click="rollback">
          <RotateCcwIcon class="h-3 w-3 mr-1" />{{ t.rollback }}
        </Button>
      </div>

      <!-- Editor body -->
      <div class="flex flex-1 min-h-0 overflow-hidden rounded border border-border/60">
        <div class="flex-1 overflow-auto p-6">
          <div ref="editorContainer" class="h-full" />
        </div>
        <EditorOutline v-if="showOutline" :body="body" :on-jump="goToLine" />
      </div>

      <!-- Status bar -->
      <div class="flex items-center gap-2 py-2 text-[11px] text-muted-foreground/70">
        <span>{{ wordCount }} {{ t.wordCount }}</span>
        <span class="text-border/40">·</span>
        <span>{{ readingTime }} {{ t.readingTime }}</span>
        <span class="text-border/40">·</span>
        <span :class="{ 'text-destructive': dirty, 'text-ok': !dirty }">{{ savedLabel() }}</span>
        <div class="flex-1" />
        <Button size="sm" variant="ghost" class="h-7 text-xs text-muted-foreground" :aria-label="t.saveDraft" :disabled="!dirty" @click="saveDraft">
          <SaveIcon class="h-3 w-3 mr-1" />{{ t.saveDraft }}
        </Button>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :aria-label="t.preview" :title="t.preview" @click="previewPost">
          <EyeIcon class="h-3 w-3" />
        </Button>
        <Button size="sm" class="h-7 text-xs rounded-full px-4" :aria-label="t.publish" :disabled="publishing" @click="publishBlog()">
          <Loader2Icon v-if="publishing" class="h-3 w-3 animate-spin mr-1" />
          <SendIcon v-else class="h-3 w-3 mr-1" />{{ t.publish }}
        </Button>
      </div>
    </template>

    <!-- Error -->
    <div v-else class="flex-1 flex flex-col items-center justify-center gap-3 animate-fade-up">
      <AlertCircleIcon class="h-10 w-10 text-destructive/40" />
      <p class="font-serif text-muted-foreground">{{ t.postLoadFailed }}</p>
      <Button class="rounded-full px-5" @click="loadPost">{{ t.retry }}</Button>
    </div>

    <KeybindingHelp v-model:open="keybindingHelpOpen" />
  </div>
</template>
