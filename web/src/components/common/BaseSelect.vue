<template>
  <div class="base-select" ref="selectRef">
    <!-- Label -->
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--select-label-text-color)] mb-1"
    >
      {{ label }}
    </label>

    <!-- Select Button -->
    <div class="relative inline-block" ref="triggerRef">
      <button
        :id="id"
        type="button"
        :disabled="disabled"
        :class="[
          'inline-flex items-center justify-between px-3 py-2 rounded-[var(--radius-md)] border border-[var(--select-border-color)] focus:outline-none focus:ring-2 focus:ring-[var(--select-focus-ring-color)] transition duration-150 ease-in-out shadow-[var(--shadow-sm)] sm:text-sm text-left',
          disabled
            ? 'bg-[var(--select-disabled-bg-color)] cursor-not-allowed opacity-70'
            : 'bg-[var(--select-bg-color)] hover:border-[var(--color-border-strong)] cursor-pointer',
          customClass,
        ]"
        @click="onToggle"
        @keydown.space.prevent="onToggle"
        @keydown.enter.prevent="onToggle"
        @keydown.up.prevent="onNavigate(-1)"
        @keydown.down.prevent="onNavigate(1)"
        @keydown.escape="onClose"
      >
        <!-- Selected Value Display -->
        <span
          :class="[
            'truncate',
            !selectedOption && placeholder
              ? 'text-[var(--color-text-muted)]'
              : 'text-[var(--color-text-secondary)]',
          ]"
        >
          {{ displayValue }}
        </span>

        <!-- Dropdown Arrow -->
        <svg
          :class="[
            'w-8 text-[var(--select-icon-color)] transition-transform duration-200',
            isOpen ? 'rotate-180' : '',
          ]"
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
        >
          <!-- Icon from Material Symbols by Google - https://github.com/google/material-design-icons/blob/master/LICENSE -->
          <path fill="currentColor" d="m12 15.4l-6-6L7.4 8l4.6 4.6L16.6 8L18 9.4z" />
        </svg>
      </button>

      <!-- Dropdown Menu -->
      <Teleport to="body">
        <Transition
          enter-active-class="transition ease-out duration-100"
          enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100"
          leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100"
          leave-to-class="transform opacity-0 scale-95"
        >
          <div
            v-show="isOpen"
            ref="dropdownRef"
            class="fixed z-5000 bg-[var(--select-bg-color)] shadow-[var(--shadow-md)] max-h-70 rounded-[var(--radius-md)] border border-[var(--select-border-color)] overflow-auto focus:outline-none"
            :style="dropdownStyle"
            @wheel.stop
          >
            <div
              v-for="(option, index) in normalizedOptions"
              :key="String(getOptionValue(option) ?? index)"
              :class="[
                'cursor-pointer select-none relative px-3 py-2 text-sm',
                index === highlightedIndex
                  ? 'bg-[var(--select-label-hover-bg-color)] text-[var(--select-option-active-color)]'
                  : 'text-[var(--color-text-primary)] hover:bg-[var(--select-label-clicked-bg-color)]',
                isSelected(option) ? 'font-medium' : 'font-normal',
              ]"
              @click="onSelect(option)"
              @mouseenter="highlightedIndex = index"
            >
              <div class="flex items-center justify-between">
                <span class="truncate text-[var(--color-text-muted)] font-bold">{{
                  getOptionLabel(option)
                }}</span>
                <!-- Check Icon for Selected -->
                <svg
                  v-if="isSelected(option)"
                  class="text-[var(--color-accent)]"
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                >
                  <!-- Icon from Typicons by Stephen Hutchings - https://creativecommons.org/licenses/by-sa/4.0/ -->
                  <path
                    fill="currentColor"
                    d="M16.972 6.251a2 2 0 0 0-2.72.777l-3.713 6.682l-2.125-2.125a2 2 0 1 0-2.828 2.828l4 4c.378.379.888.587 1.414.587l.277-.02a2 2 0 0 0 1.471-1.009l5-9a2 2 0 0 0-.776-2.72"
                  />
                </svg>
              </div>
            </div>

            <!-- Empty State -->
            <div
              v-if="normalizedOptions.length === 0"
              class="px-3 py-2 text-sm text-[var(--color-text-muted)] text-center"
            >
              {{ emptyText }}
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

// 定义值的类型
type SelectValue = string | number | boolean | null | undefined

// 定义选项接口
interface SelectOption {
  label: string
  value: SelectValue
  disabled?: boolean
}

// 定义通用对象类型用于自定义键名
interface CustomKeyOption {
  [key: string]: unknown
}

// 定义选项类型联合
type OptionType = SelectOption | string | number | CustomKeyOption

const props = defineProps<{
  modelValue: SelectValue
  options: OptionType[]
  id?: string
  label?: string
  placeholder?: string
  disabled?: boolean
  class?: string
  emptyText?: string
  labelKey?: string
  valueKey?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: SelectValue): void
  (e: 'change', value: SelectValue): void
  (e: 'open'): void
  (e: 'close'): void
}>()
const { t } = useI18n()

