<template>
  <div
    class="mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
    :class="compact ? 'pl-1 pr-0 max-w-full' : 'px-2 sm:px-4 md:px-6 max-w-full'"
  >
    <!-- Echos - 使用 TransitionGroup 实现入场动画 -->
    <TransitionGroup
      v-if="echoStore.echoList"
      name="list"
      tag="div"
      class="relative"
      @before-enter="onBeforeEnter"
      @enter="onEnter"
      @before-leave="onBeforeLeave"
      @leave="onLeave"
    >
      <div v-for="(echo, index) in echoStore.echoList" :key="echo.id" class="will-change-transform">
        <TheEchoCard :echo="echo" :index="index" @refresh="handleRefresh" />
      </div>
    </TransitionGroup>
    <!-- 加载更多 -->
    <Transition name="fade">
      <div
        v-if="echoStore.hasMore && !echoStore.isLoading"
        class="mb-4 mt-1 flex items-center justify-between echos-toolbar"
      >
        <BaseButton
          v-if="!isZenMode"
          @click="handleLoadMore"
          class="rounded-full bg-[var(--btn-bg-color)] !active:bg-[var(--btn-hover-bg-color)] mr-2"
        >
          <span class="text-[var(--btn-text-color)] text-md text-center px-2 py-1">{{
            t('homeFeed.loadMore')
          }}</span>
        </BaseButton>
        <TheBackTop class="w-8 h-8 p-1" :target="scrollTarget" />
      </div>
    </Transition>
    <!-- 没有更多 -->
    <Transition name="fade">
      <div
        v-if="!echoStore.hasMore && !echoStore.isLoading"
        class="mx-auto my-5 text-center echos-toolbar"
      >
        <p class="text-xl text-[var(--color-text-muted)]">
          {{ echoStore.isFilteringMode ? t('homeFeed.noMoreFiltered') : t('homeFeed.noMore') }}
        </p>
      </div>
    </Transition>
    <!-- 加载中 -->
    <Transition name="fade">
      <TheLoadingIndicator
        v-if="echoStore.isLoading"
        class="mx-auto my-5 echos-toolbar"
        size="lg"
        :label="t('homeFeed.loading')"
      />
    </Transition>
    <!-- 自定义页脚（紧跟时间线内容之后） -->
    <div v-if="footerContent" class="mt-6 text-center">
      <a v-if="footerLink" :href="footerLink" target="_blank" rel="noopener noreferrer">
        <span class="text-[var(--color-text-muted)] text-sm">
          {{ footerContent }}
        </span>
      </a>
      <span v-else class="text-[var(--color-text-muted)] text-sm">
        {{ footerContent }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import { computed, onBeforeUnmount, onMounted, nextTick, ref, watch } from 'vue'
import { useEchoStore, useSettingStore, useZenStore } from '@/stores'
import BaseButton from '@/components/common/BaseButton.vue'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import { storeToRefs } from 'pinia'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  scrollTarget?: HTMLElement | null
  /** 首页窄栏：减少左右留白以贴合参考图时间线宽度 */
  compact?: boolean
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const zenStore = useZenStore()
const { t } = useI18n()
const { SystemSetting } = storeToRefs(settingStore)
const { isZenMode } = storeToRefs(zenStore)
const { isFilteringMode } = storeToRefs(echoStore)
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)

// 瀑布式入场：首屏与后续刷新 / load-more 都走"从上往下依次落下"的级联。
// 通过批内计数器而不是全局 index 来计算 stagger，
// 这样 load-more 追加的新卡也能在自己的批次里正确瀑布展开，而不是一起顶到封顶延迟。
const hasInitialRendered = ref(false)

let enterBatchIndex = 0
let enterBatchResetTimer: number | null = null

const ENTER_DURATION = 420
const ENTER_STAGGER = 100
const ENTER_STAGGER_CAP = 600
const ENTER_EASING = 'cubic-bezier(0.22, 1, 0.36, 1)' // ease-out-quart，轻柔落地

const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.opacity = '0'
  element.style.transform = 'translateY(18px)'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  const indexInBatch = enterBatchIndex++
  if (enterBatchResetTimer !== null) {
    window.clearTimeout(enterBatchResetTimer)
  }
  // 一批 enter 结束后重置计数器，保证下一批（刷新、load-more）从 0 重新起跳
  enterBatchResetTimer = window.setTimeout(() => {
    enterBatchIndex = 0
    enterBatchResetTimer = null
  }, 400)

  // 过滤切换 / 点赞刷新等场景有 leave 同时进行，预留 80ms 让 leave 脱流 + move 起步
  const baseDelay = hasInitialRendered.value ? 80 : 0
  const staggerDelay = Math.min(indexInBatch * ENTER_STAGGER, ENTER_STAGGER_CAP)

  window.setTimeout(() => {
    element.style.transition = `opacity ${ENTER_DURATION}ms ${ENTER_EASING}, transform ${ENTER_DURATION}ms ${ENTER_EASING}`
    element.style.opacity = '1'
    element.style.transform = 'translateY(0)'
    window.setTimeout(done, ENTER_DURATION)
  }, baseDelay + staggerDelay)
}

