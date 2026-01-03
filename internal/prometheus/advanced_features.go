package prometheus

import (
	"context"
	"fmt"
	"time"
)

// IncidentAnalyzer 故障分析器
type IncidentAnalyzer struct {
	incidentReplay *IncidentReplayQuery
	semanticEngine *SemanticEngine
}

// NewIncidentAnalyzer 创建故障分析器
func NewIncidentAnalyzer(incidentReplay *IncidentReplayQuery, semanticEngine *SemanticEngine) *IncidentAnalyzer {
	return &IncidentAnalyzer{
		incidentReplay: incidentReplay,
		semanticEngine: semanticEngine,
	}
}

// AnalyzeIncident 分析故障
func (ia *IncidentAnalyzer) AnalyzeIncident(ctx context.Context, incidentTime time.Time, host string) (map[string]interface{}, error) {
	timeline, err := ia.incidentReplay.QueryIncidentMetrics(ctx, incidentTime, host)
	if err != nil {
		return nil, err
	}

	report := ia.incidentReplay.GenerateIncidentReport(timeline)

	analysis := map[string]interface{}{
		"incident_time":     incidentTime,
		"host":              host,
		"affected_metrics":  report.AffectedMetrics,
		"root_causes":       report.RootCauses,
		"recommendations":   report.Recommendations,
		"timeline":          timeline,
		"severity":          ia.calculateSeverity(report),
		"estimated_impact":  ia.estimateImpact(report),
	}

	return analysis, nil
}

// calculateSeverity 计算严重程度
func (ia *IncidentAnalyzer) calculateSeverity(report *IncidentReport) string {
	if len(report.AffectedMetrics) > 3 {
		return "critical"
	} else if len(report.AffectedMetrics) > 1 {
		return "high"
	} else if len(report.AffectedMetrics) > 0 {
		return "medium"
	}
	return "low"
}

// estimateImpact 估计影响
func (ia *IncidentAnalyzer) estimateImpact(report *IncidentReport) map[string]interface{} {
	impact := map[string]interface{}{
		"affected_services": []string{},
		"estimated_users":   0,
		"data_loss_risk":    false,
	}

	for _, metric := range report.AffectedMetrics {
		switch metric {
		case "node_cpu_seconds_total":
			impact["affected_services"] = append(impact["affected_services"].([]string), "compute")
		case "node_memory_MemAvailable_bytes":
			impact["affected_services"] = append(impact["affected_services"].([]string), "memory")
		case "node_disk_io_time_seconds_total":
			impact["affected_services"] = append(impact["affected_services"].([]string), "storage")
			impact["data_loss_risk"] = true
		}
	}

	return impact
}

// AnomalyDetectionEngine 异常检测引擎
type AnomalyDetectionEngine struct {
	thresholdDetector *ThresholdDetector
	semanticEngine    *SemanticEngine
}

// NewAnomalyDetectionEngine 创建异常检测引擎
func NewAnomalyDetectionEngine(thresholdDetector *ThresholdDetector, semanticEngine *SemanticEngine) *AnomalyDetectionEngine {
	return &AnomalyDetectionEngine{
		thresholdDetector: thresholdDetector,
		semanticEngine:    semanticEngine,
	}
}

// DetectAndTranslate 检测并翻译异常
func (ade *AnomalyDetectionEngine) DetectAndTranslate(metric MetricSample, baseline BaselineMetrics) map[string]interface{} {
	anomaly := ade.thresholdDetector.DetectAnomaly(metric, baseline)

	translation := ade.semanticEngine.TranslateMetric(metric)
	recommendations := ade.semanticEngine.GenerateRecommendations(anomaly)

	return map[string]interface{}{
		"is_anomaly":      anomaly.IsAnomaly,
		"severity":        anomaly.Severity,
		"deviation":       anomaly.Deviation,
		"recommendation":  anomaly.Recommendation,
		"translation":     translation,
		"recommendations": recommendations,
		"baseline":        baseline,
	}
}

// PlaybookRecommender 剧本推荐器
type PlaybookRecommender struct {
	playbookEngine *PlaybookEngine
}

