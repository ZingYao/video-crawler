<template>
  <AppLayout :page-title="`播放 - ${displayTitle}`">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <div class="header-info">
            站点：
            <div v-if="currentSourceName" class="source-info">
              <a-tag color="blue" size="small">
                <template #icon>
                  <PlayCircleOutlined />
                </template>
                {{ currentSourceName }}
              </a-tag>
            </div>
          </div>
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
              />


            </div>
            <div class="player-actions">
              <!-- 第一行：基础控制 -->
              <div class="player-actions-row">
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
                  <a-button
                    size="small"
                    type="primary"
                    @click="searchOtherSites"
                    :loading="searchingOtherSites"
                  >
                    <template #icon>
                      <SearchOutlined />
                    </template>
                    其他站点
                  </a-button>
                </a-space>
              </div>

              <!-- 第二行：跳过片首控制 -->
              <div class="player-actions-row">
                <a-space wrap>
                  <span class="skip-label">跳过片首：</span>
                  <a-switch
                    v-model:checked="skipIntro.enabled"
                    size="small"
                    @change="handleSkipIntroChange"
                  />
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
                  <span v-if="skipIntro.enabled" class="skip-unit">秒</span>
                </a-space>
              </div>

              <!-- 第三行：跳过片尾控制 -->
              <div class="player-actions-row">
                <a-space wrap>
                  <span class="skip-label">跳过片尾：</span>
                  <a-switch
                    v-model:checked="skipOutro.enabled"
                    size="small"
                    @change="handleSkipOutroChange"
                  />
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
                  <span v-if="skipOutro.enabled" class="skip-unit">秒</span>
                </a-space>
              </div>
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

  <!-- 其他站点搜索结果弹窗 -->
  <a-modal
    v-model:open="otherSitesModalVisible"
    :title="`${base.name || displayTitle || '视频'} 搜索结果`"
    width="90%"
    :footer="null"
    :destroyOnClose="true"
  >
          <div class="other-sites-search">
        <div class="search-header">
          <a-button
            type="primary"
            @click="handleSearchOtherSites"
            :loading="searchingOtherSites"
            size="large"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
            刷新
          </a-button>
        </div>

      <div class="search-results" v-if="otherSitesResults.length > 0">
        <div class="results-header">
          <span>找到 {{ otherSitesResults.length }} 个结果</span>
        </div>
        <div class="results-grid">
          <div
            v-for="result in otherSitesResults"
            :key="`${result.sourceId}-${result.url}`"
            class="result-card"
            @click="playFromOtherSite(result)"
          >
            <div class="card-cover">
              <img
                :src="result.cover || result.poster || '/favicon.ico'"
                :alt="result.name"
                @error="handleImageError"
              />
            </div>
            <div class="card-content">
              <h4 class="card-title">{{ result.name || result.title || '未知标题' }}</h4>
              <p class="card-source">来源：{{ result.sourceName }}</p>

              <!-- 视频信息 -->
              <div class="card-info">
                <p v-if="result.director" class="card-director">导演：{{ result.director }}</p>
                <p v-if="result.actor" class="card-actor">主演：{{ result.actor }}</p>
                <p v-if="result.release_date" class="card-date">上映：{{ result.release_date }}</p>
                <p v-if="result.region" class="card-region">地区：{{ result.region }}</p>
              </div>

              <!-- 描述信息 -->
              <p class="card-desc" v-if="result.description">{{ result.description }}</p>

              <!-- 评分和类型 -->
              <div class="card-meta">
                <span v-if="result.rate || result.score" class="rating">评分：{{ result.rate || result.score }}</span>
                <span v-if="result.type" class="type">{{ result.type }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="hasSearchedOtherSites && !searchingOtherSites" class="no-results">
        <a-empty description="未找到相关视频" />
      </div>
    </div>
  </a-modal>

</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppLayout from '@/components/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { useSettingsStore } from '@/stores/settings'
import { videoAPI, videoSourceAPI } from '@/api'
import { localHistoryManager } from '@/utils/localHistory'
import Plyr from 'plyr'
import Hls from 'hls.js'
import 'plyr/dist/plyr.css'
import { ThunderboltOutlined, PlayCircleOutlined, ClockCircleOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const configStore = useConfigStore()
const settingsStore = useSettingsStore()

const loading = ref(false)
const error = ref('')
const detailData = ref<any>(null)
const fromCache = ref(false)
const downloading = ref(false)

// 其他站点搜索相关变量
const otherSitesModalVisible = ref(false)
const searchingOtherSites = ref(false)
const otherSitesResults = ref<any[]>([])
const hasSearchedOtherSites = ref(false)

// 站点名称缓存
const sourceNameCache = ref<Map<string, string>>(new Map())

const networkSpeed = ref('')
const originalRate = ref(1) // 保存原始播放速率
const isLongPressActive = ref(false) // 长按状态

// 跳过片首片尾相关变量
const skipIntro = ref({ enabled: false, seconds: 30 })
const skipOutro = ref({ enabled: false, seconds: 30 })
const skipOutroTriggered = ref(false) // 跟踪当前剧集是否已触发跳过片尾
const skipOutroCurrentUrl = ref<string>('') // 跟踪当前剧集的URL，用于判断是否切换了剧集
const skipOutroLastTriggerTime = ref<number>(0) // 上次触发跳过片尾的时间戳
const skipOutroCooldownTime = 10000 // 10秒冷却时间（毫秒）

// 根据 original_url 隔离的跳过片首片尾缓存键
const skipSettingsKey = computed(() => {
  const originalUrl = String(route.query.original_url || route.query.url || '')
  return `skip_settings:${encodeURIComponent(originalUrl)}`
})

// 全局播放倍速缓存键
const globalRateKey = 'global_playback_rate'

// 下一集预加载相关变量
const nextEpisodeUrl = ref<string>('') // 下一集的播放链接
const nextEpisodePreloaded = ref(false) // 是否已预加载下一集
const nextEpisodePreloadTimer = ref<number | null>(null) // 预加载定时器
const nextEpisodeToastTimer = ref<number | null>(null) // 提示定时器
const nextEpisodeToastShown = ref(false) // 是否已显示切换提示

// 网速计算相关变量
let lastLoadedBytes = 0
let lastSpeedCheckTime = 0
let speedCheckInterval: any = null
let longPressTimer: any = null // 长按定时器

// 倍速监听相关变量
let rateCheckInterval: any = null
let lastVideoRate = 1

// 新增：视频播放控制相关变量
const isProgressVisible = ref(false) // 进度条是否可见
const progressBarTimer = ref<number | null>(null) // 进度条显示定时器
const clickTimer = ref<number | null>(null) // 点击定时器
const dragStartTime = ref(0) // 拖动开始时间
const dragStartX = ref(0) // 拖动开始X坐标
const dragStartY = ref(0) // 拖动开始Y坐标
const isHorizontalDrag = ref(false) // 是否为横向拖动
const isVerticalDrag = ref(false) // 是否为纵向拖动
// 长按倍速隐藏进度条定时器
let longPressHideTimer: number | null = null
// 滑动进度显示进度条定时器
let dragShowTimer: number | null = null

const sourceId = computed(() => String(route.params.sourceId || ''))
const videoUrl = computed(() => String(route.query.original_url || route.query.url || ''))
const currentPlayUrl = ref<string>('')
const originalUrl = computed(() => String(route.query.original_url || route.query.url || videoUrl.value || ''))
// 提前定义 base，避免初始化顺序问题导致 TDZ 报错
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

const displayTitle = computed(() => {
  const ep = flatEpisodes.value.find(e => e.url === currentPlayUrl.value)
  if (ep) return `${base.value.name || ''} - ${ep.name}`.trim()
  return String(route.query.title || base.value.name || '')
})

// 获取站点名称
async function getSourceName(sourceId: string): Promise<string> {
  // 先从缓存获取
  if (sourceNameCache.value.has(sourceId)) {
    return sourceNameCache.value.get(sourceId) || ''
  }

  try {
    const token = auth.token!
    const response: any = await videoSourceAPI.getVideoSourceDetail(token, sourceId)
    if (response?.code === 0 && response?.data) {
      const sourceName = response.data.name || ''
      // 缓存站点名称
      sourceNameCache.value.set(sourceId, sourceName)
      return sourceName
    }
  } catch (error) {
    console.error('获取站点名称失败:', error)
  }

  return ''
}

// 当前站点名称
const currentSourceName = ref('')

// 预加载所有站点名称（仅预加载一次）
let sourceNamesPreloaded = false
async function preloadAllSourceNames() {
  if (sourceNamesPreloaded) return
  try {
    const token = auth.token!
    const response: any = await videoSourceAPI.getVideoSourceList(token)
    if (response?.code === 0 && response?.data) {
      const sources = Array.isArray(response.data) ? response.data : []
      sources.forEach((source: any) => {
        if (source.id && source.name) {
          sourceNameCache.value.set(source.id, source.name)
        }
      })
      console.log(`[SourceName] 预加载了 ${sources.length} 个站点名称`)
      sourceNamesPreloaded = true
    }
  } catch (error) {
    console.error('预加载站点名称失败:', error)
  }
}

// 更新当前站点名称
async function updateCurrentSourceName() {
  if (sourceId.value) {
    currentSourceName.value = await getSourceName(sourceId.value)
  }
}

const cacheKey = computed(() => `watch_detail:${sourceId.value}:${encodeURIComponent(videoUrl.value)}`)
// 使用 sourceId + 当前播放的剧集URL 作为进度键，确保每个剧集都有独立的缓存
const playStateKey = computed(() => {
  const keyUrl = String(currentPlayUrl.value || route.query.original_url || videoUrl.value || '')
  return `watch_state:${sourceId.value}:${encodeURIComponent(keyUrl)}`
})
// 播放链接缓存键生成函数，每个剧集都有独立的播放链接缓存
function getPlayUrlCacheKey(episodeUrl: string): string {
  return `play_url:${sourceId.value}:${encodeURIComponent(episodeUrl)}`
}



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
  if (!playerSource.value || typeof playerSource.value !== 'string') return '未加载'

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
  if (!playerSource.value || typeof playerSource.value !== 'string') return '未加载'

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

// 同步倍速到页面显示
function syncRateToUI(newRate: number) {
  rate.value = newRate
  // 更新lastVideoRate，避免倍速监听器误判
  lastVideoRate = newRate
}

// 处理播放速率变化
function handleRateChange(value: number) {
  syncRateToUI(value)
  // 保存全局倍速设置
  saveGlobalRate()
  try {
    if (plyr) {
      plyr.speed = value
    } else if (videoRef.value) {
      // 原生 video 兜底
      ;(videoRef.value as any).playbackRate = value as any
    }
  } catch {}
}

// 保存跳过片首片尾设置（根据 original_url 隔离）
function saveSkipSettings() {
  try {
    const settings = {
      skipIntro: {
        enabled: skipIntro.value.enabled,
        seconds: skipIntro.value.seconds
      },
      skipOutro: {
        enabled: skipOutro.value.enabled,
        seconds: skipOutro.value.seconds
      }
    }
    localStorage.setItem(skipSettingsKey.value, JSON.stringify(settings))
    console.log('[SkipSettings] 已保存跳过设置:', settings)
  } catch (e) {
    console.error('[SkipSettings] 保存跳过设置失败:', e)
  }
}

// 加载跳过片首片尾设置（根据 original_url 隔离）
function loadSkipSettings() {
  try {
    const raw = localStorage.getItem(skipSettingsKey.value)
    if (!raw) return
    
    const settings = JSON.parse(raw)
    if (settings.skipIntro) {
      skipIntro.value.enabled = settings.skipIntro.enabled
      skipIntro.value.seconds = settings.skipIntro.seconds
    }
    if (settings.skipOutro) {
      skipOutro.value.enabled = settings.skipOutro.enabled
      skipOutro.value.seconds = settings.skipOutro.seconds
    }
    console.log('[SkipSettings] 已加载跳过设置:', settings)
  } catch (e) {
    console.error('[SkipSettings] 加载跳过设置失败:', e)
  }
}

// 保存全局播放倍速
function saveGlobalRate() {
  try {
    localStorage.setItem(globalRateKey, JSON.stringify(rate.value))
    console.log('[GlobalRate] 已保存全局倍速:', rate.value)
  } catch (e) {
    console.error('[GlobalRate] 保存全局倍速失败:', e)
  }
}

// 加载全局播放倍速
function loadGlobalRate() {
  try {
    const raw = localStorage.getItem(globalRateKey)
    if (!raw) return
    
    const savedRate = JSON.parse(raw)
    if (typeof savedRate === 'number' && rates.includes(savedRate)) {
      rate.value = savedRate
      console.log('[GlobalRate] 已加载全局倍速:', savedRate)
    }
  } catch (e) {
    console.error('[GlobalRate] 加载全局倍速失败:', e)
  }
}

// 处理跳过片首变化
function handleSkipIntroChange() {
  saveSkipSettings()
}

// 处理跳过片尾变化
function handleSkipOutroChange() {
  saveSkipSettings()
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
    // 优化控件显示逻辑
    hideControls: true, // 启用自动隐藏控件
    resetOnEnd: false, // 播放结束时不要重置
  })

  // 立即禁用 Plyr 的双击全屏功能
  disablePlyrDoubleClick()

  // 绑定 Plyr 的播放完成事件
  plyr.on('ended', () => {
    try {
      // 删除当前剧集的播放进度缓存
      deletePlayStateCache()

      if (canNext.value) {
        playNext()
      }
    } catch (e) {
      console.error('[Plyr] 自动切换下一集失败:', e)
    }
  })

  // Plyr 视频等待数据事件（卡住检测）
  plyr.on('waiting', () => {
    console.log('[Plyr] 视频等待数据')
  })

  // Plyr 视频可以播放事件（恢复检测）
  plyr.on('canplay', () => {
    console.log('[Plyr] 视频可以播放')
  })

  // Plyr 视频可以流畅播放事件
  plyr.on('canplaythrough', () => {
    console.log('[Plyr] 视频可以流畅播放')
  })

  // Plyr 播放状态 -> Screen Wake Lock
  plyr.on('play', async () => { await requestWakeLock() })
  plyr.on('pause', async () => { await releaseWakeLock() })
  plyr.on('ended', async () => { await releaseWakeLock() })

  // 优化控件显示逻辑
  plyr.on('enterfullscreen', () => {
    console.log('[Plyr] 进入全屏')
    // 全屏时允许控件自动隐藏
    try {
      const container = plyr.elements.container
      container.classList.remove('dragging-show-progress')
      container.classList.remove('longpress-hide-progress')
    } catch {}
  })

  plyr.on('exitfullscreen', () => {
    console.log('[Plyr] 退出全屏')
    // 退出全屏时恢复正常控件显示
  })

  bindPlayerEvents()

  // 手势：左右滑动调节进度（Plyr 容器）
  try {
    const container = plyr?.elements?.container as HTMLElement
    if (container) {
      attachProgressDrag(container)
    }
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
    const container = plyr?.elements?.container as HTMLElement | undefined
    const video = plyr?.elements?.video as HTMLElement | undefined

    // 克隆元素来移除所有事件监听器
    if (container && video) {
      const newContainer = container.cloneNode(true) as HTMLElement
      const newVideo = video.cloneNode(true) as HTMLElement

      container.parentNode?.replaceChild(newContainer, container)
      newContainer.appendChild(newVideo)

      // 重新设置 Plyr 的元素引用
      ;(plyr.elements as any).container = newContainer
      ;(plyr.elements as any).video = newVideo
    }
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
  try { (plyr?.elements?.container as HTMLElement)?.addEventListener('dblclick', preventDoubleClick, true) } catch {}
  try { (plyr?.elements?.video as HTMLElement)?.addEventListener('dblclick', preventDoubleClick, true) } catch {}

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

  const containerEl = (plyr?.elements?.container as HTMLElement | undefined)
  if (!containerEl) return

  containerEl.addEventListener('click', (e: any) => {
    // 避免与Plyr的控件点击事件冲突
    const target = e.target as HTMLElement
    if (target && (target.closest('.plyr__control') || target.closest('.plyr__progress') || target.closest('.plyr__controls'))) {
      return
    }
    
    // 避免与拖动冲突
    if (isDraggingProgress) {
      return
    }
    
    const currentTime = Date.now()
    if (currentTime - plyrLastClickTime < plyrDoubleClickThreshold) {
      // 双击事件 - 播放/暂停
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
      // 单击事件 - 切换进度条显示
      plyrLastClickTime = currentTime
      console.log('[PROGRESS_CTRL] [Plyr Click] 双击检测失败，执行单击切换进度条')
      toggleProgressBar()
    }
  })

  // 长按2倍速播放功能
  let plyrLongPressTimer: any = null
  let plyrIsTouchActive = false

  // 触摸开始
  containerEl.addEventListener('touchstart', (e: any) => {
    if (isDraggingProgress) return
    plyrIsTouchActive = true
    if (plyrLongPressTimer) clearTimeout(plyrLongPressTimer)
    
    // 设置单击定时器
    if (clickTimer.value) {
      clearTimeout(clickTimer.value)
    }
    clickTimer.value = setTimeout(() => {
      if (!isDraggingProgress && !isLongPressActive.value && plyrIsTouchActive) {
        console.log('[PROGRESS_CTRL] [Plyr Touch] 触摸单击定时器触发，执行切换进度条')
        toggleProgressBar()
      }
    }, 200)
    
    plyrLongPressTimer = setTimeout(() => {
      if (plyrIsTouchActive && !isDraggingProgress) {
        originalRate.value = plyr.speed
        plyr.speed = longPressSpeed.value
        syncRateToUI(longPressSpeed.value)
        isLongPressActive.value = true
        console.log(`[Plyr LongPress] 启动${longPressSpeed.value}倍速播放`)
        // 震动反馈
        vibrateFeedback()
        // 显示toast提示
        console.log(`[Plyr LongPress] 显示Toast: 已启动${longPressSpeed.value}倍速播放`)
        message.info(`已启动${longPressSpeed.value}倍速播放`, 3)
        
        // 长按倍速期间，每3秒隐藏一次进度条
        console.log('[PROGRESS_CTRL] [Plyr LongPress] 启动长按倍速隐藏进度条定时器')
        // 立即执行一次隐藏进度条
        console.log('[PROGRESS_CTRL] [Plyr LongPress] 立即执行：隐藏进度条')
        hideProgressBar()
        // 启动定时器，每3秒隐藏一次
        longPressHideTimer = setInterval(() => {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 定时器触发：隐藏进度条')
          hideProgressBar()
        }, 3000)
      }
    }, 500)
  })

  // 触摸移动
  containerEl.addEventListener('touchmove', (e: any) => {
    // 如果有移动，取消点击事件
    if (clickTimer.value) {
      clearTimeout(clickTimer.value)
      clickTimer.value = null
    }
  })

  // 触摸结束
  containerEl.addEventListener('touchend', (e: any) => {
    plyrIsTouchActive = false
    if (plyrLongPressTimer) {
      clearTimeout(plyrLongPressTimer)
      plyrLongPressTimer = null
    }
    if (isLongPressActive.value) {
      plyr.speed = originalRate.value
      syncRateToUI(originalRate.value)
      isLongPressActive.value = false
      console.log('[Plyr LongPress] 恢复原始播放速率:', originalRate.value)
      // 震动反馈
      vibrateFeedback()
      // 显示toast提示
      console.log('[Plyr LongPress] 显示Toast: 已恢复倍速播放')
      message.info(`已恢复${originalRate.value}x倍速播放`, 3)
      
              // 清理长按倍速隐藏进度条定时器
        if (longPressHideTimer) {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 清理长按倍速隐藏进度条定时器')
          clearInterval(longPressHideTimer)
          longPressHideTimer = null
        }
    }
    
    // 清除单击定时器
    if (clickTimer.value) {
      clearTimeout(clickTimer.value)
      clickTimer.value = null
    }
  })

  // 触摸取消
  containerEl.addEventListener('touchcancel', (e: any) => {
    plyrIsTouchActive = false
    if (plyrLongPressTimer) {
      clearTimeout(plyrLongPressTimer)
      plyrLongPressTimer = null
    }
    if (isLongPressActive.value) {
      plyr.speed = originalRate.value
      syncRateToUI(originalRate.value)
      isLongPressActive.value = false
      
              // 清理长按倍速隐藏进度条定时器
        if (longPressHideTimer) {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 触摸取消：清理长按倍速隐藏进度条定时器')
          clearInterval(longPressHideTimer)
          longPressHideTimer = null
        }
    }
    
    // 清除单击定时器
    if (clickTimer.value) {
      clearTimeout(clickTimer.value)
      clickTimer.value = null
    }
  })

  // 鼠标按下（桌面端）
  containerEl.addEventListener('mousedown', (e: any) => {
    if (e.button === 0) {
      if (isDraggingProgress) return
      plyrIsTouchActive = true
      if (plyrLongPressTimer) clearTimeout(plyrLongPressTimer)
      
      // 设置单击定时器
      if (clickTimer.value) {
        clearTimeout(clickTimer.value)
      }
      clickTimer.value = setTimeout(() => {
        if (!isDraggingProgress && !isLongPressActive.value && plyrIsTouchActive) {
          toggleProgressBar()
        }
      }, 200)
      
      plyrLongPressTimer = setTimeout(() => {
        if (plyrIsTouchActive && !isDraggingProgress) {
          originalRate.value = plyr.speed
          plyr.speed = longPressSpeed.value
          syncRateToUI(longPressSpeed.value)
          isLongPressActive.value = true
          console.log(`[Plyr LongPress] 启动${longPressSpeed.value}倍速播放`)
          // 震动反馈
          vibrateFeedback()
          // 显示toast提示
          console.log(`[Plyr LongPress] 显示Toast: 已启动${longPressSpeed.value}倍速播放`)
          message.info(`已启动${longPressSpeed.value}倍速播放`, 3)
          
                  // 长按倍速期间，每3秒隐藏一次进度条
        console.log('[PROGRESS_CTRL] [Plyr LongPress] 启动长按倍速隐藏进度条定时器')
        // 立即执行一次隐藏进度条
        console.log('[PROGRESS_CTRL] [Plyr LongPress] 立即执行：隐藏进度条')
        hideProgressBar()
        // 启动定时器，每3秒隐藏一次
        longPressHideTimer = setInterval(() => {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 定时器触发：隐藏进度条')
          hideProgressBar()
        }, 3000)
        }
      }, 500)
    }
  })

  // 鼠标松开（桌面端）
  containerEl.addEventListener('mouseup', (e: any) => {
    if (e.button === 0) {
      plyrIsTouchActive = false
      if (plyrLongPressTimer) {
        clearTimeout(plyrLongPressTimer)
        plyrLongPressTimer = null
      }
      if (isLongPressActive.value) {
        plyr.speed = originalRate.value
        syncRateToUI(originalRate.value)
        isLongPressActive.value = false
        console.log('[Plyr LongPress] 恢复原始播放速率:', originalRate.value)
        // 震动反馈
        vibrateFeedback()
        // 显示toast提示
        console.log('[Plyr LongPress] 显示Toast: 已恢复倍速播放')
        message.info(`已恢复${originalRate.value}x倍速播放`, 3)
        
        // 清理长按倍速隐藏进度条定时器
        if (longPressHideTimer) {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 鼠标松开：清理长按倍速隐藏进度条定时器')
          clearInterval(longPressHideTimer)
          longPressHideTimer = null
        }
      }
      
      // 清除单击定时器
      if (clickTimer.value) {
        clearTimeout(clickTimer.value)
        clickTimer.value = null
      }
    }
  })

  // 鼠标离开（桌面端）
  containerEl.addEventListener('mouseleave', (e: any) => {
    if (plyrIsTouchActive) {
      plyrIsTouchActive = false
      if (plyrLongPressTimer) {
        clearTimeout(plyrLongPressTimer)
        plyrLongPressTimer = null
      }
      if (isLongPressActive.value) {
        plyr.speed = originalRate.value
        syncRateToUI(originalRate.value)
        isLongPressActive.value = false
        
        // 清理长按倍速隐藏进度条定时器
        if (longPressHideTimer) {
          console.log('[PROGRESS_CTRL] [Plyr LongPress] 鼠标离开：清理长按倍速隐藏进度条定时器')
          clearInterval(longPressHideTimer)
          longPressHideTimer = null
        }
      }
    }
  })

  // 鼠标点击事件已合并到上面的双击播放/暂停事件中
}
let lastSavedSecond = 0
let playerBound = false
let lastVideoW = 0
let lastVideoH = 0
let orientationLocked = false
// 全屏期间键盘快进/快退
let fullscreenKeyHandler: ((e: KeyboardEvent) => void) | null = null
const fullscreenSeekStep = 10 // 秒
// 获取当前应播放的剧集 URL：根据 URL 参数和缓存决定播放逻辑
function getSelectedEpisodeUrl(): string {
  let url = ''
  // 获取 URL 参数中的 title 和 source（优先按显式参数匹配具体剧集）
  const titleParam = String(route.query.title || '')
  const sourceParam = String((route.query as any).source || '')

  try { console.log('[Play] URL参数检查 - title:', titleParam, 'source:', sourceParam) } catch {}

  // 逻辑1: URL 中有 source 和 title 时，优先播放对应剧集
  if (titleParam && sourceParam) {
    try { console.log('[Play] 逻辑1: URL中有source和title，查找对应剧集') } catch {}
    const targetSource = sourcesByTab.value.find(s => s.name === sourceParam)
    if (targetSource && Array.isArray(targetSource.episodes)) {
      const targetEpisode = targetSource.episodes.find(ep => ep.name === titleParam)
      if (targetEpisode?.url) {
        url = targetEpisode.url
        try { console.log('[Play] 逻辑1成功: 找到对应Source和Title的剧集:', sourceParam, titleParam, '->', url) } catch {}
        return url
      } else {
        try { console.log('[Play] 逻辑1失败: 在Source中找到的剧集URL无效或不存在') } catch {}
      }
    } else {
      try { console.log('[Play] 逻辑1失败: 找不到对应的Source:', sourceParam) } catch {}
    }
  }

  // 逻辑2: 当前已有播放URL且属于已解析的剧集列表时使用
  const cur = String(currentPlayUrl.value || '')
  if (cur && flatEpisodes.value.some(e => e.url === cur)) {
    try { console.log('[Play] 逻辑2: 使用当前已在剧集列表中的播放URL:', cur) } catch {}
    return cur
  }
  
  // 逻辑3: URL 中没有 source 和 title 时，查看缓存中是否存在播放的 source 和剧集
  if (!titleParam && !sourceParam) {
    try { console.log('[Play] 逻辑3: URL中没有source和title，查看缓存') } catch {}
    
    const state = loadPlayState()
    if (state?.url && state?.source && state?.title) {
      try { console.log('[Play] 缓存中存在播放信息:', state) } catch {}
      
      // 查找缓存中对应的 source
      const cachedSource = sourcesByTab.value.find(s => s.name === state.source)
      if (cachedSource && Array.isArray(cachedSource.episodes)) {
        // 在缓存对应的 source 中查找对应标题的剧集
        const cachedEpisode = cachedSource.episodes.find(ep => ep.name === state.title)
        if (cachedEpisode && cachedEpisode.url) {
          url = cachedEpisode.url
          try { console.log('[Play] 逻辑3成功: 使用缓存中的Source和Title:', state.source, state.title, '->', url) } catch {}
          return url
        } else {
          try { console.log('[Play] 逻辑3失败: 缓存中的剧集在当前Source中找不到或URL无效') } catch {}
        }
      } else {
        try { console.log('[Play] 逻辑3失败: 缓存中的Source不存在:', state.source) } catch {}
      }
    } else {
      try { console.log('[Play] 逻辑3失败: 缓存中不存在有效的播放信息') } catch {}
    }
  } else {
    try { console.log('[Play] 逻辑3跳过: URL中有参数，不使用缓存逻辑') } catch {}
  }
  
  // 逻辑4: 兜底逻辑 - 播放第一个资源的第一个剧集
  if (sourcesByTab.value.length > 0) {
    const firstSource = sourcesByTab.value[0]
    if (firstSource && Array.isArray(firstSource.episodes) && firstSource.episodes.length > 0) {
      url = String(firstSource.episodes[0]?.url || '')
      try { console.log('[Play] 逻辑4: 兜底 - 播放第一个Source的第一集:', firstSource.name, '->', url) } catch {}
      return url
    }
  }
  
  // 逻辑5: 最后兜底 - 使用 flatEpisodes 的第一集
  if (flatEpisodes.value.length > 0) {
    url = String(flatEpisodes.value[0]?.url || '')
    try { console.log('[Play] 逻辑5: 最后兜底 - 使用 flatEpisodes 第一集:', url) } catch {}
    return url
  }
  
  // 逻辑6: 最终回退到 original_url
  url = String(videoUrl.value || '')
  try { console.log('[Play] 逻辑6: 最终回退到 original_url:', url) } catch {}
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

        // 更新观看历史（无需登录模式下使用本地缓存）
        if (!configStore.needsLogin() && sourceId.value && videoUrl.value) {
          const videoId = `${sourceId.value}|${videoUrl.value}`
          const progress = dur > 0 ? ct / dur : 0
          localHistoryManager.updateVideoProgress(videoId, ct, progress)
        }
      }

      // 检查是否需要跳过片首
      if (skipIntro.value.enabled && ct < skipIntro.value.seconds) {
        v.currentTime = skipIntro.value.seconds
        console.log(`跳过片首，跳转到 ${skipIntro.value.seconds} 秒`)
      }

      // 下一集预加载和提示逻辑
      if (canNext.value && dur > 0) {
        const remainingTime = dur - ct

        // 提前20秒预加载下一集
        if (remainingTime <= 20 && !nextEpisodePreloaded.value && !nextEpisodePreloadTimer.value) {
          nextEpisodePreloadTimer.value = window.setTimeout(() => {
            preloadNextEpisode()
          }, 100) // 延迟100ms避免频繁调用
        }

        // 提前5秒显示切换提示
        if (remainingTime <= 5 && !nextEpisodeToastShown.value && !nextEpisodeToastTimer.value) {
          nextEpisodeToastTimer.value = window.setTimeout(() => {
            const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
            message.info(`即将切换到下一集：${nextEpisode?.name || '下一集'}`, 4)
            nextEpisodeToastShown.value = true
          }, 100) // 延迟100ms避免频繁调用
        }
      }

      // 检查是否需要跳过片尾
      if (skipOutro.value.enabled && dur > 0) {
        // 计算片尾触发点：如果视频长度比跳过秒数短，则在视频播放到80%时触发
        const outroTriggerPoint = dur <= skipOutro.value.seconds
          ? dur * 0.8  // 视频长度较短时，在80%处触发
          : dur - skipOutro.value.seconds  // 正常情况，在片尾前指定秒数触发

        if (ct > outroTriggerPoint) {
          // 检查当前剧集URL是否发生变化，如果变化了说明切换了剧集，需要重置状态
          if (skipOutroCurrentUrl.value !== currentPlayUrl.value) {
            skipOutroTriggered.value = false
            skipOutroCurrentUrl.value = currentPlayUrl.value
          }

          // 检查冷却时间
          const currentTime = Date.now()
          const timeSinceLastTrigger = currentTime - skipOutroLastTriggerTime.value
          if (timeSinceLastTrigger < skipOutroCooldownTime) {
            return
          }

          // 如果已经触发过跳过片尾，则不再触发
          if (skipOutroTriggered.value) {
            return
          }

          // 标记已触发，避免重复触发
          skipOutroTriggered.value = true
          skipOutroLastTriggerTime.value = currentTime

                      // 如果接近片尾，自动切换到下一集
            if (canNext.value) {
              // 删除当前剧集的播放进度缓存
              deletePlayStateCache()

              // 如果有预加载的URL，直接使用
              if (nextEpisodeUrl.value) {
                const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
                playEpisodeWithUrl(nextEpisode, nextEpisodeUrl.value)
              } else {
                playNext()
              }
                      } else {
              // 没有下一集，跳转到片尾前指定秒数
              const jumpToTime = dur <= skipOutro.value.seconds
                ? Math.max(0, dur * 0.9)  // 视频长度较短时，跳转到90%处
                : Math.max(0, dur - skipOutro.value.seconds)  // 正常情况，跳转到片尾前指定秒数
              v.currentTime = jumpToTime
            }
        }
      }
    } catch {}
  })

  // 播放完成自动切换下一集（原生事件）
  v.addEventListener('ended', () => {
    try {
      if (canNext.value) {
        playNext()
      }
    } catch (e) {
      console.error('[Video] 自动切换下一集失败:', e)
    }
  })

  // 视频等待数据事件（卡住检测）
  v.addEventListener('waiting', () => {
    console.log('[Video] 视频等待数据')
  })

  // 视频可以播放事件（恢复检测）
  v.addEventListener('canplay', () => {
    console.log('[Video] 视频可以播放')
  })

  // 视频可以流畅播放事件
  v.addEventListener('canplaythrough', () => {
    console.log('[Video] 视频可以流畅播放')
  })

  // 双击播放/暂停功能（原生 video 容器点击处理与 Plyr 一致）
  let lastClickTime = 0
  let clickCount = 0
  const doubleClickThreshold = 300 // 双击时间阈值（毫秒）

  v.addEventListener('click', (e) => {
    // 如果已经初始化了Plyr，让Plyr处理点击事件
    if (plyr) {
      return
    }
    
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
    if (container) {
      attachProgressDrag(container)
    }
  } catch {}

  // 启动网速监控
  if (!speedCheckInterval) {
    startSpeedMonitoring()
  }

  // 启动倍速监听
  if (!rateCheckInterval) {
    startRateMonitoring()
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

  // Android: 仅在全屏时监听横竖屏变化，非全屏不监听
  try {
    // 避免重复绑定
    removeAndroidOrientationListener()
    addAndroidOrientationListener()
  } catch {}

  // 绑定全屏键盘左右快进/快退（TV 遥控器/键盘）
  try {
    if (!fullscreenKeyHandler) {
      fullscreenKeyHandler = (e: KeyboardEvent) => {
        // 仅在处于全屏时生效
        const d: any = document as any
        const isFullscreen = !!(d.fullscreenElement || d.webkitFullscreenElement || d.mozFullScreenElement || d.msFullscreenElement)
        if (!isFullscreen) return
        // 仅处理左右方向键与等价按键
        if (e.key === 'ArrowLeft' || e.key === 'ArrowRight' || e.key === 'MediaRewind' || e.key === 'MediaFastForward') {
          e.preventDefault(); e.stopPropagation()
          const v = videoRef.value
          if (!v || isNaN(v.duration)) return
          const step = fullscreenSeekStep
          const delta = (e.key === 'ArrowRight' || e.key === 'MediaFastForward') ? step : -step
          const next = Math.max(0, Math.min((v.currentTime || 0) + delta, v.duration || 0))
          try { v.currentTime = next } catch {}
          try { message.open({ type: 'info', content: `${delta > 0 ? '快进' : '快退'} ${Math.abs(delta)} 秒`, duration: 1 }) } catch {}
        }
      }
    }
    window.addEventListener('keydown', fullscreenKeyHandler as any, { capture: true })
  } catch {}
}

async function handleExitFullscreen() {
  try { console.log('[Fullscreen] 退出全屏，尝试解锁屏幕方向') } catch {}
  await unlockOrientation()
  // 退出全屏后移除方向监听
  try { removeAndroidOrientationListener() } catch {}
  // 解绑全屏键盘监听
  try { if (fullscreenKeyHandler) window.removeEventListener('keydown', fullscreenKeyHandler as any, { capture: true } as any) } catch {}
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

// —— Android 全屏期间的横竖屏监听 ——
let androidOrientationHandler: ((e: Event) => void) | null = null

function addAndroidOrientationListener() {
  // 仅在 Android 下有意义（简易判断）
  const isAndroid = !!(window as any).AndroidKV || /Android/.test(navigator.userAgent)
  if (!isAndroid) return
  androidOrientationHandler = () => {
    // 全屏期间，根据当前视频/容器的宽高估算并锁定
    const ori = estimateOrientation()
    try { console.log('[Fullscreen] orientationchange ->', ori) } catch {}
    void lockOrientation(ori)
  }
  window.addEventListener('orientationchange', androidOrientationHandler, { passive: true })
}

function removeAndroidOrientationListener() {
  if (androidOrientationHandler) {
    window.removeEventListener('orientationchange', androidOrientationHandler)
    androidOrientationHandler = null
  }
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
// 仅判断"当前来源"是否还有下一集
const canNext = computed(() => currentIndex.value >= 0 && currentIndex.value < currentSourceEpisodes.value.length - 1)

function isCurrentEpisode(ep: { url: string }) {
  return String(ep?.url || '') === currentPlayUrl.value
}

// 使用预加载的URL播放剧集
async function playEpisodeWithUrl(ep: { name: string; url: string }, preloadedUrl: string, sourceName?: string) {
  if (!ep?.url || !preloadedUrl) {
    return
  }

  // 清理预加载状态
  clearNextEpisodePreload()

  // 立刻更新当前剧集，用于后续解析播放链接与按钮状态
  currentPlayUrl.value = ep.url

  // 重置跳过片尾状态，新剧集可以重新触发
  skipOutroTriggered.value = false
  skipOutroCurrentUrl.value = ep.url
  skipOutroLastTriggerTime.value = 0 // 重置冷却时间

  // 更新地址栏中标题与来源，不修改 original_url 参数，避免影响回显
  const q = { 
    ...route.query, 
    title: ep.name, 
    source: sourceName || (ep as any).__sourceName,
    // 保持 original_url 不变，避免触发路由监听器
    original_url: route.query.original_url || route.query.url
  }
  router.replace({ name: 'watch', params: route.params, query: q })
  
  // 立即缓存当前播放的剧集信息
  savePlayState({ url: ep.url, title: ep.name, source: q.source })

  try {
    if (!preloadedUrl || typeof preloadedUrl !== 'string') {
      console.error('[playEpisodeWithUrl] 预加载URL无效')
      return
    }
    playerSource.value = preloadedUrl
    await nextTick()
    ensurePlyr()
    if (videoRef.value) {
      // 切换播放源
      if (Hls.isSupported()) {
        if (hls) { try { hls.destroy() } catch {} }
        hls = new Hls({ maxBufferLength: 30 })
        hls.loadSource(preloadedUrl)
        hls.attachMedia(videoRef.value)
      } else {
        videoRef.value.src = preloadedUrl
      }
      // 设置倍速
      if (plyr) {
        try {
          plyr.speed = rate.value
        } catch {}
      } else if (videoRef.value) {
        try {
          videoRef.value.playbackRate = rate.value
        } catch {}
      }
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

      // 滚动到视频播放器
      await nextTick()
      if (videoRef.value) {
        videoRef.value.scrollIntoView({
          behavior: 'smooth',
          block: 'center'
        })
      }
    }
    
    // 播放成功后，预加载下一集的播放链接
    setTimeout(() => {
      preloadNextEpisodeUrl()
    }, 1000) // 延迟1秒执行，避免影响当前播放
    
    // 注：剧集信息已在函数开始时缓存，此处无需重复缓存
  } catch (error) {
    console.error('播放剧集失败:', error)
  }
}

async function playEpisode(ep: { name: string; url: string }, sourceName?: string) {
  if (!ep?.url) {
    return
  }

  // 清理预加载状态
  clearNextEpisodePreload()

  // 立刻更新当前剧集，用于后续解析播放链接与按钮状态
  currentPlayUrl.value = ep.url

  // 重置跳过片尾状态，新剧集可以重新触发
  skipOutroTriggered.value = false
  skipOutroCurrentUrl.value = ep.url
  skipOutroLastTriggerTime.value = 0 // 重置冷却时间
  // 更新地址栏中标题与来源，不修改 original_url 参数，避免影响回显
  const q = { 
    ...route.query, 
    title: ep.name, 
    source: sourceName || (ep as any).__sourceName,
    // 保持 original_url 不变，避免触发路由监听器
    original_url: route.query.original_url || route.query.url
  }
  router.replace({ name: 'watch', params: route.params, query: q })
  
  // 立即缓存当前播放的剧集信息
  savePlayState({ url: ep.url, title: ep.name, source: q.source })
  
  try {
    // 先尝试从缓存加载播放链接
    let url = loadPlayUrlCache(ep.url)

    if (!url) {
      // 缓存未命中，请求新的播放链接
      const token = auth.token!
      const res: any = await videoAPI.playUrl(token, sourceId.value, ep.url)
      url = res?.data?.video_url || res?.data || ''
      if (!url) return

      // 缓存播放链接
      savePlayUrlCache(ep.url, url)
    }

    if (!url || typeof url !== 'string') {
      console.error('[playEpisode] 获取播放链接失败')
      return
    }

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
      if (plyr) {
        try {
          plyr.speed = rate.value
        } catch {}
      } else if (videoRef.value) {
        try {
          videoRef.value.playbackRate = rate.value
        } catch {}
      }
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

      // 滚动到视频播放器
      await nextTick()
      if (videoRef.value) {
        videoRef.value.scrollIntoView({
          behavior: 'smooth',
          block: 'center'
        })
      }
    }
    
    // 播放成功后，预加载下一集的播放链接
    setTimeout(() => {
      preloadNextEpisodeUrl()
    }, 1000) // 延迟1秒执行，避免影响当前播放
    
    // 注：剧集信息已在函数开始时缓存，此处无需重复缓存
  } catch (error) {
    console.error('播放剧集失败:', error)
  }
}

function playPrev() {
  if (!canPrev.value) return
  playEpisode(currentSourceEpisodes.value[currentIndex.value - 1])
}
// 预加载下一集播放链接（优化版本，用于同时加载）
async function preloadNextEpisodeUrl() {
  if (!canNext.value) return

  try {
    const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
    if (!nextEpisode?.url) return

    console.log('[PreloadNext] 开始预加载下一集播放链接:', nextEpisode.url)

    // 先尝试从缓存加载播放链接
    let url = loadPlayUrlCache(nextEpisode.url)

    if (!url) {
      // 缓存未命中，请求新的播放链接
      console.log('[PreloadNext] 缓存未命中，请求下一集播放链接:', nextEpisode.url)
      const token = auth.token!
      const res: any = await videoAPI.playUrl(token, sourceId.value, nextEpisode.url)
      url = res?.data?.video_url || res?.data || ''

      if (url) {
        // 缓存播放链接
        savePlayUrlCache(nextEpisode.url, url)
        console.log('[PreloadNext] 下一集播放链接已缓存:', nextEpisode.url)
      }
    } else {
      console.log('[PreloadNext] 使用缓存的下一集播放链接:', nextEpisode.url)
    }

    if (url) {
      nextEpisodeUrl.value = url
      nextEpisodePreloaded.value = true
      console.log('[PreloadNext] 下一集播放链接预加载成功')
    }
  } catch (error) {
    console.error('[PreloadNext] 预加载下一集失败:', error)
  }
}

// 预加载下一集播放链接（原有版本，保持兼容性）
async function preloadNextEpisode() {
  if (!canNext.value || nextEpisodePreloaded.value) return

  try {
    const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
    if (!nextEpisode?.url) return

    console.log('开始预加载下一集播放链接:', nextEpisode.url)

    // 先尝试从缓存加载播放链接
    let url = loadPlayUrlCache(nextEpisode.url)

    if (!url) {
      // 缓存未命中，请求新的播放链接
      console.log('缓存未命中，请求下一集播放链接:', nextEpisode.url)
      const token = auth.token!
      const res: any = await videoAPI.playUrl(token, sourceId.value, nextEpisode.url)
      url = res?.data?.video_url || res?.data || ''

      if (url) {
        // 缓存播放链接
        savePlayUrlCache(nextEpisode.url, url)
      }
    } else {
      console.log('使用缓存的下一集播放链接:', nextEpisode.url)
    }

    if (url) {
      nextEpisodeUrl.value = url
      nextEpisodePreloaded.value = true
      console.log('下一集播放链接预加载成功')
    }
  } catch (error) {
    console.error('预加载下一集失败:', error)
  }
}

// 清理下一集预加载状态
function clearNextEpisodePreload() {
  nextEpisodeUrl.value = ''
  nextEpisodePreloaded.value = false
  nextEpisodeToastShown.value = false

  if (nextEpisodePreloadTimer.value) {
    clearTimeout(nextEpisodePreloadTimer.value)
    nextEpisodePreloadTimer.value = null
  }

  if (nextEpisodeToastTimer.value) {
    clearTimeout(nextEpisodeToastTimer.value)
    nextEpisodeToastTimer.value = null
  }
}

function playNext() {
  if (!canNext.value) {
    return
  }

  // 清理预加载状态
  clearNextEpisodePreload()

  // 如果有预加载的URL，直接使用
  if (nextEpisodeUrl.value) {
    const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
    playEpisodeWithUrl(nextEpisode, nextEpisodeUrl.value)
    // 播放下一集后，立即预加载下下集的播放链接
    setTimeout(() => {
      preloadNextEpisodeUrl()
    }, 1000) // 延迟1秒执行，避免影响当前播放
  } else {
    const nextEpisode = currentSourceEpisodes.value[currentIndex.value + 1]
    playEpisode(nextEpisode)
  }
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

  // 网速监控逻辑（移除loading相关代码）
}

function stopSpeedMonitoring() {
  if (speedCheckInterval) {
    clearInterval(speedCheckInterval)
    speedCheckInterval = null
  }
  networkSpeed.value = ''
}



// 震动反馈函数
function vibrateFeedback() {
  try {
    // 检查是否支持震动API
    if ('vibrate' in navigator) {
      // 震动两次：每次50ms，间隔100ms
      navigator.vibrate([50, 100, 50])
    }
  } catch (e) {
    console.log('[Vibrate] 震动功能不可用:', e)
  }
}

// 倍速监听函数
function startRateMonitoring() {
  if (rateCheckInterval) {
    clearInterval(rateCheckInterval)
  }

  lastVideoRate = 1
  rateCheckInterval = setInterval(() => {
    if (!videoRef.value) return

    const currentRate = videoRef.value.playbackRate || 1

    // 如果倍速发生变化且不是长按状态，同步到页面
    if (Math.abs(currentRate - lastVideoRate) > 0.01 && !isLongPressActive.value) {
      lastVideoRate = currentRate
      // 找到最接近的倍速选项
      const closestRate = rates.find(r => Math.abs(r - currentRate) < 0.1) || 1
      if (Math.abs(closestRate - rate.value) > 0.01) {
        syncRateToUI(closestRate)
        console.log('[RateMonitor] 检测到倍速变化，同步到页面:', currentRate, '->', closestRate)
      }
    }
  }, 500) // 每500ms检查一次
}

function stopRateMonitoring() {
  if (rateCheckInterval) {
    clearInterval(rateCheckInterval)
    rateCheckInterval = null
  }
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
    // 设置为2倍速并同步到UI
    syncRateToUI(2)
    isLongPressActive.value = true

    // 应用2倍速
    try {
      if (plyr) {
        plyr.speed = 2
      } else if (videoRef.value) {
        (videoRef.value as any).playbackRate = 2
      }
      // 长按期间强制隐藏进度条（Plyr 与 原生 video 均支持）
      try {
        const container = plyr
          ? plyr.elements.container
          : (videoRef.value ? videoRef.value.parentElement : null)
        if (container) {
          console.log('[LongPress] 强制隐藏进度条')
          console.log('[LongPress] 隐藏前容器类名:', container.className)
          container.classList.add('longpress-hide-progress')
          container.classList.remove('show-progress-bar')
          
          // 使用Plyr API强制隐藏控件
          if (plyr) {
            // 使用Plyr的内置方法隐藏控件
            if (typeof plyr.hideControls === 'function') {
              plyr.hideControls()
              console.log('[LongPress] 使用Plyr内置hideControls方法')
            } else {
              plyr.elements.container.classList.add('plyr--hide-controls')
              // 直接操作Plyr控件元素
              if (plyr.elements.controls) {
                plyr.elements.controls.style.setProperty('display', 'none', 'important')
                plyr.elements.controls.style.setProperty('opacity', '0', 'important')
                plyr.elements.controls.style.setProperty('visibility', 'hidden', 'important')
                plyr.elements.controls.style.setProperty('pointer-events', 'none', 'important')
                console.log('[LongPress] 强制隐藏控件1', plyr.elements.controls, plyr.elements.controls.style)
              }
            }
            
            console.log('[LongPress] 使用Plyr API隐藏控件')
          }
          
          isProgressVisible.value = false
          console.log('[LongPress] 隐藏后容器类名:', container.className)
        }
      } catch (e) {
        console.error('[LongPress] 隐藏进度条失败:', e)
      }
      console.log('[LongPress] 启动2倍速播放')
      // 震动反馈
      vibrateFeedback()
      // 显示toast提示
      console.log('[LongPress] 显示Toast: 已启动2倍速播放')
      message.info('已启动2倍速播放', 3)
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
    // 恢复原始播放速率并同步到UI
    syncRateToUI(originalRate.value)
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
          if (container) {
            container.classList.remove('longpress-hide-progress')
            
            // 使用Plyr API恢复控件显示
            if (plyr) {
              // 直接操作Plyr控件元素 - 强制设置显示样式
              if (plyr.elements.controls) {
                // 强制设置显示样式，覆盖Plyr的隐藏逻辑
                plyr.elements.controls.style.setProperty('display', 'flex', 'important')
                plyr.elements.controls.style.setProperty('opacity', '1', 'important')
                plyr.elements.controls.style.setProperty('visibility', 'visible', 'important')
                plyr.elements.controls.style.setProperty('pointer-events', 'auto', 'important')
                
                console.log('[LongPress] 强制显示控件', plyr.elements.controls, plyr.elements.controls.style)
              }
              
              // 使用Plyr的内置方法显示控件
              if (typeof plyr.showControls === 'function') {
                plyr.showControls()
                console.log('[LongPress] 使用Plyr内置showControls方法')
              } else {
                plyr.elements.container.classList.remove('plyr--hide-controls')
                console.log('[LongPress] 使用Plyr CSS类恢复控件显示')
              }
              
              console.log('[LongPress] 使用Plyr API恢复控件显示')
            }
            
            // 如果之前是手动显示的进度条，恢复显示状态
            if (isProgressVisible.value) {
              container.classList.add('show-progress-bar')
              if (plyr) {
                plyr.elements.container.classList.add('plyr--controls-active')
              }
            }
          }
        } catch (e) {
          console.error('[LongPress] 恢复进度条失败:', e)
        }
      console.log('[LongPress] 恢复原始播放速率:', originalRate.value)
      // 震动反馈
      vibrateFeedback()
      // 显示toast提示
      console.log('[LongPress] 显示Toast: 已恢复倍速播放')
      message.info(`已恢复${originalRate.value}x倍速播放`, 3)
    } catch (e) {
      console.error('[LongPress] 恢复原始速率失败:', e)
    }
  }
}

