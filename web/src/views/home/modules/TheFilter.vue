<template>
  <div class="w-full sm:px-2 mb-1 sm:mb-0">
    <div class="w-full flex flex-col gap-2">
      <div class="flex justify-start items-center gap-2 w-full flex-wrap">
        <div v-if="!isFilteringMode" class="home-filter__search-shell">
          <BaseInput
            v-tooltip="t('homeTop.searchTitle')"
            type="text"
            v-model="searchContent"
            :placeholder="t('homeTop.searchPlaceholder')"
            class="h-9 w-full! max-w-none rounded-[var(--radius-xs)]! border-[var(--filter-search-input-border)]! bg-transparent! shadow-none! text-[var(--color-text-secondary)] focus:ring-0!"
            @keyup.enter="($event.target as HTMLInputElement).blur()"
            @blur="handleSearch"
          />
          <button
            type="button"
            class="home-filter__kbd-hint"
            :aria-label="t('commandPalette.title')"
            v-tooltip="t('commandPalette.title')"
            @click="emit('openPalette')"
          >
            {{ shortcutBadge }}
          </button>
        </div>
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelTapFilter"
          class="home-filter__chip"
        >
          <span class="home-filter__chip-hash">#</span>
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
      </div>

      <div v-if="isDateRangeActive || selectedTagChips.length > 0" class="flex flex-wrap gap-1.5">
        <div
          v-if="isDateRangeActive"
          class="home-filter__chip"
          v-tooltip="t('commandPalette.reset')"
          @click="handleClearDateRange"
        >
          <p class="text-nowrap truncate">
            {{ t('commandPalette.activeChipDatePrefix') }} · {{ dateRangeSummary }}
          </p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
        <div
          v-for="chip in selectedTagChips"
          :key="chip.id"
          class="home-filter__chip"
          @click="handleRemoveTag(chip.id)"
        >
          <span class="home-filter__chip-hash">#</span>
          <p class="text-nowrap truncate">{{ chip.name }}</p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import { useEchoStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'

const emit = defineEmits<{
  (e: 'openPalette'): void
}>()

const echoStore = useEchoStore()
const {
  refreshForSearch,
  fetchCurrentPage,
  refreshEchos,
  resetDateRange,
  removeSelectedTag,
  ensureTagsLoaded,
} = echoStore
const {
  searchingMode,
  filteredTag,
  isFilteringMode,
  searchValue,
  dateFrom,
  dateTo,
  isDateRangeActive,
  selectedTagIds,
  tagList,
} = storeToRefs(echoStore)
const { t } = useI18n()

const searchContent = ref<string>(searchValue.value)

const isMac =
  typeof navigator !== 'undefined' && /Mac|iPhone|iPad|iPod/.test(navigator.platform || '')
const shortcutBadge = computed(() => (isMac ? '⌘K' : 'Ctrl+K'))

const formatDate = (sec: number | null): string => {
  if (sec === null) return '…'
  const d = new Date(sec * 1000)
  if (Number.isNaN(d.getTime())) return '…'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const dateRangeSummary = computed(() => {
  const from = formatDate(dateFrom.value)
  const to = formatDate(dateTo.value)
  return `${from} ${t('commandPalette.rangeSeparator')} ${to}`
})

const selectedTagChips = computed(() => {
  if (!selectedTagIds.value.length) return []
  const byId = new Map(tagList.value.map((tag) => [tag.id, tag.name]))
  return selectedTagIds.value.map((id) => ({ id, name: byId.get(id) ?? id }))
})

const handleSearch = () => {
  echoStore.searchValue = searchContent.value
  if (searchingMode.value) {
    refreshForSearch()
    fetchCurrentPage()
  }
}

const handleCancelTapFilter = () => {
  echoStore.isFilteringMode = false
}

const handleClearDateRange = () => {
  resetDateRange()
  refreshEchos()
}

const handleRemoveTag = (id: string) => {
  removeSelectedTag(id)
  refreshEchos()
}

watch(searchValue, (value) => {
  if (value !== searchContent.value) {
    searchContent.value = value
  }
})

// 有 selectedTagIds 时确保 tag 元数据已载入（用于显示 chip 名字）
onMounted(() => {
  if (selectedTagIds.value.length > 0) {
    void ensureTagsLoaded()
  }
})
watch(selectedTagIds, (ids) => {
  if (ids.length > 0 && tagList.value.length === 0) {
    void ensureTagsLoaded()
  }
})
</script>

<style scoped>
.home-filter__search-shell {
  position: relative;
  width: 100%;
  padding: 0.3rem;

  /* 留出 kbd 徽章所需的右内边距，避免输入内容与按钮视觉重叠 */
  padding-right: 3.25rem;
  border-radius: var(--radius-xs);
  background: var(--filter-search-shell-bg);
  box-shadow: inset 0 0 0 1px var(--color-border-subtle);
}

/* kbd 徽章：扩大可点击触摸区域（≥32×28），同时保证 AA 级对比度 */
.home-filter__kbd-hint {
  position: absolute;
  top: 50%;
  right: 0.4rem;
  transform: translateY(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.75rem;
  min-width: 2.25rem;
  padding: 0.25rem 0.5rem;
  font-family: var(--font-family-mono, monospace);
  font-size: 0.72rem;
  font-weight: 500;
  line-height: 1;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
  border-bottom-width: 2px;
  border-radius: 6px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background 0.15s ease,
    transform 0.08s ease;
}

.home-filter__kbd-hint:hover {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
}

.home-filter__kbd-hint:active {
  transform: translateY(calc(-50% + 1px));
}

.home-filter__chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 0.125rem 0.5rem;
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border-strong);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease,
    text-decoration-color 0.15s ease;
}

.home-filter__chip:hover {
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  text-decoration: line-through;
}

.home-filter__chip-hash {
  color: inherit;
  opacity: 0.6;
  margin-right: 0.15rem;
  font-weight: 500;
}
</style>
