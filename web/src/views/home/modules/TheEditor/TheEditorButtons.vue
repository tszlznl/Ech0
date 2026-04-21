<template>
  <div class="editor-actions">
    <div class="editor-actions__left">
      <!-- ShowMore -->
      <BaseButton
        :icon="currentMode === Mode.ECH0 ? Advance : Back"
        @click="handleChangeMode"
        :class="['w-8 h-8 sm:w-9 sm:h-9 rounded-xs'].join(' ')"
        :tooltip="currentMode === Mode.ECH0 ? t('editor.more') : t('editor.backToEditor')"
      />
      <!-- Photo Upload -->
      <BaseButton
        v-if="currentMode === Mode.ECH0"
        :icon="ImageUpload"
        @click="handleAddImageMode"
        class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs"
        :tooltip="t('editor.addImage')"
      />
      <!-- Privacy Set -->
      <BaseButton
        v-if="currentMode === Mode.ECH0"
        :icon="echoToAdd.private ? Private : Public"
        @click="handlePrivate"
        class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs"
        :tooltip="t('editor.togglePrivacy')"
      />
      <!-- Tag Add or Select -->
      <div v-if="currentMode === Mode.ECH0" class="editor-actions__tag">
        <div class="editor-actions__tag-box">
          <span class="text-[var(--color-text-muted)]">#</span>
          <BaseCombobox
            :key="tagOptions.length"
            v-model="tagToAdd"
            :multiple="false"
            :options="tagOptions"
            :placeholder="t('editor.tagPlaceholder')"
            wrapper-class="h-full border-transparent shadow-none bg-transparent rounded-[var(--radius-xs)]"
            input-class="w-20 h-full text-[var(--color-text-secondary)]"
          />
        </div>
      </div>
    </div>

    <div class="editor-actions__right">
      <!-- Published Info -->
      <div v-if="hasContent || hasFile || hasExtension" class="relative group">
        <Info class="w-6 h-6 text-[var(--color-text-muted)] hover:text-[var(--color-text-muted)]" />
        <div class="editor-actions__info-pop">
          <div v-if="infoTooltipLines.length > 0">
            <div class="mb-1 font-medium text-[var(--color-text-secondary)]">
              {{ t('editor.addedLabel') }}
            </div>
            <div
              v-for="line in infoTooltipLines"
              :key="line.label"
              class="flex items-center gap-1 text-[var(--color-text-muted)]"
            >
              <component v-if="line.icon" :is="line.icon" class="w-3.5 h-3.5" />
              <span>{{ line.label }}</span>
            </div>
          </div>
          <div v-else class="text-[var(--color-text-muted)]">{{ t('editor.noContentAdded') }}</div>
        </div>
      </div>

      <!-- Publish -->
      <div
        v-if="
          currentMode !== Mode.Panel && currentMode !== Mode.TagManage && isUpdateMode === false
        "
      >
        <BaseButton
          :icon="Publish"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs editor-actions__cta"
          :tooltip="t('editor.publishEcho')"
        />
      </div>
      <!-- Exit Update -->
      <div v-if="currentMode !== Mode.Panel && isUpdateMode === true">
        <BaseButton
          :icon="ExitUpdate"
          @click="handleExitUpdateMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs editor-actions__cta"
          :tooltip="t('editor.exitUpdateMode')"
        />
      </div>
      <!-- Update -->
      <div v-if="currentMode !== Mode.Panel && isUpdateMode === true">
        <BaseButton
          :icon="Update"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs editor-actions__cta"
          :tooltip="t('editor.updateEcho')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Advance from '@/components/icons/advance.vue'
