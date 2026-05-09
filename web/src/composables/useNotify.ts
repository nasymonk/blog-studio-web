import { toast } from 'vue-sonner'
import type { AppError } from '@/services/api'

export interface NotifyErrorOptions {
  sticky?: boolean
  onRetry?: () => void
}

export function useNotify() {
  function success(message: string, duration = 3000) {
    toast.success(message, { duration })
  }

  function error(err: AppError | string, options: NotifyErrorOptions | boolean = true) {
    const sticky = typeof options === 'boolean' ? options : (options.sticky ?? true)
    const onRetry = typeof options === 'object' ? options.onRetry : undefined
    const duration = sticky ? Infinity : 8000

    const action = onRetry
      ? { label: '重试', onClick: onRetry }
      : undefined

    if (typeof err === 'string') {
      toast.error(err, { duration, action })
      return
    }
    const description = [err.technicalDetail, err.suggestion && `建议：${err.suggestion}`].filter(Boolean).join('\n') || undefined
    toast.error(err.message, {
      description,
      duration,
      action,
    })
  }

  function warn(message: string) {
    toast.warning(message, { duration: 6000 })
  }

  function dismiss() {
    toast.dismiss()
  }

  return { success, error, warn, dismiss }
}
