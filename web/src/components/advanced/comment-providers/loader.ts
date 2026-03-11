const scriptCache = new Map<string, Promise<void>>()
const styleCache = new Set<string>()

export function loadScript(src: string): Promise<void> {
  const normalized = (src || '').trim()
  if (!normalized) return Promise.resolve()
  if (scriptCache.has(normalized)) {
    return scriptCache.get(normalized)!
  }

  const existing = document.querySelector(`script[src="${normalized}"]`)
  if (existing) {
    const ready = Promise.resolve()
    scriptCache.set(normalized, ready)
    return ready
  }

  const task = new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.src = normalized
    script.async = true
    script.onload = () => resolve()
    script.onerror = (err) => reject(err)
    document.head.appendChild(script)
  })

  scriptCache.set(normalized, task)
  return task
}

export function loadStyle(href: string): void {
  const normalized = (href || '').trim()
  if (!normalized || styleCache.has(normalized)) return

  const exists = document.querySelector(`link[rel="stylesheet"][href="${normalized}"]`)
  if (exists) {
    styleCache.add(normalized)
    return
  }

  const link = document.createElement('link')
  link.rel = 'stylesheet'
  link.href = normalized
  document.head.appendChild(link)
  styleCache.add(normalized)
}

export function getProviderSetting(
  setting: App.Api.Setting.CommentSetting,
  provider: string,
): App.Api.Setting.CommentProviderSetting {
  return setting.providers?.[provider] || { config: {} }
}
