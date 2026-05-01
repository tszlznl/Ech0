// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createI18n } from 'vue-i18n'
import enUS from '../../web/src/locales/messages/en-US.json'

const enUSHub = {
  ...enUS,
  hub: {
    ...enUS.hub,
    emptyConnectHint:
      'Nothing here yet—public instances may have no posts, or feeds are still loading.',
    layoutSwitch: 'Feed layout',
    layoutList: 'Single column',
    layoutMasonry: 'Masonry grid',
  },
}

export function createHubI18n() {
  return createI18n({
    legacy: false,
    locale: 'en-US',
    fallbackLocale: 'en-US',
    messages: {
      'en-US': enUSHub,
    },
  })
}
