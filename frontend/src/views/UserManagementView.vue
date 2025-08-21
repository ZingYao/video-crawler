<template>
  <AppLayout page-title="用户管理">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2 class="card-header">用户管理</h2>
          <p>管理系统用户和权限设置</p>
        </div>
      </template>

      <!-- 用户列表 -->
      <div class="user-list-container">
        <div class="list-header">
          <h3>用户列表</h3>
          <div class="list-actions">
            <a-button @click="refreshUserList" type="primary" :loading="loading">
              <template #icon>
                <ReloadOutlined />
              </template>
              刷新
            </a-button>
          </div>
        </div>

        <!-- 加载状态 -->
        <a-spin v-if="loading" size="large" />

        <!-- 错误状态 -->
        <a-result
          v-else-if="error"
          status="error"
          :title="error"
          :sub-title="'请检查网络连接或联系管理员'"
        >
          <template #extra>
            <a-button type="primary" @click="refreshUserList">重试</a-button>
          </template>
        </a-result>

        <!-- 用户列表 -->
        <div class="table-responsive" v-else>
        <a-table
          :data-source="userList"
          :columns="columns"
          :pagination="false"
          :row-key="(record: User) => record.id"
          size="small"
          :scroll="{ x: 900 }"
        >
          <template #bodyCell="{ column, record }: { column: any, record: User }">
            <template v-if="column.key === 'id'">
                             <a-typography-text
                 copyable
                 :copy-text="record.id"
                 @copy="() => message.success('用户ID已复制到剪贴板')"
               >
                {{ truncateId(record.id) }}
              </a-typography-text>
            </template>

            <template v-else-if="column.key === 'role'">
              <a-tag :color="record.is_admin ? 'red' : (record.is_site_admin ? 'purple' : 'blue')">
                {{ record.is_admin ? '管理员' : (record.is_site_admin ? '资源站点管理员' : '普通用户') }}
              </a-tag>
            </template>

            <template v-else-if="column.key === 'loginStatus'">
                             <a-switch
                 :checked="record.allow_login"
                 @change="(checked: boolean) => toggleLoginStatus(record.id, checked)"
                 :disabled="record.id === currentUserId"
               />
              <span class="status-text">
                {{ record.allow_login ? '允许' : '禁止' }}
              </span>
            </template>

            <template v-else-if="column.key === 'createdAt'">
              {{ formatDate(record.created_at) }}
            </template>

            <template v-else-if="column.key === 'lastLogin'">
              {{ formatDate(record.last_login_at) || '-' }}
            </template>

            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button
                  type="primary"
                  size="small"
                  @click="editUser(record.id)"
                  :disabled="record.id === currentUserId"
                >
                  <template #icon>
                    <EditOutlined />
                  </template>
                  编辑
                </a-button>
                <a-popconfirm
                  title="确定要删除这个用户吗？"
                  description="此操作不可恢复"
                  @confirm="deleteUser(record.id)"
                  ok-text="确定"
                  cancel-text="取消"
                >
                  <a-button
                    type="primary"
                    danger
                    size="small"
                    :disabled="record.id === currentUserId"
                  >
                    <template #icon>
                      <DeleteOutlined />
                    </template>
                    删除
                  </a-button>
                </a-popconfirm>
                <a-button
                  size="small"
                  @click="impersonate(record.id)"
                  v-if="authStore.user?.isAdmin"
                >
                  登录为该用户
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
        </div>

        <!-- 空状态 -->
        <a-empty v-if="!loading && !error && userList.length === 0" description="暂无用户数据" />
      </div>
    </a-card>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/api'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'

// 类型定义
interface User {
  id: string
  username: string
  nickname?: string
  is_admin?: boolean
  is_site_admin?: boolean
  allow_login?: boolean
  created_at?: string
  last_login_at?: string
  login_count?: number
}

