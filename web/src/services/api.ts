export type AppError = {
  code: string
  message: string
  technicalDetail: string
  suggestion: string
  retryable: boolean
}

export type PostState = {
  slug: string
  title: string
  date: string
  draft: boolean
  tags: string[]
  categories: string[]
  syncStatus: string
  remoteMtime: string
  cachedRemoteMtime: string
  latestBackupId: string
  large: boolean
}

export type PostDraft = {
  slug: string
  frontMatter: {
    title: string
    date?: string
    draft: boolean
    tags: string[]
    categories: string[]
    description?: string
    image?: string
    math?: boolean
    scheduledAt?: string
  }
  body: string
  assets: Array<{ name: string; size: number }>
  large?: boolean
}

export type PublishResult = {
  target: string
  status: string
  backupId?: string
  conflicts?: Array<{ message: string; remoteMtime: string; cachedRemoteMtime: string }>
  auditId: string
  buildResult?: { success: boolean; exitCode: number; stdout: string; stderr: string }
  channelResult?: unknown
  error?: AppError
}

export type PreviewResult = {
  previewId: string
  url: string
  expiresAt: string
  buildResult: { success: boolean; exitCode: number; stderr: string }
  error?: AppError
}

export type Config = {
  basePath: string
  site: {
    id: string
    name: string
    theme: string
    blogRoot: string
    contentRoot: string
    postSection: string
    postRoot: string
    buildCommand: string
    publicRoot: string
  }
  preview: { ttlMinutes: number }
  wechat: { enabled: boolean; appId?: string }
  comments: { enabled: boolean; twikooUrl: string; adminUrl: string; dataPath?: string }
  maxUploadBytes: number
}

export type SiteInfo = {
  description: string
  profileImage: string
}

export type PostStats = {
  views: number
  lastViewed: string
  readingTime: number
}

export type TrashItem = {
  id: string
  slug: string
  siteId: string
  deletedAt: string
  size: number
}

export type AuditEntry = {
  auditId: string
  timestamp: string
  actor: string
  siteId: string
  slug: string
  operation: string
  target: string
  result: string
  backupId: string
  diffPath: string
  errorCode: string
  errorBrief: string
  buildResult: { success: boolean; exitCode: number; stdout: string; stderr: string }
}

let csrfToken = ''
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(handler: () => void) { onUnauthorized = handler }
const base = `${window.location.pathname.split('/').filter(Boolean)[0] ? '/' + window.location.pathname.split('/').filter(Boolean)[0] : ''}/api`

// Post list SWR cache
let postListCache: { data: PostState[]; timestamp: number } | null = null
const POST_LIST_TTL = 30_000 // 30 seconds

async function request<T>(path: string, options: RequestInit = {}, signal?: AbortSignal): Promise<T> {
  const headers = new Headers(options.headers)
  if (!(options.body instanceof FormData)) headers.set('Content-Type', 'application/json')
  if (csrfToken && options.method && options.method !== 'GET') headers.set('X-CSRF-Token', csrfToken)
  const response = await fetch(`${base}${path}`, { credentials: 'same-origin', signal, ...options, headers })
  if (response.status === 401 && path !== '/auth/login' && path !== '/session') {
    if (onUnauthorized) onUnauthorized()
    return new Promise(() => {})
  }
  const payload = await response.json().catch(() => ({ ok: false, error: { message: response.statusText } }))
  if (!response.ok || !payload.ok) {
    const err = payload.error || new Error(response.statusText)
    if (response.status) (err as any).status = response.status
    throw err
  }
  return payload.data as T
}

// Cache invalidation helper
function invalidatePostCache() {
  postListCache = null
}

