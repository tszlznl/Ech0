<template>
  <ExtensionCardShell :header-label="t('extensionCard.location')">
    <template #header-icon><MapPin /></template>

    <div class="location-card__body">
      <button
        ref="triggerEl"
        type="button"
        class="location-card__trigger"
        :aria-expanded="open"
        :aria-label="t('extensionCard.location')"
        @click="togglePopover"
      >
        <span class="location-card__meta">
          <span class="location-card__text">{{ displayText }}</span>
          <span class="location-card__coords">{{ coordsText }}</span>
        </span>
      </button>

      <Transition name="location-popover">
        <div v-if="open" ref="popoverEl" class="location-card__popover">
          <LocationPicker :lat-lng="latLng" readonly />
        </div>
      </Transition>
    </div>
  </ExtensionCardShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import MapPin from '@/components/icons/mappin.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'
import LocationPicker from '../shared/LocationPicker.vue'

const { t } = useI18n()

const props = defineProps<{
  location: { placeholder: string; latitude: number; longitude: number }
}>()

const open = ref(false)
const triggerEl = ref<HTMLButtonElement | null>(null)
const popoverEl = ref<HTMLDivElement | null>(null)

const latLng = computed(() => ({
  lat: props.location.latitude,
  lng: props.location.longitude,
}))

const displayText = computed(() => props.location.placeholder || coordsText.value)

const coordsText = computed(() => {
  const lat = Number(props.location.latitude).toFixed(2)
  const lng = Number(props.location.longitude).toFixed(2)
  return `${lat}°, ${lng}°`
})

function togglePopover() {
  open.value = !open.value
}

onClickOutside(
  popoverEl,
  () => {
    open.value = false
  },
  { ignore: [triggerEl] },
)
</script>

<style scoped>
.location-card__body {
  position: relative;
  padding: 0.5rem 0.65rem;
}

.location-card__trigger {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  width: 100%;
  padding: 0.35rem 0.45rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-muted);
  color: var(--color-text-secondary);
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
}

.location-card__trigger:hover {
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
}

.location-card__trigger:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-focus-ring);
}

.location-card__meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.location-card__text {
  font-size: 0.85rem;
  color: var(--color-text-primary);
  font-weight: 500;
  line-height: 1.25;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.location-card__coords {
  font-size: 0.7rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  line-height: 1.25;
}

.location-card__popover {
  margin-top: 0.5rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

.location-popover-enter-active,
.location-popover-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.location-popover-enter-from,
.location-popover-leave-to {
  opacity: 0;
  transform: translateY(-0.2rem);
}
</style>
