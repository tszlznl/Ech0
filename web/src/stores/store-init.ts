import { useThemeStore } from './theme'
import { useUserStore } from './user'
import { useSettingStore } from './setting'
import { useEditorStore } from './editor'
import { useInitStore } from './init'

export async function initStores() {
  const themeStore = useThemeStore()
  const userStore = useUserStore()
  const settingStore = useSettingStore()
  const editorStore = useEditorStore()
  const initStore = useInitStore()

  themeStore.init()
  editorStore.init()

  // Fire non-critical settings fetches in parallel — they update reactive state
  // when they arrive but don't block the homepage first paint.
  settingStore.getSystemSetting().catch(() => undefined)
  settingStore.getAgentInfo().catch(() => undefined)
  settingStore.getHelloEch0().catch(() => undefined)

  // initStore & userStore must resolve before routing decisions, but they can
  // run concurrently (userStore's autoLogin does not depend on init status).
  await Promise.all([initStore.init(), userStore.init()])

  if (initStore.initialized && userStore.isLogin && userStore.user?.is_admin) {
    settingStore.getS3Setting().catch(() => undefined)
  }
}
