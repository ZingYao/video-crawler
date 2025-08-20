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
            <a-button type="primary" class="teal-btn" @click="handleSave" :loading="saveLoading">{{ isEdit ? '保存' : '创建' }}</a-button>
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
            <a-form-item label="站点名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入站点名称，例如：示例影视站" />
            </a-form-item>
        <a-form-item label="站点域名" name="domain">
          <a-input v-model:value="formData.domain" placeholder="请输入站点域名，如：http://example.com" />
        </a-form-item>
        <a-form-item label="排序值" name="sort">
          <a-input-number 
            v-model:value="formData.sort" 
            placeholder="请输入排序值，数字越大越靠前" 
            :min="0" 
            :max="9999"
            style="width: 100%"
          />
            </a-form-item>

          <div class="editor-logs-wrap" :style="gridStyle">
            <div class="editor-panel">
              <div class="panel-title">
                <div class="title-left">
                  <span class="title-text">Lua 脚本</span>
                </div>
                <div class="title-actions">
                  <a-button class="teal-btn" size="small" @click="onFillDefault">默认代码</a-button>
                  <a-button class="teal-btn" size="small" @click="openDocs">打开文档</a-button>
                  <a-button class="teal-btn" size="small" @click="onFillDemo">填充完整 Demo</a-button>
                  <a-button class="teal-btn" size="small" :loading="debugLoading" @click="runScript">脚本调试</a-button>
                </div>
              </div>
              <div class="editor-gradient">
                <MonacoEditor
                  class="monaco"
                  :theme="monacoTheme"
                  language="lua"
                  :options="monacoOptions"
                  v-model:value="scriptContent"
                  @mount="onEditorMount"
                />
              </div>
            </div>

            <!-- 可拖拽分隔条（仅PC显示） -->
            <div class="split-gutter" v-show="isPc" @mousedown="startDrag"></div>

            <div class="logs-panel side">
              <div class="panel-title">
                <div class="title-left"><span class="title-text">调试输出</span></div>
                <div class="title-actions">
                  <a-button class="teal-btn" size="small" @click="clearLogs">清空</a-button>
                </div>
              </div>
              <div class="logs-box gradient scrollable" ref="logsRef">
                <div v-for="(line, idx) in coloredLines" :key="idx" class="log-line" :class="line.type">
                  {{ line.text }}
                </div>
              </div>
            </div>
        </div>
      </a-form>
    </a-card>
      <LuaDocs v-model:open="docsOpen" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick, defineAsyncComponent } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'
import { message, Modal } from 'ant-design-vue'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'
import AppLayout from '@/components/AppLayout.vue'
import LuaDocs from '@/components/LuaDocs.vue'
import { defaultTemplateLua, defaultDemo } from '@/constants/luaTemplates'

import MonacoEditor, { loader } from '@guolao/vue-monaco-editor'

// 配置 Monaco 静态资源路径
loader.config({ paths: { vs: '/monaco/vs' } })

let monaco: any = null

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const formRef = ref()
const saveLoading = ref(false)
const debugLoading = ref(false)
const outputText = ref('')
const docsOpen = ref(false)
const logsRef = ref<HTMLDivElement | null>(null)
// 已移除最大化功能

const isEdit = computed(() => !!route.params.id)

const formData = ref({ id: '', name: '', domain: '', source_type: 0, sort: 0 })

const rules = {
  source_type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入站点名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入站点域名', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序值', trigger: 'blur' }]
}

const sourceTypeOptions = [
  { label: '综合', value: 0 },
  { label: '短剧', value: 1 },
  { label: '电影', value: 2 },
  { label: '电视剧', value: 3 },
  { label: '综艺', value: 4 },
  { label: '动漫', value: 5 },
  { label: '纪录片', value: 6 },
  { label: '其他', value: 7 },
]

// 自定义 Monaco 主题（偏亮青绿）
const monacoTheme = ref('teal-light')
const defineTealTheme = async () => {
  if (!monaco) {
    monaco = await loader.init()
  }
  monaco.editor.defineTheme('teal-light', {
    base: 'vs-dark', inherit: true,
    rules: [ { token: '', foreground: 'F1FFFB' } ],
    colors: {
      'editor.background': '#00000000',
      'editor.foreground': '#0b2e2a',
      'editorLineNumber.foreground': '#0f766e',
      'editorLineNumber.activeForeground': '#064e3b',
      'editorGutter.background': '#00000000',
      'editor.selectionBackground': '#1cc8a066',
      'editor.inactiveSelectionBackground': '#12a88a55',
      'editorCursor.foreground': '#065f46',
      'editorLineHighlightBackground': '#10b98144',
      'minimap.background': '#00000000'
    }
  })
  monaco.editor.setTheme('teal-light')
}

