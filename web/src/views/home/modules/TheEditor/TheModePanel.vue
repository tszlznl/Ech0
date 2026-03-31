<template>
  <div
    class="editor-mode-panel p-3 my-3 border border-dashed border-[var(--dash-line-color)] rounded-xs"
  >
    <!-- 扩展附加内容 -->
    <div class="mb-1">
      <h2 class="text-[var(--color-text-muted)] font-bold mb-1">{{ t('editor.extraContent') }}</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 添加音乐 -->
        <BaseButton
          :icon="Music"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addMusic')"
          @click="handleAddExtension(ExtensionType.MUSIC)"
        />
        <!-- 添加视频 -->
        <BaseButton
          :icon="Video"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addVideo')"
          @click="handleAddExtension(ExtensionType.VIDEO)"
        />
        <!-- 添加Github项目 -->
        <BaseButton
          :icon="Githubproj"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addGithubProject')"
          @click="handleAddExtension(ExtensionType.GITHUBPROJ)"
        />
        <!-- 添加网站链接 -->
        <BaseButton
          :icon="Weblink"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addWebsiteLink')"
          @click="handleAddExtension(ExtensionType.WEBSITE)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Weblink from '@/components/icons/weblink.vue'
import Music from '@/components/icons/music.vue'
import Video from '@/components/icons/video.vue'
import Githubproj from '@/components/icons/githubproj.vue'
import BaseButton from '@/components/common/BaseButton.vue'

import { Mode, ExtensionType } from '@/enums/enums'
import { useEditorStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const { t } = useI18n()

const handleAddExtension = (extensiontype: ExtensionType) => {
  editorStore.currentMode = Mode.EXTEN
  editorStore.currentExtensionType = extensiontype
  editorStore.extensionToAdd.extension_type = extensiontype
}
</script>

<style scoped>
.editor-mode-panel {
  border-color: var(--md-editor-mini-border);
  --btn-bg-color: var(--md-editor-mini-btn-bg);
  --btn-ring-color: color-mix(
    in srgb,
    var(--md-editor-mini-btn-border),
    var(--color-bg-canvas) 26%
  );
  --btn-hover-bg-color: color-mix(
    in srgb,
    var(--md-editor-mini-btn-hover-bg),
    var(--color-bg-canvas) 38%
  );
  --btn-hover-border-color: color-mix(
    in srgb,
    var(--md-editor-toolbar-btn-hover-border),
    var(--color-bg-canvas) 20%
  );
}
</style>
