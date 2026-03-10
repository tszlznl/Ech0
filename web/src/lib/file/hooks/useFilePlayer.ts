import { computed } from 'vue'
import { globalStreamPlayer } from '../player/stream-player'

export function useFilePlayer() {
  const streamUrl = computed(() => globalStreamPlayer.streamUrl.value)

  return {
    playingFileId: globalStreamPlayer.playingFileId,
    playingFileUrl: globalStreamPlayer.playingFileUrl,
    shouldReload: globalStreamPlayer.shouldReload,
    streamUrl,
    restoreFromStorage: globalStreamPlayer.restoreFromStorage.bind(globalStreamPlayer),
    refreshPlayingFile: globalStreamPlayer.refresh.bind(globalStreamPlayer),
    setPlayingFile: globalStreamPlayer.setPlayingFile.bind(globalStreamPlayer),
    clearPlayingFile: globalStreamPlayer.clear.bind(globalStreamPlayer),
    clearAndDeleteCurrent: globalStreamPlayer.clearAndDeleteCurrent.bind(globalStreamPlayer),
  }
}
