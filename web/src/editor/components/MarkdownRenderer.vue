<template>
  <div ref="rootRef" class="echo-markdown">
    <div
      v-if="!rendererReady && props.content"
      class="markdown-renderer-placeholder"
      aria-hidden="true"
    ></div>
    <div v-else v-html="html"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import '../styles/markdown.scss'

type RenderMarkdown = (typeof import('../core/markdown'))['renderMarkdown']

let renderMarkdownFn: RenderMarkdown | null = null
let renderMarkdownPromise: Promise<RenderMarkdown> | null = null

const loadRenderMarkdown = async (): Promise<RenderMarkdown> => {
  if (renderMarkdownFn) return renderMarkdownFn
  if (!renderMarkdownPromise) {
    renderMarkdownPromise = import('../core/markdown').then((module) => {
      renderMarkdownFn = module.renderMarkdown
      return module.renderMarkdown
    })
  }
  return renderMarkdownPromise
}

const props = defineProps<{
  content: string
}>()

const rootRef = ref<HTMLElement | null>(null)
const { t } = useI18n()
const copyResetTimers = new WeakMap<HTMLButtonElement, ReturnType<typeof setTimeout>>()
const expandLabel = computed(() => String(t('markdown.expand')))
const collapseLabel = computed(() => String(t('markdown.collapse')))
const copyLabel = computed(() => String(t('markdown.copy')))
const copiedLabel = computed(() => String(t('markdown.copied')))
const taskCheckboxLabel = computed(() => String(t('markdown.taskCheckboxLabel')))
const html = ref('')
const rendererReady = ref(Boolean(renderMarkdownFn))
let renderSequence = 0

const renderContent = async () => {
  const currentSequence = ++renderSequence
  const content = props.content
  if (!content) {
    html.value = ''
    return
  }

  const render = await loadRenderMarkdown()
  if (currentSequence !== renderSequence) return

  const rendered = await render(content, {
    expandLabel: expandLabel.value,
    collapseLabel: collapseLabel.value,
    copyLabel: copyLabel.value,
    copiedLabel: copiedLabel.value,
    taskCheckboxLabel: taskCheckboxLabel.value,
  })
  if (currentSequence !== renderSequence) return

  rendererReady.value = true
  html.value = rendered
}

function onRootClick(event: Event) {
  if (!rendererReady.value) return
  const target = event.target
  if (!(target instanceof HTMLElement)) return

  const toggleButton = target.closest<HTMLButtonElement>('.code-block-toggle')
  if (toggleButton && rootRef.value?.contains(toggleButton)) {
    const block = toggleButton.closest<HTMLElement>('.code-block--collapsible')
    if (!block) return

    const isCollapsed = block.classList.toggle('code-block--collapsed')
    const expandLabel = toggleButton.dataset.expandLabel ?? String(t('markdown.expand'))
    const collapseLabel = toggleButton.dataset.collapseLabel ?? String(t('markdown.collapse'))

    toggleButton.setAttribute('aria-expanded', String(!isCollapsed))
    toggleButton.textContent = isCollapsed ? expandLabel : collapseLabel
    return
  }

  const copyButton = target.closest<HTMLButtonElement>('.code-block-copy')
  if (copyButton && rootRef.value?.contains(copyButton)) {
    const block = copyButton.closest<HTMLElement>('.code-block')
    const codeEl = block?.querySelector<HTMLElement>('pre code')
    if (!codeEl) return

    const text = codeEl.textContent ?? ''
    const copyLabel = copyButton.dataset.copyLabel ?? String(t('markdown.copy'))
    const copiedLabel = copyButton.dataset.copiedLabel ?? String(t('markdown.copied'))

    const markCopied = () => {
      copyButton.classList.add('is-copied')
      copyButton.textContent = copiedLabel
      const prev = copyResetTimers.get(copyButton)
      if (prev) clearTimeout(prev)
      copyResetTimers.set(
        copyButton,
        setTimeout(() => {
          copyButton.classList.remove('is-copied')
          copyButton.textContent = copyLabel
          copyResetTimers.delete(copyButton)
        }, 1800),
      )
    }

    if (navigator.clipboard?.writeText) {
      navigator.clipboard
        .writeText(text)
        .then(markCopied)
        .catch(() => {})
    }
  }
}

onMounted(() => {
  rootRef.value?.addEventListener('click', onRootClick)
})

onBeforeUnmount(() => {
  rootRef.value?.removeEventListener('click', onRootClick)
})

watch(
  [() => props.content, expandLabel, collapseLabel, taskCheckboxLabel],
  () => {
    void renderContent()
  },
  { immediate: true },
)
</script>

<style scoped>
.markdown-renderer-placeholder {
  min-height: 2.75rem;
  border-radius: 0.5rem;
  background:
    linear-gradient(
      90deg,
      rgb(140 140 140 / 8%) 25%,
      rgb(140 140 140 / 18%) 37%,
      rgb(140 140 140 / 8%) 63%
    ),
    linear-gradient(180deg, rgb(120 120 120 / 5%), rgb(120 120 120 / 8%));
  background-size:
    240% 100%,
    100% 100%;
  animation: markdown-placeholder-wave 1.6s ease-in-out infinite;
}

@keyframes markdown-placeholder-wave {
  0% {
    background-position:
      100% 0,
      0 0;
  }

  100% {
    background-position:
      -100% 0,
      0 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .markdown-renderer-placeholder {
    animation: none;
  }
}
</style>
