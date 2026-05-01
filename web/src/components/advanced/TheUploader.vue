<template>
  <div class="rounded-md overflow-hidden">
    <!-- Drop zone (also click-to-pick) -->
    <button
      type="button"
      class="w-full flex flex-col items-center justify-center gap-1 py-5 px-3 rounded-md border-1 border-dashed transition-colors cursor-pointer text-center select-none"
      :class="[
        isDropOverZone
          ? 'border-[var(--color-accent)] bg-[var(--color-accent-soft)]'
          : 'border-[var(--md-editor-mini-border)] bg-[var(--color-bg-surface-alt,transparent)] hover:border-[var(--color-border-strong)]',
      ]"
      @click="openFilePicker"
      @dragover.prevent="onZoneDragOver"
      @dragenter.prevent="onZoneDragOver"
      @dragleave.prevent="isDropOverZone = false"
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

    <!-- Per-file thumbnail cards -->
    <ul v-if="items.length > 0" class="mt-2 grid grid-cols-2 gap-2">
      <li
        v-for="item in items"
        :key="item.id"
        class="group relative flex gap-2 p-2 rounded-md ring-1 ring-inset ring-[var(--md-editor-mini-border)] bg-[var(--color-bg-surface-alt,transparent)] transition-colors"
        :class="[
          isReorderable(item) ? 'cursor-grab active:cursor-grabbing' : '',
          dragOverId === item.id && draggingId !== item.id
            ? 'ring-[var(--color-accent)] bg-[var(--color-accent-soft)]'
            : '',
          draggingId === item.id ? 'opacity-50' : '',
        ]"
        :draggable="isReorderable(item)"
        @dragstart="(e) => onItemDragStart(item, e)"
        @dragenter.prevent="onItemDragEnter(item)"
        @dragover.prevent="onItemDragOver(item, $event)"
        @dragleave="onItemDragLeave(item)"
        @drop.prevent="onItemDrop(item)"
        @dragend="onItemDragEnd"
      >
        <!-- Thumbnail with status overlay -->
        <div
          class="relative w-16 h-16 shrink-0 rounded-sm overflow-hidden bg-[var(--color-bg-muted)]"
        >
          <img :src="item.preview" alt="" class="w-full h-full object-cover" />
          <div
            v-if="item.status === UPLOAD_STATUS.SUCCESS"
            class="absolute inset-0 bg-[var(--color-success,#16a34a)]/15 flex items-center justify-center"
          >
            <span class="text-base text-[var(--color-success,#16a34a)] font-bold">✓</span>
          </div>
          <div
            v-else-if="item.status === UPLOAD_STATUS.ERROR"
            class="absolute inset-0 bg-[var(--color-danger,#dc2626)]/20 flex items-center justify-center"
          >
            <span class="text-base text-[var(--color-danger,#dc2626)] font-bold">!</span>
          </div>
          <div
            v-else-if="item.status === UPLOAD_STATUS.CANCELLED"
            class="absolute inset-0 bg-black/30 flex items-center justify-center"
          >
            <span class="text-sm text-white">×</span>
          </div>
        </div>

        <!-- Info column -->
        <div class="flex-1 min-w-0 flex flex-col justify-between">
          <div class="min-w-0">
            <div
              class="text-sm truncate text-[var(--color-text-primary)] leading-tight"
              :title="item.file.name"
            >
              {{ item.file.name }}
            </div>
            <div class="mt-0.5 text-xs text-[var(--color-text-muted)] flex items-center gap-1.5">
              <span>{{ formatSize(item.file.size) }}</span>
              <span
                v-if="compressionDelta(item)"
                class="px-1 rounded bg-[var(--color-accent-soft)] text-[var(--color-accent)]"
                :title="
                  t('uploader.compressionTooltip', {
                    from: formatSize(item.originalSize),
                    to: formatSize(item.file.size),
                  })
                "
              >
                {{ compressionDelta(item) }}
              </span>
            </div>
          </div>

          <!-- Progress bar OR status text OR error -->
          <div class="mt-1">
            <div
              v-if="
                item.status === UPLOAD_STATUS.UPLOADING || item.status === UPLOAD_STATUS.COMPRESSING
              "
              class="h-1 rounded-full bg-[var(--color-border-subtle,#e5e7eb)] overflow-hidden"
            >
              <div
                class="h-full bg-[var(--color-accent)] transition-all"
                :style="{
                  width: (item.status === UPLOAD_STATUS.COMPRESSING ? 100 : item.progress) + '%',
                }"
              />
            </div>
            <div
              v-else-if="item.status === UPLOAD_STATUS.ERROR && item.error"
              class="text-xs text-[var(--color-danger,#dc2626)] truncate"
              :title="item.error"
            >
              {{ item.error }}
            </div>
            <div v-else class="text-xs" :class="statusColorClass(item.status)">
              {{ statusLabel(item) }}
            </div>
          </div>
        </div>

        <!-- Actions (top-right floating) -->
        <div class="absolute top-1 right-1 flex items-center gap-1">
          <button
            v-if="item.status === UPLOAD_STATUS.ERROR || item.status === UPLOAD_STATUS.CANCELLED"
            type="button"
            class="text-xs px-1.5 py-0.5 rounded cursor-pointer text-[var(--color-accent)] bg-[var(--color-bg-surface)] hover:bg-[var(--color-accent-soft)] shadow-sm"
            @click.stop="retry(item.id)"
          >
            {{ t('uploader.retry') }}
          </button>
          <button
            type="button"
            class="w-5 h-5 flex items-center justify-center rounded-full cursor-pointer text-[var(--color-text-muted)] bg-[var(--color-bg-surface)] hover:text-[var(--color-danger,#dc2626)] shadow-sm"
            @click.stop="remove(item.id)"
            :title="t('uploader.removeItem')"
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

    <!-- Reorder hint (only once any image is uploaded) -->
    <div
      v-else-if="hasReorderable"
      class="mt-1.5 text-xs text-[var(--color-text-muted)] text-center"
    >
      {{ t('uploader.reorderHint') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, toRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorStore } from '@/stores'
import { useUpload, UPLOAD_STATUS, type QueueItem, type UploadStatus } from '@/lib/file/useUpload'

const props = defineProps<{
  fileStorageType: App.Api.File.StorageType
  EnableCompressor: boolean
  fileCategory?: App.Api.File.Category
  allowedFileTypes?: string[]
}>()

const editorStore = useEditorStore()
const { t } = useI18n()

const fileInputRef = ref<HTMLInputElement | null>(null)
const isDropOverZone = ref(false)

const allowedTypes = computed(() =>
  props.allowedFileTypes?.length ? props.allowedFileTypes : ['image/*'],
)
const acceptAttr = computed(() => allowedTypes.value.join(','))

const maxFiles = 6

const { items, isUploading, totalActive, addFiles, retry, remove, moveItem, cancelAll } = useUpload(
  {
    storageType: toRef(props, 'fileStorageType'),
    enableCompressor: toRef(props, 'EnableCompressor'),
    category: props.fileCategory,
    allowedTypes: allowedTypes.value,
    maxFiles,
    concurrency: 3,
    onAllComplete: (results) => {
      editorStore.handleUppyUploaded(results)
    },
  },
)

watch(
  isUploading,
  (v) => {
    editorStore.fileUploading = v
  },
  { immediate: true },
)

const hasReorderable = computed(
  () => items.value.filter((i) => i.status === UPLOAD_STATUS.SUCCESS).length >= 2,
)

function isReorderable(item: QueueItem): boolean {
  return item.status === UPLOAD_STATUS.SUCCESS
}

function openFilePicker() {
  fileInputRef.value?.click()
}

function onZoneDragOver(e: DragEvent) {
  // Show "drop to add" highlight only when external files are being dragged in,
  // not when the user is reordering an internal queue item.
  if (draggingId.value) return
  if (e.dataTransfer?.types?.includes('Files')) {
    isDropOverZone.value = true
  }
}

function handleInputChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    addFiles(Array.from(input.files))
    input.value = ''
  }
}

