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
        <h3>必须实现的三个方法</h3>
        <div class="doc-item"><b>search_video(keyword: string)</b> → <code>array</code> 搜索视频，返回搜索结果数组。</div>
        <div class="doc-item"><b>get_video_detail(video_url: string)</b> → <code>object</code> 获取视频详情，返回视频详情结构。</div>
        <div class="doc-item"><b>get_play_video_detail(video_url: string)</b> → <code>object</code> 获取播放详情，返回播放详情结构。</div>
        
        <h4>数据结构</h4>
        <div class="doc-item"><b>搜索视频结果 (search_video_result)</b></div>
        <pre class="doc-code">{
  cover: '',        // 视频封面
  name: '',         // 视频名称
  type: '',         // 视频类型
  url: '',          // 视频链接
  actor: '',        // 演员
  director: '',     // 导演
  release_date: '', // 上映日期
  region: '',       // 地区
  language: '',     // 语言
  description: '',  // 描述
  score: ''         // 评分
}</pre>
        
        <div class="doc-item"><b>视频详情结果 (video_detail_result)</b></div>
        <pre class="doc-code">{
  cover: '',        // 视频封面
  name: '',         // 视频名称
  url: '',          // 视频链接
  score: '',        // 评分
  release_date: '', // 上映日期
  region: '',       // 地区
  actor: '',        // 演员
  director: '',     // 导演
  description: '',  // 描述
  language: '',     // 语言
  source: []        // 数组：来源站点及剧集列表
}</pre>
        
        <div class="doc-item"><b>来源站点对象 (source_item)</b></div>
        <pre class="doc-code">{
  name: '',         // 来源站点名称（如：'线路1'、'线路2'、'备用线路'等）
  episodes: []      // 剧集列表数组
}</pre>
        
        <div class="doc-item"><b>剧集对象 (episode_item)</b></div>
        <pre class="doc-code">{
  name: '',         // 剧集名称（如：'第1集'、'第2集'、'大结局'等）
  url: ''           // 剧集播放链接
}</pre>
        
        <div class="doc-item"><b>播放详情结果 (play_video_detail)</b></div>
        <pre class="doc-code">{
  video_url: ''     // 视频链接
}</pre>
        
        <h4>构造函数</h4>
        <div class="doc-item">提供了三个构造函数来创建标准数据结构：</div>
        <pre class="doc-code">// 创建搜索视频结果
const result = new_search_video_result()
result.name = '电影名称'
result.url = 'https://example.com/movie'

// 创建视频详情结果
const detail = new_video_detail_result()
detail.name = '电影名称'
detail.description = '电影描述'

// 构建来源站点和剧集列表
detail.source = [
  // 来源站点1
  { 
    name: '线路1', 
    episodes: [
      { name: '第1集', url: 'https://example.com/ep1' },
      { name: '第2集', url: 'https://example.com/ep2' },
      { name: '第3集', url: 'https://example.com/ep3' },
      { name: '大结局', url: 'https://example.com/final' }
    ]
  },
  // 来源站点2（备用线路）
  { 
    name: '线路2', 
    episodes: [
      { name: '第1集', url: 'https://example2.com/ep1' },
      { name: '第2集', url: 'https://example2.com/ep2' },
      { name: '第3集', url: 'https://example2.com/ep3' }
    ]
  },
  // 来源站点3（高清线路）
  { 
    name: '高清线路', 
    episodes: [
      { name: '第1集', url: 'https://hd.example.com/ep1' },
      { name: '第2集', url: 'https://hd.example.com/ep2' }
    ]
  }
]

// 创建播放详情结果
const play = new_play_video_detail()
play.video_url = 'https://example.com/video.mp4'
</pre>
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
        <h3>URL 库</h3>
        <div class="doc-item"><b>url.encode(str: string)</b> → <code>string</code> 将字符串进行 URL 编码。</div>
        <div class="doc-item"><b>url.decode(str: string)</b> → <code>string</code> 将 URL 编码的字符串解码。</div>
        <div class="doc-item"><b>url.parse(url: string)</b> → <code>object</code> 解析URL为组件对象。</div>
        <div class="doc-item"><b>url.build(object: object)</b> → <code>string</code> 从组件对象构建URL。</div>
        <pre class="doc-code">// URL 编码示例
