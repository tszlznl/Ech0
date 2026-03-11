<template>
  <button
    :class="[
      'cursor-pointer p-1.5 rounded-[var(--btn-radius)] ring-inset ring-1 ring-[var(--btn-ring-color)] text-[var(--btn-text-color)] outline-none shadow-[var(--btn-shadow)] transition-colors duration-200',
      hasBg ? '' : 'bg-[var(--btn-bg-color)]',
      disabled
        ? 'cursor-not-allowed opacity-70'
        : 'hover:bg-[var(--btn-hover-bg-color)] hover:ring-[var(--btn-hover-border-color)] focus-visible:ring-2 focus-visible:ring-[var(--btn-focus-ring-color)]',
      props.class,
    ]"
    :disabled="disabled"
    @click="onClick"
  >
    <span v-if="icon" class="flex items-center justify-center">
      <component :is="icon" class="w-full h-full" />
    </span>
    <span><slot /></span>
  </button>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed } from 'vue'

const props = defineProps<{
  icon?: Component
  disabled?: boolean
  class?: string // 接收父组件传递的 class
}>()

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

// const customClass = props.class
const hasBg = computed(() => props.class?.includes('bg-') || props.class?.includes('!bg-'))

function onClick(event: MouseEvent) {
  if (!props.disabled) emit('click', event)
}
</script>

<style scoped></style>
