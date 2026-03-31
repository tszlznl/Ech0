<template>
  <section class="home-banner" aria-label="Intro">
    <div class="home-banner__top">
      <p class="home-banner__text">{{ t('homeBio.tagline') }}</p>
    </div>
    <div class="home-banner__meta">
      <span class="home-banner__version">v{{ settingStore.hello?.version || '--' }}</span>
      <span class="home-banner__powered">Powered by Ech0</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()
const { t } = useI18n()
</script>

<style scoped>
.home-banner {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 0.5rem;
  margin-bottom: 0.75rem;
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

.home-banner__meta {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.home-banner__version {
  font-size: 0.6875rem;
  line-height: 1;
  color: var(--color-text-muted);
  opacity: 0.68;
}

.home-banner__powered {
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
</style>
