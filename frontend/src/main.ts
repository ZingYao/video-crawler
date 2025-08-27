import './assets/main.css'
import './styles/index.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import { ConfigProvider } from 'ant-design-vue'

import App from './App.vue'
import router from './router'
import { isWails } from './utils/wails'

const app = createApp(App)

// 配置Ant Design主题
const theme = {
  token: {
    colorPrimary: '#10b981',
    colorPrimaryHover: '#34d399',
    colorPrimaryActive: '#059669',
    colorSuccess: '#10b981',
    colorInfo: '#10b981',
    colorWarning: '#f59e0b',
    colorError: '#ef4444',
  },
}

app.use(createPinia())
app.use(router)
app.use(Antd)

// 全局配置主题
app.use(ConfigProvider, theme)

// 检查是否在Wails环境中，如果是则重定向到启动页面
const checkWailsEnvironment = () => {
  const isWailsEnv = isWails()

  // 调试信息（生产环境构建时会自动移除）
  console.error('=== main.ts 环境检测 ===')
  console.error('isWails:', isWailsEnv)
  console.error('window.location.pathname:', window.location.pathname)

  if (isWailsEnv && window.location.pathname === '/') {
    console.error('检测到Wails环境，重定向到启动页面')
    // 延迟重定向，确保路由已经初始化
    setTimeout(() => {
      router.push('/startup')
    }, 100)
  } else {
    console.error('非Wails环境或不在根路径，不重定向')
  }
}

// 只在页面初始加载时检查一次，避免重复检查导致循环跳转
let hasCheckedWailsEnvironment = false

const performWailsEnvironmentCheck = () => {
  if (hasCheckedWailsEnvironment) {
    return
  }
  
  const isWailsEnv = isWails()
  if (isWailsEnv && window.location.pathname === '/') {
    hasCheckedWailsEnvironment = true
    console.error('检测到Wails环境，重定向到启动页面')
    setTimeout(() => {
      router.push('/startup')
    }, 100)
  }
}

// 立即检查一次
performWailsEnvironmentCheck()

// 延迟检查，以防 window.go 在页面加载后才被注入
setTimeout(() => {
  console.error('=== 延迟环境检测 ===')
  performWailsEnvironmentCheck()
}, 1000)

// 再次延迟检查
setTimeout(() => {
  console.error('=== 再次延迟环境检测 ===')
  performWailsEnvironmentCheck()
}, 3000)

app.mount('#app')