const monacoOptions = {
  fontSize: 13,
  minimap: { enabled: false },
  smoothScrolling: true,
  scrollBeyondLastLine: false,
  wordWrap: 'on' as const,
  automaticLayout: true,
  readOnly: false,
  lineNumbers: (lineNumber: number) => String(lineNumber),
  renderLineHighlight: 'all' as const,
  stickyScroll: { enabled: false }, // 关闭顶部白色预览条
}

// 资源编辑默认模板（已抽离常量文件，但保留以避免引用失败时回退）
/*
const defaultTemplateLua = `-- 搜索视频结果结构体
local search_video_result = {
    ['source_id'] = '', -- 资源ID
    ['cover'] = '', -- 视频封面
    ['name'] = '', -- 视频名称
    ['url'] = '', -- 视频链接
    ['actor'] = '', -- 演员
    ['director'] = '', -- 导演
    ['release_date'] = '', -- 上映日期
    ['region'] = '', -- 地区
    ['language'] = '', -- 语言
    ['description'] = '', -- 描述
    ['score'] = '', -- 评分
}

-- 视频详情结构体
local video_detail_result = {
    ['source_id'] = '', -- 资源ID
    ['cover'] = '', -- 视频封面
    ['name'] = '', -- 视频名称
    ['url'] = '', -- 视频链接
    ['actor'] = '', -- 演员
    ['director'] = '', -- 导演
    ['release_date'] = '', -- 上映日期
    ['region'] = '', -- 地区
    ['language'] = '', -- 语言
    ['description'] = '', -- 描述
    ['score'] = '', -- 评分
    ['source'] = {
        ['name'] = '', -- 来源名称
        ['episodes'] = {
            ['name'] = '', -- 剧集名称
            ['url'] = '', -- 剧集链接
        }
    }
}

-- 视频播放详情
local play_video_detail = {
    ['source_id'] = '', -- 资源ID
    ['video_url'] = '', -- 视频链接
}

function search_video(search_content)
    -- 数组 Array(search_video_result)
    local result = {}
    local err = nil
    -- 完成这个代码
    return result, err
end

function get_video_detail(video_url)
    -- video_detail_result 结构体
    local result = {}
    local err = nil
    -- 完成这个代码
    return result, err
end

function get_play_video_detail(video_url)
    -- play_video_detail 结构体
    local result = {}
    local err = nil
    -- 完成这个代码
    return result, err
end
`

const defaultDemo = `-- 链式调用 Demo：请求页面，querySelector 并读取 attr / text / html
print('[Demo] 启动')
set_user_agent('Lua-Demo-Agent/1.0')
set_headers({ ['Accept'] = 'text/html' })

-- 获取并打印当前 UA
local current_ua = get_user_agent()
print('当前 User-Agent:', current_ua)

-- 1) 请求示例站点
local resp, reqErr = http_get('https://example.com')
if reqErr then
  log('请求错误:', reqErr)
else
  print('HTTP 状态码:', resp.status_code)

  -- 2) 解析 HTML
  local doc, perr = parse_html(resp.body)
  if perr then
    log('解析错误:', perr)
  else
    -- 3) 执行 querySelector（链式 select_one）
    local link, selErr = doc:select_one('a')
    if selErr then
      log('选择器错误:', selErr)
    else
      -- 4) 读取 attr / text / html 并打印
      local href, aerr = link:attr('href')
      if aerr then
        log('attr 错误:', aerr)
      else
        print('href 属性 =', href)
      end
      print('text 文本 =', link:text())
      print('inner HTML =', link:html())
    end
  end
end
print('[Demo] 完成')
`
*/

// 草稿相关
const DRAFT_KEY = 'video_source_draft'
const DRAFT_INTERVAL = 3000 // 3秒自动保存
let draftTimer: number | null = null
const scriptContent = ref<string>(defaultTemplateLua)
const editorRef = ref<any>(null)

