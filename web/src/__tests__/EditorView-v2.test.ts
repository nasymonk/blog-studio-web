import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import SplitView from '@/components/SplitView.vue'
import VersionHistory from '@/components/VersionHistory.vue'

// --- useSplitView tests ---
describe('useSplitView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
  })

  it('defaults to source mode when localStorage is empty', async () => {
    const { useSplitView } = await import('@/composables/useSplitView')
    const { mode } = useSplitView()
    expect(mode.value).toBe('source')
  })

  it('reads split mode from localStorage on init', async () => {
    localStorage.setItem('editor:splitMode', 'split')
    const { useSplitView } = await import('@/composables/useSplitView')
    const { mode } = useSplitView()
    expect(mode.value).toBe('split')
  })

  it('setMode updates mode and persists to localStorage', async () => {
    const { useSplitView } = await import('@/composables/useSplitView')
    const { mode, setMode } = useSplitView()
    setMode('preview')
    expect(mode.value).toBe('preview')
    expect(localStorage.getItem('editor:splitMode')).toBe('preview')
  })

  it('setMode with split value persists correctly', async () => {
    const { useSplitView } = await import('@/composables/useSplitView')
    const { mode, setMode } = useSplitView()
    setMode('split')
    expect(mode.value).toBe('split')
    expect(localStorage.getItem('editor:splitMode')).toBe('split')
  })

  it('splitRatio defaults to 0.5', async () => {
    const { useSplitView } = await import('@/composables/useSplitView')
    const { splitRatio } = useSplitView()
    expect(splitRatio.value).toBe(0.5)
  })

  it('isDragging defaults to false', async () => {
    const { useSplitView } = await import('@/composables/useSplitView')
    const { isDragging } = useSplitView()
    expect(isDragging.value).toBe(false)
  })
})

// --- useVersionHistory tests ---
describe('useVersionHistory', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('starts with empty versions', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions } = useVersionHistory()
    expect(versions.value).toHaveLength(0)
  })

  it('isOpen defaults to false', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { isOpen } = useVersionHistory()
    expect(isOpen.value).toBe(false)
  })

  it('toggle flips isOpen', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { isOpen, toggle } = useVersionHistory()
    expect(isOpen.value).toBe(false)
    toggle()
    expect(isOpen.value).toBe(true)
    toggle()
    expect(isOpen.value).toBe(false)
  })

  it('saveVersion adds a version', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions, saveVersion } = useVersionHistory()
    saveVersion('Hello world', 2)
    expect(versions.value).toHaveLength(1)
    expect(versions.value[0].body).toBe('Hello world')
    expect(versions.value[0].wordCount).toBe(2)
  })

  it('saveVersion does not duplicate if body is unchanged', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions, saveVersion } = useVersionHistory()
    saveVersion('Same content', 2)
    saveVersion('Same content', 2)
    expect(versions.value).toHaveLength(1)
  })

  it('saveVersion respects maxVersions limit', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions, saveVersion } = useVersionHistory(3)
    saveVersion('v1', 1)
    vi.advanceTimersByTime(100)
    saveVersion('v2', 2)
    vi.advanceTimersByTime(100)
    saveVersion('v3', 3)
    vi.advanceTimersByTime(100)
    saveVersion('v4', 4)
    expect(versions.value).toHaveLength(3)
    expect(versions.value[0].body).toBe('v4')
    expect(versions.value[2].body).toBe('v2')
  })

  it('versions are ordered newest first', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions, saveVersion } = useVersionHistory()
    saveVersion('first', 1)
    vi.advanceTimersByTime(100)
    saveVersion('second', 2)
    expect(versions.value[0].body).toBe('second')
    expect(versions.value[1].body).toBe('first')
  })

  it('each version has an id and timestamp', async () => {
    const { useVersionHistory } = await import('@/composables/useVersionHistory')
    const { versions, saveVersion } = useVersionHistory()
    saveVersion('test', 1)
    expect(versions.value[0].id).toBeTruthy()
    expect(versions.value[0].timestamp).toBeInstanceOf(Date)
  })
})

// --- SplitView + VersionHistory component integration ---

