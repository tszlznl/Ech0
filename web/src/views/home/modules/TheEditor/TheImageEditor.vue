<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="editor-image-panel">
    <h2 class="text-[var(--color-text-secondary)] font-bold my-2">
      {{ t('editor.imageSectionTitle') }}
    </h2>
    <div v-if="!fileUploading" class="flex items-center gap-2 mb-3">
      <div class="flex items-center gap-2">
        <span class="text-[var(--color-text-secondary)]">{{ t('editor.imageAddMethod') }}</span>
        <!-- 直链 -->
        <BaseButton
          :icon="Url"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetFileSource(FILE_STORAGE_TYPE.EXTERNAL)"
          :tooltip="t('editor.imageSourceExternal')"
        />
        <!-- 上传本地 -->
        <BaseButton
          :icon="Upload"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetFileSource(FILE_STORAGE_TYPE.LOCAL)"
          :tooltip="t('editor.imageSourceLocal')"
        />
        <!-- S3 存储 -->
        <BaseButton
          v-if="S3Setting.enable"
          :icon="Bucket"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetFileSource(FILE_STORAGE_TYPE.OBJECT)"
          :tooltip="t('editor.imageSourceObject')"
        />
      </div>
    </div>

    <!-- 布局方式选择 -->
    <div class="mb-2 flex items-center gap-2">
      <span class="text-[var(--color-text-secondary)]">{{ t('editor.imageLayout') }}</span>
      <BaseSelect
        v-model="echoToAdd.layout"
        :options="layoutOptions"
        class="w-36 h-7"
        :placeholder="t('editor.imageLayoutPlaceholder')"
      />
    </div>

    <!-- 智能压缩 -->
    <div
      v-if="fileToAdd.storage_type !== FILE_STORAGE_TYPE.EXTERNAL"
      class="mb-3 flex items-center"
    >
      <span class="text-[var(--color-text-secondary)]">{{ t('editor.imageSmartCompress') }}</span>
      <BaseSwitch v-model="enableCompressor" />
    </div>

    <!-- 当前上传方式与状态 -->
    <div class="text-[var(--color-text-muted)] text-sm mb-2">
      {{ t('editor.currentUploadMode') }}
      <span class="font-bold">
        {{
          fileToAdd.storage_type === FILE_STORAGE_TYPE.EXTERNAL
            ? t('editor.imageSourceExternal')
            : fileToAdd.storage_type === FILE_STORAGE_TYPE.LOCAL
              ? t('editor.imageSourceLocal')
              : t('editor.imageSourceObject')
        }}</span
      >
      {{ !fileUploading ? '' : t('editor.uploadingSuffix') }}
    </div>

    <div class="my-1">
      <!-- 图片上传 -->
      <TheUploader
        v-if="fileToAdd.storage_type !== FILE_STORAGE_TYPE.EXTERNAL"
        :fileStorageType="fileToAdd.storage_type"
        :EnableCompressor="enableCompressor"
        :maxFileSize="IMAGE_MAX_FILE_SIZE"
      />

      <!-- 图片直链 -->
      <div
        v-if="fileToAdd.storage_type === FILE_STORAGE_TYPE.EXTERNAL"
        class="flex items-center gap-2"
      >
        <BaseInput
          v-model="fileToAdd.url"
          class="rounded-lg h-auto flex-1"
          :placeholder="t('editor.imageUrlPlaceholder')"
        />
        <BaseButton
          v-if="fileToAdd.url != ''"
          :icon="Addmore"
          class="w-8 h-8 sm:w-8 sm:h-8 rounded-md shrink-0"
          @click="editorStore.handleAddMoreFile"
          :tooltip="t('editor.addMoreImages')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ImageLayout } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import Url from '@/components/icons/url.vue'
import Upload from '@/components/icons/upload.vue'
import Bucket from '@/components/icons/bucket.vue'
import Addmore from '@/components/icons/addmore.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheUploader from '@/components/advanced/TheUploader.vue'
import { localStg } from '@/utils/storage'
import { useI18n } from 'vue-i18n'

// Mirror the backend's default image upload cap (config.go: ImageMaxSize = 20 MiB)
// so an oversized file is rejected up-front instead of after a wasted upload.
const IMAGE_MAX_FILE_SIZE = 20 * 1024 * 1024

const editorStore = useEditorStore()
const { fileToAdd, fileUploading, echoToAdd } = storeToRefs(editorStore)
const settingStore = useSettingStore()
const { S3Setting } = storeToRefs(settingStore)
const enableCompressor = ref<boolean>(false)
const { t } = useI18n()

const handleSetFileSource = (source: App.Api.File.StorageType) => {
  fileToAdd.value.storage_type = source

  // 记忆上传方式
  localStg.setItem('file_storage_type', source)
}

// 布局选择
const layoutOptions = computed(() => [
  { label: String(t('editor.layoutWaterfall')), value: ImageLayout.WATERFALL },
  { label: String(t('editor.layoutGrid')), value: ImageLayout.GRID },
  { label: String(t('editor.layoutCarousel')), value: ImageLayout.CAROUSEL },
  { label: String(t('editor.layoutHorizontal')), value: ImageLayout.HORIZONTAL },
  { label: String(t('editor.layoutStack')), value: ImageLayout.STACK },
])
</script>

<style scoped>
.editor-image-panel {
  margin: 0.75rem 0;
  padding: 0.75rem;
  border: 1px dashed var(--color-border-strong);
  border-radius: var(--radius-xs);
}
</style>
