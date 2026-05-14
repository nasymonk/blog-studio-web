<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import {
  SearchIcon, XIcon, ChevronUpIcon, ChevronDownIcon,
  ReplaceIcon, TextIcon,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

defineProps<{
  searchText: string
  replaceText: string
  caseSensitive: boolean
  wholeWord: boolean
  useRegex: boolean
  matchCount: number
  currentMatch: number
}>()

const emit = defineEmits<{
  'update:searchText': [value: string]
  'update:replaceText': [value: string]
  'update:caseSensitive': [value: boolean]
  'update:wholeWord': [value: boolean]
  'update:useRegex': [value: boolean]
  close: []
  findNext: []
  findPrev: []
  replace: []
  replaceAll: []
}>()

const searchInputRef = ref<HTMLInputElement | null>(null)
const showReplace = ref(false)

onMounted(() => {
  nextTick(() => searchInputRef.value?.focus())
})

function onSearchKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    emit('findNext')
  }
  if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}

function onReplaceKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    emit('replace')
  }
  if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}

function toggleReplace() {
  showReplace.value = !showReplace.value
}
</script>

<template>
  <div
    class="flex flex-col bg-muted/50 border-b border-border px-2 py-1.5 gap-1 animate-fade-in"
    @keydown.escape="$emit('close')"
  >
    <!-- Search row -->
    <div class="flex items-center gap-1.5">
      <button
        class="flex items-center justify-center h-6 w-6 rounded text-muted-foreground hover:text-foreground hover:bg-accent/50 transition-colors cursor-pointer"
        aria-label="Toggle replace"
        title="Toggle replace"
        @click="toggleReplace"
      >
        <ChevronDownIcon
          class="h-3 w-3 transition-transform"
          :class="{ 'rotate-90': showReplace }"
          aria-hidden="true"
        />
      </button>

      <div class="relative flex-1 flex items-center">
        <SearchIcon class="absolute left-2 h-3.5 w-3.5 text-muted-foreground/50 pointer-events-none" />
        <label for="find-search" class="sr-only">Search</label>
        <input
          id="find-search"
          ref="searchInputRef"
          :value="searchText"
          class="w-full h-7 pl-7 pr-2 text-sm bg-background border border-input rounded focus:outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/40"
          placeholder="Search..."
          @input="emit('update:searchText', ($event.target as HTMLInputElement).value)"
          @keydown="onSearchKeydown"
        />
      </div>

      <!-- Toggle buttons -->
      <div class="flex items-center gap-0.5">
        <button
          class="flex items-center justify-center h-6 w-6 rounded text-xs font-mono font-semibold transition-colors cursor-pointer"
          :class="caseSensitive ? 'bg-primary/15 text-primary' : 'text-muted-foreground/60 hover:text-foreground hover:bg-accent/50'"
          aria-label="Case sensitive (Aa)"
          title="Case sensitive (Aa)"
          @click="emit('update:caseSensitive', !caseSensitive)"
        >Aa</button>
        <button
          class="flex items-center justify-center h-6 w-6 rounded text-xs font-mono font-semibold transition-colors cursor-pointer"
          :class="wholeWord ? 'bg-primary/15 text-primary' : 'text-muted-foreground/60 hover:text-foreground hover:bg-accent/50'"
          aria-label="Whole word"
          title="Whole word"
          @click="emit('update:wholeWord', !wholeWord)"
        ><TextIcon class="h-3 w-3" aria-hidden="true" /></button>
        <button
          class="flex items-center justify-center h-6 w-6 rounded text-xs font-mono font-semibold transition-colors cursor-pointer"
          :class="useRegex ? 'bg-primary/15 text-primary' : 'text-muted-foreground/60 hover:text-foreground hover:bg-accent/50'"
          aria-label="Regular expression"
          title="Regular expression"
          @click="emit('update:useRegex', !useRegex)"
        >.*</button>
      </div>

      <!-- Match count -->
      <span class="text-[11px] text-muted-foreground/60 tabular-nums min-w-[3rem] text-center">
        {{ matchCount > 0 ? `${currentMatch}/${matchCount}` : searchText ? 'No results' : '' }}
      </span>

      <!-- Navigation -->
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 text-muted-foreground hover:text-foreground"
        :disabled="!matchCount"
        aria-label="Previous match (Shift+Enter)"
        title="Previous match (Shift+Enter)"
        @click="$emit('findPrev')"
      >
        <ChevronUpIcon class="h-3.5 w-3.5" aria-hidden="true" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 text-muted-foreground hover:text-foreground"
        :disabled="!matchCount"
        aria-label="Next match (Enter)"
        title="Next match (Enter)"
        @click="$emit('findNext')"
      >
        <ChevronDownIcon class="h-3.5 w-3.5" aria-hidden="true" />
      </Button>

      <!-- Close -->
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 text-muted-foreground hover:text-foreground"
        aria-label="Close (Escape)"
        title="Close (Escape)"
        @click="$emit('close')"
      >
        <XIcon class="h-3.5 w-3.5" aria-hidden="true" />
      </Button>
    </div>

    <!-- Replace row -->
    <div v-if="showReplace" class="flex items-center gap-1.5 pl-7">
      <div class="relative flex-1 flex items-center">
        <ReplaceIcon class="absolute left-2 h-3.5 w-3.5 text-muted-foreground/50 pointer-events-none" />
        <label for="find-replace" class="sr-only">Replace</label>
        <input
          id="find-replace"
          :value="replaceText"
          class="w-full h-7 pl-7 pr-2 text-sm bg-background border border-input rounded focus:outline-none focus:ring-1 focus:ring-ring placeholder:text-muted-foreground/40"
          placeholder="Replace..."
          @input="emit('update:replaceText', ($event.target as HTMLInputElement).value)"
          @keydown="onReplaceKeydown"
        />
      </div>
      <Button
        variant="outline"
        size="sm"
        class="h-7 text-xs px-2.5"
        :disabled="!matchCount"
        title="Replace current match"
        @click="$emit('replace')"
      >Replace</Button>
      <Button
        variant="outline"
        size="sm"
        class="h-7 text-xs px-2.5"
        :disabled="!matchCount"
        title="Replace all matches"
        @click="$emit('replaceAll')"
      >All</Button>
    </div>
  </div>
</template>
