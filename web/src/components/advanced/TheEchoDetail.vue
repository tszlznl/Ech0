<template>
  <div
    class="w-full max-w-sm bg-[var(--color-bg-surface)] h-auto p-5 shadow rounded-lg mx-auto"
  >
    <!-- 顶部Logo 和 用户名 -->
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <!-- <div class="text-xl">👾</div> -->
      <div>
        <img
          :src="logo"
          alt="logo"
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
        <span class="echo-username text-[var(--color-text-secondary)]"
          >@ {{ echo.username }}
        </span>
      </div>
    </div>

    <!-- 图片 && 内容 -->
    <div>
      <div class="py-4">
        <!-- 根据布局决定文字与图片顺序 -->
        <!-- grid 和 horizontal 时，文字在图片上；其他布局（waterfall/carousel/null/undefined）文字在图片下 -->
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <!-- 文字在上 -->
          <div class="mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>

          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />
        </template>

        <template v-else>
          <!-- 图片在上，文字在下 -->
          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />

          <div class="mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <!-- 扩展内容 -->
        <div v-if="props.echo.extension" class="my-4">
          <div v-if="props.echo.extension_type === ExtensionType.MUSIC">
            <TheAPlayerCard :echo="props.echo" />
          </div>
          <div v-if="props.echo.extension_type === ExtensionType.VIDEO">
            <TheVideoCard :videoId="props.echo.extension" class="px-2 mx-auto hover:shadow-md" />
          </div>
          <TheGithubCard
            v-if="props.echo.extension_type === ExtensionType.GITHUBPROJ"
            :GithubURL="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
          <TheWebsiteCard
            v-if="props.echo.extension_type === ExtensionType.WEBSITE"
            :website="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
        </div>
      </div>
    </div>

    <!-- 日期时间 && 操作按钮 -->
    <div class="flex justify-between items-center">
      <!-- 日期时间 -->
      <div class="flex justify-start items-center h-auto">
        <div class="flex justify-start text-sm text-[var(--color-text-muted)] mr-1">
          {{ formatDate(props.echo.created_at) }}
        </div>
        <!-- 标签 -->
        <div class="text-sm text-[var(--color-text-muted)] w-18 truncate text-nowrap">
          <span>{{ props.echo.tags ? `#${props.echo.tags[0]?.name}` : '' }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div ref="menuRef" class="relative flex items-center justify-center gap-2 h-auto">
        <!-- 分享 -->
        <div class="flex items-center justify-end" title="分享">
          <button
            @click="handleShareEcho(props.echo.id)"
            title="分享"
            :class="[
              'transform transition-transform duration-150',
              isShareAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Share class="w-4 h-4" />
          </button>
        </div>

        <!-- 打印 -->
        <div class="flex items-center justify-end" title="打印">
          <button
            @click="handlePrintEcho(props.echo)"
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
              @click="handleLikeEcho(props.echo.id)"
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
              {{ props.echo.fav_count > 99 ? '99+' : props.echo.fav_count }}
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
import Print from '../icons/print.vue'
import Share from '../icons/share.vue'
import TheAPlayerCard from './TheAPlayerCard.vue'
import TheWebsiteCard from './TheWebsiteCard.vue'
import TheImageGallery from './TheImageGallery.vue'
import { computed, ref } from 'vue'
import { fetchLikeEcho } from '@/service/api'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { storeToRefs } from 'pinia'
import { useSettingStore } from '@/stores'
import { getApiUrl } from '@/service/request/shared'
import { ExtensionType, ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import TheMdPreview from './TheMdPreview.vue'
const emit = defineEmits(['updateLikeCount', 'printEcho'])

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
  }, 250) // 对应 duration-250

  // 检查LocalStorage中是否已经点赞过
  if (hasLikedEcho(echoId)) {
    theToast.info('你已经点赞过了,感谢你的喜欢！')
    return
  }

  fetchLikeEcho(echoId).then((res) => {
    if (res.code === 1) {
      likedEchoIds.push(echoId)
      localStg.setItem(LIKE_LIST_KEY, likedEchoIds)
      // 发送更新事件
      emit('updateLikeCount', echoId)
      theToast.info('点赞成功！')
    }
  })
}

const handleShareEcho = (echoId: string) => {
  isShareAnimating.value = true
  setTimeout(() => {
    isShareAnimating.value = false
  }, 250) // 对应 duration-250

  const shareUrl = `${window.location.origin}/echo/${echoId}\n ———— 来自 Ech0 分享`
  navigator.clipboard.writeText(shareUrl).then(() => {
    theToast.info('已复制到剪贴板！')
  })
}

const handlePrintEcho = (echo: Echo) => {
  isPrintAnimating.value = true
  setTimeout(() => {
    isPrintAnimating.value = false
  }, 250)

  if (!echo.content?.trim()) {
    theToast.info('仅支持打印带有文本内容的 Echo！')
    return
  }

  emit('printEcho', echo)
}

const settingStore = useSettingStore()

const { SystemSetting } = storeToRefs(settingStore)

const apiUrl = getApiUrl()
const logo = ref<string>('/Ech0.svg')
if (
  SystemSetting.value.server_logo &&
  SystemSetting.value.server_logo !== '' &&
  SystemSetting.value.server_logo !== 'Ech0.svg'
) {
  logo.value = `${apiUrl}${SystemSetting.value.server_logo}`
}
</script>

<style scoped lang="css">
.echo-username {
  font-family: var(--font-family-display);
}

</style>