// 原位置已前移，避免 TDZ；此处删除重复定义

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
  updatedAt?: number
}
function savePlayState(partial: PlayState) {
  try {
    const raw = localStorage.getItem(playStateKey.value)
    const prev: PlayState = raw ? JSON.parse(raw) : {}
    const merged: PlayState = { ...prev, ...partial, updatedAt: Date.now() }
    localStorage.setItem(playStateKey.value, JSON.stringify(merged))
    try { console.log('[PlayState] 已保存播放状态:', merged) } catch {}

    // 管理缓存数量
    managePlayStateCache()
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

// 删除当前剧集的播放进度缓存
function deletePlayStateCache() {
  try {
    localStorage.removeItem(playStateKey.value)
    console.log(`[PlayState] 已删除播放进度缓存: ${playStateKey.value}`)
  } catch (e) {
    console.error('[PlayState] 删除播放进度缓存失败:', e)
  }
}

// 管理播放进度缓存数量，只保留最新的50条
function managePlayStateCache() {
  try {
    const allKeys = Object.keys(localStorage)
    const playStateKeys = allKeys.filter(key => key.startsWith('watch_state:'))

    if (playStateKeys.length > 50) {
      // 按时间戳排序，删除最早的缓存
      const cacheItems = playStateKeys.map(key => {
        try {
          const raw = localStorage.getItem(key)
          if (raw) {
            const parsed = JSON.parse(raw)
            return { key, updatedAt: parsed.updatedAt || 0 }
          }
        } catch {}
        return { key, updatedAt: 0 }
      }).filter(item => item.updatedAt > 0)

      // 按时间戳排序，最早的在前
      cacheItems.sort((a, b) => a.updatedAt - b.updatedAt)

      // 删除最早的缓存，保留50条
      const toDelete = cacheItems.slice(0, cacheItems.length - 50)
      toDelete.forEach(item => {
        localStorage.removeItem(item.key)
        console.log(`[PlayState] 清理过期缓存: ${item.key}`)
      })

      console.log(`[PlayState] 缓存管理完成，删除了 ${toDelete.length} 条过期缓存`)
    }
  } catch (e) {
    console.error('[PlayState] 缓存管理失败:', e)
  }
}

// 播放链接缓存相关函数
type PlayUrlCache = {
  url: string
  episodeUrl: string
  sourceId: string
  cachedAt: number
  expiresAt: number
}

function savePlayUrlCache(episodeUrl: string, playUrl: string) {
  try {
    const cache: PlayUrlCache = {
      url: playUrl,
      episodeUrl: episodeUrl,
      sourceId: sourceId.value,
      cachedAt: Date.now(),
      expiresAt: Date.now() + 30 * 60 * 1000 // 30分钟过期
    }
    const cacheKey = getPlayUrlCacheKey(episodeUrl)
    localStorage.setItem(cacheKey, JSON.stringify(cache))
    console.log(`[PlayUrlCache] 已缓存播放链接: ${episodeUrl} -> ${playUrl}`)
  } catch (e) {
    console.error('[PlayUrlCache] 保存播放链接缓存失败:', e)
  }
}

function loadPlayUrlCache(episodeUrl: string): string | null {
  try {
    const cacheKey = getPlayUrlCacheKey(episodeUrl)
    const raw = localStorage.getItem(cacheKey)
    if (!raw) {
      console.log(`[PlayUrlCache] 未发现播放链接缓存: ${episodeUrl}`)
      return null
    }

    const cache: PlayUrlCache = JSON.parse(raw)

    // 检查缓存是否过期
    if (Date.now() > cache.expiresAt) {
      console.log(`[PlayUrlCache] 播放链接缓存已过期: ${episodeUrl}`)
      localStorage.removeItem(cacheKey)
      return null
    }

    // 检查缓存是否匹配当前剧集
    if (cache.episodeUrl !== episodeUrl || cache.sourceId !== sourceId.value) {
      console.log(`[PlayUrlCache] 播放链接缓存不匹配: ${episodeUrl}`)
      return null
    }

    console.log(`[PlayUrlCache] 命中播放链接缓存: ${episodeUrl} -> ${cache.url}`)
    return cache.url
  } catch (e) {
    console.error('[PlayUrlCache] 读取播放链接缓存失败:', e)
    return null
  }
}

async function fetchDetail(force = false) {
  error.value = ''
  if (!force) {
    const hit = loadCache()
    fromCache.value = hit
    if (hit) {
      // 命中缓存也要解析播放地址
      await resolvePlayUrl()
      // 如果基础信息缺失，触发一次强制刷新
      if (!base.value.name) {
        try { console.log('[Detail] 缓存数据基础信息缺失，强制刷新') } catch {}
        await fetchDetail(true)
        return
      }
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

    // 记录观看历史（无需登录模式下使用本地缓存）
    if (!configStore.needsLogin() && detailData.value) {
      const videoTitle = displayTitle.value || '未知标题'
      const videoId = `${sourceId.value}|${videoUrl.value}`

      localHistoryManager.addVideoHistory(
        videoId,
        videoTitle,
        videoUrl.value,
        sourceId.value,
        currentSourceName.value || '未知站点'
      )
    }

    await resolvePlayUrl() // detail 成功后拉取真实播放链接
  } catch (e: any) {
    error.value = e?.message || '获取详情失败'
  } finally {
    loading.value = false
  }
}

async function resolvePlayUrl() {
  try {
    // 重置跳过片尾状态，新播放源可以重新触发
    skipOutroTriggered.value = false
    skipOutroCurrentUrl.value = currentPlayUrl.value
    skipOutroLastTriggerTime.value = 0 // 重置冷却时间
    // 优先请求"当前选中剧集"的播放链接；无则回退
    const episodeUrl = getSelectedEpisodeUrl()

    // 先尝试从缓存加载播放链接
    let url = loadPlayUrlCache(episodeUrl)

    if (!url) {
      // 缓存未命中，请求新的播放链接
      console.log(`[resolvePlayUrl] 缓存未命中，请求播放链接: ${episodeUrl}`)
      const token = auth.token!
      const res: any = await videoAPI.playUrl(token, sourceId.value, episodeUrl)
      url = res?.data?.video_url || res?.data || ''

      if (url) {
        // 缓存播放链接
        savePlayUrlCache(episodeUrl, url)
      }
    } else {
      console.log(`[resolvePlayUrl] 使用缓存的播放链接: ${episodeUrl}`)
    }

    // 同时预加载下一集的播放链接
    await preloadNextEpisodeUrl()

    if (!url || typeof url !== 'string') {
      console.error('[resolvePlayUrl] 获取播放链接失败')
      return
    }

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
        // 设置倍速
        if (plyr) {
          try {
            plyr.speed = rate.value
            console.log('[HLS] set plyr speed to', rate.value)
          } catch {}
        } else if (videoRef.value) {
          try {
            videoRef.value.playbackRate = rate.value
            console.log('[HLS] set video playbackRate to', rate.value)
          } catch {}
        }

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

// 搜索其他站点
async function searchOtherSites() {
  otherSitesModalVisible.value = true
  hasSearchedOtherSites.value = false
  otherSitesResults.value = []

  // 立即执行搜索
  await handleSearchOtherSites()
}

// 处理其他站点搜索
async function handleSearchOtherSites() {
  // 使用当前视频标题作为搜索关键词
  const keyword = base.value.name || displayTitle.value || ''
  if (!keyword.trim()) {
    message.warning('无法获取视频标题')
    return
  }

  // 加载设置
  await settingsStore.loadSettings()

  // 检查是否有启用的搜索网站
  if (!settingsStore.hasEnabledSites && !settingsStore.settings.allSitesSelected) {
    message.warning('请先在设置中选择要搜索的网站')
    return
  }

  searchingOtherSites.value = true
  hasSearchedOtherSites.value = true
  otherSitesResults.value = []

  try {
    // 1) 拉取站点列表并按 sort 降序（越大越靠前）
    const token = auth.token!
    const listResp: any = await videoSourceAPI.getVideoSourceList(token)
    const allSources: Array<{ id: string; name: string; sort: number; status: number }> = (listResp?.data || [])
      // 仅搜索正常状态的站点（status=1）
      .filter((s: any) => Number(s.status) === 1)
      // 排除当前站点（通过URL中的资源ID过滤）
      .filter((s: any) => s.id !== sourceId.value)
      // 按 sort 降序
      .sort((a: any, b: any) => b.sort - a.sort)

    // 2) 根据设置过滤启用的网站
    let sources: any[] = []

    if (settingsStore.settings.allSitesSelected) {
      // 全选模式：搜索所有正常状态的站点
      sources = allSources
    } else {
      // 手动选择模式：只搜索用户勾选的站点
      const enabledSiteIds = settingsStore.enabledSearchSites.map(site => site.id)
      sources = allSources.filter(source => {
        // 如果设置中的网站ID是数字格式，需要转换比较
        const sourceIdStr = String(source.id)
        return enabledSiteIds.some(enabledId => {
          // 支持多种ID匹配方式
          return enabledId === sourceIdStr ||
                 enabledId === String(source.id) ||
                 enabledId === source.id
        })
      })
    }

    if (sources.length === 0) {
      message.warning('没有找到启用的搜索网站，请检查设置')
      otherSitesResults.value = []
      return
    }

    // 3) 并发度=2 线程池，按顺序调度
    const concurrency = 2
    const queue = [...sources]
    const results: any[] = []

    const runner = async () => {
      while (queue.length > 0) {
        const source = queue.shift()
        if (!source) break

        try {
          const searchResp: any = await videoAPI.search(token, source.id, keyword)
          if (searchResp?.code === 0 && searchResp?.data) {
            const videos = Array.isArray(searchResp.data) ? searchResp.data : [searchResp.data]
            videos.forEach((video: any) => {
              results.push({
                ...video,
                sourceId: source.id,
                sourceName: source.name
              })
            })
          }
        } catch (error) {
          console.error(`搜索站点 ${source.name} 失败:`, error)
        }
      }
    }

    // 启动并发搜索
    const runners = Array(concurrency).fill(null).map(() => runner())
    await Promise.all(runners)

    otherSitesResults.value = results

    const searchedSitesCount = sources.length
    const totalSitesCount = allSources.length
    const modeText = settingsStore.settings.allSitesSelected ? '全选模式' : '手动选择模式'
    message.success(`搜索完成，在 ${searchedSitesCount}/${totalSitesCount} 个站点中找到 ${results.length} 个结果 (${modeText})`)
  } catch (error: any) {
    message.error(error?.message || '搜索失败')
  } finally {
    searchingOtherSites.value = false
  }
}

// 从其他站点播放视频
async function playFromOtherSite(result: any) {
  try {
    const token = auth.token!
    const detailResp: any = await videoAPI.detail(token, result.sourceId, result.url)

    if (detailResp?.code === 0 && detailResp?.data) {
      // 关闭弹窗
      otherSitesModalVisible.value = false

      // 跳转到播放页面
      await router.push({
        name: 'watch',
        params: { sourceId: result.sourceId },
        query: {
          url: result.url,
          title: result.name || result.title,
          original_url: result.url
        }
      })

      // 强制刷新页面数据
      await initializePage()
    } else {
      message.error('获取视频详情失败')
    }
  } catch (error: any) {
    message.error(error?.message || '播放失败')
  }
}

// 处理图片加载错误
function handleImageError(event: Event) {
  const img = event.target as HTMLImageElement
  img.src = '/favicon.ico'
}





// 初始化页面数据
async function initializePage() {
  // 加载设置
  await loadSettings()

  // 加载全局倍速设置
  loadGlobalRate()
  
  // 加载跳过片首片尾设置（根据 original_url 隔离）
  loadSkipSettings()

  // 获取当前站点名称
  await updateCurrentSourceName()

  await fetchDetail(false)
  
  // 根据 URL 参数确定要播放的剧集
  const titleParam = String(route.query.title || '')
  const sourceParam = String((route.query as any).source || '')
  
  // 逻辑2: 如果有 title 和 source 参数，先选中对应的 source tab
  if (titleParam && sourceParam) {
    const targetSourceIndex = sourcesByTab.value.findIndex(s => s.name === sourceParam)
    if (targetSourceIndex >= 0) {
      activeSourceKey.value = String(targetSourceIndex)
      try { console.log('[Init] 根据URL参数选中Source tab:', sourceParam, 'index:', targetSourceIndex) } catch {}
    }
  }
  
  // 逻辑3: 如果没有 title 和 source 参数，根据缓存选中对应的 source tab
  if (!titleParam && !sourceParam) {
    const state = loadPlayState()
    if (state?.source) {
      const cachedSourceIndex = sourcesByTab.value.findIndex(s => s.name === state.source)
      if (cachedSourceIndex >= 0) {
        activeSourceKey.value = String(cachedSourceIndex)
        try { console.log('[Init] 根据缓存选中Source tab:', state.source, 'index:', cachedSourceIndex) } catch {}
      }
    }
  }
  
  // 优先使用 getSelectedEpisodeUrl 获取要播放的剧集
  const selectedUrl = getSelectedEpisodeUrl()
  if (selectedUrl) {
    const ep = flatEpisodes.value.find(e => e.url === selectedUrl)
    if (ep) {
      await playEpisode(ep)
    } else {
      // 如果找不到对应的剧集，使用第一个剧集
      const first = flatEpisodes.value[0]
      if (first) await playEpisode(first)
    }
  } else {
    // 如果没有找到要播放的剧集，使用缓存
    const state = loadPlayState()
    if (state?.url) {
      const ep = flatEpisodes.value.find(e => e.url === state.url) || flatEpisodes.value[0]
      if (ep) await playEpisode(ep)
    } else {
      // 最后兜底使用第一个剧集
      const first = flatEpisodes.value[0]
      if (first) await playEpisode(first)
    }
  }
  
  // 根据当前播放URL选中对应的来源 tab（只在没有URL参数时使用缓存）
  if (!titleParam && !sourceParam) {
    const cachedUrl = loadPlayState()?.url || videoUrl.value
    if (!currentPlayUrl.value) {
      currentPlayUrl.value = cachedUrl
    }
  }
  
  const idx = sourcesByTab.value.findIndex(s => s.episodes.some(e => e.url === currentPlayUrl.value))
  if (idx >= 0) activeSourceKey.value = String(idx)
}

onMounted(async () => {
  // 初始化移动设备检测
  checkMobile()
  window.addEventListener('resize', checkMobile)

  // 预加载所有站点名称
  await preloadAllSourceNames()

  await initializePage()
})

// 监听路由变化，确保页面内容刷新
let lastRouteSnapshot: any = null
watch(
  () => [route.params.sourceId, route.query.original_url],
  async (newVals, oldVals) => {
    try {
      const [newSourceId, newOriginalUrl] = newVals || []
      const [oldSourceId, oldOriginalUrl] = oldVals || []
      const sourceIdChanged = String(newSourceId ?? '') !== String(oldSourceId ?? '')
      const originalUrlChanged = String(newOriginalUrl ?? '') !== String(oldOriginalUrl ?? '')
      const reasonParts: string[] = []
      if (sourceIdChanged) reasonParts.push(`sourceId changed: ${oldSourceId} -> ${newSourceId}`)
      if (originalUrlChanged) reasonParts.push(`original_url changed: ${oldOriginalUrl} -> ${newOriginalUrl}`)
      // 进一步打印 query 的差异（包括 url/title/source 等）
      const currentSnapshot = {
        sourceId: String(route.params.sourceId ?? ''),
        original_url: String(route.query.original_url ?? ''),
        url: String(route.query.url ?? ''),
        title: String(route.query.title ?? ''),
        source: String((route.query as any).source ?? ''),
        fullPath: String(route.fullPath ?? window.location.pathname + window.location.search)
      }
      let queryDiff = ''
      if (lastRouteSnapshot) {
        const keys = ['original_url','url','title','source','fullPath']
        const diffs: string[] = []
        for (const k of keys) {
          if (String(currentSnapshot[k]) !== String(lastRouteSnapshot[k])) {
            diffs.push(`${k} changed: ${lastRouteSnapshot[k]} -> ${currentSnapshot[k]}`)
          }
        }
        queryDiff = diffs.join(' | ')
      } else {
        queryDiff = 'first run'
      }
      const reason = [reasonParts.join(' | '), queryDiff].filter(Boolean).join(' || ')
      console.log('[Watch reload]', window.location.href, '| reason:', reason)
      lastRouteSnapshot = currentSnapshot
      // 如果 sourceId 与 original_url 都未变化，则忽略此次触发，避免无意义的页面重载
      if (!sourceIdChanged && !originalUrlChanged) {
        console.log('[Watch reload ignored] sourceId/original_url 未变化，仅 query 辅助字段变更')
        return
      }
    } catch {}
    // 重置状态
    loading.value = false
    error.value = ''
    detailData.value = null
    fromCache.value = false
    currentPlayUrl.value = ''
    playerSource.value = ''
    currentSourceName.value = '' // 重置站点名称

    // 重新初始化页面
    await initializePage()
  },
  { immediate: false }
)

// 清理事件监听
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  stopSpeedMonitoring() // 清理网速监控定时器
  stopRateMonitoring() // 清理倍速监听定时器

  // 清理长按定时器
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }

  // 清理跳过片尾状态
  skipOutroTriggered.value = false
  skipOutroCurrentUrl.value = ''
  skipOutroLastTriggerTime.value = 0 // 重置冷却时间

  // 清理下一集预加载状态
  clearNextEpisodePreload()

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
      // 强制显示进度条（包括全屏状态）
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
    
    // 新增：边缘区域检查，避免误触
    containerRect = container.getBoundingClientRect()
    const touch = e.touches[0]
    const x = touch.clientX
    const y = touch.clientY
    
    // 检查是否在屏幕边缘（左右各15%区域）
    const edgeThreshold = containerRect.width * 0.15
    if (x <= edgeThreshold || x >= containerRect.right - edgeThreshold) {
      console.log('[ProgressDrag] 触摸在边缘区域，不触发进度调节')
      return
    }
    
    // 检查是否在顶部/底部边缘区域
    const topGuard = containerRect.top + containerRect.height * screenEdgeGuardRatio
    const bottomGuard = containerRect.bottom - containerRect.height * screenEdgeGuardRatio
    if (y <= topGuard || y >= bottomGuard) {
      console.log('[ProgressDrag] 触摸在顶部/底部边缘区域，不触发进度调节')
      determined = false
      isHorizontal = false
      isDraggingProgress = false
      return
    }
    
    startX = touch.clientX
    startY = touch.clientY
    startTime = getCurrentTime()
    determined = false
    isHorizontal = false
    
    // 设置拖动状态
    isDraggingProgress = false
    isHorizontalDrag.value = false
    isVerticalDrag.value = false
    
    // 强制显示进度条
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
        isVerticalDrag.value = true
        try {
          container.classList.remove('dragging-show-progress')
          container.classList.remove('plyr--controls-active')
        } catch {}
        try { if (keepAliveTimer) { clearInterval(keepAliveTimer); keepAliveTimer = null } } catch {}
        console.log('[ProgressDrag] 检测到垂直滑动，取消进度调节')
        return
      }
      if (Math.abs(dx) > 8) {
        determined = true
        isHorizontal = true
        isHorizontalDrag.value = true
        isDraggingProgress = true
        // 进入进度拖动：强制显示进度条，且取消长按隐藏
        try {
          container.classList.add('dragging-show-progress')
          container.classList.remove('longpress-hide-progress')
          container.classList.remove('plyr--hide-controls')
          container.classList.add('plyr--controls-active')
        } catch {}
        
        // 滑动进度期间，每3秒显示一次进度条
        console.log('[PROGRESS_CTRL] [ProgressDrag] 启动滑动进度显示进度条定时器')
        // 立即执行一次显示进度条
        console.log('[PROGRESS_CTRL] [ProgressDrag] 立即执行：显示进度条')
        showProgressBar()
        // 启动定时器，每3秒显示一次
        dragShowTimer = setInterval(() => {
          console.log('[PROGRESS_CTRL] [ProgressDrag] 定时器触发：显示进度条')
          showProgressBar()
        }, 3000)
      }
    }

    if (isHorizontal && !verticalCancel) {
      // 阻止长按倍速与点击
      e.preventDefault()
      e.stopPropagation()
      // 每次 move 都保证控件与进度条可见，避免任何闪烁
      ensureProgressVisible()
      // 按容器宽度映射到时长，应用进度条敏感度设置
      const duration = getDuration()
      if (!duration || duration <= 0) return
      const w = containerRect.width || 1
      const timePerPixel = duration / w
      const nt = startTime + dx * timePerPixel * progressSensitivity.value
      setCurrentTime(nt)
    }
  }

  const onTouchEnd = (_e: TouchEvent) => {
    determined = false
    isHorizontal = false
    isDraggingProgress = false
    
    // 清理滑动进度显示进度条定时器
    if (dragShowTimer) {
      console.log('[PROGRESS_CTRL] [ProgressDrag] 滑动结束：清理滑动进度显示进度条定时器')
      clearInterval(dragShowTimer)
      dragShowTimer = null
    }
    
    try {
      container.classList.remove('dragging-show-progress')
      container.classList.remove('plyr--controls-active')
    } catch {}
    try { if (keepAliveTimer) { clearInterval(keepAliveTimer); keepAliveTimer = null } } catch {}
    // 结束拖动后，给予控件短暂显示时间，避免立即被隐藏造成的闪烁
    try {
      container.classList.remove('plyr--hide-controls')
      // 延迟清理，让Plyr的自动隐藏机制正常工作
      setTimeout(() => {
        try {
          container.classList.remove('dragging-show-progress')
        } catch {}
      }, 120)
    } catch {}
  }

  // 触摸事件
  container.addEventListener('touchstart', onTouchStart, { passive: true })
  container.addEventListener('touchmove', onTouchMove, { passive: false })
  container.addEventListener('touchend', onTouchEnd, { passive: true })
  container.addEventListener('touchcancel', onTouchEnd, { passive: true })
}

