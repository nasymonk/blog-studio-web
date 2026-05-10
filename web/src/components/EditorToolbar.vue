<script setup lang="ts">
import {
  BoldIcon, ItalicIcon, StrikethroughIcon, CodeIcon, LinkIcon,
  Heading1Icon, Heading2Icon, TextQuoteIcon,
  ListIcon, ListOrderedIcon, ListChecksIcon,
  ImageIcon, TableIcon, MinusIcon,
  FileCodeIcon, Columns2Icon, MonitorIcon,
  GalleryHorizontalEndIcon, KeyboardIcon,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import type { SplitMode } from '@/composables/useSplitView'

defineProps<{
  splitMode: SplitMode
}>()

const emit = defineEmits<{
  bold: []
  italic: []
  strikethrough: []
  code: []
  link: []
  heading1: []
  heading2: []
  blockquote: []
  unorderedList: []
  orderedList: []
  taskList: []
  image: []
  table: []
  hr: []
  setSplitMode: [mode: SplitMode]
  gallery: []
  keybindings: []
}>()
</script>

<template>
  <div class="flex items-center flex-wrap gap-1 py-2">
    <!-- Text group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Bold (Cmd+B)" @click="emit('bold')">
      <BoldIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Italic (Cmd+I)" @click="emit('italic')">
      <ItalicIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Strikethrough" @click="emit('strikethrough')">
      <StrikethroughIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Inline code" @click="emit('code')">
      <CodeIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Link (Cmd+K)" @click="emit('link')">
      <LinkIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- Block group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Heading 1" @click="emit('heading1')">
      <Heading1Icon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Heading 2" @click="emit('heading2')">
      <Heading2Icon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Blockquote" @click="emit('blockquote')">
      <TextQuoteIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- List group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Unordered list" @click="emit('unorderedList')">
      <ListIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Ordered list" @click="emit('orderedList')">
      <ListOrderedIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Task list" @click="emit('taskList')">
      <ListChecksIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- Insert group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Insert image" @click="emit('image')">
      <ImageIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Insert table" @click="emit('table')">
      <TableIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Horizontal rule" @click="emit('hr')">
      <MinusIcon class="h-3.5 w-3.5" />
    </Button>

    <Separator orientation="vertical" class="h-5 mx-1" />

    <!-- View group -->
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="splitMode === 'source' ? 'bg-accent/10 text-accent' : 'text-muted-foreground hover:text-foreground'" title="Source only" @click="emit('setSplitMode', 'source')">
      <FileCodeIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="splitMode === 'split' ? 'bg-accent/10 text-accent' : 'text-muted-foreground hover:text-foreground'" title="Split view" @click="emit('setSplitMode', 'split')">
      <Columns2Icon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0" :class="splitMode === 'preview' ? 'bg-accent/10 text-accent' : 'text-muted-foreground hover:text-foreground'" title="Preview only" @click="emit('setSplitMode', 'preview')">
      <MonitorIcon class="h-3.5 w-3.5" />
    </Button>

    <div class="flex-1" />

    <!-- Right side actions -->
    <Separator orientation="vertical" class="h-5 mx-1" />

    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Gallery" @click="emit('gallery')">
      <GalleryHorizontalEndIcon class="h-3.5 w-3.5" />
    </Button>
    <Button variant="ghost" size="sm" class="h-7 w-7 p-0 text-muted-foreground hover:text-foreground" title="Keybindings" @click="emit('keybindings')">
      <KeyboardIcon class="h-3.5 w-3.5" />
    </Button>
  </div>
</template>
