<template>
  <div class="w-full max-w-sm bg-[var(--color-bg-surface)] h-auto p-5 shadow rounded-lg mx-auto">
    <!-- 顶部Logo 和 用户名 -->
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <!-- <div class="text-xl">👾</div> -->
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

    <!-- 图片 && 内容 -->
    <div>
      <div class="py-4">
        <!-- grid 和 horizontal 时，文字在图片上；其他布局（waterfall/carousel/null/undefined）文字在图片下 -->
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <!-- 文字在上 -->
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
          <!-- 图片在上，文字在下（瀑布流 / 单图轮播 等） -->
          <TheImageGallery
            :images="echoImageFiles"
            :baseUrl="echo.server_url"
            :layout="props.echo.layout"
          />

          <div class="mx-auto w-11/12 pl-1 mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <!-- 扩展内容 -->
        <div v-if="props.echo.extension" class="my-4">
          <div v-if="props.echo.extension.type === ExtensionType.MUSIC">
            <TheAPlayerCard :echo="props.echo" />
          </div>
          <div v-if="props.echo.extension.type === ExtensionType.VIDEO">
            <TheVideoCard
              :videoId="props.echo.extension.payload.videoId"
              class="px-2 mx-auto hover:shadow-md"
            />
          </div>
          <TheGithubCard
            v-if="props.echo.extension.type === ExtensionType.GITHUBPROJ"
            :GithubURL="props.echo.extension.payload.repoUrl"
            class="px-2 mx-auto hover:shadow-md"
          />
          <TheWebsiteCard
            v-if="props.echo.extension.type === ExtensionType.WEBSITE"
            :website="props.echo.extension.payload"
            class="px-2 mx-auto hover:shadow-md"
          />
        </div>
      </div>
    </div>

    <!-- 日期时间 && 操作按钮 -->
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

      <!-- 操作按钮 -->
      <div ref="menuRef" class="relative flex h-auto flex-none items-center justify-center gap-2">
        <!-- 跳转 -->
        <a :href="`${server_url}/echo/${echo_id}`" target="_blank" title="跳转至该 Echo">
          <LinkTo class="w-4 h-4" />
        </a>

        <!-- 打印 -->
        <div class="flex items-center justify-end" title="打印">
          <button
            @click="handlePrintEcho()"
            title="打印"
            :class="[
              'transform transition-transform duration-150',
              isPrintAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Print class="w-4 h-4" />
          </button>
        </div>

        <!-- 点赞 -->
        <div class="flex items-center justify-end" title="点赞">
          <div class="flex items-center gap-1">
            <!-- 点赞按钮   -->
            <button
              @click="handleLikeEcho()"
              title="点赞"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>

            <!-- 点赞数量   -->
            <span class="text-sm text-[var(--color-text-muted)]">
              <!-- 如果点赞数不超过99，则显示数字，否则显示99+ -->
              {{ fav_count > 99 ? '99+' : fav_count }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheGithubCard from './TheGithubCard.vue'
import TheVideoCard from './TheVideoCard.vue'
import Verified from '../icons/verified.vue'
import GrayLike from '../icons/graylike.vue'
import LinkTo from '../icons/linkto.vue'
import Print from '../icons/print.vue'
import TheAPlayerCard from './TheAPlayerCard.vue'
import TheWebsiteCard from './TheWebsiteCard.vue'
import TheImageGallery from './TheImageGallery.vue'
import TheMdPreview from './TheMdPreview.vue'
import { computed, ref, watch } from 'vue'
import { ExtensionType, ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import { useZoneStore } from '@/stores'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { useRouter } from 'vue-router'

type Echo = App.Api.Hub.Echo

const props = defineProps<{
  echo: Echo
}>()
const zoneStore = useZoneStore()
const router = useRouter()

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

  // 如果已经点赞过，不再重复点赞
  const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY.value) || []
  if (likedEchoIds.includes(echo_id.value)) {
    theToast.info('你已经点赞过')
    return
  }

  // 调用后端接口，点赞
  const { error, data } = await useFetch<App.Api.Response<null>>(
    `${server_url.value}/api/echo/like/${echo_id.value}`,
  )
    .put()
    .json()

  if (error.value || data.value?.code !== 1) {
    theToast.error('点赞失败')
  } else {
    fav_count.value += 1
    likedEchoIds.push(echo_id.value)
    localStg.setItem(LIKE_LIST_KEY.value, likedEchoIds)
    theToast.success('点赞成功')
  }
}

const handlePrintEcho = () => {
  isPrintAnimating.value = true
  setTimeout(() => {
    isPrintAnimating.value = false
  }, 250)

  if (!props.echo.content?.trim()) {
    theToast.info('仅支持带有文本内容的 print')
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
