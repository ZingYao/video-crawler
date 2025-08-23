<template>
  <AppLayout :page-title="`播放 - ${displayTitle}`">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>视频详情</h2>
          <div class="header-actions">
            <a-space>
              <a-button type="primary" :loading="loading" @click="refreshDetail">重新获取</a-button>
              <a-button @click="goBack">返回</a-button>
            </a-space>
          </div>
        </div>
      </template>

      <div class="watch-view">
        <a-spin v-if="loading" />
        <a-result v-else-if="error" status="error" :title="error" />
        <template v-else>
          <a-alert v-if="fromCache" type="info" show-icon message="已使用本地缓存数据" style="margin-bottom: 12px;" />

          <!-- 播放器区域 -->
          <div class="player-container">
            <!-- 播放器方案显示 -->
            <div class="player-scheme-info">
              <a-tag color="green" size="small">
                <template #icon>
                  <PlayCircleOutlined />
                </template>
                {{ currentPlayerScheme }}
              </a-tag>
            </div>
            
            <div class="player-wrap">
              <video
                ref="videoRef"
                class="plyr-video"
                controls
                preload="metadata"
                :src="playerSource"
                @loadstart="setVideoLoading(true)"
                @canplay="setVideoLoading(false)"
                @canplaythrough="setVideoLoading(false)"
                @waiting="setVideoLoading(true)"
                @error="setVideoLoading(false)"
              />
              
              <!-- 视频加载遮罩 -->
              <div v-if="videoLoading" class="video-loading-mask">
                <div class="loading-content">
                  <a-spin size="large" />
                  <div class="loading-text">视频加载中...</div>
                  <div v-if="networkSpeed" class="network-speed">
                    网络速度: {{ networkSpeed }}
                  </div>
                </div>
              </div>
            </div>
            <div class="player-actions">
              <a-space wrap>
                <a-button size="small" @click="playPrev" :disabled="!canPrev">上一集</a-button>
                <a-button size="small" @click="playNext" :disabled="!canNext">下一集</a-button>
                <!-- 移动端播放速率选择器 -->
                <a-select
                  v-if="isMobile"
                  v-model:value="rate"
                  size="small"
                  style="width: 80px;"
                  :options="rateOptions"
                  @change="handleRateChange"
                />
                <!-- 跳过片首开关 -->
                <a-switch
                  v-model:checked="skipIntro.enabled"
                  size="small"
                  @change="handleSkipIntroChange"
                >
                  <template #checkedChildren>跳过片首</template>
                  <template #unCheckedChildren>跳过片首</template>
                </a-switch>
                <!-- 跳过片首秒数输入 -->
                <a-input-number
                  v-if="skipIntro.enabled"
                  v-model:value="skipIntro.seconds"
                  size="small"
                  :min="1"
                  :max="300"
                  style="width: 80px;"
                  placeholder="秒数"
                  @change="handleSkipIntroChange"
                />
                <!-- 跳过片尾开关 -->
                <a-switch
                  v-model:checked="skipOutro.enabled"
                  size="small"
                  @change="handleSkipOutroChange"
                >
                  <template #checkedChildren>跳过片尾</template>
                  <template #unCheckedChildren>跳过片尾</template>
                </a-switch>
                <!-- 跳过片尾秒数输入 -->
                <a-input-number
                  v-if="skipOutro.enabled"
                  v-model:value="skipOutro.seconds"
                  size="small"
                  :min="1"
                  :max="300"
                  style="width: 80px;"
                  placeholder="秒数"
                  @change="handleSkipOutroChange"
                />
                <a-button 
                  size="small" 
                  type="primary" 
                  @click="downloadWithThunder" 
                  :disabled="!playerSource"
                  :loading="downloading"
                >
                  <template #icon>
                    <ThunderboltOutlined />
                  </template>
                  迅雷下载
                </a-button>
                <a-button size="small" type="default" @click="goOriginal" :disabled="!originalUrl">原站点</a-button>
              </a-space>
            </div>
          </div>

          <div class="detail-layout">
            <div class="detail-main">
              <a-card size="small" title="站点与剧集" :bordered="true" v-if="sourcesByTab.length" style="margin-bottom: 12px;">
                <a-tabs v-model:activeKey="activeSourceKey">
                  <a-tab-pane v-for="(s, idx) in sourcesByTab" :key="String(idx)" :tab="s.name">
                    <div class="ep-list">
                      <a-button
                        v-for="(ep, eidx) in s.episodes"
                        :key="eidx"
                        size="small"
                        class="ep-btn"
                        :type="isCurrentEpisode(ep) ? 'primary' : 'default'"
                        @click="playEpisode(ep, s.name)"
                      >{{ ep.name }}</a-button>
                    </div>
                  </a-tab-pane>
                </a-tabs>
              </a-card>

              <a-card size="small" title="基础信息" :bordered="true" style="margin-bottom: 12px;">
                <div class="kv-list">
                  <div class="kv-item"><span class="k">名称</span><span class="v">{{ base.name || '-' }}</span></div>
                  <div class="kv-item"><span class="k">导演</span><span class="v">{{ base.director || '-' }}</span></div>
                  <div class="kv-item"><span class="k">主演</span><span class="v">{{ base.actor || '-' }}</span></div>
                  <div class="kv-item"><span class="k">语言</span><span class="v">{{ base.language || '-' }}</span></div>
                  <div class="kv-item"><span class="k">地区</span><span class="v">{{ base.region || '-' }}</span></div>
                  <div class="kv-item"><span class="k">上映日期</span><span class="v">{{ base.release_date || '-' }}</span></div>
                  <div class="kv-item"><span class="k">评分</span><span class="v">{{ base.rate || '-' }}</span></div>
                  <div class="kv-item"><span class="k">类型</span><span class="v">{{ base.type || '-' }}</span></div>
                </div>
              </a-card>

              <a-card size="small" title="描述" :bordered="true" style="margin-bottom: 12px;">
                <div class="desc">{{ base.description || '-' }}</div>
              </a-card>

              <a-card size="small" title="资源" :bordered="true" v-if="Array.isArray(resources) && resources.length">
                <div class="res-list">
                  <div class="res-item" v-for="(res, idx) in resources" :key="idx">
                    <div class="res-name">{{ res.name || res.title || `资源${idx+1}` }}</div>
                    <div class="res-url">{{ res.url || '-' }}</div>
                  </div>
                </div>
              </a-card>

              
            </div>
          </div>
        </template>
      </div>
    </a-card>
  </AppLayout>
  
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
 
