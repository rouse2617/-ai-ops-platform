package prometheus

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ai-pro/internal/llm"
)

// NLQueryTranslator 自然语言查询翻译器
type NLQueryTranslator struct {
	client    *Client
	llmClient llm.Client
}

// NewNLQueryTranslator 创建自然语言查询翻译器
func NewNLQueryTranslator(client *Client, llmClient llm.Client) *NLQueryTranslator {
	return &NLQueryTranslator{
		client:    client,
		llmClient: llmClient,
	}
}

// QueryResult 查询结果
type QueryResult struct {
	Query       string                   `json:"query"`        // 原始自然语言查询
	PromQL      string                   `json:"promql"`       // 生成的 PromQL
	Explanation string                   `json:"explanation"`  // 查询解释
	Data        []map[string]interface{} `json:"data"`         // 查询结果数据
	ChartType   string                   `json:"chart_type"`   // 推荐的图表类型
	TimeRange   string                   `json:"time_range"`   // 时间范围
}

// TranslateAndExecute 翻译自然语言并执行查询
func (t *NLQueryTranslator) TranslateAndExecute(ctx context.Context, query string) (*QueryResult, error) {
	// 1. 翻译为 PromQL
	promQL, explanation, chartType, timeRange, err := t.translateToPromQL(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("翻译查询失败: %w", err)
	}

	result := &QueryResult{
		Query:       query,
		PromQL:      promQL,
		Explanation: explanation,
		ChartType:   chartType,
		TimeRange:   timeRange,
		Data:        make([]map[string]interface{}, 0),
	}

	// 2. 执行查询
	if strings.Contains(promQL, "[") {
		// 范围查询
		queryResult, err := t.client.QueryRange(ctx, promQL, timeRange)
		if err != nil {
			return result, nil // 返回结果但不包含数据
		}
		result.Data = t.parseRangeResult(queryResult)
	} else {
		// 即时查询
		queryResult, err := t.client.Query(ctx, promQL)
		if err != nil {
			return result, nil
		}
		result.Data = t.parseInstantResult(queryResult)
	}

	return result, nil
}

// translateToPromQL 将自然语言翻译为 PromQL
func (t *NLQueryTranslator) translateToPromQL(ctx context.Context, query string) (promQL, explanation, chartType, timeRange string, err error) {
	// 先尝试模式匹配
	if matched, pql, exp, chart, tr := t.patternMatch(query); matched {
		return pql, exp, chart, tr, nil
	}

	// 使用 LLM 翻译
	prompt := fmt.Sprintf(`将以下自然语言查询翻译为 PromQL：

用户查询：%s

请返回以下格式（每行一个）：
PROMQL: <生成的 PromQL 查询>
EXPLANATION: <查询的中文解释>
CHART_TYPE: <推荐的图表类型：line/bar/gauge/table>
TIME_RANGE: <时间范围，如 1h/3h/24h/7d>

常用指标参考：
- CPU: node_cpu_seconds_total, rate(node_cpu_seconds_total{mode="idle"}[5m])
- 内存: node_memory_MemTotal_bytes, node_memory_MemAvailable_bytes
- 磁盘: node_filesystem_size_bytes, node_filesystem_avail_bytes
- 网络: node_network_receive_bytes_total, node_network_transmit_bytes_total
- 负载: node_load1, node_load5, node_load15

注意：
1. 使用 rate() 计算速率
2. 使用 by() 进行分组
3. 时间范围用 [5m] 等格式`, query)

	messages := []llm.Message{
		{Role: "system", Content: "你是一个 Prometheus 专家，负责将自然语言查询翻译为 PromQL。只返回要求的格式，不要添加其他内容。"},
		{Role: "user", Content: prompt},
	}

	resp, err := t.llmClient.Chat(ctx, messages)
	if err != nil {
		return "", "", "", "", err
	}

	// 解析响应
	return t.parseTranslationResponse(resp.Content)
}