import ImageUpload from '@/components/icons/image.vue'
import ImageIcon from '@/components/icons/image.vue'
import Public from '@/components/icons/public.vue'
import Private from '@/components/icons/private.vue'
import Publish from '@/components/icons/publish.vue'
import Update from '@/components/icons/update.vue'
import ExitUpdate from '@/components/icons/exitupdate.vue'
import Back from '@/components/icons/back.vue'
import Info from '@/components/icons/info.vue'
import Write from '@/components/icons/write.vue'
import Music from '@/components/icons/music.vue'
import Video from '@/components/icons/video.vue'
import GithubProj from '@/components/icons/githubproj.vue'
import Website from '@/components/icons/website.vue'
import MapPin from '@/components/icons/mappin.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCombobox from '@/components/common/BaseCombobox.vue'
import { Mode, ExtensionType } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { storeToRefs } from 'pinia'
import { useEditorStore, useEchoStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { computed, onMounted, type Component } from 'vue'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const {
  currentMode,
  isUpdateMode,
  echoToAdd,
  fileToAdd,
  tagToAdd,
  hasContent,
  hasFile,
  hasExtension,
  extensionToAdd,
} = storeToRefs(editorStore)
const echoStore = useEchoStore()
const { tagOptions } = storeToRefs(echoStore)
const { t } = useI18n()

onMounted(() => {
  echoStore.ensureTagsLoaded()
})

type TooltipLine = { label: string; icon?: Component }

const infoTooltipLines = computed<TooltipLine[]>(() => {
  const extType = extensionToAdd.value.extension_type || echoToAdd.value.extension?.type
  const extMap: Record<ExtensionType, { label: string; icon: Component }> = {
    [ExtensionType.MUSIC]: { label: String(t('editor.extMusic')), icon: Music },
    [ExtensionType.VIDEO]: { label: String(t('editor.extVideo')), icon: Video },
    [ExtensionType.GITHUBPROJ]: { label: String(t('editor.extGithubProject')), icon: GithubProj },
    [ExtensionType.WEBSITE]: { label: String(t('editor.extWebsiteLink')), icon: Website },
    [ExtensionType.LOCATION]: { label: String(t('editor.extLocation')), icon: MapPin },
  }

  const parts: TooltipLine[] = []
  if (hasContent.value) parts.push({ label: String(t('editor.extText')), icon: Write })
  if (hasFile.value) parts.push({ label: String(t('editor.extImage')), icon: ImageIcon })
  if (hasExtension.value)
    parts.push({
      label:
        extType && extMap[extType as ExtensionType]?.label
          ? extMap[extType as ExtensionType].label
          : String(t('editor.extGeneric')),
      icon:
        extType && extMap[extType as ExtensionType]?.icon
          ? extMap[extType as ExtensionType].icon
          : undefined,
    })

  return parts
})

const handleAddorUpdate = () => {
  editorStore.handleAddOrUpdate()
}

const handleChangeMode = () => {
  editorStore.toggleMode()
}

const handleAddImageMode = () => {
  fileToAdd.value.storage_type = FILE_STORAGE_TYPE.LOCAL

  // 检查localStg中是否有记忆的上传方式
  const rememberedSource = localStg.getItem<App.Api.File.StorageType>('file_storage_type')
  if (rememberedSource) {
    fileToAdd.value.storage_type = rememberedSource
  }

  editorStore.setMode(Mode.Image)
}

const handlePrivate = () => {
  editorStore.togglePrivate()
  theToast.info(
    String(
      t('editor.privacySwitched', {
        mode: echoToAdd.value.private ? t('editor.privacyPrivate') : t('editor.privacyPublic'),
      }),
    ),
  )
}

const handleExitUpdateMode = () => {
  editorStore.handleExitUpdateMode()
}
</script>

<style scoped>
.editor-actions {
  --btn-bg-color: var(--md-editor-mini-btn-bg);
  --btn-ring-color: var(--md-editor-actions-ring-color);
  --btn-hover-bg-color: var(--md-editor-actions-hover-bg);
  --btn-hover-border-color: var(--md-editor-actions-hover-border);

  display: flex;
  justify-content: space-between;
  flex-wrap: nowrap;
  gap: 0.6rem;
  align-items: center;
  padding: 0.3rem 0.35rem 0.1rem;
}

.editor-actions__left {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: nowrap;
  overflow: visible;
}

.editor-actions__tag {
  min-width: 0;
  flex: 0 1 7rem;
}

.editor-actions__tag-box {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  height: 2rem;
  border-radius: var(--radius-xs);
  border: 1px dashed var(--md-editor-mini-btn-border);
  background: var(--md-editor-mini-btn-bg);
  padding: 0 0.35rem;
}

@media (width >= 640px) {
  .editor-actions__tag-box {
    height: 2.25rem;
  }
}

.editor-actions__right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.editor-actions__info-pop {
  position: absolute;
  right: 0;
  top: 100%;
  z-index: 10;
  margin-top: 0.5rem;
  white-space: nowrap;
  border-radius: var(--radius-xs);
  border: 1px dashed var(--md-editor-mini-btn-border);
  background: var(--md-editor-mini-bg);
  padding: 0.4rem 0.55rem;
  font-size: 0.75rem;
  box-shadow: var(--md-editor-mini-shell-shadow);
  opacity: 0;
  transform: translateY(0.25rem) scale(0.96);
  pointer-events: none;
  transition: all 0.2s ease-out;
}

.group:hover .editor-actions__info-pop {
  opacity: 1;
  transform: translateY(0) scale(1);
  pointer-events: auto;
}

.editor-actions__cta {
  box-shadow: var(--md-editor-mini-shell-shadow);
}

@media (width <= 639.98px) {
  .editor-actions {
    gap: 0.45rem;
  }
}
</style>
