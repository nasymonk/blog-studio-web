import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import PostsView from '@/views/PostsView.vue'
import type { PostState } from '@/services/api'
import type { AppStore } from '@/store'

// --- Shared mock store ---

const mockStore = reactive<AppStore>({
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

vi.mock('@/store', () => ({
  useStore: () => mockStore,
  createStore: () => mockStore,
  provideStore: () => {},
}))

// --- Mocks ---

const mockPush = vi.fn()
const mockReplace = vi.fn()
const mockRouteQuery = ref<Record<string, string>>({})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush, replace: mockReplace }),
  useRoute: () => ({ query: mockRouteQuery.value }),
}))

const mockPostsFn = vi.fn()
vi.mock('@/services/api', () => ({
  get api() {
    return { posts: mockPostsFn }
  },
}))

vi.mock('lucide-vue-next', () => ({
  PlusIcon: { template: '<span />' },
  SearchIcon: { template: '<span />' },
  RefreshCwIcon: { template: '<span />' },
  EditIcon: { template: '<span />' },
  EyeIcon: { template: '<span />' },
  AlertTriangleIcon: { template: '<span />' },
  FileTextIcon: { template: '<span />' },
  Trash2Icon: { template: '<span />' },
}))

vi.mock('@vueuse/core', () => ({
  useDebounceFn: (fn: Function) => fn,
  useStorage: (_key: string, defaultValue: unknown) => ref(defaultValue),
}))

vi.mock('@/components/ui/button', () => ({
  Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
}))
vi.mock('@/components/ui/input', () => ({
  Input: {
    template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
    props: ['modelValue'],
    emits: ['update:modelValue'],
  },
}))
vi.mock('@/components/ui/badge', () => ({
  Badge: { template: '<span><slot /></span>' },
}))
vi.mock('@/components/ui/select', () => ({
  Select: { template: '<div><slot /></div>' },
  SelectContent: { template: '<div><slot /></div>' },
  SelectItem: { template: '<div><slot /></div>', props: ['value'] },
  SelectTrigger: { template: '<div><slot /></div>' },
  SelectValue: { template: '<span />' },
}))
vi.mock('@/components/ui/skeleton', () => ({
  Skeleton: { template: '<div />' },
}))

// --- Helpers ---

function makePost(overrides: Partial<PostState> = {}): PostState {
  return {
    slug: 'test-post',
    title: 'Test Post',
    date: '2026-05-10T00:00:00Z',
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

function mountView() {
  const wrapper = mount(PostsView, {
    global: {
      stubs: { teleport: true },
    },
  })
  return wrapper
}

// --- Tests ---

describe('PostsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockPostsFn.mockResolvedValue([])
    mockRouteQuery.value = {}
    mockStore.posts = []
    mockStore.postsLoading = false
  })

  it('renders post list after loading', async () => {
    const posts = [
      makePost({ slug: 'first', title: 'First Post', tags: ['vue'] }),
      makePost({ slug: 'second', title: 'Second Post', tags: ['go'] }),
    ]
    mockPostsFn.mockImplementation(async () => {
      mockStore.posts = posts
      return posts
    })

    const wrapper = mountView()
    await flushPromises()

    const cards = wrapper.findAll('.post-card')
    expect(cards).toHaveLength(2)
    expect(cards[0].text()).toContain('First Post')
    expect(cards[1].text()).toContain('Second Post')
  })

  it('shows empty state when no posts', async () => {
    mockPostsFn.mockResolvedValue([])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('.post-card').exists()).toBe(false)
    expect(wrapper.text()).toContain('还没有文章')
  })

  it('search filters posts by title', async () => {
    const posts = [
      makePost({ slug: 'vue-guide', title: 'Vue Guide' }),
      makePost({ slug: 'go-intro', title: 'Go Introduction' }),
    ]
    mockPostsFn.mockImplementation(async () => {
      mockStore.posts = posts
      return posts
    })

    const wrapper = mountView()
    await flushPromises()

    const searchInput = wrapper.find('input')
    await searchInput.setValue('Vue')
    await flushPromises()

    const cards = wrapper.findAll('.post-card')
    expect(cards).toHaveLength(1)
    expect(cards[0].text()).toContain('Vue Guide')
  })

  it('search filters posts by slug', async () => {
    const posts = [
      makePost({ slug: 'hugo-tips', title: 'Hugo Tips' }),
      makePost({ slug: 'vue-guide', title: 'Vue Guide' }),
    ]
    mockPostsFn.mockImplementation(async () => {
      mockStore.posts = posts
      return posts
    })

    const wrapper = mountView()
    await flushPromises()

    const searchInput = wrapper.find('input')
    await searchInput.setValue('hugo')
    await flushPromises()

    const cards = wrapper.findAll('.post-card')
    expect(cards).toHaveLength(1)
    expect(cards[0].text()).toContain('Hugo Tips')
  })

  it('status filter shows only drafts', async () => {
    const posts = [
      makePost({ slug: 'published', title: 'Published Post', draft: false, syncStatus: 'clean' }),
      makePost({ slug: 'draft', title: 'Draft Post', draft: true, syncStatus: '' }),
    ]
    mockPostsFn.mockImplementation(async () => {
      mockStore.posts = posts
      return posts
    })

    const wrapper = mountView()
    await flushPromises()

    const filterButtons = wrapper.findAll('button').filter(b => b.text() === '草稿')
    expect(filterButtons.length).toBeGreaterThan(0)
    await filterButtons[0].trigger('click')
    await flushPromises()

    const cards = wrapper.findAll('.post-card')
    expect(cards).toHaveLength(1)
    expect(cards[0].text()).toContain('Draft Post')
  })

  it('calls api.posts() on mount', async () => {
    mountView()
    await flushPromises()
    expect(mockPostsFn).toHaveBeenCalledTimes(1)
  })
})
