<template>
  <div>
    <!-- 音乐分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.MUSIC">
      <h2 class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.musicShare') }}
      </h2>
      <p class="text-[var(--color-text-muted)] text-sm">{{ t('editor.musicSupportHint') }}</p>
      <p class="text-[var(--color-text-muted)] text-sm mb-1">
        {{ t('editor.musicVipHint') }}
      </p>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        :placeholder="t('editor.musicUrlPlaceholder')"
      />
      <div
        v-if="
          editorStore.extensionToAdd.extension.length > 0 &&
          editorStore.extensionToAdd.extension_type === ExtensionType.MUSIC
        "
        class="mt-1 text-[var(--color-text-muted)] text-md"
      >
        {{ t('editor.parseResult') }}:
        <span
          v-if="parseMusicURL(editorStore.extensionToAdd.extension)"
          class="text-[var(--color-accent)]"
          >{{ t('editor.parseSuccess') }}</span
        >
        <span v-else class="text-[var(--color-danger)]">{{ t('editor.parseFailed') }}</span>
      </div>
    </div>
    <!-- Bilibili/YouTube视频分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.VIDEO">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.videoShare') }}
      </div>
      <div class="text-[var(--color-text-muted)] mb-1">{{ t('editor.videoExtractHint') }}</div>
      <BaseInput
        v-model="editorStore.videoURL"
        class="rounded-lg h-auto w-full my-2"
        :placeholder="t('editor.videoUrlPlaceholder')"
      />
      <div class="text-[var(--color-text-secondary)] my-1">
        {{ t('editor.videoId') }}: {{ editorStore.extensionToAdd.extension }}
      </div>
    </div>
    <!-- Github项目分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.GITHUBPROJ">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.githubShare') }}
      </div>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        :placeholder="t('editor.githubUrlPlaceholder')"
      />
    </div>
    <!-- 网站链接分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.WEBSITE">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.websiteShare') }}
      </div>
      <!-- 网站标题 -->
      <BaseInput
        v-model="editorStore.websiteToAdd.title"
        class="rounded-lg h-auto w-full mb-2"
        :placeholder="t('editor.websiteTitlePlaceholder')"
      />
      <div class="flex items-center gap-2">
        <BaseInput
          v-model="editorStore.websiteToAdd.site"
          class="rounded-lg h-auto flex-1"
          :placeholder="t('editor.websiteUrlPlaceholder')"
        />
        <BaseButton
          class="rounded-md px-3 py-2 text-sm whitespace-nowrap"
          :disabled="isFetchingWebsiteTitle"
          @click="handleFetchWebsiteTitle"
        >
          {{ isFetchingWebsiteTitle ? t('editor.fetchingTitle') : t('editor.fetchTitle') }}
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ExtensionType } from '@/enums/enums'
import { parseMusicURL, extractAndCleanMusicURL } from '@/utils/other' // 导入新函数
import { useEditorStore } from '@/stores'
import { ref, watch } from 'vue' // 从 vue 导入 watch
import { fetchGetWebsiteTitle } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const isFetchingWebsiteTitle = ref(false)
const { t } = useI18n()

const handleFetchWebsiteTitle = async () => {
  const websiteURL = (editorStore.websiteToAdd.site || '').trim()
  if (!websiteURL) {
    theToast.warning(String(t('editor.websiteInputRequired')))
    return
  }

  isFetchingWebsiteTitle.value = true
  try {
    const res = await fetchGetWebsiteTitle(websiteURL)
    if (res.code === 1) {
      editorStore.websiteToAdd.title = res.data
      theToast.success(String(t('editor.fetchTitleSuccess')))
    } else {
      theToast.error(res.msg || String(t('editor.fetchTitleFailed')))
    }
  } catch (error) {
    console.error('Failed to fetch website title', error)
    theToast.error(String(t('editor.fetchTitleFailed')))
  } finally {
    isFetchingWebsiteTitle.value = false
  }
}

// 监听音乐链接输入框的变化
watch(
  () => editorStore.extensionToAdd.extension,
  (newValue: string) => {
    // 只在当前是音乐分享模式，并且输入框有内容时才执行
    if (editorStore.currentExtensionType !== ExtensionType.MUSIC || !newValue) {
      return
    }

    const value = newValue.trim()

    // 🔒 至少看起来像个 URL 再处理，避免打字中途被干扰
    if (!/https?:\/\//i.test(value)) {
      return
    }

    // 尝试提取并清理链接
    const cleanUrl = extractAndCleanMusicURL(value)

    // 如果成功提取到干净的链接，并且这个链接和当前输入框的内容不一样
    // （防止无限循环和重复赋值）
    if (cleanUrl && cleanUrl !== value) {
      editorStore.extensionToAdd.extension = cleanUrl
    }
  },
)
</script>

<style scoped></style>
