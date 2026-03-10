type DedupeKey = 'id' | 'key' | 'url'

export type FileSelectorOptions<T extends { category?: unknown; storage_type?: unknown }> = {
  categories?: App.Api.File.Category[]
  storageTypes?: App.Api.File.StorageType[]
  predicate?: (file: T, index: number, files: T[]) => boolean
  dedupeBy?: DedupeKey
}

function normalize(value: unknown): string {
  return String(value || '').toLowerCase()
}

function passCategory(
  category: unknown,
  categories?: App.Api.File.Category[],
): boolean {
  if (!categories || categories.length === 0) return true
  const allowed = new Set(categories.map((item) => normalize(item)))
  return allowed.has(normalize(category))
}

function passStorageType(
  storageType: unknown,
  storageTypes?: App.Api.File.StorageType[],
): boolean {
  if (!storageTypes || storageTypes.length === 0) return true
  const allowed = new Set(storageTypes.map((item) => normalize(item)))
  return allowed.has(normalize(storageType))
}

function dedupeFiles<T>(files: T[], dedupeBy?: DedupeKey): T[] {
  if (!dedupeBy) return files
  const seen = new Set<string>()
  const result: T[] = []
  for (const file of files as Array<T & Record<string, unknown>>) {
    const key = String(file?.[dedupeBy] || '')
    if (!key) {
      result.push(file)
      continue
    }
    if (seen.has(key)) continue
    seen.add(key)
    result.push(file)
  }
  return result
}

export function filterFiles<T extends { category?: unknown; storage_type?: unknown }>(
  files: T[] = [],
  options: FileSelectorOptions<T> = {},
): T[] {
  const filtered = files.filter((file, index, arr) => {
    if (!passCategory(file.category, options.categories)) return false
    if (!passStorageType(file.storage_type, options.storageTypes)) return false
    if (options.predicate && !options.predicate(file, index, arr)) return false
    return true
  })
  return dedupeFiles(filtered, options.dedupeBy)
}

export const selectFiles = filterFiles
