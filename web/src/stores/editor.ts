import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { theToast } from '@/utils/toast'
import {
  fetchAddEcho,
  fetchUpdateEcho,
  fetchAddTodo,
} from '@/service/api'
import { Mode, ExtensionType, ImageLayout } from '@/enums/enums'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { useEchoStore, useTodoStore, useInboxStore } from '@/stores'
import { localStg } from '@/utils/storage'
import { getImageSize } from '@/utils/image'
import { getFileToAddUrl } from '@/utils/other'
import {
  createExternalFile,
  globalFileRegistry,
  useFileAttachments,
  useFilePlayer,
} from '@/lib/file'

export const useEditorStore = defineStore('editorStore', () => {
  const echoStore = useEchoStore()
  const todoStore = useTodoStore()
  const inboxStore = useInboxStore()

  //================================================================
  // 编辑器状态控制
  //================================================================
  const ShowEditor = ref<boolean>(true) // 是否显示编辑器

  // ================================================================
  // 主编辑模式
  // ================================================================
  const currentMode = ref<Mode>(Mode.ECH0) // 默认为Echo编辑模式
  const currentExtensionType = ref<ExtensionType>() // 当前扩展类型（可为空）

  //================================================================
  // 编辑状态
  //================================================================
  const isSubmitting = ref<boolean>(false) // 是否正在提交
  const isUpdateMode = ref<boolean>(false) // 是否为编辑更新模式
  const fileUploading = ref<boolean>(false) // 文件是否正在上传

  //================================================================
  // 编辑器数据状态管理(待添加的Echo)
  //================================================================
  const echoToAdd = ref<App.Api.Ech0.EchoToAdd>({
    content: '', // 文字板块
    echo_files: [], // 仅提交给后端的文件引用
    private: false, // 是否私密
    layout: ImageLayout.WATERFALL, // 图片布局方式，默认为 waterfall
    extension: null, // 拓展内容（对于扩展类型所需的数据）
    extension_type: null, // 拓展内容类型（音乐/视频/链接/GITHUB项目）
  })

  const hasContent = computed(() => !!echoToAdd.value.content?.trim()) // 是否已填写内容
  const hasFile = computed(() => filesToAdd.value.length > 0) // 是否已添加文件
  const hasExtension = computed(() => {
    // 适合 Music/Video/Github
    const ext = extensionToAdd.value.extension
    const extType = extensionToAdd.value.extension_type

    // Website 多一层检测
    if (extType === ExtensionType.WEBSITE) {
      const { title, site } = websiteToAdd.value
      return !!title && !!site
    }

    return !!ext && !!extType
  })

  //================================================================
  // 编辑器数据状态管理(待添加的Todo)
  //================================================================
  const todoToAdd = ref<App.Api.Todo.TodoToAdd>({ content: '' })

  //================================================================
  // 辅助Echo的添加变量（文件板块）
  //================================================================
  const fileToAdd = ref<App.Api.Ech0.FileToAdd>({
    url: '', // 文件地址(依据存储方式不同而不同)
    storage_type: FILE_STORAGE_TYPE.LOCAL, // 文件存储方式（local/object/external）
    key: '', // 对应后端 file.key (如果是直链则为空)
  })
  const {
    files: filesToAdd,
    addAttachment,
    resetAttachments,
    removeAttachment,
    validateAttachments,
  } = useFileAttachments() // 最终要添加的文件列表
  const fileIndex = ref<number>(0) // 当前文件索引（用于编辑文件时定位）

  //================================================================
  // 辅助Echo的添加变量（扩展内容板块）
  //================================================================
  const websiteToAdd = ref({ title: '', site: '' }) // 辅助生成扩展内容（网站）的变量
  const videoURL = ref('') // 辅助生成扩展内容（视频）的变量
  const musicURL = ref('') // 辅助生成扩展内容（音乐）的变量
  const githubRepo = ref('') // 辅助生成扩展内容（GitHub项目）的变量
  const extensionToAdd = ref({ extension: '', extension_type: '' }) // 最终要添加的扩展内容
  const tagToAdd = ref<string>('')

  //================================================================
  // 其它状态变量
  //================================================================
  const filePlayer = useFilePlayer()
  const playingFileURL = filePlayer.playingFileUrl // 当前正在播放的文件URL
  const shouldLoadMusic = filePlayer.shouldReload // 是否应该加载音乐（用于控制音乐播放器的加载）

  //================================================================
  // 编辑器功能函数
  //================================================================
  // 设置当前编辑模式
  const setMode = (mode: Mode) => {
    currentMode.value = mode

    if (mode === Mode.Panel) {
      todoStore.setTodoMode(false)
      inboxStore.setInboxMode(false)
    }
  }
  // 切换当前编辑模式
  const toggleMode = () => {
    if (currentMode.value === Mode.ECH0)
      setMode(Mode.Panel) // 切换到面板模式
    else if (
      currentMode.value === Mode.TODO ||
      currentMode.value === Mode.INBOX ||
      currentMode.value === Mode.PlayMusic ||
      currentMode.value === Mode.EXTEN
    )
      setMode(Mode.Panel) // 扩展模式/TODO模式/音乐播放器模式均切换到面板模式
    else setMode(Mode.ECH0) // 其他模式均切换到Echo编辑模式
  }

  // 清空并重置编辑器
  const clearEditor = () => {
    const rememberedStorageType = ref<App.Api.File.StorageType>(
      localStg.getItem<App.Api.File.StorageType>('file_storage_type') ?? FILE_STORAGE_TYPE.LOCAL,
    )

    echoToAdd.value = {
      content: '',
      echo_files: [],
      private: false,
      layout: ImageLayout.WATERFALL,
      extension: null,
      extension_type: null,
      tags: [],
    }
    fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: rememberedStorageType.value,
      key: '',
    }
    resetAttachments([])
    videoURL.value = ''
    musicURL.value = ''
    githubRepo.value = ''
    extensionToAdd.value = { extension: '', extension_type: '' }
    tagToAdd.value = ''
    todoToAdd.value = { content: '' }
  }

  const handleRefreshPlayingFile = async () => {
    await filePlayer.refreshPlayingFile()
  }

  const setCurrentPlayingFile = async (fileId: string) => {
    await filePlayer.setPlayingFile(fileId)
  }

  const clearCurrentPlayingFile = () => {
    filePlayer.clearPlayingFile()
  }

  //===============================================================
  // 文件模式功能函数
  //===============================================================
  // 添加更多文件
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
        // 图片尺寸探测失败不应阻断写入，否则会出现“上传成功但无预览”。
      }
    }

    // URL 模式先在后端落一条 external file，拿到 file_id 后才能发布。
    if (fileToAdd.value.storage_type === FILE_STORAGE_TYPE.EXTERNAL && !fileToAdd.value.id) {
      const externalUrl = String(fileToAdd.value.url || '').trim()
      if (!externalUrl) {
        theToast.error('图片链接不能为空')
        return
      }

      const created = await createExternalFile({
        url: externalUrl,
        category: FILE_CATEGORY.IMAGE,
        width: width,
        height: height,
      })
      if (!created.id) {
        theToast.error('直链入库失败，请重试')
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

  const handleUppyUploaded = async (files: App.Api.Ech0.FileToAdd[]) => {
    for (const file of files) {
      if (!file.url) {
        theToast.error('上传完成但未拿到可预览地址，请重试')
        continue
      }
      fileToAdd.value = {
        id: file.id,
        url: file.url,
        storage_type: file.storage_type,
        category: file.category,
        content_type: file.content_type,
        key: file.key ? file.key : '',
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
          width: file.width,
          height: file.height,
        })
      }
      await handleAddMoreFile()
    }

    if (isUpdateMode.value && echoStore.echoToUpdate) {
      await handleAddOrUpdateEcho(true) // 仅同步文件
    }
  }

  //===============================================================
  // 私密性切换
  //===============================================================
  const togglePrivate = () => {
    echoToAdd.value.private = !echoToAdd.value.private
  }

  //===============================================================
  // 添加或更新Echo
  //===============================================================
  const handleAddOrUpdateEcho = async (justSyncFiles: boolean) => {
    // 防止重复提交
    if (isSubmitting.value) return
    isSubmitting.value = true

    // 执行添加或更新
    try {
      if (fileUploading.value) {
        theToast.error('图片仍在上传中，请等待上传完成后再发布')
        return
      }

      const valid = validateAttachments({ requireId: true })
      if (!valid.valid) {
        theToast.error(valid.reason || '存在未绑定 file_id 的附件')
        return
      }

      // ========== 添加或更新前的检查和处理 ==========
      // 处理扩展板块
      checkEchoExtension()

      // 回填图片板块（后端只认 echo_files）
      echoToAdd.value.echo_files = filesToAdd.value
        .filter((file) => file.id)
        .map((file, index) => ({
          file_id: String(file.id),
          sort_order: index,
        }))

      // 回填标签板块
      echoToAdd.value.tags = tagToAdd.value?.trim() ? [{ name: tagToAdd.value.trim() }] : []

      // 检查Echo是否为空
      if (checkIsEmptyEcho(echoToAdd.value)) {
        const errMsg = isUpdateMode.value ? '待更新的Echo不能为空！' : '待添加的Echo不能为空！'
        theToast.error(errMsg)
        return
      }

      // ========= 添加模式 =========
      if (!isUpdateMode.value) {
        console.log('adding echo:', echoToAdd.value)
        theToast.promise(fetchAddEcho(echoToAdd.value), {
          loading: '🚀发布中...',
          success: (res) => {
            if (res.code === 1) {
              clearEditor()
              echoStore.refreshEchos()
              setMode(Mode.ECH0)
              echoStore.getTags() // 刷新标签列表
              return '🎉发布成功！'
            } else {
              return '😭发布失败，请稍后再试！'
            }
          },
          error: '😭发布失败，请稍后再试！',
        })

        isSubmitting.value = false
        return
      }

      // ======== 更新模式 =========
      if (isUpdateMode.value) {
        if (!echoStore.echoToUpdate) {
          theToast.error('没有待更新的Echo！')
          return
        }

        // 回填 echoToUpdate
        echoStore.echoToUpdate.content = echoToAdd.value.content
        echoStore.echoToUpdate.private = echoToAdd.value.private
        echoStore.echoToUpdate.layout = echoToAdd.value.layout
        echoStore.echoToUpdate.echo_files = echoToAdd.value.echo_files
        echoStore.echoToUpdate.extension = echoToAdd.value.extension
        echoStore.echoToUpdate.extension_type = echoToAdd.value.extension_type
        echoStore.echoToUpdate.tags = echoToAdd.value.tags

        // 更新 Echo
        theToast.promise(fetchUpdateEcho(echoStore.echoToUpdate), {
          loading: justSyncFiles ? '🔁同步附件中...' : '🚀更新中...',
          success: (res) => {
            if (res.code === 1 && !justSyncFiles) {
              clearEditor()
              echoStore.refreshEchos()
              isUpdateMode.value = false
              echoStore.echoToUpdate = null
              setMode(Mode.ECH0)
              echoStore.getTags() // 刷新标签列表
              return '🎉更新成功！'
            } else if (res.code === 1 && justSyncFiles) {
              return '🔁发现附件更改，已自动更新同步Echo！'
            } else {
              return '😭更新失败，请稍后再试！'
            }
          },
          error: '😭更新失败，请稍后再试！',
        })
      }
    } finally {
      isSubmitting.value = false
    }
  }

  function checkIsEmptyEcho(echo: App.Api.Ech0.EchoToAdd): boolean {
    return (
      !echo.content &&
      (!echo.echo_files || echo.echo_files.length === 0) &&
      !echo.extension &&
      !echo.extension_type
    )
  }

  function checkEchoExtension() {
    // 检查是否有设置扩展类型
    const { extension_type } = extensionToAdd.value
    if (extension_type) {
      // 设置了扩展类型，检查扩展内容是否为空

      switch (extension_type) {
        case ExtensionType.WEBSITE: // 处理网站扩展
          if (!handleWebsiteExtension()) {
            return
          }
          break
        default: // 其他扩展类型暂不处理
          break
      }

      // 同步至echo
      syncEchoExtension()
    } else {
      // 没有设置扩展类型，清空扩展内容
      clearExtension()
    }
  }

  function handleWebsiteExtension(): boolean {
    const { title, site } = websiteToAdd.value

    // 存在标题但无链接
    if (title && !site) {
      theToast.error('网站链接不能为空！')
      return false
    }

    // 如果有链接但没标题，补默认标题
    const finalTitle = title || (site ? '外部链接' : '')
    if (!finalTitle || !site) {
      clearExtension()
      return true
    }

    // 构建扩展内容
    extensionToAdd.value.extension = JSON.stringify({ title: finalTitle, site })
    extensionToAdd.value.extension_type = ExtensionType.WEBSITE

    return true
  }

  // 清空扩展内容
  function clearExtension() {
    extensionToAdd.value.extension = ''
    extensionToAdd.value.extension_type = ''
    echoToAdd.value.extension = null
    echoToAdd.value.extension_type = null
  }

  // 同步Echo的扩展内容
  function syncEchoExtension() {
    const { extension, extension_type } = extensionToAdd.value
    if (extension && extension_type) {
      echoToAdd.value.extension = extension
      echoToAdd.value.extension_type = extension_type
    } else {
      echoToAdd.value.extension = null
      echoToAdd.value.extension_type = null
    }
  }

  //===============================================================
  // 添加Todo
  //===============================================================
  const handleAddTodo = async () => {
    // 防止重复提交
    if (isSubmitting.value) return
    isSubmitting.value = true

    // 执行添加
    try {
      // 检查待办事项是否为空
      console.log('todo content:', todoToAdd.value.content)
      if (todoToAdd.value.content.trim() === '') {
        theToast.error('待办事项不能为空！')
        return
      }

      // 执行添加
      const res = await fetchAddTodo(todoToAdd.value)
      if (res.code === 1) {
        theToast.success('🎉添加成功！')
        todoToAdd.value = { content: '' }
        todoStore.getTodos()
      }
    } finally {
      isSubmitting.value = false
    }
  }

  //===============================================================
  // 退出更新模式
  //===============================================================
  const handleExitUpdateMode = () => {
    isUpdateMode.value = false
    echoStore.echoToUpdate = null
    clearEditor()
    setMode(Mode.ECH0)
    theToast.info('已退出更新模式')
  }

  //===============================================================
  // 处理不同模式下的添加或更新
  //===============================================================
  const handleAddOrUpdate = () => {
    if (todoStore.todoMode) handleAddTodo()
    else handleAddOrUpdateEcho(false)
  }

  const init = () => {
    filePlayer.restoreFromStorage()
    handleRefreshPlayingFile()
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
        width: file.width,
        height: file.height,
      })),
    )
  }

  const removeFileAt = (index: number) => {
    removeAttachment(index)
  }

  return {
    // 状态
    ShowEditor,

    currentMode,
    currentExtensionType,

    isSubmitting,
    isUpdateMode,
    fileUploading,

    echoToAdd,
    todoToAdd,

    hasContent,
    hasFile,
    hasExtension,

    fileToAdd,
    filesToAdd,
    fileIndex,

    websiteToAdd,
    videoURL,
    musicURL,
    githubRepo,
    extensionToAdd,
    tagToAdd,

    playingFileURL,
    shouldLoadMusic,

    // 方法
    init,
    setMode,
    toggleMode,
    clearEditor,
    handleRefreshPlayingFile,
    setCurrentPlayingFile,
    clearCurrentPlayingFile,
    handleAddMoreFile,
    togglePrivate,
    handleAddTodo,
    handleAddOrUpdateEcho,
    handleAddOrUpdate,
    handleExitUpdateMode,
    checkIsEmptyEcho,
    checkEchoExtension,
    syncEchoExtension,
    clearExtension,
    handleUppyUploaded,
    setFilesToAdd,
    removeFileAt,
  }
})
