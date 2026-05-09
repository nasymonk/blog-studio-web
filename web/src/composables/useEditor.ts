import { ref, watch, onBeforeUnmount, type Ref } from 'vue'
import { countWords } from '@/utils/words'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState, Compartment } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history, indentWithTab } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { bracketMatching, indentOnInput } from '@codemirror/language'
import { GFM } from '@lezer/markdown'
import { oneDark } from '@codemirror/theme-one-dark'

// Lazy-loaded modules — resolved once on first use.
let searchModule: typeof import('@codemirror/search') | null = null
let wysiwygModule: typeof import('./useWysiwyg') | null = null

async function loadSearch() {
  if (!searchModule) searchModule = await import('@codemirror/search')
  return searchModule
}

async function loadWysiwygModule() {
  if (!wysiwygModule) wysiwygModule = await import('./useWysiwyg')
  return wysiwygModule
}

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
  const mode = ref<'wysiwyg' | 'source'>('wysiwyg')
  let view: EditorView | null = null
  const themeCompartment = new Compartment()
  const wysiwygCompartment = new Compartment()
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

  async function save() {
    if (saveTimer) { clearTimeout(saveTimer); saveTimer = null }
    saving.value = true
    try {
      await onSave(body.value)
      dirty.value = false
      savedAt.value = new Date()
    } finally {
      saving.value = false
    }
    return true
  }

  async function mount(el: HTMLElement) {
    // Lazy-load search and wysiwyg extensions on first mount.
    const [searchMod, wysiwygMod] = await Promise.all([loadSearch(), loadWysiwygModule()])

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
      doc: body.value || initialBody,
      extensions: [
        markdown({ extensions: [GFM, wysiwygMod.mathExtension()] }),
        history(),
        lineNumbers(),
        bracketMatching(),
        highlightActiveLine(),
        indentOnInput(),
        EditorView.lineWrapping,
        EditorView.contentAttributes.of({ spellcheck: 'false', autocorrect: 'off', autocapitalize: 'off' }),
        keymap.of([
          ...defaultKeymap,
          ...historyKeymap,
          ...searchMod.searchKeymap,
          indentWithTab,
          { key: 'Mod-s', run: () => { save(); return true } },
          { key: 'Mod-b', run: (v) => wrapSelection(v, '**', '**') },
          { key: 'Mod-i', run: (v) => wrapSelection(v, '*', '*') },
          { key: 'Mod-k', run: (v) => wrapSelection(v, '[', '](url)') },
        ]),
        themeCompartment.of(theme.value === 'dark' ? oneDark : []),
        wysiwygCompartment.of(wysiwygMod.wysiwyg()),
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

  function toggleMode() {
    if (!view) return
    const newMode = mode.value === 'wysiwyg' ? 'source' : 'wysiwyg'
    mode.value = newMode
    // wysiwygModule is already loaded by the time mount completes.
    view.dispatch({
      effects: wysiwygCompartment.reconfigure(newMode === 'wysiwyg' && wysiwygModule ? wysiwygModule.wysiwyg() : []),
    })
    view.focus()
  }

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

  return { body, dirty, saving, savedAt, wordCount, mode, mount, destroy, save, toggleMode, execBold, execItalic, execLink, execImage, execCode, execHeading, insertText, goToLine }
}