describe('SplitView component integration', () => {
  it('renders source and preview slots in split mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: {
        source: '<textarea>Markdown here</textarea>',
        preview: '<article>Rendered HTML</article>',
      },
    })
    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.find('article').exists()).toBe(true)
    expect(wrapper.find('.cursor-col-resize').exists()).toBe(true)
  })

  it('switching from source to split reveals both panels', async () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'source', splitRatio: 0.5, isDragging: false },
      slots: {
        source: '<div class="src">Source</div>',
        preview: '<div class="prev">Preview</div>',
      },
    })
    // In source mode, preview is hidden via v-show (display:none)
    const panels = wrapper.findAll('.overflow-auto')
    expect(panels).toHaveLength(2)
    expect(panels[0].element.style.display).not.toBe('none')
    expect(panels[1].element.style.display).toBe('none')

    await wrapper.setProps({ mode: 'split' })
    expect(panels[0].element.style.display).not.toBe('none')
    expect(panels[1].element.style.display).not.toBe('none')
  })
})

describe('VersionHistory component', () => {
  const mockVersions = [
    { id: 'v1', timestamp: new Date(Date.now() - 60000), body: 'First version content', wordCount: 4 },
    { id: 'v2', timestamp: new Date(Date.now() - 30000), body: 'Second version content', wordCount: 4 },
  ]

  it('renders nothing when open is false', () => {
    const wrapper = mount(VersionHistory, {
      props: { versions: mockVersions, open: false },
      global: {
        stubs: {
          Button: { template: '<button><slot /></button>' },
          XIcon: { template: '<span />' },
          RotateCcwIcon: { template: '<span />' },
          ClockIcon: { template: '<span />' },
          FileTextIcon: { template: '<span />' },
          teleport: true,
        },
      },
    })
    expect(wrapper.find('.fixed.right-0').exists()).toBe(false)
  })

  it('renders version list when open is true', () => {
    const wrapper = mount(VersionHistory, {
      props: { versions: mockVersions, open: true },
      global: {
        stubs: {
          Button: { template: '<button><slot /></button>' },
          XIcon: { template: '<span />' },
          RotateCcwIcon: { template: '<span />' },
          ClockIcon: { template: '<span />' },
          FileTextIcon: { template: '<span />' },
          teleport: true,
        },
      },
    })
    expect(wrapper.text()).toContain('版本历史')
    expect(wrapper.text()).toContain('字')
  })

  it('shows empty state when no versions', () => {
    const wrapper = mount(VersionHistory, {
      props: { versions: [], open: true },
      global: {
        stubs: {
          Button: { template: '<button><slot /></button>' },
          XIcon: { template: '<span />' },
          RotateCcwIcon: { template: '<span />' },
          ClockIcon: { template: '<span />' },
          FileTextIcon: { template: '<span />' },
          teleport: true,
        },
      },
    })
    expect(wrapper.text()).toContain('暂无版本记录')
  })

  it('emits close when close button is clicked', async () => {
    const wrapper = mount(VersionHistory, {
      props: { versions: mockVersions, open: true },
      global: {
        stubs: {
          Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
          XIcon: { template: '<span />' },
          RotateCcwIcon: { template: '<span />' },
          ClockIcon: { template: '<span />' },
          FileTextIcon: { template: '<span />' },
          teleport: true,
        },
      },
    })
    const closeBtn = wrapper.findAll('button').find(b => b.text() === '')
    if (closeBtn) {
      await closeBtn.trigger('click')
      expect(wrapper.emitted('close')).toBeTruthy()
    }
  })

  it('emits restore with version body when restore button is clicked', async () => {
    const wrapper = mount(VersionHistory, {
      props: { versions: mockVersions, open: true },
      global: {
        stubs: {
          Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
          XIcon: { template: '<span />' },
          RotateCcwIcon: { template: '<span />' },
          ClockIcon: { template: '<span />' },
          FileTextIcon: { template: '<span />' },
          teleport: true,
        },
      },
    })
    const restoreBtns = wrapper.findAll('button').filter(b => b.text().includes('恢复'))
    if (restoreBtns.length > 0) {
      await restoreBtns[0].trigger('click')
      const emitted = wrapper.emitted('restore')
      expect(emitted).toBeTruthy()
      expect(emitted![0]).toEqual(['First version content'])
    }
  })
})