// 设置相关变量
const longPressSpeed = ref(2.0)
const progressSensitivity = ref(0.7)
const isLongPressing = ref(false)
const originalPlaybackRate = ref(1.0)

// 加载设置
const loadSettings = async () => {
  await settingsStore.loadSettings()
  longPressSpeed.value = settingsStore.settings.longPressPlaybackSpeed
  progressSensitivity.value = settingsStore.settings.progressBarSensitivity
}

// 新增：视频播放控制功能
// 隐藏进度条方法
function hideProgressBar() {
  try {
    console.log('[PROGRESS_CTRL] [HideProgressBar] 开始隐藏进度条')
    let container: HTMLElement | null = null
    
    if (plyr) {
      container = plyr.elements.container
      console.log('[PROGRESS_CTRL] [HideProgressBar] 使用Plyr容器')
    } else if (videoRef.value) {
      container = videoRef.value.parentElement
      console.log('[PROGRESS_CTRL] [HideProgressBar] 使用原生video容器')
    }
    
    if (!container) {
      console.log('[PROGRESS_CTRL] [HideProgressBar] 容器不存在')
      return
    }
    
    // 隐藏进度条样式
    container.classList.remove('show-progress-bar')
    container.classList.add('hide-progress-bar')
    
    // 使用Plyr API强制隐藏控件
    if (plyr) {
      try {
        if (typeof plyr.hideControls === 'function') {
          plyr.hideControls()
          console.log('[PROGRESS_CTRL] [HideProgressBar] 使用Plyr内置hideControls方法')
        } else {
          plyr.elements.container.classList.add('plyr--hide-controls')
          if (plyr.elements.controls) {
            plyr.elements.controls.style.setProperty('display', 'none', 'important')
            plyr.elements.controls.style.setProperty('opacity', '0', 'important')
            plyr.elements.controls.style.setProperty('visibility', 'hidden', 'important')
            plyr.elements.controls.style.setProperty('pointer-events', 'none', 'important')
            console.log('[PROGRESS_CTRL] [HideProgressBar] 强制隐藏Plyr控件')
          }
        }
      } catch (e) {
        console.error('[PROGRESS_CTRL] [HideProgressBar] Plyr API隐藏控件失败:', e)
      }
    }
    
    isProgressVisible.value = false
    console.log('[PROGRESS_CTRL] [HideProgressBar] 进度条已隐藏')
  } catch (e) {
    console.error('[PROGRESS_CTRL] [HideProgressBar] 隐藏进度条失败:', e)
  }
}

