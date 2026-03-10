<template>
  <div class="flex flex-row items-center justify-between px-2">
    <div class="flex flex-row items-center gap-2">
      <!-- ShowMore -->
      <div>
        <BaseButton
          :icon="currentMode === Mode.ECH0 ? Advance : Back"
          @click="handleChangeMode"
          :class="['w-8 h-8 sm:w-9 sm:h-9 rounded-md'].join(' ')"
          title="其它"
        />
      </div>
      <!-- Photo Upload -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="ImageUpload"
          @click="handleAddImageMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="添加图片"
        />
      </div>
      <!-- Privacy Set -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="echoToAdd.private ? Private : Public"
          @click="handlePrivate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="是否私密"
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
            placeholder="标签"
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
            <div class="mb-1 font-medium text-[var(--color-text-secondary)]">已添加：</div>
            <div
              v-for="line in infoTooltipLines"
              :key="line.label"
              class="flex items-center gap-1 text-[var(--color-text-muted)]"
            >
              <component v-if="line.icon" :is="line.icon" class="w-3.5 h-3.5" />
              <span>{{ line.label }}</span>
            </div>
          </div>
          <div v-else class="text-[var(--color-text-muted)]">尚未添加内容</div>
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
          title="发布Echo"
        />
      </div>
      <!-- Exit Update -->
      <div
        v-if="
          currentMode !== Mode.Panel &&
          currentMode !== Mode.INBOX &&
          isUpdateMode === true
        "
      >
        <BaseButton
          :icon="ExitUpdate"
          @click="handleExitUpdateMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="退出更新模式"
        />
      </div>
      <!-- Update -->
      <div
        v-if="
          currentMode !== Mode.Panel &&
          isUpdateMode === true
        "
      >
        <BaseButton
          :icon="Update"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="更新Echo"
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

type TooltipLine = { label: string; icon?: Component }

const infoTooltipLines = computed<TooltipLine[]>(() => {
  const extType = extensionToAdd.value.extension_type || echoToAdd.value.extension?.type
  const extMap: Record<ExtensionType, { label: string; icon: Component }> = {
    [ExtensionType.MUSIC]: { label: '音乐', icon: Music },
    [ExtensionType.VIDEO]: { label: '视频', icon: Video },
    [ExtensionType.GITHUBPROJ]: { label: 'GitHub 项目', icon: GithubProj },
    [ExtensionType.WEBSITE]: { label: '网站链接', icon: Website },
  }

  const parts: TooltipLine[] = []
  if (hasContent.value) parts.push({ label: '文字', icon: Write })
  if (hasFile.value) parts.push({ label: '图片', icon: ImageIcon })
  if (hasExtension.value)
    parts.push({
      label:
        extType && extMap[extType as ExtensionType]?.label
          ? extMap[extType as ExtensionType].label
          : '扩展',
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
  theToast.info('已切换为 ' + (echoToAdd.value.private ? '私密' : '公开') + ' 状态')
}

const handleExitUpdateMode = () => {
  editorStore.handleExitUpdateMode()
}
</script>

<style scoped></style>
