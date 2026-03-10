import { computed, ref } from 'vue'
import { fileApiAdapter } from '../api/adapter'
import type { FileApiAdapter } from '../api/adapter'

const PLAYER_FILE_ID_KEY = 'playing_file_id'

export class StreamPlayer {
  readonly playingFileId = ref<string>('')
  readonly playingFileUrl = ref<string>('')
  readonly shouldReload = ref<boolean>(true)

  constructor(private adapter: FileApiAdapter = fileApiAdapter) {}

  get streamUrl() {
    return computed(() =>
      this.playingFileId.value ? this.adapter.buildStreamUrl(this.playingFileId.value) : '',
    )
  }

  restoreFromStorage(storage = localStorage) {
    const fileId = storage.getItem(PLAYER_FILE_ID_KEY) || ''
    this.playingFileId.value = fileId
  }

  async refresh(storage = localStorage) {
    const id = this.playingFileId.value || storage.getItem(PLAYER_FILE_ID_KEY) || ''
    if (!id) {
      this.clear(storage)
      return
    }
    const detail = await this.adapter.getFileById(id)
    if (!detail?.id) {
      this.clear(storage)
      return
    }
    this.playingFileId.value = detail.id
    this.playingFileUrl.value = detail.url
    this.shouldReload.value = !this.shouldReload.value
    storage.setItem(PLAYER_FILE_ID_KEY, detail.id)
  }

  async setPlayingFile(fileId: string, storage = localStorage) {
    this.playingFileId.value = fileId
    storage.setItem(PLAYER_FILE_ID_KEY, fileId)
    await this.refresh(storage)
  }

  async clearAndDeleteCurrent(storage = localStorage) {
    const id = this.playingFileId.value || storage.getItem(PLAYER_FILE_ID_KEY) || ''
    if (id) {
      await this.adapter.deleteFileById(id)
    }
    this.clear(storage)
  }

  clear(storage = localStorage) {
    this.playingFileId.value = ''
    this.playingFileUrl.value = ''
    storage.removeItem(PLAYER_FILE_ID_KEY)
    this.shouldReload.value = !this.shouldReload.value
  }
}

export const globalStreamPlayer = new StreamPlayer()
