import { computed, ref } from 'vue'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
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

  const handleAddMoreFile = async () => {
    let width: number | undefined = fileToAdd.value.width
    let height: number | undefined = fileToAdd.value.height
    if (width === undefined || height === undefined) {
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
        category: FILE_CATEGORY.IMAGE,
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
      category: fileToAdd.value.category,
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
    resetAttachments([])
  }

  return {
    // state
    fileUploading,
    fileToAdd,
    filesToAdd,
    fileIndex,
    // computed
    hasFile,
    // methods
    handleAddMoreFile,
    setFilesToAdd,
    removeFileAt,
    resetFilesState,
    // re-exports used by other modules
    resetAttachments,
    validateAttachments,
  }
}
