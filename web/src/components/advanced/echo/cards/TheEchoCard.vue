<template>
  <div class="echo-timeline w-full">
    <div class="echo-header-sticky flex justify-between items-center">
      <div class="flex justify-start items-center h-9">
        <div class="flex items-center h-full pr-1">
          <div class="timeline-marker" :class="{ 'is-first': props.index === 0 }">
            <div class="w-2 h-2 rounded-full bg-[var(--color-accent)]"></div>
          </div>
          <div
            @click="handleExpandEcho(echo.id)"
            class="flex items-center h-full justify-start leading-none text-sm text-nowrap text-[var(--color-accent)] hover:underline hover:decoration-offset-3 hover:decoration-1 mr-1"
          >
            {{ formatDate(props.echo.created_at) }}
          </div>
        </div>
        <div
          v-if="!showMenu"
          @click="handleFilterByTag"
          class="text-sm text-[var(--color-text-muted)] w-24 px-1 truncate text-nowrap hover:cursor-pointer hover:text-[var(--color-text-muted)] hover:underline hover:decoration-offset-3 hover:decoration-1"
        >
          <span>{{ props.echo.tags ? `#${props.echo.tags[0]?.name}` : '' }}</span>
        </div>
      </div>

      <div
        v-if="userStore.isLogin || props.echo.private"
        ref="menuRef"
        class="relative flex items-center justify-center gap-1 h-auto"
      >
        <div
          v-if="!showMenu"
          @click.stop="toggleMenu"
          class="w-7 h-7 flex items-center justify-center bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-full shadow-sm transition"
        >
          <More class="w-5 h-5" />
        </div>

        <div
          v-if="showMenu"
          class="flex items-center gap-4 bg-[var(--color-bg-surface)] rounded-full px-2 py-1 shadow-sm ring-1 ring-[var(--color-border-subtle)] ring-inset"
        >
          <span v-if="props.echo.private" v-tooltip="t('echoCard.privateStatus')">
            <Lock />
          </span>

          <template v-if="userStore.isLogin">
            <button @click="handleDeleteEcho(props.echo.id)" v-tooltip="t('echoCard.delete')">
              <Roll />
            </button>

            <button @click="handleUpdateEcho()" v-tooltip="t('echoCard.update')">
              <EditEcho />
            </button>
          </template>
        </div>
      </div>
    </div>

    <div class="timeline-content">
      <div class="px-4 py-3">
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID ||
            props.echo.layout === ImageLayout.HORIZONTAL ||
            props.echo.layout === ImageLayout.STACK
          "
        >
          <div class="mx-auto w-11/12 pl-1 mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>

          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />
        </template>

        <template v-else>
          <TheImageGallery :images="echoImageFiles" :layout="props.echo.layout" />

          <div class="mx-auto w-11/12 pl-1 mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <div v-if="props.echo.extension" class="my-2">
          <TheExtensionRenderer :echo="props.echo" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, onBeforeUnmount, computed, defineAsyncComponent } from 'vue'
import { fetchDeleteEcho, fetchGetEchoById } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useUserStore, useEchoStore, useEditorStore } from '@/stores'
import { TheMdPreview } from '@/components/advanced/md'
import Roll from '@/components/icons/roll.vue'
import Lock from '@/components/icons/lock.vue'
import More from '@/components/icons/more.vue'
import EditEcho from '@/components/icons/editecho.vue'
import { useRouter } from 'vue-router'
import { ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useI18n } from 'vue-i18n'

const TheImageGallery = defineAsyncComponent(
  () => import('@/components/advanced/gallery/TheImageGallery.vue'),
)
const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const { openConfirm } = useBaseDialog()
const { t } = useI18n()

const emit = defineEmits(['refresh'])

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
  index?: number
}>()

const userStore = useUserStore()
const echoImageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)

const echoStore = useEchoStore()
const editorStore = useEditorStore()
const router = useRouter()

const handleDeleteEcho = (echoId: string) => {
  openConfirm({
    title: String(t('echoCard.deleteConfirmTitle')),
    description: String(t('echoCard.deleteConfirmDesc')),
    onConfirm: () => {
      fetchDeleteEcho(echoId).then(() => {
        theToast.success(String(t('echoCard.deleteSuccess')))
        emit('refresh')
      })
    },
  })
}

const handleUpdateEcho = async () => {
  if (editorStore.isUpdateMode) {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    theToast.warning(String(t('echoCard.exitUpdateModeFirst')))
    return
  }

  const res = await fetchGetEchoById(String(props.echo.id))
  if (res.code === 1 && res.data) {
    echoStore.echoToUpdate = res.data
  } else {
    echoStore.echoToUpdate = props.echo
  }

  editorStore.isUpdateMode = true
  await router.push({
    name: 'home',
    query: { tab: 'publish' },
  })
}

const handleExpandEcho = (echoId: string) => {
  router.push({
    name: 'echo',
    params: { echoId: echoId },
  })
}

const showMenu = ref(false)
const menuRef = ref<HTMLElement | null>(null)

const toggleMenu = () => {
  showMenu.value = !showMenu.value
}

const handleClickOutside = (event: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
    showMenu.value = false
  }
}

const handleFilterByTag = () => {
  if (
    props.echo.tags &&
    props.echo.tags.length > 0 &&
    props.echo.tags[0] &&
    props.echo.tags[0].id
  ) {
    echoStore.filteredTag = props.echo.tags[0]
    echoStore.isFilteringMode = true
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped lang="css">
.echo-header-sticky {
  position: relative;
  z-index: 1;
  background-color: var(--color-bg-canvas);
  overflow: hidden;
}

.echo-timeline {
  --axis-offset: calc(0.25rem + 1px);
  --axis-line-width: 2px;
  --axis-dot-size: 0.5rem;
  --axis-dot-gap: 0.3rem;
  max-width: 100%;
  overflow-x: clip;
  /* 纵向允许溢出绘制，避免时间线内图片（如照片流 hover 放大）被裁切 */
  overflow-y: visible;
}

.timeline-marker {
  position: relative;
  width: var(--axis-dot-size);
  height: 100%;
  margin-right: 0.5rem;
  margin-left: calc(var(--axis-offset) - (var(--axis-dot-size) / 2));
  display: flex;
  align-items: center;
  justify-content: center;
}

.timeline-content {
  position: relative;
  margin-left: var(--axis-offset);
  max-width: 100%;
  min-width: 0;
  overflow-x: clip;
  overflow-y: visible;
}

.timeline-content::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: var(--axis-line-width);
  background-color: var(--color-border-subtle);
  pointer-events: none;
}
</style>
