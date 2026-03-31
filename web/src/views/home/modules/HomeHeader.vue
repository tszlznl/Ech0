<template>
  <div class="home-header">
    <div class="home-header__brand">
      <div class="home-header__logo-wrap">
        <img
          :src="logo"
          alt=""
          loading="lazy"
          class="home-header__logo"
        />
      </div>
      <h1 class="home-header__title">
        {{ SystemSetting.server_name }}
      </h1>
    </div>

    <div class="home-header__actions">
      <div class="home-header__links">
        <a href="/rss" v-tooltip="t('homeTop.rssTitle')" class="home-header__link-icon">
          <Rss class="w-6 h-6" />
        </a>
        <div class="sm:hidden">
          <RouterLink to="/widget" v-tooltip="t('homeTop.widgetTitle')" class="home-header__link-icon">
            <Widget class="w-6 h-6" />
          </RouterLink>
        </div>
        <RouterLink to="/hub" v-tooltip="t('homeTop.hubTitle')" class="home-header__link-icon">
          <HubIcon class="w-6 h-6" />
        </RouterLink>
        <RouterLink to="/panel" v-tooltip="t('homeTop.panelTitle')" class="home-header__link-icon">
          <Panel class="w-6 h-6" />
        </RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Panel from '@/components/icons/panel.vue'
import Rss from '@/components/icons/rss.vue'
import HubIcon from '@/components/icons/hub.vue'
import Widget from '@/components/icons/widget.vue'
import { RouterLink } from 'vue-router'
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore, useUserStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'

const settingStore = useSettingStore()
const userStore = useUserStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)
const { t } = useI18n()

const logo = computed(() => {
  if (isLogin.value && user.value?.avatar) {
    return resolveAvatarUrl(user.value.avatar)
  }
  return resolveAvatarUrl(SystemSetting.value?.server_logo)
})

</script>

<style scoped>
.home-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.25rem;
}

.home-header__brand {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 0;
}

.home-header__logo-wrap {
  flex-shrink: 0;
}

.home-header__logo {
  width: 2.125rem;
  height: 2.125rem;
  border-radius: 9999px;
  object-fit: cover;
  border: 2px solid var(--color-bg-surface);
  box-shadow:
    0 0 0 1px var(--color-border-subtle),
    0 1px 2px rgb(0 0 0 / 0.06);
}

.home-header__title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (min-width: 640px) {
  .home-header__title {
    font-size: 1.25rem;
  }
}

.home-header__actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.home-header__links {
  display: flex;
  align-items: center;
  gap: 0.125rem;
}

.home-header__link-icon {
  display: inline-flex;
  padding: 0.25rem;
  color: var(--color-text-muted);
  border-radius: 0.375rem;
  transition:
    color 0.2s,
    background 0.2s;
}

.home-header__link-icon:hover {
  color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}
</style>
