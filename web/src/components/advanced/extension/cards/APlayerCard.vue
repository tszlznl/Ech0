<template>
  <ExtensionCardShell
    v-if="musicInfo && musicInfo.server !== MusicProvider.APPLE && metingAPI.length > 0"
    size="wide"
    padding="compact"
  >
    <div class="music-player-wrap">
      <ExtensionCardSkeleton v-if="showMetingSkeleton" :min-height="104" />
      <meting-js
        ref="metingRef"
        class="block w-full transition-opacity duration-200"
        :class="isMetingReady ? 'opacity-100' : 'opacity-0'"
        :api="metingAPI"
        :server="musicInfo.server"
        :type="musicInfo.type"
        :id="musicInfo.id"
        :auto="musicAuto"
      >
      </meting-js>
    </div>
  </ExtensionCardShell>
  <ExtensionCardShell
    v-else-if="musicInfo && musicInfo.server === MusicProvider.APPLE && musicInfo.id"
    size="wide"
  >
    <div class="music-player-wrap">
      <ExtensionCardSkeleton v-if="showAppleSkeleton" :min-height="175" />
      <iframe
        allow="autoplay *; encrypted-media *; fullscreen *; clipboard-write"
        frameborder="0"
        height="175"
        class="apple-frame transition-opacity duration-200"
        :class="isAppleReady ? 'opacity-100' : 'opacity-0'"
        sandbox="allow-forms allow-popups allow-same-origin allow-scripts allow-storage-access-by-user-activation allow-top-navigation-by-user-activation"
        :src="`https://embed.music.apple.com/cn/${musicInfo.type}/${musicInfo.id}`"
        @load="isAppleReady = true"
      >
      </iframe>
    </div>
  </ExtensionCardShell>
  <div v-else class="extension-card-invalid">
    <Music class="w-4 h-4" />
    <div class="invalid-copy">
      <p class="invalid-title">该音乐播放源已失效</p>
      <p class="invalid-subtitle">请检查链接或更换平台来源</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import Music from '@/components/icons/music.vue'
import { parseMusicURL } from '@/utils/other'
import { useSettingStore } from '@/stores'
import { ExtensionType, MusicProvider } from '@/enums/enums'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'
import ExtensionCardSkeleton from '../shared/ExtensionCardSkeleton.vue'

const { SystemSetting, loading } = storeToRefs(useSettingStore())

const props = defineProps<{
  echo: {
    extension?: App.Api.Ech0.EchoExtension | null
  }
}>()

const metingRef = ref<HTMLElement | null>(null)
const isMetingReady = ref(false)
const isAppleReady = ref(false)
let metingObserver: MutationObserver | null = null

const musicInfo = computed(() => {
  if (props.echo.extension?.type !== ExtensionType.MUSIC) return null
  return parseMusicURL(props.echo.extension.payload.url)
})

const musicAuto = computed(() => {
  if (props.echo.extension?.type !== ExtensionType.MUSIC) return ''
  return props.echo.extension.payload.url
})

const metingAPI = computed(() => {
  if (!loading.value && SystemSetting.value?.meting_api?.length) {
    return `${SystemSetting.value.meting_api}?server=:server&type=:type&id=:id&auth=:auth&r=:r`
  }
  return 'https://meting.soopy.cn/api?server=:server&type=:type&id=:id&auth=:auth&r=:r'
})

const showMetingSkeleton = computed(() => !isMetingReady.value)
const showAppleSkeleton = computed(() => !isAppleReady.value)
const musicSourceKey = computed(() => {
  const type = musicInfo.value?.type ?? ''
  const id = musicInfo.value?.id ?? ''
  const server = musicInfo.value?.server ?? ''
  return `${server}|${type}|${id}|${musicAuto.value}|${metingAPI.value}`
})

const resetReadyState = () => {
  isMetingReady.value = false
  isAppleReady.value = false
}

const syncMetingReady = () => {
  const host = metingRef.value
  if (!host) return
  if (host.querySelector('.aplayer')) {
    isMetingReady.value = true
  }
}

const observeMetingReady = () => {
  metingObserver?.disconnect()
  metingObserver = null

  const host = metingRef.value
  if (!host) return

  syncMetingReady()
  if (isMetingReady.value) return

  metingObserver = new MutationObserver(() => {
    syncMetingReady()
    if (isMetingReady.value) {
      metingObserver?.disconnect()
      metingObserver = null
    }
  })

  metingObserver.observe(host, { childList: true, subtree: true })
}

let metingDisconnectPatched = false

const patchMetingDisconnectGuard = () => {
  if (typeof window === 'undefined' || metingDisconnectPatched) return
  const ctor = customElements.get('meting-js') as
    | (CustomElementConstructor & {
        prototype?: { disconnectedCallback?: (...args: unknown[]) => void }
      })
    | undefined
  const original = ctor?.prototype?.disconnectedCallback
  if (!ctor?.prototype || !original) return

  ctor.prototype.disconnectedCallback = function (...args: unknown[]) {
    try {
      original.apply(this, args)
    } catch {
      // swallow meting-js teardown error to avoid polluting console during route transitions
    }
  }
  metingDisconnectPatched = true
}

watch(
  musicSourceKey,
  async () => {
    resetReadyState()
    await nextTick()
    observeMetingReady()
  },
  { immediate: true },
)

onMounted(() => {
  patchMetingDisconnectGuard()
  observeMetingReady()
})

onBeforeUnmount(() => {
  metingObserver?.disconnect()
  metingObserver = null
})
</script>

<style scoped>
.music-player-wrap {
  position: relative;
  min-height: 104px;
}

.apple-frame {
  width: 100%;
  max-width: 660px;
  min-height: 175px;
  overflow: hidden;
  border-radius: inherit;
  display: block;
}

.extension-card-invalid {
  width: 100%;
  max-width: 24rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.75rem;
  color: var(--color-text-muted);
}

.invalid-copy {
  min-width: 0;
}

.invalid-title {
  margin: 0;
  font-size: 0.86rem;
  line-height: 1.35;
  color: var(--color-text-secondary);
}

.invalid-subtitle {
  margin: 0.12rem 0 0;
  font-size: 0.74rem;
  line-height: 1.35;
}

:deep(.aplayer) {
  border-radius: calc(var(--radius-md) - 0.15rem);
  border: 1px solid var(--color-border-subtle);
  box-shadow: var(--shadow-sm);
  transition: border-color 0.2s ease;
}
</style>
