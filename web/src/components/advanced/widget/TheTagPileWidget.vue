<template>
  <div class="px-2">
    <div class="widget bg-transparent! w-full max-w-[19rem] mx-auto rounded-md p-1">
      <div class="tag-pile-head mb-2">
        <div class="tag-pile-chip">TAGS</div>
        <div class="tag-pile-title-wrap">
          <div class="tag-pile-title">{{ t('tagPileWidget.title') }}</div>
          <div class="tag-pile-title-accent">{{ t('tagPileWidget.accent') }}</div>
        </div>
      </div>
      <div class="tag-pile-stage" :style="{ minHeight }">
        <div v-if="layoutItems.length === 0" class="tag-pile-empty">
          {{ t('tagPileWidget.empty') }}
        </div>
        <span
          v-for="item in layoutItems"
          :key="item.key"
          class="tag-pill"
          :style="{
            left: item.left,
            bottom: item.bottom,
            transform: item.transform,
            backgroundColor: item.backgroundColor,
            color: item.color,
            zIndex: item.zIndex,
          }"
        >
          {{ item.label }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

interface Props {
  layoutSeed?: number | string
  minHeight?: string
  ech0Version?: string
}

type LayoutTag = {
  key: string
  label: string
  left: string
  bottom: string
  transform: string
  zIndex: string
  backgroundColor: string
  color: string
}

const props = withDefaults(defineProps<Props>(), {
  layoutSeed: 'ech0-tag-pile',
  minHeight: '8.85rem',
  ech0Version: '--',
})

const { t } = useI18n()

const palette = [
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-accent) 14%)',
    color: 'var(--color-text-secondary)',
  },
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-text-muted) 12%)',
    color: 'var(--color-text-secondary)',
  },
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-border-strong) 16%)',
    color: 'var(--color-text-primary)',
  },
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-accent-soft) 14%)',
    color: 'var(--color-text-secondary)',
  },
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-text-primary) 10%)',
    color: 'var(--color-text-secondary)',
  },
  {
    bg: 'color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-accent) 12%)',
    color: 'var(--color-text-secondary)',
  },
]

const fixedTags = computed(() => [
  `Ech0 ${props.ech0Version}`,
  'Go',
  'Gin',
  'Gorm',
  'SQLite',
  'Vue',
  'Vite',
  'TypeScript',
  'Pinia',
  'Vue-I18n',
  'Pnpm',
])

const normalizedTags = computed(() =>
  fixedTags.value
    .map((tag) => tag.trim())
    .filter(Boolean)
    .slice(0, 24),
)

const normalizeSeed = (seed: number | string) => {
  if (typeof seed === 'number') return Math.abs(seed) || 1
  let value = 2166136261
  for (let i = 0; i < seed.length; i += 1) {
    value ^= seed.charCodeAt(i)
    value = Math.imul(value, 16777619)
  }
  return value >>> 0 || 1
}

const createRandom = (seed: number) => {
  let value = seed >>> 0
  return () => {
    value += 0x6d2b79f5
    let t = value
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const layoutItems = computed<LayoutTag[]>(() => {
  const tagSource = normalizedTags.value
  if (tagSource.length === 0) return []

  const seed = normalizeSeed(props.layoutSeed) ^ normalizeSeed(fixedTags.value.join('|'))
  const random = createRandom(seed)
  const rows = 3
  const cols = Math.max(1, Math.ceil(tagSource.length / rows))

  return tagSource.map((label, index) => {
    const color = palette[Math.floor(random() * palette.length)] ?? palette[0]
    const row = index % rows
    const col = Math.floor(index / rows)
    const baseLeft = ((col + 0.5) / cols) * 100
    const rowOffset = row === 1 ? 3 : row === 2 ? -3 : 0
    const jitterLeft = (random() - 0.5) * 3.5
    const left = Math.min(92, Math.max(8, baseLeft + rowOffset + jitterLeft))
    const bottom = 6 + row * 27 + random() * 4
    const rotate = -12 + random() * 24

    return {
      key: `${label}-${index}`,
      label: label.toUpperCase(),
      left: `${left.toFixed(2)}%`,
      bottom: `${bottom.toFixed(1)}px`,
      transform: `translateX(-50%) rotate(${rotate.toFixed(2)}deg)`,
      zIndex: String(index + 1),
      backgroundColor: color.bg,
      color: color.color,
    }
  })
})

const minHeight = computed(() => props.minHeight)
</script>

<style scoped>
.tag-pile-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
}

.tag-pile-chip {
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-size: 0.66rem;
  letter-spacing: 0.15em;
  padding: 0.08rem 0.45rem;
  font-family: var(--font-family-mono);
  transform: rotate(-1.8deg);
}

.tag-pile-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.tag-pile-title {
  font-family: Georgia, 'Times New Roman', var(--font-family-display);
  color: var(--color-text-primary);
  font-size: 1.3rem;
  font-weight: 600;
}

.tag-pile-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 0.95rem;
  margin-top: -2px;
}

.tag-pile-stage {
  position: relative;
  margin-top: 0.65rem;
  border-radius: 0.75rem;
  background: transparent;
  border: 1px dashed var(--color-border-subtle);
  overflow: hidden;
  padding: 0.25rem;
}

.tag-pile-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  letter-spacing: 0.02em;
}

.tag-pill {
  position: absolute;
  padding: 0.32rem 0.78rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 78%, transparent);
  box-shadow:
    0 1px 2px color-mix(in srgb, var(--color-bg-mask) 18%, transparent),
    inset 0 1px 0 color-mix(in srgb, white 12%, transparent);
  font-family: var(--font-family-mono);
  font-size: 0.68rem;
  line-height: 1;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  white-space: nowrap;
  user-select: none;
}
</style>