import AppLayout from '@/components/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { videoAPI } from '@/api'
import Plyr from 'plyr'
import Hls from 'hls.js'
import 'plyr/dist/plyr.css'
import { ThunderboltOutlined, PlayCircleOutlined, ClockCircleOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const error = ref('')
const detailData = ref<any>(null)
const fromCache = ref(false)
const downloading = ref(false)
const videoLoading = ref(false)
const networkSpeed = ref('')
const originalRate = ref(1) // 保存原始播放速率
const isLongPressActive = ref(false) // 长按状态

// 跳过片首片尾相关变量
const skipIntro = ref({ enabled: false, seconds: 30 })
const skipOutro = ref({ enabled: false, seconds: 30 })

// 网速计算相关变量
let lastLoadedBytes = 0
let lastSpeedCheckTime = 0
let speedCheckInterval: any = null
let longPressTimer: any = null // 长按定时器

const sourceId = computed(() => String(route.params.sourceId || ''))
const videoUrl = computed(() => String(route.query.original_url || route.query.url || ''))
const currentPlayUrl = ref<string>('')
const originalUrl = computed(() => String(route.query.original_url || route.query.url || videoUrl.value || ''))
const displayTitle = computed(() => {
  const ep = flatEpisodes.value.find(e => e.url === currentPlayUrl.value)
  if (ep) return `${base.value.name || ''} - ${ep.name}`.trim()
  return String(route.query.title || base.value.name || '')
})

const cacheKey = computed(() => `watch_detail:${sourceId.value}:${encodeURIComponent(videoUrl.value)}`)
// 使用 sourceId + original_url 作为进度键，避免标题变化导致无法命中
const playStateKey = computed(() => {
  const keyUrl = String(route.query.original_url || videoUrl.value || '')
  return `watch_state:${sourceId.value}:${encodeURIComponent(keyUrl)}`
})

 

// 播放器相关
const videoRef = ref<HTMLVideoElement | null>(null)
const playerSource = ref('')
let plyr: any = null
let hls: any = null
let isDraggingProgress = false
let plyrLongPressTimerRef: any = null
let wakeLock: any = null

// Screen Wake Lock API
async function requestWakeLock(): Promise<void> {
  try {
    const anyNav: any = navigator as any
    if (!anyNav.wakeLock || typeof anyNav.wakeLock.request !== 'function') return
    if (wakeLock && !wakeLock.released) return
    wakeLock = await anyNav.wakeLock.request('screen')
    try { console.log('[WakeLock] acquired') } catch {}
    wakeLock.addEventListener('release', () => {
      try { console.log('[WakeLock] released') } catch {}
    })
  } catch (e) {
    try { console.log('[WakeLock] request failed:', e) } catch {}
  }
}

async function releaseWakeLock(): Promise<void> {
  try {
    if (wakeLock && typeof wakeLock.release === 'function' && !wakeLock.released) {
      await wakeLock.release()
    }
  } catch {}
  wakeLock = null
}
const basePoster = computed(() => String((detailData.value?.cover || detailData.value?.poster || ''))) 

// 播放方案显示
const playerScheme = computed(() => {
  if (!playerSource.value) return '未加载'
  
  const url = playerSource.value.toLowerCase()
  if (url.includes('.m3u8')) {
    return 'HLS 流媒体'
  } else if (url.includes('.mp4')) {
    return 'MP4 直链'
  } else if (url.includes('.flv')) {
    return 'FLV 流媒体'
  } else if (url.includes('.webm')) {
    return 'WebM 格式'
  } else if (url.includes('rtmp://')) {
    return 'RTMP 流媒体'
  } else if (url.includes('http')) {
    return 'HTTP 直链'
  } else {
    return '未知格式'
  }
})

// 当前播放器方案显示
const currentPlayerScheme = computed(() => {
  if (!playerSource.value) return '未加载'
  
  const url = playerSource.value.toLowerCase()
  
  // 检查是否使用HLS
  if (url.includes('.m3u8')) {
    if (Hls.isSupported()) {
      return 'Plyr + HLS.js'
    } else if (videoRef.value?.canPlayType('application/vnd.apple.mpegurl')) {
      return 'Plyr + 原生HLS'
    } else {
      return 'Plyr + 兜底方案'
    }
  }
  
  // 其他格式都使用Plyr
  return 'Plyr 播放器'
})
const rates = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 2.5, 3]
const rate = ref(1)

// 播放速率选项
const rateOptions = computed(() => {
  return rates.map(r => ({
    label: `${r}x`,
    value: r
  }))
})

// 处理播放速率变化
function handleRateChange(value: number) {
  rate.value = value
  try {
    if (plyr) {
      plyr.speed = value
      savePlayState({ rate: value })
    } else if (videoRef.value) {
      // 原生 video 兜底
      ;(videoRef.value as any).playbackRate = value as any
      savePlayState({ rate: value })
    }
  } catch {}
}

// 处理跳过片首变化
function handleSkipIntroChange() {
  savePlayState({ 
    skipIntro: { 
      enabled: skipIntro.value.enabled, 
      seconds: skipIntro.value.seconds 
    } 
  })
}

// 处理跳过片尾变化
function handleSkipOutroChange() {
  savePlayState({ 
    skipOutro: { 
      enabled: skipOutro.value.enabled, 
      seconds: skipOutro.value.seconds 
    } 
  })
}

// 检测是否为移动设备
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// （已上方定义）

// 初始化 Plyr 实例
function ensurePlyr() {
  if (plyr || !videoRef.value) return
  const controls = isMobile.value
    ? ['play', 'progress', 'current-time', 'duration', 'mute', 'settings', 'fullscreen']
    : ['play', 'progress', 'current-time', 'duration', 'mute', 'volume', 'settings', 'fullscreen']
  plyr = new Plyr(videoRef.value!, {
    controls,
    settings: ['speed'],
    speed: { selected: rate.value, options: rates },
    clickToPlay: true,
  })
  
  // 立即禁用 Plyr 的双击全屏功能
  disablePlyrDoubleClick()
  
  // 绑定 Plyr 的播放完成事件
  plyr.on('ended', () => {
    try {
      console.log('[Plyr] 视频播放完成，尝试切换下一集')
      if (canNext.value) {
        playNext()
      } else {
        console.log('[Plyr] 没有下一集，播放结束')
      }
    } catch (e) {
      console.error('[Plyr] 自动切换下一集失败:', e)
    }
  })
  
  // Plyr 视频等待数据事件（卡住检测）
  plyr.on('waiting', () => {
    console.log('[Plyr] 视频等待数据，显示loading')
    setVideoLoading(true)
  })
  
  // Plyr 视频可以播放事件（恢复检测）
  plyr.on('canplay', () => {
    console.log('[Plyr] 视频可以播放，隐藏loading')
    setVideoLoading(false)
  })
  
  // Plyr 视频可以流畅播放事件
  plyr.on('canplaythrough', () => {
    console.log('[Plyr] 视频可以流畅播放，隐藏loading')
    setVideoLoading(false)
  })
  
  // Plyr 播放状态 -> Screen Wake Lock
  plyr.on('play', async () => { await requestWakeLock() })
  plyr.on('pause', async () => { await releaseWakeLock() })
  plyr.on('ended', async () => { await releaseWakeLock() })
  
  bindPlayerEvents()

  // 手势：左右滑动调节进度（Plyr 容器）
  try {
    const container = plyr?.elements?.container as HTMLElement
    if (container) attachProgressDrag(container)
  } catch {}
}

// 禁用 Plyr 双击全屏的专用函数
function disablePlyrDoubleClick() {
  if (!plyr) return
  
  // 方法1: 通过CSS禁用双击选择
  const style = document.createElement('style')
  style.textContent = `
    .plyr__video-wrapper {
      -webkit-user-select: none !important;
      -moz-user-select: none !important;
      -ms-user-select: none !important;
      user-select: none !important;
    }
    .plyr__video-wrapper * {
      -webkit-user-select: none !important;
      -moz-user-select: none !important;
      -ms-user-select: none !important;
      user-select: none !important;
    }
  `
  document.head.appendChild(style)
  
  // 方法2: 直接移除 Plyr 的双击事件监听器
  try {
    const container = plyr.elements.container
    const video = plyr.elements.video
    
    // 克隆元素来移除所有事件监听器
    const newContainer = container.cloneNode(true)
    const newVideo = video.cloneNode(true)
    
    container.parentNode?.replaceChild(newContainer, container)
    newContainer.appendChild(newVideo)
    
    // 重新设置 Plyr 的元素引用
    plyr.elements.container = newContainer
    plyr.elements.video = newVideo
  } catch (e) {
    console.warn('无法移除Plyr事件监听器:', e)
  }
  
  // 方法3: 使用事件捕获阶段阻止双击
  const preventDoubleClick = (e: Event) => {
    e.preventDefault()
    e.stopPropagation()
    e.stopImmediatePropagation()
    return false
  }
  
  // 在捕获阶段阻止双击事件
  plyr.elements.container.addEventListener('dblclick', preventDoubleClick, true)
  plyr.elements.video.addEventListener('dblclick', preventDoubleClick, true)
  
  // 方法4: 覆盖 Plyr 的内部双击处理函数
  if (plyr.config && typeof plyr.config === 'object') {
    (plyr.config as any).doubleClick = false
  }
}

