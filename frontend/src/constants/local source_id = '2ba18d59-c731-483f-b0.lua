local domain = 'https://www.hongling8.com'
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
        ['source'] = {
            -- {
            --     ['name'] = '', -- 来源名称
            --     ['episodes'] = {
            --         ['name'] = '', -- 剧集名称
            --         ['url'] = '' -- 剧集链接
            --     }
            -- }
        }
    }
end

-- 视频播放详情构造函数
function new_play_video_detail()
    return {
        ['video_url'] = '' -- 视频链接
    }
end

function request_get_response_body(url) 
    -- 设置并打印当前 UA
    print('当前 UA', set_ua_2_current_request_ua())
    
    -- 请求视频详情
    local resp, reqErr = http_get(url)
    if reqErr then
        err = reqErr
        log('请求视频详情失败:', reqErr)
        return nil, err
    end
    if resp.status_code ~= 200 then
        err = '请求视频详情失败:' .. resp.status_code
        log('请求视频详情失败:', resp.status_code)
        return nil, err
    end
    return resp.body, nil
end

function request_get_html_doc(url)
    local resp_body, err = request_get_response_body(url)
    if err then
        log('请求视频详情失败:', err)
        return nil, err
    end
    -- 解析 HTML
    local doc, perr = parse_html(resp_body)
    if perr then
        err = '解析视频详情失败:' .. perr
        log('解析视频详情失败:', perr)
        return nil, err
    end
    
    return doc, nil
end

function search_video(search_content)
    -- 数组 Array(search_video_result)
    local result = {}
    local err = nil

    local doc, err = request_get_html_doc(domain .. '/yhsearch/-------------.html?wd=' ..
                                 search_content .. '&submit=')
    if err then
        log('请求查询视频失败:', err)
        return result, err
    end

    -- 查询所有的 Item 这里只处理第一页，需要用户做精准搜索
    local items = doc:select('ul > li.searchlist_item')
    local index = 0
    while true do
        local elem = items:eq(index);
        if not elem or elem:html() == '' then
            log('查找元素结束')
            break
        end
        log('查找元素', index)
        local result_item = new_search_video_result();

        -- 视频名称查询
        local title_dom = elem:select_one(
                              'div.searchlist_titbox > h4.vodlist_title > a')
        if title_dom then
            result_item.name = title_dom:attr('title')
            result_item.type = title_dom:select_one('span.info_right'):text()
            log('video_name', result_item.name)

            -- 封面,链接以及评分查询
            local cover_dom = elem:select_one(
                                  'div.searchlist_img > a.vodlist_thumb')
            result_item.cover = cover_dom:attr('data-original')
            result_item.url = domain .. cover_dom:attr('href')
            result_item.score = cover_dom:select_one('span:last-child'):text()

            -- 演员查询
            local actor_dom = elem:select_one(
                                  'div.searchlist_titbox > p:nth-child(2)')
            result_item.actor = actor_dom:text():gsub('主演：', '')

            -- 导演查询
            local director_dom = elem:select_one(
                                     'div.searchlist_titbox > p:nth-child(3)')
            result_item.director = director_dom:text():gsub('导演：', '')

            -- 简介查询
            local description_dom = elem:select_one(
                                        'div.searchlist_titbox > p:nth-child(4)')
            result_item.description = description_dom:text():gsub('简介：',
                                                                  '')

            -- 其他参数当前站点没有

            -- 添加到结果数组
            table.insert(result, result_item)
        end

        -- 增加索引
        index = index + 1
    end
    return result, err
end

function get_video_detail(video_url)
    -- video_detail_result 结构体
    local result = new_video_detail_result()
    local err = nil

    local doc, err = request_get_html_doc(video_url)
    if err then
        log('请求视频详情失败:', err)
        return result, err
    end

    -- 先解析视频详情信息
    local video_detail_dom = doc:select_one('div.detail_list > div.content_box')
    if video_detail_dom then
        local cover_dom = video_detail_dom:select_one('div:nth-child(1) > a')
        if cover_dom then
            result.cover = cover_dom:attr('data-original')
            result.name = cover_dom:attr('title')
            result.url = domain .. cover_dom:attr('href')
        end

        -- 评分
        local score_dom = video_detail_dom:select_one(
                              'div:nth-child(2) > span.star_tips')
        if score_dom then result.score = score_dom:text() end

        -- 描述 Dom
        local description_dom = video_detail_dom:select_one(
                                    'div:nth-child(3) > ul')
        if description_dom then
            -- 上映年份 地区 类型 在第4个 li
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

            -- 演员 在第6个 li
            local actor_dom = description_dom:select_one('li:nth-child(6)')
            if actor_dom then
                local actor_text = actor_dom:text()
                result.actor = trim(actor_text:gsub('演员：', ''))
            end

            -- 导演 在第7个 li
            local director_dom = description_dom:select_one('li:nth-child(7)')
            if director_dom then
                local director_text = director_dom:text()
                result.director = trim(director_text:gsub('导演：', ''))
            end
        end
    end

    -- 简介
    -- 查找所有 h2 标签，获取 text 为剧情介绍的 parent 
    local h2_dom_list = doc:select('h2')
    local index = 0
    while true do
        local h2_dom = h2_dom_list:eq(index)
        if not h2_dom or h2_dom:html() == '' then break end
        if h2_dom:text() == '剧情介绍' then
            local parent_dom = h2_dom:parent()
            if parent_dom then
                result.description = parent_dom:select_one('div > span'):text()
                                         :match("剧情简介：(.+)")
                break
            end
        end
        index = index + 1
    end

    -- 资源 Dom 
    local resource_dom = doc:select_one('div.play_source')
    if not resource_dom then
        err = '解析视频详情失败: 没有找到资源 Dom'
        log('解析视频详情失败: 没有找到资源 Dom')
        return result, err
    end

    -- 得到资源站点DomList
    local source_list = resource_dom:select('div.play_source_tab > a')
    if not source_list then
        err = '解析视频详情失败: 没有找到资源站点'
        log('解析视频详情失败: 没有找到资源站点')
        return result, err
    end

    -- 遍历资源站点
    local index = 0
    while true do
        local source_item = source_list:eq(index)
        if not source_item or source_item:html() == '' then break end
        local source_item_result = {
            ['name'] = source_item:attr('alt'),
            ['episodes'] = {}
        }

        -- 剧集列表
        local episode_list = doc:select(
                                 'div.play_list_box:nth-child(' .. index + 2 ..
                                     ') > div.playlist_full > ul > li')
        if episode_list then
            local episode_index = 0
            while true do
                local episode_item = episode_list:eq(episode_index)
                if not episode_item or (episode_item:html() == '' and episode_index >= 10) then
                    break
                end
                if episode_item:html() ~= '' then
                local episode_name = episode_item:select_one('a'):text()
                local episode_url = domain ..
                                        episode_item:select_one('a')
                                            :attr('href')
                table.insert(source_item_result.episodes,
                             {['name'] = episode_name, ['url'] = episode_url})
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

    local body, err = request_get_response_body(video_url)
    if err then
        log('请求视频详情失败:', err)
        return result, err
    end
    
    -- 查找 body 中带有视频地址的 json 串
    local video_url = body:match('<script .-aaaa=(.-)</script>')
    print('body', video_url)
    if video_url then
        result.video_url = video_url
        print('视频地址', result.video_url)
    end

    return result, err
end

return get_play_video_detail('https://www.hongling8.com/yhplay/91138-2-6.html')
