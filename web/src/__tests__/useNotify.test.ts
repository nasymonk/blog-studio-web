import { describe, it, expect, vi, beforeEach } from 'vitest'
import { toast } from 'vue-sonner'
import { useNotify } from '../composables/useNotify'

describe('useNotify', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('success() calls toast.success with default 3s duration', () => {
    const { success } = useNotify()
    success('已保存')
    expect(toast.success).toHaveBeenCalledWith('已保存', { duration: 3000 })
  })

  it('success() accepts custom duration', () => {
    const { success } = useNotify()
    success('Done', 5000)
    expect(toast.success).toHaveBeenCalledWith('Done', { duration: 5000 })
  })

  it('warn() calls toast.warning with 6s duration', () => {
    const { warn } = useNotify()
    warn('注意事项')
    expect(toast.warning).toHaveBeenCalledWith('注意事项', { duration: 6000 })
  })

  it('error(string) calls toast.error sticky by default', () => {
    const { error } = useNotify()
    error('Something went wrong')
    expect(toast.error).toHaveBeenCalledWith('Something went wrong', expect.objectContaining({ duration: Infinity }))
  })

  it('error(string, false) calls toast.error with 8s duration', () => {
    const { error } = useNotify()
    error('Transient error', false)
    expect(toast.error).toHaveBeenCalledWith('Transient error', expect.objectContaining({ duration: 8000 }))
  })

  it('error(AppError) shows message with technicalDetail as description', () => {
    const { error } = useNotify()
    error({ code: 'HUGO_FAIL', message: '构建失败', technicalDetail: 'exit code 1', suggestion: '检查 hugo 安装', retryable: true })
    expect(toast.error).toHaveBeenCalledWith(
      '构建失败',
      expect.objectContaining({ description: 'exit code 1' })
    )
  })

  it('error(AppError) uses suggestion as action label', () => {
    const { error } = useNotify()
    error({ code: 'ERR', message: 'Failed', technicalDetail: '', suggestion: '重启服务', retryable: false })
    const call = vi.mocked(toast.error).mock.calls[0]
    expect((call[1]?.action as any)?.label).toBe('重启服务')
  })

  it('error(err, { onRetry }) adds retry action button', () => {
    const { error } = useNotify()
    const retryFn = vi.fn()
    error('API error', { onRetry: retryFn })
    const call = vi.mocked(toast.error).mock.calls[0]
    expect((call[1]?.action as any)?.label).toBe('重试')
    ;(call[1]?.action as any)?.onClick()
    expect(retryFn).toHaveBeenCalledOnce()
  })

  it('dismiss() calls toast.dismiss', () => {
    const { dismiss } = useNotify()
    dismiss()
    expect(toast.dismiss).toHaveBeenCalled()
  })
})
