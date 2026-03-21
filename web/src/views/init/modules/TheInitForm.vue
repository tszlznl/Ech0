<template>
  <section
    class="w-full rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/40 p-3 sm:p-3.5"
  >
    <form class="space-y-3" @submit.prevent="emit('submit')">
      <BaseInput
        :model-value="email"
        @update:model-value="emit('update:email', String($event))"
        type="email"
        :placeholder="t('init.ownerEmailPlaceholder')"
        autocomplete="email"
        required
      />

      <BaseInput
        :model-value="username"
        @update:model-value="emit('update:username', String($event))"
        type="text"
        :placeholder="t('init.ownerUsernamePlaceholder')"
        autocomplete="username"
        required
      />

      <BaseInput
        :model-value="password"
        @update:model-value="emit('update:password', String($event))"
        type="password"
        :placeholder="t('init.ownerPasswordPlaceholder')"
        autocomplete="new-password"
        required
      />

      <BaseButton
        type="submit"
        :disabled="submitting"
        class="w-full h-8.5 rounded-md disabled:opacity-60 disabled:cursor-not-allowed"
      >
        <span class="text-[var(--color-text-secondary)]">
          {{ submitting ? t('init.initializing') : t('init.initialize') }}
        </span>
      </BaseButton>
    </form>
  </section>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  email: string
  username: string
  password: string
  submitting: boolean
}>()

const emit = defineEmits<{
  (e: 'update:email', value: string): void
  (e: 'update:username', value: string): void
  (e: 'update:password', value: string): void
  (e: 'submit'): void
}>()
</script>

<style scoped></style>
