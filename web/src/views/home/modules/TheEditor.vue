<template>
  <div
    v-if="ShowEditor"
    class="bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-[var(--radius-lg)] mx-auto shadow-xs hover:shadow-sm"
  >
    <div class="mx-auto w-full px-3 py-4">
      <div class="rounded-[var(--radius-md)] p-2 sm:p-3 mb-1">
        <TheMdEditor v-if="currentMode === Mode.ECH0" class="rounded-[var(--radius-md)]" />
        <TheImageEditor v-if="currentMode === Mode.Image" />
        <TheInboxModeEditor v-if="currentMode === Mode.INBOX" />
        <TheModePanel v-if="currentMode === Mode.Panel" />
        <TheExtensionEditor v-if="currentMode === Mode.EXTEN" />
        <TheTagsManager v-if="currentMode === Mode.TagManage" />
      </div>

      <TheEditorButtons />
      <TheEditorImage />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEditorImage from './TheEditor/TheEditorImage.vue'
import TheEditorButtons from './TheEditor/TheEditorButtons.vue'

import { theToast } from '@/utils/toast'
import { defineAsyncComponent, watch } from 'vue'
import { useEchoStore, useEditorStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Mode, ExtensionType, ImageLayout } from '@/enums/enums'
import { getEchoFiles } from '@/utils/echo'
import { useI18n } from 'vue-i18n'

const TheMdEditor = defineAsyncComponent(() => import('@/components/advanced/md/TheMdEditor.vue'))
const TheModePanel = defineAsyncComponent(() => import('./TheEditor/TheModePanel.vue'))
const TheImageEditor = defineAsyncComponent(() => import('./TheEditor/TheImageEditor.vue'))
const TheInboxModeEditor = defineAsyncComponent(() => import('./TheEditor/TheInboxModeEditor.vue'))
const TheExtensionEditor = defineAsyncComponent(() => import('./TheEditor/TheExtensionEditor.vue'))
const TheTagsManager = defineAsyncComponent(() => import('./TheEditor/TheTagsManager.vue'))

const echoStore = useEchoStore()
const editorStore = useEditorStore()
const { echoToUpdate } = storeToRefs(echoStore)
const {
  ShowEditor,
  currentMode,
  isUpdateMode,
  echoToAdd,
  videoURL,
  extensionToAdd,
  websiteToAdd,
  tagToAdd,
  currentExtensionType,
} = storeToRefs(editorStore)
const { t } = useI18n()

watch(
  () => videoURL.value,
  (newVal) => {
    if (newVal.length > 0) {
      const bvRegex = /(BV[0-9A-Za-z]{10})/
      const ytRegex =
        /(?:https?:\/\/(?:www\.)?)?(?:youtu\.be\/|youtube\.com\/(?:(?:watch)?\?(?:.*&)?v(?:i)?=|(?:embed)\/))([\w-]+)/
      let match = newVal.match(bvRegex)
      if (match) {
        extensionToAdd.value.extension = match[0]
      } else {
        match = newVal.match(ytRegex)
        if (match) {
          extensionToAdd.value.extension = match[1] ?? ''
        } else {
          theToast.error(String(t('editor.videoShareLinkInvalid')))
        }
      }
    }
  },
)

watch(
  () => isUpdateMode.value,
  (newVal) => {
    if (newVal) {
      editorStore.clearEditor()
      echoToAdd.value.content = echoToUpdate.value?.content || ''

      const existingImages = getEchoFiles(echoToUpdate.value)
      if (existingImages.length > 0) {
        editorStore.setFilesToAdd(
          existingImages.map((img) => ({
            id: String(img.id || ''),
            url: img.url || '',
            storage_type: img.storage_type || 'local',
            key: img.key || '',
          })),
        )
      } else {
        editorStore.setFilesToAdd([])
      }

      if (echoToUpdate.value?.extension) {
        currentExtensionType.value = echoToUpdate.value.extension.type as ExtensionType
        extensionToAdd.value.extension_type = echoToUpdate.value.extension.type
        switch (echoToUpdate.value.extension.type) {
          case ExtensionType.MUSIC:
            extensionToAdd.value.extension = echoToUpdate.value.extension.payload.url || ''
            break
          case ExtensionType.VIDEO:
            extensionToAdd.value.extension = echoToUpdate.value.extension.payload.videoId || ''
            videoURL.value = echoToUpdate.value.extension.payload.videoId || ''
            break
          case ExtensionType.GITHUBPROJ:
            extensionToAdd.value.extension = echoToUpdate.value.extension.payload.repoUrl || ''
            break
          case ExtensionType.WEBSITE:
            websiteToAdd.value.title = echoToUpdate.value.extension.payload.title || ''
            websiteToAdd.value.site = echoToUpdate.value.extension.payload.site || ''
            break
        }
      }

      const tags = echoToUpdate.value?.tags
      tagToAdd.value = Array.isArray(tags) && tags.length > 0 ? (tags[0]?.name ?? '') : ''
      echoToAdd.value.private = echoToUpdate.value?.private || false
      echoToAdd.value.layout = echoToUpdate.value?.layout || ImageLayout.WATERFALL
      window.scrollTo({ top: 0, behavior: 'smooth' })
      theToast.info(String(t('editor.enteredUpdateMode')))
    } else {
      echoToUpdate.value = null
    }
  },
)
</script>
