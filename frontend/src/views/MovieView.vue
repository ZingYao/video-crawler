<template>
    <AppLayout page-title="观影">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>观影</h2>
          <p>搜索和观看您喜欢的视频内容</p>
        </div>
      </template>

      <div class="movie-view">
        <!-- 搜索区域 -->
        <div class="search-section">
          <a-card class="search-card">
        <div class="search-form">
          <a-row :gutter="16">
            <a-col :xs="24" :sm="12" :md="14" :lg="16">
              <a-auto-complete
                v-model:value="searchKeyword"
                :options="searchHistoryOptions"
                placeholder="请输入影片名称"
                size="large"
                @keydown.enter="handleSearch"
                @select="handleSelectHistory"
                @focus="loadSearchHistory"
              >
                <template #suffix>
                  <SearchOutlined />
                </template>
              </a-auto-complete>
            </a-col>
            <a-col :xs="12" :sm="8" :md="6" :lg="6">
              <a-select
                v-model:value="selectedSourceType"
                placeholder="选择站点类型"
                size="large"
                style="width: 100%"
              >
                <a-select-option value="">所有</a-select-option>
                <a-select-option value="0">综合</a-select-option>
                <a-select-option value="1">短剧</a-select-option>
                <a-select-option value="2">电影</a-select-option>
                <a-select-option value="3">电视剧</a-select-option>
                <a-select-option value="4">综艺</a-select-option>
                <a-select-option value="5">动漫</a-select-option>
                <a-select-option value="6">纪录片</a-select-option>
                <a-select-option value="7">其他</a-select-option>
              </a-select>
            </a-col>
            <a-col :xs="12" :sm="4" :md="4" :lg="2">
              <a-button
                type="primary"
                size="large"
                @click="handleSearch"
                :loading="searching"
                style="width: 100%"
              >
                搜索
              </a-button>
            </a-col>
          </a-row>
        </div>
      </a-card>
    </div>

    <!-- 搜索结果区域 -->
    <div class="results-section" v-if="searchResults.length > 0 && hasSearched">
      <a-row :gutter="[16, 16]">
        <a-col
          v-for="movie in searchResults"
          :key="movie.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
          :xl="4"
        >
          <a-card class="movie-card" hoverable>
            <template #cover>
              <div class="movie-cover">
                <img :src="movie.cover_url" :alt="movie.title" />
                <div class="movie-overlay">
                  <a-space>
                    <a-button
                      type="primary"
                      size="small"
                      @click="startWatching(movie)"
                    >
                      开始观看
                    </a-button>
                    <a-button
                      size="small"
                      @click="goToOriginal(movie)"
                    >
                      原站点
                    </a-button>
                  </a-space>
                </div>
              </div>
            </template>
            
            <a-card-meta>
              <template #title>
                <a-tooltip :title="movie.title" :overlayStyle="tooltipOverlayStyle">
                  <div class="movie-title">{{ movie.title }}</div>
                </a-tooltip>
              </template>
              <template #description>
                <div class="movie-info">
                  <div class="movie-meta">
                    <div class="meta-item">
                      <span class="label">导演：</span>
                      <a-tooltip :title="movie.director || '未知'" :overlayStyle="tooltipOverlayStyle">
                        <span class="value">{{ movie.director || '未知' }}</span>
                      </a-tooltip>
                    </div>
                    <div class="meta-item">
                      <span class="label">主演：</span>
                      <a-tooltip :title="movie.actors || '未知'" :overlayStyle="tooltipOverlayStyle">
                        <span class="value">{{ movie.actors || '未知' }}</span>
                      </a-tooltip>
                    </div>
                    <div class="meta-item">
                      <span class="label">上映：</span>
                      <a-tooltip :title="movie.release_date || '未知'" :overlayStyle="tooltipOverlayStyle">
                        <span class="value">{{ movie.release_date || '未知' }}</span>
                      </a-tooltip>
                    </div>
                    <div class="meta-item">
                      <span class="label">地区：</span>
                      <a-tooltip :title="movie.region || '未知'" :overlayStyle="tooltipOverlayStyle">
                        <span class="value">{{ movie.region || '未知' }}</span>
                      </a-tooltip>
                    </div>
                    <div class="meta-item">
                      <span class="label">来源：</span>
                      <a-tag color="green">{{ movie.source_name }}</a-tag>
                    </div>
                  </div>
                  <a-tooltip :title="movie.description || '暂无简介'" :overlayStyle="tooltipOverlayStyle">
                    <div class="movie-description">
                      {{ movie.description || '暂无简介' }}
                    </div>
                  </a-tooltip>
                </div>
              </template>
            </a-card-meta>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 空状态 -->
    <div class="empty-section" v-else-if="hasSearched && searchResults.length === 0">
      <a-empty description="暂无搜索结果" />
    </div>

    <!-- 初始状态 -->
    <div class="welcome-section" v-else-if="!hasSearched">
      <a-card class="welcome-card">
        <a-result
          status="success"
          title="欢迎来到观影页面"
          sub-title="请输入影片名称开始搜索"
        >
          <template #icon>
            <PlayCircleOutlined style="font-size: 64px; color: #52c41a;" />
          </template>
        </a-result>
      </a-card>
    </div>
      </div>
    </a-card>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { SearchOutlined, PlayCircleOutlined } from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI, videoAPI } from '@/api'

