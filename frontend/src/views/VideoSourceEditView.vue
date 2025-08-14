<template>
  <AppLayout :page-title="isEdit ? '编辑视频源' : '添加视频源'">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>{{ isEdit ? '编辑视频源' : '添加视频源' }}</h2>
          <p>{{ isEdit ? '修改视频源站点配置' : '创建新的视频源站点' }}</p>
        </div>
      </template>

      <div class="back-button-container">
        <a-button @click="goBack" class="back-button">
          <template #icon>
            <ArrowLeftOutlined />
          </template>
          返回列表
        </a-button>
      </div>

      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
        class="video-source-form"
        @finish="handleSave"
      >
        <!-- 基本信息 -->
        <a-divider>基本信息</a-divider>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="站点名称" name="name">
              <a-input v-model:value="formData.name" placeholder="请输入站点名称" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="站点状态" name="status">
              <a-select v-model:value="formData.status" placeholder="请选择状态">
                <a-select-option :value="0">禁用</a-select-option>
                <a-select-option :value="1">正常</a-select-option>
                <a-select-option :value="2">维护中</a-select-option>
                <a-select-option :value="3">不可用</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="资源类型" name="source_type">
              <a-select v-model:value="formData.source_type" placeholder="请选择资源类型">
                <a-select-option :value="0">综合</a-select-option>
                <a-select-option :value="1">短剧</a-select-option>
                <a-select-option :value="2">电影</a-select-option>
                <a-select-option :value="3">电视剧</a-select-option>
                <a-select-option :value="4">综艺</a-select-option>
                <a-select-option :value="5">动漫</a-select-option>
                <a-select-option :value="6">纪录片</a-select-option>
                <a-select-option :value="7">其他</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="站点域名" name="domain">
          <a-input v-model:value="formData.domain" placeholder="请输入站点域名，如：http://example.com" />
        </a-form-item>

        <!-- 搜索配置 -->
        <a-divider>搜索配置</a-divider>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="搜索路径" name="search_config.search_path">
              <a-input v-model:value="formData.search_config.search_path" placeholder="/search.php" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="搜索方法" name="search_config.search_method">
              <a-select v-model:value="formData.search_config.search_method" placeholder="GET">
                <a-select-option value="get">GET</a-select-option>
                <a-select-option value="post">POST</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="搜索关键字位置" name="search_config.search_key_position">
              <a-select v-model:value="formData.search_config.search_key_position" placeholder="url">
                <a-select-option value="url">URL</a-select-option>
                <a-select-option value="body">Body</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="搜索关键字" name="search_config.search_key">
              <a-input v-model:value="formData.search_config.search_key" placeholder="searchword" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="页码位置" name="search_config.page_key_position">
              <a-select v-model:value="formData.search_config.page_key_position" placeholder="url">
                <a-select-option value="url">URL</a-select-option>
                <a-select-option value="body">Body</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="页码关键字" name="search_config.page_key">
              <a-input v-model:value="formData.search_config.page_key" placeholder="page" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="搜索类型关键字位置" name="search_config.search_type_key_position">
              <a-select v-model:value="formData.search_config.search_type_key_position" placeholder="url">
                <a-select-option value="url">URL</a-select-option>
                <a-select-option value="body">Body</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="搜索类型关键字" name="search_config.search_type_key">
              <a-input v-model:value="formData.search_config.search_type_key" placeholder="searchtype" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="视频URL是否绝对路径" name="search_config.video_url_is_absolute">
              <a-switch v-model:checked="formData.search_config.video_url_is_absolute" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频卡片CSS选择器" name="search_config.video_card_css_filter">
              <a-input v-model:value="formData.search_config.video_card_css_filter" placeholder="ul > li" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频标题CSS选择器" name="search_config.video_title_css_filter">
              <a-input v-model:value="formData.search_config.video_title_css_filter" placeholder="h3.title > a" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频封面CSS选择器" name="search_config.video_cover_image_css_filter">
              <a-input v-model:value="formData.search_config.video_cover_image_css_filter" placeholder=".thumb > a" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频详情URL CSS选择器" name="search_config.video_detail_url_css_filter">
              <a-input v-model:value="formData.search_config.video_detail_url_css_filter" placeholder="p.margin-0 > a.btn.btn-min.btn-default" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频播放URL CSS选择器" name="search_config.video_player_url_css_filter">
              <a-input v-model:value="formData.search_config.video_player_url_css_filter" placeholder="p.margin-0 > a.btn.btn-min.btn-primary" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="总数量CSS选择器" name="search_config.total_count_css_filter">
              <a-input v-model:value="formData.search_config.total_count_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="总数量正则" name="search_config.total_count_regex">
              <a-input v-model:value="formData.search_config.total_count_regex" placeholder="共有&quot;([0-9]+)&quot;部影片" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="当前页码CSS选择器" name="search_config.current_page_css_filter">
              <a-input v-model:value="formData.search_config.current_page_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="当前页码正则" name="search_config.current_page_regex">
              <a-input v-model:value="formData.search_config.current_page_regex" placeholder="当前第&quot;([0-9]+)&quot;页" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频导演CSS选择器" name="search_config.video_director_css_filter">
              <a-input v-model:value="formData.search_config.video_director_css_filter" placeholder="div.detail > p:first-of-type" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频导演正则" name="search_config.video_director_regex">
              <a-input v-model:value="formData.search_config.video_director_regex" placeholder="导演：(.+)" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频演员CSS选择器" name="search_config.video_actor_css_filter">
              <a-input v-model:value="formData.search_config.video_actor_css_filter" placeholder="div.detail > p:nth-child(3) > a" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频演员正则" name="search_config.video_actor_regex">
              <a-input v-model:value="formData.search_config.video_actor_regex" placeholder="演员：(.+)" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频年份CSS选择器" name="search_config.video_year_css_filter">
              <a-input v-model:value="formData.search_config.video_year_css_filter" placeholder="div.detail > p:nth-child(4) > span:nth-child(4)" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频年份正则" name="search_config.video_year_regex">
              <a-input v-model:value="formData.search_config.video_year_regex" placeholder="年份：([0-9]{4})" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频地区CSS选择器" name="search_config.video_region_css_filter">
              <a-input v-model:value="formData.search_config.video_region_css_filter" placeholder="div.detail > p:nth-child(4) > span:nth-child(3)" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频地区正则" name="search_config.video_region_regex">
              <a-input v-model:value="formData.search_config.video_region_regex" placeholder="地区：(.+)" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频类型CSS选择器" name="search_config.video_type_css_filter">
              <a-input v-model:value="formData.search_config.video_type_css_filter" placeholder="div.detail > p:nth-child(4) > span:nth-child(2)" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频类型正则" name="search_config.video_type_regex">
              <a-input v-model:value="formData.search_config.video_type_regex" placeholder="类型：(.+)" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频语言CSS选择器" name="search_config.video_language_css_filter">
              <a-input v-model:value="formData.search_config.video_language_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频描述CSS选择器" name="search_config.video_description_css_filter">
              <a-input v-model:value="formData.search_config.video_description_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <!-- 详情页配置 -->
        <a-divider>详情页配置</a-divider>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频卡片CSS选择器" name="video_desc_page_config.video_card_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_card_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频封面CSS选择器" name="video_desc_page_config.video_cover_image_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_cover_image_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频标题CSS选择器" name="video_desc_page_config.video_title_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_title_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频播放URL CSS选择器" name="video_desc_page_config.video_player_url_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_player_url_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频URL是否绝对路径" name="video_desc_page_config.video_url_is_absolute">
              <a-switch v-model:checked="formData.video_desc_page_config.video_url_is_absolute" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频导演正则" name="video_desc_page_config.video_director_regex">
              <a-input v-model:value="formData.video_desc_page_config.video_director_regex" placeholder="正则表达式" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频导演CSS选择器" name="video_desc_page_config.video_director_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_director_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频演员CSS选择器" name="video_desc_page_config.video_actor_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_actor_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频年份CSS选择器" name="video_desc_page_config.video_year_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_year_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频地区CSS选择器" name="video_desc_page_config.video_area_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_area_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频类型CSS选择器" name="video_desc_page_config.video_type_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_type_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="视频语言CSS选择器" name="video_desc_page_config.video_language_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_language_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="视频描述CSS选择器" name="video_desc_page_config.video_description_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.video_description_css_filter" placeholder="CSS选择器" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="源卡片CSS选择器" name="video_desc_page_config.source_card_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.source_card_css_filter" placeholder="div.stui-pannel.stui-pannel-bg.clearfix > div.stui-pannel-box:has(div.stui-pannel_hd)" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="源名称CSS选择器" name="video_desc_page_config.source_name_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.source_name_css_filter" placeholder="div.stui-pannel_hd > h3 >img" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="剧集列表CSS选择器" name="video_desc_page_config.episode_list_css_filter">
              <a-input v-model:value="formData.video_desc_page_config.episode_list_css_filter" placeholder="ul > li > a" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="剧集名称正则" name="video_desc_page_config.episode_name_regex">
              <a-input v-model:value="formData.video_desc_page_config.episode_name_regex" placeholder="title=&quot;([^&quot;]+)&quot;" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="剧集URL正则" name="video_desc_page_config.episode_url_regex">
              <a-input v-model:value="formData.video_desc_page_config.episode_url_regex" placeholder="href=&quot;([^&quot;]+)&quot;" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="剧集URL是否绝对路径" name="video_desc_page_config.episode_url_is_absolute">
              <a-switch v-model:checked="formData.video_desc_page_config.episode_url_is_absolute" />
            </a-form-item>
          </a-col>
        </a-row>

        <!-- 播放页配置 -->
        <a-divider>播放页配置</a-divider>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="播放器CSS选择器" name="video_player_page_config.video_player_css_filter">
              <a-input v-model:value="formData.video_player_page_config.video_player_css_filter" placeholder="script" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="播放URL正则" name="video_player_page_config.video_player_url_regex">
              <a-input v-model:value="formData.video_player_page_config.video_player_url_regex" placeholder="var\\s+now\\s*=\\s*base64decode\\(\\s*&quot;([^&quot;]+)&quot;\\s*\\)" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="播放URL编码" name="video_player_page_config.video_player_url_encode">
              <a-select v-model:value="formData.video_player_page_config.video_player_url_encode" placeholder="base64">
                <a-select-option value="base64">Base64</a-select-option>
                <a-select-option value="url">URL编码</a-select-option>
                <a-select-option value="none">无编码</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <!-- 表单操作 -->
        <div class="form-actions">
          <a-space>
            <a-button @click="goBack">取消</a-button>
            <a-button type="primary" html-type="submit" :loading="saveLoading">
              {{ isEdit ? '更新' : '创建' }}
            </a-button>
          </a-space>
        </div>
      </a-form>
    </a-card>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'
