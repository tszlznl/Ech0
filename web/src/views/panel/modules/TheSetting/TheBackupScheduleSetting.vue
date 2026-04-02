<template>
  <PanelCard>
    <div class="w-full space-y-3">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="space-y-1">
          <h1 class="text-[var(--color-text-primary)] font-bold text-lg leading-none">
            {{ t('backupScheduleSetting.title') }}
          </h1>
          <p class="text-[var(--color-text-secondary)] text-sm">
            {{ t('backupScheduleSetting.description') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <BaseButton
            @click="handleCreateSnapshot"
            :disabled="isSnapshotCreating"
            class="px-3 py-1.5 text-sm! rounded-[var(--radius-md)]"
            :tooltip="t('backupScheduleSetting.createSnapshot')"
          >
            {{
              isSnapshotCreating
                ? t('backupScheduleSetting.creatingSnapshot')
                : t('backupScheduleSetting.createSnapshot')
            }}
          </BaseButton>
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

      <div
        class="flex flex-wrap items-center justify-start text-[var(--color-text-secondary)] gap-2 min-h-10"
      >
        <h2 class="font-semibold w-36 shrink-0">
          {{ t('backupScheduleSetting.enableAutoBackup') }}
        </h2>
        <BaseSwitch v-model="BackupSchedule.enable" :disabled="!scheduleEditMode" />
      </div>

      <div
        class="flex flex-wrap items-center justify-start text-[var(--color-text-secondary)] gap-2 min-h-10"
      >
        <h2 class="font-semibold w-36 shrink-0">
          {{ t('backupScheduleSetting.crontab') }}
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
import BaseButton from '@/components/common/BaseButton.vue'
import { computed, ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateBackupScheduleSetting } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getBackupSchedule, startSnapshotTask, restoreSnapshotTaskFromStorage } = settingStore
const { BackupSchedule, snapshotStatus, snapshotError } = storeToRefs(settingStore)

const scheduleEditMode = ref<boolean>(false)
const isSnapshotCreating = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
)

const handleUpdateBackupSchedule = async () => {
  const res = await fetchUpdateBackupScheduleSetting(BackupSchedule.value)
  if (res.code === 1) {
    theToast.success(res.msg)
  }

  scheduleEditMode.value = false
  await getBackupSchedule()
}

const handleCreateSnapshot = async () => {
  if (isSnapshotCreating.value) return
  try {
    const res = await startSnapshotTask()
    if (!res) return
    if (res.code === 1) {
      theToast.success(res.msg || String(t('backupScheduleSetting.creatingSnapshot')))
      return
    }

    theToast.error(res.msg || String(t('backupScheduleSetting.createSnapshotFailed')))
  } catch (error) {
    console.error(String(t('backupScheduleSetting.createSnapshotFailed')), error)
    theToast.error(String(t('backupScheduleSetting.createSnapshotFailed')))
  }
}

watch(
  () => snapshotStatus.value,
  (status, prevStatus) => {
    if (status === prevStatus) return
    if (status === 'success') {
      theToast.success(String(t('backupScheduleSetting.createSnapshotSuccess')))
      return
    }
    if (status === 'failed') {
      theToast.error(snapshotError.value || String(t('backupScheduleSetting.createSnapshotFailed')))
    }
  },
)

onMounted(async () => {
  await getBackupSchedule()
  await restoreSnapshotTaskFromStorage()
})
</script>

<style scoped></style>