interface MovieResult {
  id: string
  title: string
  cover_url: string
  director: string
  actors: string
  release_date: string
  region: string
  source_name: string
  description: string
  video_url: string
  original_url: string
  source_id: string
}

// 搜索相关状态
const searchKeyword = ref('')
const selectedSourceType = ref('0')
const searching = ref(false)
const hasSearched = ref(false)

// 搜索历史
const searchHistory = ref<string[]>([])
const searchHistoryOptions = computed(() => {
  return searchHistory.value.map(item => ({
    value: item,
    label: item
  }))
})

// 搜索结果数据
const searchResults = ref<MovieResult[]>([])
const router = useRouter()
const auth = useAuthStore()

// Tooltip 弹层样式：限制最大宽度、允许换行/长词换行
const tooltipOverlayStyle = { maxWidth: '70vw', whiteSpace: 'normal', wordBreak: 'break-word' }

// 加载搜索历史
const loadSearchHistory = () => {
  const history = localStorage.getItem('searchHistory')
  if (history) {
    searchHistory.value = JSON.parse(history)
  }
}

// 保存搜索历史
const saveSearchHistory = (keyword: string) => {
  if (!keyword.trim()) return
  
  const history = searchHistory.value.filter(item => item !== keyword)
  history.unshift(keyword)
  
  if (history.length > 10) {
    history.splice(10)
  }
  
  searchHistory.value = history
  localStorage.setItem('searchHistory', JSON.stringify(history))
}

// 处理搜索
const handleSearch = async () => {
  if (!searchKeyword.value.trim()) {
    message.warning('请输入搜索关键词')
    return
  }

  searching.value = true
  hasSearched.value = true
  
  try {
    // 1) 拉取站点列表并按 sort 降序（越大越靠前）
    const token = auth.token!
    const listResp: any = await videoSourceAPI.getVideoSourceList(token)
    const sources: Array<{ id: string; name: string; sort: number; source_type: number; status: number }>= (listResp?.data || [])
      // 仅搜索正常状态的站点（status=1）
      .filter((s: any) => Number(s.status) === 1)
      // 站点类型过滤（允许“所有”）
      .filter((s: any) => selectedSourceType.value === '' || String(s.source_type) === String(selectedSourceType.value))
      // 按 sort 降序
      .sort((a: any, b: any) => b.sort - a.sort)

    // 2) 并发度=2 线程池，按顺序调度
    const concurrency = 2
    const queue = [...sources]
    const results: MovieResult[] = []

    const runner = async () => {
      while (queue.length > 0) {
        const src = queue.shift()!
        try {
          const res: any = await videoAPI.search(token, src.id, searchKeyword.value)
          const items: any[] = Array.isArray(res?.data) ? res.data : []
          // 统一映射到 MovieResult
          for (const item of items) {
            results.push({
              id: `${src.id}:${item.url || item.name}`,
              title: String(item.name || ''),
              cover_url: String(item.cover || ''),
              director: String(item.director || ''),
              actors: String(item.actor || ''),
              release_date: String(item.release_date || ''),
              region: String(item.region || ''),
              source_name: String(src.name || ''),
              description: String(item.description || ''),
              video_url: String(item.url || ''),
              original_url: String(item.url || ''),
              source_id: String(src.id || ''),
            })
          }
        } catch (e) {
          // 单站点失败不中断
          console.warn('search failed for source', src?.id, e)
        }
      }
    }

    const workers: Promise<void>[] = []
    for (let i = 0; i < Math.min(concurrency, queue.length); i++) {
      workers.push(runner())
    }
    await Promise.all(workers)

    searchResults.value = results
    saveSearchHistory(searchKeyword.value)
    message.success(`搜索完成，找到 ${results.length} 个结果`)
  } catch (error) {
    message.error('搜索失败，请重试')
  } finally {
    searching.value = false
  }
}

// 选择搜索历史：填充并触发搜索
const handleSelectHistory = (value: string) => {
  searchKeyword.value = value
  handleSearch()
}

// 开始观看
const startWatching = (movie: MovieResult) => {
  if (!movie.source_id || !(movie.video_url || movie.original_url)) {
    message.error('缺少播放所需信息')
    return
  }
  router.push({
    name: 'watch',
    query: {
      source_id: movie.source_id,
      url: movie.video_url || movie.original_url,
      title: movie.title || '',
    },
  })
}

// 跳转原站点
const goToOriginal = (movie: MovieResult) => {
  window.open(movie.original_url, '_blank')
}
</script>

<style scoped>
@import './MovieView.css';
</style>
