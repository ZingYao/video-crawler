package entities

type (
	VideoSourceEntity struct {
		Id                    string                `json:"id"`
		Name                  string                `json:"name"`
		Domain                string                `json:"domain"`
		Sort                  int                   `json:"sort"`        // 排序 数字越大越靠前
		Status                int                   `json:"status"`      // 0: 禁用 1: 正常 2: 维护中 3: 不可用
		SourceType            int                   `json:"source_type"` // 0: 综合 1: 短剧 2: 电影 3: 电视剧 4: 综艺 5: 动漫 6: 纪录片 7: 其他
		SearchConfig          SearchConfig          `json:"search_config"`
		VideoDescPageConfig   VideoDescPageConfig   `json:"video_desc_page_config"`
		VideoPlayerPageConfig VideoPlayerPageConfig `json:"video_player_page_config"`
	}

	SearchConfig struct {
		SearchPath                string `json:"search_path"`                  // 搜索页面路径
		SearchMethod              string `json:"search_method"`                // get or post
		SearchKeyPosition         string `json:"search_key_position"`          // url or body
		SearchKey                 string `json:"search_key"`                   // 搜索关键字
		PageKeyPosition           string `json:"page_key_position"`            // url or body
		PageKey                   string `json:"page_key"`                     // 页码关键字
		SearchTypeKeyPosition     string `json:"search_type_key_position"`     // url or body
		SearchTypeKey             string `json:"search_type_key"`              // 搜索类型关键字
		TotalCountCssFilter       string `json:"total_count_css_filter"`       // 总页数css过滤器
		TotalCountRegex           string `json:"total_count_regex"`            // 总页数正则匹配
		CurrentPageCssFilter      string `json:"current_page_css_filter"`      // 当前页码css过滤器
		CurrentPageRegex          string `json:"current_page_regex"`           // 当前页码正则匹配
		VideoCardCssFilter        string `json:"video_card_css_filter"`        // 视频卡片css过滤器
		VideoCoverImageCssFilter  string `json:"video_cover_image_css_filter"` // 视频封面图片css过滤器
		VideoTitleCssFilter       string `json:"video_title_css_filter"`       // 视频标题css过滤器
		VideoDetailUrlCssFilter   string `json:"video_detail_url_css_filter"`  // 视频详情url css过滤器
		VideoPlayerUrlCssFilter   string `json:"video_player_url_css_filter"`  // 视频播放url css过滤器
		VideoUrlIsAbsolute        bool   `json:"video_url_is_absolute"`        // 视频 Url 是否绝对路径
		VideoDirectorCssFilter    string `json:"video_director_css_filter"`    // 视频导演css过滤器
		VideoDirectorRegex        string `json:"video_director_regex"`         // 视频导演正则匹配
		VideoActorCssFilter       string `json:"video_actor_css_filter"`       // 视频演员css过滤器
		VideoActorRegex           string `json:"video_actor_regex"`            // 视频演员正则匹配
		VideoYearCssFilter        string `json:"video_year_css_filter"`        // 视频年份css过滤器
		VideoYearRegex            string `json:"video_year_regex"`             // 视频年份正则匹配
		VideoRegionCssFilter      string `json:"video_region_css_filter"`      // 视频地区css过滤器
		VideoRegionRegex          string `json:"video_region_regex"`           // 视频地区正则匹配
		VideoTypeCssFilter        string `json:"video_type_css_filter"`        // 视频类型css过滤器
		VideoLanguageCssFilter    string `json:"video_language_css_filter"`    // 视频语言css过滤器
		VideoDescriptionCssFilter string `json:"video_description_css_filter"` // 视频描述css过滤器
	}

	VideoDescPageConfig struct {
		VideoCardCssFilter        string `json:"video_card_css_filter"`        // 视频卡片css过滤器
		VideoCoverImageCssFilter  string `json:"video_cover_image_css_filter"` // 视频封面图片css过滤器
		VideoTitleCssFilter       string `json:"video_title_css_filter"`       // 视频标题css过滤器
		VideoPlayerUrlCssFilter   string `json:"video_player_url_css_filter"`  // 视频播放url css过滤器
		VideoUrlIsAbsolute        bool   `json:"video_url_is_absolute"`        // 视频 Url 是否绝对路径
		VideoDirectorRegex        string `json:"video_director_regex"`         // 视频导演正则匹配
		VideoDirectorCssFilter    string `json:"video_director_css_filter"`    // 视频导演css过滤器
		VideoActorCssFilter       string `json:"video_actor_css_filter"`       // 视频演员css过滤器
		VideoYearCssFilter        string `json:"video_year_css_filter"`        // 视频年份css过滤器
		VideoAreaCssFilter        string `json:"video_area_css_filter"`        // 视频地区css过滤器
		VideoTypeCssFilter        string `json:"video_type_css_filter"`        // 视频类型css过滤器
		VideoLanguageCssFilter    string `json:"video_language_css_filter"`    // 视频语言css过滤器
		VideoDescriptionCssFilter string `json:"video_description_css_filter"` // 视频描述css过滤器
		SourceCardCssFilter       string `json:"source_card_css_filter"`       // 源卡片css过滤器
		SourceNameCssFilter       string `json:"source_name_css_filter"`       // 源名称css过滤器
		EpisodeListCssFilter      string `json:"episode_list_css_filter"`      // 剧集列表css过滤器
		EpisodeNameRegex          string `json:"episode_name_regex"`           // 剧集名称正则匹配
		EpisodeUrlRegex           string `json:"episode_url_regex"`            // 剧集url正则匹配
		EpisodeUrlIsAbsolute      bool   `json:"episode_url_is_absolute"`      // 剧集url是否绝对路径
	}

	VideoPlayerPageConfig struct {
		VideoPlayerCssFilter string `json:"video_player_css_filter"` // 视频播放器css过滤器
		VideoPlayerUrlRegex  string `json:"video_player_url_regex"`  // 视频播放url正则匹配
		VideoPlayerUrlEncode string `json:"video_player_url_encode"` // 视频播放url编码
	}
)

type VideoSourceListResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	Status     int    `json:"status"`
	SourceType int    `json:"source_type"`
}
