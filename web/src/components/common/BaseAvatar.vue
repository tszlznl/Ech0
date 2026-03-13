<template>
  <img :src="avatarSrc" :alt="alt" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

type AvatarOptionItem = string | number | boolean
type AvatarOptionValue = AvatarOptionItem | AvatarOptionItem[] | null | undefined

const props = withDefaults(
  defineProps<{
    seed?: string
    size?: number | string
    style?: string
    alt?: string
    src?: string
    options?: Record<string, AvatarOptionValue>
  }>(),
  {
    seed: 'guest',
    size: 128,
    style: 'micah',
    alt: 'avatar',
    src: '',
    options: () => ({}),
  },
)

const avatarSrc = computed(() => {
  const customSrc = props.src.trim()
  if (customSrc) return customSrc

  const seed = props.seed.trim() || 'guest'
  const url = new URL(`https://api.dicebear.com/9.x/${props.style}/svg`)
  url.searchParams.set('seed', seed)
  url.searchParams.set('size', String(props.size))

  Object.entries(props.options).forEach(([key, value]) => {
    if (value === undefined || value === null) return
    if (Array.isArray(value)) {
      if (!value.length) return
      url.searchParams.set(key, value.map(String).join(','))
      return
    }
    url.searchParams.set(key, String(value))
  })

  return url.href
})
</script>
