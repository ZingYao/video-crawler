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
            :sort-directions="['descend', 'ascend']" :scroll="{ x: 1000 }">
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
                <a-space>
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
import { videoSourceAPI } from '@/utils/api'
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
function onStatusBlur() { editingStatusId.value = '' }

const columns = [
  {
    title: '站点ID',
    key: 'id',
    width: 120
  },
  {
    title: '站点名称',
    dataIndex: 'name',
    key: 'name',
    width: 150
  },
  {
    title: '站点域名',
    key: 'domain',
    width: 200
  },
  {
    title: '资源类型',
    key: 'source_type',
    width: 100
  },
  {
    title: '排序',
    key: 'sort',
    width: 80,
    sorter: (a: VideoSource, b: VideoSource) => b.sort - a.sort,
    defaultSortOrder: 'descend' as const
  },
  {
    title: '状态',
    key: 'status',
    width: 100
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
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
  // 超出范围一律按不可用处理
  return 3
}

const getStatusColor = (status: number) => {
  const s = normalizeStatus(status)
  const colors = ['red', 'green', 'orange', 'red']
  return colors[s]
}

const getStatusText = (status: number) => {
  const s = normalizeStatus(status)
  const texts = ['禁用', '正常', '维护中', '不可用']
  return texts[s]
}

async function updateStatus(id: string, status: number) {
  try {
    const resp = await videoSourceAPI.setStatus(id, status)
    if (resp && resp.code === 0) {
      message.success('状态已更新')
      await refreshVideoSourceList()
    } else {
      message.error(resp?.message || '更新失败')
    }
  } catch (e: any) {
    message.error(e?.message || '网络错误')
  }
}

const getSourceTypeColor = (sourceType: number) => {
  const colors = ['blue', 'purple', 'red', 'orange', 'green', 'cyan', 'geekblue', 'default']
  return colors[sourceType] || 'default'
}

const getSourceTypeText = (sourceType: number) => {
  const texts = ['综合', '短剧', '电影', '电视剧', '综艺', '动漫', '纪录片', '其他']
  return texts[sourceType] || '未知'
}

const fetchVideoSourceList = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await videoSourceAPI.getList()
    if (response.code === 0) {
      // 按 sort 字段降序排序
      const data = response.data || []
      videoSourceList.value = data.sort((a: VideoSource, b: VideoSource) => (b.sort || 0) - (a.sort || 0))
    } else {
      error.value = response.message || '获取视频源列表失败'
    }
  } catch (err: any) {
    error.value = err.message || '网络错误'
  } finally {
    loading.value = false
  }
}

const refreshVideoSourceList = () => {
  fetchVideoSourceList()
}

const showCreateModal = () => {
  router.push('/video-source-edit')
}

const editVideoSource = (id: string) => {
  router.push(`/video-source-edit/${id}`)
}

const deleteVideoSource = async (id: string) => {
  try {
    const response = await videoSourceAPI.delete(id)
    if (response.code === 0) {
      message.success('删除成功')
      fetchVideoSourceList()
    } else {
      message.error(response.message || '删除失败')
    }
  } catch (err: any) {
    message.error(err.message || '网络错误')
  }
}

// 导出视频源配置
const exportVideoSources = async () => {
  try {
    const response = await videoSourceAPI.exportVideoSources()
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'video-sources.json'
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
    message.success('配置导出成功')
  } catch (err: any) {
    message.error(err.message || '导出失败')
  }
}

// 导入视频源配置
const importVideoSources = async (file: File) => {
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    if (!Array.isArray(data)) {
      message.error('文件格式错误，请选择正确的JSON配置文件')
      return false
    }
    
    const response = await videoSourceAPI.importVideoSources(data)
    if (response.code === 0) {
      const importedCount = response.data.imported_count
      if (importedCount > 0) {
        message.success(`导入成功，新增 ${importedCount} 个站点`)
        // 刷新列表
        await fetchVideoSourceList()
      } else {
        message.info('没有新增站点，所有站点已存在')
      }
    } else {
      message.error(response.message || '导入失败')
    }
  } catch (err: any) {
    message.error(err.message || '导入失败')
  }
  
  return false // 阻止默认上传行为
}

// 批量检查资源状态（逐个检查）
const checkAllStatus = async () => {
  checking.value = true
  try {
    for (const item of videoSourceList.value) {
      await checkStatus(item)
    }
    notification.success({
      message: '批量检查完成',
      description: `共检查 ${videoSourceList.value.length} 个站点`,
      placement: 'topRight'
    })
    // 刷新一次，确保状态与服务端同步
    fetchVideoSourceList()
  } finally {
    checking.value = false
  }
}

// 单个检查（调用后端接口，若接口待实现可先占位）
const checkStatus = async (item: VideoSource) => {
  try {
    checkingIds.value.add(item.id)
    const res: any = await videoSourceAPI.checkStatus(item.id)
    if (res.code !== 0) {
      notification.error({
        message: `检查失败 - ${item.name}`,
        description: res.message || '未知错误',
        placement: 'topRight'
      })
      return
    }
    // 期望后端返回 data.status
    const newStatusRaw = res?.data
    const newStatus = normalizeStatus(newStatusRaw)
    item.status = newStatus
    const texts = ['禁用', '正常', '维护中', '不可用']
    const statusText = typeof newStatus === 'number' ? (texts[newStatus] || '未知') : '已完成'
    const notify = (type: 'success' | 'info' | 'warning' | 'error') =>
      notification[type]({
        message: `检查完成 - ${item.name}`,
        description: `状态：${statusText}`,
        placement: 'topRight'
      })
    if (newStatus === 1) notify('success')
    else if (newStatus === 2) notify('warning')
    else if (newStatus === 0 || newStatus === 3) notify('error')
    else notify('info')
  } catch (e: any) {
    notification.error({
      message: `检查异常 - ${item.name}`,
      description: e?.message || '网络错误',
      placement: 'topRight'
    })
  }
  finally {
    checkingIds.value.delete(item.id)
  }
}

const forceStyles = `
  .card-header h2 {
    text-align: center !important;
    margin: 0 0 8px 0 !important;
    color: #1e293b !important;
    font-size: 24px !important;
    font-weight: 600 !important;
  }
  
  .list-actions .ant-btn,
  .list-actions .ant-btn-primary {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%) !important;
    border: none !important;
    color: white !important;
  }
  
  .list-actions .ant-btn:hover,
  .list-actions .ant-btn-primary:hover {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%) !important;
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-primary {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%) !important;
    border: none !important;
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-primary:hover {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%) !important;
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-dangerous,
  .ant-table-tbody .ant-btn-dangerous span,
  .ant-table-tbody .ant-btn-dangerous * {
    background: linear-gradient(135deg, #fca5a5 0%, #f87171 100%) !important;
    border: none !important;
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-dangerous:hover,
  .ant-table-tbody .ant-btn-dangerous:hover span,
  .ant-table-tbody .ant-btn-dangerous:hover * {
    background: linear-gradient(135deg, #fecaca 0%, #fca5a5 100%) !important;
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-dangerous .anticon {
    color: white !important;
  }
  
  .ant-table-tbody .ant-btn-dangerous:hover .anticon {
    color: white !important;
  }
  
  .sort-value {
    font-weight: 600;
    color: #10b981;
    background: #ecfdf5;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 12px;
  }
`

// 动态注入样式
const injectStyles = () => {
  const style = document.createElement('style')
  style.textContent = forceStyles
  document.head.appendChild(style)
}

onMounted(() => {
  injectStyles()
  fetchVideoSourceList()
})
</script>

<style scoped>
@import './VideoSourceManagementView.css'
</style>