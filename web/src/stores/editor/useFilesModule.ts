// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref } from 'vue'
import { FILE_CATEGORY, FILE_STORAGE_TYPE, type FileCategory } from '@/constants/file'
import { createExternalFile, globalFileRegistry, useFileAttachments } from '@/lib/file'
import { getImageSize } from '@/utils/image'
import { getFileToAddUrl } from '@/utils/other'
import { theToast } from '@/utils/toast'
import type { Translate } from './types'

type FilesModuleDeps = {
  t: Translate
}

export function useFilesModule({ t }: FilesModuleDeps) {
  const fileUploading = ref<boolean>(false)
  const fileIndex = ref<number>(0)

  const fileToAdd = ref<App.Api.Ech0.FileToAdd>({
    url: '',
    storage_type: FILE_STORAGE_TYPE.LOCAL,
    key: '',
  })

  const {
    files: filesToAdd,
    addAttachment,
    resetAttachments,
    removeAttachment,
    validateAttachments,
  } = useFileAttachments()

  const hasFile = computed(() => filesToAdd.value.length > 0)

  // 用户在空状态下选择的上传类别（图片/音频/视频）。有附件后由 mediaCategory 锁定。
  const selectedCategory = ref<FileCategory>(FILE_CATEGORY.IMAGE)

  // 已锁定的媒体类别：一旦有附件，取第一个附件的类别；否则为 null（未锁定）。
  // 一条 Echo 只能包含单一类别的文件，后端亦有硬校验。
  const mediaCategory = computed<FileCategory | null>(
    () => (filesToAdd.value[0]?.category as FileCategory | undefined) ?? null,
  )

  // 编辑器实际生效的类别：锁定优先，否则用用户所选。驱动上传器的 accept / 大小上限。
  const effectiveCategory = computed<FileCategory>(
    () => mediaCategory.value ?? selectedCategory.value,
  )

  // 切换上传类别；已有附件（类别已锁定）时拒绝切换并提示。
  const setSelectedCategory = (category: FileCategory) => {
    if (mediaCategory.value && mediaCategory.value !== category) {
      theToast.info(t('editor.categoryLockedHint'))
      return
    }
    selectedCategory.value = category
  }

  const handleAddMoreFile = async () => {
    // 单类别约束：待加入文件的类别与已锁定类别不一致时拒绝，避免一条 Echo 混挂多类文件。
    const incomingCategory =
      (fileToAdd.value.category as FileCategory | undefined) ?? effectiveCategory.value
    if (mediaCategory.value && mediaCategory.value !== incomingCategory) {
      theToast.error(t('editor.mixedCategoryRejected'))
      return
    }

    let width: number | undefined = fileToAdd.value.width
    let height: number | undefined = fileToAdd.value.height
    // 仅图片需要宽高（画廊比例占位）；音视频跳过探测。
    if (incomingCategory === FILE_CATEGORY.IMAGE && (width === undefined || height === undefined)) {
      try {
        const previewUrl = getFileToAddUrl(fileToAdd.value)
        const size = await getImageSize(previewUrl || fileToAdd.value.url)
        width = size.width
        height = size.height
      } catch {
        // 图片尺寸探测失败不应阻断写入，否则会出现"上传成功但无预览"。
      }
    }

    // URL 模式先在后端落一条 external file，拿到 file_id 后才能发布。
    if (fileToAdd.value.storage_type === FILE_STORAGE_TYPE.EXTERNAL && !fileToAdd.value.id) {
      const externalUrl = String(fileToAdd.value.url || '').trim()
      if (!externalUrl) {
        theToast.error(t('editor.imageUrlRequired'))
        return
      }

      const created = await createExternalFile({
        url: externalUrl,
        category: incomingCategory,
        width: width,
        height: height,
      })
      if (!created.id) {
        theToast.error(t('editor.externalRegisterFailed'))
        return
      }

      fileToAdd.value.id = created.id
      fileToAdd.value.key = created.key
      fileToAdd.value.url = created.url || externalUrl
      globalFileRegistry.upsert(created)
    }

    addAttachment({
      id: fileToAdd.value.id,
      url: fileToAdd.value.url,
      storage_type: fileToAdd.value.storage_type,
      category: incomingCategory,
      content_type: fileToAdd.value.content_type,
      key: fileToAdd.value.key ? fileToAdd.value.key : '',
      size: fileToAdd.value.size,
      width,
      height,
    })

    fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: fileToAdd.value.storage_type
        ? fileToAdd.value.storage_type
        : FILE_STORAGE_TYPE.LOCAL, // 记忆存储方式
      key: '',
    }
  }

  const setFilesToAdd = (files: App.Api.Ech0.FileToAdd[]) => {
    resetAttachments(
      files.map((file) => ({
        id: file.id,
        key: file.key,
        url: file.url,
        storage_type: file.storage_type,
        category: file.category,
        content_type: file.content_type,
        size: file.size,
        width: file.width,
        height: file.height,
      })),
    )
  }

  // Reorder a subset of filesToAdd by id, leaving entries whose id is NOT in
  // `orderedIds` anchored at their original positions. Used by the uploader's
  // drag-to-reorder so that mixing uploaded files with EXTERNAL/URL-mode entries
  // in the same editing session doesn't clobber the latter.
  const reorderFilesByIds = (orderedIds: string[]) => {
    if (orderedIds.length === 0) return
    const idSet = new Set(orderedIds)
    const rank = new Map(orderedIds.map((id, i) => [id, i]))
    const current = filesToAdd.value
    const positions: number[] = []
    const managed: App.Api.Ech0.FileToAdd[] = []
    current.forEach((file, idx) => {
      if (file.id && idSet.has(file.id)) {
        positions.push(idx)
        managed.push(file)
      }
    })
    if (managed.length === 0) return
    managed.sort((a, b) => (rank.get(a.id!) ?? 0) - (rank.get(b.id!) ?? 0))
    const next = current.slice()
    positions.forEach((pos, i) => {
      next[pos] = managed[i]
    })
    setFilesToAdd(next)
  }

  const removeFileAt = (index: number) => {
    removeAttachment(index)
  }

  const resetFilesState = () => {
    fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: fileToAdd.value.storage_type
        ? fileToAdd.value.storage_type
        : FILE_STORAGE_TYPE.LOCAL,
      key: '',
    }
    selectedCategory.value = FILE_CATEGORY.IMAGE
    resetAttachments([])
  }

  return {
    // state
    fileUploading,
    fileToAdd,
    filesToAdd,
    fileIndex,
    selectedCategory,
    // computed
    hasFile,
    mediaCategory,
    effectiveCategory,
    // methods
    handleAddMoreFile,
    setSelectedCategory,
    setFilesToAdd,
    reorderFilesByIds,
    removeFileAt,
    resetFilesState,
    // re-exports used by other modules
    resetAttachments,
    validateAttachments,
  }
}
