import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { configAPI } from '@/api'

export interface Config {
  requireLogin: boolean
  env: string
  serverPort?: number
}

export const useConfigStore = defineStore('config', () => {
  const config = ref<Config>({
    requireLogin: true,
    env: 'production'
  })

  const isLoading = ref(false)
  const isLoaded = ref(false)

  const needsLogin = () => config.value.requireLogin

  const loadConfig = async () => {
    isLoading.value = true
    try {
      const result = await configAPI.getConfig()
      console.error('loadConfig result:', result)
      if (result?.code === 0 && result.data) {
        config.value = {
          requireLogin: result.data.require_login || false ,
          env: result.data.env || 'production',
          serverPort: result.data.server_port
        }
      }
      isLoaded.value = true
    } catch (error) {
      console.error('Failed to load config:', error)
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  return {
    config: computed(() => config.value),
    isLoading: computed(() => isLoading.value),
    isLoaded: computed(() => isLoaded.value),
    needsLogin,
    loadConfig
  }
})
