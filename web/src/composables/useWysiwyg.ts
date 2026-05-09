import { EditorView, Decoration, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'
import { syntaxTree } from '@codemirror/language'

const headingLevels: Record<string, number> = {
  ATXHeading1: 1, ATXHeading2: 2, ATXHeading3: 3,
  ATXHeading4: 4, ATXHeading5: 5, ATXHeading6: 6,
}

const headingMark = Decoration.mark({ class: 'cm-headingMark' })
const strongMark = Decoration.mark({ class: 'cm-strongMark' })
const emphasisMark = Decoration.mark({ class: 'cm-emphasisMark' })
const codeMark = Decoration.mark({ class: 'cm-codeMark' })
const linkUrlMark = Decoration.mark({ class: 'cm-linkUrl' })
const quoteMark = Decoration.mark({ class: 'cm-quoteMark' })
const listMark = Decoration.mark({ class: 'cm-listMark' })
const strikethroughMark = Decoration.mark({ class: 'cm-strikeMark' })

function headingClass(level: number): string {
  return `cm-heading cm-heading${level}`
}

class HrWidget extends WidgetType {
  toDOM() {
    const hr = document.createElement('hr')
    hr.className = 'cm-hr'
    return hr
  }
}

function buildDecorations(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  const doc = view.state.doc

  for (const { from, to } of view.visibleRanges) {
    syntaxTree(view.state).iterate({
      from, to,
      enter(node) {
        const name = node.name
        const nodeFrom = node.from
        const nodeTo = node.to

        // Headings
        const headingLevel = headingLevels[name]
        if (headingLevel) {
          // Find HeaderMark child to hide the # prefix
          let markEnd = nodeFrom
          const text = doc.sliceString(nodeFrom, nodeTo)
          const match = text.match(/^#{1,6}\s+/)
          if (match) markEnd = nodeFrom + match[0].length

          if (markEnd > nodeFrom) {
            builder.add(nodeFrom, markEnd, headingMark)
          }
          // Mark the content area with heading style (skip if it's just the mark)
          if (markEnd < nodeTo) {
            builder.add(markEnd, nodeTo, Decoration.mark({ class: headingClass(headingLevel) }))
          }
          return false // Don't recurse into children
        }

        // Strong emphasis (**bold**)
        if (name === 'StrongEmphasis') {
          let pos = nodeFrom
          node.node.getChild('EmphasisMark')?.to
          // Iterate children to find EmphasisMark nodes
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              builder.add(child.from, child.to, strongMark)
            }
            child = child.nextSibling
          }
          // Mark content (non-EmphasisMark children)
          child = node.node.firstChild
          while (child) {
            if (child.name !== 'EmphasisMark') {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-strong' }))
            }
            child = child.nextSibling
          }
          return false
        }

        // Emphasis (*italic*)
        if (name === 'Emphasis') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              builder.add(child.from, child.to, emphasisMark)
            } else {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-emphasis' }))
            }
            child = child.nextSibling
          }
          return false
        }

        // Inline code (`code`)
        if (name === 'InlineCode') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'CodeMark') {
              builder.add(child.from, child.to, codeMark)
            } else if (child.name === 'CodeText') {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-inlineCode' }))
            }
            child = child.nextSibling
          }
          return false
        }

        // Fenced code block
        if (name === 'FencedCode') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'CodeMark' || child.name === 'CodeInfo') {
              builder.add(child.from, child.to, codeMark)
            } else if (child.name === 'CodeText') {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-codeBlock' }))
            }
            child = child.nextSibling
          }
          return false
        }

        // Indented code block
        if (name === 'CodeBlock') {
          builder.add(nodeFrom, nodeTo, Decoration.mark({ class: 'cm-codeBlock' }))
          return false
        }

        // Blockquote
        if (name === 'Blockquote') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'QuoteMark') {
              builder.add(child.from, child.to, quoteMark)
            }
            child = child.nextSibling
          }
          builder.add(nodeFrom, nodeTo, Decoration.mark({ class: 'cm-blockquote' }))
          return false
        }

        // Link [text](url)
        if (name === 'Link') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'LinkMark') {
              builder.add(child.from, child.to, linkUrlMark)
            } else if (child.name === 'LinkLabel') {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-linkText' }))
            } else if (child.name === 'URL') {
              builder.add(child.from, child.to, linkUrlMark)
            } else if (child.name === 'LinkTitle') {
              builder.add(child.from, child.to, linkUrlMark)
            }
            child = child.nextSibling
          }
          return false
        }

        // Image ![alt](url)
        if (name === 'Image') {
          let child = node.node.firstChild
          let url = ''
          while (child) {
            if (child.name === 'LinkMark') {
              builder.add(child.from, child.to, linkUrlMark)
            } else if (child.name === 'LinkLabel') {
              builder.add(child.from, child.to, linkUrlMark)
            } else if (child.name === 'URL') {
              url = doc.sliceString(child.from + 1, child.to - 1)
              builder.add(child.from, child.to, linkUrlMark)
            }
            child = child.nextSibling
          }
          // Insert image widget after the Image node
          if (url) {
            builder.add(nodeTo, nodeTo, Decoration.widget({
              widget: new ImageWidget(url),
              side: 1,
            }))
          }
          return false
        }

        // Horizontal rule
        if (name === 'HorizontalRule') {
          builder.add(nodeFrom, nodeTo, Decoration.replace({
            widget: new HrWidget(),
          }))
          return false
        }

        // Strikethrough ~~text~~
        if (name === 'Strikethrough') {
          let child = node.node.firstChild
          while (child) {
            if (child.name === 'EmphasisMark') {
              builder.add(child.from, child.to, strikethroughMark)
            } else {
              builder.add(child.from, child.to, Decoration.mark({ class: 'cm-strikethrough' }))
            }
            child = child.nextSibling
          }
          return false
        }

        // List markers
        if (name === 'ListMark') {
          builder.add(nodeFrom, nodeTo, listMark)
          return false
        }

        // Task list checkbox
        if (name === 'TaskList') {
          builder.add(nodeFrom, nodeTo, Decoration.mark({ class: 'cm-taskList' }))
          return false
        }
      }
    })
  }

  return builder.finish()
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

