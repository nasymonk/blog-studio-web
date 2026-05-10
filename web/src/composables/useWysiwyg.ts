import { EditorView, Decoration, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'
import { syntaxTree } from '@codemirror/language'
import type { InlineParser, BlockParser, MarkdownExtension } from '@lezer/markdown'

// ── Heading helpers ──────────────────────────────────────────
const headingLevels: Record<string, number> = {
  ATXHeading1: 1, ATXHeading2: 2, ATXHeading3: 3,
  ATXHeading4: 4, ATXHeading5: 5, ATXHeading6: 6,
}

function headingClass(level: number): string {
  return `cm-heading cm-heading${level}`
}

// ── Widgets ─────────────────────────────────────────────────
class HrWidget extends WidgetType {
  toDOM() {
    const hr = document.createElement('hr')
    hr.className = 'cm-hr'
    return hr
  }
}

class ImageWidget extends WidgetType {
  constructor(readonly url: string) { super() }
  toDOM() {
    const img = document.createElement('img')
    img.src = this.url
    img.className = 'cm-inlineImage'
    img.loading = 'lazy'
    img.alt = ''
    return img
  }
  eq(other: ImageWidget) { return this.url === other.url }
}

// ── Math extension (custom Lezer parsers) ───────────────────
const mathInlineParser: InlineParser = {
  name: 'mathInline',
  parse(cx, next, pos) {
    if (next !== 36 /* $ */ || cx.char(pos + 1) === 36) return -1
    for (let i = pos + 1; i < cx.end; i++) {
      const ch = cx.char(i)
      if (ch === 92 /* \ */) { i++; continue }
      if (ch === 36 /* $ */) {
        if (i === pos + 1) return -1
        return cx.addElement(cx.elt('MathInline', pos, i + 1, [
          cx.elt('MathMark', pos, pos + 1),
          cx.elt('MathContent', pos + 1, i),
          cx.elt('MathMark', i, i + 1),
        ]))
      }
    }
    return -1
  },
}

const mathBlockParser: BlockParser = {
  name: 'mathBlock',
  parse(cx, line) {
    if (!line.text.startsWith('$$')) return false
    const start = cx.lineStart
    // Single-line: $$...$$
    const trimmed = line.text.trim()
    if (trimmed.length > 2 && trimmed.endsWith('$$') && trimmed !== '$$$$') {
      cx.addElement(cx.elt('MathBlock', start, start + line.text.length))
      return true
    }
    // Multi-line: consume until closing $$
    while (cx.nextLine()) {
      if (line.text.trim() === '$$') {
        const end = cx.lineStart + line.text.length
        cx.addElement(cx.elt('MathBlock', start, end))
        return true
      }
    }
    // No closing $$ found — treat whole thing as a block
    cx.addElement(cx.elt('MathBlock', start, cx.lineStart + line.text.length))
    return true
  },
}

export function mathExtension(): MarkdownExtension {
  return {
    defineNodes: [
      'MathInline',
      { name: 'MathBlock', block: true },
      'MathMark',
      'MathContent',
    ],
    parseInline: [mathInlineParser],
    parseBlock: [mathBlockParser],
  }
}

// ── Cursor-aware decoration builder ─────────────────────────
function isInside(from: number, to: number, selFrom: number, selTo: number): boolean {
  return selFrom <= to && selTo >= from
}

function buildDecorations(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  const doc = view.state.doc
  const sel = view.state.selection.main
  const selFrom = sel.from
  const selTo = sel.to

  // Collect all decorations then sort by from position
  const decos: Array<{ from: number; to: number; deco: Decoration }> = []

  for (const { from, to } of view.visibleRanges) {
    syntaxTree(view.state).iterate({
      from, to,
      enter(node) {
        const name = node.name
        const nodeFrom = node.from
        const nodeTo = node.to

        // ── Headings ──
        const headingLevel = headingLevels[name]
        if (headingLevel) {
          let markEnd = nodeFrom
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'HeaderMark') {
              markEnd = child.to
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-headingMark cm-headingMark-show' }) })
            }
            child = child.nextSibling
          }
          if (markEnd < nodeTo) {
            decos.push({ from: markEnd, to: nodeTo, deco: Decoration.mark({ class: headingClass(headingLevel) }) })
          }
          return false
        }

        // ── Strong emphasis (**bold**) ──
        if (name === 'StrongEmphasis') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-strongMark cm-strongMark-show' }) })
            } else {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-strong' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── Emphasis (*italic*) ──
        if (name === 'Emphasis') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-emphasisMark cm-emphasisMark-show' }) })
            } else {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-emphasis' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── Inline code (`code`) ──
        if (name === 'InlineCode') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'CodeMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-codeMark cm-codeMark-show' }) })
            } else if (child.name === 'CodeText') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-inlineCode' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── Fenced code block ──
        if (name === 'FencedCode') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'CodeMark' || child.name === 'CodeInfo') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-codeMark cm-codeMark-show' }) })
            } else if (child.name === 'CodeText') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-codeBlock' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── Indented code block ──
        if (name === 'CodeBlock') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-codeBlock' }) })
          return false
        }

        // ── Blockquote ──
        if (name === 'Blockquote') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'QuoteMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-quoteMark cm-quoteMark-show' }) })
            }
            child = child.nextSibling
          }
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-blockquote' }) })
          return false
        }

        // ── Link [text](url) ──
        if (name === 'Link') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'LinkMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-linkUrl cm-linkUrl-show' }) })
            } else if (child.name === 'LinkLabel') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-linkText' }) })
            } else if (child.name === 'URL' || child.name === 'LinkTitle') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-linkUrl cm-linkUrl-show' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── Image ![alt](url) ──
        if (name === 'Image') {
          let child = node.node.firstChild
          let url = ''
          while (child) {
            if (child.name === 'LinkMark' || child.name === 'LinkLabel') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-linkUrl cm-linkUrl-show' }) })
            } else if (child.name === 'URL') {
              url = doc.sliceString(child.from + 1, child.to - 1)
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-linkUrl cm-linkUrl-show' }) })
            }
            child = child.nextSibling
          }
          if (url) {
            decos.push({ from: nodeTo, to: nodeTo, deco: Decoration.widget({ widget: new ImageWidget(url), side: 1 }) })
          }
          return false
        }

        // ── Horizontal rule ──
        if (name === 'HorizontalRule') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.replace({ widget: new HrWidget() }) })
          return false
        }

        // ── Strikethrough ~~text~~ ──
        if (name === 'Strikethrough') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-strikeMark cm-strikeMark-show' }) })
            } else {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-strikethrough' }) })
            }
            child = child.nextSibling
          }
          return false
        }

        // ── List markers ──
        if (name === 'ListMark') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-listMark' }) })
          return false
        }

        // ── Task list ──
        if (name === 'TaskList') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-taskList' }) })
          return false
        }

        // ── Table ──
        if (name === 'Table') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-table' }) })
        }
        if (name === 'TableRow') {
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-tableRow' }) })
        }
        if (name === 'TableCell') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'TableDelimiter') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-tableDelimiter' }) })
            }
            child = child.nextSibling
          }
          decos.push({ from: nodeFrom, to: nodeTo, deco: Decoration.mark({ class: 'cm-tableCell' }) })
          return false
        }

        // ── Math ──
        if (name === 'MathInline') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'MathMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-mathMark cm-mathMark-show' }) })
            } else if (child.name === 'MathContent') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-mathInline' }) })
            }
            child = child.nextSibling
          }
          return false
        }
        if (name === 'MathBlock') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'MathMark') {
              const hide = !isInside(child.from, child.to, selFrom, selTo)
              decos.push({ from: child.from, to: child.to, deco: hide ? Decoration.replace({}) : Decoration.mark({ class: 'cm-mathMark cm-mathMark-show' }) })
            } else if (child.name === 'MathContent') {
              decos.push({ from: child.from, to: child.to, deco: Decoration.mark({ class: 'cm-mathBlock' }) })
            }
            child = child.nextSibling
          }
          return false
        }
      },
    })
  }

  // Sort by from position (required by RangeSetBuilder)
  decos.sort((a, b) => a.from - b.from || a.deco.startSide - b.deco.startSide)
  for (const d of decos) {
    builder.add(d.from, d.to, d.deco)
  }
  return builder.finish()
}

