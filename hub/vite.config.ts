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
      includeAssets: ['favicon.svg', 'icons.svg'],
      manifest: {
        name: 'Ech0 Hub',
        short_name: 'Ech0 Hub',
        description: 'Aggregate public Ech0 instances',
        start_url: '/',
        display: 'standalone',
        background_color: '#f4f1ec',
        theme_color: '#f4f1ec',
        icons: [
          {
            src: '/favicon.svg',
            sizes: 'any',
            type: 'image/svg+xml',
            purpose: 'any',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,svg,woff2}'],
      },
      devOptions: {
        enabled: true,
      },
    }),
  ],
})
