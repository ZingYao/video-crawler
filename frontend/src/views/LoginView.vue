<template>
  <div class="login-container">
    <a-card class="login-card">
      <div class="login-header">
        <span class="logo">🎬</span>
        <h1>视频爬虫系统</h1>
        <p>欢迎使用现代化的视频数据采集平台</p>
      </div>

      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        class="login-form"
        @finish="handleLogin"
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

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            {{ loading ? '登录中...' : '登录' }}
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
          <span>还没有账户？</span>
          <router-link to="/register" class="register-link">立即注册</router-link>
        </div>
      </a-form>

      <div class="tech-info">
        <h3>技术栈</h3>
        <div class="tech-stack">
          <a-tag color="green">Vue 3.5.18</a-tag>
          <a-tag color="green">TypeScript 5.8.0</a-tag>
          <a-tag color="green">Gin 1.10.1</a-tag>
          <a-tag color="green">Vite 7.0.6</a-tag>
          <a-tag color="green">Ant Design Vue 4.x</a-tag>
        </div>
        <div class="open-source">
          <a-button type="link" href="https://github.com/ZingYao/video-crawler" target="_blank">GitHub</a-button>
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import type { FormInstance } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const error = ref('')
const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.login(form.username, form.password)
    const q = router.currentRoute.value.query
    const redirect = typeof q.redirect === 'string' ? q.redirect : '/'
    router.replace(redirect)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
@import './LoginView.css';
</style>
