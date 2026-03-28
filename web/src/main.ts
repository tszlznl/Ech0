import 'virtual:uno.css'
import '@/themes/index.scss'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import 'floating-vue/dist/style.css'

import { initStores } from './stores'
import { useSettingStore } from './stores/setting'
import { setupI18n } from './locales'
import FloatingVue from 'floating-vue'

// 自定义组件
import BaseDialog from '@/components/common/BaseDialog.vue'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// init
await initStores().catch((e) => {
  console.error('Failed to initialize stores:', e)
})

const settingStore = useSettingStore()
const i18n = await setupI18n(settingStore.SystemSetting.default_locale)

app.use(router)
app.use(i18n)
app.use(FloatingVue, {
  themes: {
    tooltip: {
      triggers: ['hover', 'focus'],
      hideTriggers: ['hover', 'focus', 'click'],
      placement: 'top',
      delay: { show: 300, hide: 80 },
      distance: 10,
      container: 'body',
      autoHide: true,
    },
  },
})

// 全局注册组件
app.component('BaseDialog', BaseDialog)

app.mount('#app')

// 启动加载页淡出并恢复页面滚动
const appLoader = document.getElementById('app-loader')
let loaderCleared = false
const clearStartupLoader = () => {
  if (loaderCleared) return
  loaderCleared = true
  appLoader?.remove()
  document.documentElement.classList.remove('app-loading')
}

if (appLoader) {
  window.requestAnimationFrame(() => {
    appLoader.classList.add('fade-out')
  })
  appLoader.addEventListener('transitionend', clearStartupLoader, { once: true })
  window.setTimeout(clearStartupLoader, 400)
} else {
  clearStartupLoader()
}
