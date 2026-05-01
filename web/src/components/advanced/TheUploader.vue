<template>
  <div class="rounded-md overflow-hidden">
    <!-- Drop zone (also click-to-pick) -->
    <button
      type="button"
      class="w-full flex flex-col items-center justify-center gap-1 py-5 px-3 rounded-md border-1 border-dashed transition-colors cursor-pointer text-center select-none"
      :class="[
        isDragOver
          ? 'border-[var(--color-accent)] bg-[var(--color-accent-soft)]'
          : 'border-[var(--md-editor-mini-border)] bg-[var(--color-bg-surface-alt,transparent)] hover:border-[var(--color-border-strong)]',
      ]"
      @click="openFilePicker"
      @dragover.prevent="isDragOver = true"
      @dragenter.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
      @drop.prevent="handleDrop"
    >
      <span class="text-[var(--color-text-primary)] font-medium">
        {{ t('uploader.dropHere') }}
      </span>
      <span class="text-xs text-[var(--color-text-muted)]">
        {{ t('uploader.dropHint', { max: maxFiles }) }}
      </span>
    </button>

    <input
      ref="fileInputRef"
      type="file"
      multiple
      :accept="acceptAttr"
      class="hidden"
      @change="handleInputChange"
    />

    <!-- Per-file rows -->
    <ul v-if="items.length > 0" class="mt-2 flex flex-col gap-1.5">
      <li
        v-for="item in items"
        :key="item.id"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md bg-[var(--color-bg-surface-alt,transparent)] ring-1 ring-inset ring-[var(--md-editor-mini-border)]"
      >
        <img
          :src="item.preview"
          alt=""
          class="w-9 h-9 object-cover rounded-sm shrink-0 bg-[var(--color-bg-muted)]"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between gap-2">
            <span class="text-sm truncate text-[var(--color-text-primary)]" :title="item.file.name">
              {{ item.file.name }}
            </span>
            <span class="text-xs shrink-0" :class="statusColorClass(item.status)">
              {{ statusLabel(item) }}
            </span>
          </div>
          <div
            v-if="item.status === 'uploading' || item.status === 'compressing'"
            class="mt-1 h-1 rounded-full bg-[var(--color-border-subtle,#e5e7eb)] overflow-hidden"
          >
            <div
              class="h-full bg-[var(--color-accent)] transition-all"
              :style="{ width: item.progress + '%' }"
            />
          </div>
          <div
            v-else-if="item.status === 'error' && item.error"
            class="text-xs text-[var(--color-danger,#dc2626)] truncate mt-0.5"
            :title="item.error"
          >
            {{ item.error }}
          </div>
        </div>
        <div class="flex items-center gap-1 shrink-0">
          <button
            v-if="item.status === 'error' || item.status === 'cancelled'"
            type="button"
            class="text-xs px-2 py-0.5 rounded cursor-pointer text-[var(--color-accent)] hover:bg-[var(--color-accent-soft)]"
            @click="retry(item.id)"
          >
            {{ t('uploader.retry') }}
          </button>
          <button
            type="button"
            class="text-xs px-2 py-0.5 rounded cursor-pointer text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)]"
            @click="remove(item.id)"
          >
            ✕
          </button>
        </div>
      </li>
    </ul>

    <!-- Global status -->
    <div
      v-if="isUploading"
      class="mt-2 flex items-center justify-between text-xs text-[var(--color-text-muted)]"
    >
      <span>{{ t('uploader.uploadingProgress', { active: totalActive }) }}</span>
      <button
        type="button"
        class="cursor-pointer hover:text-[var(--color-danger,#dc2626)]"
        @click="cancelAll"
      >
        {{ t('uploader.cancelAll') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, toRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorStore } from '@/stores'
import { useUpload, type QueueItem, type UploadStatus } from '@/lib/file/useUpload'

const props = defineProps<{
  fileStorageType: App.Api.File.StorageType
  EnableCompressor: boolean
  fileCategory?: App.Api.File.Category
  allowedFileTypes?: string[]
}>()

const editorStore = useEditorStore()
const { t } = useI18n()

const fileInputRef = ref<HTMLInputElement | null>(null)
const isDragOver = ref(false)

const allowedTypes = computed(() =>
  props.allowedFileTypes?.length ? props.allowedFileTypes : ['image/*'],
)
const acceptAttr = computed(() => allowedTypes.value.join(','))

const maxFiles = 6

const { items, isUploading, totalActive, addFiles, retry, remove, cancelAll } = useUpload({
  storageType: toRef(props, 'fileStorageType'),
  enableCompressor: toRef(props, 'EnableCompressor'),
  category: props.fileCategory,
  allowedTypes: allowedTypes.value,
  maxFiles,
  concurrency: 3,
  onAllComplete: (results) => {
    editorStore.handleUppyUploaded(results)
  },
})

watch(
  isUploading,
  (v) => {
    editorStore.fileUploading = v
  },
  { immediate: true },
)

function openFilePicker() {
  fileInputRef.value?.click()
}

function handleInputChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    addFiles(Array.from(input.files))
    input.value = ''
  }
}

function handleDrop(e: DragEvent) {
  isDragOver.value = false
  const dt = e.dataTransfer
  if (!dt) return
  const files: File[] = []
  if (dt.items) {
    for (const item of dt.items) {
      if (item.kind === 'file') {
        const f = item.getAsFile()
        if (f) files.push(f)
      }
    }
  } else if (dt.files) {
    for (const f of dt.files) files.push(f)
  }
  if (files.length > 0) addFiles(files)
}

function handlePaste(e: ClipboardEvent) {
  if (!e.clipboardData) return
  const files: File[] = []
  for (const item of e.clipboardData.items) {
    if (item.kind === 'file' && item.type.startsWith('image/')) {
      const f = item.getAsFile()
      if (f) {
        // Preserve lastModified so the dedup id matches across paste + pick of the same image.
        files.push(new File([f], f.name, { type: f.type, lastModified: f.lastModified }))
      }
    }
  }
  if (files.length > 0) addFiles(files)
}

onMounted(() => {
  document.addEventListener('paste', handlePaste)
})
onBeforeUnmount(() => {
  document.removeEventListener('paste', handlePaste)
  cancelAll()
})

function statusLabel(item: QueueItem): string {
  switch (item.status) {
    case 'pending':
      return t('uploader.statusPending')
    case 'compressing':
      return t('uploader.statusCompressing')
    case 'uploading':
      return `${item.progress}%`
    case 'success':
      return t('uploader.statusSuccess')
    case 'cancelled':
      return t('uploader.statusCancelled')
    case 'error':
      return t('uploader.statusError')
    default:
      return ''
  }
}

function statusColorClass(status: UploadStatus): string {
  switch (status) {
    case 'success':
      return 'text-[var(--color-success,#16a34a)]'
    case 'error':
      return 'text-[var(--color-danger,#dc2626)]'
    case 'uploading':
    case 'compressing':
      return 'text-[var(--color-accent)]'
    default:
      return 'text-[var(--color-text-muted)]'
  }
}
</script>
