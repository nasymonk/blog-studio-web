<script setup lang="ts">
import type { SplitMode } from '@/composables/useSplitView'

defineProps<{
  mode: SplitMode
  splitRatio: number
  isDragging: boolean
}>()

const emit = defineEmits<{
  dragStart: [e: MouseEvent]
}>()
</script>

<template>
  <div class="flex h-full overflow-hidden" :class="{ 'select-none': isDragging }">
    <!-- Source panel -->
    <div
      v-show="mode !== 'preview'"
      class="overflow-auto"
      :style="mode === 'split' ? { width: `${splitRatio * 100}%` } : { width: '100%' }"
    >
      <slot name="source" />
    </div>

    <!-- Divider -->
    <div
      v-if="mode === 'split'"
      class="w-1 shrink-0 cursor-col-resize bg-border/60 hover:bg-accent/40 transition-colors relative group"
      @mousedown="emit('dragStart', $event)"
    >
      <div class="absolute inset-y-0 -left-1 -right-1" />
    </div>

    <!-- Preview panel -->
    <div
      v-show="mode !== 'source'"
      class="overflow-auto"
      :style="mode === 'split' ? { width: `${(1 - splitRatio) * 100}%` } : { width: '100%' }"
    >
      <slot name="preview" />
    </div>
  </div>
</template>