// 显示进度条方法
function showProgressBar() {
  try {
    console.log('[PROGRESS_CTRL] [ShowProgressBar] 开始显示进度条')
    let container: HTMLElement | null = null
    
    if (plyr) {
      container = plyr.elements.container
      console.log('[PROGRESS_CTRL] [ShowProgressBar] 使用Plyr容器')
    } else if (videoRef.value) {
      container = videoRef.value.parentElement
      console.log('[PROGRESS_CTRL] [ShowProgressBar] 使用原生video容器')
    }
    
    if (!container) {
      console.log('[PROGRESS_CTRL] [ShowProgressBar] 容器不存在')
      return
    }
    
    // 显示进度条样式
    container.classList.add('show-progress-bar')
    container.classList.remove('hide-progress-bar')
    
    // 使用Plyr API强制显示控件
    if (plyr) {
      try {
        if (plyr.elements.controls) {
          plyr.elements.controls.style.setProperty('display', 'flex', 'important')
          plyr.elements.controls.style.setProperty('opacity', '1', 'important')
          plyr.elements.controls.style.setProperty('visibility', 'visible', 'important')
          plyr.elements.controls.style.setProperty('pointer-events', 'auto', 'important')
          console.log('[PROGRESS_CTRL] [ShowProgressBar] 强制显示Plyr控件')
        }
        
        if (typeof plyr.showControls === 'function') {
          plyr.showControls()
          console.log('[PROGRESS_CTRL] [ShowProgressBar] 使用Plyr内置showControls方法')
        } else {
          plyr.elements.container.classList.remove('plyr--hide-controls')
          plyr.elements.container.classList.add('plyr--controls-active')
          console.log('[PROGRESS_CTRL] [ShowProgressBar] 使用Plyr CSS类显示控件')
        }
      } catch (e) {
        console.error('[PROGRESS_CTRL] [ShowProgressBar] Plyr API显示控件失败:', e)
      }
    }
    
    isProgressVisible.value = true
    console.log('[PROGRESS_CTRL] [ShowProgressBar] 进度条已显示')
  } catch (e) {
    console.error('[PROGRESS_CTRL] [ShowProgressBar] 显示进度条失败:', e)
  }
}

