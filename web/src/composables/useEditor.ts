import { ref, watch, onBeforeUnmount, type Ref } from 'vue'
import { countWords } from '@/utils/words'
import { EditorView, keymap, lineNumbers } from '@codemirror/view'
import { EditorState, Compartment } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history, indentWithTab } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'

export type SaveFn = (body: string) => Promise<void>
export type PasteImageFn = (file: File) => Promise<string>

function wrapSelection(view: EditorView, before: string, after: string) {
  const { from, to } = view.state.selection.main
  const sel = view.state.sliceDoc(from, to)
  view.dispatch(view.state.update({
    changes: { from, to, insert: before + sel + after },
    selection: { anchor: from + before.length, head: from + before.length + sel.length }
  }))
  return true
}

export function useEditor(
  containerRef: Ref<HTMLElement | null>,
  initialBody: string,
  theme: Ref<'light' | 'dark'>,
  onSave: SaveFn,
  onPasteImage?: PasteImageFn
) {
  const body = ref(initialBody)
  const dirty = ref(false)
  const saving = ref(false)
  const savedAt = ref<Date | null>(null)
  const wordCount = ref(0)
  let view: EditorView | null = null
  const themeCompartment = new Compartment()
  let saveTimer: ReturnType<typeof setTimeout> | null = null

  function scheduleAutoSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(async () => {
      if (!dirty.value) return
      saving.value = true
      try {
        await onSave(body.value)
        dirty.value = false
        savedAt.value = new Date()
      } finally {
        saving.value = false
      }
    }, 1500)
  }

  function save() {
    if (saveTimer) { clearTimeout(saveTimer); saveTimer = null }
    onSave(body.value).then(() => {
      dirty.value = false
      savedAt.value = new Date()
    })
    return true
  }

  function mount(el: HTMLElement) {
    const updateListener = EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        body.value = update.state.doc.toString()
        wordCount.value = countWords(body.value)
        dirty.value = true
        scheduleAutoSave()
      }
    })

    const pasteHandler = EditorView.domEventHandlers({
      paste(event, editorView) {
        if (!onPasteImage) return false
        const items = event.clipboardData?.items
        if (!items) return false
        for (const item of Array.from(items)) {
          if (item.type.startsWith('image/')) {
            const file = item.getAsFile()
            if (!file) continue
            event.preventDefault()
            onPasteImage(file).then((filename) => {
              const insert = `![](${filename})`
              const pos = editorView.state.selection.main.head
              editorView.dispatch(editorView.state.update({
                changes: { from: pos, insert },
                selection: { anchor: pos + insert.length }
              }))
            })
            return true
          }
        }
        return false
      }
    })

    const state = EditorState.create({
      doc: initialBody,
      extensions: [
        markdown(),
        history(),
        lineNumbers(),
        EditorView.lineWrapping,
        EditorView.contentAttributes.of({ spellcheck: 'false', autocorrect: 'off', autocapitalize: 'off' }),
        keymap.of([
          ...defaultKeymap,
          ...historyKeymap,
          indentWithTab,
          { key: 'Mod-s', run: () => { save(); return true } },
          { key: 'Mod-b', run: (v) => wrapSelection(v, '**', '**') },
          { key: 'Mod-i', run: (v) => wrapSelection(v, '*', '*') },
          { key: 'Mod-k', run: (v) => wrapSelection(v, '[', '](url)') },
        ]),
        themeCompartment.of(theme.value === 'dark' ? oneDark : []),
        updateListener,
        pasteHandler,
      ]
    })

    view = new EditorView({ state, parent: el })
    wordCount.value = countWords(initialBody)
  }

  watch(theme, (val) => {
    if (view) {
      view.dispatch({ effects: themeCompartment.reconfigure(val === 'dark' ? oneDark : []) })
    }
  })

  function destroy() {
    if (saveTimer) { clearTimeout(saveTimer); saveTimer = null }
    view?.destroy()
    view = null
  }

  const beforeUnloadHandler = (e: BeforeUnloadEvent) => {
    if (dirty.value) {
      e.preventDefault()
    }
  }
  window.addEventListener('beforeunload', beforeUnloadHandler)
  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', beforeUnloadHandler)
    destroy()
  })

  function execBold() { if (view) wrapSelection(view, '**', '**') }
  function execItalic() { if (view) wrapSelection(view, '*', '*') }
  function execLink() { if (view) wrapSelection(view, '[', '](url)') }
  function execImage() { if (view) wrapSelection(view, '![', '](url)') }
  function execCode() { if (view) wrapSelection(view, '`', '`') }
  function execHeading() {
    if (!view) return
    const { from } = view.state.selection.main
    const line = view.state.doc.lineAt(from)
    const prefix = line.text.startsWith('## ') ? '' : line.text.startsWith('# ') ? '## ' : '# '
    const cleanText = line.text.replace(/^#+\s*/, '')
    view.dispatch(view.state.update({ changes: { from: line.from, to: line.to, insert: prefix + cleanText } }))
  }

  function insertText(text: string) {
    if (!view) return
    const pos = view.state.selection.main.head
    view.dispatch(view.state.update({
      changes: { from: pos, insert: text },
      selection: { anchor: pos + text.length }
    }))
  }

  function goToLine(lineNumber: number) {
    if (!view) return
    const doc = view.state.doc
    if (lineNumber < 1 || lineNumber > doc.lines) return
    const line = doc.line(lineNumber)
    view.dispatch({
      selection: { anchor: line.from },
      effects: EditorView.scrollIntoView(line.from, { y: 'center' }),
    })
    view.focus()
  }

  return { body, dirty, saving, savedAt, wordCount, mount, destroy, save, execBold, execItalic, execLink, execImage, execCode, execHeading, insertText, goToLine }
}
