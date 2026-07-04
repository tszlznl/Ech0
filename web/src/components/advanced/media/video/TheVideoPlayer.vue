<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  视频内联展示：一条 Echo 至多一个视频。卡片里渲染一张「首帧封面」，左上角有半透明磨砂标签
  标明这是视频、右上角是可点的「全屏」按钮、右下角是时长角标——不再在正中盖一个挡画面的播放按钮。
  首帧解码就绪前显示一层「暗色主题底 + Apple 菊花」的占位、就绪后封面淡入，避免「先黑屏后出内容」的生硬闪烁。
  桌面端 hover 时内联静音预览播放（prefers-reduced-motion 下不自动播放）；点击封面就地「有声」播放/暂停
  （点击是用户手势、浏览器才允许有声；hover 预览中点一下即取消静音继续）。移动端无 hover，点封面即就地有声播放。
  右上角全屏按钮（桌面 hover 淡入 / 触摸设备常显）才在暗色灯箱里放大播放——点封面本身不进灯箱。
  设计上是单视频，但对历史/异常的多视频数据也稳妥兼容：共用一个灯箱、由 activeIndex 驱动。
-->
<template>
  <div v-if="items.length" class="video-media">
    <div
      v-for="(item, idx) in items"
      :key="item.id || idx"
      class="video-media__poster"
      :style="aspectStyle(item, idx)"
      @mouseenter="onEnter"
      @mouseleave="onLeave"
    >
      <!-- 首帧未就绪的占位：暗色主题底 + Apple 菊花 loading（纯装饰，aria-hidden） -->
      <span
        class="video-media__skeleton"
        :class="{ 'is-loaded': isLoaded(idx) }"
        aria-hidden="true"
      >
        <span class="video-media__spinner">
          <i v-for="seg in spinnerSegments" :key="seg" :style="{ '--seg': seg - 1 }"></i>
        </span>
      </span>
      <!-- 点封面即就地「有声」播放/暂停；想大屏看点右上角全屏按钮 -->
      <video
        class="video-media__frame"
        :class="{ 'is-loaded': isLoaded(idx) }"
        :src="posterSrc(item.src)"
        :aria-label="t('videoPlayer.play')"
        preload="metadata"
        muted
        playsinline
        tabindex="-1"
        @loadedmetadata="onMeta(idx, $event)"
        @loadeddata="onLoaded(idx)"
        @click="onClickVideo($event)"
        @play="onPlay(idx)"
        @pause="onPause(idx)"
        @ended="onPause(idx)"
        @timeupdate="onTimeUpdate(idx, $event)"
      ></video>
      <span class="video-media__tag">
        <Video color="currentColor" />
        {{ t('videoPlayer.label') }}
      </span>
      <!-- 右上角全屏按钮：与左上 tag 对称的图标+文字矩形，是进灯箱大屏播放的唯一入口 -->
      <button
        type="button"
        class="video-media__fullscreen"
        :class="{ 'is-visible': playingByIndex[idx] }"
        @click.stop="open(idx, $event)"
      >
        <Full color="currentColor" />
        {{ t('videoPlayer.fullscreen') }}
      </button>
      <span
        v-if="badgeLabel(idx)"
        class="video-media__badge"
        :class="{ 'is-visible': playingByIndex[idx] }"
        >{{ badgeLabel(idx) }}</span
      >
    </div>

    <TheVideoLightbox :visible="activeIndex !== null" :src="activeSrc" @close="close" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFileUrl, getHubFileUrl } from '@/utils/other'
import { formatMediaTime } from '../shared/time'
import Video from '@/components/icons/video.vue'
import Full from '@/components/icons/full.vue'
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

// Apple 菊花 loading 的 12 根辐射条（占位用，纯装饰）
const spinnerSegments = Array.from({ length: 12 }, (_, index) => index + 1)

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

// 首帧是否已解码可见：loadeddata（HAVE_CURRENT_DATA）才代表 #t=0.1 那帧画出来了，
// 据此把封面从 skeleton 占位淡入，避免 metadata 阶段的黑屏闪一下。
const loadedByIndex = ref<Record<number, boolean>>({})

function onLoaded(idx: number) {
  if (loadedByIndex.value[idx]) return
  loadedByIndex.value = { ...loadedByIndex.value, [idx]: true }
}

function isLoaded(idx: number) {
  return Boolean(loadedByIndex.value[idx])
}

// 播放态与当前进度：右下角时长在播放时切成剩余时间倒计时并常显，静止时仍显示总时长。
const playingByIndex = ref<Record<number, boolean>>({})
const currentTimeByIndex = ref<Record<number, number>>({})

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

// 桌面端 hover：停留满 300ms 才内联静音预览播放；已在播放（预览或用户主动）则不重复触发。
function onEnter(event: MouseEvent) {
  if (prefersReducedMotion()) return
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video || !video.paused) return
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
  // 用户点击后的「有声」播放不因鼠标移开而中断；只回收 hover 静音预览、复位到首帧。
  if (!video.paused && !video.muted) return
  video.pause()
  video.currentTime = 0
  video.muted = true
}

// 点封面只「播放」不暂停：暂停中→有声播放；hover 静音预览中→取消静音继续；已有声播放→不打断。
function onClickVideo(event: MouseEvent) {
  clearHoverTimer()
  const video = event.currentTarget as HTMLVideoElement
  if (video.paused) {
    video.muted = false
    void video.play().catch(() => {})
  } else if (video.muted) {
    video.muted = false
  }
}

