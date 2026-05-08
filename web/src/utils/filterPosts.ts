import type { PostState } from '@/services/api'

export type { PostState }

export function filterPosts(
  posts: PostState[],
  query: string,
  statusFilter: string,
  selectedTags: string[],
  sortBy: string,
): PostState[] {
  let result = [...posts]

  const q = query.trim().toLowerCase()
  if (q) {
    result = result.filter(
      p =>
        p.title.toLowerCase().includes(q) ||
        p.slug.toLowerCase().includes(q) ||
        (p.tags || []).some(tag => tag.toLowerCase().includes(q)),
    )
  }

  if (statusFilter !== 'all') {
    result = result.filter(p => {
      if (statusFilter === 'drafts') return p.draft
      if (statusFilter === 'published') return !p.draft && p.syncStatus === 'clean'
      if (statusFilter === 'conflicts') return p.syncStatus === 'conflict'
      if (statusFilter === 'stale') return p.syncStatus === 'stale'
      return true
    })
  }

  if (selectedTags.length > 0) {
    result = result.filter(p => selectedTags.every(tag => (p.tags || []).includes(tag)))
  }

  result.sort((a, b) => {
    if (sortBy === 'date-asc') return a.date.localeCompare(b.date)
    if (sortBy === 'title') return a.title.localeCompare(b.title)
    if (sortBy === 'status') return a.syncStatus.localeCompare(b.syncStatus)
    return b.date.localeCompare(a.date)
  })

  return result
}