// 单击切换进度条显示/隐藏
function toggleProgressBar() {
  console.log('[PROGRESS_CTRL] [ToggleProgressBar] 单击触发进度条切换')
  if (isProgressVisible.value) {
    console.log('[PROGRESS_CTRL] [ToggleProgressBar] 当前进度条可见，执行隐藏')
    hideProgressBar()
  } else {
    console.log('[PROGRESS_CTRL] [ToggleProgressBar] 当前进度条隐藏，执行显示')
    showProgressBar()
  }
}



// 进度拖动手势（同时适配 Plyr 和原生 video 容器）
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
  margin-bottom: 16px;
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

.longpress-hide-progress :deep(.plyr__progress),
.longpress-hide-progress :deep(.plyr__controls),
.longpress-hide-progress :deep(.plyr__control),
.longpress-hide-progress :deep(.plyr__time),
.longpress-hide-progress :deep(.plyr__menu) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
}

/* 长按倍速时强制隐藏进度条 */
.longpress-hide-progress.plyr :deep(.plyr__controls) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
  transform: translateY(100%) !important;
}

/* 进度拖动时，强制显示 Plyr 进度条 */
.dragging-show-progress :deep(.plyr__progress) {
  display: block !important;
}

/* 全屏时优化控件显示 */
.player-wrap :deep(.plyr--fullscreen) {
  /* 全屏时允许Plyr的自动隐藏机制正常工作 */
}