// 为 Plyr 添加自定义事件处理
function addPlyrCustomEvents() {
  if (!plyr) return
  
  // 双击播放/暂停功能
  let plyrLastClickTime = 0
  const plyrDoubleClickThreshold = 300
  
  plyr.elements.container.addEventListener('click', (e: any) => {
    const currentTime = Date.now()
    if (currentTime - plyrLastClickTime < plyrDoubleClickThreshold) {
      // 双击事件
      plyrLastClickTime = 0
      try {
        if (plyr.playing) {
          plyr.pause()
          console.log('[Plyr] 双击暂停')
        } else {
          plyr.play()
          console.log('[Plyr] 双击播放')
        }
      } catch {}
    } else {
      // 单击事件
      plyrLastClickTime = currentTime
    }
  })
  
  // 长按2倍速播放功能
  let plyrLongPressTimer: any = null
  let plyrIsTouchActive = false
  
  // 触摸开始
  plyr.elements.container.addEventListener('touchstart', (e: any) => {
    if (isDraggingProgress) return
    plyrIsTouchActive = true
    if (plyrLongPressTimer) clearTimeout(plyrLongPressTimer)
    plyrLongPressTimer = setTimeout(() => {
      if (plyrIsTouchActive && !isDraggingProgress) {
        originalRate.value = plyr.speed
        plyr.speed = 2
        isLongPressActive.value = true
        console.log('[Plyr LongPress] 启动2倍速播放')
      }
    }, 500)
  })
  
  // 触摸结束
  plyr.elements.container.addEventListener('touchend', (e: any) => {
    plyrIsTouchActive = false
    if (plyrLongPressTimer) {
      clearTimeout(plyrLongPressTimer)
      plyrLongPressTimer = null
    }
    if (isLongPressActive.value) {
      plyr.speed = originalRate.value
      isLongPressActive.value = false
      console.log('[Plyr LongPress] 恢复原始播放速率:', originalRate.value)
    }
  })
  
  // 触摸取消
  plyr.elements.container.addEventListener('touchcancel', (e: any) => {
    plyrIsTouchActive = false
    if (plyrLongPressTimer) {
      clearTimeout(plyrLongPressTimer)
      plyrLongPressTimer = null
    }
    if (isLongPressActive.value) {
      plyr.speed = originalRate.value
      isLongPressActive.value = false
    }
  })
  
  // 鼠标按下（桌面端）
  plyr.elements.container.addEventListener('mousedown', (e: any) => {
    if (e.button === 0) {
      if (isDraggingProgress) return
      plyrIsTouchActive = true
      if (plyrLongPressTimer) clearTimeout(plyrLongPressTimer)
      plyrLongPressTimer = setTimeout(() => {
        if (plyrIsTouchActive && !isDraggingProgress) {
          originalRate.value = plyr.speed
          plyr.speed = 2
          isLongPressActive.value = true
          console.log('[Plyr LongPress] 启动2倍速播放')
        }
      }, 500)
    }
  })
  
  // 鼠标松开（桌面端）
  plyr.elements.container.addEventListener('mouseup', (e: any) => {
    if (e.button === 0) {
      plyrIsTouchActive = false
      if (plyrLongPressTimer) {
        clearTimeout(plyrLongPressTimer)
        plyrLongPressTimer = null
      }
      if (isLongPressActive.value) {
        plyr.speed = originalRate.value
        isLongPressActive.value = false
        console.log('[Plyr LongPress] 恢复原始播放速率:', originalRate.value)
      }
    }
  })
  
  // 鼠标离开（桌面端）
  plyr.elements.container.addEventListener('mouseleave', (e: any) => {
    if (plyrIsTouchActive) {
      plyrIsTouchActive = false
      if (plyrLongPressTimer) {
        clearTimeout(plyrLongPressTimer)
        plyrLongPressTimer = null
      }
      if (isLongPressActive.value) {
        plyr.speed = originalRate.value
        isLongPressActive.value = false
      }
    }
  })
}
let lastSavedSecond = 0
let playerBound = false
let lastVideoW = 0
let lastVideoH = 0
let orientationLocked = false
// 获取当前应播放的剧集 URL：优先 currentPlayUrl，其次当前来源的第一集，再次所有剧集第一集，最后回退 original_url
function getSelectedEpisodeUrl(): string {
  let url = String(currentPlayUrl.value || '')
  if (!url) {
    const idx = Number(activeSourceKey.value || 0)
    const src = sourcesByTab.value[idx]
    if (src && Array.isArray(src.episodes) && src.episodes.length > 0) {
      url = String(src.episodes[0]?.url || '')
    }
  }
  if (!url && flatEpisodes.value.length > 0) {
    url = String(flatEpisodes.value[0]?.url || '')
  }
  if (!url) {
    url = String(videoUrl.value || '')
  }
  try { console.log('[Play] getSelectedEpisodeUrl =>', url) } catch {}
  return url
}


watch(rate, (v) => {
  try { if (plyr) plyr.speed = v } catch {}
})

