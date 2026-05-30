<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  DiffText —— 逐 token 入场动画的文本块。

  防闪烁的关键：
  - 用一个持久化的 chunk 累加器（实例内部状态），每次只把「新增的尾巴」
    input.slice(已收长度) push 进去，并赋一个永不变的自增 id 作为 key。
  - 旧 chunk 不可变、key 不变 → Vue 复用其 DOM 节点，绝不重挂、绝不重播动画；
    只有新 push 的那一块挂载时播放一次入场动画。
  - 返回的是一个扁平、全部带 key 的 span 数组，不混入无 key 文本节点
    （混入无 key 节点会让 Vue 的 keyed diff 退化，导致整批重挂 → 整块闪烁）。

  组件实例必须被父组件按稳定 key 复用，内部累加状态才能跨重渲染存活。
-->
<script lang="ts">
import { defineComponent, h, ref, watch } from 'vue'

export default defineComponent({
  name: 'DiffText',
  props: {
    text: { type: String, default: '' },
    // 入场动画的 class 名（见下方 keyframes）
    animationClass: { type: String, default: 'am-blur-in' },
    duration: { type: Number, default: 600 },
    animate: { type: Boolean, default: true },
  },
  setup(props) {
    const chunks = ref<{ id: number; text: string }[]>([])
    let collected = ''
    let nextId = 0

    // 把一段新增文本切成「词 + 空白」逐个入列，每个词独立淡入，
    // granularity 更细、大块到达也是一词一词顺滑冒出
    const pushPieces = (delta: string) => {
      for (const piece of delta.split(/(\s+)/)) {
        if (piece.length) chunks.value.push({ id: nextId++, text: piece })
      }
    }

    // 内容变短（是旧内容的前缀）时，截断已收 chunk 而非重建——保留 id 不重播动画。
    // 行内标记闭合（如 **加粗** 把前缀文字切出去）会触发这种「变短」，不该让前缀重闪。
    const truncateTo = (len: number) => {
      const kept: { id: number; text: string }[] = []
      let acc = 0
      for (const c of chunks.value) {
        if (acc + c.text.length <= len) {
          kept.push(c)
          acc += c.text.length
        } else {
          if (len > acc) kept.push({ id: c.id, text: c.text.slice(0, len - acc) })
          break
        }
      }
      chunks.value = kept
    }

    const sync = (input: string) => {
      if (!props.animate) {
        chunks.value = input ? [{ id: 0, text: input }] : []
        collected = input
        return
      }
      if (input === collected) return
      // 只追加：取新增的尾巴，按词切分入列
      if (input.startsWith(collected)) {
        pushPieces(input.slice(collected.length))
        collected = input
        return
      }
      // 内容变短（旧内容的前缀）：截断保留，不重播
      if (collected.startsWith(input)) {
        truncateTo(input.length)
        collected = input
        return
      }
      // 真正分叉（内容被替换）→ 重置
      chunks.value = []
      pushPieces(input)
      collected = input
    }

    watch(() => props.text, sync, { immediate: true })

    return () =>
      chunks.value.map((c) =>
        h(
          'span',
          {
            key: c.id,
            class: ['am-tok', props.animate ? props.animationClass : null],
            style: props.animate ? { animationDuration: `${props.duration}ms` } : undefined,
          },
          c.text,
        ),
      )
  },
})
</script>

<!-- 非 scoped：DiffText 渲染的是 fragment（多个 span），用全局类确保稳定着色；
     名称统一 am- 前缀避免冲突 -->
<style>
.am-tok {
  /* inline-block + pre-wrap：blur/transform 渲染更干净且支持位移 */
  display: inline-block;
  white-space: pre-wrap;
  animation-iteration-count: 1;
  animation-fill-mode: both;

  /* ease-in-out 起落都柔 */
  animation-timing-function: ease-in-out;
}

.am-tok.am-blur-in {
  animation-name: am-blur-in;
}

.am-tok.am-fade-in {
  animation-name: am-fade-in;
}

/* blurIn：从左微微飘入 + 模糊聚焦，逐词顺序揭示即「从左到右浮动出现」 */
@keyframes am-blur-in {
  from {
    opacity: 0;
    filter: blur(5px);
    transform: translateX(-0.12em);
  }

  to {
    opacity: 1;
    filter: blur(0);
    transform: translateX(0);
  }
}

@keyframes am-fade-in {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}
</style>
