package builtin

import (
	"context"
	"fmt"
	"time"

	"ai-ops/internal/prometheus"
	"ai-ops/internal/tool"
)

// PrometheusIncidentReplayTool Prometheus 故障复盘工具
type PrometheusIncidentReplayTool struct {
	promClient *prometheus.PrometheusClient
}

// NewPrometheusIncidentReplayTool 创建故障复盘工具
func NewPrometheusIncidentReplayTool(promAddr string) (*PrometheusIncidentReplayTool, error) {
	client, err := prometheus.NewPrometheusClient(promAddr)
	if err != nil {
		return nil, err
	}
	return &PrometheusIncidentReplayTool{promClient: client}, nil
}

func (t *PrometheusIncidentReplayTool) Name() string {
	return "prometheus_incident_replay"
}

func (t *PrometheusIncidentReplayTool) Description() string {
	return `# Prometheus 故障复盘时光机

## 功能
通过 Prometheus 查询指定时间范围的历史指标数据，实现故障复盘分析。

## 参数
- incident_time: 故障发生时间 (RFC3339 格式)
- host: 目标主机名
- lookback_minutes: 回溯时间（分钟，默认15）

## 返回结果
- 故障时间线数据
- 关键指标变化趋势
- 异常检测结果
- 根因分析建议

适用场景：故障分析、性能调优、容量规划。`
}

func (t *PrometheusIncidentReplayTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "incident_time",
			Type:        "string",
			Description: "故障发生时间 (RFC3339 格式，如 2026-01-03T10:30:00Z)",
			Required:    true,
		},
		{
			Name:        "host",
			Type:        "string",
			Description: "目标主机名或 IP",
			Required:    true,
		},
		{
			Name:        "lookback_minutes",
			Type:        "integer",
			Description: "回溯时间（分钟，默认15）",
			Required:    false,
		},
	}
}

func (t *PrometheusIncidentReplayTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	incidentTimeStr := tool.GetStringParam(params, "incident_time", "")
	host := tool.GetStringParam(params, "host", "")
	_ = tool.GetIntParam(params, "lookback_minutes", 15) // 保留参数定义但不使用

	if incidentTimeStr == "" || host == "" {
		return tool.NewErrorResult("参数 incident_time 和 host 不能为空"), nil
	}

	incidentTime, err := time.Parse(time.RFC3339, incidentTimeStr)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("时间格式错误: %v", err)), nil
	}

	// 查询故障时间段的指标
	irq := prometheus.NewIncidentReplayQuery(t.promClient)
	timeline, err := irq.QueryIncidentMetrics(context.Background(), incidentTime, host)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("查询失败: %v", err)), nil
	}

	// 生成故障报告
	report := irq.GenerateIncidentReport(timeline)

	result := map[string]interface{}{
		"incident_time":     incidentTime.Format(time.RFC3339),
		"host":              host,
		"affected_metrics":  report.AffectedMetrics,
		"root_causes":       report.RootCauses,
		"recommendations":   report.Recommendations,
		"metrics_timeline":  t.formatTimeline(timeline),
	}

	return tool.NewResult(result, fmt.Sprintf("故障复盘完成，发现 %d 项异常", len(report.AffectedMetrics))), nil
}

func (t *PrometheusIncidentReplayTool) formatTimeline(timeline *prometheus.IncidentTimeline) map[string]interface{} {
	metrics := make(map[string]interface{})
	for name, ts := range timeline.Metrics {
		metrics[name] = map[string]interface{}{
			"baseline": ts.Baseline,
			"peak":     ts.Peak,
			"anomaly":  ts.Anomaly,
			"points":   len(ts.Values),
		}
	}
	return metrics
}

// PrometheusCapacityPlanningTool Prometheus 容量规划工具
type PrometheusCapacityPlanningTool struct {
	promClient *prometheus.PrometheusClient
}

