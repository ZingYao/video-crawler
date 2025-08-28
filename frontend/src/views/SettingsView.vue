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
        <h2>显示与缩放</h2>
        <div class="setting-item">
          <label>页面缩放（解决大屏/高DPI比例异常）</label>
          <div class="slider-container">
            <input 
              type="range" 
              min="0.5" 
              max="1.5" 
              step="0.01" 
              v-model.number="pageScale"
              @change="onScaleChangeEnd"
              class="slider"
            />
            <span class="value-display">{{ pageScale.toFixed(2) }}x</span>
          </div>
        </div>
      </div>

      <div class="settings-section">
        <h2>虚拟光标</h2>
        <div class="setting-item">
          <label>启用虚拟光标（默认开启）</label>
          <label class="switch">
            <input type="checkbox" :checked="virtualCursorEnabled" @change="onToggleVirtualCursor">
            <span class="slider-round"></span>
          </label>
        </div>
        <div class="setting-item">
          <button class="control-btn" @click="showCursorUsage">查看使用说明</button>
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
import { ref, computed, onMounted, h } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import AppLayout from '@/components/AppLayout.vue'
import type { SearchSite } from '@/stores/settings'
import { Modal, message } from 'ant-design-vue'
import { applyPageScale } from '@/utils/zoom'

const settingsStore = useSettingsStore()

// 响应式数据
const playbackSpeed = ref(2.0)
const progressSensitivity = ref(0.7)
const pageScale = ref(1)

// 计算属性
const searchSites = computed(() => settingsStore.settings.searchSites)
const enabledSitesCount = computed(() => 
  searchSites.value.filter(site => site.enabled).length
)
const hasEnabledSites = computed(() => enabledSitesCount.value > 0)
const isAllSitesSelected = computed(() => settingsStore.isAllSitesSelected)
const lastUpdateTime = computed(() => settingsStore.lastUpdateTime)
const virtualCursorEnabled = computed(() => settingsStore.settings.virtualCursorEnabled)

// 初始化
onMounted(async () => {
  // 加载设置（包括缓存数据和API数据合并）
  await settingsStore.loadSettings()
  
  // 更新本地状态
  playbackSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
  pageScale.value = settingsStore.settings.pageScale ?? 1
})

// 更新播放倍速
const updatePlaybackSpeed = async () => {
  await settingsStore.updatePlaybackSpeed(playbackSpeed.value)
}

// 更新进度条敏感度
const updateProgressSensitivity = async () => {
  await settingsStore.updateProgressSensitivity(progressSensitivity.value)
}

// 更新缩放（滑动结束生效）
const onScaleChangeEnd = async () => {
  await settingsStore.updatePageScale(pageScale.value)
  applyPageScale(pageScale.value)
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

// 切换虚拟光标
const onToggleVirtualCursor = async (e: Event) => {
  const target = e.target as HTMLInputElement
  const enabled = !!target.checked
  
  // 保存设置
  await settingsStore.setVirtualCursorEnabled(enabled)
  
  // 动态启用/禁用虚拟光标
  try {
    const { enableVirtualMouse, disableVirtualMouse } = await import('@/utils/virtualMouse')
    
    if (enabled) {
      // 启用虚拟光标
      const initOptions = { baseSpeed: 120, maxSpeed: 600, accelerateIntervalMs: 240, accelerateFactor: 1.35, cursorSize: 22 }
      enableVirtualMouse(initOptions)
      message.success('虚拟光标已启用')
    } else {
      // 禁用虚拟光标
      disableVirtualMouse()
      message.success('虚拟光标已禁用')
    }
  } catch (error) {
    console.error('虚拟光标切换失败:', error)
    message.error('虚拟光标切换失败，请刷新页面重试')
  }
}

// 使用说明弹窗（更美观）
const showCursorUsage = () => {
  Modal.info({
    title: '虚拟光标使用说明',
    width: 520,
    content: h('div', { style: 'line-height:1.75;color:#374151' }, [
      h('p', { style: 'margin:4px 0 10px;color:#6b7280' }, '用遥控器/键盘即可操控页面元素：'),
      h('ul', { style: 'padding-left:18px;margin:0' }, [
        h('li', '上下左右：移动光标（指数加速，越界限制）'),
        h('li', '回车 / Enter / OK：在光标位置点击'),
        h('li', '双击 上/下：页面上下翻动'),
        h('li', '双击 左/右：页面左右翻动（优先滚动横向容器）'),
        h('li', 'Esc / 返回：退出输入框焦点并恢复光标'),
        h('li', '输入框聚焦：光标隐藏，方向键仅移动文本光标'),
        h('li', '无操作10秒自动隐藏（1秒淡出），操作后0.25秒淡入'),
      ]),
      h('div', { style: 'margin-top:15px;padding:10px;background:#f3f4f6;border-radius:6px;border-left:4px solid #10b981' }, [
        h('p', { style: 'margin:0 0 8px;font-weight:600;color:#374151' }, '💡 智能隐藏功能'),
        h('p', { style: 'margin:0;color:#6b7280;font-size:14px' }, '• 检测到鼠标活动时自动隐藏虚拟光标'),
        h('p', { style: 'margin:0;color:#6b7280;font-size:14px' }, '• 鼠标无活动3秒后重新显示虚拟光标'),
        h('p', { style: 'margin:0;color:#6b7280;font-size:14px' }, '• 如何关闭：进入"设置"页面，关闭"启用虚拟光标"开关'),
      ]),
    ]),
    okText: '已了解',
    maskClosable: true,
    centered: true,
  })
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
.switch {
  position: relative;
  display: inline-block;
  width: 46px;
  height: 26px;
}
.switch input { display:none; }
.slider-round {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #ccc;
  transition: .2s;
  border-radius: 26px;
}
.slider-round:before {
  position: absolute;
  content: "";
  height: 20px; width: 20px;
  left: 3px; bottom: 3px;
  background-color: white;
  transition: .2s;
  border-radius: 50%;
}
.switch input:checked + .slider-round {
  background-color: #10b981;
}
.switch input:checked + .slider-round:before {
  transform: translateX(20px);
}
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
