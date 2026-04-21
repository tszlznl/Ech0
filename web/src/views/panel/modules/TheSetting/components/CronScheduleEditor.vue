<template>
  <div class="cron-editor">
    <div class="cron-editor__grid">
      <div class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.frequency') }}</label>
        <select v-model="frequency" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in frequencyOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'weekly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.weekday') }}</label>
        <select v-model.number="weekday" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in weekdayOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'monthly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.dayOfMonth') }}</label>
        <select v-model.number="monthday" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in monthdayOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div
        v-if="frequency === 'daily' || frequency === 'weekly' || frequency === 'monthly'"
        class="cron-editor__field"
      >
        <label class="cron-editor__label">{{ t('cronEditor.time') }}</label>
        <input
          type="time"
          v-model="timeString"
          :disabled="disabled"
          class="cron-editor__control cron-editor__control--time"
        />
      </div>

      <div v-if="frequency === 'hourly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.everyInterval') }}</label>
        <select v-model.number="hourlyInterval" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in hourlyOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'custom'" class="cron-editor__field cron-editor__field--wide">
        <label class="cron-editor__label">{{ t('cronEditor.customExpression') }}</label>
        <input
          v-model="customExpression"
          type="text"
          :disabled="disabled"
          :placeholder="t('cronEditor.customPlaceholder')"
          class="cron-editor__control cron-editor__control--input"
        />
      </div>
    </div>

    <div class="cron-editor__preview">
      <span class="cron-editor__preview-label">{{ t('cronEditor.preview') }}</span>
      <code class="cron-editor__preview-value">{{ modelValue || '—' }}</code>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

type Frequency = 'daily' | 'weekly' | 'monthly' | 'hourly' | 'custom'

const props = defineProps<{
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
}>()

const { t } = useI18n()

const frequency = ref<Frequency>('daily')
const hour = ref<number>(2)
const minute = ref<number>(0)
const weekday = ref<number>(0)
const monthday = ref<number>(1)
const hourlyInterval = ref<number>(6)
const customExpression = ref<string>('')

let suppressNextEmit = false

const frequencyOptions = computed(() => [
  { label: t('cronEditor.freqDaily'), value: 'daily' as Frequency },
  { label: t('cronEditor.freqWeekly'), value: 'weekly' as Frequency },
  { label: t('cronEditor.freqMonthly'), value: 'monthly' as Frequency },
  { label: t('cronEditor.freqHourly'), value: 'hourly' as Frequency },
  { label: t('cronEditor.freqCustom'), value: 'custom' as Frequency },
])

const weekdayOptions = computed(() => [
  { label: t('cronEditor.weekdaySun'), value: 0 },
  { label: t('cronEditor.weekdayMon'), value: 1 },
  { label: t('cronEditor.weekdayTue'), value: 2 },
  { label: t('cronEditor.weekdayWed'), value: 3 },
  { label: t('cronEditor.weekdayThu'), value: 4 },
  { label: t('cronEditor.weekdayFri'), value: 5 },
  { label: t('cronEditor.weekdaySat'), value: 6 },
])

const monthdayOptions = computed(() =>
  Array.from({ length: 28 }, (_, i) => ({
    label: t('cronEditor.dayN', { n: i + 1 }),
    value: i + 1,
  })),
)

const hourlyOptions = computed(() =>
  [1, 2, 3, 4, 6, 8, 12].map((n) => ({
    label: t('cronEditor.everyNHours', { n }),
    value: n,
  })),
)

const timeString = computed<string>({
  get: () => `${pad(hour.value)}:${pad(minute.value)}`,
  set: (v: string) => {
    const m = /^(\d{1,2}):(\d{1,2})$/.exec(v || '')
    if (!m) return
    hour.value = clamp(parseInt(m[1], 10), 0, 23)
    minute.value = clamp(parseInt(m[2], 10), 0, 59)
  },
})

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

function clamp(n: number, min: number, max: number): number {
  if (Number.isNaN(n)) return min
  return Math.min(Math.max(n, min), max)
}

