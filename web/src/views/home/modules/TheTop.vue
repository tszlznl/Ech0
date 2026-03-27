<template>
  <div class="mx-auto mb-1 sm:mb-0">
    <div class="w-full flex justify-between items-center">
      <!-- 搜索与过滤 -->
      <div class="flex justify-start items-center gap-2">
        <BaseInput
          v-if="!isFilteringMode"
          v-tooltip="t('homeTop.searchTitle')"
          type="text"
          v-model="searchContent"
          :placeholder="t('homeTop.searchPlaceholder')"
          class="w-42! h-10 bg-[var(--input-bg-color)]"
          @keyup.enter="$event.target.blur()"
          @blur="handleSearch"
        />
        <!-- 过滤条件 -->
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelFilter"
          class="w-34 text-nowrap flex items-center justify-between px-1 py-0.5 text-[var(--color-text-muted)] border border-dashed border-[var(--color-border-strong)] rounded-md hover:cursor-pointer hover:line-through hover:text-[var(--color-text-secondary)]"
        >
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1" />
        </div>
        <button
          v-if="isZenMode"
          type="button"
          v-tooltip="t('homeTop.exitZenMode')"
          class="h-8 px-2 text-xs text-[var(--color-text-muted)] border border-[var(--color-border-subtle)] rounded-md hover:line-through hover:text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)] transition-colors duration-200"
          @click="handleExitZenMode"
        >
          Zen
        </button>
      </div>

      <!-- 右侧图标组 -->
      <div class="flex justify-end items-center gap-1">
        <!-- RSS -->
        <div>
          <a href="/rss" v-tooltip="t('homeTop.rssTitle')">
            <!-- icon -->
            <Rss class="w-8 h-8 text-[var(--color-text-muted)]" />
          </a>
        </div>
        <!-- Ech0 Widget（移动端入口） -->
        <div class="sm:hidden">
          <RouterLink to="/widget" v-tooltip="t('homeTop.widgetTitle')">
            <Widget class="w-8 h-8 text-[var(--color-text-muted)]" />
          </RouterLink>
        </div>
        <!-- Ech0 Hub -->
        <div>
          <RouterLink to="/hub" v-tooltip="t('homeTop.hubTitle')">
            <!-- icon -->
            <HubIcon class="w-8 h-8 text-[var(--color-text-muted)]" />
          </RouterLink>
        </div>
        <!-- PanelPage -->
        <div>
          <RouterLink to="/panel" v-tooltip="t('homeTop.panelTitle')">
            <!-- icon -->
            <Panel class="w-8 h-8 text-[var(--color-text-muted)]" />
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import Panel from '@/components/icons/panel.vue'
import Rss from '@/components/icons/rss.vue'
import HubIcon from '@/components/icons/hub.vue'
import Widget from '@/components/icons/widget.vue'
import { RouterLink } from 'vue-router'
import { useEchoStore, useZenStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'
const echoStore = useEchoStore()
const zenStore = useZenStore()
const { refreshForSearch, getEchosByPage } = echoStore
const { searchingMode, filteredTag, isFilteringMode } = storeToRefs(echoStore)
const { isZenMode } = storeToRefs(zenStore)
const { t } = useI18n()

const searchContent = ref<string>('')

const handleSearch = () => {
  // 设置搜索内容

  echoStore.searchValue = searchContent.value

  // 判断是否是搜索模式
  if (searchingMode.value) {
    // 初始化搜索
    refreshForSearch()
    // 开始搜索
    getEchosByPage()
  }
}

const handleCancelFilter = () => {
  echoStore.isFilteringMode = false
  echoStore.filteredTag = null
  echoStore.refreshEchosForFilter()
}

const handleExitZenMode = () => {
  zenStore.setZenMode(false)
}
</script>
