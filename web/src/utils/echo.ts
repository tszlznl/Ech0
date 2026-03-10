import { FILE_STORAGE_TYPE } from '@/constants/file'

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
      key: String(file?.key || ''),
      width: file?.width,
      height: file?.height,
    }
  })
}

// backward-compatible alias
export const getEchoImages = getEchoFiles

function normalizeStorageType(raw: unknown): App.Api.File.StorageType {
  const value = String(raw || '').toLowerCase()
  if (value === FILE_STORAGE_TYPE.OBJECT) return FILE_STORAGE_TYPE.OBJECT
  if (value === FILE_STORAGE_TYPE.EXTERNAL) return FILE_STORAGE_TYPE.EXTERNAL
  return FILE_STORAGE_TYPE.LOCAL
}

