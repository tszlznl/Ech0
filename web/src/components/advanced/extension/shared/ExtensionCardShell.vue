<template>
  <div class="extension-card-shell" :class="[sizeClass, paddingClass]">
    <div v-if="hasHeader" class="extension-card-shell__header">
      <slot name="header">
        <span class="extension-card-shell__header-icon">
          <slot name="header-icon"></slot>
        </span>
        <span class="extension-card-shell__header-label">{{ headerLabel }}</span>
        <span
          v-if="headerBadge !== undefined && headerBadge !== ''"
          class="extension-card-shell__header-badge"
        >
          {{ headerBadge }}
        </span>
      </slot>
    </div>
    <div class="extension-card-shell__body">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue'

const props = withDefaults(
  defineProps<{
    size?: 'default' | 'wide'
    padding?: 'none' | 'compact' | 'content'
    headerLabel?: string
    headerBadge?: string | number
  }>(),
  {
    size: 'default',
    padding: 'none',
    headerLabel: '',
    headerBadge: undefined,
  },
)

const slots = useSlots()

const hasHeader = computed(
  () => Boolean(slots.header) || Boolean(slots['header-icon']) || Boolean(props.headerLabel),
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

.extension-card-shell--default,
.extension-card-shell--wide {
  max-width: 100%;
}

.extension-card-shell__header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.65rem;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  background: var(--color-bg-muted);
  border-bottom: 1px solid var(--color-border-subtle);
  user-select: none;
}

.extension-card-shell__header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 0.95rem;
  height: 0.95rem;
  color: var(--color-text-secondary);
}

.extension-card-shell__header-icon :deep(svg) {
  width: 100%;
  height: 100%;
}

.extension-card-shell__header-label {
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.01em;
}

.extension-card-shell__header-badge {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.1rem;
  height: 1.1rem;
  padding: 0 0.4rem;
  border-radius: 9999px;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.extension-card-shell--padding-none .extension-card-shell__body {
  padding: 0;
}

.extension-card-shell--padding-compact .extension-card-shell__body {
  padding: 0.25rem;
}

.extension-card-shell--padding-content .extension-card-shell__body {
  padding: 0.75rem;
}
</style>
