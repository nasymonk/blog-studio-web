import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { countWords } from '../utils/words'

// CodeMirror requires a real DOM with layout capabilities (getBoundingClientRect, etc.)
// that happy-dom does not provide. We test the composable's exported shape by importing
// the module and verifying it exports the expected function, and test the pure utility
// functions (countWords, wysiwyg, mathExtension) directly.

// ── countWords (also tested in words.test.ts; these cover editor-specific edge cases) ──

describe('countWords (editor context)', () => {
  it('counts a single Chinese character', () => {
    expect(countWords('你')).toBe(1)
  })

  it('counts a single English word', () => {
    expect(countWords('hello')).toBe(1)
  })

  it('counts mixed inline Chinese and English without spaces', () => {
    // 你好 = 2 Chinese chars, world = 1 English word
    expect(countWords('你好world')).toBe(3)
  })

  it('counts markdown-formatted text', () => {
    // ** = punctuation (0), bold = 1 word, 你好 = 2 chars
    expect(countWords('**bold** 你好')).toBe(1 + 2)
  })

  it('handles multi-line content', () => {
    expect(countWords('line one\nline two\n你好')).toBe(4 + 2)
  })

  it('ignores markdown syntax characters but counts words inside them', () => {
    // # [] () are punctuation, ignored; title + url = 2 English words
    expect(countWords('# [title](url)')).toBe(2)
  })

  it('counts repeated words separately', () => {
    expect(countWords('test test test')).toBe(3)
  })
})

// ── useEditor composable shape ──

describe('useEditor', () => {
  it('exports useEditor as a function', async () => {
    const mod = await import('../composables/useEditor')
    expect(typeof mod.useEditor).toBe('function')
  })

  it('returns an object with all expected properties', async () => {
    const { useEditor } = await import('../composables/useEditor')

    const containerRef = ref<HTMLElement | null>(null)
    const theme = ref<'light' | 'dark'>('light')
    const onSave = vi.fn()

    const result = useEditor(containerRef, 'initial text', theme, onSave)

    // Reactive state
    expect(result.body).toBeDefined()
    expect(result.body.value).toBe('initial text')
    expect(result.dirty).toBeDefined()
    expect(result.dirty.value).toBe(false)
    expect(result.saving).toBeDefined()
    expect(result.saving.value).toBe(false)
    expect(result.savedAt).toBeDefined()
    expect(result.savedAt.value).toBeNull()
    expect(result.wordCount).toBeDefined()
    expect(result.wordCount.value).toBe(0) // not counted until mount
    expect(result.mode).toBeDefined()
    expect(result.mode.value).toBe('wysiwyg')

    // Functions
    expect(typeof result.mount).toBe('function')
    expect(typeof result.destroy).toBe('function')
    expect(typeof result.save).toBe('function')
    expect(typeof result.toggleMode).toBe('function')
    expect(typeof result.execBold).toBe('function')
    expect(typeof result.execItalic).toBe('function')
    expect(typeof result.execLink).toBe('function')
    expect(typeof result.execImage).toBe('function')
    expect(typeof result.execCode).toBe('function')
    expect(typeof result.execHeading).toBe('function')
    expect(typeof result.insertText).toBe('function')
    expect(typeof result.goToLine).toBe('function')
  })

  it('accepts optional onPasteImage parameter', async () => {
    const { useEditor } = await import('../composables/useEditor')

    const containerRef = ref<HTMLElement | null>(null)
    const theme = ref<'light' | 'dark'>('dark')
    const onSave = vi.fn()
    const onPasteImage = vi.fn(async () => '/img/test.png')

    // Should not throw with optional parameter
    const result = useEditor(containerRef, '', theme, onSave, onPasteImage)
    expect(result).toBeDefined()
    expect(typeof result.mount).toBe('function')
  })

  it('initial body is set from constructor argument', async () => {
    const { useEditor } = await import('../composables/useEditor')

    const containerRef = ref<HTMLElement | null>(null)
    const theme = ref<'light' | 'dark'>('light')
    const onSave = vi.fn()

    const result = useEditor(containerRef, '# Hello World', theme, onSave)
    expect(result.body.value).toBe('# Hello World')
  })

  it('mode defaults to wysiwyg', async () => {
    const { useEditor } = await import('../composables/useEditor')

    const containerRef = ref<HTMLElement | null>(null)
    const theme = ref<'light' | 'dark'>('light')
    const onSave = vi.fn()

    const result = useEditor(containerRef, '', theme, onSave)
    expect(result.mode.value).toBe('wysiwyg')
  })

  it('save function returns a promise', async () => {
    const { useEditor } = await import('../composables/useEditor')

    const containerRef = ref<HTMLElement | null>(null)
    const theme = ref<'light' | 'dark'>('light')
    const onSave = vi.fn().mockResolvedValue(undefined)

    const result = useEditor(containerRef, 'test content', theme, onSave)
    // save() should work even without mount (it calls onSave with body.value)
    const promise = result.save()
    expect(promise).toBeInstanceOf(Promise)
    await promise
    expect(onSave).toHaveBeenCalledWith('test content')
    expect(result.dirty.value).toBe(false)
    expect(result.saving.value).toBe(false)
  })
})

// ── useWysiwyg exports ──

describe('useWysiwyg', () => {
  it('exports wysiwyg as a function', async () => {
    const mod = await import('../composables/useWysiwyg')
    expect(typeof mod.wysiwyg).toBe('function')
  })

  it('exports mathExtension as a function', async () => {
    const mod = await import('../composables/useWysiwyg')
    expect(typeof mod.mathExtension).toBe('function')
  })

  it('wysiwyg() returns an array of extensions', async () => {
    const { wysiwyg } = await import('../composables/useWysiwyg')
    const extensions = wysiwyg()
    expect(Array.isArray(extensions)).toBe(true)
    expect(extensions.length).toBe(2) // ViewPlugin + baseTheme
  })

  it('mathExtension() returns a MarkdownExtension object', async () => {
    const { mathExtension } = await import('../composables/useWysiwyg')
    const ext = mathExtension()
    expect(ext).toBeDefined()
    expect(ext.defineNodes).toBeDefined()
    expect(Array.isArray(ext.defineNodes)).toBe(true)
    expect(ext.defineNodes.length).toBeGreaterThan(0)
    expect(ext.parseInline).toBeDefined()
    expect(Array.isArray(ext.parseInline)).toBe(true)
    expect(ext.parseBlock).toBeDefined()
    expect(Array.isArray(ext.parseBlock)).toBe(true)
  })

  it('mathExtension defines MathInline and MathBlock nodes', async () => {
    const { mathExtension } = await import('../composables/useWysiwyg')
    const ext = mathExtension()
    const nodeNames = ext.defineNodes.map((n: string | { name: string }) =>
      typeof n === 'string' ? n : n.name,
    )
    expect(nodeNames).toContain('MathInline')
    expect(nodeNames).toContain('MathBlock')
    expect(nodeNames).toContain('MathMark')
    expect(nodeNames).toContain('MathContent')
  })
})