// 表格列定义
const columns = [
  {
    title: '用户ID',
    key: 'id',
    width: 120
  },
  {
    title: '用户名',
    dataIndex: 'username',
    key: 'username',
    width: 120
  },
  {
    title: '昵称',
    dataIndex: 'nickname',
    key: 'nickname',
    width: 120,
    customRender: ({ text }: { text: string }) => text || '-'
  },
  {
    title: '角色',
    key: 'role',
    width: 100
  },
  {
    title: '登录状态',
    key: 'loginStatus',
    width: 120
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 150
  },
  {
    title: '最后登录',
    key: 'lastLogin',
    width: 150
  },
  {
    title: '登录次数',
    dataIndex: 'login_count',
    key: 'loginCount',
    width: 100,
    customRender: ({ text }: { text: number }) => text || 0
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    fixed: 'right'
  }
]

// 响应式数据
const router = useRouter()
const authStore = useAuthStore()
const userList = ref<User[]>([])
const loading = ref(false)
const error = ref('')

// 计算属性
const currentUserId = computed(() => authStore.user?.id)

// 方法
const loadUserList = async () => {
  if (!authStore.token) {
    error.value = '未登录或登录已过期'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await userAPI.getUserList(authStore.token)
    const data: User[] = response.data || []
    // 按注册时间降序（新到旧）
    userList.value = data.sort((a: User, b: User) => {
      const ta = a.created_at ? Date.parse(a.created_at) : 0
      const tb = b.created_at ? Date.parse(b.created_at) : 0
      return tb - ta
    })
  } catch (err: any) {
    error.value = err.message || '加载用户列表失败'
    console.error('加载用户列表失败:', err)
  } finally {
    loading.value = false
  }
}

const refreshUserList = () => {
  loadUserList()
}

const editUser = (userId: string) => {
  router.push(`/user/edit/${userId}`)
}

const deleteUser = async (userId: string) => {
  if (!authStore.token) {
    message.error('未登录或登录已过期')
    return
  }

  try {
    await userAPI.deleteUser(authStore.token, userId)
    message.success('用户删除成功')
    loadUserList() // 重新加载列表
  } catch (err: any) {
    message.error(err.message || '删除用户失败')
    console.error('删除用户失败:', err)
  }
}

const toggleLoginStatus = async (userId: string, allowLogin: boolean) => {
  if (!authStore.token) {
    message.error('未登录或登录已过期')
    return
  }

  try {
    await userAPI.changeLoginStatus(authStore.token, userId, allowLogin)
    message.success(`已${allowLogin ? '允许' : '禁止'}用户登录`)
    loadUserList() // 重新加载列表
  } catch (err: any) {
    message.error(err.message || '更新登录状态失败')
    console.error('更新登录状态失败:', err)
  }
}

const impersonate = async (userId: string) => {
  if (!authStore.token || !authStore.user?.isAdmin) { message.error('无权限'); return }
  try {
    const data: any = await userAPI.adminImpersonateLogin(authStore.token, userId)
    if (data.code !== 0) { message.error(data.message || '切换失败'); return }
    const loginResp = data.data
    const newUser = {
      id: loginResp.id,
      username: '',
      nickname: loginResp.nickname,
      isAdmin: loginResp.is_admin || false,
      isSiteAdmin: loginResp.is_site_admin || false,
      allowLogin: true
    }
    localStorage.setItem('token', loginResp.token)
    localStorage.setItem('user', JSON.stringify(newUser))
    ;(authStore as any).user = newUser
    ;(authStore as any).token = loginResp.token
    message.success('已切换为该用户（不记录登录历史）')
  } catch (e:any) {
    message.error(e?.message || '切换失败')
  }
}

const truncateId = (id: string) => {
  return id.length > 8 ? `${id.substring(0, 8)}...` : id
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}



// 强制样式覆盖
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
`

// 动态注入样式
const injectStyles = () => {
  const style = document.createElement('style')
  style.textContent = forceStyles
  document.head.appendChild(style)
}

// 生命周期
onMounted(() => {
  injectStyles()
  loadUserList()
})
</script>

<style scoped>
@import './UserManagementView.css'
</style>
