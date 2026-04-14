<template>
  <ExtensionCardShell v-if="safeGithubURL">
    <a :href="safeGithubURL" target="_blank" rel="noopener noreferrer" class="github-card__link">
      <div class="github-card__body">
        <div class="github-avatar-wrap">
          <img
            v-if="cardData?.owner?.avatar_url"
            :src="cardData?.owner?.avatar_url"
            :alt="t('githubCard.avatarAlt')"
            class="w-12 h-12 rounded-full ring-1 ring-[var(--color-border-subtle)] object-cover"
          />
          <Githubproj v-else class="w-12 h-12 text-[var(--color-text-muted)]" />
        </div>

        <div class="github-meta">
          <span class="github-title">{{ cardData?.name || repo }}</span>
          <p class="github-desc line-clamp-2 break-all" v-tooltip="cardData?.description">
            {{ cardData?.description || `${owner}/${repo}` }}
          </p>
          <div v-if="cardData" class="github-stats">
            <span class="github-stat">
              <Star class="w-4 h-4" />
              <span>{{ cardData?.stargazers_count }}</span>
            </span>
            <span class="github-stat-divider" aria-hidden="true"></span>
            <span class="github-stat">
              <Fork class="w-4 h-4" />
              <span>{{ cardData?.forks_count }}</span>
            </span>
          </div>
        </div>
      </div>
    </a>
  </ExtensionCardShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetGithubRepo } from '@/service/api'
import Githubproj from '@/components/icons/githubproj.vue'
import Star from '@/components/icons/star.vue'
import Fork from '@/components/icons/fork.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'

const githubRepoCache = new Map<string, App.Api.Ech0.GithubCardData | null>()
const githubRepoInFlight = new Map<string, Promise<App.Api.Ech0.GithubCardData | null>>()

const props = defineProps<{
  githubUrl?: string
}>()
const { t } = useI18n()

const safeGithubURL = computed(() => String(props.githubUrl ?? '').trim())
const githubUrlSegments = computed(() => safeGithubURL.value.split('/').filter(Boolean))
const owner = computed(() => githubUrlSegments.value.slice(-2)[0] ?? '')
const repo = computed(() => githubUrlSegments.value.slice(-2)[1] ?? '')
const cardData = ref<App.Api.Ech0.GithubCardData>()
const repoKey = computed(() => `${owner.value}/${repo.value}`)

const loadGithubRepo = async () => {
  if (!owner.value || !repo.value) return

  if (githubRepoCache.has(repoKey.value)) {
    const cachedData = githubRepoCache.get(repoKey.value)
    if (cachedData) {
      cardData.value = cachedData
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
    cardData.value = repoData
  }
}

onMounted(() => {
  void loadGithubRepo()
})
</script>

<style scoped>
.github-card__link {
  display: block;
  border-radius: inherit;
}

.github-card__link:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 1px var(--color-focus-ring),
    0 0 0 4px var(--card-focus-ring-outer);
}

.github-card__body {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  min-width: 0;
}

.github-avatar-wrap {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3.25rem;
  height: 3.25rem;
  border-radius: 9999px;
  background: var(--color-bg-muted);
}

.github-meta {
  min-width: 0;
  flex: 1;
}

.github-title {
  display: block;
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.github-desc {
  margin-top: 0.15rem;
  font-size: 0.8rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  line-height: 1.45;
}

.github-stats {
  margin-top: 0.45rem;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.github-stat {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.github-stat-divider {
  width: 1px;
  height: 0.75rem;
  background: var(--color-border-subtle);
}
</style>
