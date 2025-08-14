package entities

import "time"

// 搜索历史
type SearchHistory struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	Keyword   string    `json:"keyword"`
	SourceId  string    `json:"source_id"`
	CreatedAt time.Time `json:"created_at"`
}

// 视频观看历史
type VideoHistory struct {
	Id         string    `json:"id"`
	UserId     string    `json:"user_id"`
	VideoId    string    `json:"video_id"`
	VideoTitle string    `json:"video_title"`
	VideoUrl   string    `json:"video_url"`
	SourceId   string    `json:"source_id"`
	SourceName string    `json:"source_name"`
	WatchTime  int64     `json:"watch_time"` // 观看时长（秒）
	Progress   float64   `json:"progress"`   // 观看进度（0-1）
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// 登录历史
type LoginHistory struct {
	LoginAt   time.Time `json:"login_at"`
	Ip        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Password  string    `json:"password,omitempty"`
	Success   bool      `json:"success"`
	Token     string    `json:"token"`
}

// 历史记录请求参数
type HistoryRequest struct {
	UserId string `json:"user_id" form:"user_id"`
}

// 历史记录响应
type HistoryResponse struct {
	Total int         `json:"total"`
	Data  interface{} `json:"data"`
}
