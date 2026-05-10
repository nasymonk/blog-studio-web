import { ref } from 'vue'

export type SplitMode = 'source' | 'preview' | 'split'

export function useSplitView() {
  const mode = ref<SplitMode>(
    (localStorage.getItem('editor:splitMode') as SplitMode) || 'source'
  )
  const splitRatio = ref(0.5)
  const isDragging = ref(false)

  function setMode(newMode: SplitMode) {
    mode.value = newMode
    localStorage.setItem('editor:splitMode', newMode)
  }

  function onDragStart(e: MouseEvent) {
    isDragging.value = true
    const startX = e.clientX
    const startRatio = splitRatio.value
    const container = (e.target as HTMLElement).parentElement!
    const containerWidth = container.getBoundingClientRect().width

    function onMove(e: MouseEvent) {
      const delta = e.clientX - startX
      splitRatio.value = Math.max(0.2, Math.min(0.8, startRatio + delta / containerWidth))
    }

    function onUp() {
      isDragging.value = false
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  return { mode, splitRatio, isDragging, setMode, onDragStart }
}
