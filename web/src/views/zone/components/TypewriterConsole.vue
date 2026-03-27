<template>
  <div
    class="relative pointer-events-auto scale-90 md:scale-100 origin-bottom transition-transform duration-300"
  >
    <div
      class="console-shell relative w-[380px] md:w-[440px] rounded-[3rem] p-6 border-t"
      style="
        box-shadow:
          0 50px 60px -20px rgba(0, 0, 0, 0.6),
          inset 0 2px 10px rgba(255, 255, 255, 0.4),
          inset 0 -10px 20px rgba(0, 0, 0, 0.2);
      "
    >
      <div
        class="absolute inset-0 rounded-[3rem] bg-[url('https://www.transparenttextures.com/patterns/noise-lines.png')] opacity-20 pointer-events-none mix-blend-overlay"
      ></div>
      <div
        ref="paperSlotRef"
        class="console-paper-slot absolute -top-3 left-1/2 -translate-x-1/2 w-64 h-5 rounded-full shadow-inner border-b z-0"
      ></div>
      <div
        class="absolute top-2 left-10 right-10 h-1 bg-gradient-to-r from-transparent via-white/60 to-transparent rounded-full blur-[1px]"
      ></div>

      <div class="console-panel relative rounded-3xl p-4 border">
        <div class="flex justify-between items-end mb-3 px-2">
          <div class="flex flex-col">
            <div class="flex items-center gap-1">
              <div class="console-status-dot w-1.5 h-1.5 rounded-full animate-pulse"></div>
              <span class="console-label text-[10px] font-black tracking-widest uppercase">
                Auto-Feed
              </span>
            </div>
            <div class="console-meta text-[9px] font-bold tracking-[0.2em] uppercase mt-0.5">
              Series 9000
            </div>
          </div>
          <div class="console-label flex items-center gap-1 opacity-60">
            <span class="text-xs">5G</span>
          </div>
        </div>

        <div class="console-screen rounded-xl p-1 pb-0 border-b-2 relative overflow-hidden">
          <div
            class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] z-10 bg-[length:100%_2px,3px_100%] pointer-events-none"
          ></div>

          <div class="relative z-20 p-2">
            <div class="console-screen-meta flex justify-between text-xs mb-1 border-b pb-1">
              <span>COMPOSE_MODE</span>
              <span>{{ modelLength }} chars</span>
            </div>

            <textarea
              :value="modelValue"
              class="console-input w-full h-20 bg-transparent resize-none outline-none text-xl tracking-wider leading-tight"
              placeholder="TYPE MESSAGE HERE..."
              spellcheck="false"
              @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
            />
          </div>

          <div
            v-if="deleteConfirm"
            class="console-alert-layer absolute inset-0 z-30 flex flex-col items-center justify-center"
          >
            <span class="console-alert-text text-lg animate-pulse"
              >CLICK AGAIN TO CLEAR ALL PRINTS...</span
            >
            <div class="console-alert-count text-3xl mt-2 font-bold">{{ countdown }}</div>
          </div>
        </div>

        <div class="mt-5 grid grid-cols-5 gap-3 items-center">
          <button
            class="console-round-btn col-span-1 aspect-square rounded-full active:shadow-none active:translate-y-1 transition-all border-t flex items-center justify-center group relative overflow-hidden"
            v-tooltip="stampEnabled ? 'Stamp Enabled' : 'Enable Stamp'"
            @click="stampEnabled = !stampEnabled"
          >
            <div
              class="absolute inset-0 bg-gradient-to-tr from-transparent to-white/10 rounded-full"
            ></div>
            <span
              :class="[
                'text-base transition-all duration-200',
                stampEnabled
                  ? 'text-yellow-400 drop-shadow-[0_0_8px_rgba(250,204,21,0.8)]'
                  : 'text-[var(--color-text-muted)]',
              ]"
            >
              印
            </span>
          </button>

          <button
            class="console-round-btn col-span-1 aspect-square rounded-full active:shadow-none active:translate-y-1 transition-all border-t flex items-center justify-center group relative overflow-hidden"
            v-tooltip="deleteConfirm ? 'Click again to confirm' : 'Clear all prints'"
            @click="handleDeleteClick"
          >
            <div
              class="absolute inset-0 bg-gradient-to-tr from-transparent to-white/10 rounded-full"
            ></div>
            <span
              :class="[
                'text-base transition-all duration-200',
                deleteConfirm
                  ? 'text-[var(--zone-console-alert)] drop-shadow-[0_0_8px_rgba(239,68,68,0.8)]'
                  : 'text-[var(--color-text-muted)]',
              ]"
            >
              删
            </span>
          </button>

          <div class="col-span-1 flex flex-col items-center gap-1 px-2">
            <div class="console-divider-line w-full h-1 rounded-full"></div>
            <div class="console-divider-line w-full h-1 rounded-full"></div>
            <div class="console-divider-line w-full h-1 rounded-full"></div>
            <div class="console-divider-line w-full h-1 rounded-full"></div>
          </div>

          <button
            class="console-print-btn col-span-2 h-14 rounded-lg active:shadow-none active:translate-y-[5px] transition-all border-t flex items-center justify-center gap-2 relative overflow-hidden disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="!canPrint"
            @click="handlePrint"
          >
            <div
              class="absolute top-0 left-0 w-full h-1/2 bg-gradient-to-b from-white/20 to-transparent"
            ></div>
            <span class="console-print-label text-2xl font-bold drop-shadow-sm mt-1">PRINT</span>
          </button>
        </div>
      </div>

      <div
        class="console-badge absolute bottom-3 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded text-[8px] font-mono tracking-widest border shadow-sm"
      >
        ECH0
      </div>
    </div>

    <div
      class="console-bottom-glow absolute -bottom-4 left-10 right-10 h-8 blur-xl rounded-full z-[-1]"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  print: [text: string, withStamp: boolean]
  'clear-all': []
}>()