.player-wrap :deep(.plyr--fullscreen .plyr__controls) {
  transition: opacity 0.3s ease;
}

/* 确保点击视频时能正常显示控件 */
.player-wrap :deep(.plyr__video-wrapper) {
  cursor: pointer;
}

/* 优化控件显示时机 */
.player-wrap :deep(.plyr__controls) {
  opacity: 1;
  transition: opacity 0.3s ease;
}

.player-wrap :deep(.plyr--hide-controls .plyr__controls) {
  opacity: 0;
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

/* 手动显示/隐藏进度条样式 */
.show-progress-bar :deep(.plyr__progress),
.show-progress-bar :deep(.plyr__controls),
.show-progress-bar :deep(.plyr__control),
.show-progress-bar :deep(.plyr__time),
.show-progress-bar :deep(.plyr__menu) {
  display: block !important;
  opacity: 1 !important;
  visibility: visible !important;
}

.hide-progress-bar :deep(.plyr__progress),
.hide-progress-bar :deep(.plyr__controls),
.hide-progress-bar :deep(.plyr__control),
.hide-progress-bar :deep(.plyr__time),
.hide-progress-bar :deep(.plyr__menu) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
}

/* 更强的进度条显示规则 */
.show-progress-bar.plyr :deep(.plyr__controls) {
  display: flex !important;
  opacity: 1 !important;
  visibility: visible !important;
  transform: none !important;
  pointer-events: auto !important;
}

