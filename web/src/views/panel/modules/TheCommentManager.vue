<template>
  <div class="w-full px-2 comment-manager-page">
    <PanelCard class="mb-4">
      <div class="mb-4 flex items-center justify-between gap-3">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">评论系统设置</h1>
          <p class="text-xs text-[var(--color-text-muted)]">统一管理评论开关、审核策略与验证码配置。</p>
        </div>
        <BaseButton
          class="comment-btn comment-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
          @click="saveSetting"
          :disabled="settingSaving"
        >
          {{ settingSaving ? '保存中...' : '保存' }}
        </BaseButton>
      </div>

      <div class="space-y-2 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/50 p-3">
        <div class="setting-row">
          <div>
            <h3 class="setting-title">启用评论系统</h3>
            <p class="setting-desc">关闭后前台将不展示评论区。</p>
          </div>
          <BaseSwitch v-model="setting.enable_comment" />
        </div>
        <div class="setting-row">
          <div>
            <h3 class="setting-title">游客评论需审核</h3>
            <p class="setting-desc">开启后游客评论默认进入待审核状态。</p>
          </div>
          <BaseSwitch v-model="setting.require_approval" :disabled="!setting.enable_comment" />
        </div>
        <div class="setting-row">
          <div>
            <h3 class="setting-title">启用验证码</h3>
            <p class="setting-desc">按需接入第三方验证码服务。</p>
          </div>
          <BaseSwitch v-model="setting.captcha_enabled" :disabled="!setting.enable_comment" />
        </div>
      </div>

      <div v-if="setting.captcha_enabled" class="mt-3 grid gap-2">
        <BaseInput
          v-model.trim="setting.captcha_verify_url"
          type="text"
          :disabled="!setting.enable_comment"
          placeholder="Cap siteverify 地址（例如 https://your-cap-service/<key_id>/siteverify）"
        />
        <BaseInput
          v-model.trim="setting.captcha_secret"
          type="text"
          :disabled="!setting.enable_comment"
          placeholder="Cap key secret（用于 siteverify 校验）"
        />
      </div>
      <div v-else class="mt-3 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/60 px-3 py-2 text-xs text-[var(--color-text-muted)]">
        当前未开启验证码。项目使用
        <a
          href="https://github.com/tiagozip/cap"
          target="_blank"
          rel="noreferrer"
          class="underline underline-offset-2 hover:text-[var(--color-text-secondary)]"
        >
          cap
        </a>
        验证服务，启用前需要你先自行部署并配置 siteverify 地址与 key secret。
      </div>
    </PanelCard>

    <PanelCard>
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <BaseInput
          v-model.trim="query.keyword"
          type="text"
          class="min-w-56 md:w-64"
          placeholder="搜索昵称、邮箱、内容"
        />
        <BaseSelect
          v-model="query.status"
          class="h-9 min-w-28"
          :options="statusOptions"
          placeholder="全部状态"
        >
        </BaseSelect>
        <BaseSelect
          v-model="hotFilter"
          class="h-9 min-w-28"
          :options="hotOptions"
          placeholder="Hot 筛选"
        >
        </BaseSelect>
        <BaseButton class="comment-btn px-3 py-1.5 text-sm" @click="reload">
          查询
        </BaseButton>
        <BaseButton
          class="comment-btn px-3 py-1.5 text-sm"
          @click="runBatch('approve')"
          :disabled="selectedIds.length === 0"
        >
          批量通过
        </BaseButton>
        <BaseButton
          class="comment-btn px-3 py-1.5 text-sm"
          @click="runBatch('reject')"
          :disabled="selectedIds.length === 0"
        >
          批量拒绝
        </BaseButton>
        <BaseButton
          class="comment-btn-danger px-3 py-1.5 text-sm"
          @click="runBatch('delete')"
          :disabled="selectedIds.length === 0"
        >
          批量删除
        </BaseButton>
      </div>

      <div class="overflow-x-auto rounded-lg border border-[var(--color-border-subtle)]">
        <table class="w-full min-w-[760px] text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="py-2 pl-3"><input v-model="allChecked" type="checkbox" /></th>
              <th class="py-2">昵称</th>
              <th class="py-2">邮箱</th>
              <th class="py-2">状态</th>
              <th class="py-2">Hot</th>
              <th class="py-2">时间</th>
              <th class="py-2 pr-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in list.items"
              :key="item.id"
              class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
            >
              <td class="py-2 pl-3">
                <input v-model="selectedIds" type="checkbox" :value="item.id" />
              </td>
              <td class="py-2">{{ item.nickname }}</td>
              <td class="py-2">{{ item.email }}</td>
              <td class="py-2">
                <span class="status-pill" :class="statusClass(item.status)">
                  {{ statusLabelMap[item.status] || item.status }}
                </span>
              </td>
              <td class="py-2">
                <span class="status-pill" :class="hotClass(item.hot)">
                  {{ item.hot ? '精选' : '普通' }}
                </span>
              </td>
              <td class="py-2">{{ formatDate(item.created_at) }}</td>
              <td class="py-2 pr-3">
                <div class="flex items-center gap-2">
                  <button class="table-action text-sky-500" @click="openDetail(item.id)">详情</button>
                  <button class="table-action text-violet-500" @click="updateHot(item.id, !item.hot)">
                    {{ item.hot ? '取消 hot' : '设为 hot' }}
                  </button>
                  <button class="table-action text-emerald-500" @click="updateStatus(item.id, 'approved')">
                    通过
                  </button>
                  <button class="table-action text-amber-500" @click="updateStatus(item.id, 'rejected')">
                    拒绝
                  </button>
                  <button class="table-action text-red-500" @click="remove(item.id)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="list.items.length === 0">
              <td colspan="7" class="px-3 py-8 text-center text-[var(--color-text-muted)]">暂无评论数据</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-3 flex items-center justify-between text-sm">
        <span class="text-[var(--color-text-muted)]">共 {{ list.total }} 条</span>
        <div class="flex items-center gap-2">
          <BaseButton class="comment-btn px-2 py-1 text-sm" :disabled="query.page <= 1" @click="prevPage">
            上一页
          </BaseButton>
          <span class="text-[var(--color-text-secondary)]">第 {{ query.page }} / {{ totalPages }} 页</span>
          <BaseButton
            class="comment-btn px-2 py-1 text-sm"
            :disabled="query.page * query.page_size >= list.total"
            @click="nextPage"
          >
            下一页
          </BaseButton>
        </div>
      </div>
    </PanelCard>

    <div
      v-if="detailOpen && current"
      class="fixed inset-0 z-30 flex items-end justify-center bg-black/30 p-3 md:items-center"
      @click.self="detailOpen = false"
    >
      <div class="w-full max-w-lg rounded-xl border border-[var(--color-border-subtle)] bg-[var(--color-bg-canvas)] p-4 shadow-[var(--shadow-md)]">
        <div class="mb-2 flex items-center justify-between">
          <h3 class="text-base font-semibold text-[var(--color-text-primary)]">评论详情</h3>
          <button class="table-action text-sm text-[var(--color-text-muted)]" @click="detailOpen = false">关闭</button>
        </div>
        <div class="space-y-1 text-sm text-[var(--color-text-secondary)]">
          <p><b>昵称：</b>{{ current.nickname }}</p>
          <p><b>邮箱：</b>{{ current.email }}</p>
          <p><b>网址：</b>{{ current.website || '-' }}</p>
          <p><b>状态：</b>{{ statusLabelMap[current.status] || current.status }}</p>
          <p><b>Hot：</b>{{ current.hot ? '是' : '否' }}</p>
          <p><b>来源：</b>{{ current.source }}</p>
          <p><b>时间：</b>{{ formatDate(current.created_at) }}</p>
          <p class="mt-2 whitespace-pre-wrap break-words rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/50 p-2">
            {{ current.content }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  fetchBatchPanelComments,
  fetchDeletePanelComment,
  fetchGetCommentSystemSetting,
  fetchGetPanelCommentById,
  fetchGetPanelComments,
  fetchUpdateCommentSystemSetting,
  fetchUpdatePanelCommentHot,
  fetchUpdatePanelCommentStatus,
} from '@/service/api'
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'

