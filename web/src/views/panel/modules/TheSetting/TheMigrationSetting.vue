<template>
  <PanelCard>
    <div class="migration-wrap">
      <div class="migration-header">
        <h1 class="migration-title">数据导入迁移</h1>
        <p class="migration-desc">支持从 Memos 和 Ech0 v3 导入，统一进入异步迁移任务队列。</p>
      </div>

      <div class="migration-source-grid">
        <button
          v-for="source in sourceCards"
          :key="source.value"
          class="migration-source-card"
          :class="{ active: sourceType === source.value }"
          @click="sourceType = source.value"
        >
          <h3>{{ source.title }}</h3>
          <p>{{ source.desc }}</p>
        </button>
      </div>

      <div class="migration-form">
        <div class="migration-row">
          <span class="migration-label">来源版本</span>
          <BaseInput
            v-model="sourceVersion"
            type="text"
            placeholder="例如: 0.24.x / 3.0.x"
            class="migration-input"
          />
        </div>
        <div class="migration-row migration-row-top">
          <span class="migration-label">来源数据(JSON)</span>
          <BaseTextArea
            v-model="sourcePayloadText"
            :rows="7"
            placeholder='{"items":[{"id":"1","content":"hello ech0"}]}'
            class="migration-textarea"
          />
        </div>
      </div>

      <div class="migration-actions">
        <BaseButton title="开始迁移" @click="handleCreateJob">开始迁移</BaseButton>
        <BaseButton title="刷新状态" @click="handleRefreshJob">刷新状态</BaseButton>
        <BaseButton title="取消任务" @click="handleCancelJob">取消任务</BaseButton>
        <BaseButton title="重试失败项" @click="handleRetryFailed">重试失败项</BaseButton>
      </div>

      <div class="migration-job" v-if="jobId">
        <p class="migration-job-id">任务ID: {{ jobId }}</p>
        <p class="migration-job-status">状态: {{ jobStatus }}</p>
        <p class="migration-job-status">阶段: {{ jobPhase }}</p>
        <p class="migration-job-status">
          进度: {{ processed }}/{{ total }} | 成功 {{ successCount }} | 失败 {{ failCount }}
        </p>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import {
  fetchCancelMigrationJob,
  fetchCreateMigrationJob,
  fetchGetMigrationJob,
  fetchRetryFailedMigrationJob,
} from '@/service/api'
import { theToast } from '@/utils/toast'

const sourceCards = [
  { value: 'memos', title: 'Memos', desc: '从 Memos 导出数据 JSON 迁移到 Ech0' },
  { value: 'ech0_v3', title: 'Ech0 v3', desc: '将 Ech0 v3 数据映射并导入当前版本' },
]

const sourceType = ref<'memos' | 'ech0_v3'>('memos')
const sourceVersion = ref('')
const sourcePayloadText = ref('{"items":[]}')
const jobId = ref('')
const jobStatus = ref('-')
const jobPhase = ref('-')
const total = ref(0)
const processed = ref(0)
const successCount = ref(0)
const failCount = ref(0)

const parseSourcePayload = () => {
  try {
    return JSON.parse(sourcePayloadText.value)
  } catch (_error) {
    theToast.error('来源数据 JSON 格式不正确')
    return null
  }
}

const handleCreateJob = async () => {
  const sourcePayload = parseSourcePayload()
  if (!sourcePayload) return

  const res = await fetchCreateMigrationJob({
    source_type: sourceType.value,
    source_version: sourceVersion.value.trim(),
    source_payload: sourcePayload,
  })
  if (res.code !== 1) {
    theToast.error(res.msg || '创建迁移任务失败')
    return
  }

  jobId.value = res.data?.id ?? ''
  jobStatus.value = res.data?.status ?? '-'
  jobPhase.value = res.data?.current_phase ?? '-'
  total.value = Number(res.data?.total ?? 0)
  processed.value = Number(res.data?.processed ?? 0)
  successCount.value = Number(res.data?.success_count ?? 0)
  failCount.value = Number(res.data?.fail_count ?? 0)
  theToast.success('迁移任务已创建')
}

const handleRefreshJob = async () => {
  if (!jobId.value) {
    theToast.info('请先创建迁移任务')
    return
  }
  const res = await fetchGetMigrationJob(jobId.value)
  if (res.code !== 1) {
    theToast.error(res.msg || '查询迁移任务失败')
    return
  }
  jobStatus.value = res.data?.status ?? '-'
  jobPhase.value = res.data?.current_phase ?? '-'
  total.value = Number(res.data?.total ?? 0)
  processed.value = Number(res.data?.processed ?? 0)
  successCount.value = Number(res.data?.success_count ?? 0)
  failCount.value = Number(res.data?.fail_count ?? 0)
  theToast.success('状态已更新')
}

const handleCancelJob = async () => {
  if (!jobId.value) {
    theToast.info('请先创建迁移任务')
    return
  }
  const res = await fetchCancelMigrationJob(jobId.value)
  if (res.code !== 1) {
    theToast.error(res.msg || '取消任务失败')
    return
  }
  jobStatus.value = 'cancelled'
  theToast.success('任务已取消')
}

const handleRetryFailed = async () => {
  if (!jobId.value) {
    theToast.info('请先创建迁移任务')
    return
  }
  const res = await fetchRetryFailedMigrationJob(jobId.value)
  if (res.code !== 1) {
    theToast.error(res.msg || '重试失败项失败')
    return
  }
  jobStatus.value = 'pending'
  jobPhase.value = 'extracting'
  processed.value = 0
  total.value = 0
  successCount.value = 0
  failCount.value = 0
  theToast.success('失败项已重新入队')
}
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
  grid-template-columns: repeat(2, minmax(0, 1fr));
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

.migration-input,
.migration-textarea {
  width: 100%;
}

.migration-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.migration-job {
  border: 1px dashed var(--color-border-subtle);
  border-radius: var(--radius-md);
  padding: 0.75rem;
  background: var(--color-bg-canvas);
}

.migration-job-id,
.migration-job-status {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  margin-bottom: 0.35rem;
}

@media (max-width: 768px) {
  .migration-source-grid {
    grid-template-columns: 1fr;
  }
}
</style>
