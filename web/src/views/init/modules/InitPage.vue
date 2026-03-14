<template>
  <main class="min-h-screen w-full px-4 flex items-center justify-center">
    <div class="w-full max-w-[18rem]">
      <TheInitIntro />
      <TheInitForm
        :username="form.username"
        :password="form.password"
        :submitting="submitting"
        @update:username="form.username = $event"
        @update:password="form.password = $event"
        @submit="onSubmit"
      />
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useInitStore } from '@/stores'
import { theToast } from '@/utils/toast'
import TheInitForm from './TheInitForm.vue'
import TheInitIntro from './TheInitIntro.vue'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const initStore = useInitStore()
const { t } = useI18n()
const submitting = ref(false)
const form = reactive({
  username: '',
  password: '',
})

onMounted(async () => {
  await initStore.getStatus().catch(() => undefined)
  if (initStore.initialized || initStore.ownerExists) {
    await router.replace({ name: 'auth' })
  }
})

const onSubmit = async () => {
  if (!form.username || !form.password || submitting.value) {
    return
  }

  submitting.value = true
  try {
    const res = await initStore.initOwner({
      username: form.username,
      password: form.password,
    })
    if (res.code === 1) {
      theToast.success(res.msg || String(t('init.initDone')))
      await router.replace({ name: 'auth' })
      return
    }
    if (initStore.initialized || initStore.ownerExists) {
      theToast.success(String(t('init.alreadyInitializedRedirect')))
      await router.replace({ name: 'auth' })
      return
    }
    theToast.error(res.msg || String(t('init.initFailed')))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped></style>
