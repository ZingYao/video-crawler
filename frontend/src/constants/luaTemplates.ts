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

export const defaultTemplateJS = `// 默认 JS 模板：实现三个必须方法
// 可用辅助：http_get(url), set_headers(map), set_user_agent(ua), set_random_user_agent()

function search_video(keyword) {
  // TODO: 实现站点搜索逻辑，返回数组
  return []
}

function get_video_detail(video_url) {
  // TODO: 实现详情解析，返回包含来源与剧集
  return { name: '', description: '', sources: [] }
}

function get_play_video_detail(video_url) {
  // TODO: 返回可播放链接（可直接回传或二次解析）
  return { video_url }
}
`

export const demoTemplateJS = `// Demo JS：演示请求与简单结构返回（按需改写）
set_random_user_agent()

function search_video(keyword) {
  const q = encodeURIComponent(keyword)
  const resp = http_get('https://example.com/search?q=' + q)
  return []
}

function get_video_detail(video_url) {
  const resp = http_get(video_url)
  return { name: 'Demo', description: '示例', sources: [ { name: '默认', episodes: [ { name: '第1集', url: video_url } ] } ] }
}

function get_play_video_detail(video_url) {
  return { video_url }
}
`