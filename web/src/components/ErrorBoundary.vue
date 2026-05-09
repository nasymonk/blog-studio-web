<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import { AlertCircleIcon, RotateCcwIcon } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const error = ref<Error | null>(null)
const info = ref('')

onErrorCaptured((err, _, stack) => {
  error.value = err
  info.value = stack ?? ''
  return false
})

function retry() {
  error.value = null
  info.value = ''
}
</script>

<template>
  <div v-if="error" class="flex-1 flex flex-col items-center justify-center gap-4 p-8 animate-fade-up">
    <AlertCircleIcon class="h-12 w-12 text-destructive/40" />
    <div class="text-center space-y-1">
      <p class="font-serif font-semibold text-lg">Something went wrong</p>
      <p class="text-sm text-muted-foreground max-w-md">{{ error.message }}</p>
      <p v-if="info" class="text-[11px] font-mono text-muted-foreground/50 max-w-md truncate">{{ info }}</p>
    </div>
    <Button class="rounded-full px-5" @click="retry">
      <RotateCcwIcon class="h-3.5 w-3.5 mr-1.5" />
      Retry
    </Button>
  </div>
  <slot v-else />
</template>
