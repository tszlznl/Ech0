<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

const canvasRef = ref<HTMLCanvasElement | null>(null)
const wrapperRef = ref<HTMLElement | null>(null)

function mulberry32(seed: number) {
  let a = seed
  return function () {
    let t = (a += 0x6d2b79f5)
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function readCanvasBg(): string {
  if (typeof document === 'undefined') return '#f4f1ec'
  const v = getComputedStyle(document.documentElement).getPropertyValue('--color-bg-canvas').trim()
  if (v && (v.startsWith('#') || v.startsWith('rgb') || v.startsWith('oklch'))) return v
  return '#f4f1ec'
}

function draw() {
  const canvas = canvasRef.value
  const parent = wrapperRef.value
  if (!canvas || !parent) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const rect = parent.getBoundingClientRect()
  const W = Math.max(1, Math.floor(rect.width))
  const H = Math.max(1, Math.floor(rect.height))
  const dpr = Math.min(window.devicePixelRatio ?? 1, 2)

  canvas.width = Math.floor(W * dpr)
  canvas.height = Math.floor(H * dpr)
  canvas.style.width = `${W}px`
  canvas.style.height = `${H}px`

  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, W, H)

  const bg = readCanvasBg()
  ctx.fillStyle = bg
  ctx.fillRect(0, 0, W, H)

  const rnd = mulberry32(20260405)
  const lowGrain =
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  const greens = [
    'rgba(75, 98, 72, 0.11)',
    'rgba(95, 120, 86, 0.09)',
    'rgba(118, 142, 105, 0.08)',
    'rgba(58, 82, 58, 0.07)',
    'rgba(130, 155, 118, 0.07)',
  ]

  for (let i = 0; i < 95; i++) {
    const x = rnd() * W
    const y = rnd() * H * 0.94
    const rx = 18 + rnd() * (W * 0.2)
    const ry = 10 + rnd() * (H * 0.16)
    const rot = rnd() * Math.PI
    ctx.fillStyle = greens[Math.floor(rnd() * greens.length)]!
    ctx.save()
    ctx.translate(x, y)
    ctx.rotate(rot)
    ctx.scale(1, 0.42 + rnd() * 0.38)
    ctx.beginPath()
    ctx.ellipse(0, 0, rx, ry, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
  }

  ctx.strokeStyle = 'rgba(48, 62, 46, 0.07)'
  ctx.lineWidth = 0.75
  for (let i = 0; i < 28; i++) {
    ctx.beginPath()
    let x = rnd() * W
    let y = rnd() * H * 0.45
    ctx.moveTo(x, y)
    for (let s = 0; s < 6; s++) {
      x += (rnd() - 0.5) * W * 0.14
      y += 8 + rnd() * (H * 0.12)
      ctx.lineTo(x, y)
    }
    ctx.stroke()
  }

  const grainMul = lowGrain ? 0.09 : 0.2
  const grainCount = Math.floor(W * H * grainMul)
  for (let i = 0; i < grainCount; i++) {
    const gx = rnd() * W
    const gy = rnd() * H
    const t = rnd()
    ctx.fillStyle = t > 0.52 ? 'rgba(255,255,255,0.07)' : 'rgba(0,0,0,0.035)'
    ctx.fillRect(Math.floor(gx), Math.floor(gy), 1, 1)
  }

  const g = ctx.createRadialGradient(
    W / 2,
    H / 2,
    Math.min(W, H) * 0.15,
    W / 2,
    H / 2,
    Math.max(W, H) * 0.72,
  )
  g.addColorStop(0, 'rgba(0,0,0,0)')
  g.addColorStop(1, 'rgba(0,0,0,0.035)')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, W, H)
}

let ro: ResizeObserver | null = null

function scheduleDraw() {
  requestAnimationFrame(() => draw())
}

onMounted(async () => {
  await nextTick()
  scheduleDraw()
  requestAnimationFrame(scheduleDraw)

  ro = new ResizeObserver(() => scheduleDraw())
  if (wrapperRef.value) ro.observe(wrapperRef.value)
  window.addEventListener('resize', scheduleDraw)
})

onBeforeUnmount(() => {
  ro?.disconnect()
  window.removeEventListener('resize', scheduleDraw)
})
</script>

<template>
  <div
    ref="wrapperRef"
    class="hub-home-foliage aspect-[4/3] w-full max-w-[min(100%,18rem)] overflow-hidden rounded-xl shadow-[0_16px_40px_-12px_rgba(33,32,28,0.12)] ring-1 ring-[var(--color-border-subtle)]/80 sm:max-w-[20rem]"
    role="img"
    aria-label="Decorative illustration of soft leaf shadows"
  >
    <canvas ref="canvasRef" class="block h-full w-full" />
  </div>
</template>
