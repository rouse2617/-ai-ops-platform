package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-pro/internal/llm"
)

// SemanticTranslator 语义化翻译器 - 将 Prometheus 告警转换为自然语言
type SemanticTranslator struct {
	client    *Client
	llmClient llm.Client
}

// NewSemanticTranslator 创建语义化翻译器
func NewSemanticTranslator(client *Client, llmClient llm.Client) *SemanticTranslator {
	return &SemanticTranslator{
		client:    client,
		llmClient: llmClient,
	}
}

// AlertContext 告警上下文
type AlertContext struct {
	AlertName   string                 `json:"alert_name"`
	Instance    string                 `json:"instance"`
	Severity    string                 `json:"severity"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	RelatedMetrics map[string]float64  `json:"related_metrics"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SemanticAlert 语义化告警
type SemanticAlert struct {
	Original    AlertContext `json:"original"`
	Summary     string       `json:"summary"`      // 简短摘要
	Description string       `json:"description"`  // 详细描述
	RootCause   string       `json:"root_cause"`   // 可能的根因
	Suggestion  string       `json:"suggestion"`   // 建议操作
	Confidence  float64      `json:"confidence"`   // 置信度
}

// TranslateAlert 将告警翻译为自然语言
func (t *SemanticTranslator) TranslateAlert(ctx context.Context, alert AlertContext) (*SemanticAlert, error) {
	// 1. 获取相关指标
	relatedMetrics, err := t.fetchRelatedMetrics(ctx, alert)
	if err != nil {
		// 继续处理，即使获取相关指标失败
		relatedMetrics = make(map[string]float64)
	}
	alert.RelatedMetrics = relatedMetrics

	// 2. 构建 LLM 提示
	prompt := t.buildTranslationPrompt(alert)

	// 3. 调用 LLM 生成语义化描述
	messages := []llm.Message{
		{Role: "system", Content: semanticTranslatorSystemPrompt},
		{Role: "user", Content: prompt},
	}

	resp, err := t.llmClient.Chat(ctx, messages)
	if err != nil {
		// 如果 LLM 调用失败，返回基础翻译
		return t.fallbackTranslation(alert), nil
	}

	// 4. 解析响应
	return t.parseResponse(alert, resp.Content), nil
}

// fetchRelatedMetrics 获取相关指标
func (t *SemanticTranslator) fetchRelatedMetrics(ctx context.Context, alert AlertContext) (map[string]float64, error) {
	metrics := make(map[string]float64)
	instance := alert.Instance

	// 根据告警类型获取相关指标
	queries := t.getRelatedQueries(alert.AlertName, instance)

	for name, query := range queries {
		result, err := t.client.Query(ctx, query)
		if err != nil {
			continue
		}
		if len(result.Data.Result) > 0 {
			if v, ok := result.Data.Result[0].Value[1].(string); ok {
				var val float64
				fmt.Sscanf(v, "%f", &val)
				metrics[name] = val
			}
		}
	}

	return metrics, nil
}

// getRelatedQueries 根据告警类型获取相关查询
func (t *SemanticTranslator) getRelatedQueries(alertName, instance string) map[string]string {
	queries := make(map[string]string)

	switch {
	case strings.Contains(strings.ToLower(alertName), "cpu"):
		queries["memory_usage"] = fmt.Sprintf(`100 - (node_memory_MemAvailable_bytes{instance="%s"} / node_memory_MemTotal_bytes{instance="%s"} * 100)`, instance, instance)
		queries["load_1m"] = fmt.Sprintf(`node_load1{instance="%s"}`, instance)
		queries["process_count"] = fmt.Sprintf(`count(node_processes_state{instance="%s"})`, instance)
		queries["io_wait"] = fmt.Sprintf(`rate(node_cpu_seconds_total{instance="%s",mode="iowait"}[5m]) * 100`, instance)

	case strings.Contains(strings.ToLower(alertName), "memory"):
		queries["cpu_usage"] = fmt.Sprintf(`100 - (avg(rate(node_cpu_seconds_total{instance="%s",mode="idle"}[5m])) * 100)`, instance)
		queries["swap_usage"] = fmt.Sprintf(`(node_memory_SwapTotal_bytes{instance="%s"} - node_memory_SwapFree_bytes{instance="%s"}) / node_memory_SwapTotal_bytes{instance="%s"} * 100`, instance, instance, instance)
		queries["oom_kills"] = fmt.Sprintf(`increase(node_vmstat_oom_kill{instance="%s"}[1h])`, instance)

	case strings.Contains(strings.ToLower(alertName), "disk"):
		queries["inode_usage"] = fmt.Sprintf(`100 - (node_filesystem_files_free{instance="%s"} / node_filesystem_files{instance="%s"} * 100)`, instance, instance)
		queries["io_util"] = fmt.Sprintf(`rate(node_disk_io_time_seconds_total{instance="%s"}[5m]) * 100`, instance)
		queries["write_rate"] = fmt.Sprintf(`rate(node_disk_written_bytes_total{instance="%s"}[5m])`, instance)

	case strings.Contains(strings.ToLower(alertName), "network"):
		queries["tcp_connections"] = fmt.Sprintf(`node_netstat_Tcp_CurrEstab{instance="%s"}`, instance)
		queries["tcp_retrans"] = fmt.Sprintf(`rate(node_netstat_Tcp_RetransSegs{instance="%s"}[5m])`, instance)
		queries["rx_errors"] = fmt.Sprintf(`rate(node_network_receive_errs_total{instance="%s"}[5m])`, instance)
	}

	return queries
}

// buildTranslationPrompt 构建翻译提示
func (t *SemanticTranslator) buildTranslationPrompt(alert AlertContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("告警名称: %s\n", alert.AlertName))
	sb.WriteString(fmt.Sprintf("实例: %s\n", alert.Instance))
	sb.WriteString(fmt.Sprintf("严重程度: %s\n", alert.Severity))
	sb.WriteString(fmt.Sprintf("当前值: %.2f\n", alert.Value))
	sb.WriteString(fmt.Sprintf("阈值: %.2f\n", alert.Threshold))
	sb.WriteString(fmt.Sprintf("时间: %s\n", alert.Timestamp.Format("2006-01-02 15:04:05")))

	if len(alert.RelatedMetrics) > 0 {
		sb.WriteString("\n相关指标:\n")
		for name, value := range alert.RelatedMetrics {
			sb.WriteString(fmt.Sprintf("- %s: %.2f\n", name, value))
		}
	}

	sb.WriteString("\n请分析这个告警，给出：\n")
	sb.WriteString("1. 简短摘要（一句话）\n")
	sb.WriteString("2. 详细描述（结合相关指标分析）\n")
	sb.WriteString("3. 可能的根因\n")
	sb.WriteString("4. 建议操作\n")

	return sb.String()
}

