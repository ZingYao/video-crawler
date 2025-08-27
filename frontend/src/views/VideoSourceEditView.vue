<template>
  <AppLayout :page-title="isEdit ? '编辑视频源' : '添加视频源'">
    <div class="page-wrap">
      <a-card class="content-card">
        <template #title>
          <div class="card-header">
            <div class="header-left">
              <a-button class="back-btn" @click="handleBack" type="text">
                <template #icon>
                  <ArrowLeftOutlined />
                </template>
                返回
              </a-button>
              <h2>{{ isEdit ? '编辑视频源' : '添加视频源' }}</h2>
            </div>
            <div class="header-actions">
              <a-button type="primary" class="teal-btn" @click="handleSave" :loading="saveLoading">
                {{ isEdit ? '保存' : '创建' }}
              </a-button>
            </div>
          </div>
        </template>

        <a-form ref="formRef" :model="formData" :rules="rules" layout="vertical" class="video-source-form" @finish="handleSave">
          <a-form-item label="资源类型" name="source_type">
            <a-select
              v-model:value="formData.source_type"
              :options="sourceTypeOptions"
              placeholder="请选择资源类型"
              style="width: 100%"
            />
          </a-form-item>
          <a-form-item label="脚本类型" name="engine_type">
            <a-select v-model:value="formData.engine_type" :options="engineTypeOptions" style="width: 100%" />
          </a-form-item>
          <a-form-item label="站点名称" name="name">
            <a-input v-model:value="formData.name" placeholder="请输入站点名称，例如：示例影视站" />
          </a-form-item>
          <a-form-item label="站点域名" name="domain">
            <a-input v-model:value="formData.domain" placeholder="请输入站点域名，如：http://example.com" />
          </a-form-item>
          <a-form-item label="排序值" name="sort">
            <a-input-number v-model:value="formData.sort" :min="0" :max="9999" style="width: 100%" placeholder="数字越大排序越靠前" />
          </a-form-item>
          <a-form-item label="状态" name="status">
            <a-select v-model:value="formData.status" :options="statusOptions" style="width: 100%" />
          </a-form-item>
          <a-form-item label="Lua脚本" name="lua_script">
            <a-textarea
              v-model:value="formData.lua_script"
              :rows="20"
              placeholder="请输入Lua脚本内容"
              style="font-family: 'Courier New', monospace;"
            />
          </a-form-item>
        </a-form>
      </a-card>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'
import { message } from 'ant-design-vue'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const saveLoading = ref(false)

const isEdit = computed(() => !!route.params.id)

const formData = ref({
  id: '',
  name: '',
  domain: '',
  source_type: 0,
  sort: 0,
  engine_type: 0,
  status: 0,
  lua_script: ''
})

const rules = {
  source_type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
  engine_type: [{ required: true, message: '请选择脚本类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入站点名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入站点域名', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序值', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

const sourceTypeOptions = [
  { label: '综合', value: 0 },
  { label: '电影', value: 1 },
  { label: '电视剧', value: 2 },
  { label: '动漫', value: 3 },
  { label: '综艺', value: 4 },
  { label: '纪录片', value: 5 }
]

const engineTypeOptions = [
  { label: 'Lua', value: 0 },
  { label: 'JavaScript', value: 1 }
]

const statusOptions = [
  { label: '禁用', value: 0 },
  { label: '正常', value: 1 },
  { label: '维护中', value: 2 },
  { label: '不可用', value: 3 }
]

const handleBack = () => {
  router.back()
}

const handleSave = async () => {
  try {
    await formRef.value.validate()
    saveLoading.value = true
    
    const token = authStore.token!
    const response = await videoSourceAPI.saveVideoSource(token, formData.value)
    
    if (response?.code === 0) {
      message.success(isEdit.value ? '保存成功' : '创建成功')
      router.push('/video-source-management')
    } else {
      message.error(response?.message || '保存失败')
    }
  } catch (error: any) {
    message.error(error.message || '保存失败')
  } finally {
    saveLoading.value = false
  }
}

const loadVideoSource = async () => {
  if (!isEdit.value) return
  
  try {
    const token = authStore.token!
    const response = await videoSourceAPI.getVideoSourceDetail(token, route.params.id as string)
    
    if (response?.code === 0 && response.data) {
      formData.value = { ...response.data }
    } else {
      message.error('加载视频源信息失败')
      router.push('/video-source-management')
    }
  } catch (error: any) {
    message.error(error.message || '加载失败')
    router.push('/video-source-management')
  }
}

onMounted(() => {
  loadVideoSource()
})
</script>

<style scoped>
.page-wrap {
  padding: 20px;
}

.content-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.back-btn {
  color: #666;
}

.header-left h2 {
  margin: 0;
  color: #333;
}

.teal-btn {
  background: #10b981;
  border-color: #10b981;
}

.teal-btn:hover {
  background: #059669;
  border-color: #059669;
}

.video-source-form {
  margin-top: 20px;
}

@media (max-width: 768px) {
  .page-wrap {
    padding: 10px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .header-actions {
    width: 100%;
    display: flex;
    justify-content: flex-end;
  }
}
</style>

