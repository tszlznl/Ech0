<template>
  <PanelCard>
    <div class="migration-wrap">
      <div class="migration-header">
        <h1 class="migration-title">数据导入</h1>
        <p class="migration-desc">支持从 Ech0 v4、Ech0 v3及以下、Memos 导入数据。</p>
      </div>

      <div class="migration-source-grid">
        <button
          v-for="source in sourceCards"
          :key="source.value"
          class="migration-source-card"
          :class="{ active: sourceType === source.value }"
          @click="handleSelectSource(source.value)"
        >
          <h3>{{ source.title }}</h3>
          <p>{{ source.desc }}</p>
        </button>
      </div>

      <div class="migration-form">
        <div class="migration-row migration-row-top">
          <span class="migration-label">来源压缩包</span>
          <div class="migration-upload-wrap">
            <BaseButton title="选择 zip 文件" @click="handlePickZip">选择 zip 文件</BaseButton>
            <p class="migration-file-name">{{ selectedZipName || '未选择文件' }}</p>
          </div>
        </div>
      </div>

      <div class="migration-actions">
        <BaseButton title="开始迁移" @click="handleStartMigration">开始迁移</BaseButton>
        <BaseButton title="刷新状态" @click="handleRefreshJob">刷新状态</BaseButton>
        <BaseButton title="取消任务" @click="handleCancelJob">取消任务</BaseButton>
        <BaseButton
          v-if="migrationStore.canCleanup"
          :title="migrationStore.isSuccess ? '完成' : '结束/清理迁移'"
          @click="handleCleanupMigration"
        >
          {{ migrationStore.isSuccess ? '完成' : '结束/清理迁移' }}
        </BaseButton>
      </div>

      <div class="migration-job" v-if="migrationStore.hasJob">
        <div class="migration-job-header">
          <div class="migration-job-title-wrap">
            <h3 class="migration-job-title">迁移任务</h3>
            <p class="migration-job-subtitle">
              来源 {{ sourceLabelMap[migrationStore.state.source_type] || migrationStore.state.source_type }}
            </p>
          </div>
          <span class="migration-status-pill" :class="`status-${migrationStore.state.status}`">
            {{ statusLabelMap[migrationStore.state.status] || migrationStore.state.status }}
          </span>
        </div>

        <p class="migration-job-error" v-if="migrationStore.state.error_message">
          {{ migrationStore.state.error_message }}
        </p>

        <div class="migration-job-metrics" v-if="hasMetrics">
          <div class="metric-item">
            <span class="metric-label">总处理</span>
            <span class="metric-value">{{ migrationProcessed }}</span>
          </div>
          <div class="metric-item">
            <span class="metric-label">成功</span>
            <span class="metric-value">{{ migrationSuccess }}</span>
          </div>
          <div class="metric-item">
            <span class="metric-label">失败</span>
            <span class="metric-value">{{ migrationFail }}</span>
          </div>
        </div>

        <div class="migration-job-meta">
          <p v-if="migrationJobId">任务ID: {{ migrationJobId }}</p>
          <p v-if="formattedStartedAt">开始时间: {{ formattedStartedAt }}</p>
          <p v-if="formattedFinishedAt">结束时间: {{ formattedFinishedAt }}</p>
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { fetchUploadMigrationSourceZip } from '@/service/api'
import { useMigrationStore } from '@/stores'
import { theToast } from '@/utils/toast'

type MigrationSourceType = 'ech0_v4' | 'ech0_v3' | 'memos'

const sourceCards = [
  { value: 'ech0_v4', title: 'Ech0', desc: '支持最新版 Ech0 v4 及以后' },
  { value: 'ech0_v3', title: 'Ech0 v3', desc: '支持 Ech0 v3及更早版本' },
  { value: 'memos', title: 'Memos', desc: '支持 Memos' },
]

const sourceType = ref<MigrationSourceType>('ech0_v4')
const selectedZip = ref<File | null>(null)
const selectedZipName = ref('')
const migrationStore = useMigrationStore()
const statusLabelMap: Record<string, string> = {
  idle: '空闲',
  pending: '等待中',
  running: '迁移中',
  success: '已完成',
  failed: '失败',
  cancelled: '已取消',
}
const sourceLabelMap: Record<string, string> = {
  ech0_v4: 'Ech0',
  ech0_v3: 'Ech0 v3',
  memos: 'Memos',
}
const migrationReport = computed(
  () => (migrationStore.state.source_payload?.report as Record<string, unknown> | undefined) ?? {},
)
const migrationJobId = computed(
  () =>
    (migrationStore.state.source_payload?.migration_job_id as string | undefined) ||
    (migrationReport.value.job_id as string | undefined),
)
const migrationProcessed = computed(() => migrationReport.value.processed)
const migrationSuccess = computed(() => migrationReport.value.success_count)
const migrationFail = computed(() => migrationReport.value.fail_count)
const hasMetrics = computed(
  () =>
    migrationProcessed.value !== undefined ||
    migrationSuccess.value !== undefined ||
    migrationFail.value !== undefined,
)
const formattedStartedAt = computed(() => formatTime(migrationStore.state.started_at))
const formattedFinishedAt = computed(() => formatTime(migrationStore.state.finished_at))

