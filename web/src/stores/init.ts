import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchGetInitStatus, fetchInitOwner } from '@/service/api'
import { localStg } from '@/utils/storage'

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
    const res = await fetchInitOwner(payload)
    if (res.code === 1) {
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