const wysiwygDecoration = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }
    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  { decorations: v => v.decorations }
)

export function wysiwyg() {
  return [
    wysiwygDecoration,
    EditorView.baseTheme({
      '.cm-heading1': { fontSize: '1.5em', fontWeight: 'bold', fontFamily: 'var(--font-serif)' },
      '.cm-heading2': { fontSize: '1.3em', fontWeight: 'bold', fontFamily: 'var(--font-serif)' },
      '.cm-heading3': { fontSize: '1.15em', fontWeight: 'bold', fontFamily: 'var(--font-serif)' },
      '.cm-heading4': { fontSize: '1.05em', fontWeight: '600', fontFamily: 'var(--font-serif)' },
      '.cm-heading5': { fontSize: '1em', fontWeight: '600', fontFamily: 'var(--font-serif)' },
      '.cm-heading6': { fontSize: '0.95em', fontWeight: '600', fontFamily: 'var(--font-serif)' },
      '.cm-headingMark': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-strongMark': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-emphasisMark': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-codeMark': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-linkUrl': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-quoteMark': { opacity: '0.4', fontSize: '0.85em', userSelect: 'none' },
      '.cm-listMark': { opacity: '0.5', userSelect: 'none' },
      '.cm-strikeMark': { opacity: '0', fontSize: '0', userSelect: 'none' },
      '.cm-strong': { fontWeight: 'bold' },
      '.cm-emphasis': { fontStyle: 'italic' },
      '.cm-strikethrough': { textDecoration: 'line-through' },
      '.cm-inlineCode': {
        fontFamily: 'var(--font-mono)',
        fontSize: '0.9em',
        background: 'var(--color-muted)',
        borderRadius: '3px',
        padding: '1px 4px',
      },
      '.cm-codeBlock': {
        fontFamily: 'var(--font-mono)',
        fontSize: '0.9em',
        background: 'var(--color-muted)',
        display: 'block',
        borderRadius: '4px',
        padding: '0.5em 1em',
        margin: '0.25em 0',
      },
      '.cm-blockquote': {
        borderLeft: '3px solid var(--color-accent)',
        paddingLeft: '0.75em',
        color: 'var(--color-muted-foreground)',
        fontStyle: 'italic',
      },
      '.cm-linkText': {
        color: 'var(--color-link)',
        textDecoration: 'underline',
        cursor: 'pointer',
      },
      '.cm-hr': {
        border: 'none',
        borderTop: '1px solid var(--color-border)',
        margin: '1em 0',
      },
      '.cm-inlineImage': {
        maxWidth: '100%',
        borderRadius: '4px',
        display: 'block',
        margin: '0.5em 0',
      },
      '.cm-taskList': {
        color: 'var(--color-muted-foreground)',
      },
      '&dark .cm-headingMark, &dark .cm-strongMark, &dark .cm-emphasisMark, &dark .cm-codeMark, &dark .cm-linkUrl, &dark .cm-strikeMark': {},
    }),
  ]
}
