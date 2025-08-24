package entities

// AdvancedTestResult 高级调试结果
type AdvancedTestResult struct {
	Original  interface{} `json:"original"`  // 原始结果
	Converted interface{} `json:"converted"` // 转换后的结果
}

// AdvancedTestRequest 高级调试请求
type AdvancedTestRequest struct {
	Script string                 `json:"script"` // 脚本内容
	Method string                 `json:"method"` // 方法名称
	Params map[string]interface{} `json:"params"` // 参数
}
