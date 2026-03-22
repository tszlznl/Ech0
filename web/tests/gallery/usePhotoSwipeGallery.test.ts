import { defineComponent, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePhotoSwipeGallery } from '@/components/advanced/gallery/composables/usePhotoSwipeGallery'

type MockPhotoSwipeInstance = {
  options: Record<string, unknown>
  init: ReturnType<typeof vi.fn>
  destroy: ReturnType<typeof vi.fn>
  on: ReturnType<typeof vi.fn>
  addFilter: ReturnType<typeof vi.fn>
}

const photoSwipeInstances = vi.hoisted(() => [] as MockPhotoSwipeInstance[])

vi.mock('photoswipe', () => ({
  default: class MockPhotoSwipe {
    options: Record<string, unknown>
    init = vi.fn()
    destroy = vi.fn()
    on = vi.fn()
    addFilter = vi.fn()

    constructor(options: Record<string, unknown>) {
      this.options = options
      photoSwipeInstances.push(this as MockPhotoSwipeInstance)
    }
  },
}))

const createHost = (items: Array<{ src: string; width?: number; height?: number; alt?: string }>) =>
  mount(
    defineComponent({
      setup() {
        const { open } = usePhotoSwipeGallery(ref(items))
        return { open }
      },
      template: '<div />',
    }),
  )

describe('usePhotoSwipeGallery', () => {
  beforeEach(() => {
    photoSwipeInstances.length = 0
  })

  it('normalizes index and fallback dimensions before opening', () => {
    const wrapper = createHost([
      { src: '/a.jpg', width: 0, height: undefined, alt: 'a' },
      { src: '/b.jpg', width: 800, height: 600, alt: 'b' },
    ])
    ;(wrapper.vm as { open: (index: number) => void }).open(99)

    const options = photoSwipeInstances[0]?.options as {
      index: number
      dataSource: Array<{ src: string; width: number; height: number }>
    }

    expect(options.index).toBe(1)
    expect(options.dataSource[0]).toMatchObject({
      src: '/a.jpg',
      width: 1600,
      height: 1600,
    })
    expect(photoSwipeInstances[0]?.init).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })

  it('injects thumbnail msrc and source element for the current slide', () => {
    const wrapper = createHost([
      { src: '/first.jpg', width: 500, height: 300 },
      { src: '/second.jpg', width: 700, height: 400 },
    ])

    const sourceElement = document.createElement('button')
    const img = document.createElement('img')
    img.src = '/thumb.jpg'
    sourceElement.appendChild(img)

    ;(wrapper.vm as { open: (index: number, sourceElement?: HTMLElement | null) => void }).open(
      0,
      sourceElement,
    )

    const options = photoSwipeInstances[0]?.options as {
      dataSource: Array<Record<string, unknown>>
    }
    const currentSlide = options.dataSource[0]

    expect(currentSlide.element).toBe(sourceElement)
    expect(String(currentSlide.msrc)).toContain('/thumb.jpg')

    const addFilter = photoSwipeInstances[0]?.addFilter
    expect(addFilter).toHaveBeenCalledWith('placeholderSrc', expect.any(Function))
    const placeholderFilter = addFilter?.mock.calls[0]?.[1] as (
      placeholderSrc: string,
      content: { slide?: { index?: number } },
    ) => string

    expect(placeholderFilter('fallback', { slide: { index: 0 } })).toContain('/thumb.jpg')
    expect(placeholderFilter('fallback', { slide: { index: 1 } })).toBe('fallback')

    wrapper.unmount()
  })
})