.hide-progress-bar.plyr :deep(.plyr__controls) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
  transform: translateY(100%) !important;
  pointer-events: none !important;
}

/* 针对Plyr控件的更强规则 */
.hide-progress-bar :deep(.plyr__controls),
.hide-progress-bar :deep(.plyr__control),
.hide-progress-bar :deep(.plyr__progress),
.hide-progress-bar :deep(.plyr__time),
.hide-progress-bar :deep(.plyr__menu),
.hide-progress-bar :deep(.plyr__video-wrapper + .plyr__controls) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
  pointer-events: none !important;
  width: 0 !important;
  height: 0 !important;
  overflow: hidden !important;
}

.show-progress-bar :deep(.plyr__controls),
.show-progress-bar :deep(.plyr__control),
.show-progress-bar :deep(.plyr__progress),
.show-progress-bar :deep(.plyr__time),
.show-progress-bar :deep(.plyr__menu) {
  display: block !important;
  opacity: 1 !important;
  visibility: visible !important;
  pointer-events: auto !important;
}

/* 长按倍速时的更强隐藏规则 */
.longpress-hide-progress :deep(.plyr__controls),
.longpress-hide-progress :deep(.plyr__control),
.longpress-hide-progress :deep(.plyr__progress),
.longpress-hide-progress :deep(.plyr__time),
.longpress-hide-progress :deep(.plyr__menu),
.longpress-hide-progress :deep(.plyr__video-wrapper + .plyr__controls) {
  display: none !important;
  opacity: 0 !important;
  visibility: hidden !important;
  pointer-events: none !important;
  width: 0 !important;
  height: 0 !important;
  overflow: hidden !important;
}

