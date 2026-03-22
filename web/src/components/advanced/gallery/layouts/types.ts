export type GalleryOpenHandler = (
  startIndex: number,
  sourceElement?: HTMLElement | null,
) => void

export type GalleryImageHelperProps = {
  images: App.Api.Ech0.FileObject[]
  resolvedSrcs: string[]
  getAlt: (idx: number) => string
  isLoaded: (image: App.Api.Ech0.FileObject, idx: number) => boolean
  markLoaded: (image: App.Api.Ech0.FileObject, idx: number) => void
  open: GalleryOpenHandler
}

export type GalleryWithImageKeyProps = GalleryImageHelperProps & {
  getImageKey: (image: App.Api.Ech0.FileObject, idx: number) => string
}

export type GalleryWithAspectRatioProps = GalleryImageHelperProps & {
  getAspectRatioStyle: (image: App.Api.Ech0.FileObject) => Record<string, string> | undefined
}

export type GalleryHorizontalProps = GalleryWithImageKeyProps & {
  scrollHintText: string
  getHorizontalAspectStyle: (
    image: App.Api.Ech0.FileObject,
  ) => Record<string, string> | undefined
}

export type GalleryWaterfallProps = GalleryWithImageKeyProps & {
  getAspectRatioStyle: (image: App.Api.Ech0.FileObject) => Record<string, string> | undefined
}
