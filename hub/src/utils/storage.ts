/** 与 web/src/utils/storage.ts 一致，避免从 web 拉取无关依赖 */
export const localStg = {
  setItem<T>(key: string, obj: T) {
    localStorage.setItem(key, JSON.stringify(obj))
  },
  getItem<T>(key: string): T | null {
    const item = localStorage.getItem(key)
    if (!item) return null
    try {
      return JSON.parse(item) as T
    } catch {
      return null
    }
  },
  removeItem(key: string) {
    localStorage.removeItem(key)
  },
  clear() {
    localStorage.clear()
  },
}
