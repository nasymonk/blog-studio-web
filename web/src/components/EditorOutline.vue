<script setup lang="ts">
import { computed } from 'vue'

interface Heading { level: number; text: string; line: number }

const props = defineProps<{
  headings: Heading[]
  activeLine: number
}>()

const emit = defineEmits<{
  jump: [line: number]
}>()

const activeIndex = computed(() => {
  const hs = props.headings
  if (!hs.length || props.activeLine < 1) return -1
  for (let i = hs.length - 1; i >= 0; i--) {
    if (props.activeLine >= hs[i].line) return i
  }
  return -1
})

function indentClass(level: number) {
  if (level <= 1) return 'pl-2'
  if (level === 2) return 'pl-5'
  return 'pl-8'
}
</script>

<template>
  <div class="w-[200px] min-w-[160px] border-l border-border overflow-y-auto text-[13px] bg-muted/50 shrink-0 hidden md:block">
    <div class="px-3 pt-2.5 pb-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">大纲</div>
    <div v-if="headings.length === 0" class="px-3 py-3 text-muted-foreground">暂无标题</div>
    <ul v-else class="list-none m-0 p-0 pb-4">
      <li
        v-for="(h, i) in headings"
        :key="h.line"
        :aria-current="i === activeIndex ? 'true' : undefined"
        class="flex items-baseline gap-1.5 py-1 pr-2 cursor-pointer rounded leading-snug break-words transition-colors"
        :class="[
          indentClass(h.level),
          i === activeIndex
            ? 'bg-accent/10 text-accent font-medium'
            : 'text-foreground hover:bg-muted/80',
        ]"
        :title="`第 ${h.line} 行`"
        @click="emit('jump', h.line)"
      >
        <span class="text-[10px] opacity-40 shrink-0 font-semibold">H{{ h.level }}</span>
        <span class="flex-1">{{ h.text }}</span>
      </li>
    </ul>
  </div>
</template>