// parseResponse 解析 LLM 响应
func (t *SemanticTranslator) parseResponse(alert AlertContext, content string) *SemanticAlert {
	result := &SemanticAlert{
		Original:   alert,
		Confidence: 0.8,
	}

	// 简单解析，按段落分割
	parts := strings.Split(content, "\n\n")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		switch i {
		case 0:
			result.Summary = part
		case 1:
			result.Description = part
		case 2:
			result.RootCause = part
		case 3:
			result.Suggestion = part
		}
	}

	// 如果解析不完整，使用整个内容作为描述
	if result.Summary == "" {
		result.Summary = content
	}

	return result
}

// fallbackTranslation 降级翻译
func (t *SemanticTranslator) fallbackTranslation(alert AlertContext) *SemanticAlert {
	summary := fmt.Sprintf("检测到 %s 在 %s 上触发告警", alert.AlertName, alert.Instance)

	var description strings.Builder
	description.WriteString(fmt.Sprintf("告警 %s 已触发。", alert.AlertName))
	description.WriteString(fmt.Sprintf("当前值 %.2f 超过阈值 %.2f。", alert.Value, alert.Threshold))

	if len(alert.RelatedMetrics) > 0 {
		description.WriteString("\n\n相关指标情况：")
		for name, value := range alert.RelatedMetrics {
			description.WriteString(fmt.Sprintf("\n- %s: %.2f", name, value))
		}
	}

	return &SemanticAlert{
		Original:    alert,
		Summary:     summary,
		Description: description.String(),
		RootCause:   "需要进一步分析",
		Suggestion:  "请检查相关服务和日志",
		Confidence:  0.5,
	}
}

const semanticTranslatorSystemPrompt = `你是一个运维专家，负责将 Prometheus 告警翻译成易于理解的自然语言。

你的任务是：
1. 分析告警及其相关指标
2. 用通俗易懂的语言描述问题
3. 结合相关指标推断可能的根因
4. 给出具体可行的建议

输出格式：
- 第一段：简短摘要（一句话概括问题）
- 第二段：详细描述��结合相关指标分析）
- 第三段：可能的根因
- 第四段：建议操作

注意：
- 使用中文
- 避免技术术语，用通俗语言
- 结合相关指标进行关联分析
- 给出具体可操作的建议`
