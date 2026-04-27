import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchGetInitStatus, fetchInitOwner } from '@/service/api'
import { i18n } from '@/locales'
import { localStg } from '@/utils/storage'

const INIT_ALREADY_DONE = 'INIT_ALREADY_DONE'
const INIT_OWNER_EXISTS = 'INIT_OWNER_EXISTS'

export const useInitStore = defineStore('initStore', () => {
  const initialized = ref<boolean>(false)
  const ownerExists = ref<boolean>(false)
  const ready = ref<boolean>(false)

  const saveCache = () => {
    localStg.setItem('initialized', initialized.value)
  }

  const clearCache = () => {
    localStg.removeItem('initialized')
  }

  const getStatus = async () => {
    const res = await fetchGetInitStatus()
    if (res.code === 1) {
      initialized.value = res.data.initialized
      ownerExists.value = res.data.owner_exists
      ready.value = true
      saveCache()
    } else {
      ready.value = false
      clearCache()
    }
    return res
  }

  const initOwner = async (payload: App.Api.Auth.SignupParams) => {
    // 把部署者当前页面生效的 locale（来自 navigator 检测或手动切换）一起提交，
    // 后端会用它作为 owner.locale 与站点 default_locale，避免新部署被锁成 zh-CN。
    const enriched: App.Api.Auth.SignupParams = {
      ...payload,
      locale: payload.locale || String(i18n.global.locale.value),
    }
    const res = await fetchInitOwner(enriched)
    if (res.code === 1) {
      initialized.value = true
      ownerExists.value = true
      ready.value = true
      saveCache()
    } else if (res.error_code === INIT_ALREADY_DONE || res.error_code === INIT_OWNER_EXISTS) {
      // 服务端明确返回“已初始化/Owner已存在”时，直接更新本地状态，避免被陈旧缓存卡在 init 页。
      initialized.value = true
      ownerExists.value = true
      ready.value = true
      saveCache()
    } else {
      // 提交初始化失败时同步一次服务端状态，避免并发初始化后前端状态滞后。
      await getStatus().catch(() => undefined)
    }
    return res
  }

  const init = async () => {
    const cached = localStg.getItem<boolean>('initialized')
    if (cached !== null) {
      initialized.value = cached
    }
    await getStatus()
  }

  return {
    initialized,
    ownerExists,
    ready,
    getStatus,
    initOwner,
    init,
  }
})
