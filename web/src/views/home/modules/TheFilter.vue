<template>
  <div class="w-full px-2 mb-1 sm:mb-0">
    <div class="w-full flex flex-col gap-3">
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
        </div>
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelFilter"
          class="w-34 text-nowrap flex items-center justify-between px-1 py-0.5 text-[var(--color-text-muted)] border border-dashed border-[var(--color-border-strong)] rounded-md hover:cursor-pointer hover:line-through hover:text-[var(--color-text-secondary)]"
        >
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import { useEchoStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'

const echoStore = useEchoStore()
const { refreshForSearch, getEchosByPage } = echoStore
const { searchingMode, filteredTag, isFilteringMode, searchValue } = storeToRefs(echoStore)
const { t } = useI18n()

const searchContent = ref<string>(searchValue.value)

const handleSearch = () => {
  echoStore.searchValue = searchContent.value
  if (searchingMode.value) {
    refreshForSearch()
    getEchosByPage()
  }
}

const handleCancelFilter = () => {
  echoStore.isFilteringMode = false
}

watch(searchValue, (value) => {
  if (value !== searchContent.value) {
    searchContent.value = value
  }
})
</script>

<style scoped>
.home-filter__search-shell {
  width: 100%;
  padding: 0.3rem;
  border-radius: var(--radius-xs);
  background: var(--filter-search-shell-bg);
  box-shadow: inset 0 0 0 1px var(--color-border-subtle);
}
</style>
