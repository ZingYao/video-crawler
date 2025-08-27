<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'
import { systemAPI } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()

// 响应式数据
const authStore = useAuthStore()

const healthStatus = ref({
  status: 'unknown',
  lastCheck: '未检查'
})

const apiStatus = ref({
  status: 'unknown',
  lastCheck: '未检查'
})

const apiResult = ref<any>(null)

const openGithub = () => {
  window.open('https://github.com/ZingYao/video-crawler', '_blank')
}

// 方法
const testHealth = async () => {
  try {
    const data = await systemAPI.health()
    healthStatus.value = {
      status: 'healthy',
      lastCheck: new Date().toLocaleString('zh-CN')
    }
  } catch (error) {
    healthStatus.value = {
      status: 'unhealthy',
      lastCheck: new Date().toLocaleString('zh-CN')
    }
    console.error('Health check failed:', error)
  }
}

const testApi = async () => {
  try {
    if (!authStore.token) {
      throw new Error('未登录，无法获取API信息')
    }
    const data = await systemAPI.apiInfo(authStore.token)
    apiStatus.value = {
      status: 'healthy',
      lastCheck: new Date().toLocaleString('zh-CN')
    }
    apiResult.value = data
  } catch (error) {
    apiStatus.value = {
      status: 'unhealthy',
      lastCheck: new Date().toLocaleString('zh-CN')
    }
    console.error('API info failed:', error)
  }
}

// 生命周期
onMounted(() => { testHealth() })
</script>

