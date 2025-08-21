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
              <VideoPlayer
                v-if="playerSource"
                class="video-player vjs-default-skin"
                :src="playerSource"
                :poster="basePoster"
                :playsinline="true"
                :controls="true"
                :volume="1"
                :playbackRates="rates"
                :fluid="true"
                :autoplay="false"
                :options="playerOptions"
                ref="playerRef"
              />
              <div v-if="isScrubbing" class="scrub-overlay">{{ scrubLabel }}</div>
            </div>
            <div class="player-actions">
              <a-space wrap>
                <a-button size="small" @click="playPrev" :disabled="!canPrev">上一集</a-button>
                <a-button size="small" @click="playNext" :disabled="!canNext">下一集</a-button>
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
import { VideoPlayer } from '@videojs-player/vue'
import 'video.js/dist/video-js.css'

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

const cacheKey = computed(() => `watch_detail:${sourceId.value}:${encodeURIComponent(videoUrl.value)}`)
// 使用 sourceId + original_url 作为进度键，避免标题变化导致无法命中
const playStateKey = computed(() => {
  const keyUrl = String(route.query.original_url || videoUrl.value || '')
  return `watch_state:${sourceId.value}:${encodeURIComponent(keyUrl)}`
})

 

// 播放器相关
const playerRef = ref()
const playerSource = ref('')
const basePoster = computed(() => String((detailData.value?.cover || detailData.value?.poster || ''))) 
const rates = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 2, 2.5, 3]
const rate = ref(1)

// 检测是否为移动设备
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// 根据设备类型生成播放器配置
const playerOptions = computed(() => {
  const baseOptions = {
    controlBar: {
      children: [
        'playToggle',
        'volumePanel',
        'currentTimeDisplay',
        'timeDivider',
        'durationDisplay',
        'progressControl',
        'fullscreenToggle'
      ]
    }
  }
  
  if (isMobile.value) {
    // 移动端：使用切换按钮而不是下拉菜单
    baseOptions.controlBar.children.splice(6, 0, 'playbackRateMenuButton')
  } else {
    // 桌面端：使用下拉菜单
    baseOptions.controlBar.children.splice(6, 0, 'playbackRateMenuButton')
  }
  
  return baseOptions
})
let lastSavedSecond = 0
let playerBound = false
let lastVideoW = 0
let lastVideoH = 0
let orientationLocked = false
// 拖动进度条（视频区域手势）
const isScrubbing = ref(false)
const scrubLabel = ref('00:00 / 00:00')
let startX = 0
let startTime = 0
let totalDur = 0

watch(rate, (v) => {
  const player = playerRef.value?.player
  if (player && typeof v === 'number') {
    try { player.playbackRate(v) } catch {}
  }
})

function bindPlayerEvents() {
  const player = playerRef.value?.player
  if (!player || playerBound) return
  playerBound = true
  player.on('timeupdate', () => {
    try {
      const ct = Math.floor(player.currentTime() || 0)
      const dur = Math.floor(player.duration() || 0)
      if (dur > 0 && Math.abs(ct - lastSavedSecond) >= 5) {
        lastSavedSecond = ct
        savePlayState({ currentTime: ct })
      }
    } catch {}
  })
  player.on('ratechange', () => {
    try { savePlayState({ rate: player.playbackRate() }) } catch {}
  })
  // 元数据就绪时记录视频宽高
  player.on('loadedmetadata', () => {
    try {
      const vw = typeof player.videoWidth === 'function' ? player.videoWidth() : 0
      const vh = typeof player.videoHeight === 'function' ? player.videoHeight() : 0
      if (vw && vh) { lastVideoW = vw; lastVideoH = vh }
      totalDur = Math.floor(player.duration() || 0)
    } catch {}
  })
  // 全屏切换监听（大多数安卓/部分浏览器）
  player.on('fullscreenchange', async () => {
    try {
      if (typeof player.isFullscreen === 'function' && player.isFullscreen()) {
        await handleEnterFullscreen()
      } else {
        await handleExitFullscreen()
      }
    } catch {}
  })
  // 在尝试进入全屏的瞬间（按钮点击）提前锁定，兼容部分浏览器必须在用户手势中调用 lock()
  try {
    const fsBtn = player.controlBar?.fullscreenToggle?.el?.() as HTMLElement | undefined
    if (fsBtn) {
      fsBtn.addEventListener('click', async () => {
        try { await handleEnterFullscreen() } catch {}
      }, { passive: true })
    }
  } catch {}
  // iOS Safari 原生全屏事件（通过原生 video 元素）
  try {
    const videoEl: any = player.el()?.querySelector?.('video')
    if (videoEl) {
      videoEl.addEventListener('loadedmetadata', () => {
        if (videoEl.videoWidth && videoEl.videoHeight) { lastVideoW = videoEl.videoWidth; lastVideoH = videoEl.videoHeight }
      })
      videoEl.addEventListener('webkitbeginfullscreen', handleEnterFullscreen as any)
      videoEl.addEventListener('webkitendfullscreen', handleExitFullscreen as any)
      // 绑定横向滑动快进/快退
      bindScrubGesture(videoEl)
    }
  } catch {}
}

