<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  视频内联展示：一条 Echo 至多一个视频。卡片里渲染一张「首帧封面」。左上角是一块「信息标签」——
  静止时标明这是视频、播放时切成「视频图标 + 剩余时间倒计时」、用户主动暂停后常显「⏸ 已暂停」；
  右上角是可点的「全屏」按钮——不再在正中盖一个挡画面的播放按钮，也不再单设右下角时长角标（时间已并入左上标签）。
  首帧解码就绪前显示一层「暗色主题底 + Apple 菊花」的占位、就绪后封面淡入，避免「先黑屏后出内容」的生硬闪烁。
  桌面端 hover 时内联静音预览播放（prefers-reduced-motion 下不自动播放）；点击封面切换有声播放/暂停
  （首次点击是用户手势、浏览器才允许有声：暂停中→有声播放、hover 静音预览中→取消静音继续、有声播放中→暂停）。
  播放中控件（左右上角两块 tag）走 auto-hide：指针静止满 2 秒即一并淡出、保持画面干净，指针移动/进入或键盘 focus 再淡回；
  用户主动暂停时左上标签常显「已暂停」、全屏按钮也常显（无 hover 的触摸设备点暂停后仍够得到全屏入口）。
  右上角全屏按钮才在暗色灯箱里放大播放——点封面本身不进灯箱。
  设计上是单视频，但对历史/异常的多视频数据也稳妥兼容：共用一个灯箱、由 activeIndex 驱动。
-->
<template>
  <div v-if="items.length" class="video-media">
    <div
      v-for="(item, idx) in items"
      :key="item.id || idx"
      class="video-media__poster"
      :style="aspectStyle(item, idx)"
      @mouseenter="onEnter(idx, $event)"
      @mousemove="onMove(idx, $event)"
      @mouseleave="onLeave(idx, $event)"
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
        @click="onClickVideo(idx, $event)"
        @touchstart.passive="onTouchStart(idx, $event)"
        @touchmove.passive="onTouchMove(idx, $event)"
        @touchend="onTouchEnd(idx, $event)"
        @touchcancel="onTouchEnd(idx, $event)"
        @contextmenu.prevent
        @play="onPlay(idx)"
        @pause="onPause(idx)"
        @ended="onPause(idx)"
        @timeupdate="onTimeUpdate(idx, $event)"
      ></video>
      <!-- 左上角信息标签：静止=「视频」，播放=「图标 + 剩余时间倒计时」，主动暂停=「⏸ 已暂停」常显；
           播放中静止满 2 秒随控件一起淡出（is-hidden），指针活动 / 键盘 focus 时淡回 -->
      <span class="video-media__tag" :class="{ 'is-hidden': tagHidden(idx) }">
        <Pause v-if="pausedByIndex[idx]" color="currentColor" />
        <Video v-else color="currentColor" />
        {{ tagText(idx) }}
      </span>
      <!-- 右上角全屏按钮：与左上 tag 对称的图标+文字矩形，是进灯箱大屏播放的唯一入口 -->
      <button
        type="button"
        class="video-media__fullscreen"
        :class="{ 'is-visible': controlsVisible(idx) }"
        @click.stop="open(idx, $event)"
      >
        <Full color="currentColor" />
        {{ t('videoPlayer.fullscreen') }}
      </button>
    </div>

    <TheVideoLightbox
      :visible="activeIndex !== null"
      :src="activeSrc"
      :start-time="resumeTime"
      @close="close"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFileUrl, getHubFileUrl } from '@/utils/other'
import { formatMediaTime } from '../shared/time'
import Video from '@/components/icons/video.vue'
import Pause from '@/components/icons/pause.vue'
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

// 播放态与当前进度：左上标签在播放时切成剩余时间倒计时。
const playingByIndex = ref<Record<number, boolean>>({})
const currentTimeByIndex = ref<Record<number, number>>({})
// 用户「主动暂停」（点击有声播放中的视频）标记：据此左上标签常显「已暂停」，
// 并让 hover 不擅自续播、鼠标移开不复位到首帧。区别于 hover 预览结束/离开的自动暂停。
const pausedByIndex = ref<Record<number, boolean>>({})

