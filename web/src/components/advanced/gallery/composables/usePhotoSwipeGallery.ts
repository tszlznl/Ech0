import { onBeforeUnmount, shallowRef, type Ref } from 'vue'
import PhotoSwipe from 'photoswipe'
import 'photoswipe/style.css'

type PhotoSwipeItem = {
  src: string
  width?: number
  height?: number
  alt?: string
}

const FALLBACK_DIMENSION = 1600

const normalizeDimension = (value?: number) => {
  if (!value || Number.isNaN(value) || value <= 0) return FALLBACK_DIMENSION
  return value
}

const resolveThumbImageSrc = (sourceElement?: HTMLElement | null) => {
  if (!sourceElement) return undefined

  const thumbnail =
    sourceElement instanceof HTMLImageElement
      ? sourceElement
      : (sourceElement.querySelector('img') as HTMLImageElement | null)

  if (!thumbnail) return undefined

  return thumbnail.currentSrc || thumbnail.src || undefined
}

export const usePhotoSwipeGallery = (items: Ref<PhotoSwipeItem[]>) => {
  const instance = shallowRef<PhotoSwipe | null>(null)

  const destroy = () => {
    instance.value?.destroy()
    instance.value = null
  }

  const open = (startIndex: number, sourceElement?: HTMLElement | null) => {
    if (!items.value.length) return

    destroy()

    const normalizedIndex = Math.min(Math.max(startIndex, 0), items.value.length - 1)
    const thumbImageSrc = resolveThumbImageSrc(sourceElement)
    const dataSource = items.value.map((item, index) => {
      const baseItem = {
        src: item.src,
        width: normalizeDimension(item.width),
        height: normalizeDimension(item.height),
        alt: item.alt,
      }

      if (index === normalizedIndex && sourceElement) {
        return {
          ...baseItem,
          element: sourceElement,
          msrc: thumbImageSrc || item.src,
        }
      }

      if (index === normalizedIndex && thumbImageSrc) {
        return {
          ...baseItem,
          msrc: thumbImageSrc,
        }
      }

      return baseItem
    })

    const pswp = new PhotoSwipe({
      dataSource,
      index: normalizedIndex,
      bgOpacity: 0.85,
      showHideAnimationType: 'zoom',
      preloadFirstSlide: true,
      escKey: true,
      arrowKeys: true,
      bgClickAction: 'close',
      closeOnVerticalDrag: true,
      clickToCloseNonZoomable: true,
    })

    pswp.on('destroy', () => {
      if (instance.value === pswp) {
        instance.value = null
      }
    })

    if (thumbImageSrc) {
      pswp.addFilter('placeholderSrc', (placeholderSrc, content) => {
        if (content.slide?.index === normalizedIndex) {
          return thumbImageSrc
        }
        return placeholderSrc
      })
    }

    instance.value = pswp
    pswp.init()
  }

  onBeforeUnmount(() => {
    destroy()
  })

  return {
    open,
    destroy,
  }
}
