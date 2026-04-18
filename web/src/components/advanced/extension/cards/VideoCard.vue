<template>
  <ExtensionCardShell :header-label="providerLabel">
    <template #header-icon><Video /></template>
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
import Video from '@/components/icons/video.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'

const props = defineProps<{
  videoId: string
}>()

const isBilibili = computed(() => props.videoId.startsWith('BV'))
const providerLabel = computed(() => (isBilibili.value ? 'Bilibili' : 'YouTube'))
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
</style>
