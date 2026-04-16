/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import UnoCSS from 'unocss/vite'
import viteCompression from 'vite-plugin-compression'

import { welcomePlugin } from './src/plugins/welcome-plugin'

// Inject modulepreload / prefetch hints into index.html for known chunk groups.
// Preload target (floating-vue) is on the bootstrap critical path; prefetching
// `highlight` + `uppy` populates HTTP cache during idle so later usage is free.
function preloadHintsPlugin(): Plugin {
  const PRELOAD_RE = /^floating-vue-[^.]+\.js$/
  const PREFETCH_RE = /^(highlight|uppy)-[^.]+\.js$/
  return {
    name: 'ech0-preload-hints',
    apply: 'build',
    transformIndexHtml: {
      order: 'post',
      handler(_html, ctx) {
        if (!ctx.bundle) return
        const tags: Array<{
          tag: string
          attrs: Record<string, string | boolean>
          injectTo: 'head'
        }> = []
        for (const fileName of Object.keys(ctx.bundle)) {
          const base = fileName.split('/').pop() ?? ''
          if (PRELOAD_RE.test(base)) {
            tags.push({
              tag: 'link',
              attrs: { rel: 'modulepreload', href: `/${fileName}` },
              injectTo: 'head',
            })
          } else if (PREFETCH_RE.test(base)) {
            tags.push({
              tag: 'link',
              attrs: {
                rel: 'prefetch',
                href: `/${fileName}`,
                as: 'script',
                crossorigin: '',
              },
              injectTo: 'head',
            })
          }
        }
        return tags
      },
    },
  }
}

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
      threshold: 10240,
      filter: (file) => /\.(js|mjs|css|html|svg)$/i.test(file),
    }),
    preloadHintsPlugin(),

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
    reportCompressedSize: false,
    rollupOptions: {
      output: {
        // 代码分割：将重型库打包到单独的 chunk 中，利用浏览器缓存
        manualChunks(id) {
          const normalizedId = id.replaceAll('\\', '/')
          if (normalizedId.includes('/node_modules/@uppy/')) {
            return 'uppy'
          }
          if (normalizedId.includes('/node_modules/floating-vue/')) {
            return 'floating-vue'
          }
          if (normalizedId.includes('/node_modules/highlight.js/')) {
            return 'highlight'
          }
          if (
            normalizedId.includes('/node_modules/markdown-it/') ||
            normalizedId.includes('/node_modules/linkify-it/') ||
            normalizedId.includes('/node_modules/mdurl/') ||
            normalizedId.includes('/node_modules/uc.micro/')
          ) {
            return 'markdown'
          }
          return undefined
        },
      },
    },
  },
}))