function bindPlayerEvents() {
  if (!videoRef.value || playerBound) return
  playerBound = true
  // 进度保存
  const v = videoRef.value!
  v.addEventListener('timeupdate', () => {
    try {
      const ct = Math.floor(v.currentTime || 0)
      const dur = Math.floor(v.duration || 0)
      if (dur > 0 && Math.abs(ct - lastSavedSecond) >= 5) {
        lastSavedSecond = ct
        savePlayState({ currentTime: ct })
      }
      
      // 检查是否需要跳过片首
      if (skipIntro.value.enabled && ct < skipIntro.value.seconds) {
        v.currentTime = skipIntro.value.seconds
        console.log(`跳过片首，跳转到 ${skipIntro.value.seconds} 秒`)
      }
      
      // 检查是否需要跳过片尾
      if (skipOutro.value.enabled && dur > 0 && ct > dur - skipOutro.value.seconds) {
        // 如果接近片尾，自动切换到下一集
        if (canNext.value) {
          playNext()
        } else {
          // 没有下一集，跳转到片尾前指定秒数
          v.currentTime = Math.max(0, dur - skipOutro.value.seconds)
          console.log(`跳过片尾，跳转到 ${Math.max(0, dur - skipOutro.value.seconds)} 秒`)
        }
      }
    } catch {}
  })
  
  // 播放完成自动切换下一集（原生事件）
  v.addEventListener('ended', () => {
    try {
      console.log('[Video] 视频播放完成，尝试切换下一集')
      if (canNext.value) {
        playNext()
      } else {
        console.log('[Video] 没有下一集，播放结束')
      }
    } catch (e) {
      console.error('[Video] 自动切换下一集失败:', e)
    }
  })
  
  // 视频等待数据事件（卡住检测）
  v.addEventListener('waiting', () => {
    console.log('[Video] 视频等待数据，显示loading')
    setVideoLoading(true)
  })
  
  // 视频可以播放事件（恢复检测）
  v.addEventListener('canplay', () => {
    console.log('[Video] 视频可以播放，隐藏loading')
    setVideoLoading(false)
  })
  
  // 视频可以流畅播放事件
  v.addEventListener('canplaythrough', () => {
    console.log('[Video] 视频可以流畅播放，隐藏loading')
    setVideoLoading(false)
  })
  
  // 双击播放/暂停功能（原生 video 容器点击处理与 Plyr 一致）
  let lastClickTime = 0
  let clickCount = 0
  const doubleClickThreshold = 300 // 双击时间阈值（毫秒）
  
  v.addEventListener('click', (e) => {
    const currentTime = Date.now()
    if (currentTime - lastClickTime < doubleClickThreshold) {
      // 双击事件
      clickCount = 0
      lastClickTime = 0
      try {
        if (!v.paused) {
          v.pause()
          console.log('[Video] 双击暂停')
        } else {
          v.play()
          console.log('[Video] 双击播放')
        }
      } catch {}
      
      // 阻止默认行为
      e.preventDefault()
      e.stopPropagation()
      e.stopImmediatePropagation()
    } else {
      // 单击事件
      clickCount = 1
      lastClickTime = currentTime
    }
  })
  
  // 禁用默认的双击全屏行为
  v.addEventListener('dblclick', (e) => {
    e.preventDefault()
    e.stopPropagation()
    e.stopImmediatePropagation()
    return false
  })
  
  // 在视频元素上添加 CSS 样式禁用双击选择
  v.style.userSelect = 'none'
  ;(v.style as any).webkitUserSelect = 'none'
  ;(v.style as any).mozUserSelect = 'none'
  ;(v.style as any).msUserSelect = 'none'
  
  // 长按2倍速播放事件监听
  let touchStartTime = 0
  let isTouchActive = false
  let longPressTimer: any = null
  
  // 触摸开始
  v.addEventListener('touchstart', (e) => {
    if (isDraggingProgress) return
    e.preventDefault() // 阻止默认行为
    touchStartTime = Date.now()
    isTouchActive = true
    if (longPressTimer) clearTimeout(longPressTimer)
    longPressTimer = setTimeout(() => {
      if (isTouchActive && !isDraggingProgress) {
        startLongPress()
      }
    }, 500)
  }, { passive: false })
  
  // 触摸结束
  v.addEventListener('touchend', (e) => {
    e.preventDefault() // 阻止默认行为
    isTouchActive = false
    if (longPressTimer) {
      clearTimeout(longPressTimer)
      longPressTimer = null
    }
    endLongPress()
  }, { passive: false })
  
  // 触摸取消
  v.addEventListener('touchcancel', (e) => {
    e.preventDefault() // 阻止默认行为
    isTouchActive = false
    if (longPressTimer) {
      clearTimeout(longPressTimer)
      longPressTimer = null
    }
    endLongPress()
  }, { passive: false })
  
  // 鼠标按下（桌面端）
  v.addEventListener('mousedown', (e) => {
    if (e.button === 0) { // 左键
      if (isDraggingProgress) return
      e.preventDefault() // 阻止默认行为
      touchStartTime = Date.now()
      isTouchActive = true
      if (longPressTimer) clearTimeout(longPressTimer)
      longPressTimer = setTimeout(() => {
        if (isTouchActive && !isDraggingProgress) {
          startLongPress()
        }
      }, 500)
    }
  }, { passive: false })
  
  // 鼠标松开（桌面端）
  v.addEventListener('mouseup', (e) => {
    if (e.button === 0) { // 左键
      e.preventDefault() // 阻止默认行为
      isTouchActive = false
      if (longPressTimer) {
        clearTimeout(longPressTimer)
        longPressTimer = null
      }
      endLongPress()
    }
  }, { passive: false })
  
  // 鼠标离开（桌面端）
  v.addEventListener('mouseleave', (e) => {
    if (isTouchActive) {
      isTouchActive = false
      if (longPressTimer) {
        clearTimeout(longPressTimer)
        longPressTimer = null
      }
      endLongPress()
    }
  })
  
  // 倍速变更（通过 plyr 统一）
  // 元数据
  v.addEventListener('loadedmetadata', () => {
    try {
      if (v.videoWidth && v.videoHeight) { lastVideoW = v.videoWidth; lastVideoH = v.videoHeight }
    } catch {}
  })
  
  // 为 Plyr 添加双击快进快退和长按功能
  if (plyr) {
    addPlyrCustomEvents()
  }
  // 手势：左右滑动调节进度（原生容器）
  try {
    const container = v.parentElement
    if (container) attachProgressDrag(container)
  } catch {}
  
  // 启动网速监控
  if (!speedCheckInterval) {
    startSpeedMonitoring()
  }
  
  // 页面可见性变化时恢复/释放 Wake Lock
  try {
    document.addEventListener('visibilitychange', async () => {
      if (document.visibilityState === 'visible') {
        if (!videoRef.value?.paused) { await requestWakeLock() }
      } else {
        await releaseWakeLock()
      }
    })
  } catch {}

  // 容器与 document 级别全屏事件
  try {
    const container = v.parentElement
    const handleFs = () => {
      try { console.log('[Fullscreen] document/container fullscreenchange') } catch {}
      const d: any = document as any
      const fsEl = d.fullscreenElement || d.webkitFullscreenElement || d.mozFullScreenElement || d.msFullscreenElement
      if (fsEl) {
        void handleEnterFullscreen()
      } else {
        void handleExitFullscreen()
      }
    }
    if (container) {
      container.addEventListener('fullscreenchange', handleFs)
      container.addEventListener('webkitfullscreenchange', handleFs as any)
    }
    document.addEventListener('fullscreenchange', handleFs)
    document.addEventListener('webkitfullscreenchange', handleFs as any)
    document.addEventListener('mozfullscreenchange', handleFs as any)
    document.addEventListener('MSFullscreenChange', handleFs as any)
  } catch {}
  // iOS webkit 原生事件
  try {
    const handleElFs = (_e: Event) => { void handleEnterFullscreen() }
    const handleElExit = (_e: Event) => { void handleExitFullscreen() }
    v.addEventListener('fullscreenchange', handleElFs)
    v.addEventListener('webkitfullscreenchange', handleElFs as any)
    v.addEventListener('mozfullscreenchange', handleElFs as any)
    v.addEventListener('MSFullscreenChange', handleElFs as any)
    v.addEventListener('webkitbeginfullscreen', () => { console.log('[Fullscreen] webkitbeginfullscreen'); handleEnterFullscreen() })
    v.addEventListener('webkitendfullscreen', () => { console.log('[Fullscreen] webkitendfullscreen'); handleExitFullscreen() })
  } catch {}
}

async function handleEnterFullscreen() {
  console.log('handleEnterFullscreen')
  // 每次进入全屏时都重新获取视频尺寸
  const v = videoRef.value
  if (v) {
    try {
      // 重新获取视频尺寸
      const vw = v.videoWidth || 0
      const vh = v.videoHeight || 0
      if (vw && vh) {
        lastVideoW = vw
        lastVideoH = vh
        try { console.log(`[Fullscreen] 重新获取视频尺寸: ${vw}x${vh}`) } catch {}
      }
    } catch (e: any) {
      try { console.log('[Fullscreen] 获取视频尺寸失败:', e) } catch {}
    }
  }
  
  const orientation = estimateOrientation()
  try { console.log(`[Fullscreen] 根据视频尺寸 ${lastVideoW}x${lastVideoH} 设置屏幕方向: ${orientation}`) } catch {}
  await lockOrientation(orientation)
}

async function handleExitFullscreen() {
  try { console.log('[Fullscreen] 退出全屏，尝试解锁屏幕方向') } catch {}
  await unlockOrientation()
}

