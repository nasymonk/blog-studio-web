import { reactive, provide, inject } from 'vue'
import type { PostState, Config, SiteInfo } from '@/services/api'

export interface EditorState {
  dirty: boolean
  saving: boolean
  savedAt: Date | null
}

export interface AppStore {
  session: { authenticated: boolean; csrfToken: string }
  posts: PostState[]
  postsLoading: boolean
  config: Config | null
  siteInfo: SiteInfo | null
  health: { status: string; checks: Array<{ name: string; status: string; message: string; technicalDetail: string; suggestion: string }> } | null
  audit: unknown[]
  editor: EditorState
  banner: { type: 'ok' | 'warn' | 'error'; message: string } | null
}

const STORE_KEY = Symbol('app-store')

export function createStore(): AppStore {
  return reactive<AppStore>({
    session: { authenticated: false, csrfToken: '' },
    posts: [],
    postsLoading: false,
    config: null,
    siteInfo: null,
    health: null,
    audit: [],
    editor: { dirty: false, saving: false, savedAt: null },
    banner: null,
  })
}

export function provideStore(store: AppStore) {
  provide(STORE_KEY, store)
}

export function useStore(): AppStore {
  const store = inject<AppStore>(STORE_KEY)
  if (!store) throw new Error('useStore called outside of app')
  return store
}
