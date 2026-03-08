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
    const imageSource = storageType === 'object' ? 's3' : storageType === 'external' ? 'url' : 'local'
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      image_source: imageSource,
      key: String(file?.key || ''),
      width: file?.width,
      height: file?.height,
    }
  })
}

