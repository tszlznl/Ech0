<template>
  <nav class="home-sidebar-nav" aria-label="Primary">
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
</template>

<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'

const { t } = useI18n()
const route = useRoute()
const userStore = useUserStore()
const { isLogin } = storeToRefs(userStore)

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
      route.query.tab === 'publish' || route.query.tab === 'status' || route.query.tab === 'hub'
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
  gap: 0.25rem;
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
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.05);
}
</style>
