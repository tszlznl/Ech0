<template>
  <slot v-if="isDisabled" />
  <VTooltip v-else :placement="placementValue" :delay="delay" :distance="distance" :theme="theme">
    <slot />
    <template #popper>
      {{ normalizedContent }}
    </template>
  </VTooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Tooltip as VTooltip } from 'floating-vue'

type TooltipDelay = number | { show: number; hide: number }
type TooltipPlacement =
  | 'top'
  | 'top-start'
  | 'top-end'
  | 'right'
  | 'right-start'
  | 'right-end'
  | 'bottom'
  | 'bottom-start'
  | 'bottom-end'
  | 'left'
  | 'left-start'
  | 'left-end'

const props = withDefaults(
  defineProps<{
    content?: string | number | null
    placement?: TooltipPlacement
    delay?: TooltipDelay
    distance?: number
    disabled?: boolean
    theme?: string
  }>(),
  {
    content: '',
    placement: 'top',
    delay: 0,
    distance: 10,
    disabled: false,
    theme: 'tooltip',
  },
)

const normalizedContent = computed(() => (props.content == null ? '' : String(props.content)))
const placementValue = computed<TooltipPlacement>(() => props.placement ?? 'top')
const isDisabled = computed(() => props.disabled || normalizedContent.value.trim().length === 0)
</script>
