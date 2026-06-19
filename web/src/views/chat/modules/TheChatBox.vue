<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootEl" class="hairline-chat" :class="{ 'hairline-chat--empty': isEmpty }">
    <!-- 角落控件：无边框幽灵图标 -->
    <button class="ghost-ctrl ghost-ctrl--back" :title="t('commonNav.backHome')" @click="goHome">
      <Back class="ghost-ctrl__icon" />
    </button>
    <button
      v-if="messages.length > 0"
      class="ghost-ctrl ghost-ctrl--clear"
      :title="t('chatPanel.clear')"
      @click="handleClear"
    >
      <Close class="ghost-ctrl__icon" />
    </button>

    <!-- 对话区：从线的上方向上生长，贴底排列 -->
    <div ref="scrollArea" class="transcript">
      <div
        ref="transcriptInner"
        class="transcript__inner"
        :style="{ '--tail-space': tailSpace + 'px' }"
      >
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="turn"
          :class="msg.role === 'user' ? 'turn--user' : 'turn--ai'"
        >
          <!-- 用户：柔和的 accent-soft 气泡，右对齐 -->
          <p v-if="msg.role === 'user'" class="bubble">{{ msg.content }}</p>

          <!-- AI：无气泡的 markdown 正文 -->
          <template v-else>
            <!-- 推理折叠块：推理模型才有，默认折叠成「已思考（用时 X 秒）」，思考中自动展开 -->
            <ChatReasoning
              v-if="msg.reasoning !== undefined"
              :text="msg.reasoning"
              :active="msg.reasoningActive"
              :duration-ms="msg.reasoning_ms"
            />

            <!-- 检索状态条：Agent 自主检索时逐条显示关键词（设计 §9 searching 事件） -->
            <div v-if="msg.searches && msg.searches.length > 0" class="searching">
              <span
                v-for="(query, qi) in msg.searches"
                :key="qi"
                class="searching__chip"
                :class="{
                  'searching__chip--live': isStreaming(idx) && qi === msg.searches.length - 1,
                }"
              >
                {{ t('chatPanel.searching', { query }) }}
              </span>
            </div>

            <!-- 覆盖度状态条：summarize_echos 区间聚合（年终/月度总结）时如实展示覆盖范围，杜绝静默截断 -->
            <div v-if="msg.coverage" class="searching">
              <span class="searching__chip">
                {{
                  msg.coverage.truncated
                    ? t('chatPanel.coverageTruncated', { returned: msg.coverage.returned })
                    : t('chatPanel.coverage', { total: msg.coverage.total })
                }}
              </span>
            </div>

            <div
              v-if="msg.content.length === 0 && isStreaming(idx) && !msg.reasoningActive"
              class="thinking"
              :aria-label="t('chatPanel.send')"
            >
              <span class="thinking__dot" />
              <span class="thinking__dot" />
              <span class="thinking__dot" />
            </div>
            <div v-else-if="msg.content.length > 0" class="answer">
              <!-- 流式 + 揭示未追平时都用逐 token 动画；等揭示真正播完再切到带复制/折叠的完整渲染器，
                   避免慢节奏下尾巴整坨弹出 -->
              <AnimatedMarkdown
                v-if="showAnimated(idx)"
                :content="msg.content"
                :streaming="isStreaming(idx)"
                animation="blurIn"
                @update:revealing="assistantRevealing = $event"
              />
              <TheMdPreview v-else :content="msg.content" />
            </div>

            <!-- 失败/空回复：就地重发入口（仅最后一轮）。空回复时附一句轻提示，避免“跟没发一样” -->
            <div v-if="isRetryable(idx)" class="retry">
              <span v-if="msg.content.trim().length === 0" class="retry__hint">
                {{ t('chatPanel.noResponse') }}
              </span>
              <button class="retry__btn" :title="t('chatPanel.retry')" @click="retryLast">
                <svg
                  class="retry__icon"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M21 12a9 9 0 1 1-2.64-6.36" />
                  <path d="M21 3v6h-6" />
                </svg>
                <span>{{ t('chatPanel.retry') }}</span>
              </button>
            </div>
          </template>

          <!-- 引用来源：默认展示前三条，其余折叠（ChatSources 内部管理展开态） -->
          <ChatSources
            v-if="msg.sources && msg.sources.length > 0"
            :sources="msg.sources"
            @open="goToEcho"
          />
        </div>
      </div>
    </div>

    <!-- 右侧问题导航：默认是贴右边缘的小胶囊，hover 整条导航才展开提问文字；
         当前阅读所在的问题高亮，点击直接滚动跳转。仅在已有提问且非空态时出现 -->
    <nav v-if="questionNav.length > 0" class="qnav" :aria-label="t('chatPanel.navLabel')">
      <ul class="qnav__list">
        <li v-for="item in questionNav" :key="item.idx">
          <button
            class="qnav__item"
            :class="{ 'qnav__item--active': item.idx === activeQuestionIdx }"
            :title="item.content"
            :aria-current="item.idx === activeQuestionIdx ? 'true' : undefined"
            @click="scrollToQuestion(item.idx)"
          >
            <span class="qnav__label">{{ item.content }}</span>
            <span class="qnav__pill" />
          </button>
        </li>
      </ul>
    </nav>

    <!-- 输入区：textarea 的下边框就是那条横线（钉在 75vh） -->
    <div
      ref="composerEl"
      class="composer"
      :class="{
        'composer--active': input.trim().length > 0 || loading,
        'composer--loading': loading,
      }"
    >
      <textarea
        ref="inputEl"
        v-model="input"
        class="composer__field"
        rows="1"
        :placeholder="t('chatPanel.inputPlaceholder')"
        @input="autoGrow"
        @keydown="handleKeydown"
      />
      <Transition name="send-pop">
        <button
          v-if="loading"
          class="composer__action composer__action--stop"
          :title="t('chatPanel.clear')"
          @click="handleStop"
        >
          <span class="composer__stop-glyph" />
        </button>
        <button
          v-else-if="canSend"
          class="composer__action composer__action--send"
          :title="t('chatPanel.send')"
          @click="send(input)"
        >
          <Send class="composer__send-icon" />
        </button>
      </Transition>
    </div>

    <!-- 线下留白区：输入框为空时常驻预设场景，开始输入则淡出 -->
    <div class="understory">
      <Transition name="understory-fade">
        <div v-if="showSuggestions" class="understory__list">
          <p class="understory__hint">{{ t('chatPanel.suggestionsTitle') }}</p>
          <button
            v-for="(s, i) in suggestions"
            :key="i"
            class="understory__suggestion"
            @click="send(s)"
          >
            {{ s }}
          </button>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import Back from '@/components/icons/back.vue'
