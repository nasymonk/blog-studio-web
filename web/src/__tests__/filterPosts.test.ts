import { describe, it, expect } from 'vitest'
import { filterPosts } from '../utils/filterPosts'
import type { PostState } from '../services/api'

function makePost(overrides: Partial<PostState> & Pick<PostState, 'title' | 'slug'>): PostState {
  return {
    date: '2024-01-01',
    draft: false,
    tags: [],
    categories: [],
    syncStatus: 'clean',
    remoteMtime: '',
    cachedRemoteMtime: '',
    latestBackupId: '',
    large: false,
    ...overrides,
  }
}

const posts: PostState[] = [
  makePost({ title: 'Hugo 学习笔记', slug: 'hugo-notes', date: '2024-01-15', tags: ['hugo', 'web'] }),
  makePost({ title: '日记：旅行', slug: 'diary-travel', date: '2024-03-01', draft: true, syncStatus: 'stale', tags: ['日记'] }),
  makePost({ title: 'JavaScript 进阶', slug: 'js-advanced', date: '2023-12-20', syncStatus: 'conflict', tags: ['js', 'web'] }),
  makePost({ title: 'Reading List', slug: 'reading', date: '2024-02-10' }),
]

describe('filterPosts — search', () => {
  it('returns all posts when query is empty', () => {
    expect(filterPosts(posts, '', 'all', [], 'date-desc')).toHaveLength(4)
  })

  it('matches on title (case-insensitive)', () => {
    const result = filterPosts(posts, 'hugo', 'all', [], 'date-desc')
    expect(result).toHaveLength(1)
    expect(result[0].slug).toBe('hugo-notes')
  })

  it('matches on slug', () => {
    const result = filterPosts(posts, 'diary', 'all', [], 'date-desc')
    expect(result).toHaveLength(1)
    expect(result[0].slug).toBe('diary-travel')
  })

  it('matches on tag', () => {
    expect(filterPosts(posts, 'web', 'all', [], 'date-desc')).toHaveLength(2)
  })

  it('returns empty when no match', () => {
    expect(filterPosts(posts, 'xyznotfound', 'all', [], 'date-desc')).toHaveLength(0)
  })
})

describe('filterPosts — status filter', () => {
  it('drafts', () => {
    const result = filterPosts(posts, '', 'drafts', [], 'date-desc')
    expect(result.every(p => p.draft)).toBe(true)
  })

  it('published (clean + not draft)', () => {
    const result = filterPosts(posts, '', 'published', [], 'date-desc')
    expect(result.every(p => !p.draft && p.syncStatus === 'clean')).toBe(true)
  })

  it('conflicts', () => {
    const result = filterPosts(posts, '', 'conflicts', [], 'date-desc')
    expect(result.every(p => p.syncStatus === 'conflict')).toBe(true)
  })

  it('stale', () => {
    const result = filterPosts(posts, '', 'stale', [], 'date-desc')
    expect(result.every(p => p.syncStatus === 'stale')).toBe(true)
  })
})

describe('filterPosts — tag filter', () => {
  it('single tag', () => {
    const result = filterPosts(posts, '', 'all', ['hugo'], 'date-desc')
    expect(result).toHaveLength(1)
    expect(result[0].slug).toBe('hugo-notes')
  })

  it('multiple tags (AND)', () => {
    const result = filterPosts(posts, '', 'all', ['web', 'js'], 'date-desc')
    expect(result).toHaveLength(1)
    expect(result[0].slug).toBe('js-advanced')
  })

  it('tag with no match', () => {
    expect(filterPosts(posts, '', 'all', ['nonexistent'], 'date-desc')).toHaveLength(0)
  })
})

describe('filterPosts — sorting', () => {
  it('date-desc (default) puts newest first', () => {
    const result = filterPosts(posts, '', 'all', [], 'date-desc')
    expect(result[0].date >= result[1].date).toBe(true)
  })

  it('date-asc puts oldest first', () => {
    const result = filterPosts(posts, '', 'all', [], 'date-asc')
    expect(result[0].date <= result[1].date).toBe(true)
  })

  it('title sorts alphabetically', () => {
    const result = filterPosts(posts, '', 'all', [], 'title')
    const titles = result.map(p => p.title)
    expect(titles).toEqual([...titles].sort())
  })

  it('status sorts by syncStatus', () => {
    const result = filterPosts(posts, '', 'all', [], 'status')
    const statuses = result.map(p => p.syncStatus)
    expect(statuses).toEqual([...statuses].sort())
  })
})
