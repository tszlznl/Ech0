// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import { VitePWA } from 'vite-plugin-pwa'

const hubDir = path.dirname(fileURLToPath(import.meta.url))
const webSrc = path.resolve(hubDir, '../web/src')
const webOtherShim = path.resolve(hubDir, 'src/shims/utils-other.ts')
const webOtherReal = path.resolve(webSrc, 'utils/other.ts')

// https://vite.dev/config/
export default defineConfig({
  /** 默认 5193，避免与 web（5173）、site（5183）同时开发时抢端口 */
  server: {
    port: 5193,
    strictPort: false,
  },
  resolve: {
    /** 必须在 `@` 之前：仅打包时把 web/utils/other 换成精简 shim，避免拉入 i18n 与整站依赖 */
    alias: [
      { find: webOtherReal, replacement: webOtherShim },
      { find: '@', replacement: webSrc },
    ],
  },
  css: {
    preprocessorOptions: {
      scss: {
        silenceDeprecations: ['legacy-js-api'],
      },
    },
  },
  plugins: [
    vue(),
    UnoCSS(),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'auto',
      /** 与 web/public 同套 PNG/ico（hub/public 内为拷贝） */
      includeAssets: [
        'favicon.ico',
        'favicon.svg',
        'logo.svg',
        'icons.svg',
        'android-chrome-192x192.png',
        'android-chrome-512x512.png',
        'apple-touch-icon.png',
        'maskable-icon.png',
        'web-app-manifest-192x192.png',
        'web-app-manifest-512x512.png',
      ],
      manifest: {
        id: '/',
        name: 'Ech0 Hub',
        short_name: 'Ech0 Hub',
        description:
          'Discover and connect with resonating voices from public Ech0 instances — one feed, many sites.',
        start_url: '/',
        scope: '/',
        display: 'standalone',
        background_color: '#f4f1ec',
        theme_color: '#f4f1ec',
        icons: [
          {
            src: '/web-app-manifest-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/web-app-manifest-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/maskable-icon.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
          {
            src: '/apple-touch-icon.png',
            sizes: '180x180',
            type: 'image/png',
            purpose: 'any',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,svg,png,woff2}'],
        cleanupOutdatedCaches: true,
      },
      devOptions: {
        enabled: false,
      },
    }),
  ],
})
