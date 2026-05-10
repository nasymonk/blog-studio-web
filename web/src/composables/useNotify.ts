import { useToast } from '@/composables/useToast'
import type { AppError } from '@/services/api'

export interface NotifyErrorOptions {
  sticky?: boolean
  onRetry?: () => void
}

export function useNotify() {
  const toast = useToast()

  function success(message: string, duration = 3000) {
    toast.add({ type: 'success', title: message, duration })
  }

  function error(err: AppError | string, options: NotifyErrorOptions | boolean = true) {
    const sticky = typeof options === 'boolean' ? options : (options.sticky ?? true)
    const onRetry = typeof options === 'object' ? options.onRetry : undefined
    const duration = sticky ? Infinity : 8000

    const action = onRetry
      ? { label: '重试', handler: onRetry }
      : undefined

    if (typeof err === 'string') {
      toast.add({ type: 'error', title: err, duration, action })
      return
    }
    const description = [err.technicalDetail, err.suggestion && `建议：${err.suggestion}`].filter(Boolean).join('\n') || undefined
    toast.add({
      type: 'error',
      title: err.message,
      description,
      duration,
      action,
    })
  }

  function warn(message: string) {
    toast.add({ type: 'warning', title: message, duration: 6000 })
  }

  function dismiss() {
    toast.dismissAll()
  }

  return { success, error, warn, dismiss }
}
