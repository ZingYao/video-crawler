<template>
  <AppLayout page-title="设置">
    <div class="settings-container">
      <div class="global-hint">
        提示：可点击各滑动条右侧的数值直接输入精准数值，按回车提交，Esc 取消。
      </div>
      <div class="settings-section">
        <h2>播放控制</h2>
        
        <div class="setting-item">
          <p class="item-hint">仅在“播放页面”中生效：长按（触摸/鼠标左键）临时加速到该倍速，松开恢复。</p>
          <label class="setting-title">
            <span>长按播放倍速</span>
            <a-tooltip title="重置当前项目">
              <a-button danger type="primary" size="small" @click="resetPlaybackSpeed" class="reset-btn">
                <template #icon>
                  <ReloadOutlined />
                </template>
              </a-button>
            </a-tooltip>
          </label>
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
            <template v-if="!isEditingPlayback">
              <span class="value-display" @click="startPlaybackEdit">{{ playbackSpeed }}x</span>
            </template>
            <template v-else>
              <input
                type="number"
                class="number-input"
                min="0.5"
                max="5.0"
                step="0.1"
                v-model.number="tempPlaybackSpeed"
                @blur="commitPlaybackEdit"
                @keyup.enter="commitPlaybackEdit"
                @keyup.esc="cancelPlaybackEdit"
                ref="playbackInputRef"
              />
            </template>
          </div>
        </div>

        <div class="setting-item">
          <p class="item-hint">仅在“播放页面”中生效：影响拖动进度条/左右方向键快进快退的灵敏度，数值越大移动越快。</p>
          <label class="setting-title">
            <span>进度条移动倍率</span>
            <a-tooltip title="重置当前项目">
              <a-button danger type="primary" size="small" @click="resetProgressSensitivity" class="reset-btn">
                <template #icon>
                  <ReloadOutlined />
                </template>
              </a-button>
            </a-tooltip>
          </label>
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
            <template v-if="!isEditingProgress">
              <span class="value-display" @click="startProgressEdit">{{ progressSensitivity }}x</span>
            </template>
            <template v-else>
              <input
                type="number"
                class="number-input"
                min="0.1"
                max="1.5"
                step="0.1"
                v-model.number="tempProgressSensitivity"
                @blur="commitProgressEdit"
                @keyup.enter="commitProgressEdit"
                @keyup.esc="cancelProgressEdit"
                ref="progressInputRef"
              />
            </template>
          </div>
        </div>
      </div>

      <div class="settings-section">
        <h2>显示与缩放</h2>
        <p class="section-description">此功能通过缩放整个页面来适配大屏/高 DPI。可能会导致个别页面布局异常，请谨慎使用。</p>
        <div class="setting-item">
          <p class="item-hint">全局生效：缩放整个页面元素，适配大屏或高 DPI 显示比例。</p>
          <label class="setting-title">
            <span>页面缩放</span>
            <a-tooltip title="重置当前项目">
              <a-button danger type="primary" size="small" @click="resetPageScale" class="reset-btn">
                <template #icon>
                  <ReloadOutlined />
                </template>
              </a-button>
            </a-tooltip>
          </label>
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
            <template v-if="!isEditingScale">
              <span class="value-display" @click="startScaleEdit">{{ pageScale.toFixed(2) }}x</span>
            </template>
            <template v-else>
              <input
                type="number"
                class="number-input"
                min="0.5"
                max="1.5"
                step="0.01"
                v-model.number="tempPageScale"
                @blur="commitScaleEdit"
                @keyup.enter="commitScaleEdit"
                @keyup.esc="cancelScaleEdit"
                ref="scaleInputRef"
              />
            </template>
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
import { ref, computed, onMounted, h, nextTick } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import AppLayout from '@/components/AppLayout.vue'
import type { SearchSite } from '@/stores/settings'
import { Modal, message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { applyPageScale } from '@/utils/zoom'

const settingsStore = useSettingsStore()

// 响应式数据
const playbackSpeed = ref(2.0)
const progressSensitivity = ref(0.7)
const pageScale = ref(1)
const isEditingScale = ref(false)
const tempPageScale = ref(1)
const isEditingPlayback = ref(false)
const tempPlaybackSpeed = ref(2.0)
const isEditingProgress = ref(false)
const tempProgressSensitivity = ref(0.7)
const scaleInputRef = ref<HTMLInputElement | null>(null)
const playbackInputRef = ref<HTMLInputElement | null>(null)
const progressInputRef = ref<HTMLInputElement | null>(null)

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
  tempPageScale.value = pageScale.value
})

// 更新播放倍速
const updatePlaybackSpeed = async () => {
  await settingsStore.updatePlaybackSpeed(playbackSpeed.value)
}

