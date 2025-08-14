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
                @search="handleSearch"
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
            
            <a-card-meta :title="movie.title">
              <template #description>
                <div class="movie-info">
                  <div class="movie-meta">
                    <div class="meta-item">
                      <span class="label">导演：</span>
                      <span class="value">{{ movie.director || '未知' }}</span>
                    </div>
                    <div class="meta-item">
                      <span class="label">主演：</span>
                      <span class="value">{{ movie.actors || '未知' }}</span>
                    </div>
                    <div class="meta-item">
                      <span class="label">上映：</span>
                      <span class="value">{{ movie.release_date || '未知' }}</span>
                    </div>
                    <div class="meta-item">
                      <span class="label">地区：</span>
                      <span class="value">{{ movie.region || '未知' }}</span>
                    </div>
                    <div class="meta-item">
                      <span class="label">来源：</span>
                      <a-tag color="green">{{ movie.source_name }}</a-tag>
                    </div>
                  </div>
                  <div class="movie-description">
                    {{ movie.description || '暂无简介' }}
                  </div>
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
import { message } from 'ant-design-vue'
import { SearchOutlined, PlayCircleOutlined } from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'

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

// Mock 搜索结果数据
const searchResults = ref<MovieResult[]>([])

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
    // 模拟API延迟
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // 模拟搜索结果
    const mockResults: MovieResult[] = [
      {
        id: '1',
        title: '流浪地球2',
        cover_url: 'https://via.placeholder.com/300x400/52c41a/ffffff?text=流浪地球2',
        director: '郭帆',
        actors: '吴京、刘德华、李雪健',
        release_date: '2023-01-22',
        region: '中国大陆',
        source_name: '综合站点',
        description: '太阳即将毁灭，人类在地球表面建造出巨大的推进器，寻找新的家园。',
        video_url: 'https://example.com/movie1',
        original_url: 'https://original-site.com/movie1'
      },
      {
        id: '2',
        title: '满江红',
        cover_url: 'https://via.placeholder.com/300x400/1890ff/ffffff?text=满江红',
        director: '张艺谋',
        actors: '沈腾、易烊千玺、张译',
        release_date: '2023-01-22',
        region: '中国大陆',
        source_name: '短剧站点',
        description: '南宋绍兴年间，岳飞死后四年，秦桧率兵与金国会谈。',
        video_url: 'https://example.com/movie2',
        original_url: 'https://original-site.com/movie2'
      }
    ]
    
    // 根据搜索关键词过滤结果
    const filteredResults = mockResults.filter(movie => 
      movie.title.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      movie.director.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      movie.actors.toLowerCase().includes(searchKeyword.value.toLowerCase())
    )
    
    searchResults.value = filteredResults
    saveSearchHistory(searchKeyword.value)
    message.success(`搜索完成，找到 ${filteredResults.length} 个结果`)
  } catch (error) {
    message.error('搜索失败，请重试')
  } finally {
    searching.value = false
  }
}

// 选择搜索历史
const handleSelectHistory = (value: string) => {
  searchKeyword.value = value
  handleSearch()
}

// 开始观看
const startWatching = (movie: MovieResult) => {
  message.info(`开始观看：${movie.title}`)
}

// 跳转原站点
const goToOriginal = (movie: MovieResult) => {
  window.open(movie.original_url, '_blank')
}
</script>

<style scoped>
@import './MovieView.css';
</style>
