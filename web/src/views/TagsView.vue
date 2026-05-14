<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { TagsIcon, PencilIcon, Trash2Icon, Loader2Icon, SearchIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter
} from '@/components/ui/dialog'

const store = useStore()
const { t } = useI18n()
const notify = useNotify()

const loading = ref(true)
const searchQuery = ref('')
const renamingTag = ref<string | null>(null)
const newName = ref('')
const deletingTag = ref<string | null>(null)
const processing = ref(false)

type TagInfo = { name: string; count: number }

const tags = computed<TagInfo[]>(() => {
  const map = new Map<string, number>()
  for (const post of store.posts) {
    for (const tag of (post.tags || [])) {
      map.set(tag, (map.get(tag) || 0) + 1)
    }
  }
  return [...map.entries()]
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
})

const filteredTags = computed(() => {
  if (!searchQuery.value) return tags.value
  const q = searchQuery.value.toLowerCase()
  return tags.value.filter(t => t.name.toLowerCase().includes(q))
})

async function loadPosts() {
  loading.value = true
  try { store.posts = await api.posts() }
  catch (e: any) { notify.error(e) }
  finally { loading.value = false }
}

function startRename(tag: string) {
  renamingTag.value = tag
  newName.value = tag
}

async function confirmRename() {
  if (!renamingTag.value || !newName.value || renamingTag.value === newName.value) return
  processing.value = true
  try {
    await api.renameTag(renamingTag.value, newName.value)
    notify.success(t.value.tagRenamed(renamingTag.value, newName.value))
    renamingTag.value = null
    await loadPosts()
  } catch (e: any) { notify.error(e) }
  finally { processing.value = false }
}

function startDelete(tag: string) {
  deletingTag.value = tag
}

async function confirmDelete() {
  if (!deletingTag.value) return
  processing.value = true
  try {
    await api.deleteTag(deletingTag.value)
    notify.success(t.value.tagDeleted(deletingTag.value))
    deletingTag.value = null
    await loadPosts()
  } catch (e: any) { notify.error(e) }
  finally { processing.value = false }
}

onMounted(loadPosts)
</script>

<template>
  <div class="flex flex-col gap-5 max-w-3xl">
    <!-- Header -->
    <div class="flex items-center gap-3">
      <TagsIcon class="h-5 w-5 text-muted-foreground" />
      <h2 class="font-serif font-semibold text-lg">{{ t.tagManagement }}</h2>
      <span class="text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground">{{ tags.length }}</span>
    </div>

    <!-- Search -->
    <div class="relative max-w-[300px]">
      <label for="tags-search" class="sr-only">{{ t.search }}</label>
      <SearchIcon class="absolute left-0 bottom-3 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
      <Input
        id="tags-search"
        v-model="searchQuery"
        class="border-0 border-b border-border rounded-none bg-transparent pl-6 pb-2 h-auto focus-visible:ring-2 focus-visible:ring-accent transition-colors"
        :placeholder="t.search"
      />
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-14 animate-fade-up" />
    </div>

    <!-- Empty -->
    <div v-else-if="filteredTags.length === 0" class="text-center py-20 text-muted-foreground space-y-3">
      <TagsIcon class="h-10 w-10 mx-auto opacity-30" />
      <p class="font-serif font-medium">{{ searchQuery ? t.noResults : t.noTags }}</p>
    </div>

    <!-- Tag list -->
    <div v-else class="space-y-3">
      <div
        v-for="tag in filteredTags" :key="tag.name"
        class="group flex items-center justify-between gap-3 px-4 py-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
      >
        <div class="flex items-center gap-3 min-w-0">
          <span class="font-deco text-sm font-medium truncate"># {{ tag.name }}</span>
          <span class="text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground">{{ t.tagUsage(tag.count) }}</span>
        </div>
        <div class="flex gap-1 shrink-0 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
          <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" :title="t.renameTag" @click="startRename(tag.name)">
            <PencilIcon class="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" class="h-7 w-7 text-destructive/60 hover:text-destructive" :title="t.deleteTag" @click="startDelete(tag.name)">
            <Trash2Icon class="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Rename dialog -->
    <Dialog :open="!!renamingTag" @update:open="(v: boolean) => { if (!v) renamingTag = null }">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{{ t.renameTag }}</DialogTitle>
          <DialogDescription>{{ t.renameTagConfirm(renamingTag || '', newName) }}</DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <Input v-model="newName" :placeholder="t.newTagName" autofocus @keydown.enter="confirmRename" />
        </div>
        <DialogFooter>
          <Button variant="ghost" @click="renamingTag = null">{{ t.cancel }}</Button>
          <Button :disabled="!newName || newName === renamingTag || processing" @click="confirmRename">
            <Loader2Icon v-if="processing" class="h-3 w-3 animate-spin mr-1" />
            {{ t.saveSettings }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete confirmation dialog -->
    <Dialog :open="!!deletingTag" @update:open="(v: boolean) => { if (!v) deletingTag = null }">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{{ t.deleteTag }}</DialogTitle>
          <DialogDescription>{{ t.deleteTagConfirm(deletingTag || '') }}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="ghost" @click="deletingTag = null">{{ t.cancel }}</Button>
          <Button variant="destructive" :disabled="processing" @click="confirmDelete">
            <Loader2Icon v-if="processing" class="h-3 w-3 animate-spin mr-1" />
            {{ t.purge }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
