<script setup lang="ts">
import type { Version } from '@/composables/useVersionHistory'
import { Button } from '@/components/ui/button'
import { XIcon, RotateCcwIcon, ClockIcon, FileTextIcon } from 'lucide-vue-next'

const props = defineProps<{
  versions: Version[]
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'restore', body: string): void
}>()

function relativeTime(date: Date): string {
  const now = Date.now()
  const diff = now - date.getTime()
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return `${seconds} 秒前`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  return `${days} 天前`
}

function preview(body: string): string {
  const text = body.replace(/[#*_`~\[\]()!>|]/g, '').trim()
  return text.length > 100 ? text.slice(0, 100) + '...' : text
}
</script>

<template>
  <Transition name="slide">
    <div
      v-if="open"
      class="fixed right-0 top-0 bottom-0 w-80 bg-background border-l border-border shadow-lg z-50 flex flex-col"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-border/60">
        <h3 class="text-sm font-semibold">版本历史</h3>
        <Button variant="ghost" size="icon" class="h-7 w-7" @click="emit('close')">
          <XIcon class="h-4 w-4" />
        </Button>
      </div>

      <!-- List -->
      <div class="flex-1 overflow-auto">
        <div v-if="versions.length === 0" class="flex flex-col items-center justify-center h-full text-muted-foreground text-xs gap-2">
          <ClockIcon class="h-8 w-8 opacity-30" />
          <span>暂无版本记录</span>
        </div>

        <div v-else class="divide-y divide-border/40">
          <div
            v-for="version in versions"
            :key="version.id"
            class="group px-4 py-3 hover:bg-accent/50 transition-colors"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs text-muted-foreground flex items-center gap-1">
                <ClockIcon class="h-3 w-3" />
                {{ relativeTime(version.timestamp) }}
              </span>
              <span class="text-[10px] text-muted-foreground/60 flex items-center gap-0.5">
                <FileTextIcon class="h-2.5 w-2.5" />
                {{ version.wordCount }} 字
              </span>
            </div>

            <p class="text-xs text-muted-foreground/80 leading-relaxed mb-2 line-clamp-2">
              {{ preview(version.body) }}
            </p>

            <Button
              variant="outline"
              size="sm"
              class="h-6 text-[11px] px-2 opacity-0 group-hover:opacity-100 transition-opacity"
              @click="emit('restore', version.body)"
            >
              <RotateCcwIcon class="h-3 w-3 mr-1" />恢复
            </Button>
          </div>
        </div>
      </div>
    </div>
  </Transition>

  <!-- Backdrop -->
  <Transition name="fade">
    <div
      v-if="open"
      class="fixed inset-0 bg-black/20 z-40"
      @click="emit('close')"
    />
  </Transition>
</template>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
