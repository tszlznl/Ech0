<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="about-page">
    <div class="about-back-wrap">
      <BaseButton @click="$router.push({ name: 'home' })" :tooltip="t('commonNav.backHome')">
        <Arrow class="text-2xl rotate-180" />
      </BaseButton>
    </div>

    <article class="about-shell">
      <header class="about-hero">
        <h1 class="about-hero__name">Ech0</h1>
        <div class="about-hero__meta">
          <span class="about-hero__version">v{{ version }}</span>
          <span v-if="hasCommit" class="about-hero__sep" aria-hidden="true">·</span>
          <span v-if="hasCommit" class="about-hero__commit">{{ commit }}</span>
        </div>
      </header>

      <hr class="about-rule" />

      <section class="about-section">
        <h2 class="about-section__heading">{{ t('about.authorHeading') }}</h2>
        <p class="about-section__line">
          <span class="about-section__value">{{ author }}</span>
          <a :href="AUTHOR_GITHUB" target="_blank" rel="noopener noreferrer" class="about-chip">
            <Github class="about-chip__icon" />
            <span>{{ t('about.authorGithub') }}</span>
          </a>
        </p>
      </section>

      <hr class="about-rule" />

      <section class="about-section">
        <h2 class="about-section__heading">{{ t('about.sourceHeading') }}</h2>
        <ul class="about-list">
          <li>
            <a
              :href="sourceURL"
              target="_blank"
              rel="noopener noreferrer"
              class="about-chip"
              :aria-label="sourceLinkLabel"
              :title="sourceLinkLabel"
            >
              <Github class="about-chip__icon" />
              <span v-if="hasCommit" class="about-chip__commit">{{ commit }}</span>
              <span v-else>{{ t('about.viewSource') }}</span>
            </a>
          </li>
          <li v-if="version && version !== '--'">
            <a
              :href="`${repoURL}/releases/tag/v${version}`"
              target="_blank"
              rel="noopener noreferrer"
              class="about-chip"
            >
              <Info class="about-chip__icon" />
              <span>{{ t('about.viewRelease', { version }) }}</span>
            </a>
          </li>
        </ul>
        <p v-if="buildTime" class="about-section__build">
          {{ t('about.buildTime', { time: buildTime }) }}
        </p>
      </section>

      <hr class="about-rule" />

      <section class="about-section">
        <h2 class="about-section__heading">{{ t('about.copyrightHeading') }}</h2>
        <p class="about-section__line">{{ copyright }}</p>
        <p class="about-section__line about-section__line--muted">
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
      </section>

      <footer class="about-footer">
        <span class="about-footer__mark" aria-hidden="true">·</span>
        <span>{{ t('about.poweredBy') }}</span>
        <span class="about-footer__mark" aria-hidden="true">·</span>
      </footer>
    </article>
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
const FALLBACK_AUTHOR = 'L1nSn0w'
const FALLBACK_LICENSE = 'AGPL-3.0-or-later'

const version = computed(() => settingStore.hello?.version || '--')
const commit = computed(() => settingStore.hello?.commit || '')
const hasCommit = computed(() => commit.value !== '' && commit.value !== 'unknown')
const buildTime = computed(() => settingStore.hello?.build_time || '')
const author = computed(() => settingStore.hello?.author || FALLBACK_AUTHOR)
const license = computed(() => settingStore.hello?.license || FALLBACK_LICENSE)
const repoURL = computed(() => settingStore.hello?.repo_url || FALLBACK_REPO)
const copyright = computed(
  () =>
    settingStore.hello?.copyright || `Copyright (C) ${new Date().getFullYear()} ${author.value}`,
)

// AGPL-3.0 §13 anchor: when we know the exact commit, link the user to /tree/<commit>
// so the source they receive matches the running binary. Falls back to repo root.
const sourceURL = computed(() =>
  hasCommit.value ? `${repoURL.value}/tree/${commit.value}` : repoURL.value,
)
const sourceLinkLabel = computed(() =>
  hasCommit.value ? t('about.viewSourceAtCommit', { commit: commit.value }) : t('about.viewSource'),
)
</script>

<style scoped>
.about-page {
  position: relative;
  display: flex;
  justify-content: center;
  width: 100%;
  min-height: 100vh;
  padding: 2.5rem 1.25rem 4rem;
}

.about-back-wrap {
  position: absolute;
  top: 1rem;
  left: 1rem;
  z-index: 1;
}

.about-shell {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: 100%;
  max-width: 36rem;
  padding: 1rem 0.25rem;
}

.about-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.55rem;
  padding: 1.5rem 0 0.75rem;
  text-align: center;
}

.about-hero__name {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 2.5rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
}

.about-hero__meta {
  display: inline-flex;
  align-items: baseline;
  gap: 0.4rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.about-hero__version {
  font-weight: 600;
  letter-spacing: 0.02em;
}

.about-hero__sep {
  opacity: 0.6;
}

.about-hero__commit {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.75rem;
  letter-spacing: 0.02em;
}

.about-hero__tagline {
  margin: 0.25rem 0 0;
  font-family: 'Songti SC', STSong, var(--font-family-display);
  font-size: 0.9375rem;
  font-weight: 500;
  letter-spacing: 0.02em;
  color: var(--color-text-secondary);
}

.about-rule {
  width: 100%;
  margin: 0;
  border: 0;
  border-top: 1px solid var(--color-text-muted);
  opacity: 0.18;
}

.about-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0 0.25rem;
}

.about-section__heading {
  margin: 0 0 0.15rem;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.about-section__line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  margin: 0;
  font-size: 0.9375rem;
  line-height: 1.6;
  color: var(--color-text-primary);
}

.about-section__line--muted {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.about-section__value {
  font-weight: 600;
  color: var(--color-text-primary);
}

.about-section__build {
  margin: 0.15rem 0 0;
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.about-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin: 0.15rem 0 0;
  padding: 0;
  list-style: none;
}

.about-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.6rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  text-decoration: none;
  background: transparent;
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-text-muted) 30%, transparent);
  transition:
    color 0.15s ease,
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.about-chip:hover {
  color: var(--color-accent);
  background: var(--color-accent-soft);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-accent) 35%, transparent);
}

.about-chip__icon {
  font-size: 0.95rem;
  color: currentColor;
}

.about-chip:hover .about-chip__icon {
  color: currentColor;
}

.about-chip__commit {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.75rem;
  letter-spacing: 0.02em;
}

.about-link {
  color: var(--color-text-primary);
  text-decoration: underline;
  text-decoration-color: color-mix(in oklab, var(--color-text-muted) 60%, transparent);
  text-underline-offset: 0.22em;
  transition: text-decoration-color 0.15s ease;
}

.about-link:hover {
  text-decoration-color: var(--color-accent);
  color: var(--color-accent);
}

.about-footer {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  margin-top: 1rem;
  font-family: 'Songti SC', STSong, var(--font-family-display);
  font-size: 0.75rem;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
  text-align: center;
}

.about-footer__mark {
  opacity: 0.5;
}

@media (width <= 480px) {
  .about-page {
    padding: 2rem 0.75rem 3rem;
  }

  .about-hero__name {
    font-size: 2.125rem;
  }
}
</style>
