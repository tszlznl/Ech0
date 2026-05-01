// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

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
  await initStore.init()
  await userStore.init()
  if (initStore.initialized) {
    await settingStore.init()
  }
  editorStore.init()
}