const resetSelectedZip = () => {
  selectedZip.value = null
  selectedZipName.value = ''
}

const handleSelectSource = (value: string) => {
  sourceType.value = value as MigrationSourceType
  resetSelectedZip()
}

const handlePickZip = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.zip,application/zip'
  input.onchange = (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return
    if (!file.name.toLowerCase().endsWith('.zip')) {
      theToast.error('仅支持上传 zip 文件')
      return
    }
    selectedZip.value = file
    selectedZipName.value = file.name
  }
  input.click()
}

const handleStartMigration = async () => {
  if (!selectedZip.value) {
    theToast.info('请先选择 zip 文件')
    return
  }

  const uploadRes = await fetchUploadMigrationSourceZip(sourceType.value, selectedZip.value)
  if (uploadRes.code !== 1) {
    theToast.error(uploadRes.msg || '上传迁移压缩包失败')
    return
  }
  const sourcePayload = uploadRes.data?.source_payload ?? {}

  const res = await migrationStore.startMigration({
    source_type: sourceType.value,
    source_payload: sourcePayload,
  })
  if (res.code !== 1) {
    theToast.error(res.msg || '创建迁移任务失败')
    return
  }
  theToast.success('迁移已开始')
}

const handleRefreshJob = async () => {
  const ok = await migrationStore.fetchStatus()
  if (!ok) {
    theToast.error('查询迁移状态失败')
    return
  }
  theToast.success('状态已更新')
}

const handleCancelJob = async () => {
  if (!migrationStore.isRunning) {
    theToast.info('当前没有进行中的迁移')
    return
  }
  const res = await migrationStore.cancelMigration()
  if (res.code !== 1) {
    theToast.error(res.msg || '取消任务失败')
    return
  }
  theToast.success('迁移已取消')
}

const handleCleanupMigration = async () => {
  const res = await migrationStore.cleanupMigration()
  if (res.code !== 1) {
    theToast.error(res.msg || '清理迁移失败')
    return
  }
  theToast.success('迁移记录已清理')
}

const formatTime = (iso?: string) => {
  if (!iso) return ''
  const dt = new Date(iso)
  if (Number.isNaN(dt.getTime())) return iso
  return dt.toLocaleString('zh-CN', { hour12: false })
}

void migrationStore.init()
</script>

<style scoped>
.migration-wrap {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.migration-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.migration-title {
  color: var(--color-text-primary);
  font-size: 1.05rem;
  font-weight: 700;
}

.migration-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.migration-source-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.migration-source-card {
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  border-radius: var(--radius-md);
  padding: 0.75rem;
  text-align: left;
  transition: all 0.2s ease;
}

.migration-source-card h3 {
  color: var(--color-text-primary);
  font-weight: 700;
  margin-bottom: 0.35rem;
}

.migration-source-card p {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
}

.migration-source-card.active {
  border-color: var(--color-nav-active-bg);
  box-shadow: inset 0 0 0 1px var(--color-nav-active-bg);
}

.migration-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.migration-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.migration-row-top {
  align-items: flex-start;
}

.migration-label {
  min-width: 6.2rem;
  color: var(--color-text-secondary);
  font-weight: 600;
}

.migration-upload-wrap {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
}


.migration-file-name {
  color: var(--color-text-secondary);
  font-size: 0.82rem;
}

.migration-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.migration-job {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  padding: 0.9rem;
  background: var(--color-bg-surface);
}

.migration-job-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.6rem;
}

.migration-job-title-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.migration-job-title {
  color: var(--color-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
}

.migration-job-subtitle {
  color: var(--color-text-secondary);
  font-size: 0.8rem;
}

.migration-status-pill {
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 700;
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
}

.status-running,
.status-pending {
  color: var(--color-nav-active-bg);
  border-color: var(--color-nav-active-bg);
}

.status-success {
  color: #1f9d55;
  border-color: #1f9d55;
}

.status-failed,
.status-cancelled {
  color: #d64545;
  border-color: #d64545;
}

.migration-job-error {
  margin-top: 0.65rem;
  color: #d64545;
  font-size: 0.83rem;
}

.migration-job-metrics {
  margin-top: 0.75rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.55rem;
}

.metric-item {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.5rem 0.6rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.metric-label {
  color: var(--color-text-secondary);
  font-size: 0.78rem;
}

.metric-value {
  color: var(--color-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
}

.migration-job-meta {
  margin-top: 0.7rem;
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

@media (max-width: 768px) {
  .migration-source-grid {
    grid-template-columns: 1fr;
  }

  .migration-job-metrics {
    grid-template-columns: 1fr;
  }
}
</style>
