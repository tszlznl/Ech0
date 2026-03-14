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
      class="absolute -top-3 -right-4 bg-[var(--color-accent-soft)] hover:bg-[var(--color-danger)]/30 text-[var(--color-text-secondary)] rounded-lg w-7 h-7 flex items-center justify-center shadow-[var(--shadow-sm)]"
      :title="t('editor.removeImage')"
    >
      <Close class="w-4 h-4" />
    </button>
    <div class="rounded-lg overflow-hidden">
      <template v-for="(img, idx) in filesToAdd" :key="idx">
        <a
          :href="getImageToAddUrl(img)"
          data-fancybox="gallery"
          :data-thumb="getImageToAddUrl(img)"
          class="block w-full"
          :class="{ hidden: idx !== fileIndex }"
        >
          <img
            :src="getImageToAddUrl(img)"
            alt="Image"
            class="w-full h-auto max-w-full object-cover"
            loading="lazy"
          />
        </a>
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
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import Next from '@/components/icons/next.vue'
import Prev from '@/components/icons/prev.vue'
import Close from '@/components/icons/close.vue'
import { getImageToAddUrl } from '@/utils/other'
import { deleteFileById } from '@/lib/file'
import { theToast } from '@/utils/toast'
import { useEchoStore, useEditorStore } from '@/stores'
import { Mode } from '@/enums/enums'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useI18n } from 'vue-i18n'

const { openConfirm } = useBaseDialog()
const { t } = useI18n()

// const images = defineModel<App.Api.Ech0.ImageToAdd[]>('imagesToAdd', { required: true })

// const { currentMode } = defineProps<{
//   currentMode: Mode
// }>()

// const emit = defineEmits(['handleAddorUpdateEcho'])

const fileIndex = ref<number>(0) // 临时文件索引变量
const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { filesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)

const handleRemoveImage = () => {
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
    onConfirm: () => {
      const fileToDelete: App.Api.Ech0.FileToDelete = {
        id: String(filesToAdd.value[index]?.id || ''),
      }

      const source = String(filesToAdd.value[index]?.storage_type || '')
      if (
        (source === FILE_STORAGE_TYPE.LOCAL || source === FILE_STORAGE_TYPE.OBJECT) &&
        fileToDelete.id
      ) {
        deleteFileById(fileToDelete.id).then(() => {
          // 这里不管图片是否远程删除成功都强制删除图片
          // 从数组中删除图片
          editorStore.removeFileAt(index)

          // 如果删除成功且当前处于Echo更新模式，则需要立马执行更新（图片删除操作不可逆，需要立马更新确保后端数据同步）
          if (isUpdateMode.value && echoToUpdate.value) {
            editorStore.handleAddOrUpdateEcho(true)
          }
        })
      } else {
        editorStore.removeFileAt(index)
      }

      fileIndex.value = 0
    },
  })
}

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})
</script>

<style scoped></style>
