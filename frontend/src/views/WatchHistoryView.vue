<template>
  <AppLayout page-title="观看历史">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>观看历史</h2>
          <p>查看用户的视频观看记录</p>
        </div>
      </template>

      <a-spin v-if="loading" />
      <a-result v-else-if="error" status="error" :title="error" />

      <a-table v-else :data-source="rows" :columns="columns" :pagination="false" :row-key="rowKey">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'created_at'">
            {{ formatDateTime(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'updated_at'">
            {{ formatDateTime(record.updated_at) }}
          </template>
          <template v-else-if="column.key === 'progress'">
            {{ Math.round((record.progress || 0) * 100) }}%
          </template>
        </template>
      </a-table>
    </a-card>
  </AppLayout>
  
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { historyAPI } from '@/api'
import AppLayout from '@/components/AppLayout.vue'

interface VideoHistory {
  id: string
  user_id: string
  video_id: string
  video_title: string
  video_url: string
  source_id: string
  source_name: string
  progress: number
  created_at: string
  updated_at: string
}

const route = useRoute()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const rows = ref<VideoHistory[]>([])

const columns = [
  { title: '标题', dataIndex: 'video_title', key: 'video_title' },
  { title: '进度', dataIndex: 'progress', key: 'progress', width: 100 },
  { title: '来源', dataIndex: 'source_name', key: 'source_name', width: 140 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 180 },
]

const formatDateTime = (s?: string) => {
  if (!s) return '-'
  try { return new Date(s).toLocaleString('zh-CN') } catch { return s }
}

const loadHistory = async () => {
  if (!auth.token) {
    error.value = '未登录或登录已过期'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const uidParam = (route.params.userId as string) || ''
    const uid = (uidParam || (auth.user?.id ?? '')).trim()
    const res: any = await historyAPI.getVideoHistory(auth.token, uid)
    rows.value = (res?.data?.data as VideoHistory[]) || []
  } catch (e: any) {
    error.value = e?.message || '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadHistory()
})

// rowKey 显式类型
const rowKey = (r: VideoHistory): string => r.id
</script>

<style scoped>
@import './UserManagementView.css';
</style>


