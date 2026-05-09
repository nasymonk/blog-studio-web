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

export const api = {
  setCSRF: (value: string) => { csrfToken = value },
  session: () => request<{ authenticated: boolean; csrfToken?: string }>('/session'),
  login: (password: string) => request<{ authenticated: boolean; csrfToken: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request<{ authenticated: boolean }>('/auth/logout', { method: 'POST' }),
  changePassword: (currentPassword: string, newPassword: string) => request<{ changed: boolean }>('/auth/password', { method: 'POST', body: JSON.stringify({ currentPassword, newPassword }) }),
  posts: () => request<PostState[]>('/posts'),
  post: (slug: string) => request<PostDraft>(`/posts/${encodeURIComponent(slug)}`),
  saveDraft: (draft: PostDraft) => request<{ saved: boolean }>(`/posts/${encodeURIComponent(draft.slug)}/draft`, { method: 'PUT', body: JSON.stringify(draft) }),
  publishBlog: (draft: PostDraft, confirmOverwrite = false, signal?: AbortSignal) => request<PublishResult>(`/posts/${encodeURIComponent(draft.slug)}/publish/blog`, { method: 'POST', body: JSON.stringify({ slug: draft.slug, draft, confirmOverwrite }) }, signal),
  preview: (draft: PostDraft, signal?: AbortSignal) => request<PreviewResult>(`/posts/${encodeURIComponent(draft.slug)}/preview`, { method: 'POST', body: JSON.stringify(draft) }, signal),
  rollback: (slug: string) => request<PublishResult>(`/posts/${encodeURIComponent(slug)}/rollback`, { method: 'POST' }),
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
  audit: (limit = 50) => request<AuditEntry[]>(`/audit?limit=${limit}`),
  config: () => request<Config>('/config'),
  saveConfig: (config: Config) => request<Config>('/config', { method: 'PUT', body: JSON.stringify(config) }),
  deletePost: (slug: string) => request<{ trashId: string }>(`/posts/${encodeURIComponent(slug)}`, { method: 'DELETE' }),
  trash: () => request<TrashItem[]>('/trash'),
  restoreTrash: (id: string) => request<{ restored: boolean }>(`/trash/${encodeURIComponent(id)}/restore`, { method: 'POST' }),
  purgeTrash: (id: string) => request<{ purged: boolean }>(`/trash/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  bulkTrash: (slugs: string[]) => request<Array<{ slug: string; success: boolean; error?: string }>>('/posts/bulk/trash', { method: 'POST', body: JSON.stringify({ slugs }) }),
  bulkPublish: (slugs: string[]) => request<Array<{ slug: string; status: string; error?: string }>>('/posts/bulk/publish', { method: 'POST', body: JSON.stringify({ slugs }) }),
  renameTag: (oldName: string, newName: string) => request<{ updated: number }>('/tags/rename', { method: 'POST', body: JSON.stringify({ oldName, newName }) }),
  deleteTag: (name: string) => request<{ updated: number }>('/tags/delete', { method: 'POST', body: JSON.stringify({ name }) }),
}
