import { describe, it, expect } from 'vitest'
import { createStore } from '../store/index'

describe('createStore', () => {
  it('returns unauthenticated session by default', () => {
    const store = createStore()
    expect(store.session.authenticated).toBe(false)
    expect(store.session.csrfToken).toBe('')
  })

  it('session state can be updated reactively', () => {
    const store = createStore()
    store.session = { authenticated: true, csrfToken: 'abc-token' }
    expect(store.session.authenticated).toBe(true)
    expect(store.session.csrfToken).toBe('abc-token')
  })

  it('posts array is empty by default', () => {
    const store = createStore()
    expect(store.posts).toEqual([])
    expect(store.postsLoading).toBe(false)
  })

  it('posts can be added and accessed', () => {
    const store = createStore()
    const post = {
      slug: 'hello-world',
      title: 'Hello World',
      date: '2026-01-01',
      draft: false,
      tags: ['go', 'hugo'],
      categories: [],
      syncStatus: 'clean',
      remoteMtime: '',
      cachedRemoteMtime: '',
      latestBackupId: '',
      large: false,
    }
    store.posts.push(post)
    expect(store.posts).toHaveLength(1)
    expect(store.posts[0].slug).toBe('hello-world')
    expect(store.posts[0].tags).toContain('go')
  })

  it('posts can be replaced wholesale', () => {
    const store = createStore()
    store.posts = [
      { slug: 'a', title: 'A', date: '', draft: true, tags: [], categories: [], syncStatus: 'clean', remoteMtime: '', cachedRemoteMtime: '', latestBackupId: '', large: false },
      { slug: 'b', title: 'B', date: '', draft: false, tags: [], categories: [], syncStatus: 'stale', remoteMtime: '', cachedRemoteMtime: '', latestBackupId: '', large: false },
    ]
    expect(store.posts).toHaveLength(2)
    store.posts = []
    expect(store.posts).toHaveLength(0)
  })

  it('editor state tracks dirty flag', () => {
    const store = createStore()
    expect(store.editor.dirty).toBe(false)
    store.editor.dirty = true
    expect(store.editor.dirty).toBe(true)
  })

  it('editor saving flag is false by default', () => {
    const store = createStore()
    expect(store.editor.saving).toBe(false)
    store.editor.saving = true
    expect(store.editor.saving).toBe(true)
    store.editor.saving = false
    expect(store.editor.saving).toBe(false)
  })

  it('editor savedAt is null by default and can be set', () => {
    const store = createStore()
    expect(store.editor.savedAt).toBeNull()
    const now = new Date()
    store.editor.savedAt = now
    expect(store.editor.savedAt).toBe(now)
  })

  it('config starts null and can be set', () => {
    const store = createStore()
    expect(store.config).toBeNull()
    store.config = {
      basePath: '/studio',
      site: { id: 's1', name: 'My Blog', theme: 'PaperMod', blogRoot: '/blog', contentRoot: '/blog/content', postSection: 'post', postRoot: '/blog/content/post', buildCommand: 'hugo --minify', publicRoot: '/blog/public' },
      preview: { ttlMinutes: 120 },
      wechat: { enabled: false },
      comments: { enabled: false, twikooUrl: '', adminUrl: '' },
      maxUploadBytes: 10485760,
    }
    expect(store.config.basePath).toBe('/studio')
    expect(store.config.site.name).toBe('My Blog')
  })

  it('banner starts null and can be set or cleared', () => {
    const store = createStore()
    expect(store.banner).toBeNull()
    store.banner = { type: 'warn', message: 'Service degraded' }
    expect(store.banner?.message).toBe('Service degraded')
    store.banner = null
    expect(store.banner).toBeNull()
  })

  it('postsLoading can be toggled', () => {
    const store = createStore()
    expect(store.postsLoading).toBe(false)
    store.postsLoading = true
    expect(store.postsLoading).toBe(true)
  })
})
