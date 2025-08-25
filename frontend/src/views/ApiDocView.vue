<template>
  <AppLayout page-title="API接口文档">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>API接口文档</h2>
          <p>详细的接口调用说明和参数文档</p>
        </div>
      </template>

      <!-- 端口信息 -->
      <div class="port-info-section">
        <a-alert
          message="当前后端服务端口信息"
          description="后端HTTP服务正在运行，您可以使用以下端口进行API调用"
          type="info"
          show-icon
          class="port-alert"
        />
        <div class="port-details">
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="IPv4端口">
              <a-tag color="blue">{{ serverPort }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="IPv6端口">
              <a-tag color="green">{{ serverPort }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="基础URL">
              <code>http://localhost:{{ serverPort }}</code>
            </a-descriptions-item>
            <a-descriptions-item label="状态">
              <a-tag color="success">运行中</a-tag>
            </a-descriptions-item>
          </a-descriptions>
        </div>
      </div>

      <!-- 接口分类 -->
      <div class="api-sections">
        <!-- 系统配置接口 -->
        <div class="api-section">
          <h3>系统配置接口</h3>
          <a-collapse v-model:activeKey="activeKeys" accordion>
            <a-collapse-panel key="config" header="获取系统配置">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/config</code>
                  <a-tag color="blue">无需认证</a-tag>
                </div>
                <div class="api-description">
                  <p>获取系统配置信息，包括登录要求和环境信息</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <p>无</p>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "require_login": false,
    "env": "dev"
  }
}</code></pre>
                  
                  <h4>响应字段说明</h4>
                  <a-descriptions :column="1" size="small">
                    <a-descriptions-item label="require_login">是否需要登录认证</a-descriptions-item>
                    <a-descriptions-item label="env">运行环境 (dev/prod)</a-descriptions-item>
                  </a-descriptions>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>

        <!-- 视频源管理接口 -->
        <div class="api-section">
          <h3>视频源管理接口</h3>
          <a-collapse v-model:activeKey="activeKeys" accordion>
            <a-collapse-panel key="video-source-list" header="获取视频源列表">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/video-source/list</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>获取所有视频源站点的列表信息</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <p>无</p>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "name": "站点名称",
      "domain": "https://example.com",
      "sort": 1,
      "status": 1,
      "source_type": 0,
      "engine_type": 0,
      "lua_script": "function search_video(keyword) return {} end",
      "js_script": ""
    }
  ]
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-detail" header="获取视频源详情">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/video-source/detail?id={id}</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>根据ID获取视频源的详细信息</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <a-descriptions :column="1" size="small">
                    <a-descriptions-item label="id">视频源ID (必填)</a-descriptions-item>
                  </a-descriptions>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "name": "站点名称",
    "domain": "https://example.com",
    "sort": 1,
    "status": 1,
    "source_type": 0,
    "engine_type": 0,
    "lua_script": "function search_video(keyword) return {} end",
    "js_script": ""
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-save" header="保存视频源">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/video-source/save</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>创建或更新视频源配置</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "id": "uuid-string", // 可选，更新时提供
  "name": "站点名称",
  "domain": "https://example.com",
  "source_type": 0, // 0-综合, 1-短剧, 2-电影, 3-电视剧, 4-综艺, 5-动漫, 6-纪录片, 7-其他
  "engine_type": 0, // 0-Lua, 1-JavaScript
  "sort": 1,
  "status": 1, // 0-禁用, 1-正常, 2-维护中, 3-不可用
  "lua_script": "function search_video(keyword) return {} end",
  "js_script": ""
}</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "message": "保存成功"
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-delete" header="删除视频源">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/video-source/delete</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>删除指定的视频源</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "id": "uuid-string"
}</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "message": "删除成功"
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-set-status" header="设置视频源状态">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/video-source/set-status</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>设置视频源的状态</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "id": "uuid-string",
  "status": 1 // 0-禁用, 1-正常, 2-维护中, 3-不可用
}</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "status": 1
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-check-status" header="检查视频源状态">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/video-source/check-status?id={id}</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>检查视频源站点的可用性状态</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <a-descriptions :column="1" size="small">
                    <a-descriptions-item label="id">视频源ID (必填)</a-descriptions-item>
                  </a-descriptions>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": 1 // 0-禁用, 1-正常, 2-维护中, 3-不可用
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-export" header="导出视频源配置">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/video-source/export</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>导出所有视频源配置为JSON文件</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <p>无</p>
                  
                  <h4>响应</h4>
                  <p>直接返回JSON文件，Content-Type: application/json</p>
                  <p>文件名: video-sources.json</p>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="video-source-import" header="导入视频源配置">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/video-source/import</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>导入视频源配置，支持增量导入（通过ID去重）</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>[
  {
    "id": "uuid-string",
    "name": "站点名称",
    "domain": "https://example.com",
    "source_type": 0,
    "engine_type": 0,
    "sort": 1,
    "status": 1,
    "lua_script": "function search_video(keyword) return {} end",
    "js_script": ""
  }
]</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "imported_count": 1,
    "message": "导入完成"
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>

        <!-- 脚本测试接口 -->
        <div class="api-section">
          <h3>脚本测试接口</h3>
          <a-collapse v-model:activeKey="activeKeys" accordion>
            <a-collapse-panel key="lua-test" header="Lua脚本测试">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/lua/test</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>测试Lua脚本的执行结果</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "script": "function search_video(keyword) return {} end"
}</code></pre>
                  
                  <h4>响应</h4>
                  <p>返回脚本执行结果，支持流式输出</p>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="js-test" header="JavaScript脚本测试">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/js/test</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>测试JavaScript脚本的执行结果</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "script": "function search_video(keyword) { return []; }"
}</code></pre>
                  
                  <h4>响应</h4>
                  <p>返回脚本执行结果，支持流式输出</p>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="lua-test-sse" header="Lua脚本高级测试(SSE)">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/lua/test-sse</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>使用Server-Sent Events进行Lua脚本高级测试</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "script": "function search_video(keyword) return {} end",
  "method": "search_video",
  "params": {
    "keyword": "测试关键词"
  }
}</code></pre>
                  
                  <h4>响应</h4>
                  <p>返回SSE事件流，包含执行日志和结果</p>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="js-test-sse" header="JavaScript脚本高级测试(SSE)">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/js/test-sse</code>
                  <a-tag color="orange">可选认证</a-tag>
                </div>
                <div class="api-description">
                  <p>使用Server-Sent Events进行JavaScript脚本高级测试</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "script": "function search_video(keyword) { return []; }",
  "method": "search_video",
  "params": {
    "keyword": "测试关键词"
  }
}</code></pre>
                  
                  <h4>响应</h4>
                  <p>返回SSE事件流，包含执行日志和结果</p>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>

        <!-- 用户管理接口 -->
        <div class="api-section">
          <h3>用户管理接口</h3>
          <a-collapse v-model:activeKey="activeKeys" accordion>
            <a-collapse-panel key="user-login" header="用户登录">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/user/login</code>
                  <a-tag color="blue">无需认证</a-tag>
                </div>
                <div class="api-description">
                  <p>用户登录认证</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "username": "用户名",
  "password": "密码"
}</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": "uuid-string",
      "username": "用户名",
      "nickname": "昵称",
      "role": "admin"
    }
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="user-register" header="用户注册">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="orange">POST</a-tag>
                  <code>/api/user/register</code>
                  <a-tag color="blue">无需认证</a-tag>
                </div>
                <div class="api-description">
                  <p>用户注册</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <pre><code>{
  "username": "用户名",
  "password": "密码",
  "nickname": "昵称"
}</code></pre>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "message": "注册成功"
  }
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>

        <!-- 历史记录接口 -->
        <div class="api-section">
          <h3>历史记录接口</h3>
          <a-collapse v-model:activeKey="activeKeys" accordion>
            <a-collapse-panel key="video-history" header="视频观看历史">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/history/video?user_id={user_id}</code>
                  <a-tag color="red">需要认证</a-tag>
                </div>
                <div class="api-description">
                  <p>获取用户的视频观看历史</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <a-descriptions :column="1" size="small">
                    <a-descriptions-item label="user_id">用户ID (必填)</a-descriptions-item>
                  </a-descriptions>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "video_id": "video-uuid",
      "video_title": "视频标题",
      "video_url": "https://example.com/video",
      "source_id": "source-uuid",
      "source_name": "站点名称",
      "watch_time": 3600,
      "progress": 0.5,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ]
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="search-history" header="搜索历史">
              <div class="api-detail">
                <div class="api-basic">
                  <a-tag color="green">GET</a-tag>
                  <code>/api/history/search?user_id={user_id}</code>
                  <a-tag color="red">需要认证</a-tag>
                </div>
                <div class="api-description">
                  <p>获取用户的搜索历史</p>
                </div>
                <div class="api-params">
                  <h4>请求参数</h4>
                  <a-descriptions :column="1" size="small">
                    <a-descriptions-item label="user_id">用户ID (必填)</a-descriptions-item>
                  </a-descriptions>
                  
                  <h4>响应示例</h4>
                  <pre><code>{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "keyword": "搜索关键词",
      "source_id": "source-uuid",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ]
}</code></pre>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>
      </div>

      <!-- 调用示例 -->
      <div class="api-examples">
        <h3>调用示例</h3>
        <a-tabs v-model:activeKey="activeTab">
          <a-tab-pane key="curl" tab="cURL">
            <pre><code># 获取系统配置
