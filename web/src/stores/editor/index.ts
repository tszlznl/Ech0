// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { theToast } from '@/utils/toast'
import { fetchAddEcho, fetchUpdateEcho } from '@/service/api'
import { Mode, ExtensionType, ImageLayout } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { useEchoStore } from '@/stores'
import { localStg } from '@/utils/storage'
import { globalFileRegistry } from '@/lib/file'
import { i18n } from '@/locales'
import router from '@/router'
import { useExtensionModule } from './useExtensionModule'
import { useFilesModule } from './useFilesModule'
import { useDraftModule } from './useDraftModule'
import type { Translate } from './types'

const HOME_TIMELINE_SCROLL_KEY = 'home:timeline:scrollTop'

export const useEditorStore = defineStore('editorStore', () => {
  const echoStore = useEchoStore()
  const t: Translate = (key, params) => String(i18n.global.t(key, params || {}))

  //================================================================
  // 基础顶层状态
  //================================================================
  const ShowEditor = ref<boolean>(true)
  const currentMode = ref<Mode>(Mode.ECH0)
  const currentExtensionType = ref<ExtensionType>()
  const isSubmitting = ref<boolean>(false)
  const isUpdateMode = ref<boolean>(false)

  const echoToAdd = ref<App.Api.Ech0.EchoToAdd>({
    content: '',
    echo_files: [],
    private: false,
    layout: ImageLayout.WATERFALL,
    extension: null,
  })
  const tagToAdd = ref<string[]>([])

  const hasContent = computed(() => !!echoToAdd.value.content?.trim())

  //================================================================
  // 子模块组合
  //================================================================
  const extension = useExtensionModule({ echoToAdd, t })
  const files = useFilesModule({ t })
  const draft = useDraftModule({
    echoToAdd,
    filesToAdd: files.filesToAdd,
    websiteToAdd: extension.websiteToAdd,
    videoURL: extension.videoURL,
    musicURL: extension.musicURL,
    githubRepo: extension.githubRepo,
    extensionToAdd: extension.extensionToAdd,
    locationToAdd: extension.locationToAdd,
    tagToAdd,
    isUpdateMode,
    resetAttachments: files.resetAttachments,
    t,
  })

  // 草稿自动保存监听：依赖所有子模块的 state，必须在全部装配完成后挂
  draft.initDraftWatchers()

  //================================================================
  // 跨域编排:编辑模式、清空、提交、Uppy 上传完成后的条件同步
  //================================================================
  const resetHomeTimelineState = () => {
    echoStore.searchValue = ''
    echoStore.filteredTag = null
    echoStore.isFilteringMode = false
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem(HOME_TIMELINE_SCROLL_KEY)
    }
  }

  const jumpToHomeTimeline = () => {
    void router.push({ name: 'home' }).catch(() => undefined)
  }

  const setMode = (mode: Mode) => {
    currentMode.value = mode
  }

  const toggleMode = () => {
    if (currentMode.value === Mode.ECH0) setMode(Mode.Panel)
    else if (currentMode.value === Mode.EXTEN) setMode(Mode.Panel)
    else setMode(Mode.ECH0)
  }

  const clearEditor = () => {
    const rememberedStorageType =
      localStg.getItem<App.Api.File.StorageType>('file_storage_type') ?? FILE_STORAGE_TYPE.LOCAL

    echoToAdd.value = {
      content: '',
      echo_files: [],
      private: false,
      layout: ImageLayout.WATERFALL,
      extension: null,
      tags: [],
    }
    files.fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: rememberedStorageType,
      key: '',
    }
    files.resetFilesState()
    extension.resetExtensionState()
    tagToAdd.value = []
    draft.clearLocalDraft()
  }

  const togglePrivate = () => {
    echoToAdd.value.private = !echoToAdd.value.private
  }

  const handleUppyUploaded = async (uploadedFiles: App.Api.Ech0.FileToAdd[]) => {
    for (const file of uploadedFiles) {
      if (!file.url) {
        theToast.error(t('editor.uploadNoPreviewUrl'))
        continue
      }
      files.fileToAdd.value = {
        id: file.id,
        url: file.url,
        storage_type: file.storage_type,
        category: file.category,
        content_type: file.content_type,
        key: file.key ? file.key : '',
        size: file.size,
        width: file.width,
        height: file.height,
      }
      if (file.id) {
        globalFileRegistry.upsert({
          id: file.id,
          key: file.key,
          url: file.url,
          category: file.category,
          contentType: file.content_type,
          storageType: file.storage_type,
          size: file.size,
          width: file.width,
          height: file.height,
        })
      }
      await files.handleAddMoreFile()
    }

    if (isUpdateMode.value && echoStore.echoToUpdate) {
      await handleAddOrUpdateEcho(true) // 仅同步文件
    }
  }

  function checkIsEmptyEcho(echo: App.Api.Ech0.EchoToAdd): boolean {
    return (
      !echo.content?.trim() && (!echo.echo_files || echo.echo_files.length === 0) && !echo.extension
    )
  }

  const handleAddOrUpdateEcho = async (justSyncFiles: boolean) => {
    if (isSubmitting.value) return
    isSubmitting.value = true

    try {
      if (files.fileUploading.value) {
        theToast.error(t('editor.fileUploadingWait'))
        return
      }

      const valid = files.validateAttachments({ requireId: true })
      if (!valid.valid) {
        theToast.error(valid.reason || t('editor.attachmentMissingFileId'))
        return
      }

      // 处理扩展板块
      extension.checkEchoExtension()

      // 回填图片板块（后端只认 echo_files）
      echoToAdd.value.echo_files = files.filesToAdd.value
        .filter((file) => file.id)
        .map((file, index) => ({
          file_id: String(file.id),
          sort_order: index,
        }))

      // 回填标签板块
      echoToAdd.value.tags = (tagToAdd.value ?? [])
        .map((name) => name.trim())
        .filter((name) => name.length > 0)
        .map((name) => ({ name }))

      if (checkIsEmptyEcho(echoToAdd.value)) {
        const errMsg = isUpdateMode.value ? t('editor.updateEchoEmpty') : t('editor.addEchoEmpty')
        theToast.error(errMsg)
        return
      }

      // ========= 添加模式 =========
      if (!isUpdateMode.value) {
        theToast.promise(fetchAddEcho(echoToAdd.value), {
          loading: t('editor.publishing'),
          success: (res) => {
            if (res.code === 1) {
              resetHomeTimelineState()
              clearEditor()
              echoStore.refreshEchos()
              setMode(Mode.ECH0)
              echoStore.getTags()
              jumpToHomeTimeline()
              return t('editor.publishSuccess')
            } else {
              return t('editor.publishFailed')
            }
          },
          error: t('editor.publishFailed'),
        })

        isSubmitting.value = false
        return
      }

      // ======== 更新模式 =========
      if (isUpdateMode.value) {
        if (!echoStore.echoToUpdate) {
          theToast.error(t('editor.noEchoToUpdate'))
          return
        }

        echoStore.echoToUpdate.content = echoToAdd.value.content
        echoStore.echoToUpdate.private = echoToAdd.value.private
        echoStore.echoToUpdate.layout = echoToAdd.value.layout
        echoStore.echoToUpdate.echo_files = echoToAdd.value.echo_files
        echoStore.echoToUpdate.extension = echoToAdd.value.extension
        echoStore.echoToUpdate.tags = echoToAdd.value.tags

        theToast.promise(fetchUpdateEcho(echoStore.echoToUpdate), {
          loading: justSyncFiles ? t('editor.syncingFiles') : t('editor.updating'),
          success: (res) => {
            if (res.code === 1 && !justSyncFiles) {
              resetHomeTimelineState()
              clearEditor()
              echoStore.refreshEchos()
              isUpdateMode.value = false
              echoStore.echoToUpdate = null
              setMode(Mode.ECH0)
              echoStore.getTags()
              jumpToHomeTimeline()
              return t('editor.updateSuccess')
            } else if (res.code === 1 && justSyncFiles) {
              return t('editor.updateSuccessWithSync')
            } else {
              return t('editor.updateFailed')
            }
          },
          error: t('editor.updateFailed'),
        })
      }
    } finally {
      isSubmitting.value = false
    }
  }

  const handleAddOrUpdate = () => {
    handleAddOrUpdateEcho(false)
  }

  const handleExitUpdateMode = () => {
    isUpdateMode.value = false
    echoStore.echoToUpdate = null
    clearEditor()
    setMode(Mode.ECH0)
    theToast.info(t('editor.exitUpdateModeSuccess'))
  }

  const init = () => {
    draft.initDraftLifecycle()
  }

  return {
    // ===== 状态 =====
    ShowEditor,

    currentMode,
    currentExtensionType,

    isSubmitting,
    isUpdateMode,
    fileUploading: files.fileUploading,

    echoToAdd,

    hasContent,
    hasFile: files.hasFile,
    hasExtension: extension.hasExtension,

    fileToAdd: files.fileToAdd,
    filesToAdd: files.filesToAdd,
    fileIndex: files.fileIndex,

    websiteToAdd: extension.websiteToAdd,
    videoURL: extension.videoURL,
    musicURL: extension.musicURL,
    githubRepo: extension.githubRepo,
    extensionToAdd: extension.extensionToAdd,
    locationToAdd: extension.locationToAdd,
    tagToAdd,

    // ===== 方法 =====
    init,
    setMode,
    toggleMode,
    clearEditor,
    handleAddMoreFile: files.handleAddMoreFile,
    togglePrivate,
    handleAddOrUpdateEcho,
    handleAddOrUpdate,
    handleExitUpdateMode,
    checkIsEmptyEcho,
    checkEchoExtension: extension.checkEchoExtension,
    syncEchoExtension: extension.syncEchoExtension,
    clearExtension: extension.clearExtension,
    handleUppyUploaded,
    setFilesToAdd: files.setFilesToAdd,
    reorderFilesByIds: files.reorderFilesByIds,
    removeFileAt: files.removeFileAt,
  }
})