function estimateOrientation(): 'landscape' | 'portrait' {
  // 优先使用真实视频宽高
  const w = lastVideoW || (videoRef.value?.videoWidth || 0)
  const h = lastVideoH || (videoRef.value?.videoHeight || 0)
  if (w > 0 && h > 0) return w >= h ? 'landscape' : 'portrait'
  // 退化为容器尺寸
  const el: any = (videoRef.value && videoRef.value.parentElement) || videoRef.value
  const cw = el?.clientWidth || window.innerWidth
  const ch = el?.clientHeight || window.innerHeight
  return cw >= ch ? 'landscape' : 'portrait'
}

async function lockOrientation(ori: 'landscape' | 'portrait') {
  try {
    const o: any = (screen as any).orientation
    if (o && typeof o.lock === 'function') {
      try { console.log('[Fullscreen] 请求锁定方向:', ori) } catch {}
      await o.lock(ori === 'landscape' ? 'landscape' : 'portrait-primary')
      orientationLocked = true
      try { console.log('[Fullscreen] 锁定方向成功') } catch {}
    }
  } catch {
    // 忽略不支持或被拒绝
  }
}

async function unlockOrientation() {
  try {
    const o: any = (screen as any).orientation
    if (orientationLocked && o && typeof o.unlock === 'function') {
      o.unlock()
      try { console.log('[Fullscreen] 已解锁方向') } catch {}
    }
  } catch {}
  orientationLocked = false
}

// Screen Wake Lock API（声明到顶部作用域）
// （重复定义已移除）



// 资源列表容错（上移，供后续 sourcesByTab 使用，避免初始化顺序问题）
const resources = computed<any[]>(() => {
  const d: any = detailData.value || {}
  if (Array.isArray(d.resources)) return d.resources
  if (Array.isArray(d.playlist)) return d.playlist
  if (Array.isArray(d.urls)) return d.urls
  if (Array.isArray(d.videos)) return d.videos
  return []
})

// 站点与剧集（source -> episodes）
const activeSourceKey = ref('0')
const sourcesByTab = computed(() => {
  const d: any = detailData.value || {}
  let raw: any = d.sources || d.source || d.playSources || d.play_sources || d.playSource || d.play_source
  const list: Array<{ name: string; episodes: Array<{ name: string; url: string }> }> = []

  const toEp = (x: any, i: number) => ({
    name: String(x?.name || x?.title || x?.text || `第${i + 1}集`),
    url: String(x?.url || x?.link || ''),
  })

  if (Array.isArray(raw)) {
    raw.forEach((s: any, idx: number) => {
      const name = String(s?.name || s?.title || `来源${idx + 1}`)
      let eps: any = s?.list || s?.episodes || s?.urls || s?.videos || s?.items || []
      if (Array.isArray(eps)) {
        const episodes = eps.map((e: any, i: number) => toEp(e, i)).filter((e: any) => e.url)
        if (episodes.length) list.push({ name, episodes })
      }
    })
  } else if (raw && typeof raw === 'object') {
    Object.keys(raw).forEach((key, idx) => {
      const arr = (raw as any)[key]
      if (Array.isArray(arr)) {
        const episodes = arr.map((e: any, i: number) => toEp(e, i)).filter((e: any) => e.url)
        if (episodes.length) list.push({ name: key, episodes })
      }
    })
  }

  // 兜底：用 resources 构造一个默认来源
  if (!list.length && resources.value.length) {
    const episodes = resources.value
      .map((r: any, i: number) => ({ name: String(r?.name || r?.title || `资源${i + 1}`), url: String(r?.url || '') }))
      .filter(e => e.url)
    if (episodes.length) list.push({ name: String(d.source || d.source_name || '默认'), episodes })
  }
  return list
})

const flatEpisodes = computed(() => sourcesByTab.value.flatMap(s => s.episodes.map(ep => ({ ...ep, __sourceName: s.name }))))
// 同步 HTML 标题与页面标题一致（放在 flatEpisodes 之后，避免初始化顺序问题）
watch(displayTitle, (t) => {
  const baseTitle = 'Video Crawler'
  const full = t ? `播放 - ${t}` : '播放'
  document.title = [full, baseTitle].filter(Boolean).join(' | ')
}, { immediate: true })
// 当前来源（优先使用选中的 tab；若不含当前剧集，则回退到包含当前剧集的来源）
const currentSource = computed(() => {
  const list = sourcesByTab.value
  if (!list.length) return null as any
  const activeIdx = Number(activeSourceKey.value || 0)
  const active = list[activeIdx]
  if (active && active.episodes?.some(e => e.url === currentPlayUrl.value)) return active
  const found = list.find(s => s.episodes?.some(e => e.url === currentPlayUrl.value))
  return found || active
})
const currentSourceEpisodes = computed(() => currentSource.value ? currentSource.value.episodes : [])
const currentIndex = computed(() => currentSourceEpisodes.value.findIndex((e: any) => e.url === currentPlayUrl.value))
const canPrev = computed(() => currentIndex.value > 0)
// 仅判断“当前来源”是否还有下一集
const canNext = computed(() => currentIndex.value >= 0 && currentIndex.value < currentSourceEpisodes.value.length - 1)

function isCurrentEpisode(ep: { url: string }) {
  return String(ep?.url || '') === currentPlayUrl.value
}

async function playEpisode(ep: { name: string; url: string }, sourceName?: string) {
  if (!ep?.url) return
  // 立刻更新当前剧集，用于后续解析播放链接与按钮状态
  currentPlayUrl.value = ep.url
  // 仅更新地址栏中标题与来源，不修改 url 参数，避免影响回显
  const q = { ...route.query, title: ep.name, source: sourceName || (ep as any).__sourceName }
  router.replace({ name: 'watch', params: route.params, query: q })
  try {
    setVideoLoading(true) // 开始加载
    const token = auth.token!
    const res: any = await videoAPI.playUrl(token, sourceId.value, ep.url)
    const url: string = res?.data?.video_url || res?.data || ''
    if (!url) return
    playerSource.value = url
    await nextTick()
    ensurePlyr()
    if (videoRef.value) {
      // 切换播放源
      if (Hls.isSupported()) {
        if (hls) { try { hls.destroy() } catch {} }
        hls = new Hls({ maxBufferLength: 30 })
        hls.loadSource(url)
        hls.attachMedia(videoRef.value)
      } else {
        videoRef.value.src = url
      }
      // 设置倍速
      try { if (plyr) plyr.speed = rate.value } catch {}
      // 恢复该剧集的缓存进度
      const state = loadPlayState()
      const seekTo = (state && String(state.url) === ep.url && state.currentTime && state.currentTime > 0) ? state.currentTime : 0
      if (seekTo > 0) {
        const doSeek = () => { try { if (videoRef.value) videoRef.value.currentTime = seekTo } catch {} }
        if ((videoRef.value?.readyState || 0) >= 1) doSeek()
        else videoRef.value?.addEventListener('loadedmetadata', doSeek, { once: true })
      }
      // 自动播放
      try { await videoRef.value.play() } catch {}
      bindPlayerEvents()
    }
    // 保存所选剧集
    savePlayState({ url: ep.url, title: ep.name, source: q.source })
  } catch (error) {
    console.error('播放剧集失败:', error)
  } finally {
    setVideoLoading(false) // 结束加载
  }
}

function playPrev() {
  if (!canPrev.value) return
  playEpisode(currentSourceEpisodes.value[currentIndex.value - 1])
}
function playNext() {
  if (!canNext.value) return
  playEpisode(currentSourceEpisodes.value[currentIndex.value + 1])
}

