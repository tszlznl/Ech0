<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  视频内联展示：一条 Echo 至多一个视频。卡片里渲染一张「首帧封面」，左上角有半透明磨砂标签
  标明这是视频、右下角是时长角标——不再在正中盖一个挡画面的播放按钮。
  桌面端 hover 时内联静音预览播放（prefers-reduced-motion 下不自动播放）；点击后在暗色灯箱里
  放大播放（带声音），与图片画廊「缩略图 → 灯箱」的体感一致。移动端无 hover，点按直接进灯箱。
  设计上是单视频，但对历史/异常的多视频数据也稳妥兼容：共用一个灯箱、由 activeIndex 驱动。
-->
<template>
  <div v-if="items.length" class="video-media">
    <button
      v-for="(item, idx) in items"
      :key="item.id || idx"
      type="button"
      class="video-media__poster"
      :style="aspectStyle(item, idx)"
      :aria-label="t('videoPlayer.play')"
      @click="open(idx, $event)"
      @mouseenter="onEnter"
      @mouseleave="onLeave"
    >
      <video
        class="video-media__frame"
        :src="posterSrc(item.src)"
        preload="metadata"
        muted
        playsinline
        tabindex="-1"
        @loadedmetadata="onMeta(idx, $event)"
      ></video>
      <span class="video-media__tag">
        <Video color="currentColor" />
        {{ t('videoPlayer.label') }}
      </span>
      <span v-if="durationLabel(idx)" class="video-media__badge">{{ durationLabel(idx) }}</span>
    </button>

    <TheVideoLightbox :visible="activeIndex !== null" :src="activeSrc" @close="close" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFileUrl, getHubFileUrl } from '@/utils/other'
import { formatMediaTime } from '../shared/time'
import Video from '@/components/icons/video.vue'
import TheVideoLightbox from './TheVideoLightbox.vue'

type VideoMeta = { duration: number; width: number; height: number }

const props = withDefaults(
  defineProps<{
    files?: App.Api.Ech0.FileObject[]
    /** 联邦(hub)卡片场景下用于把相对 URL 解析到远端服务器。 */
    baseUrl?: string
  }>(),
  { files: () => [] },
)

const { t } = useI18n()

const items = computed(() =>
  (props.files || []).map((file) => ({
    id: file.id,
    src: props.baseUrl ? getHubFileUrl(file, props.baseUrl) : getFileUrl(file),
    width: file.width,
    height: file.height,
  })),
)

// 从视频元数据兜底出来的时长 / 尺寸（当 FileObject 未带 width/height 时用于确定海报比例）。
const metaByIndex = ref<Record<number, VideoMeta>>({})

const activeIndex = ref<number | null>(null)
const activeSrc = computed(() =>
  activeIndex.value !== null ? (items.value[activeIndex.value]?.src ?? '') : '',
)

// 用媒体片段 #t=0.1 提示浏览器定位到首帧并渲染出来，作为静态封面。
function posterSrc(src: string) {
  return src ? `${src}#t=0.1` : ''
}

function onMeta(idx: number, event: Event) {
  const el = event.target as HTMLVideoElement
  metaByIndex.value = {
    ...metaByIndex.value,
    [idx]: {
      duration: Number.isFinite(el.duration) ? el.duration : 0,
      width: el.videoWidth,
      height: el.videoHeight,
    },
  }
}

function prefersReducedMotion() {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
}

// hover 停留达到该时长才开始预览播放，避免鼠标只是划过就触发拉流、浪费带宽。
const HOVER_PLAY_DELAY = 300
let hoverTimer: ReturnType<typeof setTimeout> | null = null

function clearHoverTimer() {
  if (hoverTimer !== null) {
    clearTimeout(hoverTimer)
    hoverTimer = null
  }
}

// 桌面端 hover：停留满 300ms 才内联静音预览播放；离开（或未满时长即移开）则取消/暂停并回到首帧。
function onEnter(event: MouseEvent) {
  if (prefersReducedMotion()) return
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video) return
  clearHoverTimer()
  hoverTimer = setTimeout(() => {
    hoverTimer = null
    video.muted = true
    void video.play().catch(() => {})
  }, HOVER_PLAY_DELAY)
}

function onLeave(event: MouseEvent) {
  clearHoverTimer()
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video) return
  video.pause()
  video.currentTime = 0
}

onBeforeUnmount(clearHoverTimer)

function aspectStyle(item: { width?: number; height?: number }, idx: number) {
  const meta = metaByIndex.value[idx]
  const width = item.width || meta?.width
  const height = item.height || meta?.height
  return { aspectRatio: width && height ? `${width} / ${height}` : '16 / 9' }
}

function durationLabel(idx: number) {
  const duration = metaByIndex.value[idx]?.duration
  return duration ? formatMediaTime(duration) : ''
}

function open(idx: number, event?: MouseEvent) {
  // 进灯箱前先停掉封面的 hover 静音预览：指针停在原处不会触发 mouseleave，
  // 否则封面视频会在灯箱底下继续解码/拉流，形成两个视频同时播放。
  clearHoverTimer()
  const poster = (event?.currentTarget as HTMLElement | undefined)?.querySelector('video')
  if (poster) {
    poster.pause()
    poster.currentTime = 0
  }
  activeIndex.value = idx
}

function close() {
  activeIndex.value = null
}
</script>

<style scoped>
.video-media {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.video-media__poster {
  position: relative;
  display: block;
  width: 100%;
  max-height: 70vh;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: #000;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.video-media__poster:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-soft);
}

.video-media__poster:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px var(--color-focus-ring);
}

.video-media__frame {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 左上角半透明磨砂标签：标明这是视频，不遮挡画面主体 */
.video-media__tag {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.22rem;
  padding: 0.14rem 0.45rem;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 42%);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(6px);
  pointer-events: none;
}

.video-media__tag svg {
  width: 0.9rem;
  height: 0.9rem;
}

.video-media__badge {
  position: absolute;
  right: 0.5rem;
  bottom: 0.5rem;
  padding: 0.1rem 0.4rem;
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 60%);
  border-radius: var(--radius-sm);

  /* 时长仅在 hover / 键盘 focus 时淡入，平时保持画面干净 */
  opacity: 0;
  transition: opacity 0.2s ease;
}

.video-media__poster:hover .video-media__badge,
.video-media__poster:focus-visible .video-media__badge {
  opacity: 1;
}

@media (prefers-reduced-motion: reduce) {
  .video-media__poster,
  .video-media__badge {
    transition: none;
  }
}
</style>
