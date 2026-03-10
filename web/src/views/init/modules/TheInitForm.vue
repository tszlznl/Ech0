<template>
  <section
    class="w-full rounded-xl border border-[var(--border-color-300)] bg-[var(--bg-color-200)]/70 backdrop-blur-sm p-5 sm:p-6"
  >
    <div class="mb-5">
      <h2 class="text-lg font-bold text-[var(--text-color-next-300)]">Owner 账号</h2>
      <p class="text-sm text-[var(--text-color-next-500)] mt-1">请输入一个用于首次登录后台的管理员账号。</p>
    </div>

    <form class="space-y-4" @submit.prevent="emit('submit')">
      <BaseInput
        :model-value="username"
        @update:model-value="emit('update:username', String($event))"
        type="text"
        placeholder="请输入用户名"
        autocomplete="username"
        required
      />

      <BaseInput
        :model-value="password"
        @update:model-value="emit('update:password', String($event))"
        type="password"
        placeholder="请输入密码"
        autocomplete="new-password"
        required
      />

      <BaseButton
        type="submit"
        :disabled="submitting"
        class="w-full h-10 rounded-lg text-white bg-[var(--button-primary-color)]! border-none! ring-0! hover:opacity-90 transition disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {{ submitting ? '初始化中...' : '创建 Owner' }}
      </BaseButton>
    </form>
  </section>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'

defineProps<{
  username: string
  password: string
  submitting: boolean
}>()

const emit = defineEmits<{
  (e: 'update:username', value: string): void
  (e: 'update:password', value: string): void
  (e: 'submit'): void
}>()
</script>

<style scoped></style>
