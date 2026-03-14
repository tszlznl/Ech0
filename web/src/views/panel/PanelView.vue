<script setup lang="ts">
import PanelPage from './modules/PanelPage.vue'
import { onMounted } from 'vue'
import { theToast } from '@/utils/toast'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

onMounted(() => {
  const url = new URL(window.location.href)
  const bind = url.searchParams.get('bind')
  if (bind === 'success') {
    theToast.success(String(t('oauth2Setting.bindSuccess')), { duration: 5000 })
  } else if (bind === 'failed') {
    theToast.error(String(t('oauth2Setting.bindFailed')))
  }
})
</script>

<template>
  <div class="w-full px-2">
    <PanelPage />
  </div>
</template>

<style></style>
