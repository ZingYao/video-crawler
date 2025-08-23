export const defaultTemplateLua = `local domain = ''

-- 深拷贝函数
function deep_copy(orig)
    local copy
    if type(orig) == 'table' then
        copy = {}
        for orig_key, orig_value in next, orig, nil do
            copy[deep_copy(orig_key)] = deep_copy(orig_value)
        end
        setmetatable(copy, deep_copy(getmetatable(orig)))
    else
        copy = orig
    end
    return copy
end

-- 搜索视频结果构造函数
function new_search_video_result()
    return {
        ['cover'] = '', -- 视频封面
        ['name'] = '', -- 视频名称
        ['type'] = '', -- 视频类型
        ['url'] = '', -- 视频链接
        ['actor'] = '', -- 演员
        ['director'] = '', -- 导演
        ['release_date'] = '', -- 上映日期
        ['region'] = '', -- 地区
        ['language'] = '', -- 语言
        ['description'] = '', -- 描述
        ['score'] = '' -- 评分
    }
end

-- 视频详情构造函数
function new_video_detail_result()
    return {
        ['cover'] = '', -- 视频封面
        ['name'] = '', -- 视频名称
        ['url'] = '', -- 视频链接
        ['score'] = '', -- 评分
        ['release_date'] = '', -- 上映日期
        ['region'] = '', -- 地区
        ['actor'] = '', -- 演员
        ['director'] = '', -- 导演
        ['description'] = '', -- 描述
        ['language'] = '', -- 语言
        ['source'] = { -- 数组：来源站点及剧集列表
        }
    }
end

-- 视频播放详情构造函数
function new_play_video_detail()
    return {
        ['video_url'] = '' -- 视频链接
    }
end

-- 通用请求：返回响应体字符串
function request_get_response_body(url)
    -- 设置并打印当前 UA
    print('当前 UA', set_ua_2_current_request_ua())

    local resp, reqErr = http_get(url)
    if reqErr then
        err = reqErr
        log('请求失败:', reqErr)
        return nil, err
    end
    if resp.status_code ~= 200 then
        err = '请求失败:' .. resp.status_code
        log('请求失败:', resp.status_code)
        return nil, err
    end
    return resp.body, nil
end

-- 通用请求：返回 HTML Document
function request_get_html_doc(url)
    local resp_body, err = request_get_response_body(url)
    if err then
        log('请求失败:', err)
        return nil, err
    end
    local doc, perr = parse_html(resp_body)
    if perr then
        err = '解析HTML失败:' .. perr
        log('解析HTML失败:', perr)
        return nil, err
    end
    return doc, nil
end

function search_video(search_content)
    -- 数组 Array(search_video_result)
    local result = {}
    local err = nil

    -- 实现搜索视频

    return result, err
end

function get_video_detail(video_url)
    -- video_detail_result 结构体
    local result = new_video_detail_result()
    local err = nil

    -- 实现获取视频详情

    return result, err
end

function get_play_video_detail(video_url)
    -- play_video_detail 结构体
    local result = new_play_video_detail()
    local err = nil

    -- 实现获取视频播放详情

    return result, err
end
`

