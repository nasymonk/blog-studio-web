<script setup lang="ts">
defineProps<{ lines?: number }>()
</script>

<template>
  <div class="space-y-3 p-4">
    <div
      v-for="i in (lines ?? 6)"
      :key="i"
      class="h-3 rounded bg-muted shimmer"
      :style="{ width: `${55 + Math.random() * 35}%` }"
    />
  </div>
</template>
