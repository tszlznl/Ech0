<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  统一媒体渲染器：一条 Echo 只含单一类别的 EchoFile（图片 / 音频 / 视频），
  据此分派到对应渲染器——图片走画廊(TheImageGallery)、音频走 TheAudioPlayer、
  视频走 TheVideoPlayer。卡片只需挂一个 <TheMediaPlayer>，无需各自重复分类逻辑。
  注意：本组件只负责“媒体”本身，不接管正文(TheMdPreview)——正文与媒体的
  前后排序(layout)、间距在各卡片按需处理。
-->
<template>
  <TheImageGallery
    v-if="imageFiles.length > 0"
    :images="imageFiles"
    :layout="layout"
    :baseUrl="baseUrl"
    :priority="priority"
  />
  <TheAudioPlayer v-else-if="audioFiles.length > 0" :files="audioFiles" :baseUrl="baseUrl" />
  <TheVideoPlayer v-else-if="videoFiles.length > 0" :files="videoFiles" :baseUrl="baseUrl" />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { getEchoFilesBy } from '@/utils/echo'
import TheAudioPlayer from './audio/TheAudioPlayer.vue'
import TheVideoPlayer from './video/TheVideoPlayer.vue'

// 画廊较重（含 5 种布局 + lightbox），异步加载；音视频渲染器轻量，静态引入。
const TheImageGallery = defineAsyncComponent(() => import('./image/TheImageGallery.vue'))

const props = withDefaults(
  defineProps<{
    /** 本地 Ech0.Echo 或联邦 Hub.Echo；两者都带 echo_files，getEchoFilesBy 通吃。 */
    echo: App.Api.Ech0.Echo | App.Api.Hub.Echo
    /** 图片画廊布局；音视频忽略。 */
    layout?: string
    /** 联邦(hub)卡片场景下把相对 URL 解析到远端服务器。 */
    baseUrl?: string
    /** 当本组媒体是页面 LCP 时设为 true（仅图片相关）。 */
    priority?: boolean
  }>(),
  { priority: false },
)

const imageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)
const audioFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['audio'], dedupeBy: 'id' }),
)
const videoFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['video'], dedupeBy: 'id' }),
)
</script>

<style scoped></style>
