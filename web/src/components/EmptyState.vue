<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  actionLabel?: string
}>()

defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center animate-fade-up">
    <div class="mb-4 text-muted-foreground/30">
      <slot name="icon">
        <svg class="h-16 w-16" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <rect x="12" y="8" width="40" height="48" rx="4" />
          <line x1="20" y1="20" x2="44" y2="20" />
          <line x1="20" y1="28" x2="38" y2="28" />
          <line x1="20" y1="36" x2="42" y2="36" />
          <line x1="20" y1="44" x2="32" y2="44" />
        </svg>
      </slot>
    </div>
    <h3 class="font-serif text-lg font-semibold text-foreground/80 mb-1">{{ title }}</h3>
    <p v-if="description" class="text-sm text-muted-foreground max-w-sm mb-4">{{ description }}</p>
    <button
      v-if="actionLabel"
      class="px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>
