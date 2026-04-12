<template>
  <div class="activity-log p-1">
    <div class="activity-head">
      <div class="activity-chip">7 DAYS</div>
      <div class="activity-title-wrap">
        <div class="activity-title">{{ t('dashboard.activityTitle') }}</div>
        <div class="activity-title-accent">{{ t('dashboard.activityAccent') }}</div>
      </div>
    </div>
    <div class="line-chart-wrap">
      <svg
        v-if="last7Days.length >= 2"
        class="line-chart-svg"
        :viewBox="`0 0 ${chartW} ${chartH}`"
        preserveAspectRatio="xMidYMid meet"
      >
        <path :d="chartFillPath" class="line-chart-fill" />
        <polyline :points="chartPoints" class="line-chart-line" />
        <circle
          v-for="(dot, i) in chartDotPositions"
          :key="i"
          :cx="dot.x"
          :cy="dot.y"
          r="6"
          class="line-chart-dot"
          @mouseenter="showTooltip(dot, $event)"
          @mouseleave="hideTooltip"
        />
      </svg>
      <div v-if="last7Days.length >= 2" class="line-chart-labels">
        <span v-for="(d, i) in last7Days" :key="i" class="line-chart-label">
          {{ (d.date || '').slice(5) }}
        </span>
      </div>
    </div>
    <div
      v-if="tooltip.visible"
      class="chart-tooltip"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
    >
      {{ tooltip.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetHeatMap } from '@/service/api'

const { t } = useI18n()

const heatmapData = ref<App.Api.Ech0.HeatMap>([])

const last7Days = computed(() => {
  const all = heatmapData.value
  return all.length >= 7 ? all.slice(-7) : all
})

const chartMax = computed(() => Math.max(...last7Days.value.map((d) => d.count), 1))

const chartW = 600
const chartH = 200
const chartPadX = 20
const chartPadY = 24

const chartPoints = computed(() => {
  const days = last7Days.value
  if (days.length < 2) return ''
  const stepX = (chartW - chartPadX * 2) / (days.length - 1)
  return days
    .map((d, i) => {
      const x = chartPadX + i * stepX
      const y = chartPadY + (chartH - chartPadY * 2) * (1 - d.count / chartMax.value)
      return `${x},${y}`
    })
    .join(' ')
})

const chartFillPath = computed(() => {
  const days = last7Days.value
  if (days.length < 2) return ''
  const stepX = (chartW - chartPadX * 2) / (days.length - 1)
  const pts = days.map((d, i) => {
    const x = chartPadX + i * stepX
    const y = chartPadY + (chartH - chartPadY * 2) * (1 - d.count / chartMax.value)
    return `${x},${y}`
  })
  return `M${pts[0]} L${pts.join(' L')} L${chartW - chartPadX},${chartH - chartPadY} L${chartPadX},${chartH - chartPadY} Z`
})

const chartDotPositions = computed(() => {
  const days = last7Days.value
  if (days.length < 2) return []
  const stepX = (chartW - chartPadX * 2) / (days.length - 1)
  return days.map((d, i) => ({
    x: chartPadX + i * stepX,
    y: chartPadY + (chartH - chartPadY * 2) * (1 - d.count / chartMax.value),
    label: (d.date || '').slice(5),
    count: d.count,
  }))
})

const tooltip = ref({ visible: false, text: '', x: 0, y: 0 })

function showTooltip(dot: { label: string; count: number }, event: MouseEvent) {
  tooltip.value.text = `${dot.label}: ${dot.count}`
  const rect = (event.target as SVGElement).getBoundingClientRect()
  tooltip.value.x = rect.left + rect.width / 2
  tooltip.value.y = rect.top - 28
  tooltip.value.visible = true
}

function hideTooltip() {
  tooltip.value.visible = false
}

onMounted(async () => {
  try {
    const res = await fetchGetHeatMap()
    if (res.data) heatmapData.value = res.data
  } catch {}
})
</script>

<style scoped>
.activity-log {
  min-width: 0;
}

.activity-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.activity-chip {
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-size: 0.66rem;
  letter-spacing: 0.15em;
  padding: 0.08rem 0.45rem;
  font-family: var(--font-family-mono);
  transform: rotate(-1.8deg);
}

.activity-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.activity-title {
  font-family: Georgia, 'Times New Roman', var(--font-family-display);
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.activity-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 0.95rem;
  margin-top: -2px;
}

.line-chart-wrap {
  width: 100%;
}

.line-chart-svg {
  width: 100%;
  height: 7rem;
  display: block;
}

.line-chart-fill {
  fill: var(--color-accent-soft);
  opacity: 0.35;
}

.line-chart-line {
  fill: none;
  stroke: var(--color-accent);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.line-chart-dot {
  fill: var(--color-bg-surface);
  stroke: var(--color-accent);
  stroke-width: 3;
  cursor: default;
  transition: r 0.15s ease;
}

.line-chart-dot:hover {
  r: 9;
}

.line-chart-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.3rem;
}

.line-chart-label {
  font-size: 0.62rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  letter-spacing: -0.02em;
}

.chart-tooltip {
  position: fixed;
  z-index: 50;
  padding: 0.15rem 0.45rem;
  background: var(--color-text-primary);
  color: var(--color-bg-surface);
  font-size: 0.68rem;
  border-radius: var(--radius-sm);
  pointer-events: none;
  font-family: var(--font-family-mono);
  transform: translateX(-50%);
}
</style>
