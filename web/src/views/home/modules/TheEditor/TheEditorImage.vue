<template>
  <!-- 图片预览 -->
  <div
    v-if="
      filesToAdd &&
      filesToAdd.length > 0 &&
      (currentMode === Mode.ECH0 || currentMode === Mode.Image)
    "
    class="relative rounded-lg shadow-lg w-5/6 mx-auto my-7"
  >
    <button
      @click="handleRemoveImage"
      :disabled="isDeleting"
      class="absolute -top-3 -right-4 bg-[var(--color-accent-soft)] hover:bg-[var(--color-danger)]/30 text-[var(--color-text-secondary)] rounded-lg w-7 h-7 flex items-center justify-center shadow-[var(--shadow-sm)] disabled:opacity-60 disabled:cursor-not-allowed"
      v-tooltip="t('editor.removeImage')"
    >
      <Close class="w-4 h-4" />
    </button>
    <div class="rounded-lg overflow-hidden">
      <template v-for="(img, idx) in filesToAdd" :key="idx">
        <button
          type="button"
          class="block w-full bg-transparent border-0 p-0 cursor-zoom-in"
          :class="{ hidden: idx !== fileIndex }"
          @click="openPreview(fileIndex)"
        >
          <img
            :src="getImageToAddUrl(img)"
            alt="Image"
            class="w-full h-auto max-w-full object-cover"
            loading="lazy"
          />
        </button>
      </template>
    </div>
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
import { getImageToAddUrl } from '@/utils/other'
import { deleteFileById } from '@/lib/file'
import { theToast } from '@/utils/toast'
import { useEchoStore, useEditorStore } from '@/stores'
import { Mode } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useI18n } from 'vue-i18n'
import { usePhotoSwipeGallery } from '@/components/advanced/gallery/composables/usePhotoSwipeGallery'

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
const { filesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)
const previewItems = computed(() =>
  filesToAdd.value.map((image, idx) => ({
    src: getImageToAddUrl(image),
    width: image.width,
    height: image.height,
    alt: t('imageGallery.previewImage', { index: idx + 1 }),
  })),
)
const { open: openPreview } = usePhotoSwipeGallery(previewItems)

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
