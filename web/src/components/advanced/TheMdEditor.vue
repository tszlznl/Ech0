<template>
  <MarkdownEditor class="h-auto" v-model="content" placeholder="一吐为快~" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores'
import { MarkdownEditor } from '@/editor'

const editorStore = useEditorStore()

const content = computed<string>({
  get: () => editorStore.echoToAdd.content,
  set: (val: string) => {
    editorStore.echoToAdd.content = val
  },
})
</script>
