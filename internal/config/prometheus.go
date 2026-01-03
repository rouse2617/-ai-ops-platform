package config

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Timeout int    `yaml:"timeout"` // 秒
}

// 在 Config 结构体中添加 Prometheus 字段
// type Config struct {
//     ...
//     Prometheus PrometheusConfig `yaml:"prometheus"`
// }