function handleDrop(e: DragEvent) {
  isDropOverZone.value = false
  // Internal reorder drag should never be treated as a "drop to add" event.
  if (draggingId.value) return
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

// ---------- Drag-to-reorder ----------

const draggingId = ref<string | null>(null)
const dragOverId = ref<string | null>(null)

function onItemDragStart(item: QueueItem, e: DragEvent) {
  if (!isReorderable(item)) {
    e.preventDefault()
    return
  }
  draggingId.value = item.id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    // Some browsers require setData for the drag to fire.
    e.dataTransfer.setData('text/plain', item.id)
  }
}

function onItemDragEnter(item: QueueItem) {
  if (!draggingId.value || !isReorderable(item)) return
  dragOverId.value = item.id
}

function onItemDragOver(item: QueueItem, e: DragEvent) {
  if (!draggingId.value || !isReorderable(item)) return
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOverId.value = item.id
}

function onItemDragLeave(item: QueueItem) {
  if (dragOverId.value === item.id) dragOverId.value = null
}

function onItemDrop(target: QueueItem) {
  const fromId = draggingId.value
  draggingId.value = null
  dragOverId.value = null
  if (!fromId || fromId === target.id || !isReorderable(target)) return
  moveItem(fromId, target.id)
  // Propagate to the editor's filesToAdd so the post payload reflects the new order.
  const orderedIds = items.value
    .filter((i) => i.status === UPLOAD_STATUS.SUCCESS && i.result?.id)
    .map((i) => i.result!.id!)
  if (orderedIds.length > 0) editorStore.reorderFilesByIds(orderedIds)
}

