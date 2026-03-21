<template>
  <div
    class="max-w-sm sm:max-w-full px-2 pb-4 py-2 mt-4 sm:mt-0 mb-10 sm:mb-0 mx-auto flex flex-col sm:flex-row justify-center items-start sm:items-stretch overflow-x-hidden sm:h-[100dvh] sm:overflow-hidden transition-all duration-500"
    :class="isZenMode ? 'sm:gap-0' : 'sm:gap-8'"
  >
    <div
      class="sm:max-w-sm w-full sm:min-h-0 sm:h-full sm:overflow-y-auto transition-opacity duration-900 ease-[cubic-bezier(0.22,1,0.36,1)]"
      :class="
        isZenMode
          ? 'sm:opacity-0 sm:invisible sm:pointer-events-none sm:w-0 sm:max-w-0 sm:overflow-hidden'
          : 'sm:opacity-100 sm:visible'
      "
    >
      <TheTop class="sm:hidden" />
      <TheEditor v-if="isLogin" />
      <TheBoard v-else />
      <div class="mt-3">
        <div class="hidden sm:flex sm:flex-col sm:gap-4">
          <TheHeatMap />
          <TheRecentCard v-if="AgentSetting.enable" />
          <TheConnectWidget />
          <TheCommentWidget />
        </div>
      </div>
    </div>
    <div
      ref="mainColumn"
      class="sm:max-w-lg w-full min-w-0 overflow-x-hidden sm:min-h-0 sm:h-full sm:overflow-y-auto sm:[overscroll-behavior:contain]"
      :style="{ '--date-sticky-top': echoDateStickyTop }"
      :class="isZenMode ? 'sm:mx-auto sm:shrink-0' : ''"
    >
      <div
        ref="topStickyBar"
        class="hidden sm:block sticky top-0 z-20 relative -mx-2 sm:-mx-4 md:-mx-6 px-2 sm:px-4 md:px-6 pt-2 bg-[var(--color-bg-canvas)]"
      >
        <TheTop class="sm:px-4" />
        <div class="w-full h-2 bg-[var(--color-bg-canvas)]"></div>
      </div>
      <TheEchos v-if="!isFilteringMode && !inboxMode" :scroll-target="mainColumn" />
      <TheFilteredEchos v-else-if="isFilteringMode && !inboxMode" :scroll-target="mainColumn" />
      <TheInbox v-else />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheTop from './TheTop.vue'
import TheEditor from './TheEditor.vue'
import TheBoard from './TheBoard.vue'
import TheEchos from './TheEchos.vue'
import TheFilteredEchos from './TheFilteredEchos.vue'
import TheInbox from './TheInbox.vue'
import {
  TheCommentWidget,
  TheConnectWidget,
  TheHeatMap,
  TheRecentCard,
} from '@/components/advanced/widget'
import { onMounted, ref, onBeforeUnmount } from 'vue'
import { useUserStore, useEchoStore, useSettingStore, useInboxStore, useZenStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'

const userStore = useUserStore()
const echoStore = useEchoStore()
const settingStore = useSettingStore()
const inboxStore = useInboxStore()
const zenStore = useZenStore()
const { isLogin } = storeToRefs(userStore)
const { isFilteringMode } = storeToRefs(echoStore)
const { AgentSetting } = storeToRefs(settingStore)
const { inboxMode } = storeToRefs(inboxStore)
const { isZenMode } = storeToRefs(zenStore)

const mainColumn = ref<HTMLElement | null>(null)
const topStickyBar = ref<HTMLElement | null>(null)
const echoDateStickyTop = ref('0px')
const backTopStyle = ref({ right: '100px' }) // 默认 fallback
const TIMELINE_SCROLL_KEY = 'home:timeline:scrollTop'
let timelineScrollRaf: number | null = null
let stickyBarObserver: ResizeObserver | null = null

const updateStickyTop = () => {
  if (window.innerWidth >= 640 && topStickyBar.value) {
    echoDateStickyTop.value = `${topStickyBar.value.offsetHeight - 1}px`
  } else {
    echoDateStickyTop.value = '0px'
  }
}

const updatePosition = () => {
  updateStickyTop()

  if (mainColumn.value) {
    const rect = mainColumn.value.getBoundingClientRect()
    const rightOffset = window.innerWidth - rect.right
    backTopStyle.value = {
      right: `${rightOffset - 160}px`,
    }
  }
}

const schedulePositionUpdate = () => {
  runWithBfCacheGuard(updatePosition, 120)
}

const saveTimelineScrollPosition = () => {
  if (!mainColumn.value || timelineScrollRaf !== null) return

  timelineScrollRaf = window.requestAnimationFrame(() => {
    timelineScrollRaf = null
    if (!mainColumn.value) return
    sessionStorage.setItem(TIMELINE_SCROLL_KEY, String(mainColumn.value.scrollTop))
  })
}

const restoreTimelineScrollPosition = () => {
  if (!mainColumn.value) return
  const raw = sessionStorage.getItem(TIMELINE_SCROLL_KEY)
  if (!raw) return
  const scrollTop = Number(raw)
  if (!Number.isFinite(scrollTop) || scrollTop < 0) return
  mainColumn.value.scrollTop = scrollTop
}

const { runWithBfCacheGuard } = useBfCacheRestore({
  onRestore: () => {
    schedulePositionUpdate()
  },
})

onMounted(async () => {
  schedulePositionUpdate()
  window.addEventListener('resize', schedulePositionUpdate)
  if (mainColumn.value) {
    mainColumn.value.scrollLeft = 0
    mainColumn.value.addEventListener('scroll', saveTimelineScrollPosition, { passive: true })
  }
  if (topStickyBar.value) {
    stickyBarObserver = new ResizeObserver(updateStickyTop)
    stickyBarObserver.observe(topStickyBar.value)
  }
  window.requestAnimationFrame(() => {
    restoreTimelineScrollPosition()
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', schedulePositionUpdate)
  if (mainColumn.value) {
    mainColumn.value.removeEventListener('scroll', saveTimelineScrollPosition)
  }
  if (timelineScrollRaf !== null) {
    window.cancelAnimationFrame(timelineScrollRaf)
    timelineScrollRaf = null
  }
  if (stickyBarObserver) {
    stickyBarObserver.disconnect()
    stickyBarObserver = null
  }
})
</script>