const testUrl = 'https://example.com/path?name=张三&age=25&city=北京'
const encoded = url.encode(testUrl)
console.log('编码后:', encoded)
// 输出: https%3A%2F%2Fexample.com%2Fpath%3Fname%3D%E5%BC%A0%E4%B8%89%26age%3D25%26city%3D%E5%8C%97%E4%BA%AC

// URL 解码示例
const decoded = url.decode(encoded)
console.log('解码后:', decoded)
// 输出: https://example.com/path?name=张三&age=25&city=北京

// URL 解析示例
const parsed = url.parse(testUrl)
if (parsed.error) {
  console.log('解析错误:', parsed.error)
} else {
  console.log('scheme:', parsed.scheme)    // https
  console.log('host:', parsed.host)        // example.com
  console.log('path:', parsed.path)        // /path
  console.log('query:', parsed.query)      // name=张三&age=25&city=北京
  console.log('fragment:', parsed.fragment) // (空)
}

// URL 构建示例
const components = {
  scheme: 'https',
  host: 'example.com',
  path: '/api/v1',
  query: 'id=123&type=user'
}
const builtUrl = url.build(components)
console.log('构建的URL:', builtUrl)
// 输出: https://example.com/api/v1?id=123&type=user

// 链式调用示例
const result = url.encode('测试文本')
  .replace(/%/g, '%25')  // 对%进行二次编码
  .replace(/%25/g, '%')  // 再解码回来
console.log('链式调用结果:', result)
</pre>
        <div class="doc-item"><b>参数/返回</b></div>
        <ul>
          <li><code>url.encode(str)</code>：<code>str:string</code>；返回 <code>string</code></li>
          <li><code>url.decode(str)</code>：<code>str:string</code>；返回 <code>string</code></li>
          <li><code>url.parse(url)</code>：<code>url:string</code>；返回 <code>object</code>（包含 scheme, host, path, query, fragment, raw 字段）</li>
          <li><code>url.build(object)</code>：<code>object:object</code>；返回 <code>string</code></li>
        </ul>
      </div>

      <div class="doc-section">
        <h3>Unicode 库</h3>
        <div class="doc-item"><b>unicode.encode(str: string)</b> → <code>string</code> 将字符串中的非ASCII字符编码为 \uXXXX 格式。</div>
        <div class="doc-item"><b>unicode.decode(str: string)</b> → <code>string</code> 将 \uXXXX 格式的字符串解码为原始字符。</div>
        <div class="doc-item"><b>unicode.isAscii(str: string)</b> → <code>boolean</code> 检查字符串是否只包含ASCII字符。</div>
        <div class="doc-item"><b>unicode.length(str: string)</b> → <code>number</code> 返回字符串的Unicode字符数量。</div>
        <pre class="doc-code">// Unicode 编码示例
const testText = 'Hello 世界！你好！'
const encoded = unicode.encode(testText)
console.log('编码后:', encoded)
// 输出: Hello \u4E16\u754C\uFF01\u4F60\u597D\uFF01

// Unicode 解码示例
const decoded = unicode.decode(encoded)
console.log('解码后:', decoded)
// 输出: Hello 世界！你好！

// Unicode 工具函数示例
console.log('是否为ASCII:', unicode.isAscii('Hello'))      // true
console.log('是否为ASCII:', unicode.isAscii('Hello世界'))  // false
console.log('字符长度:', unicode.length('Hello世界！'))      // 12

// 链式调用示例
const result = unicode.encode('测试文本')
  .replace(/\\\\u/g, '\\\\u')  // 对\\u进行二次编码
  .replace(/\\\\u/g, '\\\\u')  // 再解码回来
console.log('链式调用结果:', result)
</pre>
        <div class="doc-item"><b>参数/返回</b></div>
        <ul>
          <li><code>unicode.encode(str)</code>：<code>str:string</code>；返回 <code>string</code></li>
          <li><code>unicode.decode(str)</code>：<code>str:string</code>；返回 <code>string</code></li>
          <li><code>unicode.isAscii(str)</code>：<code>str:string</code>；返回 <code>boolean</code></li>
          <li><code>unicode.length(str)</code>：<code>str:string</code>；返回 <code>number</code></li>
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
