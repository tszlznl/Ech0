<template>
  <button
    @click="scrollToTop"
    id="backToTop"
    type="button"
    aria-label="返回顶部"
    title="返回顶部"
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

const scrollToTop = () => {
  if (props.target) {
    props.target.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>