// PC下可拖拽分隔条逻辑
const isPc = ref(true)
const GUTTER = 8
const MIN_LOGS = 260
const MAX_LOGS = 900
const SPLIT_KEY = 'video_source_editor_logs_width_pc'
const logsWidth = ref<number>(360)
const gridStyle = computed(() => {
  return isPc.value ? { gridTemplateColumns: `1fr ${GUTTER}px ${logsWidth.value}px` } : {}
})
let dragging = false
let startX = 0
let startLogs = 0
const applyStoredSplit = () => {
  try {
    const v = Number(localStorage.getItem(SPLIT_KEY) || '')
    if (!Number.isNaN(v) && v >= MIN_LOGS && v <= MAX_LOGS) logsWidth.value = v
  } catch {}
}
const saveSplit = () => {
  try { localStorage.setItem(SPLIT_KEY, String(logsWidth.value)) } catch {}
}
const startDrag = (e: MouseEvent) => {
  if (!isPc.value) return
  dragging = true
  startX = e.clientX
  startLogs = logsWidth.value
  document.addEventListener('mousemove', onDrag, { passive: false })
  document.addEventListener('mouseup', stopDrag, { passive: false })
}
const onDrag = (e: MouseEvent) => {
  if (!dragging) return
  e.preventDefault()
  const dx = e.clientX - startX
  // 向右拖：日志变窄；向左拖：日志变宽
  let next = startLogs - dx
  if (next < MIN_LOGS) next = MIN_LOGS
  if (next > MAX_LOGS) next = MAX_LOGS
  logsWidth.value = next
}
const stopDrag = () => {
  if (!dragging) return
  dragging = false
  document.removeEventListener('mousemove', onDrag as any)
  document.removeEventListener('mouseup', stopDrag as any)
  saveSplit()
}

// 草稿管理函数
const saveDraft = () => {
  const draft = {
    name: formData.value.name,
    domain: formData.value.domain,
    source_type: formData.value.source_type,
    sort: formData.value.sort,
    script: scriptContent.value,
    timestamp: Date.now()
  }
  localStorage.setItem(DRAFT_KEY, JSON.stringify(draft))
}

const loadDraft = () => {
  const draftStr = localStorage.getItem(DRAFT_KEY)
  if (draftStr) {
    try {
      return JSON.parse(draftStr)
    } catch {
      return null
    }
  }
  return null
}

const clearDraft = () => {
  localStorage.removeItem(DRAFT_KEY)
}

const hasDraft = () => {
  return localStorage.getItem(DRAFT_KEY) !== null
}

const isDraftDifferent = (draft: any) => {
  return draft.name !== formData.value.name || draft.domain !== formData.value.domain || draft.source_type !== formData.value.source_type || draft.sort !== formData.value.sort || draft.script !== scriptContent.value
}

// 定时保存草稿
const startDraftTimer = () => {
  if (draftTimer) clearInterval(draftTimer)
  draftTimer = setInterval(saveDraft, DRAFT_INTERVAL)
}

const stopDraftTimer = () => {
  if (draftTimer) {
    clearInterval(draftTimer)
    draftTimer = null
  }
}

const handleBack = () => {
  // 如果有未保存的草稿，提示用户
  if (hasDraft()) {
    Modal.confirm({
      title: '确认离开',
      content: '您有未保存的草稿，确定要离开吗？',
      okText: '离开',
      cancelText: '取消',
      onOk: () => {
        clearDraft()
        router.push('/video-source-management')
      }
    })
  } else {
    router.push('/video-source-management')
  }
}

const openDocs = () => { docsOpen.value = true }
const onEditorMount = async (editor: any) => {
  editorRef.value = editor
  await defineTealTheme()
}
const resetDemo = () => { scriptContent.value = defaultDemo }
const onFillDefault = () => {
  const current = scriptContent.value?.trim() || ''
  if (current.length > 0 && current !== defaultTemplateLua.trim()) {
    Modal.confirm({ title: '确认覆盖当前脚本？', content: '填充默认代码将覆盖编辑器中的现有内容。', okText: '覆盖', cancelText: '取消', onOk: () => { scriptContent.value = defaultTemplateLua } })
  } else {
    scriptContent.value = defaultTemplateLua
  }
}

const clearLogs = () => {
  outputText.value = ''
  nextTick().then(() => {
    const el = logsRef.value
    if (el) el.scrollTop = 0
  })
}

// 最大化功能移除，相关重排逻辑一并删去

const onFillDemo = () => {
  const current = scriptContent.value?.trim() || ''
  if (current.length > 0 && current !== defaultDemo.trim()) {
    Modal.confirm({ title: '确认覆盖当前脚本？', content: '填充 Demo 将覆盖编辑器中的现有内容。', okText: '覆盖', cancelText: '取消', onOk: () => resetDemo() })
  } else { resetDemo() }
}

