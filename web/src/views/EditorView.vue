<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  SaveIcon, SendIcon, EyeIcon, Loader2Icon, AlertCircleIcon,
  RotateCcwIcon, ChevronDownIcon, ChevronUpIcon, SettingsIcon,
  HistoryIcon, MaximizeIcon,
} from 'lucide-vue-next'
import { api } from '@/services/api'
import type { PostDraft, PostStats } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import { useEditor } from '@/composables/useEditor'
import { useEditorSettings } from '@/composables/useEditorSettings'
import EditorToolbar from '@/components/EditorToolbar.vue'
import EditorOutline from '@/components/EditorOutline.vue'
import FindReplace from '@/components/FindReplace.vue'
import ImageGallery from '@/components/ImageGallery.vue'
import KeybindingHelp from '@/components/KeybindingHelp.vue'
import SplitView from '@/components/SplitView.vue'
import EditorStatusBar from '@/components/EditorStatusBar.vue'
import MarkdownPreview from '@/components/MarkdownPreview.vue'
import { useSplitView } from '@/composables/useSplitView'
import { useVersionHistory } from '@/composables/useVersionHistory'
import VersionHistory from '@/components/VersionHistory.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import SkeletonLines from '@/components/SkeletonLines.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuGroup,
  DropdownMenuLabel, DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'

const route = useRoute()
const router = useRouter()
const store = useStore()
const { t } = useI18n()
const { theme } = useTheme()
const notify = useNotify()
const { mode: splitMode, splitRatio, isDragging, setMode: setSplitMode, onDragStart } = useSplitView()

const slug = computed(() => decodeURIComponent(route.params.slug as string))
const loading = ref(true)
const publishing = ref(false)
const draft = ref<PostDraft | null>(null)
const showOutline = ref(true)
const metaExpanded = ref(false)
const keybindingHelpOpen = ref(false)
const postStats = ref<PostStats | null>(null)
const galleryOpen = ref(false)
const flashSave = ref(false)
const focusMode = ref(false)
function toggleFocusMode() { focusMode.value = !focusMode.value }

// Global Escape handler for focus mode (CodeMirror may intercept keydown)
function onGlobalEscape(e: KeyboardEvent) {
  if (e.key === 'Escape' && focusMode.value) {
    e.preventDefault()
    focusMode.value = false
  }
}
watch(focusMode, (active) => {
  if (active) {
    document.addEventListener('keydown', onGlobalEscape)
  } else {
    document.removeEventListener('keydown', onGlobalEscape)
  }
})

const { settings: editorSettings, applyPreset } = useEditorSettings()
const { versions, isOpen: historyOpen, saveVersion, loadHistory: loadVersionHistory, toggle: toggleHistory } = useVersionHistory()

