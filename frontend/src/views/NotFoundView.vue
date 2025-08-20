<template>
  <div class="not-found">
    <div class="content">
      <h1>404</h1>
      <p>页面不存在或已被移动</p>
      <a-button type="primary" @click="goHome">返回首页</a-button>
    </div>
  </div>
  
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
const goHome = () => router.push('/')
</script>

<style scoped>
.not-found {
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdf4 100%);
}
.content { text-align: center; color: #064e3b; }
h1 { font-size: 72px; margin: 0 0 8px; color: #047857; }
p { margin-bottom: 16px; }
</style>



