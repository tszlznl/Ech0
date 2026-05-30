// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ref, watch, onBeforeUnmount, type Ref } from 'vue'

/**
 * useSmoothReveal —— 恒定匀速逐词揭示。
 *
 * 真实 LLM 流是一秒几十词、还经常成块突发，直接喂给动画会一坨词同时淡入、节奏忽快忽慢。
 * 这里改为**恒定节拍**：无论网络多突发，都每隔固定间隔揭示一个词；积压只是被缓冲，按
 * 同一拍子从容排出。匀速且慢能让同一时刻只有少数几个词在淡入，逐词清爽。
 *
 * 慢必然滞后于生成（生成早完、动画还在播），由调用方等揭示追平后再切完整渲染、避免
 * 尾巴整坨弹出（见 AnimatedMarkdown 的 update:revealing）。
 *
 * - displayed 始终是 target 的前缀，只增不减（DiffText 据此逐词累加动画）。
 * - 流式中只揭示到最后一个空白处（不拆未写完的尾词）；流结束后揭示到结尾。
 */
// 每词间隔（毫秒），≈ 1000/词速。200ms ≈ 5 词/秒，兼顾清爽与可用性。
// 想更慢更清爽 → 调到 280~333；嫌慢 → 调到 120~140。
const WORD_INTERVAL_MS = 200

const isSpace = (ch: string): boolean => ch === ' ' || ch === '\n' || ch === '\t' || ch === '\r'

const lastWhitespaceIndex = (s: string): number => {
  for (let i = s.length - 1; i >= 0; i -= 1) {
    if (isSpace(s[i])) return i
  }
  return -1
}

// 从 from 起、在 limit 之内推进 n 个「词」，返回新的字符下标
const advanceWords = (s: string, from: number, n: number, limit: number): number => {
  let i = from
  for (let w = 0; w < n && i < limit; w += 1) {
    while (i < limit && isSpace(s[i])) i += 1
    while (i < limit && !isSpace(s[i])) i += 1
  }
  return i
}

// 末行是否只是一个「悬空块标记」（列表 -/*/+、有序 1.、标题 #、引用 >），后面还没内容。
// 揭示若停在这种中间态，markdown-it 会把它忽而解析成段落/标题/空列表项，下一拍又变回，
// 长列表里就表现为反复闪烁。
const DANGLING_MARKER = /^[ \t]*([-*+]|\d{1,9}[.)]|#{1,6}|>)[ \t]*$/
const tailIsDanglingMarker = (s: string, idx: number): boolean => {
  const lineStart = s.lastIndexOf('\n', idx - 1) + 1
  return DANGLING_MARKER.test(s.slice(lineStart, idx))
}
const markerLineStart = (s: string, idx: number): number => s.lastIndexOf('\n', idx - 1) + 1

export function useSmoothReveal(source: Ref<string>, streaming: Ref<boolean>): Ref<string> {
  const displayed = ref('')
  let raf = 0
  let last = 0
  let acc = 0 // 累计时间（毫秒），每满一个 WORD_INTERVAL_MS 揭示一个词

  const tick = (now: number) => {
    raf = 0
    const target = source.value
    // 流式中留住未写完的尾词；流结束后揭示到结尾
    const limit = streaming.value ? lastWhitespaceIndex(target) + 1 : target.length
    if (displayed.value.length >= limit) {
      last = 0
      acc = 0
      return
    }
    if (!last) last = now
    acc += Math.min(now - last, 100) // 卡顿/切后台后不要一次性补吐
    last = now

    // 恒定节拍：每满一个间隔揭示一个词
    let idx = displayed.value.length
    while (acc >= WORD_INTERVAL_MS && idx < limit) {
      acc -= WORD_INTERVAL_MS
      idx = advanceWords(target, idx, 1, limit)
      // 不要停在悬空块标记上：连内容词一起吃，避免列表/标题中间态解析抖动
      let guard = 0
      while (idx < limit && tailIsDanglingMarker(target, idx) && guard < 8) {
        idx = advanceWords(target, idx, 1, limit)
        guard += 1
      }
    }
    if (acc > WORD_INTERVAL_MS) acc = WORD_INTERVAL_MS // 防止积压时间导致下一拍突进
    // 若仍停在悬空标记（内容尚未到齐）→ 退回该行之前，等内容到了再连标记一起揭示
    if (tailIsDanglingMarker(target, idx)) idx = markerLineStart(target, idx)
    if (idx > displayed.value.length) displayed.value = target.slice(0, idx)
    schedule()
  }

  const schedule = () => {
    if (!raf && typeof requestAnimationFrame === 'function') {
      raf = requestAnimationFrame(tick)
    }
  }

  watch(
    source,
    (s) => {
      // 内容被替换（新一轮对话）→ 重置揭示
      if (!s.startsWith(displayed.value)) displayed.value = ''
      schedule()
    },
    { immediate: true },
  )

  // 流结束时需要把留住的尾词揭示出来
  watch(streaming, () => schedule())

  onBeforeUnmount(() => {
    if (raf) cancelAnimationFrame(raf)
  })

  return displayed
}
