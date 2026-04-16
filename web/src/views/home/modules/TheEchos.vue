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
      @before-enter="onBeforeEnter"
      @enter="onEnter"
    >
      <div
        v-for="(echo, index) in echoStore.echoList"
        :key="echo.id"
        :data-index="index"
        class="will-change-transform"
      >
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
import { computed, onBeforeUnmount, onMounted, nextTick, watch } from 'vue'
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

// 列表入场动画钩子 - 交错入场效果
const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.opacity = '0'
  element.style.transform = 'translateY(20px)'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  const index = Number(element.dataset.index) || 0
  // 交错延迟：每个元素延迟 50ms，最大延迟 250ms
  const delay = Math.min(index * 50, 250)

  setTimeout(() => {
    element.style.transition = 'opacity 0.3s ease, transform 0.3s ease'
    element.style.opacity = '1'
    element.style.transform = 'translateY(0)'

    // 动画结束后调用 done
    setTimeout(done, 300)
  }, delay)
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

/* 列表项移动动画 */
.list-move {
  transition: transform 0.3s ease;
}

/* 列表项离开动画 */
.list-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.list-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

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
