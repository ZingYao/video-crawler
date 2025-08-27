<template>
  <AppLayout page-title="视频源管理">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>视频源管理</h2>
          <p>管理系统视频源站点配置</p>
        </div>
      </template>

      <div class="video-source-list-container">
        <div class="list-header">
          <h3>视频源列表</h3>
          <div class="list-actions">
            <a-button @click="showCreateModal" type="primary">
              <template #icon>
                <PlusOutlined />
              </template>
              添加站点
            </a-button>
            <a-button @click="refreshVideoSourceList" :loading="loading">
              <template #icon>
                <ReloadOutlined />
              </template>
              刷新
            </a-button>
            <a-button @click="checkAllStatus" :loading="checking">
              <template #icon>
                <ReloadOutlined />
              </template>
              批量检查
            </a-button>
            <a-button @click="exportVideoSources">
              <template #icon>
                <DownloadOutlined />
              </template>
              导出配置
            </a-button>
            <a-upload
              :show-upload-list="false"
              :before-upload="importVideoSources"
              accept=".json"
            >
              <a-button>
                <template #icon>
                  <UploadOutlined />
                </template>
                导入配置
              </a-button>
            </a-upload>
          </div>
        </div>

        <a-spin v-if="loading" size="large" />

        <a-result v-else-if="error" status="error" :title="error" :sub-title="'请检查网络连接或联系管理员'">
          <template #extra>
            <a-button type="primary" @click="refreshVideoSourceList">重试</a-button>
          </template>
        </a-result>

        <div class="table-responsive" v-else>
          <a-table :data-source="videoSourceList" :columns="columns" :pagination="false"
            :row-key="(record: VideoSource) => record.id" size="small" :default-sort-order="'descend'"
            :sort-directions="['descend', 'ascend']" :scroll="{ x: 1200 }">
            <template #bodyCell="{ column, record }: { column: any, record: VideoSource }">
              <template v-if="column.key === 'id'">
                <a-typography-text copyable :copy-text="record.id" @copy="() => message.success('站点ID已复制到剪贴板')">
                  {{ truncateId(record.id) }}
                </a-typography-text>
              </template>

              <template v-else-if="column.key === 'status'">
                <template v-if="editingStatusId === record.id">
                  <a-select size="small" style="width:140px" :value="record.status"
                    @change="(v: number) => onStatusChange(record, v)" @blur="onStatusBlur">
                    <a-select-option :value="0">禁用</a-select-option>
                    <a-select-option :value="1">正常</a-select-option>
                    <a-select-option :value="2">维护中</a-select-option>
                    <a-select-option :value="3">不可用</a-select-option>
                  </a-select>
                </template>
                <template v-else>
                  <a-tag :color="getStatusColor(record.status)" @click="() => (editingStatusId = record.id)"
                    style="cursor: pointer">
                    {{ getStatusText(record.status) }}
                  </a-tag>
                </template>
              </template>

              <template v-else-if="column.key === 'domain'">
                <a-typography-text copyable :copy-text="record.domain">
                  {{ record.domain }}
                </a-typography-text>
              </template>

              <template v-else-if="column.key === 'source_type'">
                <a-tag :color="getSourceTypeColor(record.source_type)">
                  {{ getSourceTypeText(record.source_type) }}
                </a-tag>
              </template>

              <template v-else-if="column.key === 'sort'">
                <span class="sort-value">{{ record.sort || 0 }}</span>
              </template>

              <template v-else-if="column.key === 'actions'">
                <a-space size="small">
                  <a-button type="primary" size="small" @click="editVideoSource(record.id)">
                    <template #icon>
                      <EditOutlined />
                    </template>
                    编辑
                  </a-button>
                  <a-popconfirm title="确定要删除这个视频源站点吗？" description="此操作不可恢复" @confirm="deleteVideoSource(record.id)"
                    ok-text="确定" cancel-text="取消">
                    <a-button type="primary" danger size="small">
                      <template #icon>
                        <DeleteOutlined />
                      </template>
                    删除
                    </a-button>
                  </a-popconfirm>
                  <a-button size="small" @click="checkStatus(record)">
                    <template #icon>
                      <ReloadOutlined />
                    </template>
                    检查
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>

        <a-empty v-if="!loading && !error && videoSourceList.length === 0" description="暂无视频源数据" />
      </div>
    </a-card>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'
import { message, notification } from 'ant-design-vue'
import {
  PlusOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  DownloadOutlined,
  UploadOutlined
} from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'

interface VideoSource {
  id: string
  name: string
  domain: string
  status: number
  source_type: number
  sort: number
  lua_script?: string
}

const router = useRouter()
const loading = ref(false)
const error = ref('')
const videoSourceList = ref<VideoSource[]>([])
const checking = ref(false)
const checkingIds = ref<Set<string>>(new Set())
const editingStatusId = ref<string>('')

async function onStatusChange(record: VideoSource, v: number) {
  await updateStatus(record.id, v)
  editingStatusId.value = ''
}

function onStatusBlur() { 
  editingStatusId.value = '' 
}