const resetPlaybackSpeed = async () => {
  playbackSpeed.value = 2.0
  await settingsStore.updatePlaybackSpeed(playbackSpeed.value)
}

// 更新进度条敏感度
const updateProgressSensitivity = async () => {
  await settingsStore.updateProgressSensitivity(progressSensitivity.value)
}

const resetProgressSensitivity = async () => {
  progressSensitivity.value = 0.7
  await settingsStore.updateProgressSensitivity(progressSensitivity.value)
}

// 更新缩放（滑动结束生效）
const onScaleChangeEnd = async () => {
  await settingsStore.updatePageScale(pageScale.value)
  applyPageScale(pageScale.value)
}

const resetPageScale = async () => {
  pageScale.value = 1
  await settingsStore.updatePageScale(pageScale.value)
  applyPageScale(pageScale.value)
}

// 点击数值进入编辑
const startScaleEdit = () => {
  tempPageScale.value = pageScale.value
  isEditingScale.value = true
  nextTick(() => { try { scaleInputRef.value?.focus() } catch {} })
}
// 提交编辑
const commitScaleEdit = async () => {
  let v = Number(tempPageScale.value)
  if (isNaN(v)) v = pageScale.value
  v = Math.max(0.5, Math.min(1.5, v))
  pageScale.value = Number(v.toFixed(2))
  isEditingScale.value = false
  
  // 退出输入框焦点
  if (scaleInputRef.value) {
    scaleInputRef.value.blur()
  }
  
  await onScaleChangeEnd()
}
// 取消编辑
const cancelScaleEdit = () => {
  isEditingScale.value = false
}

// 播放倍速编辑
const startPlaybackEdit = () => {
  tempPlaybackSpeed.value = playbackSpeed.value
  isEditingPlayback.value = true
  nextTick(() => { try { playbackInputRef.value?.focus() } catch {} })
}
const commitPlaybackEdit = async () => {
  let v = Number(tempPlaybackSpeed.value)
  if (isNaN(v)) v = playbackSpeed.value
  v = Math.max(0.5, Math.min(5.0, v))
  playbackSpeed.value = Number(v.toFixed(1))
  isEditingPlayback.value = false
  
  // 退出输入框焦点
  if (playbackInputRef.value) {
    playbackInputRef.value.blur()
  }
  
  await updatePlaybackSpeed()
}
const cancelPlaybackEdit = () => { isEditingPlayback.value = false }

// 进度灵敏度编辑
const startProgressEdit = () => {
  tempProgressSensitivity.value = progressSensitivity.value
  isEditingProgress.value = true
  nextTick(() => { try { progressInputRef.value?.focus() } catch {} })
}
const commitProgressEdit = async () => {
  let v = Number(tempProgressSensitivity.value)
  if (isNaN(v)) v = progressSensitivity.value
  v = Math.max(0.1, Math.min(1.5, v))
  progressSensitivity.value = Number(v.toFixed(1))
  isEditingProgress.value = false
  
  // 退出输入框焦点
  if (progressInputRef.value) {
    progressInputRef.value.blur()
  }
  
  await updateProgressSensitivity()
}
const cancelProgressEdit = () => { isEditingProgress.value = false }

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

.global-hint {
  margin: 0 0 12px 0;
  padding: 10px 12px;
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
  border-radius: 6px;
  font-size: 13px;
}

.settings-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 统一卡片内部间距与排版 */
.settings-section .setting-item + .setting-item {
  margin-top: 16px;
}
.settings-section .setting-title > span {
  font-weight: 500;
  color: #374151;
}
.settings-section .slider-container {
  align-items: center;
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

.setting-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  overflow: visible;
}

.reset-btn {
  margin-left: 8px;
}

/* 确保小按钮完整显示并垂直居中 */
.setting-title :deep(.ant-btn) {
  height: 24px;
  width: 24px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 24px;
}
.setting-title :deep(.ant-btn .anticon) {
  font-size: 14px;
  line-height: 1;
}

.slider-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.slider {
  flex: 0 0 260px; /* 统一滑块长度，与其它模块一致 */
  height: 6px;
  border-radius: 3px;
  background: #ddd;
  outline: none;
  -webkit-appearance: none;
}

.mini-btn {
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  background: #fff;
  color: #374151;
  cursor: pointer;
}
.mini-btn:hover {
  background: #f3f4f6;
}

.number-input {
  width: 80px;
  padding: 6px 8px;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  color: #000;
}

/* 显示与缩放模块优化 */

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

.item-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
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
  
  .slider-container { flex-direction: column; align-items: stretch; }
  .slider { flex: 1 1 auto; width: 100%; }
  
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
