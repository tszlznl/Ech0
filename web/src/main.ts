import 'virtual:uno.css'
import '@/themes/index.scss'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

import { initStores } from './stores'

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

app.use(router)

// 全局注册组件
app.component('BaseDialog', BaseDialog)

app.mount('#app')

// 移除启动加载动画
const appLoader = document.getElementById('app-loader')
if (appLoader) {
  appLoader.remove()
}
