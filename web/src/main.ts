import 'virtual:uno.css'
import '@/themes/index.scss'
import 'floating-vue/dist/style.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { initStores } from './stores/store-init'
import { useSettingStore } from './stores/setting'
import { setupI18n } from './locales'

// 自定义组件
import BaseDialog from '@/components/common/BaseDialog.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// init
await initStores().catch((e) => {
  console.error('Failed to initialize stores:', e)
})

const settingStore = useSettingStore()
const i18n = await setupI18n(settingStore.SystemSetting.default_locale)
const { default: FloatingVue } = await import('floating-vue')

app.use(router)
app.use(i18n)
app.use(FloatingVue, {
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
