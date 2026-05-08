<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { SaveIcon } from 'lucide-vue-next'
import { EditorView, keymap } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useTheme } from '@/composables/useTheme'
import { useNotify } from '@/composables/useNotify'

const { t } = useI18n()
const { theme } = useTheme()
const notify = useNotify()

const loading = ref(false)
const saving = ref(false)
const editorEl = ref<HTMLElement | null>(null)
let view: EditorView | null = null

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
  } catch (e: any) {
    notify.error(e)
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => { view?.destroy() })

async function save() {
  if (!view) return
  const raw = view.state.doc.toString()
  saving.value = true
  try {
    await api.saveNowPage(raw)
    notify.success(t.value.saveNow)
  } catch (e: any) {
    notify.error(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="now-page">
    <div class="now-hint">{{ t.now }} · Markdown</div>
    <div ref="editorEl" class="now-editor" style="flex:1;overflow:auto;border:1px solid var(--border);border-radius:8px;margin:10px 0 0"></div>
    <div style="padding:10px 0 16px;display:flex;gap:8px">
      <button class="btn btn-primary" :disabled="saving || loading" @click="save">
        <SaveIcon :size="14" />{{ saving ? t.saving : t.saveNow }}
      </button>
    </div>
  </div>
</template>
