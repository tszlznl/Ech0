<template>
  <section class="home-banner" aria-label="Intro">
    <div class="home-banner__top">
      <p class="home-banner__text">{{ t('homeBio.tagline') }}</p>
      <div class="home-banner__actions">
        <button
          type="button"
          v-tooltip="themeToggleTooltip"
          :aria-label="t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel })"
          class="home-banner__btn"
          @click="handleThemeToggle"
        >
          <component :is="themeIcon" class="w-5 h-5" />
        </button>
        <button
          type="button"
          v-tooltip="isZenMode ? t('homeTop.exitZenMode') : t('homeNav.enterZenMode')"
          :aria-label="isZenMode ? t('homeTop.exitZenMode') : t('homeNav.enterZenMode')"
          :class="['home-banner__btn', isZenMode ? 'home-banner__btn--active' : '']"
          @click="handleToggleZenMode"
        >
          <Zen class="block w-5 h-5" />
        </button>
      </div>
    </div>
    <span class="home-banner__powered">Powered by Ech0</span>
  </section>
</template>

<script setup lang="ts">
import LightIcon from '@/components/icons/light.vue'
import DarkIcon from '@/components/icons/dark.vue'
import AutoIcon from '@/components/icons/auto.vue'
import Zen from '@/components/icons/zen.vue'
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore, useZenStore } from '@/stores'
import { theToast } from '@/utils/toast'

const themeStore = useThemeStore()
const zenStore = useZenStore()
const { isZenMode } = storeToRefs(zenStore)
const { t } = useI18n()

const nextThemeMode = computed(() => {
  if (themeStore.mode === 'system') return 'light'
  if (themeStore.mode === 'light') return 'dark'
  return 'system'
})

const themeIcon = computed(() => {
  if (nextThemeMode.value === 'light') return LightIcon
  if (nextThemeMode.value === 'dark') return DarkIcon
  return AutoIcon
})

const nextThemeModeLabel = computed(() => {
  if (nextThemeMode.value === 'light') return String(t('homeNav.themeLight'))
  if (nextThemeMode.value === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeAuto'))
})

const themeToggleTooltip = computed(() => ({
  content: String(t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel.value })),
  triggers: ['hover'],
  hideTriggers: ['hover', 'click'],
}))

const getThemeModeLabel = () => {
  if (themeStore.mode === 'light') return String(t('homeNav.themeLight'))
  if (themeStore.mode === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeAuto'))
}

const handleThemeToggle = async (event: MouseEvent) => {
  await themeStore.toggleTheme(event)
  theToast.success(String(t('homeNav.themeSwitched')), {
    description: String(t('homeNav.themeCurrent', { mode: getThemeModeLabel() })),
    duration: 1500,
  })
}

const handleToggleZenMode = () => {
  zenStore.setZenMode(!isZenMode.value)
}

</script>

<style scoped>
.home-banner {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 0.5rem;
  min-height: 6rem;
  padding: 0.75rem;
  border-radius: var(--radius-xs);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-soft);
}

.home-banner__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.home-banner__powered {
  align-self: flex-end;
  font-size: 0.6875rem;
  line-height: 1;
  color: var(--color-text-muted);
  opacity: 0.6;
}

@media (max-width: 420px) {
  .home-banner {
    flex-wrap: wrap;
  }
}

.home-banner__text {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--color-text-secondary);
}

.home-banner__actions {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  border-radius: 9999px;
  background: var(--input-bg-color);
  overflow: hidden;
}

.home-banner__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 2rem;
  padding: 0 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
  transition:
    color 0.2s,
    background 0.2s;
}

.home-banner__btn:hover:not(:disabled) {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}

.home-banner__btn--active {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}
</style>
