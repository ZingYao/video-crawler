import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useConfigStore = defineStore('config', () => {
  const requireLogin = ref(true)
  const env = ref('dev')
  const isLoaded = ref(false)

  // 加载系统配置
  const loadConfig = async () => {
    try {
      const response = await fetch('/api/config')
      const result = await response.json()
      
      if (result.code === 0) {
        requireLogin.value = result.data.require_login
        env.value = result.data.env
        isLoaded.value = true
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
