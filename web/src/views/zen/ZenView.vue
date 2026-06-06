<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="zen-view">
    <header class="zen-view__bar">
      <div class="zen-view__pill">
        <img
          :src="logoUrl"
          :alt="serverName"
          class="zen-view__pill-logo"
          loading="lazy"
          decoding="async"
        />
        <button
          type="button"
          v-tooltip="bwToggleLabel"
          :aria-label="bwToggleLabel"
          :aria-pressed="bwMode"
          class="zen-view__pill-toggle"
          :class="{ 'zen-view__pill-toggle--active': bwMode }"
          @click="toggleBW"
        >
          <Contrast class="w-4 h-4" />
        </button>
      </div>

      <button
        type="button"
        v-tooltip="t('zenMode.exit')"
        :aria-label="t('zenMode.exit')"
        class="zen-view__exit"
        @click="exit"
      >
        <Close class="w-5 h-5" />
      </button>
    </header>

    <main
      v-if="zenStore.echoList.length > 0"
      class="zen-view__grid"
      :class="{ 'zen-view__grid--bw': bwMode }"
    >
      <div
        v-for="(echo, i) in zenStore.echoList"
        :key="echo.id"
        :ref="(el) => registerItem(el as HTMLElement | null)"
        class="zen-view__cell"
      >
        <TheZenEchoCard :echo="echo" :index="i" />
      </div>
    </main>

    <div v-if="zenStore.isLoading" class="zen-view__status">
      <TheLoadingIndicator size="md" />
    </div>
    <div
      v-else-if="!zenStore.hasMore && zenStore.echoList.length > 0"
      class="zen-view__status zen-view__status--end"
    >
      {{ t('zenMode.noMore') }}
    </div>
    <div
      v-else-if="!zenStore.isLoading && zenStore.echoList.length === 0"
      class="zen-view__status zen-view__status--empty"
    >
      {{ t('zenMode.empty') }}
    </div>

    <div ref="sentinelRef" class="zen-view__sentinel" aria-hidden="true" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { useSettingStore, useZenStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'
import TheZenEchoCard from '@/components/advanced/echo/cards/TheZenEchoCard.vue'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import Close from '@/components/icons/close.vue'
import Contrast from '@/components/icons/contrast.vue'

const { t } = useI18n()
const router = useRouter()
const zenStore = useZenStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const serverName = computed(() => String(SystemSetting.value?.server_name ?? 'Ech0'))
const logoUrl = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))

const sentinelRef = ref<HTMLElement | null>(null)
let sentinelObserver: IntersectionObserver | null = null

const bwMode = ref<boolean>(false)
const toggleBW = () => {
  bwMode.value = !bwMode.value
  theToast.success(String(bwMode.value ? t('zenMode.bwToastOn') : t('zenMode.bwToastOff')))
}
// 用两个静态 t() 调用，让 i18n unused-check 能识别 key 被引用。
const bwOnLabel = computed(() => t('zenMode.bwOn'))
const bwOffLabel = computed(() => t('zenMode.bwOff'))
const bwToggleLabel = computed(() => (bwMode.value ? bwOffLabel.value : bwOnLabel.value))

const onSentinelVisible = () => {
  if (zenStore.isLoading) return
  if (zenStore.echoList.length > 0 && !zenStore.hasMore) return
  zenStore.loadNextPage()
}

const setupSentinelObserver = () => {
  teardownSentinelObserver()
  const el = sentinelRef.value
  if (!el) return
  sentinelObserver = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        onSentinelVisible()
      }
    },
    { root: null, rootMargin: '0px 0px 600px 0px' },
  )
  sentinelObserver.observe(el)
}

const teardownSentinelObserver = () => {
  sentinelObserver?.disconnect()
  sentinelObserver = null
}

// ─────────────────────────────────────────────
// Masonry：CSS Grid + ResizeObserver 测高分配 row span
// 比 CSS columns 可靠：每张卡片测自己的高，按 1px 行单位算 span，
// 图片懒加载完成时自动重排，永远不会出现"右列空、左列堆满"的失衡。
// ─────────────────────────────────────────────
const ROW_HEIGHT_PX = 1
const ROW_GAP_PX = 16

const observedItems = new Set<HTMLElement>()
let resizeObserver: ResizeObserver | null = null