const editorContainer = ref<HTMLElement | null>(null)
const { body, dirty, saving, savedAt, saveStatus, lastSavedTime, wordCount, cursorLine, cursorCol, charCount, lineCount, readingTime, headings, activeLine, mount,
        execBold, execItalic, execLink, execCode, execHeading, execStrikethrough, execBlockquote, execUnorderedList, execOrderedList, execTaskList, execTable, execHr, insertText, goToLine, applySettings, findReplace } = useEditor(
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
watch(savedAt, () => {
  flashSave.value = true
  setTimeout(() => { flashSave.value = false }, 600)
  saveVersion(body.value, wordCount.value)
})

async function loadPost() {
  loading.value = true
  try {
    draft.value = await api.post(slug.value)
    body.value = draft.value.body
    api.postStats(slug.value).then(s => { postStats.value = s }).catch(() => {})
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
  loadVersionHistory(slug.value)
  if (editorContainer.value) mount(editorContainer.value, editorSettings.value)
})

watch(editorContainer, (el) => {
  if (el && !loading.value) mount(el, editorSettings.value)
})

watch(editorSettings, (val) => {
  applySettings(val)
}, { deep: true })

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

const metaSummary = computed(() => {
  if (!draft.value) return ''
  const parts: string[] = []
  if (draft.value.frontMatter.date) parts.push(draft.value.frontMatter.date)
  if (draft.value.frontMatter.tags.length) parts.push(t.value.tagCount(draft.value.frontMatter.tags.length))
  if (draft.value.frontMatter.draft) parts.push(t.value.draft)
  if (draft.value.frontMatter.math) parts.push(t.value.math)
  if (draft.value.frontMatter.scheduledAt) parts.push(`${t.value.scheduledAt}: ${draft.value.frontMatter.scheduledAt}`)
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

function onScheduledAtInput(e: Event) {
  if (!draft.value) return
  const val = (e.target as HTMLInputElement).value
  draft.value.frontMatter.scheduledAt = val || undefined
  store.editor.dirty = true
}

function onRootKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && focusMode.value) {
    e.preventDefault()
    focusMode.value = false
    return
  }
  if (e.key === '?' && !e.ctrlKey && !e.metaKey && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
    e.preventDefault()
    keybindingHelpOpen.value = true
  }
}

function restoreVersion(restoreBody: string) {
  body.value = restoreBody
  dirty.value = true
  historyOpen.value = false
}
</script>

<template>
  <div class="flex flex-col h-full max-w-5xl" :class="{ 'is-focus-mode': focusMode }" @keydown="onRootKeydown">
    <!-- Loading -->
    <div v-if="loading" class="h-full">
      <SkeletonLines :lines="12" />
    </div>

    <template v-else-if="draft">
      <!-- Breadcrumb -->
      <Breadcrumb
        :items="[
          { label: t.posts, to: '/posts' },
          { label: draft.frontMatter.title || slug },
        ]" />

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
            <div v-if="draft.frontMatter.draft" class="grid gap-1">
              <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.scheduledAt }}</Label>
              <Input
                :model-value="draft.frontMatter.scheduledAt || ''"
                type="datetime-local"
                class="h-8 text-sm"
                @input="onScheduledAtInput"
              />
              <span class="text-[10px] text-muted-foreground/50">{{ t.scheduledAtHint }}</span>
            </div>
          </div>
          <button class="text-[10px] text-muted-foreground/50 hover:text-muted-foreground transition-colors flex items-center gap-1 cursor-pointer" @click="metaExpanded = false">
            <ChevronUpIcon class="h-3 w-3" /> {{ t.collapse }}
          </button>
        </div>
      </div>

      <!-- Toolbar -->
      <div class="flex items-center gap-1 py-2">
        <EditorToolbar
          :split-mode="splitMode"
          @bold="execBold"
          @italic="execItalic"
          @strikethrough="execStrikethrough"
          @code="execCode"
          @link="execLink"
          @heading1="() => execHeading(1)"
          @heading2="() => execHeading(2)"
          @blockquote="execBlockquote"
          @unordered-list="execUnorderedList"
          @ordered-list="execOrderedList"
          @task-list="execTaskList"
          @image="handleImageUpload"
          @table="execTable"
          @hr="execHr"
          @set-split-mode="setSplitMode"
          @gallery="galleryOpen = true"
          @keybindings="keybindingHelpOpen = true"
        />

        <!-- Editor settings dropdown -->
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" :title="t.editorSettings">
              <SettingsIcon class="h-3.5 w-3.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-64">
            <DropdownMenuLabel>{{ t.editorSettings }}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <!-- Preset buttons -->
              <div class="px-2 py-1.5">
                <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70 mb-1.5 block">{{ t.preset }}</Label>
                <div class="flex gap-1">
                  <Button variant="outline" size="sm" class="h-7 text-xs flex-1" @click="applyPreset('default')">{{ t.presetDefault }}</Button>
                  <Button variant="outline" size="sm" class="h-7 text-xs flex-1" @click="applyPreset('typewriter')">{{ t.presetTypewriter }}</Button>
                  <Button variant="outline" size="sm" class="h-7 text-xs flex-1" @click="applyPreset('compact')">{{ t.presetCompact }}</Button>
                </div>
              </div>

              <!-- Font family -->
              <div class="px-2 py-1.5">
                <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70 mb-1 block">{{ t.fontFamily }}</Label>
                <Select :model-value="editorSettings.fontFamily" @update:model-value="(v: any) => { if (v) editorSettings.fontFamily = v }">
                  <SelectTrigger class="h-8 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="system">{{ t.fontFamilySystem }}</SelectItem>
                    <SelectItem value="mono">{{ t.fontFamilyMono }}</SelectItem>
                    <SelectItem value="serif">{{ t.fontFamilySerif }}</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <!-- Font size -->
              <div class="px-2 py-1.5">
                <div class="flex items-center justify-between mb-1">
                  <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.fontSize }}</Label>
                  <span class="text-[10px] text-muted-foreground tabular-nums">{{ editorSettings.fontSize }}px</span>
                </div>
                <input
                  type="range"
                  :value="editorSettings.fontSize"
                  min="12" max="24" step="1"
                  class="w-full h-1.5 accent-primary cursor-pointer"
                  @input="editorSettings.fontSize = Number(($event.target as HTMLInputElement).value)"
                />
              </div>

              <!-- Line height -->
              <div class="px-2 py-1.5">
                <div class="flex items-center justify-between mb-1">
                  <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.lineHeight }}</Label>
                  <span class="text-[10px] text-muted-foreground tabular-nums">{{ editorSettings.lineHeight.toFixed(1) }}</span>
                </div>
                <input
                  type="range"
                  :value="editorSettings.lineHeight"
                  min="1.2" max="2.0" step="0.1"
                  class="w-full h-1.5 accent-primary cursor-pointer"
                  @input="editorSettings.lineHeight = Number(Number(($event.target as HTMLInputElement).value).toFixed(1))"
                />
              </div>

              <!-- Line numbers toggle -->
              <div class="px-2 py-1.5 flex items-center justify-between">
                <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.lineNumbers }}</Label>
                <button
                  class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors cursor-pointer"
                  :class="editorSettings.lineNumbers ? 'bg-primary' : 'bg-input'"
                  @click="editorSettings.lineNumbers = !editorSettings.lineNumbers"
                >
                  <span
                    class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow transition-transform"
                    :class="editorSettings.lineNumbers ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                  />
                </button>
              </div>

              <!-- Code theme -->
              <div class="px-2 py-1.5">
                <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70 mb-1 block">{{ t.codeTheme }}</Label>
                <Select :model-value="editorSettings.codeTheme" @update:model-value="(v: any) => { if (v) editorSettings.codeTheme = v }">
                  <SelectTrigger class="h-8 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="oneDark">{{ t.codeThemeOneDark }}</SelectItem>
                    <SelectItem value="githubLight">{{ t.codeThemeGithubLight }}</SelectItem>
                    <SelectItem value="catppuccin">{{ t.codeThemeCatppuccin }}</SelectItem>
                    <SelectItem value="solarized">{{ t.codeThemeSolarized }}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>

        <Separator orientation="vertical" class="h-5 mx-1" />

        <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="专注模式" @click="toggleFocusMode">
          <MaximizeIcon class="h-3.5 w-3.5" />
        </Button>

        <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" :aria-label="'版本历史'" :title="'版本历史'" @click="toggleHistory">
          <HistoryIcon class="h-3.5 w-3.5" />
        </Button>

        <Button variant="ghost" size="sm" class="text-xs h-7 text-destructive/60 hover:text-destructive" :aria-label="t.rollback" :title="t.rollback" @click="rollback">
          <RotateCcwIcon class="h-3 w-3 mr-1" />{{ t.rollback }}
        </Button>
      </div>

      <!-- Find & Replace -->
      <FindReplace
        v-if="findReplace.open.value"
        :search-text="findReplace.searchText.value"
        :replace-text="findReplace.replaceText.value"
        :case-sensitive="findReplace.caseSensitive.value"
        :whole-word="findReplace.wholeWord.value"
        :use-regex="findReplace.useRegex.value"
        :match-count="findReplace.matchCount.value"
        :current-match="findReplace.currentMatch.value"
        @update:search-text="findReplace.searchText.value = $event"
        @update:replace-text="findReplace.replaceText.value = $event"
        @update:case-sensitive="findReplace.caseSensitive.value = $event"
        @update:whole-word="findReplace.wholeWord.value = $event"
        @update:use-regex="findReplace.useRegex.value = $event"
        @close="findReplace.close()"
        @find-next="findReplace.findNext()"
        @find-prev="findReplace.findPrev()"
        @replace="findReplace.replaceMatch()"
        @replace-all="findReplace.replaceAllMatches()"
      />

      <!-- Focus mode hint bar -->
      <div v-if="focusMode" class="focus-hint-bar flex items-center justify-between px-4 py-2 shrink-0">
        <span class="text-xs text-muted-foreground">专注模式</span>
        <button class="text-xs text-muted-foreground hover:text-foreground transition-colors" @click="toggleFocusMode">
          按 Esc 退出
        </button>
      </div>

      <!-- Editor body -->
      <div class="flex flex-1 min-h-0 overflow-hidden rounded border border-border/60">
        <SplitView
          :mode="splitMode"
          :split-ratio="splitRatio"
          :is-dragging="isDragging"
          class="flex-1 min-w-0"
          @drag-start="onDragStart"
        >
          <template #source>
            <div class="h-full overflow-auto p-6">
              <div ref="editorContainer" class="h-full" />
            </div>
          </template>
          <template #preview>
            <div class="h-full overflow-auto p-6">
              <MarkdownPreview :source="body" />
            </div>
          </template>
        </SplitView>
        <EditorOutline v-if="showOutline" :headings="headings" :active-line="activeLine" class="hidden md:block" @jump="goToLine" />
      </div>

      <EditorStatusBar
        :word-count="wordCount"
        :char-count="charCount"
        :line-count="lineCount"
        :cursor-line="cursorLine"
        :cursor-col="cursorCol"
        :reading-time="readingTime"
      />

      <!-- Status bar -->
      <div class="flex items-center gap-2 py-2 text-[11px] text-muted-foreground/70 flex-wrap">
        <span class="hidden sm:inline">{{ wordCount }} {{ t.wordCount }}</span>
        <span class="hidden sm:inline text-border/40">·</span>
        <span class="hidden sm:inline">{{ readingTime }}</span>
        <span v-if="postStats" class="hidden sm:inline text-border/40">·</span>
        <span v-if="postStats" class="hidden sm:inline">{{ t.views }}: {{ postStats.views }}</span>
        <span class="hidden sm:inline text-border/40">·</span>
        <div class="flex items-center gap-2 text-xs" :class="{ 'save-flash': flashSave }">
          <template v-if="saveStatus === 'saving'">
            <Loader2Icon class="h-3 w-3 animate-spin text-muted-foreground" />
            <span class="text-muted-foreground">{{ t.saving }}</span>
          </template>
          <template v-else-if="saveStatus === 'unsaved'">
            <div class="h-2 w-2 rounded-full bg-warn animate-pulse" />
            <span class="text-warn">{{ t.unsaved }}</span>
          </template>
          <template v-else-if="saveStatus === 'saved'">
            <div class="h-2 w-2 rounded-full bg-ok" />
            <span class="text-muted-foreground">{{ t.savedAt }} {{ lastSavedTime }}</span>
          </template>
        </div>
        <div class="flex-1" />
        <Button size="sm" variant="ghost" class="h-8 md:h-7 text-xs text-muted-foreground" :aria-label="t.saveDraft" :disabled="!dirty" @click="saveDraft">
          <SaveIcon class="h-3 w-3 mr-1" />{{ t.saveDraft }}
        </Button>
        <Button variant="ghost" size="icon" class="h-8 w-8 md:h-7 md:w-7 text-muted-foreground" :aria-label="t.preview" :title="t.preview" @click="previewPost">
          <EyeIcon class="h-3 w-3" />
        </Button>
        <Button size="sm" class="h-8 md:h-7 text-xs rounded-full px-4" :aria-label="t.publish" :disabled="publishing" @click="publishBlog()">
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

    <VersionHistory
      :versions="versions"
      :open="historyOpen"
      @close="historyOpen = false"
      @restore="restoreVersion"
    />
    <KeybindingHelp v-model:open="keybindingHelpOpen" />
    <ImageGallery
      v-if="draft"
      v-model:open="galleryOpen"
      :slug="slug"
      :assets="draft.assets"
      :on-insert="insertText"
      @uploaded="loadPost"
    />

  </div>
</template>

<style scoped>
.save-flash {
  animation: save-flash 0.6s ease-out;
}
@keyframes save-flash {
  0% { background-color: oklch(from var(--ok) l c h / 0.2); }
  100% { background-color: transparent; }
}

/* Focus mode */
.is-focus-mode {
  position: fixed !important;
  inset: 0 !important;
  z-index: 50 !important;
  max-width: none !important;
  background: var(--color-background);
  padding: 0 !important;
}

.is-focus-mode > *:not(.flex-1):not(.focus-hint-bar) {
  display: none !important;
}

.is-focus-mode > .flex-1 {
  border: none !important;
  border-radius: 0 !important;
}

.is-focus-mode .cm-editor {
  max-width: 700px;
  margin: 0 auto;
}

.is-focus-mode .cm-editor .cm-scroller {
  padding: 2rem 1rem;
}
</style>
