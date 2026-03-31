<template>
  <div class="w-full max-w-sm bg-[var(--color-bg-surface)] h-auto p-5 shadow rounded-md mx-auto">
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <div>
        <img
          :src="logo"
          alt="logo"
          loading="lazy"
          decoding="async"
          class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] object-cover"
        />
      </div>
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            {{ SystemSetting.server_name }}
          </h2>

          <div>
            <Verified class="text-sky-500 w-5 h-5" />
          </div>
        </div>
        <span class="echo-username text-[var(--color-text-secondary)]">@ {{ echo.username }} </span>
      </div>
    </div>

    <div>
      <div class="py-4">
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <div class="mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>

          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />
        </template>

        <template v-else>
          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />

          <div class="mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <div v-if="props.echo.extension" class="my-4">
          <TheExtensionRenderer :echo="props.echo" />
        </div>
      </div>
    </div>

    <div class="flex justify-between items-center">
      <div class="flex justify-start items-center h-auto">
        <div class="flex justify-start text-sm text-[var(--color-text-muted)] mr-1">
          {{ formatDate(props.echo.created_at) }}
        </div>
        <div class="text-sm text-[var(--color-text-muted)] w-18 truncate text-nowrap">
          <span>{{ props.echo.tags ? `#${props.echo.tags[0]?.name}` : '' }}</span>
        </div>
      </div>

      <div ref="menuRef" class="relative flex items-center justify-center gap-2 h-auto">
        <div class="flex items-center justify-end" v-tooltip="t('echoDetail.share')">
          <button
            @click="handleShareEcho(props.echo.id)"
            :class="[
              'transform transition-transform duration-150',
              isShareAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Share class="w-4 h-4" />
          </button>
        </div>

        <div class="flex items-center justify-end" v-tooltip="t('echoDetail.print')">
          <button
            @click="handlePrintEcho(props.echo)"
            :class="[
              'transform transition-transform duration-150',
              isPrintAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Print class="w-4 h-4" />
          </button>
        </div>

        <div class="flex items-center justify-end" v-tooltip="t('echoDetail.like')">
          <div class="flex items-center gap-1">
            <button
              @click="handleLikeEcho(props.echo.id)"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>

            <span class="text-sm text-[var(--color-text-muted)]">
              {{ props.echo.fav_count > 99 ? '99+' : props.echo.fav_count }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Verified from '@/components/icons/verified.vue'
import GrayLike from '@/components/icons/graylike.vue'
import Print from '@/components/icons/print.vue'
import Share from '@/components/icons/share.vue'
import TheImageGallery from '@/components/advanced/gallery/TheImageGallery.vue'
import { computed, defineAsyncComponent, ref } from 'vue'
import { fetchLikeEcho } from '@/service/api'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { storeToRefs } from 'pinia'
import { useSettingStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import { TheMdPreview } from '@/components/advanced/md'
import { useI18n } from 'vue-i18n'

const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const emit = defineEmits(['updateLikeCount', 'printEcho'])
const { t } = useI18n()

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()
const echoImageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)

const isLikeAnimating = ref(false)
const isShareAnimating = ref(false)
const isPrintAnimating = ref(false)

const LIKE_LIST_KEY = 'likedEchoIds'
const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY) || []
const hasLikedEcho = (echoId: string): boolean => {
  return likedEchoIds.includes(echoId)
}
const handleLikeEcho = (echoId: string) => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250)

  if (hasLikedEcho(echoId)) {
    theToast.info(String(t('echoDetail.alreadyLiked')))
    return
  }

  fetchLikeEcho(echoId).then((res) => {
    if (res.code === 1) {
      likedEchoIds.push(echoId)
      localStg.setItem(LIKE_LIST_KEY, likedEchoIds)
      emit('updateLikeCount', echoId)
      theToast.info(String(t('echoDetail.likeSuccess')))
    }
  })
}

const handleShareEcho = (echoId: string) => {
  isShareAnimating.value = true
  setTimeout(() => {
    isShareAnimating.value = false
  }, 250)

  const shareUrl = `${window.location.origin}/echo/${echoId}\n ———— ${t('echoDetail.shareSuffix')}`
  navigator.clipboard.writeText(shareUrl).then(() => {
    theToast.info(String(t('echoDetail.copied')))
  })
}

const handlePrintEcho = (echo: Echo) => {
  isPrintAnimating.value = true
  setTimeout(() => {
    isPrintAnimating.value = false
  }, 250)

  if (!echo.content?.trim()) {
    theToast.info(String(t('echoDetail.printTextOnly')))
    return
  }

  emit('printEcho', echo)
}

const settingStore = useSettingStore()

const { SystemSetting } = storeToRefs(settingStore)
const logo = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))
</script>

<style scoped lang="css">
.echo-username {
  font-family: var(--font-family-display);
}
</style>
