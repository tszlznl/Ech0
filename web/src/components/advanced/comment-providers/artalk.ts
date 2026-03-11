import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter } from './types'
import { getProviderSetting, loadScript, loadStyle } from './loader'

declare global {
  interface Window {
    Artalk?: {
      new (options: Record<string, unknown>): { destroy?: () => void }
    }
  }
}

let artalkInstance: { destroy?: () => void } | null = null

export function createArtalkAdapter(): CommentProviderAdapter {
  return {
    async mount(el, setting) {
      const providerSetting = getProviderSetting(setting, CommentProvider.ARTALK)
      const scriptURL = providerSetting.script_url || 'https://unpkg.com/artalk@2/dist/Artalk.js'
      const cssURL = providerSetting.css_url || 'https://unpkg.com/artalk@2/dist/Artalk.css'

      loadStyle(cssURL)
      await loadScript(scriptURL)

      const config = providerSetting.config || {}
      const server = String(config.server || '').trim()
      const site = String(config.site || '').trim()
      if (!window.Artalk || !server || !site) return

      el.innerHTML = '<div id="artalk-comment-container"></div>'
      artalkInstance = new window.Artalk({
        el: '#artalk-comment-container',
        server,
        site,
        pageKey: String(config.pageKey || window.location.pathname || ''),
        pageTitle: document.title,
      })
    },
    unmount() {
      artalkInstance?.destroy?.()
      artalkInstance = null
    },
  }
}
