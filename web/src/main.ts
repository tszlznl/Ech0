import 'virtual:uno.css'
import '@/themes/index.scss'
import 'floating-vue/dist/style.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { initStores } from './stores/store-init'
import { useEchoStore, useSettingStore } from '@/stores'
import { setupI18n, setI18nLocale, LOCALE_STORAGE_KEY } from './locales'
import { localStg } from '@/utils/storage'

// 自定义组件
import BaseDialog from '@/components/common/BaseDialog.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// Pre-warm the homepage timeline in parallel with bootstrap so the first page
// of echoes is ready by the time <TheEchos> mounts. Uses current URL because
// the router hasn't resolved yet.
const pathname = typeof window !== 'undefined' ? (window.location.pathname ?? '') : ''
if (pathname === '/' || pathname === '') {
  useEchoStore()
    .getEchosByPage()
    .catch(() => undefined)
}

// Whether the user already has a persisted locale. Checked before `setupI18n`
// because `setupI18n` writes localStorage itself; without this snapshot the
// server-side default_locale override (for first-time visitors) would be
// dropped.
const hadPersistedLocale = Boolean(localStg.getItem(LOCALE_STORAGE_KEY))

// Parallel bootstrap: stores init, i18n messages, floating-vue module.
const [, i18n, floatingVueModule] = await Promise.all([
  initStores().catch((e) => {
    console.error('Failed to initialize stores:', e)
  }),
  setupI18n(),
  import('floating-vue'),
])

if (!hadPersistedLocale) {
  const serverLocale = useSettingStore().SystemSetting.default_locale
  if (serverLocale && serverLocale !== i18n.global.locale.value) {
    await setI18nLocale(serverLocale).catch(() => undefined)
  }
}

app.use(router)
app.use(i18n)
app.use(floatingVueModule.default, {
  themes: {
    tooltip: {
      triggers: ['hover'],
      // `touch` → touchend on target; iOS Safari often omits mouseleave after tap.
      hideTriggers: ['hover', 'click', 'touch'],
      placement: 'top',
      delay: { show: 300, hide: 80 },
      distance: 10,
      container: 'body',
      // Do not move focus into the popper on show (reduces aria issues with tooltips).
      noAutoFocus: true,
      // Tap outside to dismiss (needed on iOS when autoHide was false + sticky hover).
      autoHide: true,
    },
  },
})

// 全局注册组件
app.component('BaseDialog', BaseDialog)

app.mount('#app')

// 启动加载页淡出并恢复页面滚动。
// 等待 router.isReady() 以确保首个 route 已完成导航守卫与组件解析，
// 避免 loader 在白屏阶段就开始淡出。
const appLoader = document.getElementById('app-loader')
let loaderCleared = false
const clearStartupLoader = () => {
  if (loaderCleared) return
  loaderCleared = true
  appLoader?.remove()
  document.documentElement.classList.remove('app-loading')
}

const startLoaderFade = () => {
  if (!appLoader) {
    clearStartupLoader()
    return
  }
  // 让首帧内容先绘制一次，再触发 loader 的 opacity 过渡。
  window.requestAnimationFrame(() => {
    appLoader.classList.add('fade-out')
  })
  appLoader.addEventListener('transitionend', clearStartupLoader, { once: true })
  // transitionend 有时在后台标签页或极慢渲染下不触发；兜底清理。
  window.setTimeout(clearStartupLoader, 800)
}

// 以 3 秒为安全上限，避免守卫永远 pending 导致 loader 永不消失。
const loaderTimeout = new Promise<void>((resolve) => {
  window.setTimeout(resolve, 3000)
})
Promise.race([router.isReady().catch(() => undefined), loaderTimeout]).then(() => {
  startLoaderFade()
})
