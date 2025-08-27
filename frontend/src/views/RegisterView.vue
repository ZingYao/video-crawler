<template>
  <div class="register-container">
    <a-card class="register-card">
      <div class="register-header">
        <span class="logo">🎬</span>
        <h1>用户注册</h1>
        <p>创建您的视频聚合系统账户</p>
      </div>

      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        class="register-form"
        @finish="handleRegister"
      >
        <a-form-item label="用户名" name="username">
          <a-input
            v-model:value="form.username"
            placeholder="请输入用户名"
            size="large"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item label="昵称" name="nickname">
          <a-input
            v-model:value="form.nickname"
            placeholder="请输入昵称（可选）"
            size="large"
          >
            <template #prefix>
              <SmileOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item label="密码" name="password">
          <a-input-password
            v-model:value="form.password"
            placeholder="请输入密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item label="确认密码" name="confirmPassword">
          <a-input-password
            v-model:value="form.confirmPassword"
            placeholder="请再次输入密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            {{ loading ? '注册中...' : '注册' }}
          </a-button>
        </a-form-item>

        <div v-if="error" class="error-message">
          <a-alert
            :message="error"
            type="error"
            show-icon
            closable
            @close="error = ''"
          />
        </div>

        <div class="form-footer">
          <span>已有账户？</span>
          <router-link to="/login" class="login-link">立即登录</router-link>
        </div>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserOutlined, LockOutlined, SmileOutlined } from '@ant-design/icons-vue'
import type { FormInstance } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const error = ref('')
const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  nickname: '',
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = async (_rule: any, value: string) => {
  if (value && value !== form.password) {
    throw new Error('两次输入的密码不一致')
  }
}

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  nickname: [
    { max: 50, message: '昵称不能超过 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleRegister = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.register({
      username: form.username,
      nickname: form.nickname || form.username,
      password: form.password
    })
    
    // 注册成功后跳转到登录页
    router.push('/login')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
@import './RegisterView.css';
</style>
