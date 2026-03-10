import { FILE_STORAGE_TYPE } from '@/constants/file'
import { filterFiles, type FileSelectorOptions } from '@/lib/file'

type EchoLike = {
  id?: string
  echo_files?:
    | Array<{
        echo_id?: string
        file_id?: string
        file?: {
          id?: string
          key?: string
          url?: string
          storage_type?: string
          category?: string
          content_type?: string
          width?: number
          height?: number
        }
      }>
    | null
}

export function getEchoFiles(echo?: EchoLike | null): App.Api.Ech0.FileObject[] {
  if (!echo) return []

  return (echo.echo_files || []).map((item) => {
    const file = item.file
    const storageType = normalizeStorageType(file?.storage_type)
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      storage_type: storageType,
      category: (file?.category as App.Api.File.Category | undefined) || undefined,
      content_type: file?.content_type,
      key: String(file?.key || ''),
      width: file?.width,
      height: file?.height,
    }
  })
}

export function getEchoFilesBy(
  echo?: EchoLike | null,
  options: FileSelectorOptions<App.Api.Ech0.FileObject> = {},
): App.Api.Ech0.FileObject[] {
  return filterFiles(getEchoFiles(echo), options)
}

// backward-compatible alias
export const getEchoImages = (echo?: EchoLike | null) =>
  getEchoFilesBy(echo, { categories: ['image'] })

function normalizeStorageType(raw: unknown): App.Api.File.StorageType {
  const value = String(raw || '').toLowerCase()
  if (value === FILE_STORAGE_TYPE.OBJECT) return FILE_STORAGE_TYPE.OBJECT
  if (value === FILE_STORAGE_TYPE.EXTERNAL) return FILE_STORAGE_TYPE.EXTERNAL
  return FILE_STORAGE_TYPE.LOCAL
}

