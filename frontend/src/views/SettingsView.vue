<template>
  <div class="settings-container">
    <h1>设置</h1>
    
    <div class="settings-section">
      <h2>播放控制</h2>
      
      <div class="setting-item">
        <label>长按播放倍速</label>
        <div class="slider-container">
          <input 
            type="range" 
            min="0.5" 
            max="5.0" 
            step="0.1" 
            v-model="playbackSpeed"
            @input="updatePlaybackSpeed"
            class="slider"
          />
          <span class="value-display">{{ playbackSpeed }}x</span>
        </div>
      </div>

      <div class="setting-item">
        <label>进度条移动倍率</label>
        <div class="slider-container">
          <input 
            type="range" 
            min="0.1" 
            max="1.5" 
            step="0.1" 
            v-model="progressSensitivity"
            @input="updateProgressSensitivity"
            class="slider"
          />
          <span class="value-display">{{ progressSensitivity }}x</span>
        </div>
      </div>
    </div>

    <div class="settings-section">
      <h2>搜索网站范围</h2>
      <p class="section-description">选择要搜索的网站，至少需要选择一个</p>
      
      <div class="search-sites-controls">
        <button 
          @click="selectAll" 
          class="control-btn"
          :class="{ active: isAllSitesSelected }"
        >
          全选
        </button>
        <button 
          @click="selectNone" 
          class="control-btn" 
          :disabled="!hasEnabledSites"
        >
          取消全选
        </button>
        <button 
          @click="refreshVideoSources" 
          class="control-btn refresh-btn"
          :loading="refreshing"
        >
          刷新站点列表
        </button>
        <span v-if="isAllSitesSelected" class="all-selected-indicator">
          ✓ 全选模式（新增网站将自动选中）
        </span>
      </div>

      <div class="search-sites-list">
        <div 
          v-for="site in searchSites" 
          :key="site.id" 
          class="site-item"
        >
          <label class="site-checkbox">
            <input 
              type="checkbox" 
              :checked="site.enabled"
              @change="toggleSite(site.id)"
              :disabled="site.enabled && enabledSitesCount === 1"
            />
            <span class="checkmark"></span>
            <span class="site-name">{{ site.name }}</span>
          </label>
        </div>
      </div>

      <!-- 测试区域：添加新网站 -->
      <div class="test-section">
        <h3>测试：添加新网站</h3>
        <div class="add-site-form">
          <input 
            v-model="newSiteName" 
            placeholder="新网站名称" 
            class="site-input"
          />
          <button @click="addNewSite" class="add-btn">添加网站</button>
        </div>
        <p class="test-note">
          在全选模式下，新添加的网站将自动被选中
        </p>
      </div>
    </div>

    <div class="settings-section">
      <h2>其他</h2>
      <button @click="resetSettings" class="reset-btn">重置为默认设置</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import type { SearchSite } from '@/stores/settings'

const settingsStore = useSettingsStore()

// 响应式数据
const playbackSpeed = ref(2.0)
const progressSensitivity = ref(0.7)
const searchSites = ref<SearchSite[]>([])
const newSiteName = ref('')
const refreshing = ref(false)

// 计算属性
const enabledSitesCount = computed(() => 
  searchSites.value.filter(site => site.enabled).length
)

const hasEnabledSites = computed(() => enabledSitesCount.value > 0)

const isAllSitesSelected = computed(() => 
  settingsStore.isAllSitesSelected
)

// 初始化
onMounted(async () => {
  await settingsStore.loadSettings()
  playbackSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
  searchSites.value = [...settingsStore.settings.searchSites]
})

// 更新播放倍速
const updatePlaybackSpeed = async () => {
  await settingsStore.updatePlaybackSpeed(playbackSpeed.value)
}

// 更新进度条敏感度
const updateProgressSensitivity = async () => {
  await settingsStore.updateProgressSensitivity(progressSensitivity.value)
}

// 切换网站状态
const toggleSite = async (siteId: string) => {
  await settingsStore.toggleSearchSite(siteId)
  // 更新本地状态
  const site = searchSites.value.find(s => s.id === siteId)
  if (site) {
    site.enabled = !site.enabled
  }
}

// 全选
const selectAll = async () => {
  await settingsStore.toggleAllSearchSites(true)
  searchSites.value.forEach(site => site.enabled = true)
}