// 控件（左上标签 + 右上全屏按钮）当前是否「唤出可见」。播放中鼠标静止一段时间即自动隐藏、保持画面干净，
// 指针进入 / 移动即唤出并刷新计时、移出即收起，暂停时常显。取代早期纯 CSS :hover —— 点击播放后指针停在
// 视频上会被 :hover 一直钉住、永不淡出，等于「2 秒后消失」没生效。静止（未播放）时靠进入唤出、移出收起。
const controlsShownByIndex = ref<Record<number, boolean>>({})

// 播放中，最后一次指针活动后再经过该毫秒数仍无动作就自动隐藏控件（即用户要的「越过 2 秒后消失」）。
const CONTROLS_IDLE_HIDE = 2000
const hideTimerByIndex: Record<number, ReturnType<typeof setTimeout>> = {}

// mousemove 位移阈值：只有真移动超过该像素数才算「有意唤出」。否则手搭在视频上的细微抖动、或页面滚动
// 引发的 mousemove 会不停重置隐藏计时，导致「播放 2 秒后消失」永不触发。滚动时 clientX/Y 不变、位移为 0，一并被滤掉。
const MOVE_ACTIVATE_DISTANCE = 8
let lastPointer: { x: number; y: number } | null = null

const activeIndex = ref<number | null>(null)
const activeSrc = computed(() =>
  activeIndex.value !== null ? (items.value[activeIndex.value]?.src ?? '') : '',
)
// 进灯箱时把内联封面的播放进度带过去，让大屏「接续播放」而不是从头开始。
const resumeTime = ref(0)

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

function clearHideTimer(idx: number) {
  const timer = hideTimerByIndex[idx]
  if (timer) {
    clearTimeout(timer)
    delete hideTimerByIndex[idx]
  }
}

// 唤出控件并刷新自动隐藏计时：只有「正在播放」才排定 2 秒后收起（静止时靠移出收起、暂停时常显）。
function showControls(idx: number) {
  if (!controlsShownByIndex.value[idx]) controlsShownByIndex.value[idx] = true
  clearHideTimer(idx)
  if (playingByIndex.value[idx]) {
    hideTimerByIndex[idx] = setTimeout(() => {
      delete hideTimerByIndex[idx]
      controlsShownByIndex.value[idx] = false
    }, CONTROLS_IDLE_HIDE)
  }
}

function hideControls(idx: number) {
  clearHideTimer(idx)
  controlsShownByIndex.value[idx] = false
}

// 判定一次 mousemove 是否为「真移动」：与上次记录点的距离超过阈值才算，顺带更新记录点。
// 过滤掉手部抖动（小幅来回，距离始终 < 阈值）与页面滚动（clientX/Y 不变，距离为 0）。
function pointerMovedEnough(event: MouseEvent) {
  const point = { x: event.clientX, y: event.clientY }
  if (!lastPointer) {
    lastPointer = point
    return true
  }
  if (Math.hypot(point.x - lastPointer.x, point.y - lastPointer.y) < MOVE_ACTIVATE_DISTANCE) {
    return false
  }
  lastPointer = point
  return true
}

// 移动端长按 = 静音预览：补齐桌面 hover 预览在触摸设备的缺失。按住满 350ms 且未滑动即静音预览播放，
// 松手回到首帧封面；轻点（短按）仍交给随后的 click 走有声播放/暂停。
const LONG_PRESS_DELAY = 350
const LONG_PRESS_MOVE_TOLERANCE = 10
let longPressTimer: ReturnType<typeof setTimeout> | null = null
let longPressStart: { x: number; y: number } | null = null
// 这两张表不进模板、无需响应式：longPress 记录本次触摸是否已触发预览；suppressClick 拦掉长按后补发的 click。
const longPressActiveByIndex: Record<number, boolean> = {}
const suppressClickByIndex: Record<number, boolean> = {}
// 最近一次触摸时间戳：用于识别触摸设备补发的 mouseenter，避免它再触发一套桌面 hover 预览。
let lastTouchAt = 0

function clearLongPressTimer() {
  if (longPressTimer !== null) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
}

function onTouchStart(idx: number, event: TouchEvent) {
  lastTouchAt = Date.now()
  suppressClickByIndex[idx] = false
  const touch = event.touches[0]
  longPressStart = touch ? { x: touch.clientX, y: touch.clientY } : null
  const video = event.currentTarget as HTMLVideoElement
  clearLongPressTimer()
  longPressTimer = setTimeout(() => {
    longPressTimer = null
    if (!video.paused) return // 已在播放（如轻点后又长按）就不重复起预览
    video.muted = true
    longPressActiveByIndex[idx] = true
    showControls(idx)
    void video.play().catch(() => {})
  }, LONG_PRESS_DELAY)
}

