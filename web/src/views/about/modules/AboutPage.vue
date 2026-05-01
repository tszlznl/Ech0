<template>
  <div class="about-page">
    <div class="about-shell">
      <header class="about-header">
        <BaseButton
          @click="$router.push({ name: 'home' })"
          :tooltip="t('commonNav.backHome')"
          class="about-back"
        >
          <Arrow class="text-2xl rotate-180" />
        </BaseButton>
        <div class="about-title">
          <h1 class="about-title__name">Ech0</h1>
          <span class="about-title__version">
            v{{ settingStore.hello?.version || '--' }}
          </span>
        </div>
        <p class="about-title__tagline">{{ t('about.tagline') }}</p>
      </header>

      <section class="about-card">
        <h2 class="about-card__heading">{{ t('about.copyrightHeading') }}</h2>
        <p class="about-card__text">
          {{ t('about.copyrightLine', { year: copyrightYears, holder: AUTHOR_NAME }) }}
        </p>
        <p class="about-card__text about-card__text--muted">
          {{ t('about.licenseLine') }}
          <a
            :href="`${REPO_URL}/blob/main/LICENSE`"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link"
          >
            AGPL-3.0-or-later
          </a>
        </p>
        <p class="about-card__text about-card__text--muted">
          {{ t('about.agplNotice') }}
        </p>
      </section>

      <section class="about-card">
        <h2 class="about-card__heading">{{ t('about.authorHeading') }}</h2>
        <p class="about-card__text">{{ AUTHOR_NAME }}</p>
        <div class="about-links">
          <a
            :href="AUTHOR_GITHUB"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link about-link--row"
          >
            <Github class="about-link__icon" />
            <span>{{ t('about.authorGithub') }}</span>
          </a>
        </div>
      </section>

      <section class="about-card">
        <h2 class="about-card__heading">{{ t('about.sourceHeading') }}</h2>
        <p class="about-card__text about-card__text--muted">
          {{ t('about.sourceDescription') }}
        </p>
        <div class="about-links">
          <a
            :href="REPO_URL"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link about-link--row"
          >
            <Github class="about-link__icon" />
            <span>{{ t('about.viewSource') }}</span>
          </a>
          <a
            :href="`${REPO_URL}/releases/tag/v${settingStore.hello?.version}`"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link about-link--row"
            v-if="settingStore.hello?.version"
          >
            <Info class="about-link__icon" />
            <span>{{ t('about.viewRelease', { version: settingStore.hello.version }) }}</span>
          </a>
        </div>
      </section>

      <footer class="about-footer">
        <span>{{ t('about.poweredBy') }}</span>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import Github from '@/components/icons/github.vue'
import Info from '@/components/icons/info.vue'

const { t } = useI18n()
const settingStore = useSettingStore()

const AUTHOR_NAME = 'lin-snow'
const AUTHOR_GITHUB = 'https://github.com/lin-snow'
const REPO_URL = 'https://github.com/lin-snow/Ech0'
const PROJECT_START_YEAR = 2024

const copyrightYears = computed(() => {
  const current = new Date().getFullYear()
  return current > PROJECT_START_YEAR ? `${PROJECT_START_YEAR}-${current}` : `${PROJECT_START_YEAR}`
})
</script>

<style scoped>
.about-page {
  display: flex;
  justify-content: center;
  width: 100%;
  min-height: 100vh;
  padding: 1.5rem 1rem 4rem;
}

.about-shell {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 42rem;
}

.about-header {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 1rem 1rem;
  text-align: center;
}

.about-back {
  position: absolute;
  top: 0.5rem;
  left: 0;
  border-radius: var(--radius-xs);
}

.about-title {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.about-title__name {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.about-title__version {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.about-title__tagline {
  margin: 0;
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
}

.about-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 1.125rem;
  background: var(--color-bg-surface);
  border-radius: var(--radius-xs);
  box-shadow: var(--shadow-soft);
}

.about-card__heading {
  margin: 0 0 0.25rem;
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.about-card__text {
  margin: 0;
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--color-text-primary);
}

.about-card__text--muted {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.about-links {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-top: 0.25rem;
}

.about-link {
  color: var(--color-text-primary);
  text-decoration: underline;
  text-decoration-color: var(--color-text-muted);
  text-underline-offset: 0.2em;
  transition: text-decoration-color 0.15s ease;
}

.about-link:hover {
  text-decoration-color: var(--color-text-primary);
}

.about-link--row {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  font-size: 0.9375rem;
}

.about-link--row:hover .about-link__icon {
  color: var(--color-text-primary);
}

.about-link__icon {
  font-size: 1.125rem;
  color: var(--color-text-secondary);
  transition: color 0.15s ease;
}

.about-footer {
  margin-top: 0.5rem;
  text-align: center;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}
</style>
