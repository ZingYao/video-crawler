<template>
  <div class="startup-container">
    <div class="startup-content">
      <div class="logo-section">
        <div class="logo">🎬</div>
        <h1>视频聚合</h1>
        <p>Video Crawler Desktop</p>
      </div>

      <div class="loading-section">
        <div class="spinner"></div>
        <h2>服务启动中</h2>
        <p>{{ currentMessage }}</p>
        
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>
        <span>{{ Math.round(progress * 10) / 10 }}%</span>
      </div>

      <div class="error-section" v-if="hasError">
        <h3>启动失败</h3>
        <p>{{ errorMessage }}</p>
        <button @click="retryStartup">重试</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfigStore } from '@/stores/config'
import { GetServerPort } from '../../wailsjs/go/main/App'
import { isWails } from '../utils/wails'

const router = useRouter()
const configStore = useConfigStore()
const currentMessage = ref('正在初始化应用...')
const progress = ref(0)
const hasError = ref(false)
const errorMessage = ref('')

let startupTimer: number | null = null
let progressTimer: number | null = null
let checkTimer: number | null = null
// 启动页面组件
const checkServiceStatus = async () => {
  try {
    // 检查是否在Wails环境中
    const isWailsEnv = isWails()
    console.log('checkServiceStatus - 环境检测:', {
      isWails: isWailsEnv
    })
    if (isWailsEnv) {
      // 在Wails环境中，尝试调用Wails API获取服务端口
      try {
        const port = await GetServerPort()
        if (port && port > 0) {
          // 使用获取到的端口检查服务状态
          const response = await fetch(`http://localhost:${port}/health`)
          if (response.ok) {
            // 服务启动完成，加载配置
            currentMessage.value = '正在加载配置...'
            progress.value = 95
            
            try {
              await configStore.loadConfig()
              progress.value = 100
              currentMessage.value = '启动完成，正在跳转...'
              
              // 根据配置决定跳转目标
              const targetPath = configStore.needsLogin() ? '/login' : '/'
              
              // 打印调试信息
              console.log('启动页面跳转信息:', {
                config: configStore.config,
                needsLogin: configStore.needsLogin(),
                targetPath,
                currentTime: new Date().toISOString()
              })
              
              // 使用 replace 而不是 push，避免在历史记录中留下 /startup
              setTimeout(() => router.replace(targetPath), 1000)
              return true
            } catch (configError) {
              console.log('配置加载失败:', configError)
              // 即使配置加载失败，也继续跳转（默认跳转到首页）
              progress.value = 100
              currentMessage.value = '启动完成，正在跳转...'
              setTimeout(() => router.replace('/'), 1000)
              return true
            }
          }
        }
      } catch (wailsError) {
        // Wails API调用失败，继续等待
        return false
      }
    } else {
      // 在浏览器环境中，直接检查本地服务
      const response = await fetch('/health')
      if (response.ok) {
        // 服务启动完成，加载配置
        currentMessage.value = '正在加载配置...'
        progress.value = 95
        
        try {
          await configStore.loadConfig()
          progress.value = 100
          currentMessage.value = '启动完成，正在跳转...'
          
          // 根据配置决定跳转目标
          const targetPath = configStore.needsLogin() ? '/login' : '/'
          console.log('targetPath', targetPath)
          // 打印调试信息
          console.log('启动页面跳转信息:', {
            config: configStore.config,
            needsLogin: configStore.needsLogin(),
            targetPath,
            currentTime: new Date().toISOString()
          })
          
          // 使用 replace 而不是 push，避免在历史记录中留下 /startup
          setTimeout(() => router.replace(targetPath), 1000)
          return true
        } catch (configError) {
          console.log('配置加载失败:', configError)
          // 即使配置加载失败，也继续跳转（默认跳转到首页）
          progress.value = 100
          currentMessage.value = '启动完成，正在跳转...'
          setTimeout(() => router.replace('/'), 1000)
          return true
        }
      }
    }
  } catch (error) {
    return false
  }
  return false
}

const startServiceCheck = () => {
  checkTimer = window.setInterval(async () => {
    const isReady = await checkServiceStatus()
    if (isReady && checkTimer) {
      clearInterval(checkTimer)
      checkTimer = null
    }
  }, 1000)
}

const retryStartup = () => {
  hasError.value = false
  progress.value = 0
  currentMessage.value = '正在重新启动...'
  simulateStartup()
  startServiceCheck()
}

const simulateStartup = () => {
  progressTimer = window.setInterval(() => {
    if (progress.value < 90) {
      progress.value += Math.random() * 10
    }
  }, 500)
}

onMounted(() => {
  simulateStartup()
  startServiceCheck()
  
  setTimeout(() => {
    currentMessage.value = '正在启动HTTP服务...'
  }, 1000)
  
  setTimeout(() => {
    currentMessage.value = '正在连接数据库...'
  }, 2000)
  
  startupTimer = window.setTimeout(() => {
    hasError.value = true
    errorMessage.value = '服务启动超时，请检查网络连接或重启应用'
  }, 30000)
})

onUnmounted(() => {
  if (startupTimer) clearTimeout(startupTimer)
  if (progressTimer) clearInterval(progressTimer)
  if (checkTimer) clearInterval(checkTimer)
})
</script>

<style scoped>
.startup-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  /* 与标题一致的渐变绿色 */
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.startup-content {
  background: white;
  border-radius: 20px;
  padding: 60px 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  text-align: center;
  max-width: 500px;
  width: 90%;
}

.logo {
  font-size: 60px;
  margin-bottom: 20px;
}

.logo-section h1 {
  font-size: 32px;
  font-weight: 700;
  color: #333;
  margin: 0 0 8px 0;
}

.logo-section p {
  font-size: 16px;
  color: #666;
  margin: 0 0 40px 0;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #10b981;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-section h2 {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0 0 10px 0;
}

.loading-section p {
  font-size: 16px;
  color: #666;
  margin: 0 0 20px 0;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 10px;
}

.progress-fill {
  height: 100%;
  /* 与主题一致的绿色渐变 */
  background: linear-gradient(90deg, #10b981, #059669);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.error-section {
  margin-top: 30px;
  padding: 20px;
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
}

.error-section h3 {
  color: #ff4d4f;
  margin: 0 0 10px 0;
}

.error-section p {
  color: #666;
  margin: 0 0 15px 0;
}

.error-section button {
  background: #667eea;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
}

.error-section button:hover {
  background: #5a6fd8;
}
</style>