const stampEnabled = ref(false)
const deleteConfirm = ref(false)
const countdown = ref(5)
const timer = ref<number | null>(null)
const paperSlotRef = ref<HTMLDivElement | null>(null)

const modelLength = computed(() => props.modelValue.length)
const canPrint = computed(() => props.modelValue.trim().length > 0)

const clearTimer = () => {
  if (timer.value !== null) {
    clearTimeout(timer.value)
    timer.value = null
  }
}

watch(
  () => deleteConfirm.value,
  (active) => {
    clearTimer()
    if (!active) return

    const tick = () => {
      if (!deleteConfirm.value) return
      if (countdown.value <= 0) {
        deleteConfirm.value = false
        countdown.value = 5
        return
      }
      countdown.value -= 1
      timer.value = window.setTimeout(tick, 1000)
    }

    timer.value = window.setTimeout(tick, 1000)
  },
)

const handlePrint = () => {
  const text = props.modelValue.trim()
  if (!text) return
  emit('print', text, stampEnabled.value)
  emit('update:modelValue', '')
  stampEnabled.value = false
}

const handleDeleteClick = () => {
  if (deleteConfirm.value) {
    emit('clear-all')
    deleteConfirm.value = false
    countdown.value = 5
    return
  }

  deleteConfirm.value = true
  countdown.value = 5
}

onBeforeUnmount(() => {
  clearTimer()
})

const getPaperOrigin = () => {
  const slot = paperSlotRef.value
  if (!slot) return null
  const rect = slot.getBoundingClientRect()
  return {
    x: rect.left + rect.width / 2,
    y: rect.top + Math.min(12, rect.height / 2),
  }
}

defineExpose({
  getPaperOrigin,
})
</script>

<style scoped>
.console-shell {
  background: linear-gradient(
    to bottom,
    var(--zone-console-shell-from),
    var(--zone-console-shell-to)
  );
  border-color: color-mix(in oklab, var(--zone-grid-color) 75%, white);
}

.console-paper-slot {
  background: var(--zone-console-screen-bg);
  border-color: color-mix(in oklab, var(--zone-console-screen-bg) 75%, white);
}

.console-panel {
  background: var(--zone-console-panel);
  border-color: color-mix(in oklab, var(--zone-console-shell-to) 78%, black);
  box-shadow:
    inset 0 4px 8px rgba(0, 0, 0, 0.3),
    0 2px 4px rgba(255, 255, 255, 0.2);
}

.console-status-dot {
  background: var(--zone-console-alert);
  box-shadow: 0 0 5px var(--zone-console-alert);
}

.console-label {
  color: color-mix(in oklab, var(--zone-console-screen-muted) 85%, black);
  font-family: var(--font-family-mono);
}

.console-meta {
  color: color-mix(in oklab, var(--zone-console-screen-muted) 80%, white);
}

.console-screen {
  background: var(--zone-console-screen-bg);
  border-color: color-mix(in oklab, var(--zone-grid-color) 65%, white);
  box-shadow: inset 0 0 20px rgba(0, 0, 0, 1);
}

.console-screen-meta {
  color: var(--zone-console-screen-muted);
  font-family: var(--font-family-mono);
  border-color: color-mix(in oklab, var(--zone-console-screen-muted) 35%, transparent);
}

.console-input {
  color: var(--zone-console-screen-text);
  font-family: var(--font-family-mono);
  text-shadow: 0 0 5px color-mix(in oklab, var(--zone-console-screen-text) 70%, transparent);
}

.console-input::placeholder {
  color: var(--zone-console-screen-muted);
}

.console-alert-layer {
  background: color-mix(in oklab, var(--zone-console-screen-bg) 95%, black);
}

.console-alert-text {
  color: var(--zone-console-alert);
  font-family: var(--font-family-mono);
}

.console-alert-count {
  color: color-mix(in oklab, var(--zone-console-alert) 78%, white);
  font-family: var(--font-family-mono);
}

.console-round-btn {
  background: var(--zone-console-button);
  border-color: color-mix(in oklab, var(--zone-console-button) 70%, white);
  box-shadow:
    0 4px 0 rgba(0, 0, 0, 0.88),
    0 5px 10px rgba(0, 0, 0, 0.5);
}

.console-print-btn {
  background: var(--zone-console-print-button);
  border-color: color-mix(in oklab, var(--zone-console-print-button) 70%, white);
  box-shadow:
    0 5px 0 rgba(80, 34, 10, 0.8),
    0 8px 15px rgba(0, 0, 0, 0.4);
}

.console-print-label {
  color: var(--zone-console-print-text);
  font-family: var(--font-family-mono);
}

.console-badge {
  background: color-mix(in oklab, var(--zone-console-screen-bg) 85%, black);
  color: var(--zone-console-screen-text);
  border-color: color-mix(in oklab, var(--zone-console-screen-text) 30%, transparent);
}

.console-bottom-glow {
  background: color-mix(in oklab, var(--zone-console-shell-from) 40%, transparent);
}

.console-divider-line {
  background: color-mix(in oklab, var(--zone-console-screen-muted) 20%, transparent);
}
</style>
