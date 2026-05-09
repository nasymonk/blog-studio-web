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
import { Button } from '@/components/ui/button'

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
    <p class="font-deco text-sm text-muted-foreground mb-3">{{ t.now }} · Markdown</p>
    <div ref="editorEl" class="flex-1 overflow-auto rounded border border-accent/20" />
    <div class="py-4 flex justify-center">
      <Button class="rounded-full px-6" :disabled="saving || loading" @click="save">
        <SaveIcon class="h-4 w-4 mr-1.5" />{{ saving ? t.saving : t.saveNow }}
      </Button>
    </div>
  </div>
</template>