function onTouchMove(idx: number, event: TouchEvent) {
  if (longPressTimer === null || !longPressStart) return
  const touch = event.touches[0]
  if (!touch) return
  // 手指移动超过阈值即判定为滚动、放弃长按预览，把手势让回页面滚动。
  const moved = Math.hypot(touch.clientX - longPressStart.x, touch.clientY - longPressStart.y)
  if (moved > LONG_PRESS_MOVE_TOLERANCE) clearLongPressTimer()
}

function onTouchEnd(idx: number, event: TouchEvent) {
  lastTouchAt = Date.now()
  clearLongPressTimer()
  longPressStart = null
  if (!longPressActiveByIndex[idx]) return // 只是轻点：不拦 click，交给它走有声播放/暂停
  longPressActiveByIndex[idx] = false
  suppressClickByIndex[idx] = true // 拦掉长按后浏览器补发的那次 click，避免误转有声播放
  const video = event.currentTarget as HTMLVideoElement
  video.pause()
  video.currentTime = 0
  video.muted = true
  hideControls(idx)
}

// 指针进入即唤出控件；桌面端再停留满 300ms 才内联静音预览播放（已在播放则不重复触发）。
// 用户主动暂停的视频不因 hover 擅自续播——hover 只负责把已淡出的 tag 唤回来。
function onEnter(idx: number, event: MouseEvent) {
  lastPointer = { x: event.clientX, y: event.clientY }
  showControls(idx)
  // 触摸设备在 touchend 后会补发一次 mouseenter：只唤出控件，静音预览交给长按处理，不再走桌面 hover 预览。
  if (Date.now() - lastTouchAt < 800) return
  if (prefersReducedMotion()) return
  if (pausedByIndex.value[idx]) return
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video || !video.paused) return
  clearHoverTimer()
  hoverTimer = setTimeout(() => {
    hoverTimer = null
    video.muted = true
    void video.play().catch(() => {})
  }, HOVER_PLAY_DELAY)
}

// 指针在视频上「真移动」才唤出控件并刷新隐藏计时——手一动控件就回来、停下满 2 秒又收起；
// 细微抖动 / 页面滚动被 pointerMovedEnough 滤掉，不会把「播放 2 秒后消失」续命。
function onMove(idx: number, event: MouseEvent) {
  if (pointerMovedEnough(event)) showControls(idx)
}

function onLeave(idx: number, event: MouseEvent) {
  lastPointer = null
  hideControls(idx)
  clearHoverTimer()
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video) return
  // 已进入「有声」态（正在播放或用户主动暂停，muted 均为 false）的视频不因鼠标移开而复位；
  // 只回收 hover 静音预览（muted=true）、退回首帧。
  if (!video.muted) return
  video.pause()
  video.currentTime = 0
  video.muted = true
}

// 点封面切换有声播放/暂停：暂停中 或 hover/长按的静音预览中 → 转有声播放；有声播放中 → 暂停（左上常显「已暂停」）。
// 关键：从静音预览转有声必须在同一个点击手势里「解除静音 + 重新 play」——只翻 muted=false 不重播，
// 浏览器对「以静音起播的视频」常不接通音频，表现为「点了在放却没声音」。
function onClickVideo(idx: number, event: MouseEvent) {
  // 长按预览后浏览器补发的那次 click：吞掉，避免把「松手回封面」误转成有声播放。
  if (suppressClickByIndex[idx]) {
    suppressClickByIndex[idx] = false
    return
  }
  clearHoverTimer()
  const video = event.currentTarget as HTMLVideoElement
  if (video.paused || video.muted) {
    pausedByIndex.value[idx] = false
    video.muted = false
    void video.play().catch(() => {})
  } else {
    video.pause()
    pausedByIndex.value[idx] = true
  }
  showControls(idx)
}

function onPlay(idx: number) {
  playingByIndex.value[idx] = true
  pausedByIndex.value[idx] = false
  // 起播即唤出控件并开始 2 秒隐藏计时——用户要的「播放越过 2 秒后消失」的起点。
  showControls(idx)
}

// pause 与 ended 都回到「非播放」，左上标签随之从倒计时切回「视频」（主动暂停另由 pausedByIndex 显「已暂停」）。
// 非播放态无需自动隐藏计时，清掉待触发的隐藏定时器。
function onPause(idx: number) {
  playingByIndex.value[idx] = false
  clearHideTimer(idx)
}

