<template>
  <section class="home-banner" aria-label="Intro">
    <div class="home-banner__top">
      <p class="home-banner__line">{{ t('homeBio.tagline') }}</p>
    </div>
    <div class="home-banner__meta">
      <RouterLink
        :to="{ name: 'about' }"
        class="home-banner__powered"
        :aria-label="t('about.linkAriaLabel')"
      >
        Powered by Ech0
      </RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
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

@media (width <= 420px) {
  .home-banner {
    flex-wrap: wrap;
  }
}

.home-banner__line {
  margin: 0;
  font-family: 'Songti SC', STSong, var(--font-family-display);
  font-weight: 550;
  letter-spacing: 0.01em;
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--color-text-secondary);
}

.home-banner__powered {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.35;
  color: var(--color-text-secondary);
  text-decoration: none;
  cursor: pointer;
  transition: color 0.15s ease;
}

.home-banner__powered:hover {
  color: var(--color-text-primary);
}
</style>