async function handleEnterFullscreen() {
  const orientation = estimateOrientation()
  await lockOrientation(orientation)
}

async function handleExitFullscreen() {
  await unlockOrientation()
}

function estimateOrientation(): 'landscape' | 'portrait' {
  // 优先使用真实视频宽高
  const w = lastVideoW || (playerRef.value?.player?.videoWidth?.() || 0)
  const h = lastVideoH || (playerRef.value?.player?.videoHeight?.() || 0)
  if (w > 0 && h > 0) return w >= h ? 'landscape' : 'portrait'
  // 退化为容器尺寸
  const el: any = playerRef.value?.player?.el?.() || playerRef.value?.$el
  const cw = el?.clientWidth || window.innerWidth
  const ch = el?.clientHeight || window.innerHeight
  return cw >= ch ? 'landscape' : 'portrait'
}

async function lockOrientation(ori: 'landscape' | 'portrait') {
  try {
    const o: any = (screen as any).orientation
    if (o && typeof o.lock === 'function') {
      await o.lock(ori === 'landscape' ? 'landscape' : 'portrait-primary')
      orientationLocked = true
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
    }
  } catch {}
  orientationLocked = false
}

function bindScrubGesture(videoEl: HTMLElement) {
  const fmt = (sec: number) => new Date(Math.max(sec,0)*1000).toISOString().substring(11,19)
  const onTouchStart = (e: TouchEvent) => {
    if (!playerRef.value?.player) return
    if ((playerRef.value.player as any).paused && e.touches.length === 1) {
      // 保持原行为
    }
    startX = e.touches[0].clientX
    startTime = Math.floor(playerRef.value.player.currentTime() || 0)
    totalDur = Math.floor(playerRef.value.player.duration() || 0)
    scrubLabel.value = `${fmt(startTime)} / ${fmt(totalDur)}`
    isScrubbing.value = true
  }
  const onTouchMove = (e: TouchEvent) => {
    if (!isScrubbing.value) return
    const dx = e.touches[0].clientX - startX
    const el = playerRef.value?.player?.el?.() as HTMLElement
    const width = el?.clientWidth || window.innerWidth
    const deltaSec = Math.floor((dx / Math.max(width,1)) * (totalDur || 0))
    const next = Math.min(Math.max(startTime + deltaSec, 0), totalDur)
    scrubLabel.value = `${fmt(next)} / ${fmt(totalDur)}`
  }
  const onTouchEnd = (e: TouchEvent) => {
    if (!isScrubbing.value) return
    const dx = (e.changedTouches?.[0]?.clientX || startX) - startX
    const el = playerRef.value?.player?.el?.() as HTMLElement
    const width = el?.clientWidth || window.innerWidth
    const deltaSec = Math.floor((dx / Math.max(width,1)) * (totalDur || 0))
    const next = Math.min(Math.max(startTime + deltaSec, 0), totalDur)
    try { playerRef.value?.player?.currentTime(next) } catch {}
    savePlayState({ currentTime: next })
    isScrubbing.value = false
  }
  videoEl.addEventListener('touchstart', onTouchStart, { passive: true })
  videoEl.addEventListener('touchmove', onTouchMove, { passive: true })
  videoEl.addEventListener('touchend', onTouchEnd, { passive: true })
  videoEl.addEventListener('touchcancel', onTouchEnd, { passive: true })
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
const currentIndex = computed(() => flatEpisodes.value.findIndex(e => e.url === currentPlayUrl.value))
const canPrev = computed(() => currentIndex.value > 0)
const canNext = computed(() => currentIndex.value >= 0 && currentIndex.value < flatEpisodes.value.length - 1)

function isCurrentEpisode(ep: { url: string }) {
  return String(ep?.url || '') === currentPlayUrl.value
}

async function playEpisode(ep: { name: string; url: string }, sourceName?: string) {
  if (!ep?.url) return
  // 仅更新地址栏中标题与来源，不修改 url 参数，避免影响回显
  const q = { ...route.query, title: ep.name, source: sourceName || (ep as any).__sourceName }
  router.replace({ name: 'watch', params: route.params, query: q })
  // 直接解析新的播放地址并替换当前播放器
  try {
    const token = auth.token!
    const res: any = await videoAPI.playUrl(token, sourceId.value, ep.url)
    const url: string = res?.data?.video_url || res?.data || ''
    if (url) {
      playerSource.value = url
      await nextTick()
      const player = playerRef.value?.player
      if (player) {
        try { player.playbackRate(rate.value) } catch {}
        try { player.play() } catch {}
        // 如果有缓存的进度且对应当前剧集，自动续播
        const state = loadPlayState()
        if (state && String(state.url) === ep.url && state.currentTime && state.currentTime > 0) {
          try { player.currentTime(state.currentTime) } catch {}
        }
        bindPlayerEvents()
      }
      // 保存所选剧集
      savePlayState({ url: ep.url, title: ep.name, source: q.source })
      currentPlayUrl.value = ep.url
    }
  } catch {}
}

function playPrev() {
  if (!canPrev.value) return
  playEpisode(flatEpisodes.value[currentIndex.value - 1])
}
function playNext() {
  if (!canNext.value) return
  playEpisode(flatEpisodes.value[currentIndex.value + 1])
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
  } catch {}
}
function loadPlayState(): PlayState | null {
  try {
    const raw = localStorage.getItem(playStateKey.value)
    if (!raw) return null
    return JSON.parse(raw)
  } catch { return null }
}

