<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="panel-card rounded-[var(--radius-lg)] p-4 border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)] shadow-[var(--shadow-sm)] overflow-auto"
    :class="{
      'panel-card--dashed': borderStyle === 'dashed',
      'panel-card--solid': borderStyle === 'solid',
    }"
  >
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    borderStyle?: 'dashed' | 'solid'
  }>(),
  {
    borderStyle: 'dashed',
  },
)
</script>

<style scoped>
.panel-card--dashed {
  border-style: dashed;
}

.panel-card--solid {
  border-style: solid;
}
</style>
