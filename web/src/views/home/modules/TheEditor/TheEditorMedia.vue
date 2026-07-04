<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <!-- 媒体预览（图片可放大；音频/视频复用展示端播放器，观感与发布后一致） -->
  <div
    v-if="
      filesToAdd &&
      filesToAdd.length > 0 &&
      (currentMode === Mode.ECH0 || currentMode === Mode.Media)
    "
    class="relative w-11/12 mx-auto my-7"
  >
    <button
      @click="handleRemoveImage"
      :disabled="isDeleting"
      class="absolute -top-2.5 -right-2.5 z-[2] w-7 h-7 flex items-center justify-center rounded-full bg-[var(--color-bg-surface)] text-[var(--color-text-secondary)] border border-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] transition-colors hover:bg-[var(--color-danger)] hover:text-white hover:border-[var(--color-danger)] disabled:opacity-60 disabled:cursor-not-allowed"
      v-tooltip="t('editor.removeImage')"
    >
      <Close color="currentColor" class="w-4 h-4" />
    </button>

    <!-- 图片：可点击放大预览，多图轮播 -->
    <div v-if="isImage" class="rounded-lg overflow-hidden shadow-lg">
      <template v-for="(file, idx) in filesToAdd" :key="idx">
        <button
          type="button"
          class="block w-full bg-transparent border-0 p-0 cursor-zoom-in"
          :class="{ hidden: idx !== fileIndex }"
          @click="openPreview(fileIndex)"
        >
          <img
            :src="getImageToAddUrl(file)"
            alt="Image"
            class="w-full h-auto max-w-full object-cover"
            loading="lazy"
          />
        </button>
      </template>
    </div>

    <!-- 音频 -->
    <TheAudioPlayer v-else-if="isAudio" :files="mediaFiles" />

    <!-- 视频 -->
    <TheVideoPlayer v-else :files="mediaFiles" />
  </div>
  <!-- 图片切换 -->
  <div v-if="filesToAdd.length > 1" class="flex items-center justify-center">
    <button @click="fileIndex = Math.max(fileIndex - 1, 0)">
      <Prev class="w-7 h-7" />
    </button>
    <span class="text-[var(--color-text-secondary)] text-sm mx-2">
      {{ fileIndex + 1 }} / {{ filesToAdd.length }}
    </span>
    <button @click="fileIndex = Math.min(fileIndex + 1, filesToAdd.length - 1)">
      <Next class="w-7 h-7" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import Next from '@/components/icons/next.vue'
import Prev from '@/components/icons/prev.vue'
import Close from '@/components/icons/close.vue'
import TheAudioPlayer from '@/components/advanced/media/audio/TheAudioPlayer.vue'
import TheVideoPlayer from '@/components/advanced/media/video/TheVideoPlayer.vue'
import { getImageToAddUrl } from '@/utils/other'
import { deleteFileById } from '@/lib/file'
import { theToast } from '@/utils/toast'
import { useEchoStore, useEditorStore } from '@/stores'
import { Mode } from '@/enums/enums'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useI18n } from 'vue-i18n'
import { usePhotoSwipeGallery } from '@/components/advanced/media/image/composables/usePhotoSwipeGallery'

const { openConfirm } = useBaseDialog()
const { t } = useI18n()

// const images = defineModel<App.Api.Ech0.ImageToAdd[]>('imagesToAdd', { required: true })

// const { currentMode } = defineProps<{
//   currentMode: Mode
// }>()

// const emit = defineEmits(['handleAddorUpdateEcho'])

const fileIndex = ref<number>(0) // 临时文件索引变量
const isDeleting = ref<boolean>(false)
const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { filesToAdd, currentMode, isUpdateMode, mediaCategory } = storeToRefs(editorStore)
const previewItems = computed(() =>
  filesToAdd.value.map((image, idx) => ({
    src: getImageToAddUrl(image),
    width: image.width,
    height: image.height,
    alt: t('imageGallery.previewImage', { index: idx + 1 }),
  })),
)
const { open: openPreview } = usePhotoSwipeGallery(previewItems)

const isAudio = computed(() => mediaCategory.value === FILE_CATEGORY.AUDIO)
const isVideo = computed(() => mediaCategory.value === FILE_CATEGORY.VIDEO)
// 图片是默认分类：既非音频也非视频即按图片处理（与轮播/放大预览一致）。
const isImage = computed(() => !isAudio.value && !isVideo.value)

// 编辑器里的 FileToAdd 与展示端 FileObject 仅差 echo_id 等字段；补齐后复用同一套播放器，
// URL 解析走同一个 resolveFileUrl，预览观感与发布后完全一致。音视频每条至多一个。
const mediaFiles = computed<App.Api.Ech0.FileObject[]>(() =>
  filesToAdd.value.map((file) => ({ ...file, id: file.id ?? '', echo_id: '' })),
)

const handleRemoveImage = () => {
  if (isDeleting.value) return
  if (
    fileIndex.value < 0 ||
    fileIndex.value >= filesToAdd.value.length ||
    filesToAdd.value.length === 0
  ) {
    theToast.error(String(t('editor.invalidImageIndex')))
    return
  }
  const index = fileIndex.value

  openConfirm({
    title: String(t('editor.removeImageConfirmTitle')),
    description: '',
    onConfirm: async () => {
      if (isDeleting.value) return
      const target = filesToAdd.value[index]
      const fileId = String(target?.id || '')
      const source = String(target?.storage_type || '')
      const needsRemoteDelete =
        (source === FILE_STORAGE_TYPE.LOCAL || source === FILE_STORAGE_TYPE.OBJECT) && !!fileId

      isDeleting.value = true
      try {
        if (needsRemoteDelete) {
          try {
            await deleteFileById(fileId)
          } catch {
            // Per design, local removal proceeds even if the backend delete fails;
            // surface a warning so the user knows the backend file may be orphaned.
            theToast.error(String(t('editor.removeImageRemoteFailed')))
          }
        }

        editorStore.removeFileAt(index)

        // Keep the carousel near the removed slot rather than snapping back to 0.
        const nextLen = filesToAdd.value.length
        fileIndex.value = nextLen === 0 ? 0 : Math.min(index, nextLen - 1)

        if (isUpdateMode.value && echoToUpdate.value) {
          editorStore.handleAddOrUpdateEcho(true)
        }
      } finally {
        isDeleting.value = false
      }
    },
  })
}
</script>

<style scoped></style>
