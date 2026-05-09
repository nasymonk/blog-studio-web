<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { SaveIcon, ClockIcon } from 'lucide-vue-next'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

const { t } = useI18n()
const { theme } = useTheme()
const notify = useNotify()

const loading = ref(false)
const saving = ref(false)
const editorEl = ref<HTMLElement | null>(null)
let view: EditorView | null = null

const wordCount = computed(() => {
  if (!view) return 0
  return view.state.doc.toString().replace(/\s/g, '').length
})

onMounted(async () => {
  loading.value = true
  try {
    const data = await api.getNowPage()
    if (editorEl.value) {
      view = new EditorView({
        state: EditorState.create({
          doc: data.raw,
          extensions: [
            markdown(),
            history(),
            EditorView.lineWrapping,
            keymap.of([...defaultKeymap, ...historyKeymap]),
            theme.value === 'dark' ? oneDark : [],
          ]
        }),
        parent: editorEl.value
      })
    }
  } catch (e: any) { notify.error(e) }
  finally { loading.value = false }
})

onBeforeUnmount(() => { view?.destroy() })

async function save() {
  if (!view) return
  const raw = view.state.doc.toString()
  saving.value = true
  try {
    await api.saveNowPage(raw)
    notify.success(t.value.saveNow)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}
</script>

<template>
  <div class="flex flex-col h-full max-w-3xl">
    <div class="flex items-center justify-between mb-3">
      <div>
        <h1 class="flex items-center gap-2 font-serif text-lg font-semibold">
          <ClockIcon class="h-4 w-4" /> {{ t.now }}
        </h1>
        <p class="text-xs text-muted-foreground mt-0.5">Markdown 格式的"现在"页面</p>
      </div>
      <span v-if="wordCount > 0" class="text-[11px] text-muted-foreground/60">{{ wordCount }} 字</span>
    </div>

    <div v-if="loading" class="flex-1 space-y-3 animate-fade-in">
      <Skeleton class="h-4 w-32" />
      <Skeleton class="flex-1 min-h-[400px] rounded" />
    </div>
    <div v-else class="flex flex-col flex-1 min-h-0">
      <div ref="editorEl" class="flex-1 overflow-auto rounded border border-border/60" />
      <div class="py-4 flex items-center justify-center gap-3">
        <span v-if="saving" class="text-[11px] text-muted-foreground/60 animate-pulse">保存中…</span>
        <Button class="rounded-full px-6" :disabled="saving || loading" @click="save">
          <SaveIcon class="h-4 w-4 mr-1.5" />{{ saving ? t.saving : t.saveNow }}
        </Button>
      </div>
    </div>
  </div>
</template>