// 迅雷下载功能
function downloadWithThunder() {
  if (!playerSource.value) {
    return
  }
  
  downloading.value = true
  
  try {
    // 构建文件名：视频名称 + 剧集名称
    const videoName = base.value.name || '未知视频'
    const currentEpisode = flatEpisodes.value.find(e => e.url === currentPlayUrl.value)
    const episodeName = currentEpisode?.name || ''
    const fileName = episodeName ? `${videoName} - ${episodeName}` : videoName
    
    // 清理文件名中的非法字符
    const cleanFileName = fileName.replace(/[<>:"/\\|?*]/g, '_').trim()
    
    // 构建迅雷下载链接
    const thunderUrl = `thunder://${btoa(`AA${playerSource.value}ZZ`)}`
    
    // 创建下载链接并触发下载
    const link = document.createElement('a')
    link.href = thunderUrl
    link.download = `${cleanFileName}.mp4` // 设置下载文件名
    link.style.display = 'none'
    
    // 添加到页面并触发点击
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // 显示成功提示
    message.success('已调用迅雷下载，请检查迅雷是否已启动')
    
  } catch (error) {
    console.error('迅雷下载失败:', error)
    message.error('迅雷下载失败，请检查迅雷是否已安装')
  } finally {
    downloading.value = false
  }
}

// 网速计算和加载状态管理
function startSpeedMonitoring() {
  if (speedCheckInterval) {
    clearInterval(speedCheckInterval)
  }
  
  lastLoadedBytes = 0
  lastSpeedCheckTime = Date.now()
  networkSpeed.value = ''
  
  speedCheckInterval = setInterval(() => {
    if (videoRef.value) {
      const v = videoRef.value
      const currentTime = Date.now()
      const timeDiff = (currentTime - lastSpeedCheckTime) / 1000 // 秒
      
      if (timeDiff > 0 && v.buffered.length > 0) {
        const bufferedEnd = v.buffered.end(v.buffered.length - 1)
        const currentVideoTime = v.currentTime
        const bufferedTime = bufferedEnd - currentVideoTime
        
        // 更准确的字节数估算：基于视频时长和码率
        // 假设平均码率为 2Mbps (250KB/s)
        const estimatedBitrate = 2 * 1024 * 1024 // 2Mbps
        const bufferedBytes = bufferedTime * estimatedBitrate / 8
        
        const bytesDiff = bufferedBytes - lastLoadedBytes
        const speedBps = bytesDiff / timeDiff
        
        if (speedBps > 0) {
          const speedKBps = speedBps / 1024
          const speedMBps = speedKBps / 1024
          
          if (speedMBps >= 1) {
            networkSpeed.value = `${speedMBps.toFixed(1)} MB/s`
          } else {
            networkSpeed.value = `${speedKBps.toFixed(1)} KB/s`
          }
        } else {
          networkSpeed.value = '0 KB/s'
        }
        
        lastLoadedBytes = bufferedBytes
        lastSpeedCheckTime = currentTime
        
        // 检测播放中卡住的情况
        checkVideoStuck()
      }
    }
  }, 1000) // 每秒检查一次
}

// 检测视频是否卡住
function checkVideoStuck() {
  if (!videoRef.value) return
  
  const v = videoRef.value
  
  // 如果视频正在播放但缓冲不足，显示loading
  if (!v.paused && v.readyState < 3) {
    // readyState < 3 表示缓冲不足
    if (!videoLoading.value) {
      console.log('[Video] 检测到播放卡住，显示loading')
      setVideoLoading(true)
    }
  } else if (videoLoading.value && v.readyState >= 3) {
    // readyState >= 3 表示缓冲充足
    console.log('[Video] 缓冲充足，隐藏loading')
    setVideoLoading(false)
  }
}

function stopSpeedMonitoring() {
  if (speedCheckInterval) {
    clearInterval(speedCheckInterval)
    speedCheckInterval = null
  }
  networkSpeed.value = ''
}

function setVideoLoading(loading: boolean) {
  videoLoading.value = loading
  if (loading && !speedCheckInterval) {
    // 开始加载时启动网速监控
    startSpeedMonitoring()
  }
  // 移除自动停止网速监控的逻辑，让网速监控持续运行
  // 只有在组件卸载时才停止
}

// 长按2倍速播放功能
function startLongPress() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
  }
  
  longPressTimer = setTimeout(() => {
    if (isDraggingProgress) return
    // 保存当前播放速率
    originalRate.value = rate.value
    // 设置为2倍速
    rate.value = 2
    isLongPressActive.value = true
    
    // 应用2倍速
    try {
      if (plyr) {
        plyr.speed = 2
      } else if (videoRef.value) {
        (videoRef.value as any).playbackRate = 2
      }
      // 长按期间隐藏进度条（Plyr 与 原生 video 均支持）
      try {
        const container = plyr
          ? plyr.elements.container
          : (videoRef.value ? videoRef.value.parentElement : null)
        if (container) container.classList.add('longpress-hide-progress')
      } catch {}
      console.log('[LongPress] 启动2倍速播放')
    } catch (e) {
      console.error('[LongPress] 设置2倍速失败:', e)
    }
  }, 500) // 500ms 长按触发
}

function endLongPress() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
  
  if (isLongPressActive.value) {
    // 恢复原始播放速率
    rate.value = originalRate.value
    isLongPressActive.value = false
    
    // 应用原始速率
    try {
      if (plyr) {
        plyr.speed = originalRate.value
      } else if (videoRef.value) {
        (videoRef.value as any).playbackRate = originalRate.value
      }
      // 恢复进度条显示（Plyr 与 原生 video 均支持）
      try {
        const container = plyr
          ? plyr.elements.container
          : (videoRef.value ? videoRef.value.parentElement : null)
        if (container) container.classList.remove('longpress-hide-progress')
      } catch {}
      console.log('[LongPress] 恢复原始播放速率:', originalRate.value)
    } catch (e) {
      console.error('[LongPress] 恢复原始速率失败:', e)
    }
  }
}

// 映射基础字段，容错不同脚本返回的键名
const base = computed(() => {
  const d: any = detailData.value || {}
  return {
    name: d.name || d.title || '',
    director: d.director || '',
    actor: d.actor || d.actors || '',
    language: d.language || d.lang || '',
    region: d.region || d.area || '',
    release_date: d.release_date || d.releaseDate || '',
    rate: d.rate || d.rating || '',
    type: d.type || d.category || '',
    description: d.description || d.desc || '',
  }
})

// resources 已上移

function saveCache() {
  try {
    sessionStorage.setItem(cacheKey.value, JSON.stringify({
      t: Date.now(),
      data: detailData.value,
    }))
  } catch {}
}

function loadCache(): boolean {
  try {
    const raw = sessionStorage.getItem(cacheKey.value)
    if (!raw) return false
    const obj = JSON.parse(raw)
    if (obj && 'data' in obj) {
      detailData.value = obj.data
      return true
    }
  } catch {}
  return false
}

type PlayState = { 
  url?: string; 
  title?: string; 
  source?: string; 
  currentTime?: number; 
  rate?: number; 
  skipIntro?: { enabled: boolean; seconds: number };
  skipOutro?: { enabled: boolean; seconds: number };
  updatedAt?: number 
}
function savePlayState(partial: PlayState) {
  try {
    const raw = localStorage.getItem(playStateKey.value)
    const prev: PlayState = raw ? JSON.parse(raw) : {}
    const merged: PlayState = { ...prev, ...partial, rate: rate.value, updatedAt: Date.now() }
    localStorage.setItem(playStateKey.value, JSON.stringify(merged))
    try { console.log('[PlayState] 已保存播放状态:', merged) } catch {}
  } catch {}
}
function loadPlayState(): PlayState | null {
  try {
    const raw = localStorage.getItem(playStateKey.value)
    if (!raw) {
      try { console.log('[PlayState] 未发现缓存: key =', playStateKey.value) } catch {}
      return null
    }
    const parsed = JSON.parse(raw)
    try { console.log('[PlayState] 读取到缓存:', parsed) } catch {}
    return parsed
  } catch { return null }
}

