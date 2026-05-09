<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  body: string
  onJump: (lineNumber: number) => void
}>()

interface Heading { level: number; text: string; line: number }

const headings = computed<Heading[]>(() => {
  const result: Heading[] = []
  const lines = props.body.split('\n')
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(/^(#{1,6})\s+(.+)$/)
    if (m) result.push({ level: m[1].length, text: m[2], line: i + 1 })
  }
  return result
})
</script>

<template>
  <div class="w-[200px] min-w-[160px] border-l border-border overflow-y-auto text-[13px] bg-muted/50 shrink-0 hidden md:block">
    <div class="px-3 pt-2.5 pb-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">大纲</div>
    <div v-if="headings.length === 0" class="px-3 py-3 text-muted-foreground">暂无标题</div>
    <ul v-else class="list-none m-0 p-0 pb-4">
      <li
        v-for="h in headings"
        :key="h.line"
        class="flex items-baseline gap-1.5 py-1 pr-2 cursor-pointer rounded text-foreground leading-snug break-words hover:bg-muted/80"
        :style="{ paddingLeft: (h.level - 1) * 12 + 8 + 'px' }"
        :title="`第 ${h.line} 行`"
        @click="onJump(h.line)"
      >
        <span class="text-[10px] opacity-40 shrink-0 font-semibold">H{{ h.level }}</span>
        <span class="flex-1">{{ h.text }}</span>
      </li>
    </ul>
  </div>
</template>
