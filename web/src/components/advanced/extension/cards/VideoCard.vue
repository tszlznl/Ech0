<template>
  <ExtensionCardShell>
    <div class="video-shell">
      <span class="video-provider-badge">{{ providerLabel }}</span>
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
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'

const props = defineProps<{
  videoId: string
}>()

const isBilibili = computed(() => props.videoId.startsWith('BV'))
const providerLabel = computed(() => (isBilibili.value ? 'Bilibili' : 'YouTube'))
</script>

<style scoped>
.video-shell {
  position: relative;
  padding: 0.55rem;
}

.video-provider-badge {
  position: absolute;
  top: 0.9rem;
  right: 0.95rem;
  z-index: 1;
  padding: 0.15rem 0.45rem;
  border-radius: 9999px;
  font-size: 0.68rem;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-muted) 14%);
  backdrop-filter: blur(1.2px);
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
</style>
