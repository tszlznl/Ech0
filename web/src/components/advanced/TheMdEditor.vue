<template>
  <MarkdownEditor class="h-auto" v-model="content" :placeholder="t('editor.mainPlaceholder')" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import { MarkdownEditor } from '@/editor'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const { t } = useI18n()

const content = computed<string>({
  get: () => editorStore.echoToAdd.content,
  set: (val: string) => {
    editorStore.echoToAdd.content = val
  },
})
</script>
