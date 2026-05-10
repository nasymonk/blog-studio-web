import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  description?: string
  action?: { label: string, handler: () => void }
  duration: number
}

const toasts = ref<Toast[]>([])

export function useToast() {
  function add(toast: Omit<Toast, 'id'>) {
    const id = Math.random().toString(36).slice(2)
    const duration = toast.duration ?? (toast.type === 'error' ? Infinity : 5000)
    toasts.value.push({ ...toast, id, duration })
    // Keep max 3 visible
    if (toasts.value.length > 3) toasts.value.shift()
    // Auto-dismiss
    if (duration !== Infinity && duration > 0) {
      setTimeout(() => remove(id), duration)
    }
  }

  function remove(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function dismissAll() {
    toasts.value = []
  }

  return {
    toasts,
    success: (title: string, desc?: string) => add({ type: 'success', title, description: desc, duration: 5000 }),
    error: (title: string, desc?: string) => add({ type: 'error', title, description: desc, duration: Infinity }),
    warning: (title: string, desc?: string) => add({ type: 'warning', title, description: desc, duration: 6000 }),
    info: (title: string, desc?: string) => add({ type: 'info', title, description: desc, duration: 5000 }),
    add,
    remove,
    dismissAll,
  }
}