// 草稿恢复弹窗
const showDraftRestoreDialog = (draft: any) => {
  Modal.confirm({
    title: '发现草稿',
    content: '检测到您有未保存的草稿，是否要恢复？',
    okText: '恢复草稿',
    cancelText: '删除草稿',
    onOk: () => {
      formData.value.name = draft.name || ''
      formData.value.domain = draft.domain
      formData.value.source_type = draft.source_type ?? 0
      formData.value.sort = draft.sort || 0
      scriptContent.value = draft.script
      clearDraft()
      message.success('草稿已恢复')
    },
    onCancel: () => {
      Modal.confirm({
        title: '确认删除草稿',
        content: '删除后无法恢复，确定要删除草稿吗？',
        okText: '确定删除',
        cancelText: '取消',
        okType: 'danger',
        onOk: () => {
          clearDraft()
          message.success('草稿已删除')
        }
      })
    }
  })
}

const coloredLines = computed(() => {
  const lines = (outputText.value || '').split(/\r?\n/)
  return lines.filter(Boolean).map((t) => {
    if (t.startsWith('[ERROR]') || t.includes('[ERROR]')) return { type: 'err', text: t }
    if (t.startsWith('[INFO]') || t.includes('[INFO]')) return { type: 'info', text: t }
    if (t.startsWith('[LOG]') || t.includes('[LOG]')) return { type: 'log', text: t }
    if (t.startsWith('[PRINT]') || t.includes('[PRINT]')) return { type: 'print', text: t }
    return { type: 'plain', text: t }
  })
})

watch(outputText, async () => { await nextTick(); const el = logsRef.value; if (el) el.scrollTop = el.scrollHeight })

const fetchVideoSourceDetail = async (id: string) => {
  if (!authStore.token) return
  try {
    const response = await videoSourceAPI.getVideoSourceDetail(authStore.token, id)
    if ((response as any).code === 0) {
      const data = (response as any).data || {}
      formData.value.id = data.id || ''
      formData.value.name = data.name || ''
      formData.value.domain = data.domain || ''
      formData.value.source_type = data.source_type ?? 0
      formData.value.sort = data.sort || 0
      // 加载Lua脚本到编辑器
      if (typeof data.lua_script === 'string') {
        if (data.lua_script.trim()) scriptContent.value = data.lua_script
        else scriptContent.value = defaultTemplateLua
      }
      
      // 检查是否有草稿需要恢复
      const draft = loadDraft()
      if (draft && isDraftDifferent(draft)) {
        showDraftRestoreDialog(draft)
      }
    } else { message.error((response as any).message || '获取视频源详情失败') }
  } catch (err: any) { message.error(err.message || '网络错误') }
}

