<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <article class="echo-detail w-full max-w-sm mx-auto">
    <header class="echo-detail-head">
      <img
        :src="logo"
        alt="logo"
        loading="lazy"
        decoding="async"
        class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] object-cover"
      />
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            {{ SystemSetting.server_name }}
          </h2>
          <Verified class="text-sky-500 w-5 h-5" />
        </div>
        <span class="echo-username text-[var(--color-text-secondary)]">@ {{ echo.username }}</span>
      </div>
    </header>

    <TheEchoMeta :echo="props.echo" />

    <section class="echo-detail-body">
      <template
        v-if="
          props.echo.layout === ImageLayout.GRID ||
          props.echo.layout === ImageLayout.HORIZONTAL ||
          props.echo.layout === ImageLayout.STACK
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
    </section>
  </article>
</template>

<script setup lang="ts">
import Verified from '@/components/icons/verified.vue'
import TheImageGallery from '@/components/advanced/gallery/TheImageGallery.vue'
import TheEchoMeta from '@/components/advanced/echo/cards/TheEchoMeta.vue'
import { computed, defineAsyncComponent } from 'vue'
import { storeToRefs } from 'pinia'
import { useSettingStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { ImageLayout } from '@/enums/enums'
import { getEchoFilesBy } from '@/utils/echo'
import { TheMdPreview } from '@/components/advanced/md'

const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()
const echoImageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)

const settingStore = useSettingStore()

const { SystemSetting } = storeToRefs(settingStore)
const logo = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))
</script>

<style scoped lang="css">
.echo-detail {
  background-color: transparent;
  padding: 0.5rem 0.25rem;
}

.echo-detail-head {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.echo-detail-body {
  padding: 2rem 0 2rem;
}

.echo-username {
  font-family: var(--font-family-display);
}
</style>
