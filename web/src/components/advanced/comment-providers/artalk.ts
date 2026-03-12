import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter } from './types'
import { getProviderSetting, loadScript, loadStyle } from './loader'

declare global {
  interface Window {
    Artalk?:
      | {
          init?: (options: Record<string, unknown>) => { destroy?: () => void }
        }
      | (new (options: Record<string, unknown>) => { destroy?: () => void })
  }
}

let artalkInstance: { destroy?: () => void } | null = null

export function createArtalkAdapter(): CommentProviderAdapter {
  return {
    async mount(el, setting) {
      const providerSetting = getProviderSetting(setting, CommentProvider.ARTALK)
      const scriptURL = providerSetting.script_url || 'https://unpkg.com/artalk@2.9.1/dist/Artalk.js'
      const cssURL = providerSetting.css_url || 'https://unpkg.com/artalk@2.9.1/dist/Artalk.css'

      loadStyle(cssURL)
      await loadScript(scriptURL)

      const config = providerSetting.config || {}
      const server = String(config.server || '').trim()
      const site = String(config.site || '').trim()
      if (!window.Artalk || !server || !site) return

      el.innerHTML = '<div id="artalk-comment-container"></div>'
      const options = {
        el: '#artalk-comment-container',
        server,
        site,
        pageKey: String(config.pageKey || window.location.pathname || ''),
        pageTitle: document.title,
      }

      // Prefer the official API in recent Artalk versions.
      const artalk = window.Artalk as {
        init?: (opts: Record<string, unknown>) => { destroy?: () => void }
      }
      if (typeof artalk.init === 'function') {
        artalkInstance = artalk.init(options)
        return
      }

      // Compatibility fallback for older constructor-style builds.
      artalkInstance = new (window.Artalk as new (
        opts: Record<string, unknown>,
      ) => { destroy?: () => void })(options)
    },
    unmount() {
      artalkInstance?.destroy?.()
      artalkInstance = null
    },
  }
}
