<template>
  <PanelCard>
    <div class="backup-wrap">
      <div class="backup-header">
        <h1 class="backup-title">数据导出</h1>
        <p class="backup-desc">导出当前系统快照，便于离线归档与迁移。</p>
      </div>

      <div class="backup-action">
        <BaseButton @click="handleBackupExport" class="backup-export-btn" title="导出快照">
          导出快照
        </BaseButton>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { theToast } from '@/utils/toast'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'

const userStore = useUserStore()
const { isLogin } = storeToRefs(userStore)

const handleBackupExport = async () => {
  if (!isLogin.value) {
    theToast.info('请登录后使用', { duration: 3000 })
    return
  }

  try {
    theToast.info('导出中...请稍等', {
      duration: 4000,
    })

    // 1. 获取 token
    const token = localStorage.getItem('token')
    const baseURL =
      import.meta.env.VITE_SERVICE_BASE_URL === '/'
        ? window.location.origin
        : import.meta.env.VITE_SERVICE_BASE_URL
    const downloadUrl = `${baseURL}/api/backup/export?token=${token}`

    // 创建隐藏的 a 标签触发下载
    const link = document.createElement('a')
    link.href = downloadUrl
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    theToast.success('备份导出开始！')
  } catch (error) {
    theToast.error('导出失败')
    console.error('导出备份失败:', error)
  }
}
</script>

<style scoped>
.backup-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.backup-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.backup-title {
  color: var(--color-text-primary);
  font-weight: 700;
  font-size: 1.05rem;
}

.backup-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.backup-action {
  display: flex;
  align-items: center;
}

.backup-export-btn {
  border-radius: var(--radius-md);
  color: var(--color-text-primary) !important;
}
</style>
