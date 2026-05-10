import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SplitView from '@/components/SplitView.vue'

describe('SplitView', () => {
  it('renders source panel in source mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'source', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Source')
  })

  it('renders both panels in split mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Source')
    expect(wrapper.text()).toContain('Preview')
  })

  it('renders preview panel in preview mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'preview', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.text()).toContain('Preview')
  })

  it('emits dragStart on divider mousedown', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    const divider = wrapper.find('.cursor-col-resize')
    if (divider.exists()) {
      divider.trigger('mousedown')
      expect(wrapper.emitted('dragStart')).toBeTruthy()
    }
  })

  it('applies select-none class when dragging', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: true },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.classes()).toContain('select-none')
  })

  it('does not apply select-none class when not dragging', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(wrapper.classes()).not.toContain('select-none')
  })

  it('applies split ratio as width in split mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.6, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    const panels = wrapper.findAll('.overflow-auto')
    expect(panels.length).toBeGreaterThanOrEqual(2)
    expect(panels[0].attributes('style')).toContain('width: 60%')
    expect(panels[1].attributes('style')).toContain('width: 40%')
  })

  it('divider is only visible in split mode', () => {
    const sourceWrapper = mount(SplitView, {
      props: { mode: 'source', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(sourceWrapper.find('.cursor-col-resize').exists()).toBe(false)

    const splitWrapper = mount(SplitView, {
      props: { mode: 'split', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>Source</div>', preview: '<div>Preview</div>' },
    })
    expect(splitWrapper.find('.cursor-col-resize').exists()).toBe(true)
  })

  it('source panel has display:none in preview mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'preview', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>SourceContent</div>', preview: '<div>PreviewContent</div>' },
    })
    const panels = wrapper.findAll('.overflow-auto')
    // v-show keeps both in DOM but hides source with display:none
    expect(panels).toHaveLength(2)
    expect(panels[0].element.style.display).toBe('none')
    expect(panels[1].element.style.display).not.toBe('none')
  })

  it('preview panel has display:none in source mode', () => {
    const wrapper = mount(SplitView, {
      props: { mode: 'source', splitRatio: 0.5, isDragging: false },
      slots: { source: '<div>SourceContent</div>', preview: '<div>PreviewContent</div>' },
    })
    const panels = wrapper.findAll('.overflow-auto')
    // v-show keeps both in DOM but hides preview with display:none
    expect(panels).toHaveLength(2)
    expect(panels[0].element.style.display).not.toBe('none')
    expect(panels[1].element.style.display).toBe('none')
  })
})