import { message } from 'ant-design-vue'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const formRef = ref()
const saveLoading = ref(false)

const isEdit = computed(() => !!route.params.id)

const formData = ref({
  name: '',
  domain: '',
  status: 1,
  source_type: 0,
  search_config: {
    search_path: '/search.php',
    search_method: 'get',
    search_key_position: 'url',
    search_key: 'searchword',
    page_key_position: 'url',
    page_key: 'page',
    search_type_key_position: 'url',
    search_type_key: 'searchtype',
    total_count_css_filter: '',
    total_count_regex: '',
    current_page_css_filter: '',
    current_page_regex: '',
    video_card_css_filter: '',
    video_cover_image_css_filter: '',
    video_title_css_filter: '',
    video_detail_url_css_filter: '',
    video_player_url_css_filter: '',
    video_url_is_absolute: false,
    video_director_css_filter: '',
    video_director_regex: '',
    video_actor_css_filter: '',
    video_actor_regex: '',
    video_year_css_filter: '',
    video_year_regex: '',
    video_region_css_filter: '',
    video_region_regex: '',
    video_type_css_filter: '',
    video_type_regex: '',
    video_language_css_filter: '',
    video_description_css_filter: ''
  },
  video_desc_page_config: {
    video_card_css_filter: '',
    video_cover_image_css_filter: '',
    video_title_css_filter: '',
    video_player_url_css_filter: '',
    video_url_is_absolute: false,
    video_director_regex: '',
    video_director_css_filter: '',
    video_actor_css_filter: '',
    video_year_css_filter: '',
    video_area_css_filter: '',
    video_type_css_filter: '',
    video_language_css_filter: '',
    video_description_css_filter: '',
    source_card_css_filter: '',
    source_name_css_filter: '',
    episode_list_css_filter: '',
    episode_name_regex: '',
    episode_url_regex: '',
    episode_url_is_absolute: false
  },
  video_player_page_config: {
    video_player_css_filter: '',
    video_player_url_regex: '',
    video_player_url_encode: 'base64'
  }
})

