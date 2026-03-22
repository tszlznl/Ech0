<template>
  <div class="w-[88%] mx-auto mb-4">
    <div class="horizontal-scroll-container">
      <div class="horizontal-scroll-wrapper">
        <GalleryImageItem
          v-for="(image, idx) in images"
          :key="getImageKey(image, idx)"
          :image="image"
          :src="resolvedSrcs[idx] || ''"
          :alt="getAlt(idx)"
          :loaded="isLoaded(image, idx)"
          button-class="horizontal-item rounded-lg shrink-0"
          frame-class="h-full"
          img-class="h-full w-auto object-contain"
          :frame-style="getHorizontalAspectStyle(image)"
          @click="open(idx, $event)"
          @load="markLoaded(image, idx)"
          @error="markLoaded(image, idx)"
        />
      </div>
    </div>
    <div class="scroll-hint">{{ scrollHintText }}</div>
  </div>
</template>

<script setup lang="ts">
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryHorizontalProps } from './types'

defineProps<GalleryHorizontalProps>()
</script>

<style scoped>
.horizontal-scroll-container {
  position: relative;
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.1) transparent;
}

.horizontal-scroll-container::-webkit-scrollbar {
  height: 4px;
}

.horizontal-scroll-wrapper {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  align-items: center;
}

.horizontal-item {
  height: 200px;
  width: auto;
  overflow: hidden;
}

.scroll-hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: 8px;
  animation: hint-pulse 2s infinite;
}

@keyframes hint-pulse {
  0%,
  100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}
</style>
