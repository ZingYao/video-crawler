<template>
  <AppLayout page-title="设置">
    <div class="settings-container">
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
            :disabled="settingsStore.isLoading"
          >
            <span v-if="settingsStore.isLoading">刷新中...</span>
            <span v-else>刷新站点列表</span>
          </button>
          <span v-if="isAllSitesSelected" class="all-selected-indicator">
            ✓ 全选模式（搜索时将包含所有正常状态的站点）
          </span>
        </div>

        <div v-if="settingsStore.isLoading" class="loading-indicator">
          <div class="loading-spinner"></div>
          <span>正在加载站点列表...</span>
        </div>

        <div v-else-if="searchSites.length === 0" class="empty-state">
          <p>暂无可用的视频源站点</p>
          <button @click="refreshVideoSources" class="control-btn refresh-btn">
            重新加载
          </button>
        </div>

        <div v-else class="search-sites-list">
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
              <div class="site-info">
                <span class="site-name">{{ site.name }}</span>
                <span class="site-meta">
                  <span class="site-type">{{ getSourceTypeText(site.source_type) }}</span>
                  <span class="site-sort">排序: {{ site.sort || 0 }}</span>
                </span>
              </div>
            </label>
          </div>
        </div>

        <div v-if="lastUpdateTime" class="update-info">
          <small>最后更新: {{ formatUpdateTime(lastUpdateTime) }}</small>
        </div>
      </div>

      <div class="settings-section">
        <h2>其他</h2>
        <button @click="resetSettings" class="reset-btn">重置为默认设置</button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import AppLayout from '@/components/AppLayout.vue'
import type { SearchSite } from '@/stores/settings'

const settingsStore = useSettingsStore()

// 响应式数据
const playbackSpeed = ref(2.0)
const progressSensitivity = ref(0.7)

// 计算属性
const searchSites = computed(() => settingsStore.settings.searchSites)
const enabledSitesCount = computed(() => 
  searchSites.value.filter(site => site.enabled).length
)
const hasEnabledSites = computed(() => enabledSitesCount.value > 0)
const isAllSitesSelected = computed(() => settingsStore.isAllSitesSelected)
const lastUpdateTime = computed(() => settingsStore.lastUpdateTime)

// 初始化
onMounted(async () => {
  // 加载设置（包括缓存数据和API数据合并）
  await settingsStore.loadSettings()
  
  // 更新本地状态
  playbackSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
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
}

// 全选
const selectAll = async () => {
  await settingsStore.toggleAllSearchSites(true)
}

// 取消全选
const selectNone = async () => {
  if (enabledSitesCount.value > 1) {
    await settingsStore.toggleAllSearchSites(false)
  }
}

// 刷新视频源
const refreshVideoSources = async () => {
  await settingsStore.refreshVideoSources()
}

// 重置设置
const resetSettings = async () => {
  await settingsStore.resetToDefaults()
  playbackSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
}

// 获取资源类型文本
const getSourceTypeText = (sourceType?: number) => {
  switch (sourceType) {
    case 0: return '综合'
    case 1: return '电影'
    case 2: return '电视剧'
    case 3: return '动漫'
    case 4: return '综艺'
    case 5: return '纪录片'
    default: return '未知'
  }
}

// 格式化更新时间
const formatUpdateTime = (timestamp: number) => {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN')
}
</script>

<style scoped>
.settings-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
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
  background: #f59e0b;
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

.loading-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px;
  color: #666;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #f3f3f3;
  border-top: 2px solid #10b981;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.empty-state p {
  margin-bottom: 15px;
}

.search-sites-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 10px;
  margin-bottom: 20px;
}

.site-item {
  padding: 12px;
  border: 1px solid #eee;
  border-radius: 6px;
  background: #fafafa;
  transition: all 0.2s;
}

.site-item:hover {
  background: #f0f0f0;
  border-color: #ddd;
}

.site-checkbox {
  display: flex;
  align-items: flex-start;
  cursor: pointer;
  position: relative;
  width: 100%;
}

.site-checkbox input[type="checkbox"] {
  margin-right: 10px;
  width: 18px;
  height: 18px;
  cursor: pointer;
  margin-top: 2px;
}

.site-checkbox input[type="checkbox"]:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.site-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.site-name {
  font-size: 0.95em;
  color: #333;
  font-weight: 500;
}

.site-meta {
  display: flex;
  gap: 10px;
  font-size: 0.8em;
  color: #666;
}

.site-type {
  background: #e5e7eb;
  padding: 2px 6px;
  border-radius: 3px;
}

.site-sort {
  color: #9ca3af;
}

.update-info {
  text-align: center;
  color: #9ca3af;
  font-size: 0.85em;
  margin-top: 10px;
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
  
  .site-meta {
    flex-direction: column;
    gap: 2px;
  }
}
</style>
