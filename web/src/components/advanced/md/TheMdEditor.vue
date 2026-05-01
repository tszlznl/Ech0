<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <MarkdownEditor class="h-auto" v-model="content" :placeholder="t('editor.mainPlaceholder')" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import { MarkdownEditor } from '@/editor'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  modelValue?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorStore = useEditorStore()
const { t } = useI18n()

const isControlled = computed(() => props.modelValue !== undefined)

const content = computed<string>({
  get: () => (isControlled.value ? props.modelValue || '' : editorStore.echoToAdd.content),
  set: (val: string) => {
    emit('update:modelValue', val)
    if (!isControlled.value) {
      editorStore.echoToAdd.content = val
    }
  },
})
</script>