const rules = {
  name: [{ required: true, message: '请输入站点名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入站点域名', trigger: 'blur' }],
  status: [{ required: true, message: '请选择站点状态', trigger: 'change' }]
}

const fetchVideoSourceDetail = async (id: string) => {
  if (!authStore.token) return
  
  try {
    const response = await videoSourceAPI.getVideoSourceDetail(authStore.token, id)
    if (response.code === 0) {
      formData.value = { ...response.data }
    } else {
      message.error(response.message || '获取视频源详情失败')
      goBack()
    }
  } catch (err: any) {
    message.error(err.message || '网络错误')
    goBack()
  }
}

const handleSave = async () => {
  if (!authStore.token) return
  
  try {
    await formRef.value?.validate()
  } catch (err) {
    return
  }
  
  saveLoading.value = true
  
  try {
    const response = await videoSourceAPI.saveVideoSource(authStore.token, formData.value)
    if (response.code === 0) {
      message.success(isEdit.value ? '更新成功' : '创建成功')
      goBack()
    } else {
      message.error(response.message || '保存失败')
    }
  } catch (err: any) {
    message.error(err.message || '网络错误')
  } finally {
    saveLoading.value = false
  }
}

const goBack = () => {
  router.push('/video-source-management')
}

onMounted(() => {
  if (isEdit.value && route.params.id) {
    fetchVideoSourceDetail(route.params.id as string)
  }
})
</script>

@import './VideoSourceEditView.css';