const setting = reactive<App.Api.Comment.SystemSetting>({
  enable_comment: true,
  require_approval: true,
  captcha_enabled: false,
  captcha_verify_url: '',
  captcha_secret: '',
})
const settingSaving = ref(false)

const query = reactive<App.Api.Comment.PanelListQuery>({
  page: 1,
  page_size: 20,
  keyword: '',
  status: '',
  echo_id: '',
})

const list = reactive<App.Api.Comment.PanelPageResult>({
  items: [],
  total: 0,
})
const hotFilter = ref('')

const selectedIds = ref<string[]>([])
const detailOpen = ref(false)
const current = ref<App.Api.Comment.CommentItem | null>(null)
const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
]
const hotOptions = [
  { label: '全部 Hot', value: '' },
  { label: '仅 Hot', value: 'true' },
  { label: '仅非 Hot', value: 'false' },
]
const statusLabelMap: Record<string, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
}
const totalPages = computed(() => Math.max(1, Math.ceil(list.total / query.page_size)))

const allChecked = computed({
  get() {
    return list.items.length > 0 && selectedIds.value.length === list.items.length
  },
  set(v: boolean) {
    if (v) {
      selectedIds.value = list.items.map((i) => i.id)
    } else {
      selectedIds.value = []
    }
  },
})