const columns = [
  {
    title: '站点ID',
    key: 'id',
    width: 100
  },
  {
    title: '站点名称',
    dataIndex: 'name',
    key: 'name',
    width: 120
  },
  {
    title: '站点域名',
    key: 'domain',
    width: 180
  },
  {
    title: '资源类型',
    key: 'source_type',
    width: 80
  },
  {
    title: '排序',
    key: 'sort',
    width: 60,
    sorter: (a: VideoSource, b: VideoSource) => b.sort - a.sort,
    defaultSortOrder: 'descend' as const
  },
  {
    title: '状态',
    key: 'status',
    width: 80
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right'
  }
]

const authStore = useAuthStore()
const token = computed(() => authStore.token)

const truncateId = (id: string) => {
  return id.length > 8 ? `${id.substring(0, 8)}...` : id
}

const normalizeStatus = (status: any): 0 | 1 | 2 | 3 => {
  const n = Number(status)
  if (n === 0 || n === 1 || n === 2 || n === 3) return n as 0 | 1 | 2 | 3
  return 3
}

const getStatusColor = (status: number) => {
  switch (status) {
    case 0: return 'default'
    case 1: return 'success'
    case 2: return 'warning'
    case 3: return 'error'
    default: return 'default'
  }
}

const getStatusText = (status: number) => {
  switch (status) {
    case 0: return '禁用'
    case 1: return '正常'
    case 2: return '维护中'
    case 3: return '不可用'
    default: return '未知'
  }
}

const getSourceTypeColor = (sourceType: number) => {
  switch (sourceType) {
    case 0: return 'blue'
    case 1: return 'green'
    case 2: return 'orange'
    case 3: return 'purple'
    case 4: return 'cyan'
    case 5: return 'red'
    default: return 'default'
  }
}

const getSourceTypeText = (sourceType: number) => {
  switch (sourceType) {
    case 0: return '综合'
    case 1: return '电影'
    case 2: return '电视剧'
    case 3: return '动漫'
    case 4: return '综艺'
    case 5: return '纪录片'
    default: return '未知'
  }
}

const refreshVideoSourceList = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await videoSourceAPI.getVideoSourceList(token.value!)
    if (response?.code === 0) {
      videoSourceList.value = response.data || []
    } else {
      error.value = response?.message || '获取视频源列表失败'
    }
  } catch (err: any) {
    error.value = err.message || '网络错误'
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  router.push('/video-source-edit')
}

const deleteVideoSource = async (id: string) => {
  try {
    const response = await videoSourceAPI.deleteVideoSource(token.value!, id)
    if (response?.code === 0) {
      message.success('删除成功')
      await refreshVideoSourceList()
    } else {
      message.error(response?.message || '删除失败')
    }
  } catch (err: any) {
    message.error(err.message || '删除失败')
  }
}

const updateStatus = async (id: string, status: number) => {
  try {
    const response = await videoSourceAPI.setStatus(token.value!, id, status)
    if (response?.code === 0) {
      message.success('状态更新成功')
      await refreshVideoSourceList()
    } else {
      message.error(response?.message || '状态更新失败')
    }
  } catch (err: any) {
    message.error(err.message || '状态更新失败')
  }
}

const checkStatus = async (record: VideoSource) => {
  checkingIds.value.add(record.id)
  try {
    const response = await videoSourceAPI.checkStatus(token.value!, record.id)
    if (response?.code === 0) {
      const newStatus = response.data?.status
      if (newStatus !== undefined) {
        record.status = normalizeStatus(newStatus)
        message.success('状态检查完成')
      }
    } else {
      message.error(response?.message || '检查失败')
    }
  } catch (err: any) {
    message.error(err.message || '检查失败')
  } finally {
    checkingIds.value.delete(record.id)
  }
}

const checkAllStatus = async () => {
  checking.value = true
  try {
    for (const record of videoSourceList.value) {
      await checkStatus(record)
    }
    message.success('批量检查完成')
  } finally {
    checking.value = false
  }
}

const exportVideoSources = async () => {
  try {
    const filename = `video-sources-${new Date().toISOString().slice(0, 10)}.json`
    const text = JSON.stringify(videoSourceList.value, null, 2)
    const blob = new Blob([text], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    message.success('导出成功')
  } catch (err: any) {
    message.error(err.message || '导出失败')
  }
}

const importVideoSources = async (file: File) => {
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    if (!Array.isArray(data)) {
      message.error('文件格式错误，请选择正确的JSON文件')
      return false
    }
    
    message.success(`成功导入 ${data.length} 个视频源`)
    await refreshVideoSourceList()
    return false
  } catch (err: any) {
    message.error(err.message || '导入失败')
    return false
  }
}

const editVideoSource = (id: string) => {
  router.push(`/video-source-edit/${id}`)
}

onMounted(() => {
  refreshVideoSourceList()
})
</script>

<style scoped>
.content-card {
  margin: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
  color: #333;
}

.card-header p {
  margin: 5px 0 0 0;
  color: #666;
  font-size: 14px;
}

.video-source-list-container {
  margin-top: 20px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.list-header h3 {
  margin: 0;
  color: #333;
}

.list-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.table-responsive {
  overflow-x: auto;
}

.sort-value {
  font-weight: 500;
  color: #1890ff;
}

@media (max-width: 768px) {
  .content-card {
    margin: 10px;
  }
  
  .list-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .list-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>