async function fetchDetail(force = false) {
  error.value = ''
  if (!force) {
    const hit = loadCache()
    fromCache.value = hit
    if (hit) {
      // 命中缓存也要解析播放地址
      await resolvePlayUrl()
      return
    }
  }

  if (!sourceId.value || !videoUrl.value) {
    error.value = '缺少必要参数'
    return
  }

  loading.value = true
  try {
    const token = auth.token!
    const res: any = await videoAPI.detail(token, sourceId.value, videoUrl.value)
    detailData.value = res?.data ?? res
    fromCache.value = false
    saveCache()
    await resolvePlayUrl() // detail 成功后拉取真实播放链接
  } catch (e: any) {
    error.value = e?.message || '获取详情失败'
  } finally {
    loading.value = false
  }
}

async function resolvePlayUrl() {
  try {
    setVideoLoading(true) // 开始加载
    const token = auth.token!
    // 优先请求"当前选中剧集"的播放链接；无则回退
    const episodeUrl = getSelectedEpisodeUrl()
    const res: any = await videoAPI.playUrl(token, sourceId.value, episodeUrl)
    const url: string = res?.data?.video_url || res?.data || ''
    playerSource.value = url
    await nextTick()
    ensurePlyr()
    if (videoRef.value) {
      try {
        // 使用 hls.js 播放 m3u8；Safari 原生支持时直接赋 src
        if (Hls.isSupported()) {
          try { console.log('[HLS] using hls.js, version:', (Hls as any).version) } catch {}
          if (hls) { try { hls.destroy() } catch {} }
          hls = new Hls({ maxBufferLength: 30 })
          try {
            hls.on(Hls.Events.MEDIA_ATTACHED, () => { try { console.log('[HLS] MEDIA_ATTACHED') } catch {} })
            hls.on(Hls.Events.MANIFEST_PARSED, (_: any, data: any) => { try { console.log('[HLS] MANIFEST_PARSED levels=', data?.levels?.length) } catch {} })
            hls.on(Hls.Events.ERROR, (_: any, data: any) => { try { console.log('[HLS] ERROR', data?.type, data?.details, 'fatal=', data?.fatal) } catch {} })
          } catch {}
          console.log('hls.loadSource', url)
          hls.loadSource(url)
          hls.attachMedia(videoRef.value)
        } else if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
          try { console.log('[HLS] using native HLS via canPlayType') } catch {}
          videoRef.value.src = url
        } else {
          // 兜底：直接设置
          try { console.log('[HLS] fallback: set src directly (no hls support)') } catch {}
          videoRef.value.src = url
        }
        if (plyr) try { plyr.speed = rate.value; console.log('[HLS] set speed to', rate.value) } catch {}

        // 每次播放时都检查缓存进度并跳转
        const state = loadPlayState()
        const seekTo = state?.currentTime || 0
        if (seekTo > 0) {
          const doSeek = () => { 
            try { 
              if (videoRef.value) videoRef.value.currentTime = seekTo
              console.log(`跳转到缓存进度: ${seekTo}秒`)
            } catch (e: any) {
              console.log('跳转进度失败:', e)
            } 
          }
          if ((videoRef.value?.readyState || 0) >= 1) {
            doSeek()
          } else {
            videoRef.value?.addEventListener('loadedmetadata', doSeek, { once: true })
          }
        } else {
          console.log('没有缓存进度，从头开始播放')
        }
        try { console.log('[HLS] call video.play()'); await videoRef.value.play() } catch (e: any) { try { console.log('[HLS] play() error', e) } catch {} }
        bindPlayerEvents()
      } catch {}
    }
    // 初始化情况下，将当前播放 url 与初始 url 对齐
    if (!currentPlayUrl.value) currentPlayUrl.value = videoUrl.value
  } catch (e: any) {
    // 忽略错误，保留空源
  } finally {
    setVideoLoading(false) // 结束加载
  }
}

function refreshDetail() {
  fetchDetail(true)
}

function goBack() {
  router.back()
}

function goOriginal() {
  if (originalUrl.value) {
    window.open(originalUrl.value, '_blank')
  }
}

 



onMounted(async () => {
  // 初始化移动设备检测
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  await fetchDetail(false)
  const state = loadPlayState()
  if (state?.url) {
    if (typeof state.rate === 'number') rate.value = state.rate
    // 恢复跳过片首片尾配置
    if (state.skipIntro) {
      skipIntro.value.enabled = state.skipIntro.enabled
      skipIntro.value.seconds = state.skipIntro.seconds
    }
    if (state.skipOutro) {
      skipOutro.value.enabled = state.skipOutro.enabled
      skipOutro.value.seconds = state.skipOutro.seconds
    }
    const ep = flatEpisodes.value.find(e => e.url === state.url) || flatEpisodes.value[0]
    if (ep) await playEpisode(ep)
  } else {
    const first = flatEpisodes.value[0]
    if (first) await playEpisode(first)
  }
  // 根据当前 url 选中对应的来源 tab
  currentPlayUrl.value = loadPlayState()?.url || currentPlayUrl.value || videoUrl.value
  const idx = sourcesByTab.value.findIndex(s => s.episodes.some(e => e.url === currentPlayUrl.value))
  if (idx >= 0) activeSourceKey.value = String(idx)
})

// 清理事件监听
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  stopSpeedMonitoring() // 清理网速监控定时器
  
  // 清理长按定时器
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
  // 释放 Wake Lock
  releaseWakeLock()
})

