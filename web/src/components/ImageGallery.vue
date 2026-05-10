<script setup lang="ts">
import { ref } from 'vue'
import { ImageIcon, CopyIcon, UploadIcon, CheckIcon } from 'lucide-vue-next'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { api } from '@/services/api'
import type { PostDraft } from '@/services/api'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  slug: string
  assets: PostDraft['assets']
  onInsert: (markdown: string) => void
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  uploaded: []
}>()

const { t } = useI18n()
const notify = useNotify()
const copiedName = ref<string | null>(null)

const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg']

function isImage(name: string): boolean {
  const ext = name.toLowerCase().slice(name.lastIndexOf('.'))
  return imageExts.includes(ext)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function copyMarkdown(name: string) {
  const md = `![${name}](${name})`
  navigator.clipboard.writeText(md).then(() => {
    copiedName.value = name
    notify.success(t.value.copied)
    setTimeout(() => { copiedName.value = null }, 2000)
  })
}

function insertImage(name: string) {
  props.onInsert(`![${name}](${name})`)
  emit('update:open', false)
}

async function handleUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.multiple = true
  input.onchange = async () => {
    const files = input.files
    if (!files?.length) return
    for (const file of Array.from(files)) {
      try {
        await api.uploadAsset(props.slug, file)
      } catch (e) {
        notify.error(e as any)
      }
    }
    emit('uploaded')
  }
  input.click()
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ t.gallery }}</DialogTitle>
      </DialogHeader>

      <div class="flex items-center justify-between mb-2">
        <span class="text-xs text-muted-foreground">{{ assets.length }} {{ assets.length === 1 ? 'file' : 'files' }}</span>
        <Button variant="outline" size="sm" class="h-7 text-xs" @click="handleUpload">
          <UploadIcon class="h-3 w-3 mr-1" />{{ t.assetUpload }}
        </Button>
      </div>

      <div v-if="assets.length === 0" class="py-8 text-center text-muted-foreground text-sm">
        <ImageIcon class="h-8 w-8 mx-auto mb-2 opacity-30" />
        {{ t.noImages }}
      </div>

      <div v-else class="grid grid-cols-3 gap-3 max-h-[400px] overflow-auto">
        <div
          v-for="asset in assets"
          :key="asset.name"
          class="group relative rounded-lg border border-border/60 overflow-hidden cursor-pointer hover:border-accent/60 transition-colors"
          @click="insertImage(asset.name)"
        >
          <!-- Thumbnail or icon -->
          <div class="aspect-square flex items-center justify-center bg-muted/30">
            <img
              v-if="isImage(asset.name)"
              :src="'/studio/preview/' + slug + '/' + asset.name"
              :alt="asset.name"
              class="w-full h-full object-cover"
              loading="lazy"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
            <ImageIcon v-else class="h-6 w-6 text-muted-foreground/40" />
          </div>

          <!-- Info -->
          <div class="p-2 space-y-0.5">
            <p class="text-[11px] truncate font-medium">{{ asset.name }}</p>
            <p class="text-[10px] text-muted-foreground">{{ formatSize(asset.size) }}</p>
          </div>

          <!-- Actions overlay -->
          <div class="absolute top-1 right-1 flex gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              variant="secondary" size="icon" class="h-6 w-6"
              :title="t.copyLink"
              @click.stop="copyMarkdown(asset.name)"
            >
              <CheckIcon v-if="copiedName === asset.name" class="h-3 w-3 text-ok" />
              <CopyIcon v-else class="h-3 w-3" />
            </Button>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
