import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'

export interface SearchSite {
  id: string
  name: string
  enabled: boolean
}

export interface Settings {
  longPressPlaybackSpeed: number
  progressBarSensitivity: number
  searchSites: SearchSite[]
  allSitesSelected: boolean // 新增：全选标记
}

const defaultSearchSites: SearchSite[] = [
  { id: 'site1', name: '网站1', enabled: true },
  { id: 'site2', name: '网站2', enabled: true },
  { id: 'site3', name: '网站3', enabled: true },
  { id: 'site4', name: '网站4', enabled: true },
  { id: 'site5', name: '网站5', enabled: true },
]

const defaultSettings: Settings = {
  longPressPlaybackSpeed: 2.0,
  progressBarSensitivity: 0.7,
  searchSites: defaultSearchSites,
  allSitesSelected: true, // 默认全选
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Settings>({ ...defaultSettings })

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

  // 从后端加载实际的视频源站点
  const loadActualVideoSources = async () => {
    try {
      console.log('[Settings] 开始加载实际视频源...')
      
      const authStore = useAuthStore()
      
      console.log('[Settings] Auth token:', authStore.token ? '存在' : '不存在')
      
      if (!authStore.token) {
        console.warn('[Settings] No auth token available for loading video sources')
        return
      }
      
      console.log('[Settings] 调用 videoSourceAPI.getVideoSourceList...')
      const response = await videoSourceAPI.getVideoSourceList(authStore.token)
      console.log('[Settings] API响应:', response)
      
      if (response?.code === 0 && Array.isArray(response.data)) {
        // 获取当前缓存的勾选状态
        const currentSites = settings.value.searchSites
        const currentEnabledMap = new Map(currentSites.map(site => [site.id, site.enabled]))
        
        // 从API获取站点列表，并合并勾选状态
        const actualSites: SearchSite[] = response.data
          .filter((source: any) => source.status === 1) // 只包含正常状态的站点
          .map((source: any) => ({
            id: source.id,
            name: source.name,
            enabled: currentEnabledMap.has(source.id) 
              ? currentEnabledMap.get(source.id)! // 使用缓存的勾选状态
              : settings.value.allSitesSelected // 新站点根据全选状态决定是否启用
          }))
        
        console.log('[Settings] 合并后的站点:', actualSites)
        
        if (actualSites.length > 0) {
          settings.value.searchSites = actualSites
          // 更新全选状态
          settings.value.allSitesSelected = actualSites.every(site => site.enabled)
          await saveSettings()
          console.log('[Settings] 成功更新站点列表')
        } else {
          console.warn('[Settings] 没有找到可用的视频源站点')
        }
      } else {
        console.warn('[Settings] API返回数据格式不正确:', response)
      }
    } catch (error) {
      console.error('[Settings] Failed to load actual video sources:', error)
      // 如果加载失败，保持默认设置
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
          searchSites: loaded.searchSites || defaultSearchSites,
          allSitesSelected: loaded.allSitesSelected !== undefined ? loaded.allSitesSelected : true
        }
        
        // 如果之前是全选状态，确保所有网站都被选中
        if (settings.value.allSitesSelected) {
          settings.value.searchSites.forEach(site => {
            site.enabled = true
          })
        }
        
        // 无论是否有缓存数据，都尝试获取最新的视频源列表
        await loadActualVideoSources()
      } catch (e) {
        console.error('Failed to parse settings:', e)
        settings.value = { ...defaultSettings }
        await saveSettings() // 保存默认设置到缓存
        // 尝试获取实际的视频源
        await loadActualVideoSources()
      }
    } else {
      // 首次加载，使用默认设置
      settings.value = { ...defaultSettings }
      // 尝试获取实际的视频源
      await loadActualVideoSources()
      // 确保默认设置被保存到缓存
      await saveSettings()
    }
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
    
    // 更新全选标记
    const allSelected = sites.length > 0 && sites.every(site => site.enabled)
    
    settings.value.searchSites = sites
    settings.value.allSitesSelected = allSelected
    await saveSettings()
  }

  // 添加新网站
  const addSearchSite = async (site: SearchSite) => {
    const newSites = [...settings.value.searchSites, site]
    
    // 如果当前是全选状态，新网站也应该被选中
    if (settings.value.allSitesSelected) {
      site.enabled = true
    }
    
    await updateSearchSites(newSites)
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
      
      // 更新全选标记
      settings.value.allSitesSelected = isAllSitesSelected.value
      
      await saveSettings()
    }
  }

  // 全选/取消全选搜索网站
  const toggleAllSearchSites = async (enabled: boolean) => {
    settings.value.searchSites.forEach(site => {
      site.enabled = enabled
    })
    settings.value.allSitesSelected = enabled
    await saveSettings()
  }

  // 重置为默认设置
  const resetToDefaults = async () => {
    settings.value = { ...defaultSettings }
    await saveSettings()
  }

  // 刷新视频源列表
  const refreshVideoSources = async () => {
    await loadActualVideoSources()
  }

  return {
    settings: readonly(settings),
    enabledSearchSites,
    hasEnabledSites,
    isAllSitesSelected,
    loadSettings,
    updateSettings,
    updatePlaybackSpeed,
    updateProgressSensitivity,
    updateSearchSites,
    addSearchSite,
    toggleSearchSite,
    toggleAllSearchSites,
    resetToDefaults,
    refreshVideoSources,
  }
})
