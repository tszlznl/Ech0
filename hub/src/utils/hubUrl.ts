/** 与实例清单、缓冲池 Map 键一致，避免末尾斜杠导致 logo 对不上 */
export function normalizeHubInstanceUrl(url: string): string {
  return url.trim().replace(/\/+$/, '')
}