const handleSave = async () => {
  if (!authStore.token) return
  try { await formRef.value?.validate() } catch { return }
  saveLoading.value = true
  try {
    const payload: any = { 
      id: formData.value.id || '', 
      name: formData.value.name,
      domain: formData.value.domain,
      source_type: formData.value.source_type,
      sort: formData.value.sort,
      lua_script: scriptContent.value
    }
    // 必要方法校验
    const code = String(payload.lua_script || '')
    const missing: string[] = []
    if (!/\bfunction\s+search_video\s*\(/.test(code)) missing.push('search_video(search_content)')
    if (!/\bfunction\s+get_video_detail\s*\(/.test(code)) missing.push('get_video_detail(video_url)')
    if (!/\bfunction\s+get_play_video_detail\s*\(/.test(code)) missing.push('get_play_video_detail(video_url)')
    if (missing.length) {
      message.error('缺少必要方法：' + missing.join('、'))
      return
    }
    const response = await videoSourceAPI.saveVideoSource(authStore.token, payload)
    if ((response as any).code === 0) { 
      message.success(isEdit.value ? '保存成功' : '创建成功')
      // 保存成功后清除草稿
      clearDraft()
      // 创建成功后返回列表
      if (!isEdit.value) {
        router.push('/video-source-management')
      }
    }
    else { message.error((response as any).message || '保存失败') }
  } catch (err: any) { message.error(err.message || '网络错误') }
  finally { saveLoading.value = false }
}

const runScript = async () => {
  if (!authStore.token) { message.error('未登录，无法调试脚本'); return }
  outputText.value = ''
  debugLoading.value = true
  await nextTick(); logsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  try {
    const resp = await fetch(`${window.location.origin}/api/lua/test`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${authStore.token}` }, body: JSON.stringify({ script: scriptContent.value }),
    })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    if (!resp.body) throw new Error('浏览器不支持流式响应')
    const reader = resp.body.getReader(); const decoder = new TextDecoder()
    while (true) { const { done, value } = await reader.read(); if (done) break; outputText.value += decoder.decode(value) }
  } catch (err: any) { outputText.value += `\n[ERROR] ${err?.message || String(err)}` }
  finally { debugLoading.value = false }
}

onMounted(() => {
  // 启动定时保存草稿
  startDraftTimer()
  // 设备类型与分隔条恢复
  const mq = window.matchMedia('(max-width: 900px)')
  const syncPc = () => { isPc.value = !mq.matches }
  mq.addEventListener?.('change', syncPc)
  isPc.value = !mq.matches
  if (isPc.value) applyStoredSplit()
  // 监听页面离开事件
  
  if (isEdit.value && route.params.id) {
    fetchVideoSourceDetail(route.params.id as string)
  } else {
    // 新建模式下检查草稿
    const draft = loadDraft()
    if (draft) {
      showDraftRestoreDialog(draft)
    }
  }
  
  // 监听页面离开事件
  window.addEventListener('beforeunload', saveDraft)
})

onUnmounted(() => {
  // 清理定时器
  stopDraftTimer()
  // 移除页面离开事件监听
  window.removeEventListener('beforeunload', saveDraft)
  document.removeEventListener('mousemove', onDrag as any)
  document.removeEventListener('mouseup', stopDrag as any)
})

// 全局快捷键：F5 运行脚本；屏蔽 ⌘S / Ctrl+S
document.addEventListener('keydown', (e: KeyboardEvent) => {
  // 仅处理保存与运行，不影响其它系统/编辑器快捷键（如 ⌘A / ⌘Z）
  const isSave = (e.key && e.key.toLowerCase() === 's') && (e.metaKey || e.ctrlKey)
  if (isSave) { e.preventDefault(); e.stopPropagation(); return }
  if (e.key === 'F5') { e.preventDefault(); runScript(); return }
}, { passive: false })
</script>

<style scoped>
.page-wrap { --teal: #10b981; --teal-hover: #34d399; --teal-active: #059669; padding: 12px; }
.editor-logs-wrap { display: grid; grid-template-columns: 1fr 8px 360px; gap: 0; align-items: stretch; }
.card-header { display: flex; align-items: center; justify-content: space-between; min-width: 0; }
.header-left { display: flex; align-items: center; gap: 12px; }
.back-btn { color: var(--teal-active); font-weight: 600; padding: 4px 8px; border-radius: 6px; transition: all 0.2s; }
.back-btn:hover { color: var(--teal-hover); background: rgba(16, 185, 129, 0.08); }
.back-btn:active { color: #047857; background: rgba(4, 120, 87, 0.12); }
.header-actions > * { margin-left: 8px; }
.editor-panel, .logs-panel { background: transparent; border: 1px solid #20c7ab; border-radius: 8px; overflow: hidden; margin-bottom: 12px; min-width: 0; }
.editor-panel { display: flex; flex-direction: column; height: 620px; }
.split-gutter { cursor: col-resize; background: linear-gradient(180deg, #14b8a61a, #10b9811a); border-radius: 6px; }
.split-gutter:hover { background: linear-gradient(180deg, #14b8a638, #10b98138); }
.panel-title { height: 36px; display: flex; align-items: center; justify-content: space-between; padding: 0 10px; color: #0a2f28; font-weight: 800; background: linear-gradient(90deg, #99f6e4 0%, #34d399 100%); border-bottom: 1px solid #20c7ab; font-size: 13px; letter-spacing: 0.5px; }
.title-left { display: flex; align-items: center; gap: 10px; min-width: 0; }
.title-actions { display: flex; gap: 8px; align-items: center; }
.title-text { font-weight: 800; }
.editor-gradient { background: linear-gradient(135deg, #99f6e4 0%, #5eead4 35%, #34d399 70%, #2dd4bf 100%); padding: 0; overflow-x: hidden; min-width: 0; flex: 1; min-height: 0; }
.monaco { height: 100%; min-height: 0; }
::v-deep(.monaco-editor),
::v-deep(.monaco-editor .margin),
::v-deep(.monaco-editor .monaco-editor-background) { background: transparent !important; max-width: 100% !important; }
/* 避免强制设置内部滚动容器宽度导致布局抖动或无限重排 */
::v-deep(.monaco-editor .overflow-guard),
::v-deep(.monaco-editor .editor-scrollable) { max-width: 100% !important; }
/* 可能的白色顶部预览/阴影 */
::v-deep(.monaco-editor .sticky-scroll),
::v-deep(.monaco-editor .monaco-sticky-scroll),
::v-deep(.monaco-editor .editor-sticky-scroll),
::v-deep(.monaco-scrollable-element .scroll-decoration) { background: transparent !important; box-shadow: none !important; }

.logs-box.gradient { background: linear-gradient(135deg, #99f6e4 0%, #5eead4 35%, #34d399 70%, #2dd4bf 100%); border-top: 1px solid #20c7ab; min-width: 0; }
.logs-panel.side { display: flex; flex-direction: column; height: 620px; overflow: hidden; }
.logs-box.scrollable { flex: 1; min-height: 0; overflow: auto; }
.log-line { font-family: Menlo, Monaco, Consolas, 'Courier New', monospace; font-size: 12px; color: #083942; white-space: pre-wrap; word-break: break-word; line-height: 1.6; }
.log-line.print { color: #065f46; }
.log-line.log { color: #075985; }
.log-line.info { color: #0e7490; }
.log-line.err { color: #7f1d1d; }
.log-line.plain { color: #083942; }

/* 统一按钮为主题绿色（仅限本页） */
.teal-btn, ::v-deep(.title-actions .ant-btn-primary),
::v-deep(.header-actions .ant-btn-primary) {
  background: linear-gradient(180deg, var(--teal) 0%, var(--teal-active) 100%) !important;
  border-color: var(--teal-active) !important;
  color: #fff !important;
}
.title-actions .teal-btn {
  background: linear-gradient(180deg, #10b981 0%, #059669 100%) !important;
  border: 1px solid #047857 !important;
  color: #fff !important;
  /* 强制按钮文本为白色 */
  &,
  & span,
  & * { color: #fff !important; }
  border-radius: 6px !important;
  font-weight: 700 !important;
  box-shadow: 0 2px 6px rgba(4, 120, 87, 0.25) !important;
}
.teal-btn:hover, ::v-deep(.title-actions .ant-btn-primary:hover),
::v-deep(.header-actions .ant-btn-primary:hover) {
  background: linear-gradient(180deg, var(--teal-hover) 0%, var(--teal) 100%) !important;
  border-color: var(--teal-hover) !important;
}
.title-actions .teal-btn:hover {
  background: linear-gradient(180deg, #34d399 0%, #10b981 100%) !important;
  border-color: #34d399 !important;
  box-shadow: 0 3px 8px rgba(4, 120, 87, 0.35) !important;
}
.teal-btn:active, ::v-deep(.title-actions .ant-btn-primary:active),
::v-deep(.header-actions .ant-btn-primary:active) {
  background: linear-gradient(180deg, var(--teal-active) 0%, #047857 100%) !important;
  border-color: #047857 !important;
}
.title-actions .teal-btn:active { box-shadow: 0 1px 4px rgba(4, 120, 87, 0.3) !important; }
/* 若有默认按钮类型，使用主题绿色描边 */
::v-deep(.title-actions .ant-btn-default),
::v-deep(.header-actions .ant-btn-default) {
  color: var(--teal-active) !important;
  border-color: var(--teal) !important;
}
::v-deep(.title-actions .ant-btn-default:hover),
::v-deep(.header-actions .ant-btn-default:hover) {
  color: #064e3b !important;
  border-color: var(--teal-hover) !important;
  background: rgba(16, 185, 129, 0.08) !important;
}

/* 防止在侧边栏展开时产生横向溢出 */
.page-wrap, .content-card, .video-source-form { max-width: 100%; width: 100%; box-sizing: border-box; overflow-x: hidden; }
.monaco { width: 100%; }

/* 移动端自适应（不影响 PC） */
@media (max-width: 900px) {
  .page-wrap { padding: 8px; }
  .editor-logs-wrap { grid-template-columns: 1fr; gap: 10px; }
  .split-gutter { display: none; }
  .panel-title { height: 32px; font-size: 12px; }
  .editor-panel { height: 48vh; }
  .logs-panel.side { height: 36vh; }
  .logs-box.scrollable { overflow: auto; }
  .monaco { height: 100%; min-height: 0; }
}
</style>
