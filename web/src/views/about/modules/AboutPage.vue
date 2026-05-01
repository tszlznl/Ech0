<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
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
            v{{ version }}
            <span v-if="hasCommit" class="about-title__commit"> · {{ commit }}</span>
          </span>
        </div>
        <p class="about-title__tagline">{{ t('about.tagline') }}</p>
      </header>

      <section class="about-card">
        <h2 class="about-card__heading">{{ t('about.copyrightHeading') }}</h2>
        <p class="about-card__text">{{ copyright }}</p>
        <p class="about-card__text about-card__text--muted">
          {{ t('about.licenseLine') }}
          <a
            :href="`${repoURL}/blob/main/LICENSE`"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link"
          >
            {{ license }}
          </a>
        </p>
        <p class="about-card__text about-card__text--muted">
          {{ t('about.agplNotice') }}
        </p>
      </section>

      <section class="about-card">
        <h2 class="about-card__heading">{{ t('about.authorHeading') }}</h2>
        <p class="about-card__text">{{ author }}</p>
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
            :href="sourceURL"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link about-link--row"
          >
            <Github class="about-link__icon" />
            <span>{{ sourceLinkLabel }}</span>
          </a>
          <a
            v-if="version && version !== '--'"
            :href="`${repoURL}/releases/tag/v${version}`"
            target="_blank"
            rel="noopener noreferrer"
            class="about-link about-link--row"
          >
            <Info class="about-link__icon" />
            <span>{{ t('about.viewRelease', { version }) }}</span>
          </a>
          <p v-if="buildTime" class="about-card__text about-card__text--muted about-build-time">
            {{ t('about.buildTime', { time: buildTime }) }}
          </p>
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

// Author profile URL is intentionally hardcoded — it points to the person's
// GitHub user page, not to the repository, so it is not part of /hello.
const AUTHOR_GITHUB = 'https://github.com/lin-snow'

const FALLBACK_REPO = 'https://github.com/lin-snow/Ech0'
const FALLBACK_AUTHOR = 'lin-snow'
const FALLBACK_LICENSE = 'AGPL-3.0-or-later'

const version = computed(() => settingStore.hello?.version || '--')
const commit = computed(() => settingStore.hello?.commit || '')
const hasCommit = computed(() => commit.value !== '' && commit.value !== 'unknown')
const buildTime = computed(() => settingStore.hello?.build_time || '')
const author = computed(() => settingStore.hello?.author || FALLBACK_AUTHOR)
const license = computed(() => settingStore.hello?.license || FALLBACK_LICENSE)
const repoURL = computed(() => settingStore.hello?.repo_url || FALLBACK_REPO)
const copyright = computed(
  () => settingStore.hello?.copyright || `Copyright (C) ${new Date().getFullYear()} ${author.value}`,
)

// AGPL-3.0 §13 anchor: when we know the exact commit, link the user to /tree/<commit>
// so the source they receive matches the running binary. Falls back to repo root.
const sourceURL = computed(() =>
  hasCommit.value ? `${repoURL.value}/tree/${commit.value}` : repoURL.value,
)
const sourceLinkLabel = computed(() =>
  hasCommit.value
    ? t('about.viewSourceAtCommit', { commit: commit.value })
    : t('about.viewSource'),
)
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

.about-title__commit {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-weight: 500;
  color: var(--color-text-muted);
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

.about-build-time {
  margin-top: 0.25rem;
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.about-footer {
  margin-top: 0.5rem;
  text-align: center;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}
</style>
