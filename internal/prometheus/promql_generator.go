package prometheus

import (
	"fmt"
	"strings"
)

// PromQLGenerator PromQL 生成器
type PromQLGenerator struct {
	templates map[string]string
}

// NewPromQLGenerator 创建 PromQL 生成器
func NewPromQLGenerator() *PromQLGenerator {
	return &PromQLGenerator{
		templates: initPromQLTemplates(),
	}
}

// GenerateQuery 自然语言转 PromQL
func (pg *PromQLGenerator) GenerateQuery(nl NLQuery) (PromQLQuery, error) {
	question := strings.ToLower(nl.Question)

	var query string
	var explanation string

	switch {
	case strings.Contains(question, "cpu") && strings.Contains(question, "超过"):
		threshold := extractThreshold(question, 80)
		query = fmt.Sprintf("(1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance)) * 100 > %d", int(threshold))
		explanation = fmt.Sprintf("查询 CPU 使用率超过 %d%% 的主机", int(threshold))

	case strings.Contains(question, "内存") && strings.Contains(question, "超过"):
		threshold := extractThreshold(question, 85)
		query = fmt.Sprintf("(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > %d", int(threshold))
		explanation = fmt.Sprintf("查询内存使用率超过 %d%% 的主机", int(threshold))

	case strings.Contains(question, "磁盘") && strings.Contains(question, "超过"):
		threshold := extractThreshold(question, 90)
		query = fmt.Sprintf("(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100 > %d", int(threshold))
		explanation = fmt.Sprintf("查询磁盘使用率超过 %d%% 的主机", int(threshold))

	case strings.Contains(question, "负载"):
		query = "node_load1"
		explanation = "查询系统 1 分钟负载"

	case strings.Contains(question, "网络"):
		query = "rate(node_network_receive_bytes_total[5m])"
		explanation = "查询网络接收速率"

	default:
		query = "up"
		explanation = "默认查询：主机在线状态"
	}

	return PromQLQuery{
		Query:       query,
		Range:       "1h",
		Step:        "1m",
		Explanation: explanation,
	}, nil
}

// OptimizeQuery 优化 PromQL
func (pg *PromQLGenerator) OptimizeQuery(query string) (string, error) {
	// 移除多余的空格
	query = strings.TrimSpace(query)

	// 添加聚合函数优化
	if !strings.Contains(query, "by") && !strings.Contains(query, "without") {
		if strings.Contains(query, "node_") {
			query = fmt.Sprintf("avg(%s) by (instance)", query)
		}
	}

	return query, nil
}

// ExplainQuery 解释 PromQL
func (pg *PromQLGenerator) ExplainQuery(query string) (string, error) {
	explanation := "PromQL 查询解释:\n"

	if strings.Contains(query, "node_cpu") {
		explanation += "- 查询 CPU 相关指标\n"
	}
	if strings.Contains(query, "node_memory") {
		explanation += "- 查询内存相关指标\n"
	}
	if strings.Contains(query, "node_filesystem") {
		explanation += "- 查询文件系统相关指标\n"
	}
	if strings.Contains(query, "rate") {
		explanation += "- 计算指标的速率（每秒变化量）\n"
	}
	if strings.Contains(query, "avg") {
		explanation += "- 计算平均值\n"
	}
	if strings.Contains(query, "by") {
		explanation += "- 按指定标签分组\n"
	}

	return explanation, nil
}

// 私有方法

func extractThreshold(question string, defaultValue float64) float64 {
	// 简单的阈值提取逻辑
	parts := strings.Fields(question)
	for i, part := range parts {
		if strings.Contains(part, "%") || (i > 0 && strings.Contains(parts[i-1], "超过")) {
			// 尝试解析数字
			var threshold float64
			fmt.Sscanf(part, "%f", &threshold)
			if threshold > 0 {
				return threshold
			}
		}
	}
	return defaultValue
}

func initPromQLTemplates() map[string]string {
	return map[string]string{
		"cpu_usage":     "(1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance)) * 100",
		"memory_usage":  "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
		"disk_usage":    "(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100",
		"load_average":  "node_load1",
		"network_in":    "rate(node_network_receive_bytes_total[5m])",
		"network_out":   "rate(node_network_transmit_bytes_total[5m])",
		"disk_io_read":  "rate(node_disk_read_bytes_total[5m])",
		"disk_io_write": "rate(node_disk_written_bytes_total[5m])",
	}
}
