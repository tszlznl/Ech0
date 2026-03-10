<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { fetchDownloadFileById, fetchListFiles } from '@/service/api'
import { onMounted, reactive, ref } from 'vue'
import { theToast } from '@/utils/toast'

type RootStorageType = typeof FILE_STORAGE_TYPE.LOCAL | typeof FILE_STORAGE_TYPE.OBJECT

type FileSectionState = {
  loading: boolean
  error: string
  total: number
  items: App.Api.File.FileListItem[]
}

const sections = reactive<Record<RootStorageType, FileSectionState>>({
  [FILE_STORAGE_TYPE.LOCAL]: { loading: false, error: '', total: 0, items: [] },
  [FILE_STORAGE_TYPE.OBJECT]: { loading: false, error: '', total: 0, items: [] },
})

const downloadingId = ref('')

const sectionMeta: Record<RootStorageType, { title: string }> = {
  [FILE_STORAGE_TYPE.LOCAL]: { title: 'LocalFS' },
  [FILE_STORAGE_TYPE.OBJECT]: { title: 'ObjectFS' },
}

const loadSection = async (storageType: RootStorageType) => {
  sections[storageType].loading = true
  sections[storageType].error = ''
  const res = await fetchListFiles({
    page: 1,
    pageSize: 50,
    storage_type: storageType,
  })
  if (res.code === 1) {
    sections[storageType].items = res.data?.items || []
    sections[storageType].total = res.data?.total || 0
  } else {
    sections[storageType].error = res.msg || '加载失败'
  }
  sections[storageType].loading = false
}

const loadAll = async () => {
  await Promise.all([loadSection(FILE_STORAGE_TYPE.LOCAL), loadSection(FILE_STORAGE_TYPE.OBJECT)])
}

const triggerDownload = async (item: App.Api.File.FileListItem) => {
  try {
    downloadingId.value = item.id
    const blob = await fetchDownloadFileById(item.id)
    const objectUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = item.name || item.key || 'file'
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(objectUrl)
  } catch {
    theToast.error('下载失败')
  } finally {
    downloadingId.value = ''
  }
}

onMounted(() => {
  void loadAll()
})
</script>

<template>
  <PanelCard class="mt-3">
    <div class="storage-file-list">
      <div class="header">
        <h1 class="title">文件管理</h1>
        <button class="refresh-btn" @click="loadAll">刷新</button>
      </div>

      <section
        v-for="storageType in [FILE_STORAGE_TYPE.LOCAL, FILE_STORAGE_TYPE.OBJECT]"
        :key="storageType"
        class="root-section"
      >
        <div class="root-head">
          <h2>{{ sectionMeta[storageType].title }}</h2>
          <span class="count">共 {{ sections[storageType].total }} 个文件</span>
        </div>

        <div v-if="sections[storageType].loading" class="status-text">正在加载...</div>
        <div v-else-if="sections[storageType].error" class="status-text status-error">
          {{ sections[storageType].error }}
          <button class="retry-btn" @click="loadSection(storageType)">重试</button>
        </div>
        <div v-else-if="sections[storageType].items.length === 0" class="status-text">暂无文件</div>
        <div v-else class="table-wrap">
          <table class="file-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>Key</th>
                <th class="action-col">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in sections[storageType].items" :key="item.id">
                <td class="name-cell" :title="item.name">{{ item.name }}</td>
                <td class="key-cell" :title="item.key">{{ item.key }}</td>
                <td class="action-col">
                  <button
                    class="download-btn"
                    :disabled="downloadingId === item.id"
                    @click="triggerDownload(item)"
                  >
                    {{ downloadingId === item.id ? '下载中...' : '下载' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </PanelCard>
</template>

<style scoped>
.storage-file-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  color: var(--color-text-primary);
  font-weight: 700;
  font-size: 1.1rem;
}

.refresh-btn,
.retry-btn,
.download-btn {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.2rem 0.65rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  cursor: pointer;
}

.refresh-btn:hover,
.retry-btn:hover,
.download-btn:hover:not(:disabled) {
  border-color: var(--color-border-strong);
}

.download-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.root-section {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  padding: 0.7rem;
}

.root-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.55rem;
}

.root-head h2 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.count {
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.status-text {
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.status-error {
  color: var(--color-error, #ef4444);
}

.table-wrap {
  overflow-x: auto;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.file-table th,
.file-table td {
  border-top: 1px solid var(--color-border-subtle);
  padding: 0.5rem 0.35rem;
  text-align: left;
  color: var(--color-text-secondary);
}

.name-cell,
.key-cell {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-col {
  width: 86px;
  text-align: center !important;
}
</style>
