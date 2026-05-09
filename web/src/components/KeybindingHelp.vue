<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

const open = defineModel<boolean>('open', { default: false })

const shortcuts = [
  { keys: '⌘S', description: 'Save draft' },
  { keys: '⌘B', description: 'Bold' },
  { keys: '⌘I', description: 'Italic' },
  { keys: '⌘K', description: 'Insert link' },
  { keys: '⌘F', description: 'Search' },
  { keys: 'Tab', description: 'Indent' },
  { keys: 'Shift+Tab', description: 'Outdent' },
  { keys: '?', description: 'Show this help' },
]
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Keyboard Shortcuts</DialogTitle>
      </DialogHeader>
      <div class="space-y-1">
        <div
          v-for="s in shortcuts"
          :key="s.keys"
          class="flex items-center justify-between py-1.5 text-sm"
        >
          <span class="text-muted-foreground">{{ s.description }}</span>
          <kbd class="inline-flex h-6 items-center rounded border border-border bg-muted px-1.5 font-mono text-[11px] font-medium text-foreground">
            {{ s.keys }}
          </kbd>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