import Close from '@/components/icons/close.vue'
import Send from '@/components/icons/send.vue'
import { TheMdPreview } from '@/components/advanced/md'
import AnimatedMarkdown from './AnimatedMarkdown.vue'
import ChatSources from './ChatSources.vue'
import ChatReasoning from './ChatReasoning.vue'
import { ref, computed, nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { chatStream } from '@/service/api'
import { getChatSession, clearChatSession } from '@/service/api/chat'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { openConfirm } = useBaseDialog()

const input = ref<string>('')
const loading = ref<boolean>(false)
const messages = ref<App.Api.Chat.ChatMessage[]>([])
const rootEl = ref<HTMLElement | null>(null)
const scrollArea = ref<HTMLElement | null>(null)
const transcriptInner = ref<HTMLElement | null>(null)
const composerEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLTextAreaElement | null>(null)
let abort: (() => void) | null = null

const canSend = computed<boolean>(() => !loading.value && input.value.trim().length > 0)

// 最后一条 assistant 消息的逐 token 揭示是否尚未追平（由 AnimatedMarkdown 上报）
const assistantRevealing = ref<boolean>(false)

// 是否正在流式输出最后一条 assistant 消息
const isStreaming = (idx: number): boolean => loading.value && idx === messages.value.length - 1

// 是否对该消息使用动画渲染：流式中、或流已结束但揭示还没播完，都继续动画
const showAnimated = (idx: number): boolean => {
  const last = messages.value.length - 1
  if (idx !== last || messages.value[idx]?.role !== 'assistant') return false
  return loading.value || assistantRevealing.value
}

const suggestions = computed<string[]>(() => [
  t('chatPanel.suggestion1'),
  t('chatPanel.suggestion2'),
  t('chatPanel.suggestion3'),
])

// 输入框为空且非流式时展示预设场景；开始输入或发送后淡出
const showSuggestions = computed<boolean>(() => !loading.value && input.value.trim().length === 0)

// 尚无任何对话：发丝线整组（输入框 + suggestions）居中，告别上半屏大片空白；
// 发出首条消息后 messages 非空 → 自动下沉到 75vh 钉位（CSS transition 负责平滑过渡）
const isEmpty = computed<boolean>(() => messages.value.length === 0)

// ── 右侧问题导航（ToC） ─────────────────────────
// 导轨最多保留的提问条数：只留最新的几条，避免长会话把右侧拉成一长溜显得冗杂。
// 改这一个数字即可调整（如想更精简改成 5）。
const MAX_NAV_QUESTIONS = 7

// 只取用户提问，连同其在 messages 中的下标（下标即 .transcript__inner 的子节点序号，
// 用于定位 DOM、滚动跳转与高亮）；再 slice 出最新的若干条。保留原始下标，故跳转/高亮不受裁剪影响。
const questionNav = computed<{ idx: number; content: string }[]>(() =>
  messages.value
    .map((m, idx) => ({ idx, content: m.content, role: m.role }))
    .filter((m) => m.role === 'user')
    .slice(-MAX_NAV_QUESTIONS)
    .map(({ idx, content }) => ({ idx, content })),
)

// 当前阅读所在问题（messages 下标），-1 表示无。随滚动/内容长高刷新，驱动胶囊高亮。
const activeQuestionIdx = ref<number>(-1)

// 参考线：距对话区顶部的偏移，约等于顶部渐隐带高度，让「当前问题」取阅读区顶部那一条。
const NAV_TOP_GUTTER = 84

// 高亮判定容差：点击会把目标提问精确滚到阅读线，平滑滚动落点经设备像素取整后常落在线下数像素，
// 严格比较会把它误判给上一条。放宽几像素即可稳稳命中被点中的那条（提问间距远大于此，不会越界）。
const NAV_ACTIVE_TOLERANCE = 8

// v-for 按 messages 顺序渲染，故 .transcript__inner 的第 idx 个子节点恰是第 idx 条消息。
const turnElAt = (idx: number): HTMLElement | null =>
  (transcriptInner.value?.children[idx] as HTMLElement | undefined) ?? null

// 高亮规则：参考线之上（含）最靠下的那条用户提问即为「当前」；全在参考线之下则取第一条。
const updateActiveQuestion = () => {
  const area = scrollArea.value
  const nav = questionNav.value
  if (!area || nav.length === 0) {
    activeQuestionIdx.value = -1
    return
  }
  // 点击跳转的平滑滚动期间：高亮已被 scrollToQuestion 锁定为目标项，不被流式测量/滚动回写
  if (jumping) return
  // 贴底跟随直播时你就在最新一条：直接高亮它（此时它常在阅读线下方，逐条测量会漏掉）
  if (pinned.value) {
    activeQuestionIdx.value = nav[nav.length - 1].idx
    return
  }
  const refLine = area.getBoundingClientRect().top + NAV_TOP_GUTTER + NAV_ACTIVE_TOLERANCE
  let active = nav[0].idx
  for (const item of nav) {
    const el = turnElAt(item.idx)
    if (!el) continue
    if (el.getBoundingClientRect().top <= refLine) active = item.idx
    else break
  }
  activeQuestionIdx.value = active
}

// rAF 合帧：流式逐词揭示会高频触发 onContentResize，这里只做只读测量，按帧聚合即可，
// 不与既有「事件驱动写 scrollTop」的策略冲突。
let navRaf = 0
const scheduleActiveUpdate = () => {
  if (navRaf) return
  navRaf = requestAnimationFrame(() => {
    navRaf = 0
    updateActiveQuestion()
  })
}

// 点击胶囊：把对应提问滚到阅读区顶部（让出渐隐带）。主动跳转视为放弃贴底意图，
// 免得随后 ResizeObserver 又把视图拽回底部。
const scrollToQuestion = async (idx: number) => {
  const area = scrollArea.value
  const el = turnElAt(idx)
  if (!area || !el) return
  pinned.value = false
  // 立即高亮被点中的那条，并在跳转动画期间锁住：不等平滑滚动落定、也不被流式测量回写
  activeQuestionIdx.value = idx
  jumping = true
  if (jumpTimer) clearTimeout(jumpTimer)
  const delta = el.getBoundingClientRect().top - area.getBoundingClientRect().top
  const target = area.scrollTop + delta - NAV_TOP_GUTTER
  // 目标超过真实内容可滚到的上限（多为最新提问贴底、下方无内容）→ 临时撑出一屏留白，
  // 让它也能顶到阅读线；否则（较早的提问）无需留白，置 0。
  const realMax = area.scrollHeight - tailSpace.value - area.clientHeight
  tailSpace.value = target > realMax ? area.clientHeight : 0
  await nextTick() // 等留白落到 DOM，scrollTo 才不会被旧的可滚上限钳住
  jumpTimer = window.setTimeout(() => {
    jumping = false
    jumpTimer = 0
  }, 700)
  area.scrollTo({ top: target, behavior: 'smooth' })
}

// 贴底滚动：事件驱动而非帧驱动。`pinned` 是用户「想不想贴底」的意图，只由真实滚动翻转；
// 内容长高由 ResizeObserver 感知后跟随一次。彻底告别 rAF 每帧强写 scrollTop 带来的亚像素抖动。
const STICK_THRESHOLD = 80
const pinned = ref<boolean>(true)
let resizeObserver: ResizeObserver | null = null

// ToC「临时留白跳顶」：最新提问贴底、下方无内容可滚时，点击它会临时在底部撑出一屏空白，
// 好把它也顶到阅读线。不再需要后自动收起（滚回真正底部 / 答案已长到一屏 / 重新发消息）。
// tailSpace 仅是额外撑高的像素，经 --tail-space 注入 .transcript__inner 的 padding-bottom。
const tailSpace = ref<number>(0)
let jumping = false // 正在执行点击跳转的平滑滚动：期间禁用自动收起，免得把刚撑开的留白又抹掉
let jumpTimer = 0

// 仅当移除留白不会引起视图回弹（当前 scrollTop 仍落在收起后的可滚范围内）时才收起，杜绝跳变。
const collapseTailIfSafe = () => {
  const el = scrollArea.value
  if (!el || tailSpace.value === 0) return
  const maxAfter = el.scrollHeight - tailSpace.value - el.clientHeight
  if (el.scrollTop <= maxAfter + 1) tailSpace.value = 0
}

const jumpToBottom = () => {
  nextTick(() => {
    const el = scrollArea.value
    if (el) el.scrollTop = el.scrollHeight - tailSpace.value - el.clientHeight
  })
}

// 用户滚动时更新贴底意图：离底超过阈值即视为「想自己翻看」，滚回阈值内则恢复跟随。
// 程序触发的「跳到底」滚完仍在底部 → 读出 pinned=true，不会误翻转，故无需额外守卫。
const onScroll = () => {
  const el = scrollArea.value
  if (!el) return
  pinned.value = el.scrollHeight - el.scrollTop - el.clientHeight < STICK_THRESHOLD
  // 跳转动画进行中不碰留白；落定后：滚回真正底部 → 回归贴底跟随并收起，否则在安全时机收起
  if (!jumping) {
    if (pinned.value) tailSpace.value = 0
    else collapseTailIfSafe()
  }
  scheduleActiveUpdate()
}

// 内容尺寸变化（逐 token 揭示、流结束揭示尾词、渲染器切换）→ 若意图贴底则跟随一次。
// 已在底部时把 scrollTop 设成它本来的值是 no-op，不触发 scroll 事件、不抖。
const onContentResize = () => {
  // 贴底跟随永远对准「真实内容底」（减去额外留白），故撑开留白也不会把答案推上去露出空白
  const el = scrollArea.value
  if (el && pinned.value) el.scrollTop = el.scrollHeight - tailSpace.value - el.clientHeight
  scheduleActiveUpdate()
}

// 把输入框当前高度写进 --composer-h，供对话区底边实时让位。
// CSS max-height(--composer-max) 已封顶，offsetHeight 即真实封顶后的高度。
const syncComposerHeight = () => {
  const root = rootEl.value
  const c = composerEl.value
  if (root && c) root.style.setProperty('--composer-h', `${c.offsetHeight}px`)
}

// textarea 向上生长，横线恒定钉在 75vh；封顶由 CSS max-height 负责（超出内部滚动）
const autoGrow = () => {
  const el = inputEl.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${el.scrollHeight}px`
  syncComposerHeight()
  onContentResize() // 同帧贴底，杜绝输入框增高与对话区让位错开一帧的闪现
}

const goHome = () => router.push('/')
const goToEcho = (echoId: string) => router.push(`/echo/${echoId}`)

// 把一轮 SSE 流式问答挂到给定的 assistant 消息上（reactive 数组元素，原地累积）。
// send（新建一轮）与 retryLast（就地重生最后一条失败轮）共用，确保两条路径行为一致；
// 不在此清空输入框——重发时用户可能正打着下一个问题（清空交给 send 自己做）。
const streamInto = (question: string, assistant: App.Api.Chat.ChatMessage) => {
  loading.value = true
  assistantRevealing.value = true
  pinned.value = true
  // 回到常规贴底跟随：清掉上次 ToC 跳转撑开的临时留白与未结束的跳转计时
  tailSpace.value = 0
  jumping = false
  if (jumpTimer) {
    clearTimeout(jumpTimer)
    jumpTimer = 0
  }
  nextTick(autoGrow)
  jumpToBottom()

  abort = chatStream(question, {
    onSearching: (query) => {
      if (query && !assistant.searches?.includes(query)) {
        assistant.searches?.push(query)
      }
    },
    onSources: (sources) => {
      // sources 可多次增量到达，按 echo_id 累积去重（设计 §9）
      const merged = assistant.sources ? [...assistant.sources] : []
      const seen = new Set(merged.map((s) => s.echo_id))
      for (const src of sources) {
        if (!seen.has(src.echo_id)) {
          seen.add(src.echo_id)
          merged.push(src)
        }
      }
      assistant.sources = merged
    },
    onCoverage: (coverage) => {
      // 区间聚合总结（summarize_echos）的覆盖度，供「📚 已覆盖 N 条」状态条如实展示
      assistant.coverage = coverage
    },
    onReasoning: (text) => {
      // 推理模型的思考增量：首段到达即建块并标记「思考中」（折叠块自动展开）
      if (assistant.reasoning === undefined) {
        assistant.reasoning = ''
        assistant.reasoningActive = true
      }
      assistant.reasoning += text
    },
    onReasoningDone: (durationMs) => {
      // 推理结束：定格后端权威耗时并停掉「思考中」（折叠块自动收起，舞台让回答案）
      assistant.reasoning_ms = durationMs
      assistant.reasoningActive = false
    },
    onDelta: (text) => {
      assistant.content += text
    },
    onError: (message) => {
      // 传输/服务端 error 中断：标记失败态以亮出「重发」入口，并弹一次 toast 带出具体原因。
      // 不再把 errorGeneric 写进气泡正文——失败由内联重发区表达，红字正文反而喧宾夺主。
      loading.value = false
      assistant.failed = true
      theToast.error(message || String(t('chatPanel.errorGeneric')))
    },
    onDone: () => {
      loading.value = false
    },
  })
}

const send = (question: string) => {
  const q = question.trim()
  if (q.length === 0 || loading.value) return

  messages.value.push({ role: 'user', content: q })
  messages.value.push({ role: 'assistant', content: '', sources: [], searches: [] })
  input.value = ''
  // 取数组里那条 reactive 代理（而非刚 push 的裸对象），保证流式累积能触发渲染
  streamInto(q, messages.value[messages.value.length - 1])
}

// 失败/空回复判定：仅「最后一轮」可重发——后端 persistTurn 总在会话末尾追加，唯有就地重生
// 最后一轮才能保证前后端历史一致（中间轮重发会与后端的末尾追加错位）。命中条件：
// ① 流式中传输/服务端 error（failed），或 ② 正常收尾却空回复且无来源（静默失败）。
const isRetryable = (idx: number): boolean => {
  const m = messages.value[idx]
  if (!m || m.role !== 'assistant') return false
  if (isStreaming(idx) || idx !== messages.value.length - 1) return false
  if (m.failed === true) return true
  const noText = m.content.trim().length === 0
  const noSources = !m.sources || m.sources.length === 0
  return noText && noSources
}

// 就地重生最后一轮：保留提问气泡，清空那条失败 assistant 的全部状态后以同一问题重新流式。
// 失败/空轮次未被后端持久化（见 session.go persistTurn），故重发后会话历史保持干净。
const retryLast = () => {
  if (loading.value) return
  const n = messages.value.length
  if (n < 2) return
  const assistant = messages.value[n - 1]
  const user = messages.value[n - 2]
  if (assistant.role !== 'assistant' || user.role !== 'user') return

  assistant.content = ''
  assistant.sources = []
  assistant.searches = []
  assistant.coverage = undefined
  assistant.failed = false
  assistant.reasoning = undefined
  assistant.reasoning_ms = undefined
  assistant.reasoningActive = false
  streamInto(user.content, assistant)
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send(input.value)
  }
}

// 流式进行中点击停止：中断请求但保留已生成的内容（剩余尾巴让揭示从容播完）
const handleStop = () => {
  if (abort) abort()
  abort = null
  loading.value = false
}

// 流式结束（含正常完成 / 出错 / 手动停止）后，若没有正在进行的跳转，尝试收起临时留白：
// 答案已长到不致回弹时静默收起，过短则保留至下次滚到底/发消息（避免突兀跳变）。
watch(loading, (now, prev) => {
  if (!prev || now) return
  // 流式结束（完成/出错/手动停止）：若推理还卡在「思考中」（如手动停止未收到 reasoning_done），
  // 就地定格，避免折叠块永远转圈
  const last = messages.value[messages.value.length - 1]
  if (last?.role === 'assistant' && last.reasoningActive) last.reasoningActive = false
  if (!jumping) collapseTailIfSafe()
})

const handleClear = () => {
  openConfirm({
    title: t('chatPanel.clearConfirmTitle'),
    description: t('chatPanel.clearConfirmDesc'),
    onConfirm: async () => {
      if (abort) abort()
      abort = null
      loading.value = false
      assistantRevealing.value = false
      pinned.value = true
      tailSpace.value = 0
      jumping = false
      if (jumpTimer) {
        clearTimeout(jumpTimer)
        jumpTimer = 0
      }
      try {
        await clearChatSession()
      } catch {
        // 清除失败不阻断本地清空（best-effort）
      }
      messages.value = []
      theToast.success(String(t('chatPanel.clearSuccess')))
    },
  })
}

// 进入页面恢复上次的持久化会话（仅展示）。恢复的消息走静态渲染：
// loading 与 assistantRevealing 均为 false，showAnimated 对历史消息返回 false → 走 TheMdPreview。
onMounted(async () => {
  // 贴底引擎接线：scroll 监听更新意图，ResizeObserver 感知内容长高后跟随
  const area = scrollArea.value
  if (area) area.addEventListener('scroll', onScroll, { passive: true })
  if (typeof ResizeObserver === 'function') {
    // 一个实例观两目标，但按 entries 区分来源：对话内容长高只需贴底，唯有输入框自身
    // 变化才同步 --composer-h。否则流式逐词揭示时每次长高都会白写一次 CSS 变量、再被
    // 紧接的 scrollHeight 读触发一次强制重排——而这正是热路径，省掉它收益最直接。
    resizeObserver = new ResizeObserver((entries) => {
      let composerChanged = false
      for (const entry of entries) {
        if (entry.target === composerEl.value) composerChanged = true
      }
      if (composerChanged) syncComposerHeight()
      onContentResize()
    })
    if (transcriptInner.value) resizeObserver.observe(transcriptInner.value)
    if (composerEl.value) resizeObserver.observe(composerEl.value)
  }
  syncComposerHeight() // 首帧兜底，避免对话区底边先用默认值再跳一下

  try {
    const res = await getChatSession()
    const history = res.data
    if (Array.isArray(history) && history.length > 0) {
      messages.value = history
      pinned.value = true
      jumpToBottom()
    }
  } catch {
    // 恢复失败静默忽略，保持空态
  }
  scheduleActiveUpdate() // 兜底：无 ResizeObserver 的环境下也能为恢复的会话点亮当前胶囊

  // 由快捷输入框（Cmd/Ctrl+J）带入的问题：恢复历史后自动发送，并清掉 query 防止刷新重发
  const initialQuery = route.query.q
  const q = Array.isArray(initialQuery) ? initialQuery[0] : initialQuery
  if (typeof q === 'string' && q.trim().length > 0) {
    void router.replace({ query: {} })
    send(q)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  scrollArea.value?.removeEventListener('scroll', onScroll)
  if (navRaf) cancelAnimationFrame(navRaf)
  if (jumpTimer) clearTimeout(jumpTimer)
})
</script>

<style scoped>
.hairline-chat {
  /* 输入框滚动前的最大高度（≈5 行），小屏再按视口收口；单一事实源，对话区让位与 textarea 封顶共用 */
  --composer-max: min(8.5rem, 30dvh);

  /* 输入框当前高度，autoGrow / ResizeObserver 实时写入；首帧 1 行兜底 */
  --composer-h: 1.6rem;

  /* 发丝线距视口底的距离：有对话时钉在 25dvh（即 75vh 处），
     空态时抬到视口中部，让输入框那组元素居中、消灭上半屏留白 */
  --line-pos: 25dvh;

  position: relative;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background: var(--color-bg-canvas);
  color: var(--color-text-primary);
}

/* 空态：整组上移居中。composer.bottom / understory.top 各自带 transition，
   首条消息发出后 --line-pos 切回 25dvh 即平滑下沉 */
.hairline-chat--empty {
  --line-pos: 52dvh;
}

/* ── 角落幽灵控件 ───────────────────────────── */
.ghost-ctrl {
  position: absolute;
  top: 1.25rem;
  z-index: 3;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: none;
  border-radius: 999px;

  /* 磨砂圆底：背后无论滚到什么文字都要保持可辨——底色取较实的不透明度，别让按钮糊进正文。
     注意元素级 opacity 会与底色 alpha 相乘，故这里保持接近不透明，仅靠 muted 图标色维持克制 */
  background: color-mix(in srgb, var(--color-bg-canvas) 92%, transparent);
  backdrop-filter: blur(8px);
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0.95;
  transition:
    opacity 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.ghost-ctrl--back {
  left: 1.25rem;
}

.ghost-ctrl--clear {
  right: 1.25rem;
}

.ghost-ctrl:hover {
  opacity: 1;
  background: color-mix(in srgb, var(--color-bg-canvas) 98%, transparent);
  color: var(--color-text-primary);
}

.ghost-ctrl__icon {
  width: 1.2rem;
  height: 1.2rem;
}

/* 图标内置 fill=#888888，统一改用 currentColor 以便随主题/状态着色 */
.ghost-ctrl__icon :deep(path) {
  fill: currentColor;
}

/* ── 对话区（线之上，贴底向上生长） ─────────── */
.transcript {
  position: absolute;

  /* 底边随输入框当前高度实时让位：输入框长多少，对话区底边抬多少，永不重叠。
     不加 transition——textarea 是瞬时增高，对话区须同帧锁步，过渡反而会脱拍 */
  inset: 0 0 calc(var(--line-pos) + var(--composer-h, 1.6rem) + 1rem);

  /* 在横线上方再留出一段呼吸距离，避免内容黏住输入框 */
  display: flex;
  justify-content: center;
  overflow-y: auto;

  /* 贴底滚动由 JS 手动管理：关掉浏览器原生锚定，避免两者抢着调 scrollTop 造成抖动 */
  overflow-anchor: none;

  /* 顶部与底部都做渐隐：内容靠近边界时柔和淡出 */
  mask-image: linear-gradient(
    to bottom,
    transparent 0,
    #000 4.5rem,
    #000 calc(100% - 1rem),
    transparent 100%
  );
}

.transcript__inner {
  width: 100%;
  max-width: 42rem;

  /* 顶对齐、向下生长：内容不足时不再贴底，避免每多一行整块往上跳一格；
     溢出后由 jumpToBottom + ResizeObserver 跟随粘住底部 */
  align-self: flex-start;

  /* 底部留出 > 1rem 的 gutter：贴底时让最后一条来源避开 .transcript 底部 1rem 的渐隐带，
     杜绝来源块被 mask 半透明笼罩、随滚动逐帧跳变透明度的闪烁 */
  padding: 4.5rem 1.5rem 1.5rem;

  /* --tail-space 是 ToC「临时留白跳顶」撑出的额外底部空间，让最新提问也能顶到阅读线；
     默认 0，不影响常规布局。不加 transition：撑开后须同帧可滚，过渡会让 scrollTo 被旧上限钳住 */
  padding-bottom: calc(1.5rem + var(--tail-space, 0px));
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
}

.turn {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  animation: turn-in 0.32s ease both;

  /* 隔离各 turn 的布局/样式作用域：流式时最后一条逐词长高，不必反复重排上方已定稿的
     历史 turn。不含 paint——am-tok 的 blur/translateX 会微溢出，paint 收束会裁掉。 */
  contain: layout style;
}

.turn--user {
  align-items: flex-end;
}

.turn--ai {
  align-items: flex-start;
}

@keyframes turn-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

/* 用户气泡：柔和 accent-soft，无硬边框 */
.bubble {
  max-width: 85%;
  padding: 0.55rem 0.9rem;
  border-radius: 1.1rem 1.1rem 0.25rem;
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

/* AI 回答：markdown 正文 */
.answer {
  width: 100%;
  font-size: 1rem;
}

.answer :deep(.echo-markdown) {
  line-height: 1.8;
}

/* ── 失败/空回复：就地重发入口 ─────────────── */
.retry {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  margin-top: 0.15rem;
}

.retry__hint {
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--color-text-muted);
}

/* 无边框幽灵按钮，贴合整页的克制风格；hover 才浮起 accent */
.retry__btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem 0.2rem 0.4rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  line-height: 1.5;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease;
}

.retry__btn:hover {
  background: var(--color-accent-soft);
  color: var(--color-accent);
}

.retry__icon {
  width: 0.95rem;
  height: 0.95rem;
  flex-shrink: 0;
  transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1);
}

/* hover 时图标顺时针转一圈，呼应「重试」语义 */
.retry__btn:hover .retry__icon {
  transform: rotate(180deg);
}

@media (prefers-reduced-motion: reduce) {
  .retry__icon {
    transition: none;
  }
}

/* 首 token 到达前的「思考中」动画 */
.thinking {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0;
}

.thinking__dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.5;
  animation: thinking-bounce 1.2s ease-in-out infinite;
}

.thinking__dot:nth-child(2) {
  animation-delay: 0.16s;
}

.thinking__dot:nth-child(3) {
  animation-delay: 0.32s;
}

@keyframes thinking-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.35;
  }

  40% {
    transform: translateY(-0.28rem);
    opacity: 1;
  }
}

/* ── 检索状态条（Agent 自主检索） ───────────── */
.searching {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

.searching__chip {
  display: inline-flex;
  align-items: center;
  max-width: 22rem;
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  font-size: 0.72rem;
  line-height: 1.5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.searching__chip::before {
  content: '🔍';
  margin-right: 0.3rem;
  font-size: 0.7rem;
  font-variant-emoji: text;
}

/* 最新一条仍在检索：轻微呼吸 */
.searching__chip--live {
  animation: searching-pulse 1.4s ease-in-out infinite;
}

@keyframes searching-pulse {
  0%,
  100% {
    opacity: 0.55;
  }

  50% {
    opacity: 1;
  }
}

/* ── 输入区：下边框就是 75vh 那条横线 ───────── */
.composer {
  position: absolute;
  left: 50%;
  bottom: var(--line-pos);
  transform: translateX(-50%);
  z-index: 2;
  width: min(42rem, calc(100% - 3rem));
  display: flex;
  align-items: flex-end;
  gap: 0.6rem;

  /* 仅在空态 ↔ 有对话切换时（--line-pos 变化）平滑滑动；日常打字 bottom 不变，不受影响 */
  transition: bottom 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

/* 发丝线（底色）：两端淡出的渐变，避免满宽硬边显得空荡 */
.composer::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: linear-gradient(
    to right,
    transparent,
    var(--color-border-strong) 18%,
    var(--color-border-strong) 82%,
    transparent
  );
}

/* accent 线：聚焦/有内容时从中间向两端"画"出来，覆盖在灰线之上。
   两端淡出交给 mask，颜色/流光交给 background，互不干扰 */
.composer::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1;
  height: 1px;
  background: var(--color-accent);
  mask-image: linear-gradient(to right, transparent, #000 15%, #000 85%, transparent);
  transform: scaleX(0);
  transform-origin: 50% 50%;
  transition: transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.composer--active::before,
.composer:focus-within::before {
  transform: scaleX(1);
}

/* AI 回复时让 accent 线轻微流光，把这条贯穿全页的线当作状态指示 */
.composer--loading::before {
  background-image: linear-gradient(
    90deg,
    var(--color-accent) 0%,
    color-mix(in srgb, var(--color-accent) 35%, #fff) 50%,
    var(--color-accent) 100%
  );
  background-size: 200% 100%;
  animation: composer-shimmer 1.6s linear infinite;
}

@keyframes composer-shimmer {
  from {
    background-position: 150% 0;
  }

  to {
    background-position: -150% 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .composer,
  .understory {
    transition: none;
  }

  .composer::before {
    transition: none;
  }

  .composer--loading::before {
    animation: none;
  }
}

.composer__field {
  flex: 1;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  padding: 0 0 0.7rem;
  max-height: var(--composer-max);
  font-size: 1rem;
  line-height: 1.6;
  color: var(--color-text-primary);
  overflow-y: auto;
}

.composer__field::placeholder {
  color: var(--color-text-muted);
  opacity: 0.7;
}

/* 发送 / 停止：仅在可操作时出现，一个小小的极简字形，不抢戏 */
.composer__action {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  margin-bottom: 0.5rem;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.composer__action--send {
  color: var(--color-accent);
}

.composer__action:hover {
  transform: translateY(-1px);
  filter: brightness(1.1);
}

.composer__action:active {
  transform: translateY(0);
}

.composer__send-icon {
  width: 1.15rem;
  height: 1.15rem;
}

.composer__stop-glyph {
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 0.14rem;
  background: currentColor;
}

/* 发送按钮弹入/弹出 */
.send-pop-enter-active,
.send-pop-leave-active {
  transition:
    transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1),
    opacity 0.18s ease;
}

.send-pop-enter-from,
.send-pop-leave-to {
  transform: scale(0.5);
  opacity: 0;
}

/* ── 线下留白区（25vh） ─────────────────────── */
.understory {
  position: absolute;
  left: 50%;

  /* 紧贴发丝线下方：线在距底 --line-pos 处，故距顶 = 100dvh - --line-pos */
  top: calc(100dvh - var(--line-pos));
  transform: translateX(-50%);
  width: min(42rem, calc(100% - 3rem));
  padding-top: 1.5rem;

  /* 与 composer 同步滑动 */
  transition: top 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.understory__list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
}

.understory-fade-enter-active,
.understory-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.understory-fade-enter-from,
.understory-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

.understory__hint {
  font-size: 0.72rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  opacity: 0.7;
  margin-bottom: 0.25rem;
}

.understory__suggestion {
  border: none;
  background: transparent;
  padding: 0.15rem 0;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  text-align: left;
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.understory__suggestion:hover {
  color: var(--color-accent);
  transform: translateX(3px);
}

/* ── 右侧问题导航（ToC）：默认小胶囊，hover 整条导航才展开提问文字 ─── */
.qnav {
  position: absolute;
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  z-index: 3;
  display: flex;
  align-items: center;
  max-height: 72dvh;

  /* 折叠时只占住右缘一小条，并给鼠标留出从容的命中热区 */
  padding-right: 0.4rem;
}

.qnav__list {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.3rem;
  max-height: 72dvh;
  margin: 0;
  padding: 0;
  list-style: none;

  /* 提问很多时列表内部可滚动，但不露出滚动条，保持克制 */
  overflow-y: auto;
  scrollbar-width: none;
}

.qnav__list::-webkit-scrollbar {
  display: none;
}

.qnav__item {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  width: 100%;
  padding: 0.22rem 0.3rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  cursor: pointer;
  transition: background 0.22s ease;
}

/* 文字标签：折叠时宽度归零并淡出；hover 整条导航才展开 */
.qnav__label {
  max-width: 0;
  margin-right: 0;
  overflow: hidden;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  white-space: nowrap;
  text-overflow: ellipsis;
  opacity: 0;
  transition:
    max-width 0.28s cubic-bezier(0.22, 1, 0.36, 1),
    margin 0.28s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.2s ease;
}

/* 小胶囊：默认细短的横条，居右紧贴边缘 */
.qnav__pill {
  flex-shrink: 0;
  width: 1.1rem;
  height: 0.26rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.6;
  transition:
    width 0.22s ease,
    opacity 0.22s ease,
    background 0.22s ease;
}

/* 当前所在问题：accent 高亮，胶囊更长更实 */
.qnav__item--active .qnav__pill {
  width: 1.7rem;
  background: var(--color-accent);
  opacity: 1;
}

/* 悬停单条时给胶囊一点反馈 */
.qnav__item:hover .qnav__pill {
  opacity: 0.85;
}

/* hover 整条导航：每条套一层磨砂卡片、文字展开。底色取较实的不透明度，
   保证展开的提问文字压在滚动正文之上仍清晰可读 */
.qnav:hover .qnav__item {
  background: color-mix(in srgb, var(--color-bg-canvas) 94%, transparent);
  backdrop-filter: blur(10px);
}

.qnav:hover .qnav__label {
  max-width: min(38vw, 15rem);
  margin-right: 0.5rem;
  opacity: 1;
}

.qnav:hover .qnav__item--active .qnav__label {
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 触屏与窄屏：hover 无从触发、展开还会盖住正文，直接隐去 */
@media (hover: none), (width <= 768px) {
  .qnav {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qnav__item,
  .qnav__label,
  .qnav__pill {
    transition: none;
  }
}

@media (width <= 640px) {
  .transcript__inner {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
}
</style>
