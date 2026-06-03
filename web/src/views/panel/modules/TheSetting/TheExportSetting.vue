<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="export-wrap">
    <div class="export-header">
      <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
        {{ t('exportSetting.title') }}
      </h1>
      <p class="export-desc">{{ t('exportSetting.description') }}</p>
    </div>

    <div class="export-action">
      <BaseButton
        @click="handleExport"
        :disabled="isExporting"
        class="export-download-btn"
        :tooltip="t('exportSetting.exportSnapshot')"
      >
        {{ isExporting ? t('exportSetting.exporting') : t('exportSetting.exportSnapshot') }}
      </BaseButton>
    </div>

    <JobProgressCard
      v-if="snapshotStatus !== 'idle'"
      :title="t('exportSetting.jobTitle')"
      :status="snapshotStatus"
      :status-label="statusLabelMap[snapshotStatus] || snapshotStatus"
      :steps="exportSteps"
      :current-key="exportCurrentKey"
      :error-message="snapshotStatus === 'failed' ? snapshotError : ''"
    >
      <template v-if="snapshotStatus === 'success'" #footer>
        <div class="export-artifact">
          <span class="export-artifact__label">{{ t('exportSetting.artifactLabel') }}</span>
          <span class="export-artifact__name" v-tooltip="snapshotFileName">
            {{ snapshotFileName || '—' }}
          </span>
          <span class="export-artifact__size">{{ formatBytes(snapshotSize) }}</span>
          <BaseButton :tooltip="t('exportSetting.redownload')" @click="downloadSnapshot">
            {{ t('exportSetting.redownload') }}
          </BaseButton>
        </div>
      </template>
    </JobProgressCard>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import JobProgressCard from './components/JobProgressCard.vue'
import { computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'
import { useSettingStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { fetchDownloadExport } from '@/service/api'
import { formatBytes } from '@/utils/file'

const { t } = useI18n()
const settingStore = useSettingStore()
const userStore = useUserStore()
const { startSnapshotTask, restoreSnapshotTask } = settingStore
const { snapshotStatus, snapshotError, snapshotPhase, snapshotFileName, snapshotSize } =
  storeToRefs(settingStore)
const { isLogin } = storeToRefs(userStore)

// 导出 = 异步 export 作业产出快照（与导入对称：重活在 job 里，有进度/可取消），完成后取回产物。
const isExporting = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
)

// 步进器对齐后端真实阶段：排队(pending) → 打包(packing) → 完成(completed)。
const exportSteps = computed(() => [
  { key: 'pending', label: String(t('jobProgress.exportPhasePending')) },
  { key: 'packing', label: String(t('jobProgress.exportPhasePacking')) },
  { key: 'completed', label: String(t('jobProgress.exportPhaseCompleted')) },
])

const exportCurrentKey = computed(() => {
  if (snapshotStatus.value === 'pending') return 'pending'
  if (snapshotStatus.value === 'success') return 'completed'
  return snapshotPhase.value || 'packing'
})

const statusLabelMap = computed<Record<string, string>>(() => ({
  idle: String(t('jobProgress.statusIdle')),
  pending: String(t('jobProgress.statusPending')),
  running: String(t('jobProgress.statusRunning')),
  success: String(t('jobProgress.statusSuccess')),
  failed: String(t('jobProgress.statusFailed')),
  cancelled: String(t('jobProgress.statusCancelled')),
}))

// 鉴权下载：经 fetchDownloadExport（credentials + Authorization header）取回 blob 触发下载，
// token 不进 URL（避免出现在浏览器历史/日志/Referer）。
const downloadSnapshot = async () => {
  try {
    const blob = await fetchDownloadExport()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `ech0-snapshot-${Date.now()}.zip`
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error(String(t('exportSetting.exportFailed')), error)
    theToast.error(String(t('exportSetting.exportFailed')))
  }
}

const handleExport = async () => {
  if (!isLogin.value) {
    theToast.info(String(t('exportSetting.loginRequired')), { duration: 3000 })
    return
  }
  if (isExporting.value) return
  try {
    theToast.info(String(t('exportSetting.exporting')), { duration: 4000 })
    const res = await startSnapshotTask()
    if (!res) return
    if (res.code !== 1) {
      theToast.error(res.msg || String(t('exportSetting.exportFailed')))
    }
    // 作业完成后由下方 watch(snapshotStatus) 自动触发下载。
  } catch (error) {
    console.error(String(t('exportSetting.exportFailed')), error)
    theToast.error(String(t('exportSetting.exportFailed')))
  }
}

watch(
  () => snapshotStatus.value,
  (status, prevStatus) => {
    if (status === prevStatus) return
    if (status === 'success') {
      theToast.success(String(t('exportSetting.exportStarted')))
      void downloadSnapshot()
      return
    }
    if (status === 'failed') {
      theToast.error(snapshotError.value || String(t('exportSetting.exportFailed')))
    }
  },
)

onMounted(() => {
  // 进入页面时若有进行中的导出作业则接管轮询（完成会自动下载）。
  void restoreSnapshotTask()
})
</script>

<style scoped>
.export-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.export-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.export-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.export-action {
  display: flex;
  align-items: center;
}

.export-download-btn {
  border-radius: var(--radius-md);
  color: var(--color-text-primary) !important;
}

.export-artifact {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.82rem;
}

.export-artifact__label {
  color: var(--color-text-muted);
}

.export-artifact__name {
  max-width: 16rem;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-artifact__size {
  padding: 0.05rem 0.4rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  font-variant-numeric: tabular-nums;
}
</style>
