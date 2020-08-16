package base

import "net/http"

// CollectorConfig 用于配置Collector
type CollectorConfig struct {
	Name    string                 `json:"name"`
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Header  http.Header            `json:"header"`
	Content string                 `json:"content"`
	Args    map[string]interface{} `json:"args"`
}

// CollectorConfigs Config数组，保存于一个json文件
type CollectorConfigs struct {
	Configs []CollectorConfig `json:"configs"`
}
