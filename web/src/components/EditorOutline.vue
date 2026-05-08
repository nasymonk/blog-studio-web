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
  <div class="outline-panel">
    <div class="outline-header">大纲</div>
    <div v-if="headings.length === 0" class="outline-empty">暂无标题</div>
    <ul v-else class="outline-list">
      <li
        v-for="h in headings"
        :key="h.line"
        class="outline-item"
        :style="{ paddingLeft: (h.level - 1) * 12 + 8 + 'px' }"
        :title="`第 ${h.line} 行`"
        @click="onJump(h.line)"
      >
        <span class="outline-level">H{{ h.level }}</span>
        <span class="outline-text">{{ h.text }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.outline-panel {
  width: 200px;
  min-width: 160px;
  border-left: 1px solid var(--border-color, #e5e7eb);
  overflow-y: auto;
  font-size: 13px;
  background: var(--bg-secondary, #f9fafb);
  flex-shrink: 0;
}
.outline-header {
  padding: 10px 12px 6px;
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #9ca3af);
}
.outline-list { list-style: none; margin: 0; padding: 0 0 16px; }
.outline-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding-top: 4px;
  padding-bottom: 4px;
  padding-right: 8px;
  cursor: pointer;
  border-radius: 4px;
  color: var(--text-primary, #111);
  line-height: 1.3;
  word-break: break-word;
}
.outline-item:hover { background: var(--hover-bg, rgba(0,0,0,0.05)); }
.outline-level {
  font-size: 10px;
  opacity: 0.4;
  flex-shrink: 0;
  font-weight: 600;
}
.outline-text { flex: 1; }
.outline-empty { padding: 12px; color: var(--text-muted, #9ca3af); }

@media (max-width: 768px) {
  .outline-panel { display: none; }
}
</style>
