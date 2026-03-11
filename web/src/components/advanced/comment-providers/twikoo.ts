import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter } from './types'
import { getProviderSetting, loadScript } from './loader'

declare global {
  interface Window {
    twikoo?: {
      init: (options: Record<string, unknown>) => void
    }
  }
}

let mounted = false

export function createTwikooAdapter(): CommentProviderAdapter {
  return {
    async mount(el, setting) {
      const providerSetting = getProviderSetting(setting, CommentProvider.TWIKOO)
      const scriptURL = providerSetting.script_url || '/others/scripts/twikoo.all.min.js'
      try {
        await loadScript(scriptURL)
      } catch {
        // Fallback to CDN when local script is unavailable.
        await loadScript('https://cdn.staticfile.net/twikoo/1.6.44/twikoo.all.min.js')
      }

      const config = providerSetting.config || {}
      const envId = String(config.envId || '').trim()
      if (!window.twikoo || !envId) return

      el.innerHTML = '<div id="twikoo-comment-container"></div>'
      window.twikoo.init({
        envId,
        el: '#twikoo-comment-container',
      })
      mounted = true
    },
    unmount() {
      if (!mounted) return
      mounted = false
    },
  }
}