export const defaultDemo = `-- 链式调用 Demo：请求页面，querySelector 并读取 attr / text / html
print('[Demo] 启动')
set_user_agent('Lua-Demo-Agent/1.0')
set_headers({ ['Accept'] = 'text/html' })

-- 获取并打印当前 UA
local current_ua = get_user_agent()
print('当前 User-Agent:', current_ua)

-- URL 和 Unicode 编解码演示
print('=== URL 编解码演示 ===')
local test_url = 'https://example.com/path?name=张三&age=25&city=北京'
local encoded_url = url.encode(test_url)
local decoded_url, decode_err = url.decode(encoded_url)
print('原始URL:', test_url)
print('编码后:', encoded_url)
if decode_err then
  print('解码错误:', decode_err)
else
  print('解码后:', decoded_url)
end

-- URL 解析和构建演示
print('=== URL 解析和构建演示 ===')
local parsed_url, parse_err = url.parse(test_url)
if parse_err then
  print('解析错误:', parse_err)
else
  print('URL 组件:')
  print('  scheme:', parsed_url.scheme)
  print('  host:', parsed_url.host)
  print('  path:', parsed_url.path)
  print('  query:', parsed_url.query)
  print('  fragment:', parsed_url.fragment)
  
  -- 修改组件并重新构建
  parsed_url.query = 'name=李四&age=30'
  local rebuilt_url, build_err = url.build(parsed_url)
  if build_err then
    print('构建错误:', build_err)
  else
    print('重新构建的URL:', rebuilt_url)
  end
end

print('=== Unicode 编解码演示 ===')
local test_unicode = 'Hello 世界！你好！'
local encoded_unicode = unicode.encode(test_unicode)
local decoded_unicode = unicode.decode(encoded_unicode)
print('原始文本:', test_unicode)
print('编码后:', encoded_unicode)
print('解码后:', decoded_unicode)

-- Unicode 工具函数演示
print('=== Unicode 工具函数演示 ===')
print('是否为ASCII:', unicode.is_ascii('Hello'))
print('是否为ASCII:', unicode.is_ascii('Hello世界'))
print('字符长度:', unicode.length('Hello世界！'))

-- 链式调用演示
print('=== 链式调用演示 ===')
local chain_result = unicode.encode('测试文本')
    :gsub('\\\\u', '\\\\u')  -- 对\\u进行二次编码
    :gsub('\\\\u', '\\\\u')  -- 再解码回来
print('链式调用结果:', chain_result)

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
        -- 对href进行URL编码演示
        local encoded_href = url_encode(href)
        print('编码后的href =', encoded_href)
      end
      print('text 文本 =', link:text())
      print('inner HTML =', link:html())
    end
  end
end
print('[Demo] 完成')
`

export const defaultTemplateJS = `// 默认 JS 模板（已内置 DOM 能力）：实现三个必须方法
// 可用：httpGet(url), setHeaders(h), setUserAgent(ua), setRandomUserAgent(), parseHtml(html)

// 搜索视频结果构造函数
function new_search_video_result() {
  return {
    cover: '', // 视频封面
    name: '', // 视频名称
    type: '', // 视频类型
    url: '', // 视频链接
    actor: '', // 演员
    director: '', // 导演
    release_date: '', // 上映日期
    region: '', // 地区
    language: '', // 语言
    description: '', // 描述
    score: '' // 评分
  }
}

// 视频详情构造函数
function new_video_detail_result() {
  return {
    cover: '', // 视频封面
    name: '', // 视频名称
    url: '', // 视频链接
    score: '', // 评分
    release_date: '', // 上映日期
    region: '', // 地区
    actor: '', // 演员
    director: '', // 导演
    description: '', // 描述
    language: '', // 语言
    source: [] // 数组：来源站点及剧集列表
  }
}

// 视频播放详情构造函数
function new_play_video_detail() {
  return {
    video_url: '' // 视频链接
  }
}

function search_video(keyword) {
  // TODO: 实现站点搜索逻辑
  // 示例：
  // const r = httpGet('https://example.com/?q=' + encodeURIComponent(keyword))
  // const doc = parseHtml(r.body)
  // const items = doc.querySelectorAll('ul > li.searchlist_item')
  // return [...Array(items.length)].map((_, i) => {
  //   const el = items[i]
  //   const result = new_search_video_result()
  //   result.name = el.querySelector('h4.vodlist_title a').attr('title') || ''
  //   result.url = el.querySelector('a').attr('href') || ''
  //   return result
  // })
  return []
}

function get_video_detail(video_url) {
  // TODO: 使用 DOM 提取详情，返回视频详情结构
  // 示例：
  // const r = httpGet(video_url)
  // const doc = parseHtml(r.body)
  // const result = new_video_detail_result()
  // result.name = doc.querySelector('h1.title').text() || ''
  // result.description = doc.querySelector('.description').text() || ''
  // // 解析剧集列表...
  // return result
  return new_video_detail_result()
}

function get_play_video_detail(video_url) {
  // TODO: 返回播放详情结构
  // 示例：
  // const r = httpGet(video_url)
  // const doc = parseHtml(r.body)
  // const result = new_play_video_detail()
  // result.video_url = doc.querySelector('video source').attr('src') || ''
  // return result
  return new_play_video_detail()
}
`

