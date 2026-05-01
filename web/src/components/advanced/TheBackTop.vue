<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <button
    @click="scrollToTop"
    id="backToTop"
    type="button"
    aria-label="返回顶部"
    v-tooltip="'返回顶部'"
    class="cursor-pointer rounded-full shadow hover:shadow-md bg-[var(--color-bg-surface)] ring-1 ring-inset ring-[var(--color-border-subtle)] z-50"
  >
    <Arrowup class="w-full h-full" />
  </button>
</template>

<script setup lang="ts">
import Arrowup from '../icons/arrowup.vue'

const props = defineProps<{
  target?: HTMLElement | null
}>()

const supportsSmoothScroll = () => 'scrollBehavior' in document.documentElement.style

const canScrollElement = (el: HTMLElement) => {
  if (el.scrollHeight <= el.clientHeight + 1) return false
  const overflowY = window.getComputedStyle(el).overflowY
  return overflowY === 'auto' || overflowY === 'scroll' || overflowY === 'overlay'
}

const scrollToTop = () => {
  const behavior: ScrollBehavior = supportsSmoothScroll() ? 'smooth' : 'auto'
  if (props.target && canScrollElement(props.target)) {
    props.target.scrollTo({ top: 0, behavior })
    return
  }
  window.scrollTo({ top: 0, behavior })
}
</script>
