<template>
  <div
    class="inline-flex items-center rounded-full ring-1 ring-inset ring-[var(--color-border-subtle)] bg-[var(--input-bg-color)] overflow-hidden"
  >
    <button
      v-if="editing"
      type="button"
      :title="applyTitle"
      :aria-label="applyTitle"
      class="w-7 h-7 flex items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
      @click="$emit('apply')"
    >
      <Saveupdate class="w-4 h-4" />
    </button>
    <button
      type="button"
      :title="editing ? cancelTitle : editTitle"
      :aria-label="editing ? cancelTitle : editTitle"
      class="w-7 h-7 flex items-center justify-center text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)] transition-colors duration-200"
      @click="handleSecondaryClick"
    >
      <Close v-if="editing" class="w-4 h-4" />
      <Edit v-else class="w-4 h-4" />
    </button>
  </div>
</template>

<script setup lang="ts">
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'

withDefaults(
  defineProps<{
    editing: boolean
    editTitle?: string
    cancelTitle?: string
    applyTitle?: string
  }>(),
  {
    editTitle: '编辑',
    cancelTitle: '取消',
    applyTitle: '应用',
  },
)

const emit = defineEmits<{
  apply: []
  toggle: []
}>()

const handleSecondaryClick = () => {
  emit('toggle')
}
</script>
