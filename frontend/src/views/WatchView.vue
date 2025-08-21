<template>
  <AppLayout :page-title="`播放 - ${displayTitle}`">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>视频详情</h2>
          <div class="header-actions">
            <a-space>
              <a-button type="primary" :loading="loading" @click="refreshDetail">重新获取</a-button>
              <a-button type="default" @click="goOriginal" :disabled="!originalUrl">原站点</a-button>
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
            <div class="player-wrap">
              <video
                ref="videoRef"
                class="plyr-video"
                :poster="basePoster"
                playsinline
                controls
              ></video>
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

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const error = ref('')
const detailData = ref<any>(null)
const fromCache = ref(false)

const sourceId = computed(() => String(route.params.sourceId || ''))
const videoUrl = computed(() => String(route.query.original_url || ''))
const currentPlayUrl = ref<string>('')
const originalUrl = computed(() => String(route.query.original_url || videoUrl.value || ''))
const displayTitle = computed(() => {
  const ep = flatEpisodes.value.find(e => e.url === currentPlayUrl.value)
  if (ep) return `${base.value.name || ''} - ${ep.name}`.trim()
  return String(route.query.title || base.value.name || '')
})
// 同步 HTML 标题与页面标题一致
watch(displayTitle, (t) => {
  const baseTitle = 'Video Crawler'
  const full = t ? `播放 - ${t}` : '播放'
  document.title = [full, baseTitle].filter(Boolean).join(' | ')
}, { immediate: true })

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
const basePoster = computed(() => String((detailData.value?.cover || detailData.value?.poster || ''))) 
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

// 检测是否为移动设备
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// 初始化 Plyr 实例
function ensurePlyr() {
  if (plyr || !videoRef.value) return
  plyr = new Plyr(videoRef.value!, {
    controls: ['play', 'progress', 'current-time', 'duration', 'mute', 'volume', 'settings', 'fullscreen'],
    settings: ['speed'],
    speed: { selected: rate.value, options: rates },
  })
  bindPlayerEvents()
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
    } catch {}
  })
  // 倍速变更（通过 plyr 统一）
  // 元数据
  v.addEventListener('loadedmetadata', () => {
    try {
      if (v.videoWidth && v.videoHeight) { lastVideoW = v.videoWidth; lastVideoH = v.videoHeight }
    } catch {}
  })
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
  } catch {}
}

function playPrev() {
  if (!canPrev.value) return
  playEpisode(currentSourceEpisodes.value[currentIndex.value - 1])
}
function playNext() {
  if (!canNext.value) return
  playEpisode(currentSourceEpisodes.value[currentIndex.value + 1])
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

// 资源列表容错
const resources = computed<any[]>(() => {
  const d: any = detailData.value || {}
  if (Array.isArray(d.resources)) return d.resources
  if (Array.isArray(d.playlist)) return d.playlist
  if (Array.isArray(d.urls)) return d.urls
  if (Array.isArray(d.videos)) return d.videos
  return []
})

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

type PlayState = { url?: string; title?: string; source?: string; currentTime?: number; rate?: number; updatedAt?: number }
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
    const token = auth.token!
    // 优先请求“当前选中剧集”的播放链接；无则回退
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
})
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

.player-wrap {
  position: relative;
  width: 100%;
}

/* 确保播放器控件在移动端也能正常显示 */
@media (max-width: 768px) {
  .player-container {
    margin: 0 0 12px 0; /* 与内容同宽 */
  }
  
  .player-wrap {
    width: 100%; /* 跟随 watch-view 内容宽度 */
    margin-left: 0;
  }
  
  .video-player {
    width: 100% !important; /* 占满容器宽度 */
    height: auto !important;
    aspect-ratio: 16/9; /* 保持16:9比例 */
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
}
</style>


