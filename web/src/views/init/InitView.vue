<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="w-full max-w-md rounded-lg ring-1 ring-[var(--ring-color)] p-6">
      <h1 class="text-xl font-bold mb-2">初始化系统</h1>
      <p class="text-sm mb-6 text-[var(--text-color-next-500)]">创建首个 Owner 账号后即可进入登录页面。</p>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <div>
          <label class="block text-sm mb-1">用户名</label>
          <input
            v-model="form.username"
            type="text"
            class="w-full rounded-md border px-3 py-2 bg-transparent"
            required
          />
        </div>
        <div>
          <label class="block text-sm mb-1">密码</label>
          <input
            v-model="form.password"
            type="password"
            class="w-full rounded-md border px-3 py-2 bg-transparent"
            required
          />
        </div>

        <button
          type="submit"
          class="w-full rounded-md py-2 text-white bg-[var(--button-primary-color)] disabled:opacity-60"
          :disabled="submitting"
        >
          {{ submitting ? '初始化中...' : '创建 Owner' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useInitStore } from '@/stores'
import { theToast } from '@/utils/toast'

const router = useRouter()
const initStore = useInitStore()
const submitting = ref(false)
const form = reactive({
  username: '',
  password: '',
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
    theToast.error(res.msg || '初始化失败')
  } finally {
    submitting.value = false
  }
}
</script>

