<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{ source: string }>()

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

const html = computed(() => DOMPurify.sanitize(marked.parse(props.source) as string))
</script>

<template>
  <div class="markdown-body" v-html="html"></div>
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
