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

/** 参考图：博客=主时间线；第四项为嵌入小组件 */
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
  border-radius: 0.5rem;
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
  background: rgb(0 0 0 / 0.04);
}

.home-sidebar-nav__link--active {
  color: var(--color-text-primary);
  background: rgb(255 255 255 / 0.85);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.05);
}

:global(html.dark) .home-sidebar-nav__link--active {
  background: var(--color-bg-surface);
}
</style>
