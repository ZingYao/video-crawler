import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { videoSourceAPI } from '@/api'

export interface SearchSite {
  id: string
  name: string
  enabled: boolean
  status?: number // 来自API的状态
  source_type?: number // 来自API的资源类型
  sort?: number // 来自API的排序
}

export interface Settings {
  longPressPlaybackSpeed: number
  progressBarSensitivity: number
  searchSites: SearchSite[]
  allSitesSelected: boolean // 全选标记 - true表示全选模式，false表示手动选择模式
}

const defaultSettings: Settings = {
  longPressPlaybackSpeed: 2.0,
  progressBarSensitivity: 0.7,
  searchSites: [],
  allSitesSelected: true, // 默认全选模式
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Settings>({ ...defaultSettings })
  const isLoading = ref(false)
  const lastUpdateTime = ref<number>(0)

  // 计算属性
  const enabledSearchSites = computed(() => 
    settings.value.searchSites.filter(site => site.enabled)
  )

  const hasEnabledSites = computed(() => 
    settings.value.searchSites.some(site => site.enabled)
  )

  // 检查是否所有网站都被选中
  const isAllSitesSelected = computed(() => 
    settings.value.searchSites.length > 0 && 
    settings.value.searchSites.every(site => site.enabled)
  )

  // 获取缓存的勾选状态
  const getCachedEnabledState = (siteId: string): boolean => {
    const cachedSites = settings.value.searchSites
    const cachedSite = cachedSites.find(site => site.id === siteId)
    
    // 如果是全选模式，所有站点都启用
    if (settings.value.allSitesSelected) {
      return true
    }
    
    // 如果是手动选择模式，返回缓存的勾选状态
    return cachedSite ? cachedSite.enabled : false
  }

  // 从后端加载实际的视频源站点
  const loadActualVideoSources = async () => {
    try {
      const authStore = useAuthStore()
      const configStore = useConfigStore()
      
      // 在无需登录模式下，不需要token也可以调用API
      if (configStore.needsLogin() && !authStore.token) {
        console.warn('[Settings] No auth token available for loading video sources')
        return
      }
      
      const response = await videoSourceAPI.getVideoSourceList(authStore.token || '')
      
      if (response?.code === 0 && Array.isArray(response.data)) {
        // 从API获取站点列表，只包含正常状态的站点
        const actualSites: SearchSite[] = response.data
          .filter((source: any) => source.status === 1) // 只包含正常状态的站点
          .map((source: any) => ({
            id: source.id,
            name: source.name,
            status: source.status,
            source_type: source.source_type,
            sort: source.sort,
            enabled: getCachedEnabledState(source.id) // 使用缓存的勾选状态
          }))
          .sort((a: SearchSite, b: SearchSite) => (b.sort || 0) - (a.sort || 0)) // 按sort降序排列
        
        if (actualSites.length > 0) {
          // 更新站点列表，移除缓存中不存在于API返回列表中的站点
          settings.value.searchSites = actualSites
          
          // 不自动更新全选状态，保持用户的手动设置
          lastUpdateTime.value = Date.now()
          await saveSettings()
        } else {
          console.warn('[Settings] 没有找到可用的视频源站点')
        }
      } else {
        console.warn('[Settings] API返回数据格式不正确:', response)
      }
    } catch (error) {
      console.error('[Settings] Failed to load actual video sources:', error)
      // 如果加载失败，保持现有设置
    }
  }

  // 保存设置到缓存
  const saveSettings = async () => {
    const settingsJson = JSON.stringify(settings.value)
    
    // 检查是否在 Android WebView 环境
    if (window.AndroidKV) {
      // Android 特殊渠道保存
      window.AndroidKV.setItem('app_settings', settingsJson)
    } else {
      // 浏览器环境使用 localStorage
      localStorage.setItem('video_crawler_settings', settingsJson)
    }
  }

  // 从缓存加载设置
  const loadSettings = async () => {
    let settingsJson: string | null = null
    
    // 检查是否在 Android WebView 环境
    if (window.AndroidKV) {
      // Android 特殊渠道加载
      settingsJson = window.AndroidKV.getItem('app_settings')
    } else {
      // 浏览器环境使用 localStorage
      settingsJson = localStorage.getItem('video_crawler_settings')
    }

    if (settingsJson) {
      try {
        const loaded = JSON.parse(settingsJson)
        // 合并默认设置，确保新增字段有默认值
        settings.value = {
          ...defaultSettings,
          ...loaded,
          searchSites: loaded.searchSites || [],
          allSitesSelected: loaded.allSitesSelected !== undefined ? loaded.allSitesSelected : true
        }
      } catch (e) {
        console.error('Failed to parse settings:', e)
        settings.value = { ...defaultSettings }
        await saveSettings() // 保存默认设置到缓存
      }
    } else {
      // 首次加载，使用默认设置
      settings.value = { ...defaultSettings }
      await saveSettings()
    }

    // 无论是否有缓存数据，都尝试获取最新的视频源列表
    await loadActualVideoSources()
  }

  // 更新设置
  const updateSettings = async (newSettings: Partial<Settings>) => {
    settings.value = { ...settings.value, ...newSettings }
    await saveSettings()
  }

  // 更新播放倍速
  const updatePlaybackSpeed = async (speed: number) => {
    settings.value.longPressPlaybackSpeed = Math.max(0.5, Math.min(5.0, speed))
    await saveSettings()
  }

  // 更新进度条敏感度
  const updateProgressSensitivity = async (sensitivity: number) => {
    settings.value.progressBarSensitivity = Math.max(0.1, Math.min(1.5, sensitivity))
    await saveSettings()
  }

  // 更新搜索网站
  const updateSearchSites = async (sites: SearchSite[]) => {
    // 确保至少有一个网站被选中
    if (!sites.some(site => site.enabled)) {
      sites[0].enabled = true
    }
    
    // 更新全选标记 - 只有当所有站点都被选中时，才设置为全选模式
    const allSelected = sites.length > 0 && sites.every(site => site.enabled)
    
    settings.value.searchSites = sites
    settings.value.allSitesSelected = allSelected
    await saveSettings()
  }

  // 切换单个网站状态
  const toggleSearchSite = async (siteId: string) => {
    const site = settings.value.searchSites.find(s => s.id === siteId)
    if (site) {
      // 如果这是最后一个启用的网站，不允许禁用
      if (site.enabled && enabledSearchSites.value.length === 1) {
        return
      }
      site.enabled = !site.enabled
      
      // 更新全选标记 - 如果所有站点都被选中，设置为全选模式
      settings.value.allSitesSelected = isAllSitesSelected.value
      
      await saveSettings()
    }
  }

  // 全选/取消全选搜索网站
  const toggleAllSearchSites = async (enabled: boolean) => {
    settings.value.searchSites.forEach(site => {
      site.enabled = enabled
    })
    // 设置全选标记 - true表示全选模式，false表示手动选择模式
    settings.value.allSitesSelected = enabled
    await saveSettings()
  }

  // 重置为默认设置
  const resetToDefaults = async () => {
    settings.value = { ...defaultSettings }
    await saveSettings()
    // 重新加载视频源列表
    await loadActualVideoSources()
  }

  // 刷新视频源列表
  const refreshVideoSources = async () => {
    isLoading.value = true
    try {
      await loadActualVideoSources()
    } finally {
      isLoading.value = false
    }
  }

  return {
    settings: readonly(settings),
    enabledSearchSites,
    hasEnabledSites,
    isAllSitesSelected,
    isLoading: readonly(isLoading),
    lastUpdateTime: readonly(lastUpdateTime),
    loadSettings,
    updateSettings,
    updatePlaybackSpeed,
    updateProgressSensitivity,
    updateSearchSites,
    toggleSearchSite,
    toggleAllSearchSites,
    resetToDefaults,
    refreshVideoSources,
  }
})
