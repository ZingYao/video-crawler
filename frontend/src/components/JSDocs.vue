<template>
  <a-drawer
    v-model:open="localOpen"
    title="JavaScript 扩展函数说明文档"
    placement="right"
    width="580"
    :mask="false"
    :get-container="false"
  >
    <div class="docs">
      <div class="doc-section">
        <h3>总览</h3>
        <div class="doc-item">本引擎基于 goja，已做沙箱处理：不提供 <code>os</code>、<code>fs</code>、<code>child_process</code> 等危险能力，可直接使用原生 <code>JSON.parse</code> / <code>JSON.stringify</code>。</div>
      </div>

      <div class="doc-section">
        <h3>HTTP</h3>
        <div class="doc-item"><b>setUserAgent(ua: string)</b> 设置 UA</div>
        <div class="doc-item"><b>setRandomUserAgent()</b> 随机 UA</div>
        <div class="doc-item"><b>getUserAgent()</b> → <code>string</code> 获取当前 UA</div>
        <div class="doc-item"><b>setUaToCurrentRequestUa()</b> → <code>string</code> 将当前 HTTP 客户端 UA 写入请求头并返回实际生效的 UA</div>
        <div class="doc-item"><b>setHeaders(h: Record&lt;string,string&gt;)</b> 设置通用请求头</div>
        <div class="doc-item"><b>setCookies(c: Record&lt;string,string&gt;)</b> 设置通用 Cookie（键值对）</div>
        <div class="doc-item"><b>httpGet(url: string)</b> → <code>{ status_code, url, headers, body }</code></div>
        <div class="doc-item"><b>httpPost(url: string, data: object|string)</b> → <code>{ status_code, url, headers, body }</code></div>
        <div class="doc-item"><b>fetch(url, options)</b> → <code>Response</code>（同步返回）：支持 <code>method</code>/<code>headers</code>/<code>body</code>/<code>timeout(ms)</code>/<code>redirect</code>（<code>follow|manual|error</code>）</div>
        <pre class="doc-code">// UA / Headers / Cookies
setUserAgent('JS-Demo/1.0')
setRandomUserAgent()
const uaApplied = setUaToCurrentRequestUa()
const currentUA = getUserAgent()
setHeaders({ 'Accept': 'application/json', 'X-Trace': 'demo' })
setCookies({ session: 'abc', token: 'xyz' })

// httpGet
const r1 = httpGet('https://httpbin.org/get')
console.log('GET code:', r1.status_code)

// httpPost（对象将自动 JSON 序列化）
const r2 = httpPost('https://httpbin.org/post', { q: 'js', page: 1 })
console.log('POST code:', r2.status_code)

// fetch 同步返回 Response 对象
const r3 = fetch('https://httpbin.org/anything', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: { a: 1, b: true },
  timeout: 8000,
  redirect: 'manual'
})
console.log(r3.status, r3.ok)
console.log('text length:', r3.text().length)
const json = r3.json() // 解析失败返回 undefined
console.log('url:', r3.url, 'location(if 3xx):', r3.location)
</pre>
        <div class="doc-item"><b>Response</b> 字段/方法：</div>
        <ul>
          <li><code>ok</code>、<code>status</code>、<code>statusText</code>、<code>url</code>、<code>redirected</code>、<code>type='basic'</code></li>
          <li><code>headers</code>：提供 <code>get(name)</code>、<code>has(name)</code>、<code>keys()</code>、<code>values()</code>、<code>entries()</code>、<code>forEach((v,k)=>{})</code></li>
          <li><code>text()</code> → string，<code>json()</code> → any|undefined，<code>arrayBuffer()</code> → Uint8Array</li>
        </ul>
      </div>

      <div class="doc-section">
        <h3>建议规范</h3>
        <div class="doc-item">- 优先使用 <b>fetch</b>，其语义更接近 Web/Node；需要简单快速时可用 <b>httpGet/httpPost</b>。</div>
        <div class="doc-item">- 处理重定向：<code>redirect: 'manual'</code> 时，可读取 <code>response.location</code> 自行处理跳转。</div>
        <div class="doc-item">- JSON 直接使用 <code>JSON.parse</code>/<code>JSON.stringify</code>；不需要 <code>json_encode/json_decode</code>。</div>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits(['update:open'])

const localOpen = computed({
  get: () => props.open,
  set: (v: boolean) => emit('update:open', v),
})
</script>

<style scoped>
.docs { 
  color: #052e2b; 
  padding: 16px;
  font-size: 14px;
  line-height: 1.6;
}
.doc-section { 
  margin-bottom: 24px; 
  padding-left: 12px; 
  border-left: 3px solid #10b981; 
  background: rgba(16, 185, 129, 0.02);
  padding: 16px 12px;
  border-radius: 8px;
}
.docs h3 { 
  margin: 0 0 16px 0; 
  font-weight: 700; 
  color: #065f46; 
  font-size: 18px;
  border-bottom: 2px solid #10b981;
  padding-bottom: 8px;
}
.doc-item { 
  margin: 12px 0; 
  line-height: 1.7; 
  color: #064e3b; 
  font-size: 14px;
}
.docs ul { 
  padding-left: 20px; 
  margin: 12px 0;
}
.docs li {
  margin: 8px 0;
  color: #064e3b;
}
.docs code { 
  background: #f0fdf4; 
  color: #166534; 
  padding: 2px 6px; 
  border-radius: 4px; 
  border: 1px solid #bbf7d0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}
.docs pre.doc-code { 
  background: #f0fdf4; 
  color: #166534; 
  padding: 16px; 
  border-radius: 8px; 
  overflow-x: auto; 
  border: 1px solid #bbf7d0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  margin: 12px 0;
}
::deep(.ant-drawer-body) {
  padding: 0 !important;
}
::deep(.ant-drawer-content) {
  background: #ffffff;
}
</style>