export const api = {
  setCSRF: (value: string) => { csrfToken = value },
  session: () => request<{ authenticated: boolean; csrfToken?: string }>('/session'),
  login: (password: string) => request<{ authenticated: boolean; csrfToken: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request<{ authenticated: boolean }>('/auth/logout', { method: 'POST' }),
  changePassword: (currentPassword: string, newPassword: string) => request<{ changed: boolean }>('/auth/password', { method: 'POST', body: JSON.stringify({ currentPassword, newPassword }) }),
  posts: async (): Promise<PostState[]> => {
    if (postListCache && Date.now() - postListCache.timestamp < POST_LIST_TTL) {
      // Background revalidation
      request<PostState[]>('/posts').then(data => {
        postListCache = { data, timestamp: Date.now() }
      }).catch(() => {})
      return postListCache.data
    }
    const data = await request<PostState[]>('/posts')
    postListCache = { data, timestamp: Date.now() }
    return data
  },
  post: (slug: string) => request<PostDraft>(`/posts/${encodeURIComponent(slug)}`),
  saveDraft: async (draft: PostDraft) => {
    const result = await request<{ saved: boolean }>(`/posts/${encodeURIComponent(draft.slug)}/draft`, { method: 'PUT', body: JSON.stringify(draft) })
    invalidatePostCache()
    return result
  },
  publishBlog: async (draft: PostDraft, confirmOverwrite = false, signal?: AbortSignal) => {
    const result = await request<PublishResult>(`/posts/${encodeURIComponent(draft.slug)}/publish/blog`, { method: 'POST', body: JSON.stringify({ slug: draft.slug, draft, confirmOverwrite }) }, signal)
    invalidatePostCache()
    return result
  },
  preview: (draft: PostDraft, signal?: AbortSignal) => request<PreviewResult>(`/posts/${encodeURIComponent(draft.slug)}/preview`, { method: 'POST', body: JSON.stringify(draft) }, signal),
  rollback: async (slug: string) => {
    const result = await request<PublishResult>(`/posts/${encodeURIComponent(slug)}/rollback`, { method: 'POST' })
    invalidatePostCache()
    return result
  },
  uploadAsset: (slug: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<{ name: string; size: number }>(`/posts/${encodeURIComponent(slug)}/assets`, { method: 'POST', body: form })
  },
  getSite: () => request<SiteInfo>('/site'),
  saveSite: (info: SiteInfo) => request<SiteInfo>('/site', { method: 'PUT', body: JSON.stringify(info) }),
  uploadAvatar: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<{ path: string }>('/site/avatar', { method: 'POST', body: form })
  },
  getNowPage: () => request<{ raw: string }>('/pages/now'),
  saveNowPage: (raw: string) => request<{ saved: boolean }>('/pages/now', { method: 'PUT', body: JSON.stringify({ raw }) }),
  health: () => request<{ status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> }>('/health/full'),
  audit: (limit = 50, operation?: string, search?: string) => {
    const params = new URLSearchParams({ limit: String(limit) })
    if (operation) params.set('operation', operation)
    if (search) params.set('search', search)
    return request<AuditEntry[]>(`/audit?${params.toString()}`)
  },
  config: () => request<Config>('/config'),
  saveConfig: (config: Config) => request<Config>('/config', { method: 'PUT', body: JSON.stringify(config) }),
  deletePost: async (slug: string) => {
    const result = await request<{ trashId: string }>(`/posts/${encodeURIComponent(slug)}`, { method: 'DELETE' })
    invalidatePostCache()
    return result
  },
  postStats: (slug: string) => request<PostStats>(`/posts/${encodeURIComponent(slug)}/stats`),
  trash: () => request<TrashItem[]>('/trash'),
  restoreTrash: async (id: string) => {
    const result = await request<{ restored: boolean }>(`/trash/${encodeURIComponent(id)}/restore`, { method: 'POST' })
    invalidatePostCache()
    return result
  },
  purgeTrash: (id: string) => request<{ purged: boolean }>(`/trash/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  bulkTrash: async (slugs: string[]) => {
    const result = await request<Array<{ slug: string; success: boolean; error?: string }>>('/posts/bulk/trash', { method: 'POST', body: JSON.stringify({ slugs }) })
    invalidatePostCache()
    return result
  },
  bulkPublish: async (slugs: string[]) => {
    const result = await request<Array<{ slug: string; status: string; error?: string }>>('/posts/bulk/publish', { method: 'POST', body: JSON.stringify({ slugs }) })
    invalidatePostCache()
    return result
  },
  renameTag: (oldName: string, newName: string) => request<{ updated: number }>('/tags/rename', { method: 'POST', body: JSON.stringify({ oldName, newName }) }),
  deleteTag: (name: string) => request<{ updated: number }>('/tags/delete', { method: 'POST', body: JSON.stringify({ name }) }),
  exportAll: async () => {
    const headers = new Headers()
    if (csrfToken) headers.set('X-CSRF-Token', csrfToken)
    const response = await fetch(`${base}/posts/export`, { credentials: 'same-origin', headers })
    if (response.status === 401) { if (onUnauthorized) onUnauthorized(); return }
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'blog-posts.zip'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  },
  importZip: async (file: File) => {
    const form = new FormData()
    form.append('file', file)
    const result = await request<{ imported: number }>('/posts/import', { method: 'POST', body: form })
    invalidatePostCache()
    return result
  },
  metrics: async (): Promise<string> => {
    const headers = new Headers()
    if (csrfToken) headers.set('X-CSRF-Token', csrfToken)
    const response = await fetch(`${base}/metrics`, { credentials: 'same-origin', headers })
    if (response.status === 401) {
      if (onUnauthorized) onUnauthorized()
      return new Promise(() => {})
    }
    return response.text()
  },
}