const recomputeSpan = (cell: HTMLElement) => {
  const inner = cell.firstElementChild as HTMLElement | null
  if (!inner) return
  const h = inner.getBoundingClientRect().height
  if (h <= 0) return
  const span = Math.max(1, Math.ceil((h + ROW_GAP_PX) / (ROW_HEIGHT_PX + ROW_GAP_PX)))
  cell.style.gridRowEnd = `span ${span}`
}

const ensureResizeObserver = () => {
  if (resizeObserver || typeof ResizeObserver === 'undefined') return
  resizeObserver = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const cell = (entry.target as HTMLElement).parentElement
      if (cell) recomputeSpan(cell)
    }
  })
}

// Vue ref function：每张卡的 cell wrapper mount/unmount 都会调一次。
// 用一个 Set 去重，避免对同一元素重复 observe（v-for 重渲染时 Vue 会先卸再装）。
const registerItem = (el: HTMLElement | null) => {
  if (!el) return
  if (observedItems.has(el)) return
  observedItems.add(el)
  ensureResizeObserver()
  const inner = el.firstElementChild as HTMLElement | null
  if (inner && resizeObserver) {
    resizeObserver.observe(inner)
  }
  nextTick(() => recomputeSpan(el))
}

const exit = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push({ name: 'home' })
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    event.preventDefault()
    exit()
  }
}

watch(
  () => zenStore.echoList.length,
  () => {
    nextTick(() => {
      setupSentinelObserver()
    })
  },
  { flush: 'post' },
)

onMounted(async () => {
  zenStore.reset()
  await zenStore.loadNextPage()
  await nextTick()
  setupSentinelObserver()
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  teardownSentinelObserver()
  resizeObserver?.disconnect()
  resizeObserver = null
  observedItems.clear()
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.zen-view {
  min-height: 100dvh;
  background: var(--color-bg-canvas);
  color: var(--color-text-primary);
  padding: 0 1.25rem 3rem;
}

.zen-view__bar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1.25rem;
  margin: 0 -1.25rem 0.75rem;

  /* 同 canvas 色，纯色不透明；横向拉满抵消父容器 padding，避免边缘漏出卡片 */
  background-color: var(--color-bg-canvas);
  isolation: isolate;
}

/* 左侧胶囊：logo + 黑白切换按钮 */
.zen-view__pill {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.5rem 0.25rem 0.25rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 9999px;
  box-shadow: 0 1px 2px rgb(0 0 0 / 4%);
}

.zen-view__pill-logo {
  width: 1.875rem;
  height: 1.875rem;
  border-radius: 9999px;
  object-fit: cover;
  flex-shrink: 0;
}

.zen-view__pill-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.625rem;
  height: 1.625rem;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--color-text-muted);
  border-radius: 9999px;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease,
    transform 0.18s ease;
}

.zen-view__pill-toggle:hover {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}

.zen-view__pill-toggle--active {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}

.zen-view__pill-toggle:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

/* 右上角退出按钮 */
.zen-view__exit {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  margin-left: auto;
  border: 0;
  background: transparent;
  color: var(--color-text-secondary);
  border-radius: 0.5rem;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease;
}

.zen-view__exit:hover {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}

.zen-view__exit:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

/* JS-driven masonry：每个 cell 的 grid-row span 由脚本按真实高度算
   minmax(min(100%, 22rem), 1fr)：列至少 22rem，否则一行一列；
   配合 grid-auto-rows 1px + 1rem gap，让 span 数学正好对齐 */
.zen-view__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 22rem), 1fr));
  grid-auto-rows: 1px;
  gap: 1rem;
  align-items: start;
}

.zen-view__cell {
  /* gridRowEnd 由 JS 设置；此处只声明默认占位，避免首帧抖动 */
  grid-row-end: span 30;
}

/* 黑白模式：所有卡片 grayscale，hover 单卡恢复彩色（"对比聚焦"效果） */
.zen-view__grid--bw .zen-view__cell {
  filter: grayscale(1);
  transition: filter 0.3s ease;
}

.zen-view__grid--bw .zen-view__cell:hover {
  filter: grayscale(0);
}

.zen-view__status {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1.5rem 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-display);
}

.zen-view__status--end,
.zen-view__status--empty {
  letter-spacing: 0.02em;
}

.zen-view__sentinel {
  height: 1px;
}
</style>
