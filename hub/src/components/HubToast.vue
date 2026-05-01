<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    message: string | null
    durationMs?: number
  }>(),
  { durationMs: 3000 },
)

const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

function clearTimer() {
  if (timer !== null) {
    clearTimeout(timer)
    timer = null
  }
}

watch(
  () => props.message,
  (msg) => {
    clearTimer()
    if (!msg) {
      visible.value = false
      return
    }
    visible.value = true
    timer = setTimeout(() => {
      visible.value = false
      timer = null
    }, props.durationMs)
  },
  { immediate: true },
)

onBeforeUnmount(clearTimer)
</script>

<template>
  <Transition name="hub-toast">
    <div v-if="visible && message" class="hub-toast-wrap">
      <div class="hub-toast" role="status" aria-live="polite">
        {{ message }}
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.hub-toast-wrap {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 2rem;
  z-index: 60;
  display: flex;
  justify-content: center;
  pointer-events: none;
}

.hub-toast {
  padding: 0.5rem 1rem;
  border-radius: 999px;
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  font-size: 0.75rem;
  letter-spacing: 0.04em;
  border: 1px solid var(--color-border-subtle);
  box-shadow: 0 6px 24px rgb(0 0 0 / 0.08);
  white-space: nowrap;
}

.hub-toast-enter-active,
.hub-toast-leave-active {
  transition:
    opacity 200ms ease,
    transform 200ms ease;
}
.hub-toast-enter-from,
.hub-toast-leave-to {
  opacity: 0;
  transform: translateY(0.5rem);
}
</style>
