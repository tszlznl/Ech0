<template>
  <div v-if="echo.extension" class="extension-renderer">
    <APlayerCard v-if="echo.extension.type === ExtensionType.MUSIC" :echo="echo" />
    <VideoCard
      v-else-if="echo.extension.type === ExtensionType.VIDEO"
      :video-id="echo.extension.payload.videoId"
      class="px-2 mx-auto"
    />
    <GithubCard
      v-else-if="
        echo.extension.type === ExtensionType.GITHUBPROJ && echo.extension.payload?.repoUrl
      "
      :github-url="echo.extension.payload.repoUrl"
      class="px-2 mx-auto"
    />
    <WebsiteCard
      v-else-if="echo.extension.type === ExtensionType.WEBSITE"
      :website="echo.extension.payload"
      class="px-2 mx-auto"
    />
  </div>
</template>

<script setup lang="ts">
import { ExtensionType } from '@/enums/enums'
import APlayerCard from './cards/APlayerCard.vue'
import VideoCard from './cards/VideoCard.vue'
import GithubCard from './cards/GithubCard.vue'
import WebsiteCard from './cards/WebsiteCard.vue'

defineProps<{
  echo: {
    extension?: App.Api.Ech0.EchoExtension | null
  }
}>()
</script>

<style scoped>
.extension-renderer {
  width: 100%;
}
</style>
