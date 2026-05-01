import { computed, type Ref, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { theToast } from '@/utils/toast'
import { getImageSize } from '@/utils/image'
import { compressImage, inferFileExtFromType } from './compress'
import { getPresign, updateFileMeta } from './api/adapter'
import { globalFileRegistry } from './registry/file-registry'
import { httpUpload } from './upload'

export type UploadStatus =
  | 'pending'
  | 'compressing'
  | 'uploading'
  | 'success'
  | 'error'
  | 'cancelled'

export interface QueueItem {
  id: string
  file: File
  originalFile: File
  status: UploadStatus
  progress: number
  error?: string
  preview: string
  result?: App.Api.Ech0.FileToAdd
  abort?: AbortController
  delivered?: boolean
}

export interface UseUploadOptions {
  storageType: Ref<App.Api.File.StorageType>
  enableCompressor: Ref<boolean>
  category?: App.Api.File.Category
  allowedTypes?: string[]
  maxFiles?: number
  concurrency?: number
  onAllComplete: (files: App.Api.Ech0.FileToAdd[]) => void
}

function extractUploadPayload(response: unknown): Record<string, unknown> {
  const resp = (response || {}) as Record<string, unknown>
  const data = resp.data
  if (data && typeof data === 'object') return data as Record<string, unknown>
  return resp
}

function mimeMatches(rule: string, fileType: string): boolean {
  if (rule === fileType) return true
  if (rule.endsWith('/*')) {
    return fileType.startsWith(rule.slice(0, -1))
  }
  return false
}

export function useUpload(opts: UseUploadOptions) {
  const items = ref<QueueItem[]>([])

  const userStore = useUserStore()
  const authStore = useAuthStore()
  const { t } = useI18n()

  const envURL = (import.meta.env.VITE_SERVICE_BASE_URL as string) || ''
  const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL

  const category = opts.category ?? FILE_CATEGORY.IMAGE
  const allowedTypes = opts.allowedTypes?.length ? opts.allowedTypes : ['image/*']
  const maxFiles = opts.maxFiles ?? 6
  const concurrency = opts.concurrency ?? 3

  let inFlight = 0

  const isUploading = computed(() =>
    items.value.some(
      (i) => i.status === 'pending' || i.status === 'compressing' || i.status === 'uploading',
    ),
  )

  const successCount = computed(() => items.value.filter((i) => i.status === 'success').length)
  const totalActive = computed(
    () =>
      items.value.filter(
        (i) => i.status === 'pending' || i.status === 'compressing' || i.status === 'uploading',
      ).length,
  )

  function genId(file: File): string {
    return `${file.name}|${file.size}|${file.lastModified}`
  }

  function isAllowed(file: File): boolean {
    return allowedTypes.some((rule) => mimeMatches(rule, file.type))
  }

  function liveCount(): number {
    return items.value.filter((i) => i.status !== 'cancelled' && i.status !== 'error').length
  }

  function addFiles(files: File[]) {
    if (!userStore.isLogin) {
      theToast.error(String(t('uploader.loginRequired')))
      return
    }

    const existingIds = new Set(items.value.map((i) => i.id))
    let remaining = Math.max(0, maxFiles - liveCount())
    let rejected = false

    for (const file of files) {
      if (remaining <= 0) {
        rejected = true
        break
      }
      if (!isAllowed(file)) continue
      const id = genId(file)
      if (existingIds.has(id)) continue
      existingIds.add(id)

      items.value.push({
        id,
        file,
        originalFile: file,
        status: 'pending',
        progress: 0,
        preview: URL.createObjectURL(file),
      })
      remaining--
    }

    if (rejected) {
      theToast.info(String(t('uploader.maxFilesReached', { max: maxFiles })))
    }

    pump()
  }

  function pump() {
    while (inFlight < concurrency) {
      const next = items.value.find((i) => i.status === 'pending')
      if (!next) break
      void processItem(next)
    }

    const stillWorking = items.value.some(
      (i) => i.status === 'pending' || i.status === 'compressing' || i.status === 'uploading',
    )
    if (!stillWorking && inFlight === 0) {
      const undelivered = items.value.filter(
        (i) => i.status === 'success' && i.result && !i.delivered,
      )
      if (undelivered.length > 0) {
        for (const item of undelivered) item.delivered = true
        opts.onAllComplete(undelivered.map((i) => i.result!))
      }
    }
  }

  async function processItem(item: QueueItem) {
    inFlight++
    const ctl = new AbortController()
    item.abort = ctl

    try {
      let working = item.file
      if (opts.enableCompressor.value && working.type.startsWith('image/')) {
        item.status = 'compressing'
        try {
          working = await compressImage(working)
          item.file = working
        } catch (err) {
          // Compression failure must not block upload — fall back to the original file.
          console.warn('[uploader] compression failed, using original file:', err)
        }
      }

      if (ctl.signal.aborted) {
        item.status = 'cancelled'
        return
      }

      item.status = 'uploading'
      item.progress = 0

      const result =
        opts.storageType.value === FILE_STORAGE_TYPE.OBJECT
          ? await uploadToS3(working, item, ctl.signal)
          : await uploadToLocal(working, item, ctl.signal)

      item.result = result
      item.status = 'success'
      item.progress = 100

      globalFileRegistry.upsert({
        id: result.id || '',
        key: result.key,
        url: result.url,
        category: result.category,
        contentType: result.content_type,
        storageType: result.storage_type,
        size: result.size,
        width: result.width,
        height: result.height,
      })
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        item.status = 'cancelled'
      } else {
        item.status = 'error'
        const msg =
          (err instanceof Error ? err.message : String(err)) || String(t('uploader.uploadError'))
        item.error = msg
        theToast.error(msg)
      }
    } finally {
      item.abort = undefined
      inFlight--
      pump()
    }
  }

  async function uploadToLocal(
    file: File,
    item: QueueItem,
    signal: AbortSignal,
  ): Promise<App.Api.Ech0.FileToAdd> {
    const res = await httpUpload(
      file,
      {
        kind: 'local',
        endpoint: `${backendURL}/api/files/upload`,
        authHeader: authStore.authHeader,
        fields: {
          category,
          storage_type: FILE_STORAGE_TYPE.LOCAL,
        },
      },
      {
        signal,
        onProgress: (loaded, total) => {
          if (total > 0) item.progress = Math.round((loaded / total) * 100)
        },
      },
    )

    const payload = extractUploadPayload(res.responseBody)
    const fileId = String(payload.id || payload.file_id || payload.ID || '')
    const fileKey = String(payload.key || payload.object_key || '')
    const fileUrl = String(payload.url || payload.file_url || payload.access_url || '')
    if (!fileId || !fileUrl) {
      throw new Error(String(t('uploader.missingFileIdentifier')))
    }
    return {
      id: fileId,
      url: fileUrl,
      storage_type: FILE_STORAGE_TYPE.LOCAL,
      key: fileKey,
      content_type: typeof payload.content_type === 'string' ? payload.content_type : file.type,
      size: typeof payload.size === 'number' ? payload.size : file.size,
      width: typeof payload.width === 'number' ? payload.width : undefined,
      height: typeof payload.height === 'number' ? payload.height : undefined,
      category,
    }
  }

  async function uploadToS3(
    file: File,
    item: QueueItem,
    signal: AbortSignal,
  ): Promise<App.Api.Ech0.FileToAdd> {
    const contentType = file.type || 'application/octet-stream'
    const rawName = String(file.name || '').trim()
    const fileName = rawName || `upload_${Date.now()}${inferFileExtFromType(contentType)}`

    const presign = await getPresign({
      fileName,
      contentType,
      storageType: FILE_STORAGE_TYPE.OBJECT,
    })

    if (signal.aborted) throw new DOMException('Aborted', 'AbortError')

    await httpUpload(
      file,
      {
        kind: 's3',
        presignUrl: presign.presign_url,
        contentType,
      },
      {
        signal,
        onProgress: (loaded, total) => {
          if (total > 0) item.progress = Math.round((loaded / total) * 100)
        },
      },
    )

    let width: number | undefined
    let height: number | undefined
    if (category === FILE_CATEGORY.IMAGE) {
      try {
        const dim = await getImageSize(item.preview)
        width = dim.width
        height = dim.height
      } catch {
        // Best-effort; backend can fall back without dims.
      }
    }

    let resolvedSize = file.size
    let resolvedWidth = width
    let resolvedHeight = height
    let resolvedContentType: string | undefined = contentType
    try {
      const updated = await updateFileMeta({
        id: presign.id,
        size: file.size,
        width,
        height,
        contentType,
      })
      resolvedSize = updated.size ?? resolvedSize
      resolvedWidth = updated.width
      resolvedHeight = updated.height
      resolvedContentType = updated.contentType ?? resolvedContentType
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(t('uploader.fileMetaFillbackFailed'))
      theToast.error(msg)
    }

    return {
      id: presign.id,
      url: presign.file_url,
      storage_type: FILE_STORAGE_TYPE.OBJECT,
      key: presign.key,
      content_type: resolvedContentType,
      size: resolvedSize,
      width: resolvedWidth,
      height: resolvedHeight,
      category,
    }
  }

  function retry(id: string) {
    const item = items.value.find((i) => i.id === id)
    if (!item) return
    if (item.status !== 'error' && item.status !== 'cancelled') return
    item.status = 'pending'
    item.error = undefined
    item.progress = 0
    item.file = item.originalFile
    item.delivered = false
    pump()
  }

  function remove(id: string) {
    const idx = items.value.findIndex((i) => i.id === id)
    if (idx < 0) return
    const item = items.value[idx]
    if (item.abort) item.abort.abort()
    URL.revokeObjectURL(item.preview)
    items.value.splice(idx, 1)
    pump()
  }

  function cancelAll() {
    for (const item of items.value) {
      if (item.abort) item.abort.abort()
      URL.revokeObjectURL(item.preview)
    }
    items.value = []
  }

  function reset() {
    cancelAll()
  }

  return {
    items,
    isUploading,
    successCount,
    totalActive,
    addFiles,
    retry,
    remove,
    cancelAll,
    reset,
  }
}