function onTimeUpdate(idx: number, event: Event) {
  currentTimeByIndex.value[idx] = (event.target as HTMLVideoElement).currentTime
}

onBeforeUnmount(() => {
  clearHoverTimer()
  clearLongPressTimer()
  Object.values(hideTimerByIndex).forEach((timer) => clearTimeout(timer))
})

function aspectStyle(item: { width?: number; height?: number }, idx: number) {
  const meta = metaByIndex.value[idx]
  const width = item.width || meta?.width
  const height = item.height || meta?.height
  return { aspectRatio: width && height ? `${width} / ${height}` : '16 / 9' }
}

// 左上标签文案：主动暂停时显「已暂停」；播放中显剩余时间倒计时（带「-」前缀）；其余（静止/预览未起播）显「视频」。
function tagText(idx: number) {
  if (pausedByIndex.value[idx]) return t('videoPlayer.paused')
  const duration = metaByIndex.value[idx]?.duration
  if (playingByIndex.value[idx] && duration) {
    const remaining = Math.max(0, duration - (currentTimeByIndex.value[idx] ?? 0))
    return `-${formatMediaTime(remaining)}`
  }
  return t('videoPlayer.label')
}

// 左上标签隐藏时机：播放中且控件已进入 auto-hide（静止满 2 秒收起）——静止 / 暂停时不隐藏（显「视频」/「已暂停」）。
function tagHidden(idx: number) {
  return Boolean(playingByIndex.value[idx]) && !controlsShownByIndex.value[idx]
}

// 右上全屏按钮显示时机：控件唤出时（hover / 指针活动 / 起播头 2 秒）或用户主动暂停时——
// 暂停常显让无 hover 的触摸设备点暂停后仍够得到全屏入口。
function controlsVisible(idx: number) {
  return Boolean(controlsShownByIndex.value[idx]) || Boolean(pausedByIndex.value[idx])
}

function open(idx: number, event?: MouseEvent) {
  // 进灯箱前先停掉封面播放：指针停在原处不会触发 mouseleave，否则封面视频会在灯箱底下
  // 继续解码/拉流，形成两个视频同时播放（甚至双声道）。全屏按钮是触发者，向上找同一 poster 内的封面。
  clearHoverTimer()
  const trigger = event?.currentTarget as HTMLElement | undefined
  const video = trigger?.closest('.video-media__poster')?.querySelector('video')
  // 先记下封面当前进度（含 hover/长按预览走过的时间），再复位封面——灯箱据此接续播放。
  resumeTime.value = video ? video.currentTime : 0
  if (video) {
    video.pause()
    video.currentTime = 0
    video.muted = true
  }
  pausedByIndex.value[idx] = false
  hideControls(idx)
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

  /* 抑制移动端长按视频弹出的「存储视频/拷贝」系统菜单与文字选择，长按手势留给静音预览 */
  -webkit-touch-callout: none;
  user-select: none;
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

/* 左上角半透明磨砂标签：静止标明这是视频 / 播放显剩余倒计时 / 主动暂停显「已暂停」，不遮挡画面主体 */
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
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 42%);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(6px);
  pointer-events: none;

  /* 播放中控件 auto-hide：静止满 2 秒淡出，指针活动 / focus 时淡回 —— 与右上角全屏按钮同步 */
  opacity: 1;
  transition: opacity 0.2s ease;
}

.video-media__tag.is-hidden {
  opacity: 0;
}

/* 键盘焦点进入卡片时把已隐藏的标签唤回，保证可达性（指针的唤出/收起走 JS auto-hide） */
.video-media__poster:focus-within .video-media__tag.is-hidden {
  opacity: 1;
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

  /* 平时隐藏、由 is-visible（JS auto-hide 控制）/ 键盘 focus 时淡入，保持画面干净 */
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

/* is-visible（JS auto-hide：指针唤出 / 起播头 2 秒 / 主动暂停）或键盘 focus 时淡入 —— 与左上标签同步 */
.video-media__poster:focus-within .video-media__fullscreen,
.video-media__fullscreen.is-visible {
  opacity: 1;
}

@media (prefers-reduced-motion: reduce) {
  .video-media__poster,
  .video-media__frame,
  .video-media__fullscreen,
  .video-media__tag {
    transition: none;
  }

  .video-media__spinner i {
    animation: none;
    opacity: 0.5;
  }
}
</style>
