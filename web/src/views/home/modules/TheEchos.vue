<template>
  <div class="mx-auto px-2 sm:px-4 md:px-6 mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6">
    <!-- Echos - 使用 TransitionGroup 实现入场动画 -->
    <TransitionGroup
      v-if="echoStore.echoList"
      name="list"
      tag="div"
      @before-enter="onBeforeEnter"
      @enter="onEnter"
    >
      <div
        v-for="(echo, index) in echoStore.echoList"
        :key="echo.id"
        :data-index="index"
        class="will-change-transform"
      >
        <TheEchoCard
          :echo="echo"
          :index="index"
          @refresh="handleRefresh"
          @update-like-count="handleUpdateLikeCount"
        />
      </div>
    </TransitionGroup>
    <!-- 加载更多 -->
    <Transition name="fade">
      <div
        v-if="echoStore.hasMore && !echoStore.isLoading"
        class="mb-4 mt-1 -ml-2 flex items-center justify-between echos-toolbar"
      >
        <BaseButton
          @click="handleLoadMore"
          class="rounded-full bg-[var(--btn-bg-color)] !active:bg-[var(--btn-hover-bg-color)] mr-2"
        >
          <span class="text-[var(--btn-text-color)] text-md text-center px-2 py-1">继续装填</span>
        </BaseButton>
        <TheBackTop class="w-8 h-8 p-1" :target="scrollTarget" />
      </div>
    </Transition>
    <!-- 没有更多 -->
    <Transition name="fade">
      <div
        v-if="!echoStore.hasMore && !echoStore.isLoading"
        class="mx-auto my-5 text-center echos-toolbar"
      >
        <p class="text-xl text-[var(--color-text-muted)] flex items-center justify-center">
          没有啦！<Flowers />
        </p>
      </div>
    </Transition>
    <!-- 加载中 -->
    <Transition name="fade">
      <div v-if="echoStore.isLoading" class="mx-auto my-5 text-center echos-toolbar">
        <span class="text-xl text-[var(--color-text-muted)]">加载中...</span>
      </div>
    </Transition>
    <!-- 自定义页脚 -->
    <div v-if="footerContent" class="text-center">
      <a v-if="footerLink" :href="footerLink" target="_blank" rel="noopener noreferrer">
        <span class="text-[var(--color-text-muted)] text-sm">
          {{ footerContent }}
        </span>
      </a>
      <span v-else class="text-[var(--color-text-muted)] text-sm">
        {{ footerContent }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/TheEchoCard.vue'
import { computed, onMounted } from 'vue'
import { useEchoStore, useSettingStore } from '@/stores'
import BaseButton from '@/components/common/BaseButton.vue'
import { storeToRefs } from 'pinia'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import Flowers from '@/components/icons/flowers.vue'

defineProps<{
  scrollTarget?: HTMLElement | null
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)

// 列表入场动画钩子 - 交错入场效果
const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.opacity = '0'
  element.style.transform = 'translateY(20px)'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  const index = Number(element.dataset.index) || 0
  // 交错延迟：每个元素延迟 50ms，最大延迟 250ms
  const delay = Math.min(index * 50, 250)

  setTimeout(() => {
    element.style.transition = 'opacity 0.3s ease, transform 0.3s ease'
    element.style.opacity = '1'
    element.style.transform = 'translateY(0)'

    // 动画结束后调用 done
    setTimeout(done, 300)
  }, delay)
}

const handleLoadMore = async () => {
  echoStore.current = echoStore.current + 1
  await echoStore.getEchosByPage()
}

// 刷新数据
const handleRefresh = () => {
  echoStore.refreshEchos()
}

// 刷新点赞数据
const handleUpdateLikeCount = (echoId: string) => {
  echoStore.updateLikeCount(echoId, 1)
}

onMounted(async () => {
  // 获取数据
  echoStore.getEchosByPage()
})
</script>

<style scoped>
.echos-toolbar {
  font-family: var(--font-family-display);
}

/* 列表项移动动画 */
.list-move {
  transition: transform 0.3s ease;
}

/* 列表项离开动画 */
.list-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.list-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
