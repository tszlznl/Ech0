<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  视频灯箱：与图片画廊(PhotoSwipe)一致的「缩略图 → 暗色浮层」体感，但为视频量身做的轻量实现——
  暗背景、居中大屏、原生控件（音量/进度/全屏/画中画由浏览器提供）+ 打开即自动播放，并从内联封面的进度接续播放。
  ESC 或点击背景关闭；打开期间锁定 body 滚动。用 v-if 卸载 <video>，关闭时音频自然停止。
-->
<template>
  <Teleport to="body">
    <Transition name="video-lightbox">
      <div
        v-if="visible"
        class="video-lightbox"
        role="dialog"
        aria-modal="true"
        @click.self="emitClose"
      >
        <button
          type="button"
          class="video-lightbox__close"
          :aria-label="t('videoPlayer.close')"
          @click="emitClose"
        >
          <Close color="currentColor" />
        </button>

        <div class="video-lightbox__stage">
          <video
            class="video-lightbox__video"
            :src="src"
            controls
            autoplay
            playsinline
            @loadedmetadata="onLoadedMetadata"
          >
            {{ t('videoPlayer.unsupported') }}
          </video>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Close from '@/components/icons/close.vue'

const props = withDefaults(
  defineProps<{
    visible: boolean
    src: string
    /** 从内联封面带过来的续播进度（秒）；0 表示从头播。 */
    startTime?: number
  }>(),
  { startTime: 0 },
)

// 元数据就绪即定位到内联封面走到的进度，实现「点全屏接续播放」而非从头开始。
function onLoadedMetadata(event: Event) {
  const el = event.target as HTMLVideoElement
  if (props.startTime > 0 && Number.isFinite(el.duration)) {
    el.currentTime = Math.min(props.startTime, el.duration)
  }
}

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()

let previousBodyOverflow = ''
let scrollLocked = false

function emitClose() {
  emit('close')
}

function lockScroll() {
  if (scrollLocked) return
  previousBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  scrollLocked = true
}

function unlockScroll() {
  if (!scrollLocked) return
  document.body.style.overflow = previousBodyOverflow
  scrollLocked = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') emitClose()
}

watch(
  () => props.visible,
  (open) => {
    if (open) {
      lockScroll()
      window.addEventListener('keydown', onKeydown)
    } else {
      unlockScroll()
      window.removeEventListener('keydown', onKeydown)
    }
  },
)

onBeforeUnmount(() => {
  unlockScroll()
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.video-lightbox {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(1rem, 4vw, 3rem);
  background: rgb(0 0 0 / 88%);
  backdrop-filter: blur(2px);
}

.video-lightbox__close {
  position: absolute;
  top: clamp(0.75rem, 2vw, 1.25rem);
  right: clamp(0.75rem, 2vw, 1.25rem);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.4rem;
  height: 2.4rem;
  padding: 0;
  font-size: 1.35rem;
  color: #fff;
  background: rgb(255 255 255 / 12%);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  transition:
    background 0.15s ease,
    transform 0.15s ease;
}

.video-lightbox__close:hover {
  background: rgb(255 255 255 / 22%);
}

.video-lightbox__close:active {
  transform: scale(0.92);
}

.video-lightbox__close:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgb(255 255 255 / 55%);
}

.video-lightbox__stage {
  display: flex;
  max-width: min(92vw, 1280px);
  max-height: 86vh;
  transition: transform 0.22s ease;
}

.video-lightbox__video {
  max-width: 100%;
  max-height: 86vh;
  border-radius: var(--radius-md);
  background: #000;
  box-shadow: 0 12px 48px rgb(0 0 0 / 45%);
}

/* 进出：背景淡入 + 舞台轻微放大，呼应图片灯箱的 zoom 体感 */
.video-lightbox-enter-active,
.video-lightbox-leave-active {
  transition: opacity 0.2s ease;
}

.video-lightbox-enter-from,
.video-lightbox-leave-to {
  opacity: 0;
}

.video-lightbox-enter-from .video-lightbox__stage,
.video-lightbox-leave-to .video-lightbox__stage {
  transform: scale(0.94);
}

@media (prefers-reduced-motion: reduce) {
  .video-lightbox__stage,
  .video-lightbox-enter-active,
  .video-lightbox-leave-active {
    transition: none;
  }
}
</style>
