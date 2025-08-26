<template>
  <div class="bottom-pull-wrapper" ref="wrapper">
    <slot />
    <div class="bottom-indicator" :class="{ active: isActive }" :style="indicatorStyle">
      <div class="indicator-content">
        <div class="spinner" v-if="refreshing"></div>
        <div class="hint">
          <span class="arrow">⬇</span>
          <span class="text">{{ hintText }}</span>
          <span class="progress" v-if="!refreshing && pulling > 0">
            {{ Math.min(100, Math.round((pulling / thresholdPx) * 100)) }}%
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'

interface Props {
  threshold?: number
  refreshingText?: string
  pullText?: string
  releaseText?: string
}

const props = withDefaults(defineProps<Props>(), {
  threshold: 100,
  refreshingText: '正在刷新…',
  pullText: '下拉刷新',
  releaseText: '释放刷新'
})

const emit = defineEmits<{ (e: 'refresh'): void }>()

const wrapper = ref<HTMLElement | null>(null)
let scrollEl: HTMLElement | Document | null = null
const pulling = ref(0)
const lastY = ref(0)
const atBottom = ref(false)
const refreshing = ref(false)

const thresholdPx = computed(() => Number(props.threshold || 100))
const isActive = computed(() => pulling.value > 0 || refreshing.value)

const hintText = computed(() => {
  if (refreshing.value) return props.refreshingText
  return pulling.value >= thresholdPx.value ? props.releaseText : props.pullText
})

const indicatorStyle = computed(() => ({
  transform: `translateY(${Math.min(pulling.value, thresholdPx.value)}px)`
}))

function getScrollEl(): HTMLElement | Document {
  if (document.scrollingElement) return document.scrollingElement as unknown as Document
  const all = Array.from(document.querySelectorAll<HTMLElement>('*'))
  let best: HTMLElement | null = document.documentElement
  let bestScore = 0
  for (const n of all) {
    const sh = n.scrollHeight || 0
    const ch = n.clientHeight || 0
    if (sh > ch + 20) {
      const style = getComputedStyle(n)
      const oy = style.overflowY
      if (oy === 'auto' || oy === 'scroll') {
        const score = sh - ch
        if (score > bestScore) { bestScore = score; best = n }
      }
    }
  }
  return best || document
}

function isNearBottom(): boolean {
  if (!scrollEl) return false
  if (scrollEl === document || (scrollEl as any) === document.scrollingElement) {
    const el = document.scrollingElement as HTMLElement
    return el.scrollTop + el.clientHeight >= el.scrollHeight - 2
  } else {
    const el = scrollEl as HTMLElement
    return el.scrollTop + el.clientHeight >= el.scrollHeight - 2
  }
}

function onTouchStart(e: TouchEvent) {
  if (!e.touches || e.touches.length === 0) return
  lastY.value = e.touches[0].clientY
  pulling.value = 0
}

function onTouchMove(e: TouchEvent) {
  if (!e.touches || e.touches.length === 0) return
  const y = e.touches[0].clientY
  const dy = lastY.value - y // 手指下拉: dy < 0
  lastY.value = y
  atBottom.value = isNearBottom()
  if (atBottom.value && dy < 0 && !refreshing.value) {
    pulling.value += -dy // 累加下拉距离
    if (pulling.value >= thresholdPx.value) {
      pulling.value = 0
      startRefresh()
    }
  }
}

function onWheel(e: WheelEvent) {
  atBottom.value = isNearBottom()
  if (atBottom.value && e.deltaY < 0 && !refreshing.value) { // 向上滚轮为负，等效“下拉”
    pulling.value += -e.deltaY
    if (pulling.value >= thresholdPx.value) {
      pulling.value = 0
      startRefresh()
    }
  }
}

function startRefresh() {
  if (refreshing.value) return
  refreshing.value = true
  emit('refresh')
  setTimeout(() => {
    if (refreshing.value) {
      window.location.reload()
    }
  }, 1200)
}

function done() {
  refreshing.value = false
  pulling.value = 0
}

onMounted(() => {
  scrollEl = getScrollEl()
  const el: any = scrollEl === document ? window : (scrollEl as HTMLElement)
  el.addEventListener('touchstart', onTouchStart, { passive: true })
  el.addEventListener('touchmove', onTouchMove, { passive: true })
  el.addEventListener('wheel', onWheel, { passive: true })
  ;(window as any).BottomPullRefreshDone = done
})

onBeforeUnmount(() => {
  if (!scrollEl) return
  const el: any = scrollEl === document ? window : (scrollEl as HTMLElement)
  el.removeEventListener('touchstart', onTouchStart as any)
  el.removeEventListener('touchmove', onTouchMove as any)
  el.removeEventListener('wheel', onWheel as any)
})
</script>

<style scoped>
.bottom-pull-wrapper {
  position: relative;
  min-height: 100%;
  touch-action: pan-y; /* 明确允许垂直滚动 */
}
.bottom-indicator {
  position: fixed;
  bottom: 8px;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  transition: transform 0.15s ease-out, opacity 0.2s ease;
  opacity: 0.0;
  z-index: 2147483647; /* 确保可见 */
}
.bottom-indicator.active { opacity: 1.0; }
.indicator-content {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(0,0,0,0.55);
  color: #fff;
  padding: 6px 10px;
  border-radius: 14px;
  backdrop-filter: blur(2px);
}
.spinner {
  width: 24px;
  height: 24px;
  border: 3px solid rgba(16, 185, 129, 0.3);
  border-top-color: #10b981;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
}
.hint { display: inline-flex; align-items: center; gap: 8px; }
.hint .arrow { font-size: 14px; }
.hint .text { font-size: 13px; }
.hint .progress { font-size: 12px; opacity: 0.9; }

@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