// Refs
const selectRef = ref<HTMLElement>()
const triggerRef = ref<HTMLElement>()
const dropdownRef = ref<HTMLElement>()
const isOpen = ref(false)
const highlightedIndex = ref(-1)
const dropdownStyle = ref<Record<string, string>>({})

// Computed
const customClass = props.class
const emptyText = computed(() => props.emptyText || String(t('baseSelect.empty')))

const normalizedOptions = computed((): SelectOption[] => {
  return props.options.map((option): SelectOption => {
    if (typeof option === 'string' || typeof option === 'number') {
      return { label: String(option), value: option }
    }

    // 如果是对象类型，检查是否已经是 SelectOption 格式
    if (typeof option === 'object' && option !== null) {
      const objOption = option as Record<string, unknown>

      // 如果有自定义键名
      if (props.labelKey && props.valueKey) {
        return {
          label: String(objOption[props.labelKey] || ''),
          value: objOption[props.valueKey] as SelectValue,
          disabled: objOption.disabled as boolean | undefined,
        }
      }

      // 如果是标准的 SelectOption 格式
      if ('label' in objOption && 'value' in objOption) {
        return option as SelectOption
      }
    }

    // 兜底情况
    return { label: String(option), value: option as unknown as SelectValue }
  })
})

const selectedOption = computed(() => {
  return normalizedOptions.value.find((option) => option.value === props.modelValue)
})

const displayValue = computed(() => {
  return selectedOption.value?.label || props.placeholder || String(t('baseSelect.pleaseSelect'))
})

// Methods
function getOptionLabel(option: SelectOption): string {
  return option.label
}

function getOptionValue(option: SelectOption): SelectValue {
  return option.value
}

function isSelected(option: SelectOption): boolean {
  return getOptionValue(option) === props.modelValue
}

function onToggle(): void {
  if (props.disabled) return

  if (isOpen.value) {
    onClose()
  } else {
    onOpen()
  }
}

function onOpen(): void {
  isOpen.value = true
  highlightedIndex.value = normalizedOptions.value.findIndex((option) => isSelected(option))
  void nextTick(() => {
    updateDropdownPosition()
  })
  emit('open')
}

function onClose(): void {
  isOpen.value = false
  highlightedIndex.value = -1
  emit('close')
}

function onSelect(option: SelectOption): void {
  if (option.disabled) return

  const value = getOptionValue(option)
  emit('update:modelValue', value)
  emit('change', value)
  onClose()
}

function onNavigate(direction: number): void {
  if (!isOpen.value) {
    onOpen()
    return
  }

  const optionsCount = normalizedOptions.value.length
  if (optionsCount === 0) return

  let newIndex = highlightedIndex.value + direction

  if (newIndex < 0) {
    newIndex = optionsCount - 1
  } else if (newIndex >= optionsCount) {
    newIndex = 0
  }

  highlightedIndex.value = newIndex
}

// Handle clicks outside to close dropdown
function handleClickOutside(event: Event): void {
  if (!(event.target instanceof Node)) return

  const clickedInsideSelect = !!selectRef.value?.contains(event.target)
  const clickedInsideDropdown = !!dropdownRef.value?.contains(event.target)
  if (!clickedInsideSelect && !clickedInsideDropdown) {
    onClose()
  }
}

function updateDropdownPosition(): void {
  if (!isOpen.value || !triggerRef.value) return

  const triggerRect = triggerRef.value.getBoundingClientRect()
  const gap = 4
  const menuMaxHeight = 280
  const viewportBottom = window.innerHeight
  const viewportRight = window.innerWidth
  const availableAbove = Math.max(0, triggerRect.top - gap)
  const availableBelow = Math.max(0, viewportBottom - triggerRect.bottom - gap)
  const measuredHeight = dropdownRef.value?.offsetHeight || menuMaxHeight
  const shouldOpenUpward =
    availableBelow < Math.min(menuMaxHeight, 160) && availableAbove > availableBelow
  const maxHeight = shouldOpenUpward
    ? Math.min(menuMaxHeight, availableAbove)
    : Math.min(menuMaxHeight, availableBelow)
  const top = shouldOpenUpward
    ? Math.max(gap, triggerRect.top - gap - Math.min(measuredHeight, maxHeight))
    : triggerRect.bottom + gap
  const left = Math.min(
    Math.max(gap, triggerRect.left),
    Math.max(gap, viewportRight - triggerRect.width - gap),
  )

  dropdownStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    minWidth: `${triggerRect.width}px`,
    maxHeight: `${Math.max(120, maxHeight)}px`,
  }
}

// Lifecycle
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', updateDropdownPosition)
  document.addEventListener('scroll', updateDropdownPosition, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updateDropdownPosition)
  document.removeEventListener('scroll', updateDropdownPosition, true)
})
</script>
