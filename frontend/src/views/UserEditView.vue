<template>
  <AppLayout :page-title="isEdit ? '编辑用户' : '个人中心'">
    <template #default>
      <a-card class="content-card">
        <!-- 返回按钮 -->
        <div class="back-button-container">
          <a-button @click="goBack" class="back-button">
            <template #icon>
              <ArrowLeftOutlined />
            </template>
            返回
          </a-button>
        </div>
        <template #title>
          <div class="card-header">
            <h2>{{ isEdit ? '编辑用户' : '个人中心' }}</h2>
            <p>{{ isEdit ? '修改用户信息和权限设置' : '管理您的个人信息和账户设置' }}</p>
          </div>
        </template>

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
            <a-button type="primary" @click="loadUserData">重试</a-button>
          </template>
        </a-result>

        <!-- 用户表单 -->
        <a-form
          v-else
          ref="formRef"
          :model="formData"
          :rules="rules"
          layout="vertical"
          class="user-form"
          @finish="saveUser"
        >
          <a-form-item label="用户名" name="username">
            <a-input
              v-model:value="formData.username"
              :disabled="true"
              placeholder="请输入用户名"
            />
            <template #extra>
              <span class="form-hint">用户名不可修改</span>
            </template>
          </a-form-item>

          <a-form-item label="昵称" name="nickname">
            <a-input
              v-model:value="formData.nickname"
              placeholder="请输入昵称"
            />
          </a-form-item>

          <a-form-item label="密码" name="password">
            <a-input-password
              v-model:value="formData.password"
              :required="!isEdit"
              placeholder="请输入密码"
            />
            <template #extra v-if="isEdit">
              <span class="form-hint">留空则不修改密码</span>
            </template>
          </a-form-item>

          <a-form-item v-if="isEdit && isAdmin" label="权限设置">
            <a-space direction="vertical" style="width: 100%">
              <a-checkbox v-model:checked="formData.isAdmin">
                管理员权限
              </a-checkbox>
              <a-checkbox v-model:checked="formData.allowLogin">
                允许登录
              </a-checkbox>
            </a-space>
          </a-form-item>

          <a-form-item>
            <a-space>
              <a-button @click="goBack">
                取消
              </a-button>
              <a-button type="primary" html-type="submit" :loading="saving">
                {{ saving ? '保存中...' : '保存' }}
              </a-button>
            </a-space>
          </a-form-item>
        </a-form>
        
        <!-- 登录历史 -->
        <a-divider />
        <div class="login-history-section">
          <h3>登录历史</h3>
          <a-table
            :data-source="loginHistory"
            :columns="loginHistoryColumns"
            :pagination="false"
            :row-key="rowKey"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'login_at'">
                {{ formatDateTime(record.login_at) }}
              </template>
              <template v-else-if="column.key === 'success'">
                <a-tag :color="record.success ? 'green' : 'red'">
                  {{ record.success ? '成功' : '失败' }}
                </a-tag>
              </template>
            </template>
          </a-table>
        </div>
      </a-card>
    </template>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userAPI } from '@/api'
import { message } from 'ant-design-vue'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'
import type { FormInstance } from 'ant-design-vue'
import AppLayout from '@/components/AppLayout.vue'

// 类型定义
interface UserFormData {
  user_id: string
  username: string
  nickname: string
  password: string
  isAdmin: boolean
  allowLogin: boolean
}

// 登录历史类型（与后端 snake_case 对齐）
interface UserLoginHistory {
  login_at: string
  ip: string
  user_agent: string
  success: boolean
  token?: string
  password?: string
}

// 响应式数据
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formRef = ref<FormInstance>()

const formData = ref<UserFormData>({
  user_id: '',
  username: '',
  nickname: '',
  password: '',
  isAdmin: false,
  allowLogin: true
})

// 登录历史数据
const loginHistory = ref<UserLoginHistory[]>([])

// 表格列定义
const loginHistoryColumns = [
  { title: '时间', key: 'login_at', dataIndex: 'login_at', width: 180 },
  { title: 'IP', key: 'ip', dataIndex: 'ip', width: 140 },
  { title: 'User-Agent', key: 'user_agent', dataIndex: 'user_agent' },
  { title: '结果', key: 'success', dataIndex: 'success', width: 100 }
]

// 表格 rowKey（提供显式类型，避免隐式 any）
const rowKey = (_: UserLoginHistory, index: number): string => String(index)

// 计算属性
const isEdit = computed(() => {
  return route.params.userId !== undefined
})

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { max: 50, message: '昵称不能超过 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: !isEdit.value, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 个字符', trigger: 'blur' }
  ]
}

const isAdmin = computed(() => {
  return authStore.user?.isAdmin === true
})

// 方法
const loadUserData = async () => {
  if (!authStore.token) {
    error.value = '未登录或登录已过期'
    return
  }

  loading.value = true
  error.value = ''

  try {
    let userId = route.params.userId as string
    
    // 如果是个人中心，使用当前用户ID
    if (!isEdit.value) {
      userId = authStore.user?.id || ''
    }

    const response = await userAPI.getUserDetail(authStore.token, userId)
    const userData = response.data

    formData.value = {
      user_id: userData.id,
      username: userData.username,
      nickname: userData.nickname || '',
      password: '',
      isAdmin: userData.is_admin || false,
      allowLogin: userData.allow_login !== false
    }

    // 登录历史（如后端存在该字段）
    loginHistory.value = Array.isArray(userData.login_history) ? userData.login_history : []
  } catch (err: any) {
    error.value = err.message || '加载用户数据失败'
    console.error('加载用户数据失败:', err)
  } finally {
    loading.value = false
  }
}

const saveUser = async () => {
  if (!authStore.token) {
    message.error('未登录或登录已过期')
    return
  }

  saving.value = true

  try {
    const saveData: any = {
      user_id: formData.value.user_id,
      username: formData.value.username,
      nickname: formData.value.nickname,
      is_admin: formData.value.isAdmin,
      allow_login: formData.value.allowLogin
    }

    // 如果有密码，添加到保存数据中
    if (formData.value.password) {
      saveData.password = formData.value.password
    }

    await userAPI.saveUser(authStore.token, saveData)
    
    message.success('保存成功')
    
    // 如果是编辑当前用户，更新本地状态
    if (!isEdit.value && authStore.user) {
      authStore.user.nickname = formData.value.nickname
      authStore.user.isAdmin = formData.value.isAdmin
    }
    
    goBack()
  } catch (err: any) {
    message.error(err.message || '保存失败')
    console.error('保存用户失败:', err)
  } finally {
    saving.value = false
  }
}

const goBack = () => {
  router.back()
}

const formatDateTime = (iso?: string) => {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

// 生命周期
onMounted(() => {
  loadUserData()
})
</script>

<style scoped>
@import './UserEditView.css';
</style>
