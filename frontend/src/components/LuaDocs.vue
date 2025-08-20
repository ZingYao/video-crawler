<template>
  <a-drawer
    v-model:open="localOpen"
    title="Lua 扩展函数说明文档"
    placement="right"
    width="580"
    :mask="false"
    :get-container="false"
  >
    <div class="docs">
      <!-- 默认文档：如果父级未提供插槽，则显示本内容 -->
      <slot>
        <div class="doc-section">
          <h3>总览</h3>
          <div class="doc-item">所有方法均在 Lua 脚本中直接调用；错误通过第二返回值返回（约定：成功时第二返回值为 <code>nil</code>）。</div>
        </div>

        <div class="doc-section">
          <h3>日志与工具</h3>
          <div class="doc-item">
            <b>print(...)</b>：打印到调试输出，前缀为 [PRINT]，后端已添加毫秒级时间戳。
          </div>
          <div class="doc-item">
            <b>log(...)</b>：打印到调试输出，前缀为 [LOG]，带毫秒级时间戳。
          </div>
          <div class="doc-item">
            <b>sleep(ms)</b>：暂停当前协程 <code>ms</code> 毫秒。
          </div>
          <pre class="doc-code">-- 示例
print('hello')
log('step 1')
sleep(300)
log('step 2 after 300ms')
</pre>
          <div class="doc-item"><b>参数/返回</b></div>
          <ul>
            <li><code>print(...)</code>：任意参数，<b>无返回</b></li>
            <li><code>log(...)</code>：任意参数，<b>无返回</b></li>
            <li><code>sleep(ms)</code>：<code>ms:number</code> 毫秒，<b>无返回</b></li>
          </ul>
        </div>

        <div class="doc-section">
          <h3>字符串工具</h3>
          <div class="doc-item"><b>split(s: string, sep: string)</b> → <code>array</code> 按分隔符切分；当 <code>sep</code> 为空字符串时，按字符切分（rune 安全）。</div>
          <div class="doc-item"><b>trim(s: string)</b> → <code>string</code> 去除首尾空白。</div>
          <pre class="doc-code">-- split 示例
