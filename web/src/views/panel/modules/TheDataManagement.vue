<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <!-- 分段控件：导入 / 导出 / 快照 -->
      <div class="seg" role="tablist">
        <button
          type="button"
          role="tab"
          :aria-selected="tab === 'import'"
          class="seg__btn"
          :class="{ 'seg__btn--active': tab === 'import' }"
          @click="tab = 'import'"
        >
          {{ t('dataManagement.tabImport') }}
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="tab === 'export'"
          class="seg__btn"
          :class="{ 'seg__btn--active': tab === 'export' }"
          @click="tab = 'export'"
        >
          {{ t('dataManagement.tabExport') }}
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="tab === 'snapshot'"
          class="seg__btn"
          :class="{ 'seg__btn--active': tab === 'snapshot' }"
          @click="tab = 'snapshot'"
        >
          {{ t('dataManagement.tabSnapshot') }}
        </button>
      </div>

      <!-- 内容 -->
      <div class="seg-content">
        <TheMigrationSetting v-if="tab === 'import'" />
        <TheExportSetting v-else-if="tab === 'export'" />
        <TheSnapshotScheduleSetting v-else />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PanelCard from '@/layout/PanelCard.vue'
import TheMigrationSetting from './TheSetting/TheMigrationSetting.vue'
import TheExportSetting from './TheSetting/TheExportSetting.vue'
import TheSnapshotScheduleSetting from './TheSetting/TheSnapshotScheduleSetting.vue'

const { t } = useI18n()
const tab = ref<'import' | 'export' | 'snapshot'>('import')
</script>

<style scoped>
.seg {
  display: inline-flex;
  gap: 0.25rem;
  padding: 0.25rem;
  margin-bottom: 1rem;
  background: var(--color-bg-muted);
  border-radius: var(--radius-md);
}

.seg__btn {
  padding: 0.35rem 1.1rem;
  font-size: 0.85rem;
  font-weight: 600;
  line-height: 1.2;
  color: var(--color-text-secondary);
  border-radius: var(--radius-sm);
  transition:
    background 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.seg__btn:hover:not(.seg__btn--active) {
  color: var(--color-text-primary);
}

.seg__btn--active {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

@media (width < 640px) {
  .seg {
    display: flex;
    width: 100%;
  }

  .seg__btn {
    flex: 1;
    text-align: center;
    padding: 0.4rem 0.5rem;
  }
}
</style>