// NewPrometheusCapacityPlanningTool 创建容量规划工具
func NewPrometheusCapacityPlanningTool(promAddr string) (*PrometheusCapacityPlanningTool, error) {
	client, err := prometheus.NewPrometheusClient(promAddr)
	if err != nil {
		return nil, err
	}
	return &PrometheusCapacityPlanningTool{promClient: client}, nil
}

func (t *PrometheusCapacityPlanningTool) Name() string {
	return "prometheus_capacity_planning"
}

func (t *PrometheusCapacityPlanningTool) Description() string {
	return `# Prometheus 容量规划 What-if 分析

## 功能
基于历史负载数据进行趋势分析和容量预测。

## 参数
- metric_query: PromQL 查询语句
- days: 历史数据天数（默认30）
- capacity_threshold: 容量阈值（百分比，默认90）

## 返回结果
- 负载趋势分析
- 容量耗尽预测
- What-if 场景分析
- 扩容建议

适用场景：容量规划、成本优化、资源预测。`
}

func (t *PrometheusCapacityPlanningTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "metric_query",
			Type:        "string",
			Description: "PromQL 查询语句",
			Required:    true,
		},
		{
			Name:        "days",
			Type:        "integer",
			Description: "历史数据天数（默认30）",
			Required:    false,
		},
		{
			Name:        "capacity_threshold",
			Type:        "number",
			Description: "容量阈值（百分比，默认90）",
			Required:    false,
		},
	}
}

func (t *PrometheusCapacityPlanningTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	query := tool.GetStringParam(params, "metric_query", "")
	days := tool.GetIntParam(params, "days", 30)
	threshold := tool.GetFloatParam(params, "capacity_threshold", 90.0)

	if query == "" {
		return tool.NewErrorResult("参数 metric_query 不能为空"), nil
	}

	// 获取历史数据
	cp := prometheus.NewCapacityPlanner(t.promClient)
	points, err := cp.GetHistoricalData(context.Background(), query, days)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("查询失败: %v", err)), nil
	}

	if len(points) == 0 {
		return tool.NewErrorResult("未获取到数据"), nil
	}

	// 分析趋势
	trend := cp.AnalyzeTrend(points)
	trend.MetricName = query

	// 预测容量耗尽时间
	daysToCapacity := cp.PredictCapacityExhaustion(trend, threshold)

	// What-if 场景分析
	scenarios := cp.AnalyzeWhatIfScenarios(trend, threshold)

	result := map[string]interface{}{
		"trend": map[string]interface{}{
			"type":              string(trend.Trend),
			"growth_rate":       trend.GrowthRate,
			"projected_peak":    trend.ProjectedPeak,
			"days_to_capacity":  daysToCapacity,
		},
		"scenarios": t.formatScenarios(scenarios),
		"data_points": len(points),
	}

	return tool.NewResult(result, fmt.Sprintf("容量分析完成，预计 %d 天内达到容量上限", daysToCapacity)), nil
}

func (t *PrometheusCapacityPlanningTool) formatScenarios(scenarios []prometheus.WhatIfScenario) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, s := range scenarios {
		result = append(result, map[string]interface{}{
			"name":                s.Name,
			"growth_multiplier":   s.GrowthMultiplier,
			"peak_multiplier":     s.PeakMultiplier,
			"recommended_action":  s.RecommendedAction,
			"time_to_action_days": s.TimeToAction,
		})
	}
	return result
}

// RegisterPrometheusTools 注册 Prometheus 工具
func RegisterPrometheusTools(registry *tool.Registry, promAddr string) error {
	// 故障复盘工具
	replayTool, err := NewPrometheusIncidentReplayTool(promAddr)
	if err != nil {
		return fmt.Errorf("创建故障复盘工具失败: %w", err)
	}
	registry.Register(replayTool, "builtin")

	// 容量规划工具
	planningTool, err := NewPrometheusCapacityPlanningTool(promAddr)
	if err != nil {
		return fmt.Errorf("创建容量规划工具失败: %w", err)
	}
	registry.Register(planningTool, "builtin")

	return nil
}
