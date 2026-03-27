<template>
  <div class="p-3 my-3 border border-dashed border-[var(--dash-line-color)] rounded-lg">
    <!-- 扩展附加内容 -->
    <div class="mb-1">
      <h2 class="text-[var(--color-text-muted)] font-bold mb-1">{{ t('editor.extraContent') }}</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 添加音乐 -->
        <BaseButton
          :icon="Music"
          class="w-7 h-7 rounded-md"
          :tooltip="t('editor.addMusic')"
          @click="handleAddExtension(ExtensionType.MUSIC)"
        />
        <!-- 添加视频 -->
        <BaseButton
          :icon="Video"
          class="w-7 h-7 rounded-md"
          :tooltip="t('editor.addVideo')"
          @click="handleAddExtension(ExtensionType.VIDEO)"
        />
        <!-- 添加Github项目 -->
        <BaseButton
          :icon="Githubproj"
          class="w-7 h-7 rounded-md"
          :tooltip="t('editor.addGithubProject')"
          @click="handleAddExtension(ExtensionType.GITHUBPROJ)"
        />
        <!-- 添加网站链接 -->
        <BaseButton
          :icon="Weblink"
          class="w-7 h-7 rounded-md"
          :tooltip="t('editor.addWebsiteLink')"
          @click="handleAddExtension(ExtensionType.WEBSITE)"
        />
      </div>
    </div>

    <!-- 模式切换 -->
    <div class="mb-1">
      <h2 class="text-[var(--color-text-muted)] font-bold mb-1">{{ t('editor.modeSwitch') }}</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 打开 收件箱 模式 -->
        <BaseButton
          :icon="Inbox"
          @click="handleInbox"
          class="w-7 h-7 rounded-md"
          :tooltip="t('editor.inboxMode')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Weblink from '@/components/icons/weblink.vue'
import Music from '@/components/icons/music.vue'
import Inbox from '@/components/icons/inbox.vue'
import Video from '@/components/icons/video.vue'
import Githubproj from '@/components/icons/githubproj.vue'
import BaseButton from '@/components/common/BaseButton.vue'

import { theToast } from '@/utils/toast'
import { Mode, ExtensionType } from '@/enums/enums'
import { useEditorStore, useInboxStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const { setInboxMode } = useInboxStore()
const { t } = useI18n()

const handleAddExtension = (extensiontype: ExtensionType) => {
  editorStore.currentMode = Mode.EXTEN
  editorStore.currentExtensionType = extensiontype
  editorStore.extensionToAdd.extension_type = extensiontype
}

const handleInbox = () => {
  setInboxMode(true)
  editorStore.currentMode = Mode.INBOX
  theToast.info(String(t('editor.switchedToInboxMode')))
}
</script>

<style scoped></style>
