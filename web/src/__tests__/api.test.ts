import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

function mockFetch(payload: unknown, status = 200) {
  return vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : `HTTP ${status}`,
    json: () => Promise.resolve(payload),
  } as Response)
}

describe('api', () => {
  beforeEach(() => {
    vi.resetModules()
    // Reset window.location to a known state
    Object.defineProperty(window, 'location', {
      value: { pathname: '/studio/', href: '' },
      writable: true,
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('posts() returns array of PostState on success', async () => {
    const { api } = await import('../services/api')
    mockFetch({ ok: true, data: [{ slug: 'hello', title: 'Hello', draft: false, syncStatus: 'clean', tags: [], categories: [], date: '', remoteMtime: '', cachedRemoteMtime: '', latestBackupId: '', large: false }] })
    const posts = await api.posts()
    expect(posts).toHaveLength(1)
    expect(posts[0].slug).toBe('hello')
  })

  it('throws AppError when server returns ok:false', async () => {
    const { api } = await import('../services/api')
    mockFetch({ ok: false, error: { code: 'NOT_FOUND', message: '文章不存在', technicalDetail: '', suggestion: '', retryable: false } })
    await expect(api.post('missing-slug')).rejects.toMatchObject({ code: 'NOT_FOUND', message: '文章不存在' })
  })

  it('throws on HTTP 500', async () => {
    const { api } = await import('../services/api')
    mockFetch({ ok: false, error: { code: 'INTERNAL', message: '服务器错误', technicalDetail: '', suggestion: '', retryable: true } }, 500)
    await expect(api.posts()).rejects.toMatchObject({ code: 'INTERNAL' })
  })

  it('sets X-CSRF-Token header on POST after setCSRF', async () => {
    const { api } = await import('../services/api')
    api.setCSRF('test-csrf-token')
    const spy = mockFetch({ ok: true, data: { authenticated: true } })
    await api.logout()
    const headers = spy.mock.calls[0][1]?.headers as Headers
    expect(headers.get('X-CSRF-Token')).toBe('test-csrf-token')
  })

  it('does NOT set X-CSRF-Token on GET requests', async () => {
    const { api } = await import('../services/api')
    api.setCSRF('token-xyz')
    const spy = mockFetch({ ok: true, data: [] })
    await api.posts()
    const headers = spy.mock.calls[0][1]?.headers as Headers
    expect(headers.get('X-CSRF-Token')).toBeNull()
  })

  it('handles JSON parse failure gracefully', async () => {
    const { api } = await import('../services/api')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 502,
      statusText: 'Bad Gateway',
      json: () => Promise.reject(new SyntaxError('invalid json')),
    } as Response)
    await expect(api.posts()).rejects.toMatchObject({ message: 'Bad Gateway' })
  })

  it('redirects to /login on 401 for authenticated endpoints', async () => {
    const { api } = await import('../services/api')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: () => Promise.resolve({ ok: false, error: { message: 'Unauthorized' } }),
    } as Response)
    await expect(api.posts()).rejects.toThrow('unauthenticated')
    expect(window.location.href).toBe('/studio/#/login')
  })

  it('login() does NOT redirect on 401', async () => {
    const { api } = await import('../services/api')
    mockFetch({ ok: false, error: { code: 'WRONG_PASSWORD', message: '密码错误', technicalDetail: '', suggestion: '', retryable: false } }, 401)
    await expect(api.login('wrong')).rejects.toMatchObject({ code: 'WRONG_PASSWORD' })
    expect(window.location.href).not.toBe('/studio/#/login')
  })

  it('session() returns authenticated status', async () => {
    const { api } = await import('../services/api')
    mockFetch({ ok: true, data: { authenticated: true, csrfToken: 'csrf-abc' } })
    const session = await api.session()
    expect(session.authenticated).toBe(true)
    expect(session.csrfToken).toBe('csrf-abc')
  })

  it('config() returns Config object', async () => {
    const { api } = await import('../services/api')
    const cfg = { basePath: '/studio', site: { id: 's1', name: 'Blog', theme: 'PaperMod', blogRoot: '/blog', contentRoot: '', postSection: 'post', postRoot: '', buildCommand: 'hugo', publicRoot: '' }, preview: { ttlMinutes: 60 }, wechat: { enabled: false }, comments: { enabled: false, twikooUrl: '', adminUrl: '' }, maxUploadBytes: 5242880 }
    mockFetch({ ok: true, data: cfg })
    const result = await api.config()
    expect(result.basePath).toBe('/studio')
    expect(result.site.name).toBe('Blog')
  })
})