export const demoTemplateJS = `// Demo JS：使用 parseHtml 的 DOM 解析示例
setRandomUserAgent()

// 搜索视频结果构造函数
function new_search_video_result() {
  return {
    cover: '', // 视频封面
    name: '', // 视频名称
    type: '', // 视频类型
    url: '', // 视频链接
    actor: '', // 演员
    director: '', // 导演
    release_date: '', // 上映日期
    region: '', // 地区
    language: '', // 语言
    description: '', // 描述
    score: '' // 评分
  }
}

// 视频详情构造函数
function new_video_detail_result() {
  return {
    cover: '', // 视频封面
    name: '', // 视频名称
    url: '', // 视频链接
    score: '', // 评分
    release_date: '', // 上映日期
    region: '', // 地区
    actor: '', // 演员
    director: '', // 导演
    description: '', // 描述
    language: '', // 语言
    source: [] // 数组：来源站点及剧集列表
  }
}

// 视频播放详情构造函数
function new_play_video_detail() {
  return {
    video_url: '' // 视频链接
  }
}

// URL 和 Unicode 编解码演示
console.log('=== URL 编解码演示 ===')
const testUrl = 'https://example.com/path?name=张三&age=25&city=北京'
const encodedUrl = url.encode(testUrl)
const decodedUrl = url.decode(encodedUrl)
console.log('原始URL:', testUrl)
console.log('编码后:', encodedUrl)
console.log('解码后:', decodedUrl)

// URL 解析和构建演示
console.log('=== URL 解析和构建演示 ===')
const parsedUrl = url.parse(testUrl)
if (parsedUrl.error) {
  console.log('解析错误:', parsedUrl.error)
} else {
  console.log('URL 组件:', parsedUrl)
  
  // 修改组件并重新构建
  const newComponents = {
    scheme: parsedUrl.scheme,
    host: parsedUrl.host,
    path: parsedUrl.path,
    query: 'name=李四&age=30',
    fragment: parsedUrl.fragment
  }
  const rebuiltUrl = url.build(newComponents)
  console.log('重新构建的URL:', rebuiltUrl)
}

console.log('=== Unicode 编解码演示 ===')
const testUnicode = 'Hello 世界！你好！'
const encodedUnicode = unicode.encode(testUnicode)
const decodedUnicode = unicode.decode(encodedUnicode)
console.log('原始文本:', testUnicode)
console.log('编码后:', encodedUnicode)
console.log('解码后:', decodedUnicode)

// Unicode 工具函数演示
console.log('=== Unicode 工具函数演示 ===')
console.log('是否为ASCII:', unicode.isAscii('Hello'))
console.log('是否为ASCII:', unicode.isAscii('Hello世界'))
console.log('字符长度:', unicode.length('Hello世界！'))

// 链式调用演示
console.log('=== 链式调用演示 ===')
const chainResult = unicode.encode('测试文本')
  .replace(/\\\\u/g, '\\\\u')  // 对\\u进行二次编码
  .replace(/\\\\u/g, '\\\\u')  // 再解码回来
console.log('链式调用结果:', chainResult)

function search_video(keyword) {
  const url = 'https://example.com/search?q=' + encodeURIComponent(keyword)
  const r = httpGet(url)
  const html = r.body || ''
  const doc = parseHtml(html)
  const nodes = doc.querySelectorAll('ul > li.searchlist_item')
  const out = []
  for (let i = 0; i < (nodes.length || 0); i++) {
    const el = nodes[i]
    const result = new_search_video_result()
    
    const a = el.querySelector('h4.vodlist_title a')
    result.name = a ? (a.attr('title') || a.text()) : ''
    result.url = a ? (a.attr('href') || '') : ''
    
    const coverA = el.querySelector('a.vodlist_thumb')
    result.cover = coverA ? (coverA.attr('data-original') || '') : ''
    
    const scoreSpan = el.querySelector('a.vodlist_thumb span:last-child')
    result.score = scoreSpan ? scoreSpan.text() : ''
    
    // 提取其他信息
    const actorEl = el.querySelector('.actor')
    result.actor = actorEl ? actorEl.text() : ''
    
    const directorEl = el.querySelector('.director')
    result.director = directorEl ? directorEl.text() : ''
    
    const dateEl = el.querySelector('.release_date')
    result.release_date = dateEl ? dateEl.text() : ''
    
    const regionEl = el.querySelector('.region')
    result.region = regionEl ? regionEl.text() : ''
    
    out.push(result)
  }
  return out
}

function get_video_detail(video_url) {
  const r = httpGet(video_url)
  const doc = parseHtml(r.body || '')
  const result = new_video_detail_result()
  
  // 基本信息
  const topA = doc.querySelector('div.detail_list div.content_box div:nth-child(1) > a')
  if (topA) {
    result.name = topA.attr('title') || ''
  }
  
  const coverEl = doc.querySelector('.cover img')
  result.cover = coverEl ? coverEl.attr('src') || '' : ''
  
  const scoreEl = doc.querySelector('.score')
  result.score = scoreEl ? scoreEl.text() : ''
  
  const descEl = doc.querySelector('.description')
  result.description = descEl ? descEl.text() : ''
  
  const actorEl = doc.querySelector('.actor')
  result.actor = actorEl ? actorEl.text() : ''
  
  const directorEl = doc.querySelector('.director')
  result.director = directorEl ? directorEl.text() : ''
  
  const dateEl = doc.querySelector('.release_date')
  result.release_date = dateEl ? dateEl.text() : ''
  
  const regionEl = doc.querySelector('.region')
  result.region = regionEl ? regionEl.text() : ''
  
  const langEl = doc.querySelector('.language')
  result.language = langEl ? langEl.text() : ''
  
  // 源与剧集（示例选择器，按站点实际调整）
  const tabs = doc.querySelectorAll('div.play_source_tab > a')
  const lists = doc.querySelectorAll('div.play_list_box')
  for (let i = 0; i < (tabs.length || 0); i++) {
    // 创建来源站点对象
    const src = { 
      name: tabs[i].attr('alt') || tabs[i].text() || ('线路' + (i + 1)), 
      episodes: [] 
    }
    const box = lists[i]
    if (box) {
      const lis = box.querySelectorAll('ul > li')
      for (let j = 0; j < (lis.length || 0); j++) {
        const a = lis[j].querySelector('a')
        if (!a) continue
        // 创建剧集对象
        const episode = { 
          name: a.text() || a.attr('title') || ('第' + (j + 1) + '集'), 
          url: a.attr('href') || '' 
        }
        src.episodes.push(episode)
      }
    }
    result.source.push(src)
  }
  return result
}

function get_play_video_detail(video_url) {
  // 站点若需解析中转，可在此请求并解析；否则直接返回
  const result = new_play_video_detail()
  
  // 示例：解析播放页面获取真实视频URL
  // const r = httpGet(video_url)
  // const doc = parseHtml(r.body || '')
  // const videoEl = doc.querySelector('video source')
  // result.video_url = videoEl ? videoEl.attr('src') : video_url
  
  result.video_url = video_url
  return result
}
`