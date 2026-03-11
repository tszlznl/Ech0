import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter } from './types'
import { getProviderSetting, loadScript, loadStyle } from './loader'

declare global {
  interface Window {
    Waline?: {
      init: (options: Record<string, unknown>) => {
        update?: (options?: Record<string, unknown>) => void
        destroy?: () => void
      }
    }
  }
}

let walineInstance: { destroy?: () => void } | null = null

export function createWalineAdapter(): CommentProviderAdapter {
  return {
    async mount(el, setting) {
      const providerSetting = getProviderSetting(setting, CommentProvider.WALINE)
      const scriptURL = providerSetting.script_url || 'https://unpkg.com/@waline/client@v2/dist/waline.js'
      const cssURL = providerSetting.css_url || 'https://unpkg.com/@waline/client@v2/dist/waline.css'

      loadStyle(cssURL)
      await loadScript(scriptURL)

      const config = providerSetting.config || {}
      const serverURL = String(config.serverURL || '').trim()
      if (!window.Waline || !serverURL) return

      el.innerHTML = '<div id="waline-comment-container"></div>'
      walineInstance = window.Waline.init({
        el: '#waline-comment-container',
        serverURL,
        path: String(config.path || window.location.pathname || ''),
      })
    },
    unmount() {
      walineInstance?.destroy?.()
      walineInstance = null
    },
  }
}
