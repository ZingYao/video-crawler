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

export const defaultTemplateJS = `// 默认 JS 模板（已内置 DOM 能力）：实现三个必须方法
// 可用：httpGet(url), setHeaders(h), setUserAgent(ua), setRandomUserAgent(), parseHtml(html)

function search_video(keyword) {
  // TODO: 实现站点搜索逻辑
  // 示例：
  // const r = httpGet('https://example.com/?q=' + encodeURIComponent(keyword))
  // const doc = parseHtml(r.body)
  // const items = doc.querySelectorAll('ul > li.searchlist_item')
  // return [...Array(items.length)].map((_, i) => {
  //   const el = items[i]
  //   return { name: el.querySelector('h4.vodlist_title a').attr('title') || '', url: '' }
  // })
  return []
}

function get_video_detail(video_url) {
  // TODO: 使用 DOM 提取详情，返回 { name, description, sources: [{name, episodes:[{name,url}]}] }
  return { name: '', description: '', sources: [] }
}

function get_play_video_detail(video_url) {
  // TODO: 返回 { video_url }
  return { video_url }
}
`

export const demoTemplateJS = `// Demo JS：使用 parseHtml 的 DOM 解析示例
setRandomUserAgent()

function search_video(keyword) {
  const url = 'https://example.com/search?q=' + encodeURIComponent(keyword)
  const r = httpGet(url)
  const html = r.body || ''
  const doc = parseHtml(html)
  const nodes = doc.querySelectorAll('ul > li.searchlist_item')
  const out = []
  for (let i = 0; i < (nodes.length || 0); i++) {
    const el = nodes[i]
    const a = el.querySelector('h4.vodlist_title a')
    const name = a ? (a.attr('title') || a.text()) : ''
    const coverA = el.querySelector('a.vodlist_thumb')
    const cover = coverA ? (coverA.attr('data-original') || '') : ''
    const href = coverA ? (coverA.attr('href') || '') : ''
    const scoreSpan = el.querySelector('a.vodlist_thumb span:last-child')
    const score = scoreSpan ? scoreSpan.text() : ''
    out.push({ name, cover, url: href, score })
  }
  return out
}

function get_video_detail(video_url) {
  const r = httpGet(video_url)
  const doc = parseHtml(r.body || '')
  const res = { name: '', description: '', sources: [] }
  const topA = doc.querySelector('div.detail_list div.content_box div:nth-child(1) > a')
  if (topA) {
    res.name = topA.attr('title') || ''
  }
  // 源与剧集（示例选择器，按站点实际调整）
  const tabs = doc.querySelectorAll('div.play_source_tab > a')
  const lists = doc.querySelectorAll('div.play_list_box')
  for (let i = 0; i < (tabs.length || 0); i++) {
    const src = { name: tabs[i].attr('alt') || ('来源' + (i + 1)), episodes: [] }
    const box = lists[i]
    if (box) {
      const lis = box.querySelectorAll('ul > li')
      for (let j = 0; j < (lis.length || 0); j++) {
        const a = lis[j].querySelector('a')
        if (!a) continue
        src.episodes.push({ name: a.text(), url: a.attr('href') || '' })
      }
    }
    res.sources.push(src)
  }
  return res
}

function get_play_video_detail(video_url) {
  // 站点若需解析中转，可在此请求并解析；否则直接返回
  return { video_url }
}
`