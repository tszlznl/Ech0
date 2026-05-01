<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ExtensionCardShell :header-label="t('extensionCard.website')">
    <template #header-icon><Link /></template>
    <a
      :href="websiteInfo.site"
      target="_blank"
      rel="noopener noreferrer"
      class="website-card__link website-card__body"
    >
      <div class="website-icon-wrap">
        <Link class="w-4 h-4" />
      </div>
      <div class="website-meta">
        <span class="website-title">{{ websiteInfo.title }}</span>
        <span class="website-domain">{{ displayDomain }}</span>
      </div>
    </a>
  </ExtensionCardShell>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Link from '@/components/icons/link.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'

const { t } = useI18n()

const props = defineProps<{
  website: { title: string; site: string }
}>()
const websiteInfo = props.website

const displayDomain = computed(() => {
  const site = websiteInfo.site.trim()
  if (!site) return ''
  try {
    const parsed = new URL(site)
    return parsed.hostname.replace(/^www\./, '')
  } catch {
    return site.replace(/^https?:\/\//, '').replace(/\/$/, '')
  }
})
</script>

<style scoped>
.website-card__link {
  display: block;
  border-radius: inherit;
}

.website-card__link:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 1px var(--color-focus-ring),
    0 0 0 4px var(--card-focus-ring-outer);
}

.website-card__body {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
}

.website-icon-wrap {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
}

.website-meta {
  min-width: 0;
  flex: 1;
}

.website-title {
  display: block;
  color: var(--color-text-primary);
  font-size: 0.96rem;
  font-weight: 700;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.website-domain {
  margin-top: 0.15rem;
  display: block;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  font-family: var(--font-family-mono);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
