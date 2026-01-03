package prometheus

import (
	"context"
	"fmt"
	"time"
)

// AlertTranslator 告警翻译器
type AlertTranslator struct {
	semanticEngine *SemanticEngine
}

// NewAlertTranslator 创建告警翻译器
func NewAlertTranslator(semanticEngine *SemanticEngine) *AlertTranslator {
	return &AlertTranslator{
		semanticEngine: semanticEngine,
	}
}

// TranslateAlert 翻译告警
func (at *AlertTranslator) TranslateAlert(alert AlertMetadata, context AlertContext) string {
	return at.semanticEngine.GenerateAlertDescription(alert, context)
}

// TranslateAlerts 批量翻译告警
func (at *AlertTranslator) TranslateAlerts(alerts []AlertMetadata, contexts []AlertContext) []string {
	translations := make([]string, 0)
	for i, alert := range alerts {
		if i < len(contexts) {
			translation := at.TranslateAlert(alert, contexts[i])
			translations = append(translations, translation)
		}
	}
	return translations
}

// NLToPromQL 自然语言转 PromQL
type NLToPromQL struct {
	generator *PromQLGenerator
}

// NewNLToPromQL 创建自然语言转 PromQL
func NewNLToPromQL(generator *PromQLGenerator) *NLToPromQL {
	return &NLToPromQL{
		generator: generator,
	}
}

// Convert 转换
func (np *NLToPromQL) Convert(question string) (PromQLQuery, error) {
	nlQuery := NLQuery{
		Question: question,
	}
	return np.generator.GenerateQuery(nlQuery)
}

// QueryResultVisualizer 查询结果可视化
type QueryResultVisualizer struct {
	client *PrometheusClient
}

// NewQueryResultVisualizer 创建查询结果可视化
func NewQueryResultVisualizer(client *PrometheusClient) *QueryResultVisualizer {
	return &QueryResultVisualizer{
		client: client,
	}
}

// VisualizeQuery 可视化查询
func (qrv *QueryResultVisualizer) VisualizeQuery(ctx context.Context, query string, start, end time.Time) (map[string]interface{}, error) {
	step := (end.Sub(start)) / 100 // 100 个数据点
	if step < time.Minute {
		step = time.Minute
	}

	matrix, err := qrv.client.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, err
	}

	visualization := map[string]interface{}{
		"query":  query,
		"start":  start,
		"end":    end,
		"series": make([]map[string]interface{}, 0),
	}

	for _, series := range matrix {
		seriesData := map[string]interface{}{
			"metric": series.Metric,
			"values": make([]interface{}, 0),
		}

		for _, sample := range series.Values {
			seriesData["values"] = append(seriesData["values"].([]interface{}), map[string]interface{}{
				"timestamp": sample.Timestamp,
				"value":     sample.Value,
			})
		}

		visualization["series"] = append(visualization["series"].([]map[string]interface{}), seriesData)
	}

	return visualization, nil
}

// DashboardContextAnalyzer 看板上下文分析器
type DashboardContextAnalyzer struct {
	dashboardGenerator *DashboardGenerator
}

// NewDashboardContextAnalyzer 创建看板上下文分析器
func NewDashboardContextAnalyzer(dashboardGenerator *DashboardGenerator) *DashboardContextAnalyzer {
	return &DashboardContextAnalyzer{
		dashboardGenerator: dashboardGenerator,
	}
}

// AnalyzeContext 分析上下文
func (dca *DashboardContextAnalyzer) AnalyzeContext(ctx context.Context, conversationContext map[string]interface{}) ([]PanelConfig, error) {
	// 根据对话上下文推荐相关面板
	recommendedPanels := make([]PanelConfig, 0)

	// 检查上下文中的关键词
	if keywords, ok := conversationContext["keywords"].([]string); ok {
		for _, keyword := range keywords {
			panels := dca.getPanelsByKeyword(keyword)
			recommendedPanels = append(recommendedPanels, panels...)
		}
	}

	// 检查��下文中的指标
	if metrics, ok := conversationContext["metrics"].([]string); ok {
		for _, metric := range metrics {
			panel := dca.getPanelByMetric(metric)
			if panel != nil {
				recommendedPanels = append(recommendedPanels, *panel)
			}
		}
	}

	return recommendedPanels, nil
}

// getPanelsByKeyword 根据关键词获取面板
func (dca *DashboardContextAnalyzer) getPanelsByKeyword(keyword string) []PanelConfig {
	panels := make([]PanelConfig, 0)

	switch keyword {
	case "cpu":
		panels = append(panels, PanelConfig{
			ID:    fmt.Sprintf("panel-cpu-%d", time.Now().UnixNano()),
			Title: "CPU 使用率",
			Type:  "graph",
			Query: "(1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance)) * 100",
		})
	case "memory":
		panels = append(panels, PanelConfig{
			ID:    fmt.Sprintf("panel-memory-%d", time.Now().UnixNano()),
			Title: "内存使用率",
			Type:  "graph",
			Query: "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
		})
	case "disk":
		panels = append(panels, PanelConfig{
			ID:    fmt.Sprintf("panel-disk-%d", time.Now().UnixNano()),
			Title: "磁盘使用率",
			Type:  "graph",
			Query: "(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100",
		})
	case "network":
		panels = append(panels, PanelConfig{
			ID:    fmt.Sprintf("panel-network-%d", time.Now().UnixNano()),
			Title: "网络流量",
			Type:  "graph",
			Query: "rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])",
		})
	}

	return panels
}

// getPanelByMetric 根据指标获取面板
func (dca *DashboardContextAnalyzer) getPanelByMetric(metric string) *PanelConfig {
	return &PanelConfig{
		ID:    fmt.Sprintf("panel-%s-%d", metric, time.Now().UnixNano()),
		Title: metric,
		Type:  "graph",
		Query: metric,
	}
}
