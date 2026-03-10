<template>
  <div>
    <h2 class="text-[var(--text-color-500)] font-bold mb-1">音乐播放器模式</h2>
    <div class="mb-1 flex items-center gap-2">
      <p class="text-[var(--text-color-500)]">上传音乐：</p>
      <input
        id="file-input"
        class="hidden"
        type="file"
        accept="audio/*"
        ref="fileInput"
        @change="handleUploadMusic"
      />
      <BaseButton
        :icon="UploadMusic"
        @click="handleTriggerUpload"
        class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
        title="上传音乐"
      />
    </div>
    <div class="flex items-center gap-2">
      <p class="text-[var(--text-color-500)]">删除音乐：</p>
      <BaseButton
        :icon="Delete"
        @click="handleDeleteMusic"
        class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
        title="删除音乐"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import UploadMusic from '@/components/icons/musicupload.vue'
import Delete from '@/components/icons/delete.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ref } from 'vue'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { uploadFile, useFilePlayer } from '@/lib/file'

const { openConfirm } = useBaseDialog()

const filePlayer = useFilePlayer()

const fileInput = ref<HTMLInputElement | null>(null)
const handleTriggerUpload = () => {
  if (fileInput.value) {
    fileInput.value.click()
  }
}

const handleUploadMusic = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  try {
    const uploaded = await theToast.promise(
      uploadFile({
        file,
        storageType: FILE_STORAGE_TYPE.LOCAL,
        category: FILE_CATEGORY.AUDIO,
      }),
      {
      loading: '音乐上传中...',
      success: '音乐上传成功！',
      error: '音乐上传失败，请稍后再试',
      },
    )

    if (uploaded.id) {
      await filePlayer.setPlayingFile(uploaded.id)
    }
  } catch (err) {
    console.error('音乐上传异常:', err)
  } finally {
    target.value = ''
  }
}

const handleDeleteMusic = () => {
  openConfirm({
    title: '确定要删除音乐吗？',
    description: '删除后将无法恢复，请谨慎操作',
    onConfirm: async () => {
      await filePlayer.clearAndDeleteCurrent()
      theToast.success('音乐删除成功！')
    },
  })
}
</script>

<style scoped></style>
