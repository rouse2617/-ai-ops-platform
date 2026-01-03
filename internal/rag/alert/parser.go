package alert

import (
	"fmt"
	"time"

	"ai-ops/internal/rag"
	"ai-ops/internal/rag/retriever"
)

// Alert 告警结构
type Alert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
}

// AlertParser 告警解析器
type AlertParser struct {
	queryRewriter *retriever.QueryRewriter
}

// NewAlertParser 创建告警解析器
func NewAlertParser(queryRewriter *retriever.QueryRewriter) *AlertParser {
	return &AlertParser{
		queryRewriter: queryRewriter,
	}
}

// ParseToQuery 解析告警为检索查询
func (ap *AlertParser) ParseToQuery(alert interface{}) rag.SearchQuery {
	// 类型断言
	a, ok := alert.(Alert)
	if !ok {
		return rag.SearchQuery{}
	}

	// 构建检索查询
	query := fmt.Sprintf("%s %s %s",
		a.Labels["service"],
		a.Labels["alertname"],
		a.Annotations["summary"],
	)

	// 提取过滤条件
	filters := map[string]string{
		"service":  a.Labels["service"],
		"severity": a.Labels["severity"],
		"category": ap.inferCategory(a),
	}

	// 时间范围过滤（优先检索近期案例）
	filters["occurred_after"] = time.Now().AddDate(0, -6, 0).Format(time.RFC3339)

	return rag.SearchQuery{
		Query:   query,
		TopK:    10,
		Filters: filters,
		Rerank:  true,
	}
}

// inferCategory 推断告警类别
func (ap *AlertParser) inferCategory(alert Alert) string {
	alertName := alert.Labels["alertname"]

	// 简单的规则匹配
	if contains(alertName, []string{"CPU", "Memory", "Disk"}) {
		return "system"
	}
	if contains(alertName, []string{"Database", "MySQL", "PostgreSQL", "Redis"}) {
		return "database"
	}
	if contains(alertName, []string{"Network", "Connection", "Timeout"}) {
		return "network"
	}

	return "application"
}

func contains(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(s) >= len(keyword) && s[:len(keyword)] == keyword {
			return true
		}
	}
	return false
}