function onItemDragEnd() {
  draggingId.value = null
  dragOverId.value = null
}

// ---------- Display helpers ----------

const SIZE_UNITS = ['B', 'KB', 'MB', 'GB'] as const

function formatSize(bytes: number | undefined): string {
  if (bytes == null || !Number.isFinite(bytes) || bytes < 0) return '—'
  if (bytes === 0) return '0 B'
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < SIZE_UNITS.length - 1) {
    value /= 1024
    unit++
  }
  const fixed = value >= 100 || unit === 0 ? value.toFixed(0) : value.toFixed(1)
  return `${fixed} ${SIZE_UNITS[unit]}`
}

// Returns a "−42%" badge when post-compression size is meaningfully smaller, else null.
function compressionDelta(item: QueueItem): string | null {
  const before = item.originalSize
  const after = item.file.size
  if (!before || !after || after >= before) return null
  const pct = Math.round((1 - after / before) * 100)
  if (pct < 1) return null
  return `−${pct}%`
}

function statusLabel(item: QueueItem): string {
  switch (item.status) {
    case UPLOAD_STATUS.PENDING:
      return t('uploader.statusPending')
    case UPLOAD_STATUS.COMPRESSING:
      return t('uploader.statusCompressing')
    case UPLOAD_STATUS.UPLOADING:
      return `${item.progress}%`
    case UPLOAD_STATUS.SUCCESS:
      return t('uploader.statusSuccess')
    case UPLOAD_STATUS.CANCELLED:
      return t('uploader.statusCancelled')
    case UPLOAD_STATUS.ERROR:
      return t('uploader.statusError')
    default:
      return ''
  }
}

function statusColorClass(status: UploadStatus): string {
  switch (status) {
    case UPLOAD_STATUS.SUCCESS:
      return 'text-[var(--color-success,#16a34a)]'
    case UPLOAD_STATUS.ERROR:
      return 'text-[var(--color-danger,#dc2626)]'
    case UPLOAD_STATUS.UPLOADING:
    case UPLOAD_STATUS.COMPRESSING:
      return 'text-[var(--color-accent)]'
    default:
      return 'text-[var(--color-text-muted)]'
  }
}
</script>
