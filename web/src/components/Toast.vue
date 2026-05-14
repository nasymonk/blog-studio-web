<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { CheckCircleIcon, XCircleIcon, AlertTriangleIcon, InfoIcon, XIcon } from 'lucide-vue-next'
import type { Toast } from '@/composables/useToast'

const props = defineProps<{ toast: Toast }>()
const emit = defineEmits<{ dismiss: [id: string] }>()

const iconMap = {
  success: CheckCircleIcon,
  error: XCircleIcon,
  warning: AlertTriangleIcon,
  info: InfoIcon,
}

const colorMap = {
  success: { bg: 'bg-emerald-50 dark:bg-emerald-950/40', border: 'border-emerald-200 dark:border-emerald-800', icon: 'text-emerald-500', bar: 'bg-emerald-500' },
  error: { bg: 'bg-red-50 dark:bg-red-950/40', border: 'border-red-200 dark:border-red-800', icon: 'text-red-500', bar: 'bg-red-500' },
  warning: { bg: 'bg-amber-50 dark:bg-amber-950/40', border: 'border-amber-200 dark:border-amber-800', icon: 'text-amber-500', bar: 'bg-amber-500' },
  info: { bg: 'bg-blue-50 dark:bg-blue-950/40', border: 'border-blue-200 dark:border-blue-800', icon: 'text-blue-500', bar: 'bg-blue-500' },
}

const progress = ref(100)
let animationFrame: number | null = null
let remainingMs = 0
let pauseTime = 0
let timerId: ReturnType<typeof setTimeout> | null = null

function startTimer() {
  const duration = props.toast.duration
  if (duration === Infinity || duration <= 0) return

  remainingMs = duration
  const startTime = Date.now()

  function tick() {
    if (!pauseTime) {
      const elapsed = Date.now() - startTime
      progress.value = Math.max(0, 100 - (elapsed / duration) * 100)
    }
    if (progress.value > 0) {
      animationFrame = requestAnimationFrame(tick)
    }
  }
  animationFrame = requestAnimationFrame(tick)

  timerId = setTimeout(() => {
    emit('dismiss', props.toast.id)
  }, duration)
}

function onMouseEnter() {
  if (timerId) {
    pauseTime = Date.now()
    clearTimeout(timerId)
    timerId = null
    // Calculate remaining time
    const elapsed = (100 - progress.value) / 100 * props.toast.duration
    remainingMs = Math.max(0, props.toast.duration - elapsed)
  }
}

function onMouseLeave() {
  if (pauseTime && remainingMs > 0) {
    pauseTime = 0
    timerId = setTimeout(() => {
      emit('dismiss', props.toast.id)
    }, remainingMs)
  }
}

onMounted(() => {
  startTimer()
})

function onDismiss() {
  if (animationFrame) cancelAnimationFrame(animationFrame)
  if (timerId) clearTimeout(timerId)
  emit('dismiss', props.toast.id)
}
</script>

<template>
  <div
    class="relative w-[360px] overflow-hidden rounded-lg border shadow-lg backdrop-blur-sm"
    :class="[colorMap[toast.type].bg, colorMap[toast.type].border]"
    role="alert"
    :aria-live="toast.type === 'error' ? 'assertive' : 'polite'"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
  >
    <div class="flex items-start gap-3 p-4">
      <component
        :is="iconMap[toast.type]"
        class="h-5 w-5 shrink-0 mt-0.5"
        :class="colorMap[toast.type].icon"
        aria-hidden="true"
      />
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-foreground leading-snug">{{ toast.title }}</p>
        <p v-if="toast.description" class="mt-1 text-xs text-muted-foreground leading-relaxed whitespace-pre-line">{{ toast.description }}</p>
        <button
          v-if="toast.action"
          class="mt-2 text-xs font-medium underline underline-offset-2 text-foreground hover:text-foreground/80 transition-colors"
          @click="toast.action.handler"
        >
          {{ toast.action.label }}
        </button>
      </div>
      <button
        class="shrink-0 h-6 w-6 rounded-md inline-flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        aria-label="关闭通知"
        @click="onDismiss"
      >
        <XIcon class="h-3.5 w-3.5" />
      </button>
    </div>

    <!-- Progress bar -->
    <div
      v-if="toast.duration !== Infinity && toast.duration > 0"
      class="absolute bottom-0 left-0 h-[3px] transition-none"
      :class="colorMap[toast.type].bar"
      :style="{ width: `${progress}%` }"
    />
  </div>
</template>