// 进度拖动手势（同时适配 Plyr 和原生 video 容器）
function attachProgressDrag(container: HTMLElement) {
  let startX = 0
  let startY = 0
  let startTime = 0
  let determined = false
  let isHorizontal = false
  let containerRect: DOMRect
  let keepAliveTimer: any = null
  
  const ensureProgressVisible = () => {
    try {
      container.classList.add('dragging-show-progress')
      container.classList.remove('plyr--hide-controls')
      container.classList.add('plyr--controls-active')
    } catch {}
  }

  const verticalCancelThresholdRatio = 1 / 4 // 垂直位移超过高度1/4则取消
  const screenEdgeGuardRatio = 1 / 6 // 顶/底部1/6区域内不触发

  const getMedia = () => plyr ? (plyr as any) : (videoRef.value as any)
  const getDuration = () => plyr ? (plyr.duration as number || 0) : ((videoRef.value?.duration as number) || 0)
  const getCurrentTime = () => plyr ? (plyr.currentTime as number || 0) : ((videoRef.value?.currentTime as number) || 0)
  const setCurrentTime = (t: number) => {
    const d = getDuration()
    const nt = Math.max(0, Math.min(d || 0, t))
    if (plyr) (plyr.currentTime = nt)
    else if (videoRef.value) videoRef.value.currentTime = nt
  }

  const onTouchStart = (e: TouchEvent) => {
    if (!container || isLongPressActive.value) return
    containerRect = container.getBoundingClientRect()
    const y = e.touches[0].clientY
    const topGuard = containerRect.top + containerRect.height * screenEdgeGuardRatio
    const bottomGuard = containerRect.bottom - containerRect.height * screenEdgeGuardRatio
    if (y <= topGuard || y >= bottomGuard) {
      determined = false
      isHorizontal = false
      isDraggingProgress = false
      return
    }
    startX = e.touches[0].clientX
    startY = e.touches[0].clientY
    startTime = getCurrentTime()
    determined = false
    isHorizontal = false
    // 不立刻标记为拖动，避免阻断长按倍速判定。
    // 仅先行保持进度条常亮，等判定为水平拖动后再置 isDraggingProgress = true
    ensureProgressVisible()
    // 在手指未抬起期间，定期刷新可见状态，防止 Plyr 自动隐藏
    try { if (keepAliveTimer) clearInterval(keepAliveTimer) } catch {}
    keepAliveTimer = setInterval(() => {
      ensureProgressVisible()
    }, 100)
  }

  const onTouchMove = (e: TouchEvent) => {
    if (!container) return
    if (isLongPressActive.value) {
      // 长按倍速已激活时，禁止进度拖动并清理样式
      isDraggingProgress = false
      try { container.classList.remove('dragging-show-progress') } catch {}
      return
    }
    if (!containerRect) containerRect = container.getBoundingClientRect()
    const dx = e.touches[0].clientX - startX
    const dy = e.touches[0].clientY - startY

    const verticalCancel = Math.abs(dy) > containerRect.height * verticalCancelThresholdRatio
    if (!determined) {
      if (verticalCancel) {
        determined = true
        isHorizontal = false
        isDraggingProgress = false
        try {
          container.classList.remove('dragging-show-progress')
          container.classList.remove('plyr--controls-active')
        } catch {}
        try { if (keepAliveTimer) { clearInterval(keepAliveTimer); keepAliveTimer = null } } catch {}
        return
      }
      if (Math.abs(dx) > 8) {
        determined = true
        isHorizontal = true
        isDraggingProgress = true
        // 进入进度拖动：强制显示进度条，且取消长按隐藏
        try {
          container.classList.add('dragging-show-progress')
          container.classList.remove('longpress-hide-progress')
          container.classList.remove('plyr--hide-controls')
          container.classList.add('plyr--controls-active')
        } catch {}
      }
    }

    if (isHorizontal && !verticalCancel) {
      // 阻止长按倍速与点击
      e.preventDefault()
      e.stopPropagation()
      // 每次 move 都保证控件与进度条可见，避免任何闪烁
      ensureProgressVisible()
      // 按容器宽度映射到时长
      const duration = getDuration()
      if (!duration || duration <= 0) return
      const w = containerRect.width || 1
      const timePerPixel = duration / w
      const nt = startTime + dx * timePerPixel
      setCurrentTime(nt)
    }
  }

  const onTouchEnd = (_e: TouchEvent) => {
    determined = false
    isHorizontal = false
    isDraggingProgress = false
    try {
      container.classList.remove('dragging-show-progress')
      container.classList.remove('plyr--controls-active')
    } catch {}
    try { if (keepAliveTimer) { clearInterval(keepAliveTimer); keepAliveTimer = null } } catch {}
    // 结束拖动后，给予控件短暂显示时间，避免立即被隐藏造成的闪烁
    try {
      container.classList.remove('plyr--hide-controls')
      setTimeout(() => {}, 120)
    } catch {}
  }

  // 触摸事件
  container.addEventListener('touchstart', onTouchStart, { passive: true })
  container.addEventListener('touchmove', onTouchMove, { passive: false })
  container.addEventListener('touchend', onTouchEnd, { passive: true })
  container.addEventListener('touchcancel', onTouchEnd, { passive: true })
}
</script>

<style scoped>
.watch-view {
  padding: 12px 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.card-header h2 {
  margin: 0;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.header-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.player-wrap { position: relative; width: 100%; }
.video-player :deep(.vjs-big-play-button) {
  top: 50% !important;
  left: 50% !important;
  transform: translate(-50%, -50%) !important;
}

/* 优化倍速菜单显示 */
.video-player :deep(.vjs-playback-rate-menu-button) {
  margin-right: 8px;
}

.longpress-hide-progress :deep(.plyr__progress) {
  display: none !important;
}

/* 进度拖动时，强制显示 Plyr 进度条 */
.dragging-show-progress :deep(.plyr__progress) {
  display: block !important;
}

/* 原生 video 控件隐藏进度条（WebKit 内核） */
.longpress-hide-progress video::-webkit-media-controls-timeline {
  display: none !important;
}
.longpress-hide-progress video::-webkit-media-controls-current-time-display,
.longpress-hide-progress video::-webkit-media-controls-time-remaining-display {
  display: none !important;
}

/* 原生 video 进度拖动时强制显示进度条（WebKit 内核） */
.dragging-show-progress video::-webkit-media-controls-timeline,
.dragging-show-progress video::-webkit-media-controls-current-time-display,
.dragging-show-progress video::-webkit-media-controls-time-remaining-display {
  display: block !important;
}

.video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-content) {
  background: rgba(0, 0, 0, 0.9);
  border-radius: 4px;
  padding: 4px 0;
}

.video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-item) {
  padding: 8px 16px;
  color: #fff;
  font-size: 14px;
  text-align: center;
}

.video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-item:hover) {
  background: rgba(255, 255, 255, 0.1);
}

.video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-item.vjs-selected) {
  background: #1890ff;
  color: #fff;
}

/* 播放器容器样式 */
.player-container {
  margin-bottom: 12px;
}

.player-scheme-info {
  margin-bottom: 8px;
  display: flex;
  justify-content: center;
}

.player-wrap {
  position: relative;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
}

.video-loading-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.loading-content {
  text-align: center;
  color: white;
}

.loading-text {
  margin-top: 12px;
  font-size: 14px;
  color: #fff;
}

.network-speed {
  margin-top: 8px;
  font-size: 12px;
  color: #ccc;
}

.plyr-video {
  width: 100%;
  max-width: 100%;
  height: auto;
  aspect-ratio: 16/9;
}

/* 确保播放器控件在移动端也能正常显示 */
@media (max-width: 768px) {
  .player-container {
    margin: 0 0 12px 0; /* 与内容同宽 */
  }
  
  .player-wrap {
    width: 100%; /* 跟随 watch-view 内容宽度 */
    margin-left: 0;
    max-width: 100%;
  }
  
  .plyr-video {
    width: 100% !important; /* 占满容器宽度 */
    height: auto !important;
    aspect-ratio: 16/9; /* 保持16:9比例 */
    max-width: 100% !important;
  }
  
  .video-player :deep(.vjs-control-bar) {
    height: 40px;
  }
  
  .video-player :deep(.vjs-playback-rate-menu-button) {
    font-size: 12px;
    padding: 0 4px;
  }
}
.player-actions { margin-top: 8px; }
.detail-layout {
  display: flex;
  gap: 16px;
}
.detail-main {
  flex: 1;
  min-width: 0;
}
.kv-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
}
.kv-item { display: flex; gap: 8px; }
.kv-item .k { color: #64748b; min-width: 72px; }
.kv-item .v { color: #0f172a; flex: 1; min-width: 0; overflow-wrap: anywhere; word-break: break-word; }
.desc { white-space: pre-wrap; line-height: 1.6; }
.res-list { display: flex; flex-direction: column; gap: 8px; }
.res-item { padding: 8px; border: 1px solid #e5e7eb; border-radius: 6px; }
.res-name { font-weight: 600; margin-bottom: 4px; }
.res-url { color: #334155; word-break: break-all; overflow-wrap: anywhere; }
 
.ep-list { display: flex; flex-wrap: wrap; gap: 8px; }
.ep-btn { max-width: 100%; }

@media (max-width: 768px) {
  .card-header { flex-direction: column; align-items: flex-start; }
  .kv-list { grid-template-columns: 1fr; }
  .card-header h2 { white-space: normal; font-size: 18px; }
  
  /* 移动端播放器控制区域优化 */
  .player-actions {
    flex-direction: column;
    gap: 8px;
  }
  
  .player-actions .ant-space {
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .player-actions .ant-space-item {
    margin-bottom: 4px;
  }
}
</style>


