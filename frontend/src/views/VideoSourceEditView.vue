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
          <a-input-number 
            v-model:value="formData.sort" 
            placeholder="请输入排序值，数字越大越靠前" 
            :min="0" 
            :max="9999"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status" style="width: 100%">
            <a-select-option :value="0">禁用</a-select-option>
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="2">维护中</a-select-option>
            <a-select-option :value="3">不可用</a-select-option>
          </a-select>
        </a-form-item>

          <div class="editor-logs-wrap" :style="gridStyle">
            <div class="editor-panel">
              <div class="panel-title">
                <div class="title-left">
                  <span class="title-text">{{ formData.engine_type === 1 ? 'JavaScript 脚本' : 'Lua 脚本' }}</span>
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
                  :language="formData.engine_type === 1 ? 'javascript' : 'lua'"
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
      <component :is="formData.engine_type === 1 ? JSDocs : LuaDocs" v-model:open="docsOpen" />
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
import JSDocs from '@/components/JSDocs.vue'
import { defaultTemplateLua, defaultDemo, defaultTemplateJS, demoTemplateJS } from '@/constants/scriptTemplates'

import MonacoEditor, { loader } from '@guolao/vue-monaco-editor'

// 配置 Monaco 静态资源路径（使用绝对地址，避免 Worker 环境相对路径解析失败）
const vsBase = `${window.location.origin}/monaco/vs`
loader.config({ paths: { vs: vsBase } })
;(window as any).MonacoEnvironment = {
  baseUrl: `${window.location.origin}/monaco/`,
  getWorkerUrl: function (_moduleId: string, _label: string) {
    return `${window.location.origin}/monaco/vs/base/worker/workerMain.js`
  }
}

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

const formData = ref({ id: '', name: '', domain: '', source_type: 0, sort: 0, engine_type: 0, status: 0 })

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
  { label: '短剧', value: 1 },
  { label: '电影', value: 2 },
  { label: '电视剧', value: 3 },
  { label: '综艺', value: 4 },
  { label: '动漫', value: 5 },
  { label: '纪录片', value: 6 },
  { label: '其他', value: 7 },
]

const engineTypeOptions = [
  { label: 'Lua', value: 0 },
  { label: 'JavaScript', value: 1 },
]

// 自定义 Monaco 主题（浅色背景深色文字）
const monacoTheme = ref('light-high-contrast')
const defineLightHighContrastTheme = async () => {
  if (!monaco) {
    monaco = await loader.init()
  }
  monaco.editor.defineTheme('light-high-contrast', {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'comment', foreground: '008000', fontStyle: 'italic' },
      { token: 'keyword', foreground: '0000FF', fontStyle: 'bold' },
      { token: 'string', foreground: 'A31515' },
      { token: 'number', foreground: '098658' },
      { token: 'operator', foreground: '000000' },
      { token: 'function', foreground: '795E26' },
      { token: 'variable', foreground: '001080' },
      { token: 'type', foreground: '267f99' },
      { token: 'constant', foreground: '0070C1' },
      { token: 'punctuation', foreground: '000000' },
      { token: 'delimiter', foreground: '000000' },
      { token: 'identifier', foreground: '001080' }
    ],
    colors: {
      'editor.background': '#00000000',
      'editor.foreground': '#000000',
      'editorLineNumber.foreground': '#2B91AF',
      'editorLineNumber.activeForeground': '#000000',
      'editorGutter.background': '#00000000',
      'editor.selectionBackground': '#ADD6FF',
      'editor.inactiveSelectionBackground': '#E5EBF1',
      'editorCursor.foreground': '#000000',
      'editorLineHighlightBackground': '#F7F7F7',
      'editorLineHighlightBorder': '#E7E7E7',
      'minimap.background': '#00000000',
      'minimap.selectionBackground': '#ADD6FF',
      'scrollbarSlider.background': '#C1C1C1',
      'scrollbarSlider.hoverBackground': '#A8A8A8',
      'scrollbarSlider.activeBackground': '#787878',
      'editorWidget.background': '#F3F3F3',
      'editorWidget.border': '#C8C8C8',
      'editorSuggestWidget.background': '#F3F3F3',
      'editorSuggestWidget.border': '#C8C8C8',
      'editorSuggestWidget.selectedBackground': '#C9D0D9'
    }
  })
  monaco.editor.setTheme('light-high-contrast')
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

// 草稿相关
const DRAFT_KEY = 'video_source_draft'
const DRAFT_INTERVAL = 3000 // 3秒自动保存
let draftTimer: number | null = null
const scriptContent = ref<string>(defaultTemplateLua)
const editorRef = ref<any>(null)
// 缓存两种脚本各自的代码，切换语言时优先使用已有内容
const luaCode = ref<string>('')
const jsCode = ref<string>('')

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
    engine_type: formData.value.engine_type,
    sort: formData.value.sort,
    status: formData.value.status,
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
  return draft.name !== formData.value.name || draft.domain !== formData.value.domain || draft.source_type !== formData.value.source_type || draft.engine_type !== formData.value.engine_type || draft.sort !== formData.value.sort || draft.status !== formData.value.status || draft.script !== scriptContent.value
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

