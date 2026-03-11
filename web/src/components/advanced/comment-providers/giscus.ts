import { CommentProvider } from '@/enums/enums'
import type { CommentProviderAdapter } from './types'
import { getProviderSetting, resolveResourceURL } from './loader'

let mountedScript: HTMLScriptElement | null = null

function toAttr(config: Record<string, unknown>, key: string, fallback = ''): string {
  const value = config[key]
  if (value === undefined || value === null) return fallback
  return String(value)
}

export function createGiscusAdapter(): CommentProviderAdapter {
  return {
    mount(el, setting) {
      const providerSetting = getProviderSetting(setting, CommentProvider.GISCUS)
      const config = providerSetting.config || {}
      const required = ['repo', 'repoId', 'category', 'categoryId']
      const missing = required.some((key) => !toAttr(config, key))
      if (missing) return

      el.innerHTML = ''
      const script = document.createElement('script')
      script.src = resolveResourceURL(providerSetting.script_url || 'https://giscus.app/client.js')
      script.async = true
      script.crossOrigin = 'anonymous'
      script.setAttribute('data-repo', toAttr(config, 'repo'))
      script.setAttribute('data-repo-id', toAttr(config, 'repoId'))
      script.setAttribute('data-category', toAttr(config, 'category'))
      script.setAttribute('data-category-id', toAttr(config, 'categoryId'))
      script.setAttribute('data-mapping', toAttr(config, 'mapping', 'pathname'))
      script.setAttribute('data-strict', toAttr(config, 'strict', '0'))
      script.setAttribute('data-reactions-enabled', toAttr(config, 'reactionsEnabled', '1'))
      script.setAttribute('data-input-position', toAttr(config, 'inputPosition', 'top'))
      script.setAttribute('data-lang', toAttr(config, 'lang', 'zh-CN'))
      script.setAttribute('data-theme', toAttr(config, 'theme', 'preferred_color_scheme'))
      el.appendChild(script)
      mountedScript = script
    },
    unmount() {
      if (mountedScript && mountedScript.parentNode) {
        mountedScript.parentNode.removeChild(mountedScript)
      }
      mountedScript = null
    },
  }
}
