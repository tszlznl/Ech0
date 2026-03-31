<template>
  <div class="home-page">
    <div class="home-shell" :class="{ 'home-shell--zen': isZenMode }">
      <div class="home-layout" :class="{ 'home-layout--zen': isZenMode }">
        <div
          ref="mainColumn"
          class="home-main"
        >
          <div class="home-main-track">
            <HomeHeader />
            <HomeBanner />
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
            </aside>
          </div>
        </div>

        <aside v-if="!isZenMode" class="home-aside home-aside--rail">
          <HomeSidebarNav />
          <TheFilter />
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

const TheEditor = defineAsyncComponent(() => import('./TheEditor.vue'))
const TheInbox = defineAsyncComponent(() => import('./TheInbox.vue'))

const userStore = useUserStore()
const inboxStore = useInboxStore()
const zenStore = useZenStore()
const { isLogin } = storeToRefs(userStore)
const { inboxMode } = storeToRefs(inboxStore)
const { isZenMode } = storeToRefs(zenStore)

const mainColumn = ref<HTMLElement | null>(null)
const TIMELINE_SCROLL_KEY = 'home:timeline:scrollTop'
let timelineScrollRaf: number | null = null

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

onMounted(async () => {
  if (mainColumn.value) {
    mainColumn.value.scrollLeft = 0
    mainColumn.value.addEventListener('scroll', saveTimelineScrollPosition, { passive: true })
  }
  window.requestAnimationFrame(() => {
    restoreTimelineScrollPosition()
  })
})

onBeforeUnmount(() => {
  if (mainColumn.value) {
    mainColumn.value.removeEventListener('scroll', saveTimelineScrollPosition)
  }
  if (timelineScrollRaf !== null) {
    window.cancelAnimationFrame(timelineScrollRaf)
    timelineScrollRaf = null
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

@media (min-width: 820px) {
  .home-page {
    height: 100dvh;
    overflow: hidden;
  }
}

.home-shell {
  max-width: 50rem;
  margin: 1rem auto 2.5rem;
  padding: 0 0.75rem;
}

@media (min-width: 640px) {
  .home-shell {
    margin-top: 1.25rem;
    margin-bottom: 2rem;
    padding: 0 1rem;
  }
}

@media (min-width: 820px) {
  .home-shell {
    margin: 0 auto;
    padding: 0 1rem;
    display: flex;
    flex-direction: column;
    height: 100%;
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
    flex: 1;
    min-height: 0;
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
    align-self: stretch;
    overflow-y: auto;
    overscroll-behavior: contain;
    flex: 0 1 var(--home-main-max);
    max-width: var(--home-main-max);
    padding: 1.5rem 0 2rem;
  }

  .home-layout--zen .home-main {
    flex: 0 1 min(var(--home-main-max), 100%);
    max-width: min(var(--home-main-max), 100%);
    margin-left: auto;
    margin-right: auto;
  }
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
    width: 14rem;
    flex-shrink: 0;
    align-self: flex-start;
    margin-top: 1.5rem;
  }
}

@media (max-width: 819.98px) {
  .home-aside--rail {
    display: none !important;
  }
}

</style>
