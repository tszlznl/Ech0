<template>
  <nav class="home-sidebar-nav" aria-label="Primary">
    <RouterLink
      v-for="item in items"
      :key="item.routeName"
      :to="item.to"
      class="home-sidebar-nav__link"
      active-class="home-sidebar-nav__link--active"
    >
      {{ t(item.labelKey) }}
    </RouterLink>
  </nav>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const items = [
  { routeName: 'home', to: { name: 'home' }, labelKey: 'homeSidebar.blog' },
  { routeName: 'zone', to: { name: 'zone' }, labelKey: 'homeSidebar.tags' },
  { routeName: 'hub', to: { name: 'hub' }, labelKey: 'homeSidebar.links' },
  { routeName: 'widget', to: { name: 'widget' }, labelKey: 'homeSidebar.widget' },
] as const
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
  border-radius: var(--radius-sm);
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
