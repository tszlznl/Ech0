import { useThemeStore } from './theme'
import { useUserStore } from './user'
import { useSettingStore } from './setting'
import { useEchoStore } from './echo'
import { useEditorStore } from './editor'
import { useInitStore } from './init'

export async function initStores() {
  const themeStore = useThemeStore()
  const userStore = useUserStore()
  const settingStore = useSettingStore()
  const echoStore = useEchoStore()
  const editorStore = useEditorStore()
  const initStore = useInitStore()

  themeStore.init()
  await initStore.init()
  await userStore.init()
  if (initStore.initialized) {
    await settingStore.init()
  }
  editorStore.init()
  echoStore.init()
}