curl -X GET http://localhost:{{ serverPort }}/api/config

# 获取视频源列表
curl -X GET http://localhost:{{ serverPort }}/api/video-source/list

# 导出视频源配置
curl -X GET http://localhost:{{ serverPort }}/api/video-source/export -o video-sources.json

# 导入视频源配置
curl -X POST http://localhost:{{ serverPort }}/api/video-source/import \
  -H "Content-Type: application/json" \
  -d @video-sources.json

# 测试Lua脚本
curl -X POST http://localhost:{{ serverPort }}/api/lua/test \
  -H "Content-Type: application/json" \
  -d '{"script":"function search_video(keyword) return {} end"}'

# 带认证的请求
curl -X GET http://localhost:{{ serverPort }}/api/video-source/list \
  -H "Authorization: Bearer your-jwt-token"</code></pre>
          </a-tab-pane>
          
          <a-tab-pane key="javascript" tab="JavaScript">
            <pre><code>// 获取系统配置
const response = await fetch('http://localhost:{{ serverPort }}/api/config');
const config = await response.json();

// 获取视频源列表
const response = await fetch('http://localhost:{{ serverPort }}/api/video-source/list');
const videoSources = await response.json();

// 导出视频源配置
const response = await fetch('http://localhost:{{ serverPort }}/api/video-source/export');
const blob = await response.blob();
const url = window.URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = 'video-sources.json';
a.click();

