<template>
  <div class="extension-skeleton" :style="styleVars" aria-hidden="true">
    <div class="skeleton-line w-16"></div>
    <div class="skeleton-line w-40"></div>
    <div class="skeleton-line w-32"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    minHeight?: number
  }>(),
  {
    minHeight: 104,
  },
)

const styleVars = computed(() => ({
  minHeight: `${props.minHeight}px`,
}))
</script>

<style scoped>
.extension-skeleton {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem;
}

.skeleton-line {
  height: 0.7rem;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    rgba(130, 130, 130, 0.14) 0%,
    rgba(130, 130, 130, 0.22) 50%,
    rgba(130, 130, 130, 0.14) 100%
  );
  background-size: 200% 100%;
  animation: skeleton-flow 1.8s ease-in-out infinite;
}

@keyframes skeleton-flow {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-line {
    animation: none;
  }
}
</style>