// NewPlaybookRecommender 创建剧本推荐器
func NewPlaybookRecommender(playbookEngine *PlaybookEngine) *PlaybookRecommender {
	return &PlaybookRecommender{
		playbookEngine: playbookEngine,
	}
}

// RecommendPlaybooks 推荐剧本
func (pr *PlaybookRecommender) RecommendPlaybooks(anomaly AnomalyResult) ([]Playbook, error) {
	filter := PlaybookFilter{}

	// 根据异常类型推荐剧本
	if anomaly.Deviation > 50 {
		filter.Trigger = "high_deviation"
	}

	playbooks, err := pr.playbookEngine.ListPlaybooks(filter)
	if err != nil {
		return nil, err
	}

	return playbooks, nil
}

// ExecutePlaybookWithContext 执行剧本并返回结果
func (pr *PlaybookRecommender) ExecutePlaybookWithContext(ctx context.Context, playbookID string, context map[string]interface{}) (map[string]interface{}, error) {
	result, err := pr.playbookEngine.ExecutePlaybook(ctx, playbookID, context)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"execution_id": result.ExecutionID,
		"playbook_id":  result.PlaybookID,
		"status":       result.Status,
		"steps":        result.Steps,
		"start_time":   result.StartTime,
		"end_time":     result.EndTime,
	}, nil
}

// DashboardBuilder 看板构建器
type DashboardBuilder struct {
	dashboardGenerator *DashboardGenerator
	contextAnalyzer    *DashboardContextAnalyzer
}

// NewDashboardBuilder 创建看板构建器
func NewDashboardBuilder(dashboardGenerator *DashboardGenerator, contextAnalyzer *DashboardContextAnalyzer) *DashboardBuilder {
	return &DashboardBuilder{
		dashboardGenerator: dashboardGenerator,
		contextAnalyzer:    contextAnalyzer,
	}
}

// BuildDashboardFromContext 从上下文构建看板
func (db *DashboardBuilder) BuildDashboardFromContext(ctx context.Context, userID string, conversationContext map[string]interface{}) (Dashboard, error) {
	// 分析上下文获取推荐面板
	panels, err := db.contextAnalyzer.AnalyzeContext(ctx, conversationContext)
	if err != nil {
		return Dashboard{}, err
	}

	// 创建看板配置
	config := DashboardConfig{
		ID:        fmt.Sprintf("dash-%d", time.Now().UnixNano()),
		UserID:    userID,
		Name:      fmt.Sprintf("Dynamic Dashboard - %s", time.Now().Format("2006-01-02 15:04:05")),
		Panels:    panels,
		Refresh:   "30s",
		TimeRange: "1h",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 生成看板
	dashboard, err := db.dashboardGenerator.GenerateDashboard(userID, config)
	if err != nil {
		return Dashboard{}, err
	}

	return dashboard, nil
}

// MetricsAggregator 指标聚合器
type MetricsAggregator struct {
	client *PrometheusClient
}

// NewMetricsAggregator 创建指标聚合器
func NewMetricsAggregator(client *PrometheusClient) *MetricsAggregator {
	return &MetricsAggregator{
		client: client,
	}
}

// AggregateMetrics 聚合指标
func (ma *MetricsAggregator) AggregateMetrics(ctx context.Context, queries []string, start, end time.Time) (map[string]interface{}, error) {
	aggregated := map[string]interface{}{
		"queries": make([]map[string]interface{}, 0),
		"summary": map[string]interface{}{},
	}

	step := (end.Sub(start)) / 100
	if step < time.Minute {
		step = time.Minute
	}

	for _, query := range queries {
		matrix, err := ma.client.QueryRange(ctx, query, start, end, step)
		if err != nil {
			continue
		}

		queryResult := map[string]interface{}{
			"query":  query,
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

			queryResult["series"] = append(queryResult["series"].([]map[string]interface{}), seriesData)
		}

		aggregated["queries"] = append(aggregated["queries"].([]map[string]interface{}), queryResult)
	}

	return aggregated, nil
}
