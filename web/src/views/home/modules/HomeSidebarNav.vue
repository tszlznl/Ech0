<template>
  <div class="home-sidebar-nav">
    <div class="home-sidebar-nav__mobile-row">
      <nav class="home-sidebar-nav__nav" aria-label="Primary">
        <RouterLink
          v-for="item in visibleItems"
          :key="item.id"
          :to="item.to"
          class="home-sidebar-nav__link"
          :class="{ 'home-sidebar-nav__link--active': isItemActive(item) }"
        >
          {{ t(item.labelKey) }}
        </RouterLink>
      </nav>
      <button
        type="button"
        class="home-sidebar-nav__search-trigger"
        :aria-expanded="searchOpenState"
        :aria-label="t('homeTop.searchTitle')"
        @click="searchOpenState = !searchOpenState"
      >
        <Search class="home-sidebar-nav__search-icon" />
      </button>
    </div>
    <div v-if="searchOpenState || isFilteringMode" class="home-sidebar-nav__mobile-filter">
      <TheFilter />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEchoStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import Search from '@/components/icons/search.vue'
import TheFilter from './TheFilter.vue'

const { t } = useI18n()
const route = useRoute()
const userStore = useUserStore()
const echoStore = useEchoStore()
const { isLogin } = storeToRefs(userStore)
const { isFilteringMode } = storeToRefs(echoStore)
const props = defineProps<{
  mobileSearchOpen?: boolean
}>()
const emit = defineEmits<{
  (event: 'update:mobileSearchOpen', value: boolean): void
}>()
const localSearchOpen = ref(false)
const searchOpenState = computed({
  get: () => (props.mobileSearchOpen ?? localSearchOpen.value),
  set: (value: boolean) => {
    if (props.mobileSearchOpen === undefined) {
      localSearchOpen.value = value
      return
    }
    emit('update:mobileSearchOpen', value)
  },
})

const items = [
  {
    id: 'home',
    to: { name: 'home' },
    labelKey: 'homeSidebar.home',
    kind: 'homeTab',
  },
  { id: 'panel', to: { name: 'panel' }, labelKey: 'homeSidebar.panel', kind: 'route' },
  {
    id: 'publish',
    to: { name: 'home', query: { tab: 'publish' } },
    labelKey: 'homeSidebar.publish',
    kind: 'homeTab',
  },
  {
    id: 'status',
    to: { name: 'home', query: { tab: 'status' } },
    labelKey: 'homeSidebar.status',
    kind: 'homeTab',
  },
  {
    id: 'tags',
    to: { name: 'home', query: { tab: 'tags' } },
    labelKey: 'homeSidebar.tags',
    kind: 'homeTab',
  },
  {
    id: 'hub',
    to: { name: 'home', query: { tab: 'hub' } },
    labelKey: 'homeSidebar.plaza',
    kind: 'homeTab',
  },
] as const

const visibleItems = computed(() => items.filter((item) => item.id !== 'publish' || isLogin.value))

const isItemActive = (item: (typeof items)[number]) => {
  if (item.kind === 'homeTab') {
    const tab =
      route.query.tab === 'publish' ||
      route.query.tab === 'status' ||
      route.query.tab === 'tags' ||
      route.query.tab === 'hub'
        ? route.query.tab
        : 'home'
    const itemTab = 'query' in item.to ? item.to.query.tab : 'home'
    return route.name === 'home' && tab === itemTab
  }
  return route.name === item.to.name
}
</script>

<style scoped>
.home-sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.home-sidebar-nav__mobile-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.home-sidebar-nav__nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
  flex: 1 1 auto;
}

.home-sidebar-nav__link {
  display: block;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-xs);
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition:
    color 0.2s,
    background 0.2s;
}

.home-sidebar-nav__link:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.home-sidebar-nav__link--active {
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-bg-muted), var(--color-bg-surface) 90%);
  box-shadow: var(--shadow-soft);
}

.home-sidebar-nav__search-trigger {
  display: none;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.home-sidebar-nav__mobile-filter {
  display: none;
}

.home-sidebar-nav__search-icon {
  width: 1rem;
  height: 1rem;
}

:deep(.home-sidebar-nav__search-icon path) {
  fill: currentColor;
}

@media (max-width: 819.98px) {
  .home-sidebar-nav__nav {
    flex-direction: row;
    gap: 0.2rem;
    overflow-x: auto;
    white-space: nowrap;
    scrollbar-width: none;
  }

  .home-sidebar-nav__nav::-webkit-scrollbar {
    display: none;
  }

  .home-sidebar-nav__link {
    flex: 0 0 auto;
    padding: 0.3rem 0.62rem;
    border-radius: var(--radius-xs);
    font-size: 0.9rem;
    line-height: 1.2;
  }

  .home-sidebar-nav__search-trigger {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    padding: 0.2rem;
    border-radius: 0.375rem;
    color: var(--color-text-muted);
    transition:
      color 0.2s,
      background 0.2s;
  }

  .home-sidebar-nav__search-trigger:hover {
    color: var(--color-text-secondary);
    background: var(--color-bg-muted);
  }

  .home-sidebar-nav__mobile-filter {
    display: block;
    margin: 0;
  }
}
</style>