/* 原生video进度条显示/隐藏 */
.show-progress-bar video::-webkit-media-controls-timeline,
.show-progress-bar video::-webkit-media-controls-current-time-display,
.show-progress-bar video::-webkit-media-controls-time-remaining-display {
  display: block !important;
}

.hide-progress-bar video::-webkit-media-controls-timeline,
.hide-progress-bar video::-webkit-media-controls-current-time-display,
.hide-progress-bar video::-webkit-media-controls-time-remaining-display {
  display: none !important;
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
.player-actions {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.player-actions-row {
  display: flex;
}

/* 其他站点搜索结果样式 */
.other-sites-search {
  padding: 16px 0;
}

.search-header {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.search-results {
  margin-top: 16px;
}

.results-header {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  max-height: 60vh;
  overflow-y: auto;
}

.result-card {
  display: flex;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 180px;
  overflow: hidden;
}

.result-card:hover {
  border-color: #1890ff;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.15);
  transform: translateY(-2px);
}

.card-cover {
  flex-shrink: 0;
  width: 80px;
  height: 100px;
  margin-right: 12px;
  border-radius: 4px;
  overflow: hidden;
}

.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
  overflow: hidden;
}

.card-title {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-source {
  margin: 0 0 4px 0;
  font-size: 12px;
  color: #666;
}

.card-info {
  margin: 0 0 4px 0;
  font-size: 12px;
  color: #666;
  line-height: 1.4;
}

.card-info p {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  line-height: 1.3;
}

.card-actor,
.card-director,
.card-date,
.card-region {
  margin: 0;
  font-size: 12px;
  color: #666;
}

.card-desc {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: #999;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
}

.rating {
  color: #fa8c16;
  font-weight: 500;
}

.type {
  color: #52c41a;
  background: #f6ffed;
  padding: 2px 6px;
  border-radius: 2px;
}

.no-results {
  text-align: center;
  padding: 40px 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .results-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .result-card {
    height: 160px;
    padding: 8px;
  }

  .card-cover {
    width: 60px;
    height: 80px;
    margin-right: 8px;
  }

  .card-title {
    font-size: 13px;
  }

  .card-desc {
    font-size: 11px;
  }
}

.skip-label {
  color: #666;
  font-size: 14px;
  margin-right: 8px;
  min-width: 60px;
}

.skip-unit {
  color: #666;
  font-size: 14px;
  margin-left: 4px;
}
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

.ep-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
  gap: 8px;
  min-width: 0;
}
.ep-btn {
  width: 100%;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  padding: 0 4px;
  box-sizing: border-box;
}

@media (max-width: 768px) {
  .card-header {
  flex-direction: column;
  align-items: flex-start;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  margin-top: 8px;
}

.source-info {
  display: flex;
  align-items: center;
}
  .kv-list { grid-template-columns: 1fr; }

  /* 移动端剧集列表优化 */
  .ep-list {
    grid-template-columns: repeat(auto-fit, minmax(60px, 1fr));
    gap: 6px;
  }

  .ep-btn {
    font-size: 12px;
    padding: 0 2px;
    height: 28px;
  }
  .card-header h2 { white-space: normal; font-size: 18px; }

  /* 移动端播放器控制区域优化 */
  .player-actions {
    flex-direction: column;
    gap: 8px;
  }

  .player-actions-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .player-actions .ant-space {
    flex-wrap: wrap;
    gap: 4px;
  }

  .player-actions .ant-space-item {
    margin-bottom: 4px;
  }

  .skip-label {
    min-width: auto;
    margin-bottom: 4px;
  }
}
</style>


