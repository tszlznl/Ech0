<template>
  <!-- 网易云 / QQ 音乐使用 Meting JS来展示 -->
  <TheExtensionCardShell
    v-if="musicInfo && musicInfo.server !== MusicProvider.APPLE && metingAPI.length > 0 && !loading"
    size="wide"
    padding="compact"
  >
    <meting-js
      class="block w-full"
      :api="metingAPI"
      :server="musicInfo.server"
      :type="musicInfo.type"
      :id="musicInfo.id"
      :auto="musicAuto"
    >
    </meting-js>
  </TheExtensionCardShell>
  <!-- Apple Music 使用官方IFrame -->
  <TheExtensionCardShell
    v-else-if="musicInfo && musicInfo.server === MusicProvider.APPLE && musicInfo.id"
    size="wide"
  >
    <iframe
      allow="autoplay *; encrypted-media *; fullscreen *; clipboard-write"
      frameborder="0"
      height="175"
      class="apple-frame"
      sandbox="allow-forms allow-popups allow-same-origin allow-scripts allow-storage-access-by-user-activation allow-top-navigation-by-user-activation"
      :src="`https://embed.music.apple.com/cn/${musicInfo.type}/${musicInfo.id}`"
    >
    </iframe>
  </TheExtensionCardShell>
  <div v-else class="extension-card-invalid">
    <Music class="w-4 h-4" />
    <div class="invalid-copy">
      <p class="invalid-title">该音乐播放源已失效</p>
      <p class="invalid-subtitle">请检查链接或更换平台来源</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Music from '@/components/icons/music.vue'
import { computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { parseMusicURL } from '@/utils/other'
import { useSettingStore } from '@/stores'
import { ExtensionType, MusicProvider } from '@/enums/enums'
import TheExtensionCardShell from './TheExtensionCardShell.vue'

const { SystemSetting, loading } = storeToRefs(useSettingStore())
type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()

const musicInfo = computed(() => {
  if (props.echo.extension?.type !== ExtensionType.MUSIC) return null
  return parseMusicURL(props.echo.extension.payload.url)
})
const musicAuto = computed(() => {
  if (props.echo.extension?.type !== ExtensionType.MUSIC) return ''
  return props.echo.extension.payload.url
})
const metingAPI = computed(() => {
  if (!loading.value && SystemSetting.value && SystemSetting.value.meting_api.length > 0) {
    return SystemSetting.value.meting_api + '?server=:server&type=:type&id=:id&auth=:auth&r=:r'
  } else {
    return 'https://meting.soopy.cn/api?server=:server&type=:type&id=:id&auth=:auth&r=:r'
  }
})

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

onMounted(() => {
  patchMetingDisconnectGuard()
})
</script>

<style scoped>
.apple-frame {
  width: 100%;
  max-width: 660px;
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
