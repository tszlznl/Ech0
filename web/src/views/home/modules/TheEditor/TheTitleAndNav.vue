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
          v-tooltip="themeToggleTooltip"
          :aria-label="t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel })"
          class="h-8 px-3 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
          @click="handleThemeToggle"
        >
          <component :is="themeIcon" class="w-5 h-5" />
        </button>
        <button
          type="button"
          v-tooltip="t('homeNav.enterZenMode')"
          :aria-label="t('homeNav.enterZenMode')"
          :disabled="isZenMode"
          class="hidden sm:flex items-center justify-center h-8 px-3 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-transparent disabled:hover:text-[var(--color-text-muted)]"
          @click="handleEnterZenMode"
        >
          <Zen class="block w-5 h-5" />
        </button>
        <button
          type="button"
          v-tooltip="t('homeNav.helloRequest')"
          :aria-label="t('homeNav.helloRequest')"
          class="h-8 px-2.5 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
          @click="handleHelloOnly"
        >
          <Hello class="w-5 h-5" />
        </button>
      </div>
      <!-- Github -->
      <!--
      <div>
        <a href="https://github.com/lin-snow/Ech0" target="_blank">
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
import { useI18n } from 'vue-i18n'
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
const { t } = useI18n()

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
  if (nextThemeMode.value === 'light') return String(t('homeNav.themeLight'))
  if (nextThemeMode.value === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeAuto'))
})

const themeToggleTooltip = computed(() => ({
  content: String(t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel.value })),
  triggers: ['hover'],
  hideTriggers: ['hover', 'click'],
}))

const getThemeModeLabel = () => {
  if (themeStore.mode === 'light') return String(t('homeNav.themeLight'))
  if (themeStore.mode === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeAuto'))
}

const handleThemeToggle = async (event: MouseEvent) => {
  await themeStore.toggleTheme(event)
  theToast.success(String(t('homeNav.themeSwitched')), {
    description: String(t('homeNav.themeCurrent', { mode: getThemeModeLabel() })),
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
      theToast.success(String(t('homeNav.helloToastTitle')), {
        description: String(
          t('homeNav.helloToastDesc', { version: hello.value.version, mode: modeText }),
        ),
        duration: 2000,
        action: {
          label: String(t('homeNav.githubAction')),
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
