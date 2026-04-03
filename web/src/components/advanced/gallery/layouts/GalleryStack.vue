<template>
  <div class="stack-root">
    <div class="stack-scroll">
      <div class="stack-track">
        <div v-for="(row, rowIdx) in imageRows" :key="rowIdx" class="stack-row">
          <div
            v-for="(cell, colIdx) in row"
            :key="getImageKey(cell.image, cell.idx)"
            class="stack-card"
            :class="{ 'stack-card--row-start': colIdx === 0 }"
            :style="cardStyle(cell.idx)"
          >
            <GalleryImageItem
              :image="cell.image"
              :src="resolvedSrcs[cell.idx] || ''"
              :alt="getAlt(cell.idx)"
              :loaded="isLoaded(cell.image, cell.idx)"
              button-class="stack-btn"
              frame-class="stack-frame"
              img-class="stack-img"
              @click="open(cell.idx, $event)"
              @load="markLoaded(cell.image, cell.idx)"
              @error="markLoaded(cell.image, cell.idx)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryStackProps } from './types'

/** 每行最多邮票数量 */
const STACK_MAX_PER_ROW = 5

const props = defineProps<GalleryStackProps>()

const imageRows = computed(() => {
  const list = props.images
  const rows: { image: App.Api.Ech0.FileObject; idx: number }[][] = []
  for (let i = 0; i < list.length; i++) {
    const rowIndex = Math.floor(i / STACK_MAX_PER_ROW)
    if (!rows[rowIndex]) rows[rowIndex] = []
    rows[rowIndex].push({ image: list[i], idx: i })
  }
  return rows
})

/** 稳定伪随机角度，约 [-10°, 10°] */
function stackRotateDeg(idx: number): number {
  const t = ((idx * 9301 + 49297) % 233280) / 233280
  return Math.round(-10 + t * 20)
}

/** 波浪式垂直偏移（px），与邮票尺寸成比例 */
function stackOffsetY(idx: number): number {
  return Math.round(Math.sin(idx * 1.07) * 10)
}

function cardStyle(idx: number): Record<string, string> {
  const deg = stackRotateDeg(idx)
  const y = stackOffsetY(idx)
  return {
    '--stack-rot': `${deg}deg`,
    '--stack-y': `${y}px`,
    zIndex: String(idx + 1),
  }
}
</script>

<style scoped>
/* 邮票外框总边长（含白边），勿放大：避免 flex 子项 min-width:auto 按原图撑满屏 */
.stack-root {
  --stamp-outer: 64px;
  /* 画在 img 上的白框宽度（box-sizing: border-box 含在邮票边长内） */
  --stamp-white-border: 2px;
  /* hover 放大倍数，与下方 min-height / padding 联动，避免被裁切 */
  --stack-hover-scale: 1.25;
  /* 横向重叠：仅压住一小部分（0.22 ≈ 22% 宽度），勿过大 */
  --stack-overlap: 0.22;
  /* 纵向：下一行向上叠到上一行，比例相对邮票高度 */
  --stack-row-overlap: 0.6;
  width: 100%;
  max-width: 100%;
  margin-left: auto;
  margin-right: auto;
  margin-bottom: 1rem;
}

.stack-scroll {
  position: relative;
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
  overflow-y: visible;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.12) transparent;
  box-sizing: border-box;
}

.stack-scroll::-webkit-scrollbar {
  height: 4px;
}

.stack-track {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
  width: max-content;
  max-width: 100%;
  min-width: 100%;
  box-sizing: border-box;
  padding: 6px 0 12px;
}

.stack-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  width: max-content;
  max-width: 100%;
  min-height: calc(
    var(--stamp-outer) * var(--stack-hover-scale) + var(--stamp-outer) * 0.28 + 0.75rem
  );
  padding-top: calc(var(--stamp-outer) * (var(--stack-hover-scale) - 1) * 0.38 + 8px);
  padding-bottom: calc(var(--stamp-outer) * (var(--stack-hover-scale) - 1) * 0.32 + 6px);
  padding-left: 4px;
  padding-right: 4px;
  box-sizing: border-box;
}

/* 下一行整体上移，与上一行纵向交错（全局 z-index 随 idx 递增，后行压在前行上） */
.stack-row + .stack-row {
  margin-top: calc(var(--stamp-outer) * -1 * var(--stack-row-overlap));
}

.stack-card {
  flex: 0 0 auto;
  width: var(--stamp-outer);
  margin-left: calc(var(--stamp-outer) * -1 * var(--stack-overlap));
  transform: rotate(var(--stack-rot, 0deg)) translateY(var(--stack-y, 0px));
  transform-origin: center center;
  transition: transform 0.2s ease;
}

.stack-card--row-start {
  margin-left: 0;
}

.stack-card:focus-within,
.stack-card:hover {
  z-index: 50 !important;
  transform: rotate(var(--stack-rot, 0deg)) translateY(var(--stack-y, 0px))
    scale(var(--stack-hover-scale));
}

/* 锁住点击区域，防止大图 intrinsic 宽度撑破 flex */
.stack-btn {
  display: block;
  box-sizing: border-box;
  width: var(--stamp-outer);
  height: var(--stamp-outer);
  min-width: 0;
  min-height: 0;
  padding: 0;
  overflow: visible;
}

/*
 * frame 在 GalleryImageItem 内部，非子组件根节点，scoped 必须用 :deep 才能命中。
 * 白边直接画在 img 的 border 上，不依赖与卡片同色的 padding。
 */
:deep(.gallery-image-frame.stack-frame) {
  width: var(--stamp-outer);
  height: var(--stamp-outer);
  min-width: 0;
  min-height: 0;
  padding: 0;
  box-sizing: border-box;
  background: transparent;
  border: none;
  border-radius: 0 !important;
  overflow: hidden;
  /* 仅细描边，避免灰黑投影；hover 不再加深阴影 */
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 85%, transparent);
}

:deep(.gallery-image-frame.stack-frame .image-skeleton) {
  border-radius: 0 !important;
}

:deep(.gallery-image-frame.stack-frame .echoimg.stack-img) {
  display: block;
  width: 100% !important;
  height: 100% !important;
  max-width: none !important;
  max-height: none !important;
  object-fit: cover;
  object-position: center;
  border-radius: 0 !important;
  border: var(--stamp-white-border) solid #fff;
  box-sizing: border-box;
  box-shadow: none !important;
}

@media (prefers-reduced-motion: reduce) {
  .stack-scroll {
    scroll-behavior: auto;
  }

  .stack-card {
    transform: translateY(var(--stack-y, 0px));
    transition: none;
  }

  .stack-card:focus-within,
  .stack-card:hover {
    transform: translateY(var(--stack-y, 0px));
  }
}
</style>
