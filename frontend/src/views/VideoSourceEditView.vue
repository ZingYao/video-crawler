<template>
  <AppLayout :page-title="isEdit ? '编辑视频源' : '添加视频源'">
    <div class="page-wrap">
    <a-card class="content-card">
      <template #title>
        <div class="card-header">
          <h2>{{ isEdit ? '编辑视频源' : '添加视频源' }}</h2>
            <div class="header-actions">
              <a-button type="primary" @click="handleSave" :loading="saveLoading">{{ isEdit ? '保存' : '创建' }}</a-button>
            </div>
        </div>
      </template>

        <a-form ref="formRef" :model="formData" :rules="rules" layout="vertical" class="video-source-form" @finish="handleSave">
        <a-form-item label="站点域名" name="domain">
          <a-input v-model:value="formData.domain" placeholder="请输入站点域名，如：http://example.com" />
        </a-form-item>

          <div class="editor-panel">
            <div class="panel-title">
              <div class="title-left">
                <span class="title-text">Lua 脚本</span>
              </div>
              <div class="title-actions">
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

          <div class="logs-panel">
            <div class="panel-title">调试输出</div>
            <div class="logs-box gradient" ref="logsRef">
              <div v-for="(line, idx) in coloredLines" :key="idx" class="log-line" :class="line.type">
                {{ line.text }}
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
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { videoSourceAPI } from '@/api'
import { message, Modal } from 'ant-design-vue'
import AppLayout from '@/components/AppLayout.vue'
import MonacoEditor from '@guolao/vue-monaco-editor'
import * as monaco from 'monaco-editor'
import LuaDocs from '@/components/LuaDocs.vue'

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

const formData = ref({ id: '', domain: '' })

const rules = { domain: [{ required: true, message: '请输入站点域名', trigger: 'blur' }] }

// 自定义 Monaco 主题（偏亮青绿）
const monacoTheme = ref('teal-light')
const defineTealTheme = () => {
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

const defaultDemo = `-- 链式调用 Demo：请求页面，querySelector 并读取 attr / text / html
print('[Demo] 启动')
set_user_agent('Lua-Demo-Agent/1.0')
set_headers({ ['Accept'] = 'text/html' })

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

const scriptContent = ref<string>(defaultDemo)
const editorRef = ref<any>(null)

const openDocs = () => { docsOpen.value = true }
const onEditorMount = (editor: any) => { editorRef.value = editor; defineTealTheme() }
const resetDemo = () => { scriptContent.value = defaultDemo }

// 最大化功能移除，相关重排逻辑一并删去

const onFillDemo = () => {
  const current = scriptContent.value?.trim() || ''
  if (current.length > 0 && current !== defaultDemo.trim()) {
    Modal.confirm({ title: '确认覆盖当前脚本？', content: '填充 Demo 将覆盖编辑器中的现有内容。', okText: '覆盖', cancelText: '取消', onOk: () => resetDemo() })
  } else { resetDemo() }
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
      formData.value.domain = data.domain || ''
    } else { message.error((response as any).message || '获取视频源详情失败') }
  } catch (err: any) { message.error(err.message || '网络错误') }
}

const handleSave = async () => {
  if (!authStore.token) return
  try { await formRef.value?.validate() } catch { return }
  saveLoading.value = true
  try {
    const payload: any = { id: formData.value.id || '', domain: formData.value.domain }
    const response = await videoSourceAPI.saveVideoSource(authStore.token, payload)
    if ((response as any).code === 0) { message.success(isEdit.value ? '保存成功' : '创建成功'); if (!isEdit.value) formData.value.id = (response as any).data?.id || '' }
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

onMounted(() => { if (isEdit.value && route.params.id) fetchVideoSourceDetail(route.params.id as string) })
</script>

<style scoped>
:root { --teal: #10b981; --teal-hover: #34d399; --teal-active: #059669; }
.page-wrap { padding: 12px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.header-actions > * { margin-left: 8px; }
.editor-panel, .logs-panel { background: transparent; border: 1px solid #20c7ab; border-radius: 8px; overflow: hidden; margin-bottom: 12px; }
.panel-title { height: 36px; display: flex; align-items: center; justify-content: space-between; padding: 0 10px; color: #0a2f28; font-weight: 800; background: linear-gradient(90deg, #99f6e4 0%, #34d399 100%); border-bottom: 1px solid #20c7ab; font-size: 13px; letter-spacing: 0.5px; }
.title-left { display: flex; align-items: center; gap: 10px; }
.title-actions { display: flex; gap: 8px; align-items: center; }
.title-text { font-weight: 800; }
.editor-gradient { background: linear-gradient(135deg, #99f6e4 0%, #5eead4 35%, #34d399 70%, #2dd4bf 100%); padding: 0; }
.monaco { min-height: 520px; height: auto; }
::v-deep(.monaco-editor),
::v-deep(.monaco-editor .margin),
::v-deep(.monaco-editor .monaco-editor-background) { background: transparent !important; max-width: 100% !important; }
/* 避免强制设置内部滚动容器宽度导致布局抖动或无限重排 */
/* ::v-deep(.monaco-editor .overflow-guard),
::v-deep(.monaco-editor .editor-scrollable) { max-width: 100% !important; width: 100% !important; } */
/* 可能的白色顶部预览/阴影 */
::v-deep(.monaco-editor .sticky-scroll),
::v-deep(.monaco-editor .monaco-sticky-scroll),
::v-deep(.monaco-editor .editor-sticky-scroll),
::v-deep(.monaco-scrollable-element .scroll-decoration) { background: transparent !important; box-shadow: none !important; }

.logs-box.gradient { background: linear-gradient(135deg, #99f6e4 0%, #5eead4 35%, #34d399 70%, #2dd4bf 100%); border-top: 1px solid #20c7ab; }
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
.teal-btn:hover, ::v-deep(.title-actions .ant-btn-primary:hover),
::v-deep(.header-actions .ant-btn-primary:hover) {
  background: linear-gradient(180deg, var(--teal-hover) 0%, var(--teal) 100%) !important;
  border-color: var(--teal-hover) !important;
}
.teal-btn:active, ::v-deep(.title-actions .ant-btn-primary:active),
::v-deep(.header-actions .ant-btn-primary:active) {
  background: linear-gradient(180deg, var(--teal-active) 0%, #047857 100%) !important;
  border-color: #047857 !important;
}
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

/* 编辑器全屏样式 */
.page-wrap, .content-card, .video-source-form { max-width: 100%; width: 100%; box-sizing: border-box; overflow-x: hidden; }
.monaco { width: 100%; }
</style>
