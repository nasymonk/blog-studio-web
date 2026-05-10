import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useNotify } from '../composables/useNotify'
import { useToast } from '../composables/useToast'

describe('useNotify', () => {
  beforeEach(() => {
    const { dismissAll } = useToast()
    dismissAll()
    vi.clearAllMocks()
  })

  it('success() adds a success toast with default 3s duration', () => {
    const { success } = useNotify()
    success('已保存')
    const { toasts } = useToast()
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].type).toBe('success')
    expect(toasts.value[0].title).toBe('已保存')
    expect(toasts.value[0].duration).toBe(3000)
  })

  it('success() accepts custom duration', () => {
    const { success } = useNotify()
    success('Done', 5000)
    const { toasts } = useToast()
    expect(toasts.value[0].duration).toBe(5000)
  })

  it('warn() adds a warning toast with 6s duration', () => {
    const { warn } = useNotify()
    warn('注意事项')
    const { toasts } = useToast()
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].type).toBe('warning')
    expect(toasts.value[0].title).toBe('注意事项')
    expect(toasts.value[0].duration).toBe(6000)
  })

  it('error(string) adds an error toast sticky by default', () => {
    const { error } = useNotify()
    error('Something went wrong')
    const { toasts } = useToast()
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].type).toBe('error')
    expect(toasts.value[0].title).toBe('Something went wrong')
    expect(toasts.value[0].duration).toBe(Infinity)
  })

  it('error(string, false) adds an error toast with 8s duration', () => {
    const { error } = useNotify()
    error('Transient error', false)
    const { toasts } = useToast()
    expect(toasts.value[0].duration).toBe(8000)
  })

  it('error(AppError) shows technicalDetail and suggestion in description', () => {
    const { error } = useNotify()
    error({ code: 'HUGO_FAIL', message: '构建失败', technicalDetail: 'exit code 1', suggestion: '检查 hugo 安装', retryable: true })
    const { toasts } = useToast()
    expect(toasts.value[0].title).toBe('构建失败')
    expect(toasts.value[0].description).toBe('exit code 1\n建议：检查 hugo 安装')
  })

  it('error(AppError) with suggestion only shows suggestion in description', () => {
    const { error } = useNotify()
    error({ code: 'ERR', message: 'Failed', technicalDetail: '', suggestion: '重启服务', retryable: false })
    const { toasts } = useToast()
    expect(toasts.value[0].description).toBe('建议：重启服务')
    expect(toasts.value[0].action).toBeUndefined()
  })

  it('error(err, { onRetry }) adds retry action button', () => {
    const { error } = useNotify()
    const retryFn = vi.fn()
    error('API error', { onRetry: retryFn })
    const { toasts } = useToast()
    expect(toasts.value[0].action?.label).toBe('重试')
    toasts.value[0].action?.handler()
    expect(retryFn).toHaveBeenCalledOnce()
  })

  it('dismiss() clears all toasts', () => {
    const { success, dismiss } = useNotify()
    success('one')
    success('two')
    const { toasts } = useToast()
    expect(toasts.value).toHaveLength(2)
    dismiss()
    expect(toasts.value).toHaveLength(0)
  })
})
