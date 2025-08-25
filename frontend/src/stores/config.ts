import { defineStore } from 'pinia'
import { ref } from 'vue'
import { configAPI } from '@/utils/api'

// 导入Wails类型
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetConfig(): Promise<Record<string, any>>
          GetServerPort(): Promise<number>
        }
      }
    }
  }
}

export const useConfigStore = defineStore('config', () => {
  const requireLogin = ref(true)
  const env = ref('dev')
  const isLoaded = ref(false)

  // 加载系统配置
  const loadConfig = async () => {
    try {
      console.log('开始加载系统配置...')
      const result = await configAPI.getConfig()
      requireLogin.value = result.require_login
      env.value = result.env
      isLoaded.value = true
      console.log('配置加载成功:', result)
      
      // 如果是Wails模式，记录服务器端口
      if (result.server_port) {
        console.log('Wails HTTP服务端口:', result.server_port)
      }
    } catch (error) {
      console.error('加载系统配置失败:', error)
      // 默认需要登录
      requireLogin.value = true
      isLoaded.value = true
    }
  }

  // 检查是否需要登录
  const needsLogin = () => {
    return requireLogin.value
  }

  return {
    requireLogin,
    env,
    isLoaded,
    loadConfig,
    needsLogin
  }
})