function parseCron(expr: string) {
  const src = (expr || '').trim()

  const daily = /^(\d+)\s+(\d+)\s+\*\s+\*\s+\*$/.exec(src)
  if (daily) {
    minute.value = clamp(parseInt(daily[1], 10), 0, 59)
    hour.value = clamp(parseInt(daily[2], 10), 0, 23)
    frequency.value = 'daily'
    return
  }

  const weekly = /^(\d+)\s+(\d+)\s+\*\s+\*\s+([0-6])$/.exec(src)
  if (weekly) {
    minute.value = clamp(parseInt(weekly[1], 10), 0, 59)
    hour.value = clamp(parseInt(weekly[2], 10), 0, 23)
    weekday.value = clamp(parseInt(weekly[3], 10), 0, 6)
    frequency.value = 'weekly'
    return
  }

  const monthly = /^(\d+)\s+(\d+)\s+(\d+)\s+\*\s+\*$/.exec(src)
  if (monthly) {
    minute.value = clamp(parseInt(monthly[1], 10), 0, 59)
    hour.value = clamp(parseInt(monthly[2], 10), 0, 23)
    monthday.value = clamp(parseInt(monthly[3], 10), 1, 28)
    frequency.value = 'monthly'
    return
  }

  const hourly = /^0\s+\*\/(\d+)\s+\*\s+\*\s+\*$/.exec(src)
  if (hourly) {
    hourlyInterval.value = clamp(parseInt(hourly[1], 10), 1, 23)
    frequency.value = 'hourly'
    return
  }

  frequency.value = 'custom'
  customExpression.value = src
}

function buildCron(): string {
  switch (frequency.value) {
    case 'daily':
      return `${minute.value} ${hour.value} * * *`
    case 'weekly':
      return `${minute.value} ${hour.value} * * ${weekday.value}`
    case 'monthly':
      return `${minute.value} ${hour.value} ${monthday.value} * *`
    case 'hourly':
      return `0 */${hourlyInterval.value} * * *`
    case 'custom':
      return customExpression.value.trim()
    default:
      return ''
  }
}

watch(
  () => props.modelValue,
  (v) => {
    if (buildCron() === (v || '').trim()) return
    suppressNextEmit = true
    parseCron(v || '')
  },
  { immediate: true },
)

watch(
  [frequency, hour, minute, weekday, monthday, hourlyInterval, customExpression],
  () => {
    if (suppressNextEmit) {
      suppressNextEmit = false
      return
    }
    const next = buildCron()
    if (next !== props.modelValue) {
      emit('update:modelValue', next)
    }
  },
)
</script>

<style scoped>
.cron-editor {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.cron-editor__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

@media (width < 480px) {
  .cron-editor__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.cron-editor__field {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.cron-editor__field--wide {
  grid-column: 1 / -1;
}

.cron-editor__label {
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.cron-editor__control {
  width: 100%;
  height: 1.9rem;
  padding: 0 0.55rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--input-border-color);
  background: var(--input-bg-color);
  color: var(--input-text-color);
  font-size: 0.82rem;
  line-height: 1;
  box-shadow: var(--shadow-sm);
  outline: none;
  appearance: none;
  -webkit-appearance: none;
  background-repeat: no-repeat;
  background-position: right 0.5rem center;
  background-size: 0.8rem 0.8rem;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.cron-editor__control:hover:not(:disabled) {
  border-color: var(--color-border-strong);
}

.cron-editor__control:focus-visible {
  border-color: var(--input-focus-ring-color);
  box-shadow:
    0 0 0 2px var(--input-focus-ring-color),
    var(--shadow-sm);
}

.cron-editor__control:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* select 右侧箭头 */
select.cron-editor__control {
  padding-right: 1.75rem;
  cursor: pointer;
  background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'><path fill='%23888888' d='m12 15.4l-6-6L7.4 8l4.6 4.6L16.6 8L18 9.4z'/></svg>");
}

.cron-editor__control--time {
  font-family: var(--font-family-mono);
}

.cron-editor__control--input {
  font-family: var(--font-family-mono);
  font-size: 0.8rem;
}

.cron-editor__preview {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.55rem;
  border-radius: var(--radius-sm);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
  width: fit-content;
  max-width: 100%;
}

.cron-editor__preview-label {
  font-size: 0.65rem;
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.cron-editor__preview-value {
  font-family: var(--font-family-mono);
  font-size: 0.78rem;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
