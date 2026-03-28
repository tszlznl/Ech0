<template>
  <div class="flex flex-row items-center justify-between px-2">
    <div class="flex flex-row items-center gap-2">
      <!-- ShowMore -->
      <div>
        <BaseButton
          :icon="currentMode === Mode.ECH0 ? Advance : Back"
          @click="handleChangeMode"
          :class="['w-8 h-8 sm:w-9 sm:h-9 rounded-md'].join(' ')"
          :tooltip="currentMode === Mode.ECH0 ? t('editor.more') : t('editor.backToEditor')"
        />
      </div>
      <!-- Photo Upload -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="ImageUpload"
          @click="handleAddImageMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          :tooltip="t('editor.addImage')"
        />
      </div>
      <!-- Privacy Set -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="echoToAdd.private ? Private : Public"
          @click="handlePrivate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          :tooltip="t('editor.togglePrivacy')"
        />
      </div>
      <!-- Tag Add or Select -->
      <div v-if="currentMode === Mode.ECH0">
        <div
          class="flex items-center justify-between rounded-sm border border-[var(--color-border-subtle)] border-dashed px-1"
        >
          <span class="text-[var(--color-text-muted)]">#</span>
          <BaseCombobox
            :key="tagOptions.length"
            v-model="tagToAdd"
            :multiple="false"
            :options="tagOptions"
            :placeholder="t('editor.tagPlaceholder')"
            wrapper-class="border-transparent shadow-none bg-transparent"
            input-class="w-16 h-7 text-[var(--color-text-secondary)]"
          />
        </div>
      </div>
    </div>

    <div class="flex flex-row items-center gap-2">
      <!-- Published Info -->
      <div v-if="hasContent || hasFile || hasExtension" class="relative group">
        <Info class="w-6 h-6 text-[var(--color-text-muted)] hover:text-[var(--color-text-muted)]" />
        <div
          class="absolute right-0 top-full z-10 mt-2 whitespace-nowrap rounded-md border border-[var(--color-border-subtle)] border-dashed bg-[var(--color-bg-surface)] px-2 py-1 text-xs shadow-md opacity-0 translate-y-1 scale-95 pointer-events-none transition-all duration-200 ease-out group-hover:opacity-100 group-hover:translate-y-0 group-hover:scale-100 group-hover:pointer-events-auto"
        >
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
          currentMode !== Mode.Panel &&
          currentMode !== Mode.TagManage &&
          currentMode !== Mode.INBOX &&
          isUpdateMode === false
        "
      >
        <BaseButton
          :icon="Publish"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          :tooltip="t('editor.publishEcho')"
        />
      </div>
      <!-- Exit Update -->
      <div v-if="currentMode !== Mode.Panel && currentMode !== Mode.INBOX && isUpdateMode === true">
        <BaseButton
          :icon="ExitUpdate"
          @click="handleExitUpdateMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          :tooltip="t('editor.exitUpdateMode')"
        />
      </div>
      <!-- Update -->
      <div v-if="currentMode !== Mode.Panel && isUpdateMode === true">
        <BaseButton
          :icon="Update"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
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
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCombobox from '@/components/common/BaseCombobox.vue'
import { Mode, ExtensionType } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { storeToRefs } from 'pinia'
import { useEditorStore, useEchoStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { computed, type Component } from 'vue'
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

type TooltipLine = { label: string; icon?: Component }

const infoTooltipLines = computed<TooltipLine[]>(() => {
  const extType = extensionToAdd.value.extension_type || echoToAdd.value.extension?.type
  const extMap: Record<ExtensionType, { label: string; icon: Component }> = {
    [ExtensionType.MUSIC]: { label: String(t('editor.extMusic')), icon: Music },
    [ExtensionType.VIDEO]: { label: String(t('editor.extVideo')), icon: Video },
    [ExtensionType.GITHUBPROJ]: { label: String(t('editor.extGithubProject')), icon: GithubProj },
    [ExtensionType.WEBSITE]: { label: String(t('editor.extWebsiteLink')), icon: Website },
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

<style scoped></style>
