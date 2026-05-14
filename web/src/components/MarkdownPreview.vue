<script setup lang="ts">
import { computed, nextTick, watch, useTemplateRef, onMounted } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{ source: string }>()
const previewRef = useTemplateRef<HTMLElement>('preview')

const MATH_BLOCK = 'MATHBLOCK'
const MATH_INLINE = 'MATHINLINE'
const mathRaw = new Map<string, string>()
let mathCounter = 0

// Katex renderToString — loaded lazily (~60KB) only when math is detected in content
type KatexRenderFn = (math: string, displayMode: boolean) => string
let renderKatex: KatexRenderFn | null = null
let katexLoading = false

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
    image({ href, title, text }: { href: string; title?: string | null; text: string }) {
      const titleAttr = title ? ` title="${title}"` : ''
      return `<img src="${href}" alt="${text}" loading="lazy" decoding="async"${titleAttr} />`
    },
  },
})

function extractMath(text: string): string {
  mathRaw.clear()
  mathCounter = 0
  // Block math: $$...$$
  text = text.replace(/\$\$([\s\S]*?)\$\$/g, (_, math: string) => {
    const key = `${MATH_BLOCK}${mathCounter++}MATH`
    mathRaw.set(key, math.trim())
    return key
  })
  // Inline math: $...$ (not $$)
  text = text.replace(/(?<!\$)\$(?!\$)(.*?)\$/g, (_, math: string) => {
    const key = `${MATH_INLINE}${mathCounter++}MATH`
    mathRaw.set(key, math.trim())
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
  if (!renderKatex) return
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const toReplace: { node: Text; key: string }[] = []
  while (walker.nextNode()) {
    const textNode = walker.currentNode as Text
    const text = textNode.textContent || ''
    for (const [key] of mathRaw) {
      if (text.includes(key)) {
        toReplace.push({ node: textNode, key })
        break
      }
    }
  }
  for (const { node, key } of toReplace) {
    const rawMath = mathRaw.get(key)
    if (!rawMath) continue
    try {
      const isBlock = key.startsWith(MATH_BLOCK)
      const rendered = renderKatex(rawMath, isBlock)
      const span = document.createElement('span')
      span.className = isBlock ? 'math-block' : 'math-inline'
      span.innerHTML = rendered
      const parent = node.parentNode
      if (parent) {
        const remaining = (node.textContent || '').replace(key, '')
        parent.replaceChild(span, node)
        if (remaining) {
          parent.insertBefore(document.createTextNode(remaining), span.nextSibling)
        }
      }
    } catch {
      // If KaTeX rendering fails, show raw math as inline code
      const code = document.createElement('code')
      code.textContent = rawMath
      node.parentNode?.replaceChild(code, node)
    }
  }
}

function attachImageLoadListeners(root: HTMLElement) {
  for (const img of root.querySelectorAll<HTMLImageElement>('img')) {
    if (img.complete) {
      img.classList.add('loaded')
    } else {
      img.addEventListener('load', () => img.classList.add('loaded'), { once: true })
    }
  }
}

watch(html, () => {
  nextTick(() => {
    if (previewRef.value) {
      replaceMathPlaceholders(previewRef.value)
      attachImageLoadListeners(previewRef.value)
    }
  })
})

// Lazy-load KaTeX (~60KB) only when math delimiters are detected in content
async function loadKatex() {
  if (katexLoading) return
  katexLoading = true
  try {
    const mod = await import('katex')
    await import('katex/dist/katex.min.css')
    renderKatex = (math: string, displayMode: boolean) =>
      mod.default.renderToString(math, { throwOnError: false, displayMode })
    if (previewRef.value) {
      replaceMathPlaceholders(previewRef.value)
    }
  } catch {
    // KaTeX failed to load; math placeholders remain visible as raw text
  }
}

function hasMath(text: string): boolean {
  return /\$\$/.test(text) || /(?<!\$)\$(?!\$)/.test(text)
}

onMounted(() => {
  if (hasMath(props.source)) loadKatex()
})

watch(() => props.source, (val) => {
  if (!renderKatex && hasMath(val)) loadKatex()
})
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
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
