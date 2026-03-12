<template>
  <div class="w-full px-2">
    <PanelCard class="mb-3">
      <div class="mb-3 flex items-center justify-between">
        <h1 class="text-lg font-bold text-[var(--color-text-primary)]">评论系统设置</h1>
        <button
          class="rounded-md border border-[var(--color-border-subtle)] px-3 py-1 text-sm"
          @click="saveSetting"
          :disabled="settingSaving"
        >
          {{ settingSaving ? '保存中...' : '保存设置' }}
        </button>
      </div>
      <div class="space-y-2 text-sm text-[var(--color-text-secondary)]">
        <label class="flex items-center gap-2">
          <input v-model="setting.enable_comment" type="checkbox" />
          启用评论系统
        </label>
        <label class="flex items-center gap-2">
          <input v-model="setting.require_approval" type="checkbox" />
          游客评论需审核
        </label>
        <label class="flex items-center gap-2">
          <input v-model="setting.captcha_enabled" type="checkbox" />
          启用验证码（可选）
        </label>
        <input
          v-model.trim="setting.captcha_verify_url"
          type="text"
          class="w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2"
          placeholder="验证码校验地址（例如 https://your-cap-service/verify）"
        />
        <input
          v-model.trim="setting.captcha_secret"
          type="text"
          class="w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2"
          placeholder="验证码 secret（按需填写）"
        />
      </div>
    </PanelCard>

    <PanelCard>
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <input
          v-model.trim="query.keyword"
          type="text"
          class="min-w-56 rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
          placeholder="搜索昵称、邮箱、内容"
        />
        <select
          v-model="query.status"
          class="rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
        >
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
        </select>
        <button class="rounded-md border border-[var(--color-border-subtle)] px-3 py-2 text-sm" @click="reload">
          查询
        </button>
        <button
          class="rounded-md border border-[var(--color-border-subtle)] px-3 py-2 text-sm"
          @click="runBatch('approve')"
        >
          批量通过
        </button>
        <button
          class="rounded-md border border-[var(--color-border-subtle)] px-3 py-2 text-sm"
          @click="runBatch('reject')"
        >
          批量拒绝
        </button>
        <button
          class="rounded-md border border-red-300 px-3 py-2 text-sm text-red-500"
          @click="runBatch('delete')"
        >
          批量删除
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full min-w-[760px] text-sm">
          <thead>
            <tr class="border-b border-[var(--color-border-subtle)] text-left text-[var(--color-text-muted)]">
              <th class="py-2"><input v-model="allChecked" type="checkbox" @change="toggleAll" /></th>
              <th class="py-2">昵称</th>
              <th class="py-2">邮箱</th>
              <th class="py-2">状态</th>
              <th class="py-2">时间</th>
              <th class="py-2">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in list.items" :key="item.id" class="border-b border-[var(--color-border-subtle)]">
              <td class="py-2">
                <input v-model="selectedIds" type="checkbox" :value="item.id" />
              </td>
              <td class="py-2">{{ item.nickname }}</td>
              <td class="py-2">{{ item.email }}</td>
              <td class="py-2">
                <span :class="statusClass(item.status)">{{ item.status }}</span>
              </td>
              <td class="py-2">{{ formatDate(item.created_at) }}</td>
              <td class="py-2">
                <div class="flex items-center gap-2">
                  <button class="text-sky-500 hover:underline" @click="openDetail(item.id)">详情</button>
                  <button class="text-emerald-500 hover:underline" @click="updateStatus(item.id, 'approved')">
                    通过
                  </button>
                  <button class="text-amber-500 hover:underline" @click="updateStatus(item.id, 'rejected')">
                    拒绝
                  </button>
                  <button class="text-red-500 hover:underline" @click="remove(item.id)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-3 flex items-center justify-between text-sm">
        <span class="text-[var(--color-text-muted)]">共 {{ list.total }} 条</span>
        <div class="flex items-center gap-2">
          <button class="rounded border px-2 py-1" :disabled="query.page <= 1" @click="prevPage">
            上一页
          </button>
          <span>第 {{ query.page }} 页</span>
          <button
            class="rounded border px-2 py-1"
            :disabled="query.page * query.page_size >= list.total"
            @click="nextPage"
          >
            下一页
          </button>
        </div>
      </div>
    </PanelCard>

    <div
      v-if="detailOpen && current"
      class="fixed inset-0 z-30 flex items-end justify-center bg-black/30 p-3 md:items-center"
      @click.self="detailOpen = false"
    >
      <div class="w-full max-w-lg rounded-lg bg-[var(--color-bg-canvas)] p-4">
        <div class="mb-2 flex items-center justify-between">
          <h3 class="text-base font-semibold">评论详情</h3>
          <button class="text-sm text-[var(--color-text-muted)]" @click="detailOpen = false">关闭</button>
        </div>
        <div class="space-y-1 text-sm">
          <p><b>昵称：</b>{{ current.nickname }}</p>
          <p><b>邮箱：</b>{{ current.email }}</p>
          <p><b>网址：</b>{{ current.website || '-' }}</p>
          <p><b>状态：</b>{{ current.status }}</p>
          <p><b>来源：</b>{{ current.source }}</p>
          <p><b>时间：</b>{{ formatDate(current.created_at) }}</p>
          <p class="mt-2 whitespace-pre-wrap break-words rounded border p-2">{{ current.content }}</p>
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
  fetchUpdatePanelCommentStatus,
} from '@/service/api'
import PanelCard from '@/layout/PanelCard.vue'
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

const selectedIds = ref<string[]>([])
const detailOpen = ref(false)
const current = ref<App.Api.Comment.CommentItem | null>(null)

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
  const res = await fetchGetPanelComments(query)
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

const toggleAll = () => {
  allChecked.value = !allChecked.value
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
  if (status === 'approved') return 'text-emerald-500'
  if (status === 'rejected') return 'text-amber-500'
  return 'text-sky-500'
}

onMounted(async () => {
  await Promise.all([loadSetting(), loadList()])
})
</script>