// 导入视频源配置
const response = await fetch('http://localhost:{{ serverPort }}/api/video-source/import', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(videoSourcesData)
});
const result = await response.json();

// 带认证的请求
const response = await fetch('http://localhost:{{ serverPort }}/api/video-source/list', {
  headers: {
    'Authorization': 'Bearer your-jwt-token'
  }
});</code></pre>
          </a-tab-pane>
          
          <a-tab-pane key="python" tab="Python">
            <pre><code>import requests

base_url = f"http://localhost:{{ serverPort }}"

# 获取系统配置
response = requests.get(f"{base_url}/api/config")
config = response.json()

# 获取视频源列表
response = requests.get(f"{base_url}/api/video-source/list")
video_sources = response.json()

# 导出视频源配置
response = requests.get(f"{base_url}/api/video-source/export")
with open('video-sources.json', 'wb') as f:
    f.write(response.content)

# 导入视频源配置
with open('video-sources.json', 'r') as f:
    data = f.read()
response = requests.post(f"{base_url}/api/video-source/import", 
                        data=data, 
                        headers={'Content-Type': 'application/json'})
result = response.json()

# 带认证的请求
headers = {'Authorization': 'Bearer your-jwt-token'}
response = requests.get(f"{base_url}/api/video-source/list", headers=headers)</code></pre>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-card>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getWailsServerPort } from '@/utils/api'
import AppLayout from '@/components/AppLayout.vue'

const serverPort = ref<number>(0)
const activeKeys = ref<string[]>([])
const activeTab = ref<string>('curl')

const loadServerPort = async () => {
  try {
    serverPort.value = await getWailsServerPort()
  } catch (error) {
    console.error('获取服务器端口失败:', error)
    serverPort.value = 0
  }
}

onMounted(() => {
  loadServerPort()
})
</script>

<style scoped>
.content-card {
  margin: 20px;
}

.card-header h2 {
  text-align: center;
  margin: 0 0 8px 0;
  color: #1e293b;
  font-size: 24px;
  font-weight: 600;
}

.card-header p {
  text-align: center;
  margin: 0;
  color: #64748b;
}

.port-info-section {
  margin-bottom: 30px;
}

.port-alert {
  margin-bottom: 16px;
}

.port-details {
  margin-top: 16px;
}

.api-sections {
  margin-bottom: 30px;
}

.api-section {
  margin-bottom: 24px;
}

.api-section h3 {
  color: #1e293b;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e2e8f0;
}

.api-detail {
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
  margin-top: 8px;
}

.api-basic {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.api-basic code {
  background: #e2e8f0;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
}

.api-description {
  margin-bottom: 16px;
}

.api-description p {
  margin: 0;
  color: #475569;
}

.api-params h4 {
  color: #1e293b;
  font-size: 16px;
  font-weight: 600;
  margin: 16px 0 8px 0;
}

.api-params pre {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.api-params code {
  background: #1e293b;
  color: #e2e8f0;
  padding: 2px 4px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
}

.api-examples {
  margin-top: 30px;
}

.api-examples h3 {
  color: #1e293b;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}

.api-examples pre {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}

:deep(.ant-collapse-header) {
  font-weight: 600 !important;
  color: #1e293b !important;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 600;
  color: #475569;
}

:deep(.ant-tabs-tab) {
  font-weight: 500;
}
</style>
