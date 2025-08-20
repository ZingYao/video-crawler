<template>
  <AppLayout page-title="观看">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>{{ route.query.title || '视频详情' }}</h2>
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

          <div class="detail-layout">
            <div class="detail-main">
              <a-card size="small" title="接口返回原始数据" :bordered="true">
                <a-space style="margin-bottom: 8px;">
                  <a-button size="small" @click="copyJson">复制 JSON</a-button>
                </a-space>
                <pre class="json-pre">{{ formattedJson }}</pre>
              </a-card>
            </div>
          </div>
        </template>
      </div>
    </a-card>
  </AppLayout>
  
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import AppLayout from '@/components/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { videoAPI } from '@/api'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const error = ref('')
const detailData = ref<any>(null)
const fromCache = ref(false)

const sourceId = computed(() => String(route.query.source_id || ''))
const videoUrl = computed(() => String(route.query.url || ''))

const cacheKey = computed(() => `watch_detail:${sourceId.value}:${encodeURIComponent(videoUrl.value)}`)

const formattedJson = computed(() => {
  try {
    return JSON.stringify(detailData.value, null, 2)
  } catch {
    return String(detailData.value)
  }
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
  } catch (e: any) {
    error.value = e?.message || '获取详情失败'
  } finally {
    loading.value = false
  }
}

function refreshDetail() {
  fetchDetail(true)
}

function goBack() {
  router.back()
}

function copyJson() {
  try {
    navigator.clipboard.writeText(formattedJson.value)
    message.success('已复制')
  } catch {
    message.error('复制失败')
  }
}

onMounted(() => {
  fetchDetail(false)
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
}
.card-header h2 {
  margin: 0;
}
.detail-layout {
  display: flex;
  gap: 16px;
}
.detail-main {
  flex: 1;
}
.json-pre {
  margin: 0;
  padding: 12px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 8px;
  overflow: auto;
  max-height: 70vh;
}
</style>