// 离场：先把元素位置快照并切成 absolute 脱离文档流，这样留下的卡片可以立刻
// 通过 .list-move 滑到新位置，新进来的卡片也不会和还没消失的旧卡重叠。
const onBeforeLeave = (el: Element) => {
  const element = el as HTMLElement
  const parent = element.parentElement
  if (!parent) return
  const rect = element.getBoundingClientRect()
  const parentRect = parent.getBoundingClientRect()
  element.style.position = 'absolute'
  element.style.left = `${rect.left - parentRect.left}px`
  element.style.top = `${rect.top - parentRect.top}px`
  element.style.width = `${rect.width}px`
}

const onLeave = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  window.requestAnimationFrame(() => {
    element.style.transition = 'opacity 0.18s ease'
    element.style.opacity = '0'
    window.setTimeout(done, 180)
  })
}

const handleLoadMore = async () => {
  echoStore.current = echoStore.current + 1
  await echoStore.getEchosByPage()
}

const nearBottomThreshold = 240
let scrollListenerAttachedEl: HTMLElement | null = null
let rafId: number | null = null
let ensuringScrollable = false

const getScrollMetrics = (container: HTMLElement) => ({
  scrollTop: container.scrollTop,
  viewportHeight: container.clientHeight,
  fullHeight: container.scrollHeight,
})

const checkAndLoadMoreInZen = async () => {
  const container = props.scrollTarget
  if (!container || !isZenMode.value || echoStore.isLoading || !echoStore.hasMore) return
  const { scrollTop, viewportHeight, fullHeight } = getScrollMetrics(container)
  if (scrollTop + viewportHeight + nearBottomThreshold >= fullHeight) {
    await handleLoadMore()
  }
}

const onTimelineScroll = () => {
  if (!isZenMode.value || rafId !== null) return
  rafId = window.requestAnimationFrame(async () => {
    rafId = null
    await checkAndLoadMoreInZen()
  })
}

const bindTimelineScroll = () => {
  if (scrollListenerAttachedEl === props.scrollTarget) return
  if (scrollListenerAttachedEl) {
    scrollListenerAttachedEl.removeEventListener('scroll', onTimelineScroll)
  }
  scrollListenerAttachedEl = props.scrollTarget ?? null
  if (scrollListenerAttachedEl) {
    scrollListenerAttachedEl.addEventListener('scroll', onTimelineScroll, { passive: true })
  }
}

const ensureScrollableInZen = async () => {
  if (ensuringScrollable || !isZenMode.value) return
  const container = props.scrollTarget
  if (!container) return
  ensuringScrollable = true
  try {
    const maxAutoLoads = 3
    let attempts = 0
    while (attempts < maxAutoLoads && echoStore.hasMore && !echoStore.isLoading) {
      await nextTick()
      const { viewportHeight, fullHeight } = getScrollMetrics(container)
      if (fullHeight > viewportHeight + 10) break
      attempts += 1
      await handleLoadMore()
    }
  } finally {
    ensuringScrollable = false
  }
}

// 刷新数据
const handleRefresh = () => {
  echoStore.refreshEchos()
}

onMounted(async () => {
  // 获取数据
  bindTimelineScroll()
  // main.ts 在 `/` 路由上预热了第一页请求，`getEchosByPage` 内置了
  // `current <= page` 的重复请求守卫，所以这里直接调用即可：
  // 若预热已完成则快速返回，否则接着完成加载。
  await echoStore.getEchosByPage()
  await ensureScrollableInZen()
  // 首屏交错动画跑完（最长 ~460ms）后切到淡入模式，避免后续过滤刷新抖动
  window.setTimeout(() => {
    hasInitialRendered.value = true
  }, 500)
})

watch(
  () => props.scrollTarget,
  () => {
    bindTimelineScroll()
    ensureScrollableInZen()
  },
)

watch(isZenMode, () => {
  ensureScrollableInZen()
})

watch(
  () => [echoStore.echoList.length, echoStore.hasMore, echoStore.isLoading],
  () => {
    ensureScrollableInZen()
  },
)

// 过滤模式切换时（进入/退出/切换标签），刷新列表
watch(isFilteringMode, () => {
  echoStore.refreshEchos()
})

onBeforeUnmount(() => {
  if (scrollListenerAttachedEl) {
    scrollListenerAttachedEl.removeEventListener('scroll', onTimelineScroll)
    scrollListenerAttachedEl = null
  }
  if (rafId !== null) {
    window.cancelAnimationFrame(rafId)
    rafId = null
  }
})
</script>

<style scoped>
.echos-toolbar {
  font-family: var(--font-family-display);
}

/* 留下的卡片位置变化：用于过滤/刷新时平滑补位 */
.list-move {
  transition: transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1);
}

/* 离场动画完全在 JS 里处理（onBeforeLeave / onLeave）以便快照位置后脱离文档流。 */

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