// patternMatch 模式匹配常见查询
func (t *NLQueryTranslator) patternMatch(query string) (matched bool, promQL, explanation, chartType, timeRange string) {
	query = strings.ToLower(query)

	patterns := []struct {
		keywords    []string
		promQL      string
		explanation string
		chartType   string
		timeRange   string
	}{
		{
			keywords:    []string{"cpu", "使用率", "占用"},
			promQL:      `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`,
			explanation: "计算各实例的 CPU 使用率百分比",
			chartType:   "line",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"内存", "使用率", "占用"},
			promQL:      `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`,
			explanation: "计算内存使用率百分比",
			chartType:   "line",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"磁盘", "使用率", "空间"},
			promQL:      `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)`,
			explanation: "计算根分区磁盘使用率",
			chartType:   "gauge",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"负载", "load"},
			promQL:      `node_load1`,
			explanation: "1 分钟平均负载",
			chartType:   "line",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"网络", "流量", "带宽"},
			promQL:      `rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])`,
			explanation: "网络总流量（收发合计）",
			chartType:   "line",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"tcp", "连接"},
			promQL:      `node_netstat_Tcp_CurrEstab`,
			explanation: "当前 TCP 连接数",
			chartType:   "line",
			timeRange:   "1h",
		},
		{
			keywords:    []string{"进程", "数量"},
			promQL:      `node_procs_running`,
			explanation: "运行中的进程数",
			chartType:   "line",
			timeRange:   "1h",
		},
	}

	// 提取时间范围
	timeRange = "1h"
	if strings.Contains(query, "3小时") || strings.Contains(query, "3h") {
		timeRange = "3h"
	} else if strings.Contains(query, "24小时") || strings.Contains(query, "一天") || strings.Contains(query, "1d") {
		timeRange = "24h"
	} else if strings.Contains(query, "7天") || strings.Contains(query, "一周") || strings.Contains(query, "7d") {
		timeRange = "7d"
	}

	for _, p := range patterns {
		matchCount := 0
		for _, kw := range p.keywords {
			if strings.Contains(query, kw) {
				matchCount++
			}
		}
		if matchCount >= 2 || (matchCount == 1 && len(p.keywords) == 1) {
			return true, p.promQL, p.explanation, p.chartType, timeRange
		}
	}

	return false, "", "", "", ""
}

// parseTranslationResponse 解析翻译响应
func (t *NLQueryTranslator) parseTranslationResponse(content string) (promQL, explanation, chartType, timeRange string, err error) {
	lines := strings.Split(content, "\n")

	promQLRe := regexp.MustCompile(`(?i)PROMQL:\s*(.+)`)
	expRe := regexp.MustCompile(`(?i)EXPLANATION:\s*(.+)`)
	chartRe := regexp.MustCompile(`(?i)CHART_TYPE:\s*(.+)`)
	timeRe := regexp.MustCompile(`(?i)TIME_RANGE:\s*(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := promQLRe.FindStringSubmatch(line); len(m) > 1 {
			promQL = strings.TrimSpace(m[1])
		} else if m := expRe.FindStringSubmatch(line); len(m) > 1 {
			explanation = strings.TrimSpace(m[1])
		} else if m := chartRe.FindStringSubmatch(line); len(m) > 1 {
			chartType = strings.TrimSpace(m[1])
		} else if m := timeRe.FindStringSubmatch(line); len(m) > 1 {
			timeRange = strings.TrimSpace(m[1])
		}
	}

	if promQL == "" {
		return "", "", "", "", fmt.Errorf("无法解析 PromQL")
	}

	// 默认值
	if chartType == "" {
		chartType = "line"
	}
	if timeRange == "" {
		timeRange = "1h"
	}

	return promQL, explanation, chartType, timeRange, nil
}

// parseInstantResult 解析即时查询结果
func (t *NLQueryTranslator) parseInstantResult(result *QueryResponse) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)

	for _, r := range result.Data.Result {
		item := map[string]interface{}{
			"metric": r.Metric,
		}
		if len(r.Value) >= 2 {
			item["timestamp"] = r.Value[0]
			item["value"] = r.Value[1]
		}
		data = append(data, item)
	}

	return data
}

// parseRangeResult 解析范围查询结果
func (t *NLQueryTranslator) parseRangeResult(result *QueryResponse) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)

	for _, r := range result.Data.Result {
		item := map[string]interface{}{
			"metric": r.Metric,
			"values": r.Values,
		}
		data = append(data, item)
	}

	return data
}

// GetSuggestions 获取查询建议
func (t *NLQueryTranslator) GetSuggestions(partial string) []string {
	suggestions := []string{
		"查看所有主机的 CPU 使用率",
		"对比过去 3 小时的内存增长情况",
		"显示磁盘使用率最高的主机",
		"查看网络流量趋势",
		"显示当前 TCP 连接数",
		"对比生产环境各机器的负载",
		"查看过去 24 小时的错误率",
	}

	if partial == "" {
		return suggestions
	}

	filtered := make([]string, 0)
	for _, s := range suggestions {
		if strings.Contains(strings.ToLower(s), strings.ToLower(partial)) {
			filtered = append(filtered, s)
		}
	}

	return filtered
}
