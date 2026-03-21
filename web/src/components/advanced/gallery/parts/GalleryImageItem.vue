<template>
  <button
    class="bg-transparent border-0 p-0 cursor-pointer"
    :class="buttonClass"
    type="button"
    @click="handleClick"
  >
    <div class="gallery-image-frame" :class="frameClass" :style="frameStyle">
      <div v-if="!loaded" class="image-skeleton" aria-hidden="true"></div>
      <img
        :src="src"
        :alt="alt"
        :width="image.width || undefined"
        :height="image.height || undefined"
        :loading="loading"
        class="echoimg transition-opacity duration-300"
        :class="[imgClass, loaded ? 'opacity-100' : 'opacity-0']"
        @load="$emit('load')"
        @error="$emit('error')"
      />
      <slot></slot>
    </div>
  </button>
</template>

<script setup lang="ts">
const emit = defineEmits<{
  (e: 'click', sourceElement: HTMLElement | null): void
  (e: 'load'): void
  (e: 'error'): void
}>()

const handleClick = (event: MouseEvent) => {
  const sourceElement = event.currentTarget
  emit('click', sourceElement instanceof HTMLElement ? sourceElement : null)
}

withDefaults(
  defineProps<{
    image: App.Api.Ech0.FileObject
    src: string
    alt: string
    loaded: boolean
    loading?: 'lazy' | 'eager'
    buttonClass?: string
    frameClass?: string
    imgClass?: string
    frameStyle?: Record<string, string>
  }>(),
  {
    loading: 'lazy',
    buttonClass: 'w-fit',
    frameClass: '',
    imgClass: 'block max-w-full h-auto',
    frameStyle: undefined,
  },
)
</script>

<style scoped>
.echoimg {
  border-radius: 8px;
  box-shadow:
    0 1px 2px rgba(0, 0, 0, 0.02),
    0 2px 4px rgba(0, 0, 0, 0.02),
    0 4px 8px rgba(0, 0, 0, 0.02),
    0 8px 16px rgba(0, 0, 0, 0.02);
}

.gallery-image-frame {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
}

.image-skeleton {
  position: absolute;
  inset: 0;
  border-radius: 8px;
  pointer-events: none;
  background: radial-gradient(
      120% 120% at 50% 30%,
      rgba(160, 160, 160, 0.18) 0%,
      rgba(160, 160, 160, 0.1) 55%,
      rgba(160, 160, 160, 0.07) 100%
    ),
    linear-gradient(180deg, rgba(140, 140, 140, 0.08), rgba(140, 140, 140, 0.11));
  animation: skeleton-breathe 2.2s ease-in-out infinite;
}

@keyframes skeleton-breathe {
  0%,
  100% {
    opacity: 0.42;
  }
  50% {
    opacity: 0.68;
  }
}

@media (prefers-reduced-motion: reduce) {
  .image-skeleton {
    animation: none;
    opacity: 0.55;
  }
}
</style>
