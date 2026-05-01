<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ExtensionCardShell :header-label="providerLabel">
    <template #header-icon><Video /></template>
    <template #header-actions>
      <a
        :href="videoUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="video-card__jump"
        :aria-label="t('extensionCard.jump')"
      >
        <span class="video-card__jump-text">{{ t('extensionCard.jump') }}</span>
        <Link class="video-card__jump-icon" />
      </a>
    </template>
    <div class="video-shell">
      <div class="video-frame-wrap">
        <iframe
          v-if="isBilibili"
          :src="`https://www.bilibili.com/blackboard/html5mobileplayer.html?bvid=${props.videoId}&as_wide=1&high_quality=1&danmaku=0`"
          scrolling="no"
          border="0"
          frameborder="no"
          framespacing="0"
          allowfullscreen="true"
          loading="lazy"
          class="video-frame"
        ></iframe>
        <iframe
          v-else
          :src="`https://www.youtube.com/embed/${props.videoId}`"
          frameborder="0"
          allow="
            accelerometer;
            clipboard-write;
            encrypted-media;
            gyroscope;
            picture-in-picture;
            web-share;
          "
          allowfullscreen
          loading="lazy"
          class="video-frame"
        ></iframe>
      </div>
    </div>
  </ExtensionCardShell>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Video from '@/components/icons/video.vue'
import Link from '@/components/icons/link.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'

const { t } = useI18n()

const props = defineProps<{
  videoId: string
}>()

const isBilibili = computed(() => props.videoId.startsWith('BV'))
const providerLabel = computed(() => (isBilibili.value ? 'Bilibili' : 'YouTube'))
const videoUrl = computed(() =>
  isBilibili.value
    ? `https://www.bilibili.com/video/${props.videoId}`
    : `https://www.youtube.com/watch?v=${props.videoId}`,
)
</script>

<style scoped>
.video-shell {
  padding: 0.55rem;
}

.video-frame-wrap {
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--color-bg-muted);
}

.video-frame {
  width: 100%;
  aspect-ratio: 16 / 9;
  border: 0;
  display: block;
}

.video-card__jump {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  text-decoration: none;
  border-radius: var(--radius-sm);
  padding: 0.1rem 0.3rem;
  transition:
    color 0.15s ease,
    background 0.15s ease;
}

.video-card__jump:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
}

.video-card__jump:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-focus-ring);
}

.video-card__jump-text {
  font-weight: 600;
  letter-spacing: 0.01em;
}

.video-card__jump-icon {
  width: 0.85rem;
  height: 0.85rem;
}
</style>
