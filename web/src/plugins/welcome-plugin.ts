// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Plugin } from 'vite'
import { printWelcome } from '../scripts/welcome.js'

export function welcomePlugin(): Plugin {
  let hasShown = false

  return {
    name: 'welcome-banner',
    configureServer(server) {
      // 监听服务器启动事件
      server.middlewares.use('/', (req, res, next) => {
        if (!hasShown) {
          // 延迟显示，确保 Vite 的启动信息已经输出完成
          printWelcome()
          hasShown = true
        }
        next()
      })
    },
  }
}
