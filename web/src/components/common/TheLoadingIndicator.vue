<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="loading-indicator"
    :class="[{ 'loading-indicator--center': center }, `loading-indicator--${size}`]"
    role="status"
    aria-live="polite"
  >
    <span class="loading-spinner" aria-hidden="true">
      <i v-for="segment in segments" :key="segment" :style="{ '--segment': segment - 1 }"></i>
    </span>
    <span v-if="label" class="loading-label">{{ label }}</span>
    <span v-else class="sr-only">Loading</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    size?: 'sm' | 'md' | 'lg'
    label?: string
    center?: boolean
  }>(),
  {
    size: 'md',
    label: '',
    center: true,
  },
)

const segments = Array.from({ length: 12 }, (_, index) => index + 1)
</script>

<style scoped>
.loading-indicator {
  --spinner-size: 22px;
  --spinner-color: var(--color-text-muted);
  --spinner-stroke: 2.4px;
  --label-size: 0.95rem;

  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 0.55rem;
  color: var(--spinner-color);
}

.loading-indicator--center {
  display: flex;
  width: fit-content;
  margin-inline: auto;
}

.loading-indicator--sm {
  --spinner-size: 16px;
  --spinner-stroke: 2px;
  --label-size: 0.85rem;
}

.loading-indicator--lg {
  --spinner-size: 28px;
  --spinner-stroke: 2.8px;
  --label-size: 1.05rem;
}

.loading-spinner {
  position: relative;
  width: var(--spinner-size);
  height: var(--spinner-size);
}

.loading-spinner i {
  position: absolute;
  top: 0;
  left: 50%;
  width: var(--spinner-stroke);
  height: calc(var(--spinner-size) * 0.24);
  border-radius: 999px;
  background: currentColor;
  transform-origin: center calc(var(--spinner-size) / 2);
  transform: translateX(-50%) rotate(calc(var(--segment) * 30deg));
  animation: spinner-fade 1.35s linear infinite;
  animation-delay: calc(var(--segment) * 0.11s);
}

.loading-label {
  font-size: var(--label-size);
  line-height: 1.2;
  color: var(--spinner-color);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  border: 0;
  white-space: nowrap;
}

@keyframes spinner-fade {
  0% {
    opacity: 1;
  }

  100% {
    opacity: 0.12;
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-spinner i {
    animation: none;
    opacity: 0.55;
  }
}
</style>
