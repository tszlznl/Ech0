/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import UnoCSS from 'unocss/vite'
import viteCompression from 'vite-plugin-compression'

import { welcomePlugin } from './src/plugins/welcome-plugin'

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag === 'meting-js' || tag === 'cap-widget',
        },
      },
    }),
    ...(command === 'serve' ? [vueDevTools()] : []),
    UnoCSS(),
    viteCompression({
      deleteOriginFile: false,
    }),

    welcomePlugin(), // 欢迎横幅插件
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    include: ['tests/**/*.{test,spec}.ts'],
    clearMocks: true,
    restoreMocks: true,
  },
  build: {
    // 当使用embed时则调整构建输出到后端的template/dist目录
    outDir: '../template/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        // 代码分割：将重型库打包到单独的 chunk 中，利用浏览器缓存
        manualChunks: {
          // 代码高亮
          highlight: ['highlight.js'],
        },
      },
    },
  },
}))
