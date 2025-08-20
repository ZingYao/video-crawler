export const defaultTemplateLua = `local domain = 'https://www.hongling8.com'

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

    local doc, derr = request_get_html_doc(domain .. '/yhsearch/-------------.html?wd=' .. search_content .. '&submit=')
    if derr then
        log('请求查询视频失败:', derr)
        return result, derr
    end

    local items = doc:select('ul > li.searchlist_item')
    local index = 0
    while true do
        local elem = items:eq(index)
        if not elem or elem:html() == '' then
            log('查找元素结束')
            break
        end
        log('查找元素', index)

        local result_item = new_search_video_result()
        local title_dom = elem:select_one('div.searchlist_titbox > h4.vodlist_title > a')
        if title_dom then
            result_item.name = title_dom:attr('title')
            result_item.type = title_dom:select_one('span.info_right'):text()

            local cover_dom = elem:select_one('div.searchlist_img > a.vodlist_thumb')
            result_item.cover = cover_dom:attr('data-original')
            result_item.url = domain .. cover_dom:attr('href')
            result_item.score = cover_dom:select_one('span:last-child'):text()

            local actor_dom = elem:select_one('div.searchlist_titbox > p:nth-child(2)')
            result_item.actor = actor_dom:text():gsub('主演：', '')

            local director_dom = elem:select_one('div.searchlist_titbox > p:nth-child(3)')
            result_item.director = director_dom:text():gsub('导演：', '')

            local description_dom = elem:select_one('div.searchlist_titbox > p:nth-child(4)')
            result_item.description = description_dom:text():gsub('简介：', '')

            table.insert(result, result_item)
        end
        index = index + 1
    end
    return result, err
end

function get_video_detail(video_url)
    -- video_detail_result 结构体
    local result = new_video_detail_result()
    local err = nil

    local doc, derr = request_get_html_doc(video_url)
    if derr then
        log('请求视频详情失败:', derr)
        return result, derr
    end

    local video_detail_dom = doc:select_one('div.detail_list > div.content_box')
    if video_detail_dom then
        local cover_dom = video_detail_dom:select_one('div:nth-child(1) > a')
        if cover_dom then
            result.cover = cover_dom:attr('data-original')
            result.name = cover_dom:attr('title')
            result.url = domain .. cover_dom:attr('href')
        end

        local score_dom = video_detail_dom:select_one('div:nth-child(2) > span.star_tips')
        if score_dom then result.score = score_dom:text() end

        local description_dom = video_detail_dom:select_one('div:nth-child(3) > ul')
        if description_dom then
            local li_dom = description_dom:select_one('li:nth-child(4)')
            if li_dom then
                local result_arr = split(li_dom:html(), 'split_line')
                for _, v in pairs(result_arr) do
                    local str = ''
                    for s in v:gmatch('>(.-)<') do
                        str = str .. s
                    end
                    if str:find('上映') then
                        result.release_date = str:match('(%d+)')
                    elseif str:find('地区') then
                        result.region = trim(str:gsub('地区：', ''))
                    elseif str:find('类型') then
                        result.type = trim(str:gsub('类型：', ''))
                    end
                end
            end

            local actor_dom = description_dom:select_one('li:nth-child(6)')
            if actor_dom then
                local actor_text = actor_dom:text()
                result.actor = trim(actor_text:gsub('演员：', ''))
            end

            local director_dom = description_dom:select_one('li:nth-child(7)')
            if director_dom then
                local director_text = director_dom:text()
                result.director = trim(director_text:gsub('导演：', ''))
            end
        end
    end

    local resource_dom = doc:select_one('div.play_source')
    if not resource_dom then
        err = '解析视频详情失败: 没有找到资源 Dom'
        log(err)
        return result, err
    end

    local source_list = resource_dom:select('div.play_source_tab > a')
    if not source_list then
        err = '解析视频详情失败: 没有找到资源站点'
        log(err)
        return result, err
    end

    local index = 0
    while true do
        local source_item = source_list:eq(index)
        if not source_item or source_item:html() == '' then break end
        local source_item_result = {
            ['name'] = source_item:attr('alt'),
            ['episodes'] = {}
        }

        local episode_list = doc:select('div.play_list_box:nth-child(' .. index + 2 .. ') > div.playlist_full > ul > li')
        if episode_list then
            local episode_index = 0
            while true do
                local episode_item = episode_list:eq(episode_index)
                if not episode_item or (episode_item:html() == '' and episode_index >= 10) then
                    break
                end
                if episode_item:html() ~= '' then
                    local episode_name = episode_item:select_one('a'):text()
                    local episode_url = domain .. episode_item:select_one('a'):attr('href')
                    table.insert(source_item_result.episodes, { ['name'] = episode_name, ['url'] = episode_url })
                end
                episode_index = episode_index + 1
            end
        end

        table.insert(result.source, source_item_result)
        index = index + 1
    end

    return result, err
end

function get_play_video_detail(video_url)
    -- play_video_detail 结构体
    local result = new_play_video_detail()
    local err = nil

    local body, derr = request_get_response_body(video_url)
    if derr then
        log('请求视频详情失败:', derr)
        return result, derr
    end

    -- 查找 body 中带有视频地址的 json 串
    local video_url_json = body:match('<script .-aaaa=(.-)</script>')
    if video_url_json then
        local json_obj = json_decode(video_url_json)
        result.video_url = json_obj and json_obj.url or ''
    end

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