local parts = split('a,b,c', ',')
print(#parts, parts[1], parts[2], parts[3])  -- 3 a b c

-- 按字符切分（sep 为空串）
local chars = split('汉字OK', '')
print(#chars, chars[1], chars[2])  -- 4 汉 字

-- trim 示例
print('[' .. trim('  hello  ') .. ']')  -- [hello]
print('[' .. trim('\n\t中文 空格\t') .. ']')
</pre>
          <div class="doc-item"><b>参数/返回</b></div>
          <ul>
            <li><code>split(s, sep)</code>：<code>s:string</code>、<code>sep:string</code>；返回 <code>array(table)</code></li>
            <li><code>trim(s)</code>：<code>s:string</code>；返回 <code>string</code></li>
          </ul>
        </div>

        <div class="doc-section">
          <h3>JSON 编解码</h3>
          <div class="doc-item"><b>json_encode(value[, indent])</b> → <code>string, err</code> 将 Lua 值编码为 JSON 字符串。</div>
          <div class="doc-item"><b>json_decode(json)</b> → <code>value, err</code> 将 JSON 字符串解码为 Lua 值（对象→table、数组→数组、null→nil）。</div>
          <pre class="doc-code">-- encode 紧凑
local obj = { a = 1, b = true, c = {1, 2, 3}, s = '中文' }
local s1, e1 = json_encode(obj)
if e1 then log('encode error:', e1) else print('json:', s1) end

-- encode 使用两空格缩进（indent 为 boolean=true）
local s2 = json_encode(obj, true)
print(s2)

-- encode 使用 4 空格缩进（indent 为 number）
local s3 = json_encode(obj, 4)
print(s3)

-- encode 使用制表符缩进（indent 为 string）
local s4 = json_encode(obj, "\t")
print(s4)

-- decode
local v, derr = json_decode('{"x":10,"y":[1,2,3],"ok":true,"s":"hi","n":null}')
if derr then
  log('decode error:', derr)
else
  print(v.x, v.ok)       -- 10  true
  print(#v.y, v.y[1])    -- 3   1
  print(v.s, v.n == nil) -- hi  true
end
</pre>
          <div class="doc-item"><b>indent 参数说明（可选）</b></div>
          <ul>
            <li><b>boolean</b>：<code>true</code> → 两空格缩进；<code>false</code> → 紧凑模式</li>
            <li><b>number</b>：使用给定数量的空格缩进</li>
            <li><b>string</b>：使用该字符串作为缩进（如 <code>"\t"</code>）</li>
          </ul>
        </div>

        <div class="doc-section">
          <h3>HTTP</h3>
          <div class="doc-item"><b>set_user_agent(ua: string)</b> 设置 UA</div>
          <div class="doc-item"><b>set_random_user_agent()</b> 随机 UA</div>
          <div class="doc-item"><b>get_user_agent()</b> → <code>string</code> 获取当前 UA</div>
          <div class="doc-item"><b>set_ua_2_current_request_ua()</b> → <code>string</code> 将当前 HTTP 客户端 UA 写入请求头并返回实际生效的 UA</div>
          <div class="doc-item"><b>set_headers(h: table)</b> 设置通用请求头</div>
          <div class="doc-item"><b>set_cookies(c: table)</b> 设置通用 Cookie（键值对）</div>
          <div class="doc-item"><b>http_get(url: string)</b> → <code>resp, err</code></div>
          <div class="doc-item"><b>http_post(url: string, data: table|string)</b> → <code>resp, err</code></div>
          <div class="doc-item">resp 结构：<code>{ status_code:number, url:string, headers:table, body:string }</code></div>
          <pre class="doc-code">-- set_user_agent / set_random_user_agent / get_user_agent / set_ua_2_current_request_ua
set_user_agent('Lua-Demo/1.0')
set_random_user_agent()  -- 可选：随机 UA 会覆盖上面的 UA
local ua_applied = set_ua_2_current_request_ua()  -- 将当前 UA 写入到后续 HTTP 请求
print('生效 UA:', ua_applied)

local current_ua = get_user_agent()  -- 获取当前设置的 UA
print('当前 UA:', current_ua)

-- set_headers / set_cookies
set_headers({ ['Accept'] = 'application/json', ['X-Trace'] = 'demo' })
set_cookies({ session = 'abc', token = 'xyz' })

-- http_get(url) -> resp, err
local r1, e1 = http_get('https://httpbin.org/get')
if e1 then
  log('GET 错误:', e1)
else
  print('GET 状态码:', r1.status_code)
  print('GET 最终URL:', r1.url)
  print('GET 响应体长度:', #r1.body)
end

-- http_post(url, data) -> resp, err
local payload = { q = 'lua', page = 1 }
local r2, e2 = http_post('https://httpbin.org/post', payload)
if e2 then
  log('POST 错误:', e2)
else
  print('POST 状态码:', r2.status_code)
  print('POST Body 片段:', string.sub(r2.body, 1, 60) .. '...')
end
</pre>
          <div class="doc-item"><b>参数/返回</b></div>
          <ul>
            <li><code>set_user_agent(ua)</code>：<code>ua:string</code>；<b>无返回</b></li>
            <li><code>set_random_user_agent()</code>：<b>无返回</b></li>
            <li><code>get_user_agent()</code>：<b>返回</b> <code>string</code> 当前 UA</li>
            <li><code>set_ua_2_current_request_ua()</code>：<b>返回</b> <code>string</code> 实际生效的 UA</li>
            <li><code>set_headers(h)</code>：<code>h:table</code>，示例 <code>{ ['K']='V' }</code>；<b>无返回</b></li>
            <li><code>set_cookies(c)</code>：<code>c:table</code>，示例 <code>{ name='v' }</code>；<b>无返回</b></li>
            <li><code>http_get(url)</code>：<code>url:string</code>；返回 <code>resp, err</code></li>
            <li><code>http_post(url, data)</code>：<code>data:table|string</code>；返回 <code>resp, err</code></li>
          </ul>
        </div>

        <div class="doc-section">
          <h3>HTML 解析</h3>
          <div class="doc-item"><b>parse_html(html: string)</b> → <code>doc, err</code></div>
          <div class="doc-item"><b>select(doc|el, css: string)</b> → <code>elements[], err</code></div>
          <div class="doc-item"><b>select_one(doc|el, css: string)</b> → <code>element, err</code></div>
          <div class="doc-item"><b>text(el)</b> → <code>string</code>，<b>html(el)</b> → <code>string</code>，<b>attr(el, name)</b> → <code>value, err</code></div>
          <pre class="doc-code">local htmlStr = [[
&lt;html&gt;&lt;body&gt;
  &lt;div class="card" data-id="100"&gt;Hello&lt;/div&gt;
  &lt;p id="p1"&gt;world&lt;/p&gt;
&lt;/body&gt;&lt;/html&gt;
]]
local doc, perr = parse_html(htmlStr)
if perr then
  log('parse_html 错误:', perr)
  return
end

-- select / text / html / attr
local els, selErr = select(doc, 'div.card')
if not selErr then
  for i, el in ipairs(els) do
    local id, aerr = attr(el, 'data-id')
    print('card', i, 'id=', id, 'text=', text(el), 'html=', html(el))
  end
end

-- select_one
local one, oneErr = select_one(doc, '#p1')
if not oneErr then
  print('select_one #p1 =', text(one))
end
</pre>
          <div class="doc-item"><b>参数/返回</b></div>
          <ul>
            <li><code>parse_html(html)</code>：<code>html:string</code>；返回 <code>doc, err</code></li>
            <li><code>select(ctx, css)</code>：<code>ctx:doc|element</code>、<code>css:string</code>；返回 <code>elements(table), err</code></li>
            <li><code>select_one(ctx, css)</code>：同上；返回 <code>element, err</code></li>
            <li><code>text(el)</code>：<code>el:element</code>；返回 <code>string</code></li>
            <li><code>html(el)</code>：<code>el:element</code>；返回 <code>string</code></li>
            <li><code>attr(el, name)</code>：<code>name:string</code>；返回 <code>value, err</code></li>
          </ul>
        </div>
      </slot>
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

.docs b {
  font-weight: 600;
  color: #047857;
}

/* 确保抽屉内容可见 */
:deep(.ant-drawer-body) {
  padding: 0 !important;
}

:deep(.ant-drawer-content) {
  background: #ffffff;
}
</style>
