<template>
  <div ref="rootRef" class="echo-markdown" v-html="html"></div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { renderMarkdown } from '../core/markdown'
import { useI18n } from 'vue-i18n'
import '../styles/markdown.scss'

const props = defineProps<{
  content: string
}>()

const rootRef = ref<HTMLElement | null>(null)
const { t } = useI18n()
const html = computed(() =>
  renderMarkdown(props.content, {
    expandLabel: String(t('markdown.expand')),
    collapseLabel: String(t('markdown.collapse')),
    taskCheckboxLabel: String(t('markdown.taskCheckboxLabel')),
  }),
)

function onRootClick(event: Event) {
  const target = event.target
  if (!(target instanceof HTMLElement)) return

  const toggleButton = target.closest<HTMLButtonElement>('.code-block-toggle')
  if (!toggleButton || !rootRef.value?.contains(toggleButton)) return

  const block = toggleButton.closest<HTMLElement>('.code-block--collapsible')
  if (!block) return

  const isCollapsed = block.classList.toggle('code-block--collapsed')
  const expandLabel = toggleButton.dataset.expandLabel ?? String(t('markdown.expand'))
  const collapseLabel = toggleButton.dataset.collapseLabel ?? String(t('markdown.collapse'))

  toggleButton.setAttribute('aria-expanded', String(!isCollapsed))
  toggleButton.textContent = isCollapsed ? expandLabel : collapseLabel
}

onMounted(() => {
  rootRef.value?.addEventListener('click', onRootClick)
})

onBeforeUnmount(() => {
  rootRef.value?.removeEventListener('click', onRootClick)
})
</script>
