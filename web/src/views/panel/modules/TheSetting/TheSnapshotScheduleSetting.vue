<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full space-y-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <p class="text-[var(--color-text-secondary)] text-sm">
        {{ t('snapshotScheduleSetting.description') }}
      </p>
      <BaseEditCapsule
        :editing="scheduleEditMode"
        :apply-title="t('commonUi.apply')"
        :cancel-title="t('commonUi.cancel')"
        :edit-title="t('commonUi.edit')"
        @apply="handleUpdateSnapshotSchedule"
        @toggle="scheduleEditMode = !scheduleEditMode"
      />
    </div>

    <div class="schedule-row">
      <h2 class="schedule-row__label">
        {{ t('snapshotScheduleSetting.enableAutoSnapshot') }}
      </h2>
      <div class="schedule-row__control">
        <BaseSwitch v-model="SnapshotSchedule.enable" :disabled="!scheduleEditMode" />
      </div>
    </div>

    <div class="schedule-row schedule-row--top">
      <h2 class="schedule-row__label">
        {{ t('snapshotScheduleSetting.crontab') }}
      </h2>
      <div class="schedule-row__control">
        <div v-if="!scheduleEditMode" class="schedule-display">
          <p class="schedule-display__text">
            {{ SnapshotSchedule.cron_expression.length === 0 ? t('commonUi.none') : humanizedCron }}
          </p>
          <code
            v-if="SnapshotSchedule.cron_expression.length > 0"
            class="schedule-display__code"
            v-tooltip="SnapshotSchedule.cron_expression"
          >
            {{ SnapshotSchedule.cron_expression }}
          </code>
        </div>
        <CronScheduleEditor v-else v-model="SnapshotSchedule.cron_expression" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import CronScheduleEditor from './components/CronScheduleEditor.vue'
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { humanizeCron } from '@/utils/cron'
import { fetchUpdateSnapshotScheduleSetting } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getSnapshotSchedule } = settingStore
const { SnapshotSchedule } = storeToRefs(settingStore)

const scheduleEditMode = ref<boolean>(false)
const humanizedCron = computed(() => humanizeCron(SnapshotSchedule.value.cron_expression, t))

const handleUpdateSnapshotSchedule = async () => {
  const res = await fetchUpdateSnapshotScheduleSetting(SnapshotSchedule.value)
  if (res.code === 1) {
    theToast.success(res.msg)
  }

  scheduleEditMode.value = false
  await getSnapshotSchedule()
}

onMounted(async () => {
  await getSnapshotSchedule()
})
</script>

<style scoped>
.schedule-row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.5rem;
  color: var(--color-text-secondary);
}

.schedule-row--top {
  align-items: flex-start;
}

.schedule-row__label {
  flex: 0 0 auto;
  width: 9rem;
  font-weight: 600;
  font-size: 0.9rem;
  line-height: 1.4;
}

.schedule-row--top .schedule-row__label {
  padding-top: 0.2rem;
}

.schedule-row__control {
  flex: 1;
  min-width: 0;
}

.schedule-display {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.schedule-display__text {
  font-size: 0.9rem;
  color: var(--color-text-primary);
  line-height: 1.4;
  margin: 0;
  overflow-wrap: anywhere;
}

.schedule-display__code {
  display: inline-block;
  align-self: flex-start;
  max-width: 100%;
  padding: 0.1rem 0.5rem;
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  color: var(--color-text-muted);
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  letter-spacing: 0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (width < 640px) {
  .schedule-row {
    flex-direction: column;
    align-items: stretch;
    gap: 0.35rem;
  }

  .schedule-row__label {
    width: auto;
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }

  .schedule-row--top .schedule-row__label {
    padding-top: 0;
  }
}
</style>
