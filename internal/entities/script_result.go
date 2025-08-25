package entities

import (
	"encoding/json"
	"fmt"
)

// SearchVideoResult 搜索视频结果结构体
type SearchVideoResult struct {
	Cover       string `json:"cover"`        // 视频封面
	Name        string `json:"name"`         // 视频名称
	Type        string `json:"type"`         // 视频类型
	URL         string `json:"url"`          // 视频链接
	Actor       string `json:"actor"`        // 演员
	Director    string `json:"director"`     // 导演
	ReleaseDate string `json:"release_date"` // 上映日期
	Region      string `json:"region"`       // 地区
	Language    string `json:"language"`     // 语言
	Description string `json:"description"`  // 描述
	Score       string `json:"score"`        // 评分
}

// EpisodeItem 剧集对象结构体
type EpisodeItem struct {
	Name string `json:"name"` // 剧集名称（如：'第1集'、'第2集'、'大结局'等）
	URL  string `json:"url"`  // 剧集播放链接
}

// SourceItem 来源站点对象结构体
type SourceItem struct {
	Name     string        `json:"name"`     // 来源站点名称（如：'线路1'、'线路2'、'备用线路'等）
	Episodes []EpisodeItem `json:"episodes"` // 剧集列表数组
}

// VideoDetailResult 视频详情结果结构体
type VideoDetailResult struct {
	Cover       string       `json:"cover"`        // 视频封面
	Name        string       `json:"name"`         // 视频名称
	URL         string       `json:"url"`          // 视频链接
	Score       string       `json:"score"`        // 评分
	ReleaseDate string       `json:"release_date"` // 上映日期
	Region      string       `json:"region"`       // 地区
	Actor       string       `json:"actor"`        // 演员
	Director    string       `json:"director"`     // 导演
	Description string       `json:"description"`  // 描述
	Language    string       `json:"language"`     // 语言
	Source      []SourceItem `json:"source"`       // 数组：来源站点及剧集列表
}

// PlayVideoDetailResult 播放详情结果结构体
type PlayVideoDetailResult struct {
	VideoURL string `json:"video_url"` // 视频链接
}

// ScriptResult 脚本执行结果结构体
type ScriptResult struct {
	Data interface{} `json:"data"` // 具体的数据内容
	Err  string      `json:"err"`  // 错误信息，为空表示成功
}

// ValidateSearchVideoResult 验证搜索视频结果
func ValidateSearchVideoResult(data interface{}) ([]SearchVideoResult, error) {
	if data == nil {
		return nil, nil
	}

	// 尝试转换为数组
	if results, ok := data.([]interface{}); ok {
		validResults := make([]SearchVideoResult, 0, len(results))
		for _, item := range results {
			if result, ok := item.(map[string]interface{}); ok {
				validResult := SearchVideoResult{
					Cover:       getString(result, "cover"),
					Name:        getString(result, "name"),
					Type:        getString(result, "type"),
					URL:         getString(result, "url"),
					Actor:       getString(result, "actor"),
					Director:    getString(result, "director"),
					ReleaseDate: getString(result, "release_date"),
					Region:      getString(result, "region"),
					Language:    getString(result, "language"),
					Description: getString(result, "description"),
					Score:       getString(result, "score"),
				}
				validResults = append(validResults, validResult)
			}
		}
		return validResults, nil
	}

	return nil, nil
}

// ValidateVideoDetailResult 验证视频详情结果
func ValidateVideoDetailResult(data interface{}) (*VideoDetailResult, error) {
	if data == nil {
		return nil, nil
	}

	if result, ok := data.(map[string]interface{}); ok {
		// 验证并转换 source 字段
		var sources []SourceItem
		if sourceData, exists := result["source"]; exists {
			if sourceArray, ok := sourceData.([]interface{}); ok {
				sources = make([]SourceItem, 0, len(sourceArray))
				for _, sourceItem := range sourceArray {
					if sourceMap, ok := sourceItem.(map[string]interface{}); ok {
						source := SourceItem{
							Name: getString(sourceMap, "name"),
						}

						// 验证并转换 episodes 字段
						if episodesData, exists := sourceMap["episodes"]; exists {
							if episodesArray, ok := episodesData.([]interface{}); ok {
								episodes := make([]EpisodeItem, 0, len(episodesArray))
								for _, episodeItem := range episodesArray {
									if episodeMap, ok := episodeItem.(map[string]interface{}); ok {
										episode := EpisodeItem{
											Name: getString(episodeMap, "name"),
											URL:  getString(episodeMap, "url"),
										}
										episodes = append(episodes, episode)
									}
								}
								source.Episodes = episodes
							}
						}

						sources = append(sources, source)
					}
				}
			}
		}

		validResult := &VideoDetailResult{
			Cover:       getString(result, "cover"),
			Name:        getString(result, "name"),
			URL:         getString(result, "url"),
			Score:       getString(result, "score"),
			ReleaseDate: getString(result, "release_date"),
			Region:      getString(result, "region"),
			Actor:       getString(result, "actor"),
			Director:    getString(result, "director"),
			Description: getString(result, "description"),
			Language:    getString(result, "language"),
			Source:      sources,
		}
		return validResult, nil
	}

	return nil, nil
}

// ValidatePlayVideoDetailResult 验证播放详情结果
func ValidatePlayVideoDetailResult(data interface{}) (*PlayVideoDetailResult, error) {
	if data == nil {
		return nil, nil
	}

	if result, ok := data.(map[string]interface{}); ok {
		validResult := &PlayVideoDetailResult{
			VideoURL: getString(result, "video_url"),
		}
		return validResult, nil
	}

	return nil, nil
}

// FilterSearchVideoResult 通过JSON序列化反序列化过滤搜索视频结果字段
func FilterSearchVideoResult(data interface{}) ([]SearchVideoResult, error) {
	if data == nil {
		return nil, nil
	}

	// 先序列化为JSON字节数组
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	// 反序列化到结构体，自动过滤未定义字段
	var results []SearchVideoResult
	err = json.Unmarshal(jsonBytes, &results)
	if err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	return results, nil
}

// FilterVideoDetailResult 通过JSON序列化反序列化过滤视频详情结果字段
func FilterVideoDetailResult(data interface{}) (*VideoDetailResult, error) {
	if data == nil {
		return nil, nil
	}

	// 先序列化为JSON字节数组
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	// 反序列化到结构体，自动过滤未定义字段
	var result VideoDetailResult
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	return &result, nil
}

// FilterPlayVideoDetailResult 通过JSON序列化反序列化过滤播放详情结果字段
func FilterPlayVideoDetailResult(data interface{}) (*PlayVideoDetailResult, error) {
	if data == nil {
		return nil, nil
	}

	// 先序列化为JSON字节数组
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	// 反序列化到结构体，自动过滤未定义字段
	var result PlayVideoDetailResult
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	return &result, nil
}

// getString 安全地从 map 中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
		// 如果不是字符串，尝试转换为字符串
		return toString(val)
	}
	return ""
}

// toString 将任意类型转换为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
