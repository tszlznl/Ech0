<template>
  <div class="home-page">
    <div class="home-shell" :class="{ 'home-shell--zen': isZenMode }">
      <div class="home-layout" :class="{ 'home-layout--zen': isZenMode }">
        <div
          ref="mainColumn"
          class="home-main"
          :style="{ '--date-sticky-top': echoDateStickyTop }"
        >
          <!-- 整个顶部区域作为一个 sticky 块，内部不再各自 sticky -->
          <div ref="topStickyBar" class="home-sticky">
            <div class="home-main-track">
              <HomeHeader />
              <HomeBanner />
            </div>
          </div>

          <div class="home-main-track">
            <TheEditor v-if="isLogin" />

            <TheEchos
              v-if="!inboxMode"
              compact
              :scroll-target="mainColumn"
            />
            <TheInbox v-else />

            <aside v-if="!isZenMode" class="home-aside home-aside--mobile">
              <HomeSidebarNav />
              <TheFilter />
              <p class="home-aside-foot">{{ sidebarFooterText }}</p>
            </aside>
          </div>
        </div>

        <aside v-if="!isZenMode" class="home-aside home-aside--rail">
          <HomeSidebarNav />
          <TheFilter />
          <p class="home-aside-foot">{{ sidebarFooterText }}</p>
        </aside>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import HomeHeader from './HomeHeader.vue'
import HomeBanner from './HomeBanner.vue'
import HomeSidebarNav from './HomeSidebarNav.vue'
import TheFilter from './TheFilter.vue'
import TheEchos from './TheEchos.vue'
import { defineAsyncComponent, onMounted, ref, onBeforeUnmount } from 'vue'
import { useUserStore, useInboxStore, useZenStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'
import { useI18n } from 'vue-i18n'

const TheEditor = defineAsyncComponent(() => import('./TheEditor.vue'))
const TheInbox = defineAsyncComponent(() => import('./TheInbox.vue'))

const { t } = useI18n()
const userStore = useUserStore()
const inboxStore = useInboxStore()
const zenStore = useZenStore()
const { isLogin } = storeToRefs(userStore)
const { inboxMode } = storeToRefs(inboxStore)
const { isZenMode } = storeToRefs(zenStore)

const sidebarFooterText = t('homeFooter.powered')

const mainColumn = ref<HTMLElement | null>(null)
const topStickyBar = ref<HTMLElement | null>(null)
const echoDateStickyTop = ref('0px')
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

<style scoped>
.home-page {
  --home-canvas: var(--color-bg-canvas, #f5f3ef);
  --home-accent: #e07020;
  --home-main-max: 28rem;
  min-height: 100dvh;
  background: var(--home-canvas);
  color: var(--color-text-primary);
}

.home-shell {
  max-width: 50rem;
  margin-left: auto;
  margin-right: auto;
  padding: 1rem 0.75rem 2.5rem;
}

@media (min-width: 640px) {
  .home-shell {
    padding: 1.25rem 1rem 2rem;
  }
}

@media (min-width: 820px) {
  .home-shell {
    padding: 1.5rem 1rem 2rem;
  }
}

.home-shell--zen .home-layout {
  max-width: min(var(--home-main-max), 100%);
  margin-left: auto;
  margin-right: auto;
}

.home-layout {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  align-items: stretch;
  background: transparent;
}

@media (min-width: 820px) {
  .home-layout {
    flex-direction: row;
    align-items: flex-start;
    justify-content: center;
    gap: clamp(1.25rem, 4vw, 2rem);
    padding: 0;
  }

  .home-layout--zen {
    justify-content: center;
  }
}

.home-main-track {
  width: 100%;
  max-width: var(--home-main-max);
  margin-left: auto;
  margin-right: auto;
}

.home-main {
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  overflow-x: visible;
}

@media (min-width: 820px) {
  .home-main {
    min-height: 0;
    height: 100%;
    max-height: 100dvh;
    overflow-y: auto;
    overscroll-behavior: contain;
    flex: 0 1 var(--home-main-max);
    max-width: var(--home-main-max);
  }

  .home-layout--zen .home-main {
    flex: 0 1 min(var(--home-main-max), 100%);
    max-width: min(var(--home-main-max), 100%);
    margin-left: auto;
    margin-right: auto;
  }
}

/* 整个顶部（头像 + 欢迎）作为一个 sticky 块 */
.home-sticky {
  position: sticky;
  top: 0;
  z-index: 30;
  padding-bottom: 0.25rem;
  background: var(--home-canvas);
}

.home-aside {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  width: 100%;
}

.home-aside--mobile {
  display: flex;
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border-subtle);
}

@media (min-width: 820px) {
  .home-aside--mobile {
    display: none !important;
  }

  .home-aside--rail {
    display: flex;
    position: sticky;
    top: 1rem;
    width: 14rem;
    flex-shrink: 0;
    align-self: flex-start;
    max-height: calc(100dvh - 2rem);
    overflow: auto;
  }
}

@media (max-width: 819.98px) {
  .home-aside--rail {
    display: none !important;
  }
}

.home-aside-foot {
  margin: 0;
  padding-top: 0.25rem;
  font-size: 0.75rem;
  line-height: 1.45;
  color: var(--color-text-muted);
}
</style>
