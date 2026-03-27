<template>
  <PanelCard>
    <!-- 备份计划设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('backupScheduleSetting.title') }}
        </h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="scheduleEditMode"
            :apply-title="t('commonUi.apply')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="handleUpdateBackupSchedule"
            @toggle="scheduleEditMode = !scheduleEditMode"
          />
        </div>
      </div>

      <!-- 开启自动备份 -->
      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('backupScheduleSetting.enableAutoBackup') }}:
        </h2>
        <BaseSwitch v-model="BackupSchedule.enable" :disabled="!scheduleEditMode" />
      </div>

      <!-- 备份计划表达式 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-38 w-max shrink-0 whitespace-nowrap">
          {{ t('backupScheduleSetting.crontab') }}:
        </h2>
        <span
          v-if="!scheduleEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="BackupSchedule.cron_expression"
          style="vertical-align: middle"
        >
          {{
            BackupSchedule.cron_expression.length === 0
              ? t('commonUi.none')
              : BackupSchedule.cron_expression
          }}
        </span>
        <BaseInput
          v-else
          v-model="BackupSchedule.cron_expression"
          type="text"
          :placeholder="t('backupScheduleSetting.crontabPlaceholder')"
          class="w-full py-1!"
        />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateBackupScheduleSetting } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getBackupSchedule } = settingStore
const { BackupSchedule } = storeToRefs(settingStore)

const scheduleEditMode = ref<boolean>(false)

const handleUpdateBackupSchedule = async () => {
  const res = await fetchUpdateBackupScheduleSetting(BackupSchedule.value)
  if (res.code === 1) {
    theToast.success(res.msg)
  }

  scheduleEditMode.value = false
  await getBackupSchedule()
}

onMounted(async () => {
  await getBackupSchedule()
})
</script>

<style scoped></style>