async function fetchDetail(force = false) {
  error.value = ''
  if (!force) {
    const hit = loadCache()
    fromCache.value = hit
    if (hit) return
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
    const res: any = await videoAPI.playUrl(token, sourceId.value, videoUrl.value)
    const url: string = res?.data?.video_url || res?.data || ''
    playerSource.value = url
    await nextTick()
    const player = playerRef.value?.player
    if (player) {
      try { player.src({ src: url, type: 'application/x-mpegURL' }) } catch {}
      try { player.playbackRate(rate.value) } catch {}
      // 等待 metadata 再恢复进度，避免 duration 为 0 导致 seek 失败
      const state = loadPlayState()
      const seekTo = state?.currentTime || 0
      if (seekTo > 0) {
        const doSeek = () => { try { player.currentTime(seekTo) } catch {} }
        if (player.readyState() >= 1) doSeek()
        else player.one('loadedmetadata', doSeek)
      }
      try { player.play() } catch {}
      bindPlayerEvents()
    }
    // 初始化情况下，将当前播放 url 与初始 url 对齐
    if (!currentPlayUrl.value) currentPlayUrl.value = videoUrl.value
  } catch (e) {
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

 

// 移动端长按：2 倍速（绑定到原生 video，更可靠）
let touchTimer: any = null
let longPressActive = false
function bindLongPress() {
  const videoEl: any = playerRef.value?.player?.el()?.querySelector?.('video')
  if (!videoEl) return
  const start = () => {
    if (touchTimer) clearTimeout(touchTimer)
    touchTimer = setTimeout(() => {
      const player = playerRef.value?.player
      if (!player) return
      try { player.playbackRate(2) } catch {}
      longPressActive = true
    }, 350)
  }
  const end = () => {
    const player = playerRef.value?.player
    if (!player) return
    if (touchTimer) clearTimeout(touchTimer)
    if (longPressActive) {
      try { player.playbackRate(rate.value) } catch {}
      longPressActive = false
    }
  }
  videoEl.addEventListener('touchstart', start, { passive: true })
  videoEl.addEventListener('touchend', end, { passive: true })
  videoEl.addEventListener('touchcancel', end, { passive: true })
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
  // 绑定长按 2x
  bindLongPress()
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
    margin: 0 -12px 12px -12px; /* 负边距让播放器延伸到容器边缘 */
  }
  
  .player-wrap {
    width: 100vw; /* 占满视口宽度 */
    margin-left: calc(-50vw + 50%); /* 居中显示 */
  }
  
  .video-player {
    width: 100% !important;
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
  
  /* 移动端倍速菜单优化 */
  .video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-content) {
    min-width: 80px;
    max-height: 200px;
    overflow-y: auto;
  }
  
  .video-player :deep(.vjs-playback-rate-menu-button .vjs-menu-item) {
    padding: 8px 12px;
    font-size: 14px;
    text-align: center;
    min-height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  /* 确保倍速按钮在移动端更容易点击 */
  .video-player :deep(.vjs-playback-rate-menu-button) {
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
.player-actions { margin-top: 8px; }
.scrub-overlay { position:absolute; left:50%; top:50%; transform:translate(-50%,-50%); padding:6px 10px; background:rgba(0,0,0,.6); color:#fff; border-radius:6px; font-size:12px; }
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
  .kv-list { grid-template-columns: 1fr; }
  .card-header h2 { white-space: normal; }
}
</style>


