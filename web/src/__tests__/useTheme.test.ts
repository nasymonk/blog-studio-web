import { describe, it, expect, beforeEach, vi } from 'vitest'

// useTheme uses module-level state, so we reset the module between test groups
// and control localStorage before importing.

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.classList.remove('dark')
    vi.resetModules()
  })

  it('defaults to light theme when localStorage is empty', async () => {
    const { useTheme } = await import('../composables/useTheme')
    const { theme } = useTheme()
    expect(theme.value).toBe('light')
  })

  it('reads dark theme from localStorage on init', async () => {
    localStorage.setItem('blog-studio-theme', 'dark')
    const { useTheme } = await import('../composables/useTheme')
    const { theme } = useTheme()
    expect(theme.value).toBe('dark')
  })

  it('toggle switches light → dark', async () => {
    const { useTheme } = await import('../composables/useTheme')
    const { theme, toggle } = useTheme()
    expect(theme.value).toBe('light')
    toggle()
    expect(theme.value).toBe('dark')
  })

  it('toggle switches dark → light', async () => {
    localStorage.setItem('blog-studio-theme', 'dark')
    const { useTheme } = await import('../composables/useTheme')
    const { theme, toggle } = useTheme()
    expect(theme.value).toBe('dark')
    toggle()
    expect(theme.value).toBe('light')
  })

  it('toggle persists reactively (theme value mirrors toggle calls)', async () => {
    const { useTheme } = await import('../composables/useTheme')
    const { theme, toggle } = useTheme()
    expect(theme.value).toBe('light')
    toggle()
    expect(theme.value).toBe('dark')
    toggle()
    expect(theme.value).toBe('light')
  })

  it('sets dark class on documentElement', async () => {
    const { useTheme } = await import('../composables/useTheme')
    const { toggle } = useTheme()
    toggle()
    // useStorage's watcher fires synchronously via triggerRef in happy-dom
    await Promise.resolve()
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })
})
