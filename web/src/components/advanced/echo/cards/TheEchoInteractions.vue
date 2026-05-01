<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <section class="echo-interactions w-full max-w-sm mx-auto">
    <TheComment>
      <template #title-actions>
        <div class="echo-interactions-actions">
          <TheShareEchoPanel :echo-id="props.echo.id" :echo-content="props.echo.content" />
          <button
            type="button"
            class="echo-interactions-like"
            v-tooltip="t('echoDetail.like')"
            @click="handleLikeEcho(props.echo.id)"
          >
            <span
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </span>
            <span class="text-xs text-[var(--color-text-muted)]">
              {{ props.echo.fav_count > 99 ? '99+' : props.echo.fav_count }}
            </span>
          </button>
        </div>
      </template>
    </TheComment>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import GrayLike from '@/components/icons/graylike.vue'
import TheShareEchoPanel from '@/components/advanced/echo/cards/TheShareEchoPanel.vue'
import TheComment from '@/components/advanced/TheComment.vue'
import { fetchLikeEcho } from '@/service/api'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'

const { t } = useI18n()

const props = defineProps<{
  echo: App.Api.Ech0.Echo
}>()

const emit = defineEmits<{
  (e: 'updateLikeCount', echoId: string): void
}>()

const isLikeAnimating = ref(false)

const LIKE_LIST_KEY = 'likedEchoIds'
const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY) || []
const hasLikedEcho = (echoId: string): boolean => likedEchoIds.includes(echoId)

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
</script>

<style scoped>
.echo-interactions {
  margin-top: 1rem;
  border-top: 1px dashed var(--color-border-subtle);
}

.echo-interactions-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.echo-interactions-like {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
}
</style>
