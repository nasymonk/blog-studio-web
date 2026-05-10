<script setup lang="ts">
defineProps<{
  wordCount: number
  charCount: number
  lineCount: number
  cursorLine: number
  cursorCol: number
  readingTime: string
}>()
</script>

<template>
  <div class="flex items-center gap-4 px-4 py-1.5 text-[11px] text-muted-foreground border-t border-border bg-muted/30 select-none">
    <span>{{ wordCount }} 词</span>
    <span>{{ charCount }} 字符</span>
    <span>{{ lineCount }} 行</span>
    <span class="ml-auto">行 {{ cursorLine }}, 列 {{ cursorCol }}</span>
    <span>阅读 ~{{ readingTime }}</span>
  </div>
</template>
