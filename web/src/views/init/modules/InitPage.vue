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

const router = useRouter()
const initStore = useInitStore()
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
      theToast.success(res.msg || '初始化完成')
      await router.replace({ name: 'auth' })
      return
    }
    if (initStore.initialized || initStore.ownerExists) {
      theToast.success('系统已初始化，正在跳转登录')
      await router.replace({ name: 'auth' })
      return
    }
    theToast.error(res.msg || '初始化失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped></style>
