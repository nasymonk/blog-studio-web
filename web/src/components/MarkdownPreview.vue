<script setup lang="ts">
import { computed, nextTick, watch, useTemplateRef } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import katex from 'katex'
import 'katex/dist/katex.min.css'

const props = defineProps<{ source: string }>()
const previewRef = useTemplateRef<HTMLElement>('preview')

const MATH_BLOCK = 'MATHBLOCK'
const MATH_INLINE = 'MATHINLINE'
const mathMap = new Map<string, string>()
let mathCounter = 0

// Configure DOMPurify to allow KaTeX elements and inline styles
const purifyConfig = {
  ADD_TAGS: ['svg', 'path', 'line', 'polyline', 'polygon', 'rect', 'circle', 'use'],
  ADD_ATTR: ['style', 'viewBox', 'xmlns', 'width', 'height', 'd', 'x1', 'y1', 'x2', 'y2',
    'points', 'x', 'y', 'cx', 'cy', 'r', 'fill', 'stroke', 'stroke-width',
    'transform', 'xlink:href', 'preserveAspectRatio'],
} satisfies DOMPurify.Config

marked.use({
  gfm: true,
  breaks: false,
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      const langLabel = lang ? `<span class="code-lang">${lang}</span>` : ''
      const escaped = text.replace(/</g, '&lt;').replace(/>/g, '&gt;')
      return `<pre>${langLabel}<code>${escaped}</code></pre>`
    },
  },
})

function extractMath(text: string): string {
  mathMap.clear()
  mathCounter = 0
  // Block math: $$...$$
  text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_, math: string) => {
    const key = `${MATH_BLOCK}${mathCounter++}MATH`
    try {
      mathMap.set(key, katex.renderToString(math.trim(), { throwOnError: false, displayMode: true }))
    } catch {
      mathMap.set(key, `<span class="math-error">${math}</span>`)
    }
    return key
  })
  // Inline math: $...$ (not $$)
  text = text.replace(/(?<!\$)\$(?!\$)(.*?)\$/g, (_, math: string) => {
    const key = `${MATH_INLINE}${mathCounter++}MATH`
    try {
      mathMap.set(key, katex.renderToString(math.trim(), { throwOnError: false, displayMode: false }))
    } catch {
      mathMap.set(key, `<code>${math}</code>`)
    }
    return key
  })
  return text
}

const html = computed(() => {
  const preprocessed = extractMath(props.source)
  const raw = marked.parse(preprocessed) as string
  return DOMPurify.sanitize(raw, purifyConfig)
})

function replaceMathPlaceholders(root: HTMLElement) {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const toReplace: { node: Text; key: string }[] = []
  while (walker.nextNode()) {
    const textNode = walker.currentNode as Text
    const text = textNode.textContent || ''
    for (const [key] of mathMap) {
      if (text.includes(key)) {
        toReplace.push({ node: textNode, key })
        break
      }
    }
  }
  for (const { node, key } of toReplace) {
    const mathHtml = mathMap.get(key)
    if (!mathHtml) continue
    const span = document.createElement('span')
    span.className = key.startsWith(MATH_BLOCK) ? 'math-block' : 'math-inline'
    span.innerHTML = mathHtml
    const parent = node.parentNode
    if (parent) {
      const remaining = (node.textContent || '').replace(key, '')
      parent.replaceChild(span, node)
      if (remaining) {
        parent.insertBefore(document.createTextNode(remaining), span.nextSibling)
      }
    }
  }
}

watch(html, () => {
  nextTick(() => {
    if (previewRef.value) {
      replaceMathPlaceholders(previewRef.value)
    }
  })
})
</script>

<template>
  <div ref="preview" class="markdown-body" v-html="html"></div>
</template>

<style scoped>
.markdown-body :deep(pre) {
  position: relative;
}
.markdown-body :deep(.code-lang) {
  position: absolute;
  top: 0.4em;
  right: 0.6em;
  font-size: 0.7em;
  color: var(--muted-foreground);
  opacity: 0.6;
  font-family: var(--font-sans);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  pointer-events: none;
}
</style>