const loadSetting = async () => {
  const res = await fetchGetCommentSystemSetting()
  if (res.code === 1) {
    Object.assign(setting, res.data)
  }
}

const saveSetting = async () => {
  settingSaving.value = true
  try {
    const res = await fetchUpdateCommentSystemSetting(setting)
    if (res.code === 1) {
      theToast.success('评论设置已更新')
    }
  } finally {
    settingSaving.value = false
  }
}

const loadList = async () => {
  const hot = hotFilter.value === '' ? undefined : hotFilter.value === 'true'
  const res = await fetchGetPanelComments({
    ...query,
    hot,
  })
  if (res.code === 1) {
    list.items = res.data.items || []
    list.total = res.data.total || 0
    selectedIds.value = []
  }
}

const reload = async () => {
  query.page = 1
  await loadList()
}

const runBatch = async (action: App.Api.Comment.BatchAction) => {
  if (selectedIds.value.length === 0) {
    theToast.info('请先选择评论')
    return
  }
  const res = await fetchBatchPanelComments(action, selectedIds.value)
  if (res.code === 1) {
    theToast.success('批量操作成功')
    await loadList()
  }
}

const updateStatus = async (id: string, status: App.Api.Comment.CommentStatus) => {
  const res = await fetchUpdatePanelCommentStatus(id, status)
  if (res.code === 1) {
    await loadList()
  }
}

const updateHot = async (id: string, hot: boolean) => {
  const res = await fetchUpdatePanelCommentHot(id, hot)
  if (res.code === 1) {
    await loadList()
  }
}

const remove = async (id: string) => {
  const res = await fetchDeletePanelComment(id)
  if (res.code === 1) {
    await loadList()
  }
}

const openDetail = async (id: string) => {
  const res = await fetchGetPanelCommentById(id)
  if (res.code === 1) {
    current.value = res.data
    detailOpen.value = true
  }
}

const prevPage = async () => {
  if (query.page <= 1) return
  query.page -= 1
  await loadList()
}

const nextPage = async () => {
  if (query.page * query.page_size >= list.total) return
  query.page += 1
  await loadList()
}

const statusClass = (status: string) => {
  if (status === 'approved') return 'status-approved'
  if (status === 'rejected') return 'status-rejected'
  return 'status-pending'
}

const hotClass = (hot: boolean) => {
  if (hot) return 'status-hot'
  return 'status-normal'
}

onMounted(async () => {
  await Promise.all([loadSetting(), loadList()])
})
</script>

<style scoped>
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.6rem;
}

.setting-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.setting-desc {
  margin-top: 0.1rem;
  font-size: 0.78rem;
  color: var(--color-text-muted);
}

.comment-btn {
  border: 1px solid var(--color-border-subtle) !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
}

.comment-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.comment-btn-primary {
  border-color: var(--color-border-strong) !important;
}

.comment-btn-danger {
  border: 1px solid color-mix(in srgb, #ef4444 35%, var(--color-border-subtle) 65%) !important;
  background: color-mix(in srgb, #ef4444 9%, var(--color-bg-surface) 91%) !important;
  color: #dc2626 !important;
}

.comment-btn-danger:hover {
  border-color: #f87171 !important;
  background: color-mix(in srgb, #ef4444 14%, var(--color-bg-surface) 86%) !important;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.1rem 0.5rem;
  font-size: 0.75rem;
  line-height: 1.2;
  border: 1px solid transparent;
}

.status-approved {
  color: #059669;
  border-color: color-mix(in srgb, #10b981 45%, transparent 55%);
  background: color-mix(in srgb, #10b981 12%, transparent 88%);
}

.status-rejected {
  color: #d97706;
  border-color: color-mix(in srgb, #f59e0b 45%, transparent 55%);
  background: color-mix(in srgb, #f59e0b 14%, transparent 86%);
}

.status-pending {
  color: #0369a1;
  border-color: color-mix(in srgb, #38bdf8 45%, transparent 55%);
  background: color-mix(in srgb, #38bdf8 14%, transparent 86%);
}

.status-hot {
  color: #7c3aed;
  border-color: color-mix(in srgb, #8b5cf6 45%, transparent 55%);
  background: color-mix(in srgb, #8b5cf6 14%, transparent 86%);
}

.status-normal {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-muted) 65%, transparent 35%);
}

.table-action {
  transition: opacity 0.2s ease;
}

.table-action:hover {
  opacity: 0.72;
}

@media (max-width: 768px) {
  .setting-row {
    align-items: flex-start;
  }
}
</style>

