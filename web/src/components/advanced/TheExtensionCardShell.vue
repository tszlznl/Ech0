<template>
  <div class="extension-card-shell" :class="[sizeClass, paddingClass]">
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    size?: 'default' | 'wide'
    padding?: 'none' | 'compact' | 'content'
  }>(),
  {
    size: 'default',
    padding: 'none',
  },
)

const sizeClass = computed(() =>
  props.size === 'wide' ? 'extension-card-shell--wide' : 'extension-card-shell--default',
)

const paddingClass = computed(() => {
  if (props.padding === 'compact') return 'extension-card-shell--padding-compact'
  if (props.padding === 'content') return 'extension-card-shell--padding-content'
  return 'extension-card-shell--padding-none'
})
</script>

<style scoped>
.extension-card-shell {
  width: 100%;
  min-width: 0;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.extension-card-shell--default {
  max-width: 24rem;
}

.extension-card-shell--wide {
  max-width: 41.25rem;
}

.extension-card-shell--padding-none {
  padding: 0;
}

.extension-card-shell--padding-compact {
  padding: 0.25rem;
}

.extension-card-shell--padding-content {
  padding: 0.75rem;
}
</style>
