<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="export-wrap">
    <div class="export-header">
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
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import { computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'
import { useSettingStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { fetchDownloadExport } from '@/service/api'

const { t } = useI18n()
const settingStore = useSettingStore()
const userStore = useUserStore()
const { startSnapshotTask, restoreSnapshotTask } = settingStore
const { snapshotStatus, snapshotError } = storeToRefs(settingStore)
const { isLogin } = storeToRefs(userStore)

// 导出 = 异步 export 作业产出快照（与导入对称：重活在 job 里，有进度/可取消），完成后取回产物。
const isExporting = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
)

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
</style>
