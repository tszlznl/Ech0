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

export function getEchoImages(echo?: EchoLike | null): App.Api.Ech0.FileObject[] {
  if (!echo) return []

  return (echo.echo_files || []).map((item) => {
    const file = item.file
    const storageType = String(file?.storage_type || 'local')
    const normalizedStorageType =
      storageType === 'object' || storageType === 'external' ? storageType : 'local'
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      storage_type: normalizedStorageType as App.Api.File.StorageType,
      key: String(file?.key || ''),
      width: file?.width,
      height: file?.height,
    }
  })
}