<template>
  <AppLayout page-title="首页">
    <template #default>
      <div class="content-card">
        <div class="card-header">
          <h2>欢迎使用视频爬虫系统</h2>
          <p>这是一个功能强大的视频数据采集和管理平台</p>
        </div>

        <div class="feature-grid">
          <div class="feature-card" @click="openGithub" style="cursor:pointer;">
            <div class="feature-icon">🐙</div>
            <h3>GitHub 仓库</h3>
            <p>查看源码、文档与更新日志</p>
          </div>
          <div class="feature-card" @click="$router.push('/history/watch')" style="cursor:pointer;">
            <div class="feature-icon">📺</div>
            <h3>观看历史</h3>
            <p>快速进入观看记录列表</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">⚙️</div>
            <h3>系统管理</h3>
            <p>完善的用户权限和系统配置管理</p>
          </div>
        </div>
      </div>

      <!-- 系统状态 -->
      <div class="content-card">
        <div class="card-header">
          <h3>系统状态</h3>
          <p>实时监控系统运行状态</p>
        </div>
        <div class="status-grid">
          <div class="status-card">
            <div class="status-icon">💚</div>
            <h4>后端服务</h4>
            <div class="status-info">
              <span :class="['status-badge', healthStatus.status]">{{ healthStatus.status === 'healthy' ? '正常' : '异常' }}</span>
              <p class="status-time">{{ healthStatus.lastCheck }}</p>
            </div>
          </div>
          <div class="status-card">
            <div class="status-icon">🌐</div>
            <h4>API服务</h4>
            <div class="status-info">
              <span :class="['status-badge', apiStatus.status]">{{ apiStatus.status === 'healthy' ? '正常' : '异常' }}</span>
              <p class="status-time">{{ apiStatus.lastCheck }}</p>
            </div>
          </div>
          <div class="status-card">
            <div class="status-icon">👥</div>
            <h4>用户系统</h4>
            <div class="status-info">
              <span class="status-badge healthy">正常</span>
              <p class="status-time">实时</p>
            </div>
          </div>
        </div>

        <div class="quick-actions">
          <h3>快速操作</h3>
          <div class="action-buttons">
            <button @click="router.push('/movie')" class="action-btn primary"><span>🎭</span>观影</button>
            <button @click="router.push('/history/watch')" class="action-btn primary"><span>📺</span>观看历史</button>
            <button @click="router.push('/user-management')" class="action-btn secondary"><span>👥</span>用户管理</button>
            <button @click="router.push('/video-source-management')" class="action-btn secondary"><span>🎬</span>视频资源管理</button>
          </div>
        </div>

        <div v-if="apiResult" class="api-result">
          <h4>API响应结果</h4>
          <pre class="api-response">{{ JSON.stringify(apiResult, null, 2) }}</pre>
        </div>
      </div>

      <!-- 技术栈信息 -->
      <div class="content-card">
        <div class="card-header">
          <h3>技术栈</h3>
          <p>基于最新技术栈构建的高性能系统</p>
        </div>
        <div class="tech-stack-grid">
          <div class="tech-card">
            <div class="tech-icon">⚡</div>
            <h4>前端技术</h4>
            <div class="tech-tags">
              <span class="tech-tag">Vue 3.5.18</span>
              <span class="tech-tag">TypeScript 5.8.0</span>
              <span class="tech-tag">Vite 7.0.6</span>
              <span class="tech-tag">Pinia 3.0.3</span>
              <span class="tech-tag">Ant Design Vue 4.2.6</span>
              <span class="tech-tag">Monaco Editor</span>
            </div>
          </div>
          <div class="tech-card">
            <div class="tech-icon">🔧</div>
            <h4>后端技术</h4>
            <div class="tech-tags">
              <span class="tech-tag">Go 1.24.4</span>
              <span class="tech-tag">Gin 1.10.1</span>
              <span class="tech-tag">JWT</span>
              <span class="tech-tag">Logrus</span>
              <span class="tech-tag">goja (JavaScript 引擎)</span>
            </div>
          </div>
          <div class="tech-card">
            <div class="tech-icon">🧩</div>
            <h4>脚本引擎与解析</h4>
            <div class="tech-tags">
              <span class="tech-tag">Lua (gopher-lua)</span>
              <span class="tech-tag">JavaScript (goja)</span>
              <span class="tech-tag">链式 HTML (goquery)</span>
              <span class="tech-tag">SSE/Chunked 流</span>
              <span class="tech-tag">json_encode/json_decode</span>
              <span class="tech-tag">sleep/split/trim</span>
              <span class="tech-tag">http_get/http_post</span>
            </div>
          </div>
          <div class="tech-card">
            <div class="tech-icon">🎥</div>
            <h4>播放器</h4>
            <div class="tech-tags">
              <span class="tech-tag">video.js</span>
              <span class="tech-tag">@videojs-player/vue</span>
              <span class="tech-tag">HLS.js (m3u8)</span>
              <span class="tech-tag">倍速/长按2x/续播</span>
            </div>
          </div>
          <div class="tech-card">
            <div class="tech-icon">🤝</div>
            <h4>Android + JNI + Wails</h4>
            <div class="tech-tags">
              <span class="tech-tag">JNI 桥接 Go (c-shared)</span>
              <span class="tech-tag">Android WebView 内嵌前端</span>
              <span class="tech-tag">浮窗工具：刷新/首页/退出</span>
              <span class="tech-tag">端口自检与自动重启</span>
              <span class="tech-tag">跨端口持久化历史(AndroidKV)</span>
              <span class="tech-tag">Wails 桌面集成</span>
            </div>
            <p class="tech-note">
              移动端通过 JNI 动态链接 Go 服务，启动后前端以 WebView 直连本地 HTTP。为避免端口变化导致的数据丢失，采用原生 SharedPreferences 暴露的 AndroidKV 进行前端历史的跨端口持久化。桌面端可通过 Wails 一体化构建桌面应用。
            </p>
          </div>
          <div class="tech-card">
            <div class="tech-icon">🎨</div>
            <h4>样式技术</h4>
            <div class="tech-tags">
              <span class="tech-tag">CSS3</span>
              <span class="tech-tag">Flexbox</span>
              <span class="tech-tag">Grid</span>
              <span class="tech-tag">响应式</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </AppLayout>
</template>

<style scoped>
@import './HomeView.css';
.tech-note { margin-top: 8px; color: #475569; line-height: 1.6; }
</style>