const openDocs = () => {
  console.log('[Docs] open clicked. engine_type=', formData.value.engine_type)
  docsOpen.value = true
}
watch(docsOpen, (v) => {
  console.log('[Docs] docsOpen changed:', v, 'engine_type=', formData.value.engine_type)
})
const onEditorMount = async (editor: any) => {
  editorRef.value = editor
  await defineLightHighContrastTheme()
}
const resetDemo = () => { scriptContent.value = (formData.value.engine_type === 1 ? demoTemplateJS : defaultDemo) }
const onFillDefault = () => {
  const current = scriptContent.value?.trim() || ''
  const tpl = formData.value.engine_type === 1 ? defaultTemplateJS : defaultTemplateLua
  if (current.length > 0 && current !== tpl.trim()) {
    Modal.confirm({ title: '确认覆盖当前脚本？', content: '填充默认代码将覆盖编辑器中的现有内容。', okText: '覆盖', cancelText: '取消', onOk: () => { scriptContent.value = tpl } })
  } else {
    scriptContent.value = tpl
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
  const demo = formData.value.engine_type === 1 ? demoTemplateJS : defaultDemo
  if (current.length > 0 && current !== demo.trim()) {
    Modal.confirm({ title: '确认覆盖当前脚本？', content: '填充 Demo 将覆盖编辑器中的现有内容。', okText: '覆盖', cancelText: '取消', onOk: () => resetDemo() })
  } else { resetDemo() }
}

// 同步当前脚本到对应语言的缓存
watch(scriptContent, (val) => {
  const code = String(val ?? '')
  if (formData.value.engine_type === 1) {
    jsCode.value = code
  } else {
    luaCode.value = code
  }
})

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
      formData.value.engine_type = draft.engine_type ?? 0
      formData.value.sort = draft.sort || 0
      formData.value.status = draft.status ?? 0
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
      formData.value.engine_type = data.engine_type ?? 0
      formData.value.sort = data.sort || 0
      formData.value.status = data.status ?? 0
      // 加载Lua脚本到编辑器
      if (formData.value.engine_type === 1) {
        // JS 脚本
        if (typeof (data as any).js_script === 'string' && (data as any).js_script.trim()) {
          jsCode.value = (data as any).js_script
          scriptContent.value = jsCode.value
        } else {
          jsCode.value = '// TODO: implement search_video, get_video_detail, get_play_video_detail in JS\n'
          scriptContent.value = jsCode.value
        }
      } else {
        if (typeof data.lua_script === 'string') {
          luaCode.value = data.lua_script.trim() ? data.lua_script : defaultTemplateLua
          scriptContent.value = luaCode.value
        }
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
      engine_type: formData.value.engine_type,
      sort: formData.value.sort,
      status: formData.value.status,
      lua_script: formData.value.engine_type === 0 ? scriptContent.value : '',
      js_script: formData.value.engine_type === 1 ? scriptContent.value : ''
    }
    // 必要方法校验
    const code = String(scriptContent.value || '')
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
    const isJS = formData.value.engine_type === 1
    const endpoint = isJS ? '/api/js/test' : '/api/lua/test'
    const resp = await fetch(`${window.location.origin}${endpoint}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${authStore.token}` }, body: JSON.stringify({ script: scriptContent.value }),
    })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    if (!resp.body) throw new Error('浏览器不支持流式响应')
    const reader = resp.body.getReader(); const decoder = new TextDecoder()
    while (true) { const { done, value } = await reader.read(); if (done) break; outputText.value += decoder.decode(value) }
  } catch (err: any) { outputText.value += `\n[ERROR] ${err?.message || String(err)}` }
  finally { debugLoading.value = false }
}

// 切换脚本语言类型时：若已有对应类型代码则使用该代码，否则才填充 Demo，并清理草稿
watch(() => formData.value.engine_type, (val, oldVal) => {
  if (val === oldVal) return
  if (val === 1) {
    if (jsCode.value && jsCode.value.trim()) {
      scriptContent.value = jsCode.value
      message.success('已切换到 JavaScript（使用现有代码）')
    } else {
      scriptContent.value = demoTemplateJS
      clearDraft()
      message.success('已切换到 JavaScript（填充 Demo 代码）')
    }
  } else {
    if (luaCode.value && luaCode.value.trim()) {
      scriptContent.value = luaCode.value
      message.success('已切换到 Lua（使用现有代码）')
    } else {
      scriptContent.value = defaultDemo
      clearDraft()
      message.success('已切换到 Lua（填充 Demo 代码）')
    }
  }
})

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
