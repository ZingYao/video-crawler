package entities

type (
	VideoSourceEntity struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Domain     string `json:"domain"`
		Sort       int    `json:"sort"`        // 排序 数字越大越靠前
		Status     int    `json:"status"`      // 0: 禁用 1: 正常 2: 维护中 3: 不可用
		SourceType int    `json:"source_type"` // 0: 综合 1: 短剧 2: 电影 3: 电视剧 4: 综艺 5: 动漫 6: 纪录片 7: 其他
		LuaScript  string `json:"lua_script"`  // Lua脚本内容
	}
)

type VideoSourceListResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	Sort       int    `json:"sort"` // 排序值
	Status     int    `json:"status"`
	SourceType int    `json:"source_type"`
	LuaScript  string `json:"lua_script"` // Lua脚本内容
}