function onPlay(idx: number) {
  playingByIndex.value[idx] = true
}

// pause 与 ended 都回到「非播放」，角标随之切回总时长。
function onPause(idx: number) {
  playingByIndex.value[idx] = false
}

function onTimeUpdate(idx: number, event: Event) {
  currentTimeByIndex.value[idx] = (event.target as HTMLVideoElement).currentTime
}

onBeforeUnmount(clearHoverTimer)

function aspectStyle(item: { width?: number; height?: number }, idx: number) {
  const meta = metaByIndex.value[idx]
  const width = item.width || meta?.width
  const height = item.height || meta?.height
  return { aspectRatio: width && height ? `${width} / ${height}` : '16 / 9' }
}

// 右下角角标：静止时显示总时长；播放时切成剩余时间倒计时（带「-」前缀）。
function badgeLabel(idx: number) {
  const duration = metaByIndex.value[idx]?.duration
  if (!duration) return ''
  if (playingByIndex.value[idx]) {
    const remaining = Math.max(0, duration - (currentTimeByIndex.value[idx] ?? 0))
    return `-${formatMediaTime(remaining)}`
  }
  return formatMediaTime(duration)
}

function open(idx: number, event?: MouseEvent) {
  // 进灯箱前先停掉封面播放：指针停在原处不会触发 mouseleave，否则封面视频会在灯箱底下
  // 继续解码/拉流，形成两个视频同时播放（甚至双声道）。全屏按钮是触发者，向上找同一 poster 内的封面。
  clearHoverTimer()
  const trigger = event?.currentTarget as HTMLElement | undefined
  const video = trigger?.closest('.video-media__poster')?.querySelector('video')
  if (video) {
    video.pause()
    video.currentTime = 0
    video.muted = true
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
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

/* focus-within：全屏按钮获得键盘焦点时也高亮卡片边框 */
.video-media__poster:hover,
.video-media__poster:focus-within {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-soft);
}

.video-media__frame {
  /* 提到 skeleton 之上：首帧未就绪时透明露出 skeleton，就绪后淡入盖住它，全程不露黑底 */
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.video-media__frame.is-loaded {
  opacity: 1;
}

/* 首帧解码前的占位：跟随主题的暗色底（accent-soft → 表面色，再叠一层黑压暗）+ 居中的 Apple 菊花 */
.video-media__skeleton {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  background:
    linear-gradient(rgb(0 0 0 / 66%), rgb(0 0 0 / 76%)),
    radial-gradient(
      130% 130% at 50% 25%,
      var(--color-accent-soft) 0%,
      var(--color-bg-muted) 55%,
      var(--color-bg-canvas) 100%
    );
}

/* 首帧就绪后被封面盖住，直接卸掉占位、连带停掉菊花动画省开销 */
.video-media__skeleton.is-loaded {
  display: none;
}

/* Apple 菊花：12 根辐射条绕心旋转、逐条淡出（做法借鉴 common/TheLoadingIndicator） */
.video-media__spinner {
  position: relative;
  width: 26px;
  height: 26px;
  color: rgb(255 255 255 / 82%);
}

.video-media__spinner i {
  position: absolute;
  top: 0;
  left: 50%;
  width: 2.4px;
  height: 6.4px;
  border-radius: 999px;
  background: currentColor;
  transform-origin: center 13px;
  transform: translateX(-50%) rotate(calc(var(--seg) * 30deg));
  animation: video-spinner-fade 1.1s linear infinite;
  animation-delay: calc(var(--seg) * -0.0916s);
}

@keyframes video-spinner-fade {
  0% {
    opacity: 1;
  }

  100% {
    opacity: 0.12;
  }
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

/* 右上角全屏按钮：与左上 tag 完全对称的磨砂矩形（图标 + 文字），进灯箱大屏播放的唯一入口 */
.video-media__fullscreen {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.22rem;
  padding: 0.14rem 0.45rem;
  font-family: inherit;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 42%);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(6px);
  cursor: pointer;

  /* 桌面平时隐藏、hover/focus 时淡入，保持画面干净 */
  opacity: 0;
  transition:
    opacity 0.2s ease,
    background 0.15s ease,
    transform 0.15s ease;
}

.video-media__fullscreen:hover {
  background: rgb(0 0 0 / 58%);
}

.video-media__fullscreen:active {
  transform: scale(0.96);
}

.video-media__fullscreen:focus-visible {
  outline: none;
  opacity: 1;
  box-shadow: 0 0 0 3px var(--color-focus-ring);
}

.video-media__fullscreen svg {
  width: 0.9rem;
  height: 0.9rem;
}

/* 桌面 hover/focus 时淡入；播放中（含移动端点击封面后）常显 —— 与右下角倒计时同步 */
.video-media__poster:hover .video-media__fullscreen,
.video-media__poster:focus-within .video-media__fullscreen,
.video-media__fullscreen.is-visible {
  opacity: 1;
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

/* 播放中的剩余时间倒计时常显（无需 hover），其余情况仅 hover/focus 淡入 */
.video-media__poster:hover .video-media__badge,
.video-media__poster:focus-within .video-media__badge,
.video-media__badge.is-visible {
  opacity: 1;
}

@media (prefers-reduced-motion: reduce) {
  .video-media__poster,
  .video-media__frame,
  .video-media__fullscreen,
  .video-media__badge {
    transition: none;
  }

  .video-media__spinner i {
    animation: none;
    opacity: 0.5;
  }
}
</style>
