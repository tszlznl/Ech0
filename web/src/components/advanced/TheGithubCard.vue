<template>
  <div
    v-if="safeGithubURL"
    class="w-full max-w-sm min-w-0 flex justify-center items-center bg-[var(--color-bg-surface)] rounded-lg shadow-sm ring-1 ring-inset ring-[var(--color-border-subtle)] p-2 gap-2 overflow-hidden"
  >
    <a :href="safeGithubURL" target="_blank" rel="noopener noreferrer" class="block w-full min-w-0">
      <div class="flex justify-between items-center">
        <div class="shrink-0 px-6">
          <img
            v-if="CardData?.owner?.avatar_url"
            :src="CardData?.owner?.avatar_url"
            alt="头像"
            class="w-14 h-14 rounded-full shadow"
          />
          <Githubproj v-else class="w-14 h-14" />
        </div>

        <div v-if="CardData" class="py-1 min-w-0">
          <span class="block text-lg font-bold text-[var(--color-text-secondary)] truncate">{{
            CardData?.name || repo
          }}</span>
          <p
            class="text-sm text-[var(--color-text-muted)] font-mono line-clamp-2 break-all"
            :title="CardData?.description"
          >
            {{ CardData?.description }}
          </p>
          <div class="flex justify-start items-center h-auto text-[var(--color-text-muted)]">
            <!-- star -->
            <Star class="w-4 h-4 mr-1" /> <span> {{ CardData?.stargazers_count }} </span>
            <!-- fork -->
            <Fork class="w-4 h-4 mx-1" /> <span> {{ CardData?.forks_count }} </span>
          </div>
        </div>
      </div>
    </a>
  </div>
</template>

<script setup lang="ts">
import Githubproj from '../icons/githubproj.vue'
import Star from '../icons/star.vue'
import Fork from '../icons/fork.vue'
import { fetchGetGithubRepo } from '@/service/api'
import { computed, onMounted, ref } from 'vue'

const githubRepoCache = new Map<string, App.Api.Ech0.GithubCardData | null>()
const githubRepoInFlight = new Map<string, Promise<App.Api.Ech0.GithubCardData | null>>()

const props = defineProps<{
  GithubURL?: string
}>()

const safeGithubURL = computed(() => String(props.GithubURL ?? '').trim())

// 处理GithubURL(提取owner和repo)
const githubUrlSegments = computed(() => safeGithubURL.value.split('/').filter(Boolean))
const owner = computed(() => githubUrlSegments.value.slice(-2)[0] ?? '')
const repo = computed(() => githubUrlSegments.value.slice(-2)[1] ?? '')
const CardData = ref<App.Api.Ech0.GithubCardData>()
const repoKey = computed(() => `${owner.value}/${repo.value}`)

const loadGithubRepo = async () => {
  if (!owner.value || !repo.value) {
    return
  }

  if (githubRepoCache.has(repoKey.value)) {
    const cachedData = githubRepoCache.get(repoKey.value)
    if (cachedData) {
      CardData.value = cachedData
    }
    return
  }

  if (!githubRepoInFlight.has(repoKey.value)) {
    const task = fetchGetGithubRepo({ owner: owner.value, repo: repo.value })
      .then((res) => res ?? null)
      .catch(() => null)
      .finally(() => {
        githubRepoInFlight.delete(repoKey.value)
      })
    githubRepoInFlight.set(repoKey.value, task)
  }

  const repoData = await githubRepoInFlight.get(repoKey.value)
  githubRepoCache.set(repoKey.value, repoData ?? null)
  if (repoData) {
    CardData.value = repoData
  }
}

onMounted(() => {
  void loadGithubRepo()
})
</script>

<style scoped></style>
