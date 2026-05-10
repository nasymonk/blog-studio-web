import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useToast } from '@/composables/useToast'

describe('useToast', () => {
  beforeEach(() => {
    const { dismissAll } = useToast()
    dismissAll()
    vi.clearAllMocks()
  })

  it('adds success toast', () => {
    const { toasts, success } = useToast()
    success('Saved!', 'Your changes were saved')
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].type).toBe('success')
    expect(toasts.value[0].title).toBe('Saved!')
    expect(toasts.value[0].description).toBe('Your changes were saved')
  })

  it('adds error toast with infinite duration', () => {
    const { toasts, error } = useToast()
    error('Failed', 'Something went wrong')
    expect(toasts.value[0].type).toBe('error')
    expect(toasts.value[0].duration).toBe(Infinity)
  })

  it('adds warning toast', () => {
    const { toasts, warning } = useToast()
    warning('Warning')
    expect(toasts.value[0].type).toBe('warning')
  })

  it('adds info toast', () => {
    const { toasts, info } = useToast()
    info('Info')
    expect(toasts.value[0].type).toBe('info')
  })

  it('removes toast by id', () => {
    const { toasts, success, remove } = useToast()
    success('Test')
    const id = toasts.value[0].id
    remove(id)
    expect(toasts.value).toHaveLength(0)
  })

  it('limits to 3 toasts', () => {
    const { toasts, info } = useToast()
    info('1'); info('2'); info('3'); info('4')
    expect(toasts.value).toHaveLength(3)
    expect(toasts.value[0].title).toBe('2')
  })

  it('assigns unique ids to each toast', () => {
    const { toasts, success } = useToast()
    success('First')
    success('Second')
    expect(toasts.value[0].id).not.toBe(toasts.value[1].id)
  })

  it('dismissAll clears all toasts', () => {
    const { toasts, success, dismissAll } = useToast()
    success('One')
    success('Two')
    expect(toasts.value).toHaveLength(2)
    dismissAll()
    expect(toasts.value).toHaveLength(0)
  })

  it('success toast has 5000ms default duration', () => {
    const { toasts, success } = useToast()
    success('Done')
    expect(toasts.value[0].duration).toBe(5000)
  })

  it('warning toast has 6000ms default duration', () => {
    const { toasts, warning } = useToast()
    warning('Caution')
    expect(toasts.value[0].duration).toBe(6000)
  })

  it('info toast has 5000ms default duration', () => {
    const { toasts, info } = useToast()
    info('Note')
    expect(toasts.value[0].duration).toBe(5000)
  })

  it('error toast without description works', () => {
    const { toasts, error } = useToast()
    error('Oops')
    expect(toasts.value[0].title).toBe('Oops')
    expect(toasts.value[0].description).toBeUndefined()
  })

  it('add() with custom duration overrides default', () => {
    const { toasts, add } = useToast()
    add({ type: 'success', title: 'Custom', duration: 10000 })
    expect(toasts.value[0].duration).toBe(10000)
  })

  it('add() with error type defaults to Infinity duration', () => {
    const { toasts, add } = useToast()
    add({ type: 'error', title: 'Oops' })
    expect(toasts.value[0].duration).toBe(Infinity)
  })
})