// ── ViewPlugin ──────────────────────────────────────────────
const wysiwygDecoration = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }
    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged || update.selectionSet) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  { decorations: v => v.decorations }
)

// ── Exported extension ──────────────────────────────────────
export function wysiwyg() {
  return [
    wysiwygDecoration,
    EditorView.baseTheme({
      // Hide syntax marks (when cursor is NOT inside)
      '.cm-headingMark, .cm-strongMark, .cm-emphasisMark, .cm-codeMark, .cm-linkUrl, .cm-strikeMark, .cm-quoteMark, .cm-mathMark': {
        display: 'inline-block',
        width: '0',
        overflow: 'hidden',
        opacity: '0',
        userSelect: 'none',
      },
      // Show syntax marks (when cursor IS inside)
      '& .cm-headingMark-show, & .cm-strongMark-show, & .cm-emphasisMark-show, & .cm-codeMark-show, & .cm-linkUrl-show, & .cm-strikeMark-show, & .cm-quoteMark-show, & .cm-mathMark-show': {
        display: 'inline',
        width: 'auto',
        overflow: 'visible',
        opacity: '0.4',
        userSelect: 'auto',
      },
      // Headings
      '.cm-heading1': { fontSize: '1.5em', fontWeight: 'bold', fontFamily: 'var(--font-serif)', lineHeight: '1.3' },
      '.cm-heading2': { fontSize: '1.3em', fontWeight: 'bold', fontFamily: 'var(--font-serif)', lineHeight: '1.3' },
      '.cm-heading3': { fontSize: '1.15em', fontWeight: 'bold', fontFamily: 'var(--font-serif)', lineHeight: '1.3' },
      '.cm-heading4': { fontSize: '1.05em', fontWeight: '600', fontFamily: 'var(--font-serif)', lineHeight: '1.4' },
      '.cm-heading5': { fontSize: '1em', fontWeight: '600', fontFamily: 'var(--font-serif)', lineHeight: '1.4' },
      '.cm-heading6': { fontSize: '0.95em', fontWeight: '600', fontFamily: 'var(--font-serif)', lineHeight: '1.4' },
      // Text styles
      '.cm-strong': { fontWeight: 'bold' },
      '.cm-emphasis': { fontStyle: 'italic' },
      '.cm-strikethrough': { textDecoration: 'line-through' },
      // Inline code
      '.cm-inlineCode': {
        fontFamily: 'var(--font-mono)',
        fontSize: '0.9em',
        background: 'var(--color-muted)',
        borderRadius: '3px',
        padding: '1px 4px',
      },
      // Code blocks
      '.cm-codeBlock': {
        fontFamily: 'var(--font-mono)',
        fontSize: '0.85em',
        background: 'var(--color-muted)',
        display: 'block',
        borderRadius: '4px',
        padding: '0.75em 1em',
        margin: '0.5em 0',
        lineHeight: '1.6',
      },
      // Blockquote
      '.cm-blockquote': {
        borderLeft: '3px solid var(--color-accent)',
        paddingLeft: '0.75em',
        color: 'var(--color-muted-foreground)',
        fontStyle: 'italic',
      },
      // Links
      '.cm-linkText': {
        color: 'var(--color-link)',
        textDecoration: 'underline',
        cursor: 'pointer',
      },
      // Horizontal rule
      '.cm-hr': {
        border: 'none',
        borderTop: '1px solid var(--color-border)',
        margin: '1em 0',
      },
      // Inline images
      '.cm-inlineImage': {
        maxWidth: '100%',
        borderRadius: '4px',
        display: 'block',
        margin: '0.5em 0',
      },
      // List markers
      '.cm-listMark': { opacity: '0.5' },
      // Task list
      '.cm-taskList': { color: 'var(--color-muted-foreground)' },
      // Table
      '.cm-table': { display: 'block', overflowX: 'auto', margin: '0.5em 0' },
      '.cm-tableRow': { display: 'flex', borderBottom: '1px solid var(--color-border)', minWidth: 'fit-content' },
      '.cm-tableCell': { flex: '1', padding: '0.35em 0.6em', minWidth: '60px', borderRight: '1px solid var(--color-border)' },
      '.cm-tableCell:last-child': { borderRight: 'none' },
      '.cm-tableRow:first-child .cm-tableCell': { fontWeight: '600', background: 'var(--color-muted)' },
      '.cm-tableDelimiter': { display: 'none' },
      // Math
      '.cm-mathInline': { fontStyle: 'italic', color: 'var(--color-accent)', fontFamily: 'var(--font-mono)', fontSize: '0.9em' },
      '.cm-mathBlock': { display: 'block', fontFamily: 'var(--font-mono)', fontSize: '0.9em', color: 'var(--color-accent)', background: 'var(--color-muted)', borderRadius: '4px', padding: '0.75em 1em', margin: '0.5em 0', textAlign: 'center' },
    }),
  ]
}