// 取消全选
const selectNone = async () => {
  if (enabledSitesCount.value > 1) {
    await settingsStore.toggleAllSearchSites(false)
    searchSites.value.forEach(site => site.enabled = false)
  }
}

// 添加新网站（测试功能）
const addNewSite = async () => {
  if (newSiteName.value.trim()) {
    const newSite: SearchSite = {
      id: `site${Date.now()}`,
      name: newSiteName.value.trim(),
      enabled: settingsStore.settings.allSitesSelected // 根据全选状态决定是否启用
    }
    
    await settingsStore.addSearchSite(newSite)
    searchSites.value = [...settingsStore.settings.searchSites]
    newSiteName.value = ''
  }
}

// 刷新视频源
const refreshVideoSources = async () => {
  refreshing.value = true
  try {
    await settingsStore.refreshVideoSources()
    searchSites.value = [...settingsStore.settings.searchSites]
  } catch (error) {
    console.error('Failed to refresh video sources:', error)
  } finally {
    refreshing.value = false
  }
}

// 重置设置
const resetSettings = async () => {
  await settingsStore.resetToDefaults()
  playbackSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
  searchSites.value = [...settingsStore.settings.searchSites]
}
</script>

<style scoped>
.settings-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  color: #333;
  margin-bottom: 30px;
  text-align: center;
}

.settings-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

h2 {
  color: #555;
  margin-bottom: 15px;
  font-size: 1.2em;
}

.section-description {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 15px;
}

.setting-item {
  margin-bottom: 20px;
}

.setting-item label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

.slider-container {
  display: flex;
  align-items: center;
  gap: 15px;
}

.slider {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: #ddd;
  outline: none;
  -webkit-appearance: none;
}

.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #10b981;
  cursor: pointer;
}

.slider::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #10b981;
  cursor: pointer;
  border: none;
}

.value-display {
  min-width: 50px;
  text-align: center;
  font-weight: 500;
  color: #10b981;
}

.search-sites-controls {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.control-btn {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  color: #333;
  cursor: pointer;
  transition: all 0.2s;
}

.control-btn:hover:not(:disabled) {
  background: #f5f5f5;
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.control-btn.active {
  background: #10b981;
  color: white;
  border-color: #10b981;
}

.refresh-btn {
  background: #f59e0b; /* A more distinct color for refresh */
  color: white;
  border-color: #f59e0b;
}

.refresh-btn:hover:not(:disabled) {
  background: #d97706;
  border-color: #d97706;
}

.all-selected-indicator {
  color: #10b981;
  font-size: 0.9em;
  font-weight: 500;
  margin-left: 10px;
}

.search-sites-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin-bottom: 20px;
}

.site-item {
  padding: 10px;
  border: 1px solid #eee;
  border-radius: 4px;
  background: #fafafa;
}

.site-checkbox {
  display: flex;
  align-items: center;
  cursor: pointer;
  position: relative;
}

.site-checkbox input[type="checkbox"] {
  margin-right: 8px;
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.site-checkbox input[type="checkbox"]:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.site-name {
  font-size: 0.9em;
  color: #333;
}

.test-section {
  border-top: 1px solid #eee;
  padding-top: 20px;
  margin-top: 20px;
}

.test-section h3 {
  color: #555;
  margin-bottom: 15px;
  font-size: 1.1em;
}

.add-site-form {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.site-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.9em;
}

.add-btn {
  padding: 8px 16px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9em;
  transition: background 0.2s;
}

.add-btn:hover {
  background: #2563eb;
}

.test-note {
  color: #666;
  font-size: 0.8em;
  font-style: italic;
}

.reset-btn {
  padding: 12px 24px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1em;
  transition: background 0.2s;
}

.reset-btn:hover {
  background: #dc2626;
}

@media (max-width: 768px) {
  .settings-container {
    padding: 15px;
  }
  
  .search-sites-list {
    grid-template-columns: 1fr;
  }
  
  .slider-container {
    flex-direction: column;
    align-items: stretch;
  }
  
  .value-display {
    text-align: center;
  }
  
  .search-sites-controls {
    flex-direction: column;
    align-items: stretch;
  }
  
  .all-selected-indicator {
    margin-left: 0;
    margin-top: 5px;
  }
  
  .add-site-form {
    flex-direction: column;
  }
}
</style>
