<template>
  <div class="flex justify-between items-center py-1 px-3">
    <div class="flex flex-row items-center gap-2 justify-between">
      <!-- <div class="text-xl">👾</div> -->
      <div class="inline-flex rounded-full">
        <img
          :src="logo"
          alt="logo"
          loading="lazy"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-sm object-cover"
        />
      </div>
      <h1 class="text-[var(--color-text-primary)] font-bold sm:text-xl">
        {{ SystemSetting.server_name }}
      </h1>
    </div>

    <div class="flex flex-row items-center gap-2">
      <div
        class="inline-flex items-center rounded-full ring-1 ring-inset ring-[var(--color-border-subtle)] bg-[var(--input-bg-color)] overflow-hidden"
      >
        <button
          type="button"
          :title="`切换主题（下一个：${nextThemeModeLabel}）`"
          :aria-label="`切换主题（下一个：${nextThemeModeLabel}）`"
          class="h-8 px-3 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
          @click="handleThemeToggle"
        >
          <component :is="themeIcon" class="w-5 h-5" />
        </button>
        <button
          type="button"
          title="进入 Zen Mode"
          aria-label="进入 Zen Mode"
          :disabled="isZenMode"
          class="hidden sm:flex items-center justify-center h-8 px-3 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-transparent disabled:hover:text-[var(--color-text-muted)]"
          @click="handleEnterZenMode"
        >
          <Zen class="block w-5 h-5" />
        </button>
        <button
          type="button"
          title="Hello 请求"
          aria-label="Hello 请求"
          class="h-8 px-2.5 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
          @click="handleHelloOnly"
        >
          <Hello class="w-5 h-5" />
        </button>
      </div>
      <!-- Github -->
      <!--
      <div>
        <a href="https://github.com/lin-snow/Ech0" target="_blank" title="Github">
          <Github class="w-6 sm:w-7 h-6 sm:h-7 text-[var(--color-text-muted)]" />
        </a>
      </div>
      -->
    </div>
  </div>
</template>

<script setup lang="ts">
import Hello from '@/components/icons/hello.vue'
import LightIcon from '@/components/icons/light.vue'
import DarkIcon from '@/components/icons/dark.vue'
import AutoIcon from '@/components/icons/auto.vue'
import Zen from '@/components/icons/zen.vue'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { fetchHelloEch0 } from '@/service/api'
import { useSettingStore, useThemeStore, useUserStore, useZenStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const settingStore = useSettingStore()
const themeStore = useThemeStore()
const userStore = useUserStore()
const zenStore = useZenStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)
const { isZenMode } = storeToRefs(zenStore)

const logo = computed(() => {
  if (isLogin.value && user.value?.avatar) {
    return resolveAvatarUrl(user.value.avatar)
  }
  return resolveAvatarUrl(SystemSetting.value?.server_logo)
})

const nextThemeMode = computed(() => {
  if (themeStore.mode === 'system') return 'light'
  if (themeStore.mode === 'light') return 'dark'
  return 'system'
})

const themeIcon = computed(() => {
  if (nextThemeMode.value === 'light') return LightIcon
  if (nextThemeMode.value === 'dark') return DarkIcon
  return AutoIcon
})

const nextThemeModeLabel = computed(() => {
  if (nextThemeMode.value === 'light') return 'Light'
  if (nextThemeMode.value === 'dark') return 'Dark'
  return 'Auto'
})

const getThemeModeLabel = () => {
  if (themeStore.mode === 'light') return 'Light'
  if (themeStore.mode === 'dark') return 'Dark'
  return 'Auto'
}

const handleThemeToggle = async (event: MouseEvent) => {
  await themeStore.toggleTheme(event)
  theToast.success('主题已切换', {
    description: `当前主题为：${getThemeModeLabel()}`,
    duration: 1500,
  })
}

const handleEnterZenMode = () => {
  zenStore.setZenMode(true)
}

const handleHelloOnly = () => {
  const modeText = getThemeModeLabel()

  const hello = ref<App.Api.Ech0.HelloEch0>()

  fetchHelloEch0().then((res) => {
    if (res.code === 1) {
      hello.value = res.data
      theToast.success('你好呀！ 👋', {
        description: `当前版本：v${hello.value.version} | ${modeText}`,
        duration: 2000,
        action: {
          label: 'Github',
          onClick: () => {
            window.open(hello.value?.github, '_blank')
          },
        },
      })
    }
  })
}
</script>

<style scoped></style>
