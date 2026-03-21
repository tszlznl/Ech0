<template>
  <div class="w-full max-w-sm bg-[var(--color-bg-surface)] h-auto p-5 shadow rounded-lg mx-auto">
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <div>
        <img
          :src="echo.logo"
          alt="logo"
          class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] object-cover"
        />
      </div>
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            <a :href="echo.server_url" target="_blank">{{ echo.server_name }}</a>
          </h2>

          <div>
            <Verified class="text-sky-500 w-5 h-5" />
          </div>
        </div>
        <span class="hub-echo-username text-[var(--color-text-secondary)]"
          >@ {{ echo.username }}
        </span>
      </div>
    </div>

    <div>
      <div class="py-4">
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <div class="mx-auto w-11/12 pl-1 mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>

          <TheImageGallery
            :images="echoImageFiles"
            :baseUrl="echo.server_url"
            :layout="props.echo.layout"
          />
        </template>

        <template v-else>
          <TheImageGallery
            :images="echoImageFiles"
            :baseUrl="echo.server_url"
            :layout="props.echo.layout"
          />

          <div class="mx-auto w-11/12 pl-1 mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <div v-if="props.echo.extension" class="my-4">
          <TheExtensionRenderer :echo="props.echo" />
        </div>
      </div>
    </div>

    <div class="flex items-center justify-between gap-2">
      <div class="min-w-0 flex flex-1 items-center overflow-hidden">
        <div class="min-w-0 truncate whitespace-nowrap text-sm text-slate-500">
          {{ formatDate(props.echo.created_at) }}
        </div>
        <div
          v-if="props.echo.tags?.[0]?.name"
          class="hidden min-w-0 flex-shrink truncate whitespace-nowrap text-xs text-[var(--color-text-muted)] sm:block sm:ml-1"
        >
          #{{ props.echo.tags[0]?.name }}
        </div>
      </div>

      <div ref="menuRef" class="relative flex h-auto flex-none items-center justify-center gap-2">
        <a :href="`${server_url}/echo/${echo_id}`" target="_blank" :title="t('hubEcho.jumpToEcho')">
          <LinkTo class="w-4 h-4" />
        </a>

        <div class="flex items-center justify-end" :title="t('hubEcho.print')">
          <button
            @click="handlePrintEcho()"
            :title="t('hubEcho.print')"
            :class="[
              'transform transition-transform duration-150',
              isPrintAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Print class="w-4 h-4" />
          </button>
        </div>

        <div class="flex items-center justify-end" :title="t('hubEcho.like')">
          <div class="flex items-center gap-1">
            <button
              @click="handleLikeEcho()"
              :title="t('hubEcho.like')"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>

            <span class="text-sm text-[var(--color-text-muted)]">
              {{ fav_count > 99 ? '99+' : fav_count }}
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
import LinkTo from '@/components/icons/linkto.vue'
import Print from '@/components/icons/print.vue'
import TheExtensionRenderer from '@/components/advanced/extension/TheExtensionRenderer.vue'
import TheImageGallery from '@/components/advanced/gallery/TheImageGallery.vue'
import TheMdPreview from '@/components/advanced/TheMdPreview.vue'
import { computed, ref, watch } from 'vue'
import { ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import { useZoneStore } from '@/stores'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

type Echo = App.Api.Hub.Echo

const props = defineProps<{
  echo: Echo
}>()
const zoneStore = useZoneStore()
const router = useRouter()
const { t } = useI18n()

const fav_count = ref<number>(props.echo.fav_count)
const echoImageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)
const server_url = computed(() => props.echo.server_url)
const echo_id = computed(() => props.echo.id)
const isLikeAnimating = ref(false)
const isPrintAnimating = ref(false)
const LIKE_LIST_KEY = computed(() => `${server_url.value}_liked_echo_ids`)

watch(
  () => props.echo.fav_count,
  (next) => {
    fav_count.value = next
  },
)

const handleLikeEcho = async () => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250)

  const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY.value) || []
  if (likedEchoIds.includes(echo_id.value)) {
    theToast.info(String(t('hubEcho.alreadyLiked')))
    return
  }

  const { error, data } = await useFetch<App.Api.Response<null>>(
    `${server_url.value}/api/echo/like/${echo_id.value}`,
  )
    .put()
    .json()

  if (error.value || data.value?.code !== 1) {
    theToast.error(String(t('hubEcho.likeFailed')))
  } else {
    fav_count.value += 1
    likedEchoIds.push(echo_id.value)
    localStg.setItem(LIKE_LIST_KEY.value, likedEchoIds)
    theToast.success(String(t('hubEcho.likeSuccess')))
  }
}

const handlePrintEcho = () => {
  isPrintAnimating.value = true
  setTimeout(() => {
    isPrintAnimating.value = false
  }, 250)

  if (!props.echo.content?.trim()) {
    theToast.info(String(t('hubEcho.printTextOnly')))
    return
  }

  zoneStore.setPendingPrintEcho({
    id: props.echo.id,
    content: props.echo.content,
    created_at: props.echo.created_at,
    tags: props.echo.tags,
    echo_files: props.echo.echo_files,
    extension: props.echo.extension,
  })
  router.push({ name: 'zone' })
}
</script>

<style scoped lang="css">
.hub-echo-username {
  font-family: var(--font-family-display);
}
</style>
