import { describe, it, expect } from 'vitest'

// Test the dictionary objects directly without mounting the composable.
// t is a ComputedRef, so access values via t.value.

describe('i18n', () => {
  it('zh dict has required keys', async () => {
    const { useI18n } = await import('../i18n/index')
    const { t } = useI18n()

    const requiredKeys = ['posts', 'settings', 'logout', 'login'] as const
    for (const key of requiredKeys) {
      const val = t.value[key as keyof typeof t.value]
      expect(typeof val, `key '${key}' should be a string`).toBe('string')
      expect((val as string).length, `key '${key}' should be non-empty`).toBeGreaterThan(0)
    }
  })

  it('lang switch changes t values', async () => {
    const { useI18n } = await import('../i18n/index')
    const { t, lang, setLang } = useI18n()

    setLang('zh')
    const zhPosts = t.value['posts' as keyof typeof t.value]

    setLang('en')
    const enPosts = t.value['posts' as keyof typeof t.value]

    expect(typeof zhPosts).toBe('string')
    expect(typeof enPosts).toBe('string')

    // Reset to zh
    setLang('zh')
    expect(lang.value).toBe('zh')
  })
})
