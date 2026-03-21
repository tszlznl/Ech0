<template>
  <div v-if="images.length" class="image-gallery-container">
    <GalleryWaterfall
      v-if="layoutValue === ImageLayout.WATERFALL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-aspect-ratio-style="getAspectRatioStyle"
    />

    <GalleryGrid
      v-if="layoutValue === ImageLayout.GRID"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
    />

    <GalleryCarousel
      v-if="layoutValue === ImageLayout.CAROUSEL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-aspect-ratio-style="getAspectRatioStyle"
    />

    <GalleryHorizontal
      v-if="layoutValue === ImageLayout.HORIZONTAL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :scroll-hint-text="t('imageGallery.scrollHint')"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-horizontal-aspect-style="getHorizontalAspectStyle"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { getImageUrl, getHubImageUrl } from '@/utils/other'
import { ImageLayout } from '@/enums/enums'
import { useI18n } from 'vue-i18n'
import { usePhotoSwipeGallery } from './composables/usePhotoSwipeGallery'
import GalleryWaterfall from './layouts/GalleryWaterfall.vue'
import GalleryGrid from './layouts/GalleryGrid.vue'
import GalleryCarousel from './layouts/GalleryCarousel.vue'
import GalleryHorizontal from './layouts/GalleryHorizontal.vue'

const props = defineProps<{
  images?: App.Api.Ech0.FileObject[]
  baseUrl?: string
  layout?: ImageLayout | string | undefined
}>()

const { t } = useI18n()

const images = computed(() => props.images || [])
const layoutValue = computed(() => props.layout || ImageLayout.WATERFALL)
const loadedImages = ref<Record<string, boolean>>({})
const resolvedSrcs = computed(() =>
  images.value.map((image) =>
    props.baseUrl ? getHubImageUrl(image, props.baseUrl) : getImageUrl(image),
  ),
)

const galleryItems = computed(() =>
  images.value.map((image, idx) => ({
    src: resolvedSrcs.value[idx] || '',
    width: image.width,
    height: image.height,
    alt: getAlt(idx),
  })),
)

const { open } = usePhotoSwipeGallery(galleryItems)

const getAlt = (idx: number) => t('imageGallery.previewImage', { index: idx + 1 })

const openGallery = (index: number, sourceElement?: HTMLElement | null) => {
  open(index, sourceElement)
}

const getImageKey = (image: App.Api.Ech0.FileObject, idx: number) =>
  image.id || `${image.url}-${idx}`

const getAspectRatioStyle = (image: App.Api.Ech0.FileObject): Record<string, string> | undefined => {
  if (!image.width || !image.height) return undefined
  return { aspectRatio: `${image.width} / ${image.height}` }
}

const getHorizontalAspectStyle = (
  image: App.Api.Ech0.FileObject,
): Record<string, string> | undefined => {
  if (!image.width || !image.height) return undefined
  return { aspectRatio: `${image.width} / ${image.height}`, width: 'auto' }
}

const isImageLoaded = (image: App.Api.Ech0.FileObject, idx: number) =>
  Boolean(loadedImages.value[getImageKey(image, idx)])

const markImageLoaded = (image: App.Api.Ech0.FileObject, idx: number) => {
  loadedImages.value[getImageKey(image, idx)] = true
}
</script>

<style scoped>
.image-gallery-container {
  position: relative;
}

.imgwidth {
  width: 88%;
}
</style>
