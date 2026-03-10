<template>
  <div v-if="playingFileUrl" class="px-9 md:px-11">
    <!-- 列出所有连接（列出每个连接的头像） -->
    <div
      class="widget rounded-md shadow-sm hover:shadow-md ring-1 ring-[var(--ring-color)] ring-inset p-4"
    >
      <p class="text-[var(--widget-title-color)] font-bold text-lg flex items-center">
        <Album class="mr-1" /> 最近在听：
      </p>
      <div class="flex items-center justify-between my-1">
        <div class="flex items-center gap-4">
          <button
            @click="togglePlay"
            class="w-8 h-8 flex items-center justify-center rounded-full bg-[var(--player-button-bg-color)]! shadow-sm hover:bg-[var(--player-button-hover-bg-color)] text-white transition"
          >
            <span v-if="!isPlaying">
              <Play class="w-5 h-5" :color="'#ee5b5bd9'" />
            </span>
            <span v-else>
              <Pause class="w-6 h-6" :color="'#ee5b5bd9'" />
            </span>
          </button>

          <!-- 提示 -->
          <div v-if="isPlaying" class="text-[var(--text-color-next-500)]">播放中...</div>
          <div v-else class="text-[var(--text-color-next-500)]">暂停中...</div>
        </div>

        <button
          v-if="isPlaying"
          @click="toggleLoop"
          class="w-8 h-8 flex items-center justify-center rounded-full bg-[var(--player-button-bg-color)]! shadow-sm hover:bg-[var(--player-button-hover-bg-color)] text-white transition"
          :class="{ 'bg-[var(--player-button-active-bg-color)]': isLooping }"
        >
          <Repeat class="w-5 h-5" :color="isLooping ? '#ee5b5bd9' : '#888888'" />
        </button>
      </div>

      <audio
        ref="audioRef"
        :src="url"
        @play="toggleIsPlaying(true)"
        @pause="toggleIsPlaying(false)"
        :loop="isLooping"
        preload="none"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { theToast } from '@/utils/toast'
import { useFilePlayer } from '@/lib/file'
import Album from '../icons/album.vue'
import Pause from '../icons/pause.vue'
import Play from '../icons/play.vue'
import Repeat from '../icons/repeat.vue'

const url = ref<string>('')
const isPlaying = ref<boolean>(false)
const isLooping = ref<boolean>(false)
const audioRef = ref<HTMLAudioElement | null>(null)
const filePlayer = useFilePlayer()
const { playingFileUrl, shouldReload, streamUrl } = filePlayer

watch(shouldReload, (newVal) => {
  if (newVal && audioRef.value) {
    url.value = streamUrl.value // 添加时间戳，绕过缓存
    // 强制重新加载音频
    if (isPlaying.value) {
      audioRef.value.pause()
      isPlaying.value = false
    }
    audioRef.value.load()
    audioRef.value.pause()
  }
})

watch(
  playingFileUrl,
  (val) => {
    if (!val) {
      url.value = ''
      return
    }
    url.value = streamUrl.value
  },
  { immediate: true },
)

const toggleIsPlaying = (state: boolean) => {
  isPlaying.value = state
}

function togglePlay() {
  if (!audioRef.value) return
  if (isPlaying.value) {
    audioRef.value.pause()
  } else {
    audioRef.value.play()
  }
}

function toggleLoop() {
  isLooping.value = !isLooping.value
  if (audioRef.value) {
    audioRef.value.loop = isLooping.value
  }
  if (isLooping.value) {
    theToast.info('已开启循环播放')
  } else {
    theToast.info('已关闭循环播放')
  }
}
</script>
