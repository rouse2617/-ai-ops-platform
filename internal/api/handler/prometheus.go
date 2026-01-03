package handler

import (
	"time"

	"ai-ops/internal/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PrometheusHandler Prometheus 处理器
type PrometheusHandler struct {
	promClient          *prometheus.PrometheusClient
	incidentReplay      *prometheus.IncidentReplayQuery
	capacityPlanner     *prometheus.CapacityPlanner
	semanticEngine      *prometheus.SemanticEngine
	promqlGenerator     *prometheus.PromQLGenerator
	thresholdDetector   *prometheus.ThresholdDetector
	playbookEngine      *prometheus.PlaybookEngine
	dashboardGenerator  *prometheus.DashboardGenerator
}

// NewPrometheusHandler 创建 Prometheus 处理器
func NewPrometheusHandler(promAddr string) (*PrometheusHandler, error) {
	client, err := prometheus.NewPrometheusClient(promAddr)
	if err != nil {
		return nil, err
	}

	return &PrometheusHandler{
		promClient:         client,
		incidentReplay:     prometheus.NewIncidentReplayQuery(client),
		capacityPlanner:    prometheus.NewCapacityPlanner(client),
		semanticEngine:     prometheus.NewSemanticEngine(),
		promqlGenerator:    prometheus.NewPromQLGenerator(),
		thresholdDetector:  prometheus.NewThresholdDetector(1000),
		playbookEngine:     prometheus.NewPlaybookEngine(),
		dashboardGenerator: prometheus.NewDashboardGenerator(),
	}, nil
}

// IncidentReplayRequest 故障复盘请求
type IncidentReplayRequest struct {
	IncidentTime    string `json:"incident_time" binding:"required"` // RFC3339 格式
	Host            string `json:"host" binding:"required"`
	LookbackMinutes int    `json:"lookback_minutes"`
}

// IncidentReplayResponse 故障复盘响应
type IncidentReplayResponse struct {
	IncidentTime    string                 `json:"incident_time"`
	Host            string                 `json:"host"`
	AffectedMetrics []string               `json:"affected_metrics"`
	RootCauses      []string               `json:"root_causes"`
	Recommendations []string               `json:"recommendations"`
	MetricsTimeline map[string]interface{} `json:"metrics_timeline"`
	Timeline        *prometheus.IncidentTimeline `json:"timeline,omitempty"`
}

// QueryIncidentMetrics 查询故障时间段的指标
// POST /api/prometheus/incident-replay
func (h *PrometheusHandler) QueryIncidentMetrics(c *gin.Context) {
	var req IncidentReplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	incidentTime, err := time.Parse(time.RFC3339, req.IncidentTime)
	if err != nil {
		ParamError(c, "时间格式错误: "+err.Error())
		return
	}

	if req.LookbackMinutes == 0 {
		req.LookbackMinutes = 15
	}

	// 查询故障时间段的指标
	timeline, err := h.incidentReplay.QueryIncidentMetrics(c.Request.Context(), incidentTime, req.Host)
	if err != nil {
		ExecError(c, "查询失败: "+err.Error())
		return
	}

	// 生成故障报告
	report := h.incidentReplay.GenerateIncidentReport(timeline)

	resp := IncidentReplayResponse{
		IncidentTime:    req.IncidentTime,
		Host:            req.Host,
		AffectedMetrics: report.AffectedMetrics,
		RootCauses:      report.RootCauses,
		Recommendations: report.Recommendations,
		Timeline:        timeline,
	}

	Success(c, resp)
}

// CapacityPlanningRequest 容量规划请求
type CapacityPlanningRequest struct {
	MetricQuery        string  `json:"metric_query" binding:"required"`
	Days               int     `json:"days"`
	CapacityThreshold  float64 `json:"capacity_threshold"`
	CurrentCapacity    float64 `json:"current_capacity"`
	CostPerUnit        float64 `json:"cost_per_unit"`
}

// CapacityPlanningResponse 容量规划响应
type CapacityPlanningResponse struct {
	Trend                 map[string]interface{}   `json:"trend"`
	Scenarios             []map[string]interface{} `json:"scenarios"`
	Recommendation        map[string]interface{}   `json:"recommendation"`
	ResourceAllocation    map[string]interface{}   `json:"resource_allocation"`
	DataPoints            int                      `json:"data_points"`
}

// AnalyzeCapacityPlanning 分析容量规划
// POST /api/prometheus/capacity-planning
func (h *PrometheusHandler) AnalyzeCapacityPlanning(c *gin.Context) {
	var req CapacityPlanningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if req.Days == 0 {
		req.Days = 30
	}
	if req.CapacityThreshold == 0 {
		req.CapacityThreshold = 90.0
	}
	if req.CurrentCapacity == 0 {
		req.CurrentCapacity = 100.0
	}

	// 获取历史数据
	points, err := h.capacityPlanner.GetHistoricalData(c.Request.Context(), req.MetricQuery, req.Days)
	if err != nil {
		ExecError(c, "查询失败: "+err.Error())
		return
	}

	if len(points) == 0 {
		ParamError(c, "未获取到数据")
		return
	}

	// 分析趋势
	trend := h.capacityPlanner.AnalyzeTrend(points)
	trend.MetricName = req.MetricQuery

	// 预测容量耗尽时间
	daysToCapacity := h.capacityPlanner.PredictCapacityExhaustion(trend, req.CapacityThreshold)

	// What-if 场景分析
	scenarios := h.capacityPlanner.AnalyzeWhatIfScenarios(trend, req.CapacityThreshold)

	// 生成容量建议
	recommendation := h.capacityPlanner.GenerateCapacityRecommendation(trend, req.CurrentCapacity, req.CapacityThreshold)

	// 计算资源分配
	allocation := h.capacityPlanner.CalculateResourceAllocation(trend, req.CurrentCapacity, req.CostPerUnit)

	resp := CapacityPlanningResponse{
		Trend: map[string]interface{}{
			"type":              string(trend.Trend),
			"growth_rate":       trend.GrowthRate,
			"projected_peak":    trend.ProjectedPeak,
			"days_to_capacity":  daysToCapacity,
		},
		Scenarios: h.formatScenarios(scenarios),
		Recommendation: map[string]interface{}{
			"current_usage":         recommendation.CurrentUsage,
			"projected_usage":       recommendation.ProjectedUsage,
			"recommended_capacity":  recommendation.RecommendedCapacity,
			"utilization_rate":      recommendation.UtilizationRate,
			"safety_margin":         recommendation.SafetyMargin,
			"priority":              recommendation.Priority,
			"action_items":          recommendation.ActionItems,
		},
		ResourceAllocation: map[string]interface{}{
			"current_allocation":     allocation.CurrentAllocation,
			"recommended_allocation": allocation.RecommendedAllocation,
			"allocation_ratio":       allocation.AllocationRatio,
			"timeline":               allocation.Timeline,
			"cost_estimate":          allocation.CostEstimate,
			"roi":                    allocation.ROI,
		},
		DataPoints: len(points),
	}

	Success(c, resp)
}

// PromQLQueryRequest PromQL 查询请求
type PromQLQueryRequest struct {
	Query string    `json:"query" binding:"required"`
	Start time.Time `json:"start" binding:"required"`
	End   time.Time `json:"end" binding:"required"`
	Step  string    `json:"step"` // 如 "30s", "1m", "5m"
}

// QueryPrometheus 执行 PromQL 查询
// POST /api/prometheus/query
func (h *PrometheusHandler) QueryPrometheus(c *gin.Context) {
	var req PromQLQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 解析 step
	step := 1 * time.Minute
	if req.Step != "" {
		parsedStep, err := time.ParseDuration(req.Step)
		if err != nil {
			ParamError(c, "step 格式错误: "+err.Error())
			return
		}
		step = parsedStep
	}

	// 执行查询
	result, err := h.promClient.QueryRange(c.Request.Context(), req.Query, req.Start, req.End, step)
	if err != nil {
		ExecError(c, "查询失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"query":  req.Query,
		"start":  req.Start,
		"end":    req.End,
		"step":   step.String(),
		"result": result,
	})
}

// GetMetricsMetadata 获取指标元数据
// GET /api/prometheus/metrics
func (h *PrometheusHandler) GetMetricsMetadata(c *gin.Context) {
	metadata, err := h.promClient.GetMetricMetadata(c.Request.Context())
	if err != nil {
		ExecError(c, "获取元数据失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"metrics": metadata,
	})
}

// GetLabelValues 获取标签值
// GET /api/prometheus/labels/:label
func (h *PrometheusHandler) GetLabelValues(c *gin.Context) {
	label := c.Param("label")
	if label == "" {
		ParamError(c, "标签名不能为空")
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	if startStr != "" {
		var err error
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			ParamError(c, "start 时间格式错误")
			return
		}
	} else {
		start = time.Now().Add(-24 * time.Hour)
	}

	if endStr != "" {
		var err error
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			ParamError(c, "end 时间格式错误")
			return
		}
	} else {
		end = time.Now()
	}

	values, warnings, err := h.promClient.GetLabelValues(c.Request.Context(), label, start, end)
	if err != nil {
		ExecError(c, "获取标签值失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"label":    label,
		"values":   values,
		"warnings": warnings,
	})
}

// TranslateMetric 翻译指标为自然语言
// POST /api/prometheus/translate
func (h *PrometheusHandler) TranslateMetric(c *gin.Context) {
	var req struct {
		MetricName string            `json:"metric_name" binding:"required"`
		Value      float64           `json:"value" binding:"required"`
		Labels     map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	metric := prometheus.MetricSample{
		Metric:    req.Labels,
		Value:     req.Value,
		Timestamp: time.Now().Unix(),
	}

	if metric.Metric == nil {
		metric.Metric = make(map[string]string)
	}
	metric.Metric["__name__"] = req.MetricName

	translation := h.semanticEngine.TranslateMetric(metric)
	Success(c, translation)
}

// GeneratePromQL 生成 PromQL
// POST /api/prometheus/generate-promql
func (h *PrometheusHandler) GeneratePromQL(c *gin.Context) {
	var req struct {
		Question string                 `json:"question" binding:"required"`
		Context  map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	nlQuery := prometheus.NLQuery{
		Question: req.Question,
		Context:  req.Context,
	}

	query, err := h.promqlGenerator.GenerateQuery(nlQuery)
	if err != nil {
		ExecError(c, "生成失败: "+err.Error())
		return
	}

	Success(c, query)
}

// DetectAnomaly 检测异常
// POST /api/prometheus/detect-anomaly
func (h *PrometheusHandler) DetectAnomaly(c *gin.Context) {
	var req struct {
		MetricName string  `json:"metric_name" binding:"required"`
		Value      float64 `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	baseline, err := h.thresholdDetector.GetBaseline(req.MetricName)
	if err != nil {
		ExecError(c, "获取基线失败: "+err.Error())
		return
	}

	metric := prometheus.MetricSample{
		Metric: map[string]string{"__name__": req.MetricName},
		Value:  req.Value,
	}

	result := h.thresholdDetector.DetectAnomaly(metric, *baseline)
	Success(c, result)
}

// CreatePlaybook 创建剧本
// POST /api/prometheus/playbooks
func (h *PrometheusHandler) CreatePlaybook(c *gin.Context) {
	var playbook prometheus.Playbook

	if err := c.ShouldBindJSON(&playbook); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if playbook.ID == "" {
		playbook.ID = uuid.New().String()
	}

	if err := h.playbookEngine.CreatePlaybook(playbook); err != nil {
		ExecError(c, "创建失败: "+err.Error())
		return
	}

	Success(c, playbook)
}

// ExecutePlaybook 执行剧本
// POST /api/prometheus/playbooks/:id/execute
func (h *PrometheusHandler) ExecutePlaybook(c *gin.Context) {
	playbookID := c.Param("id")

	var req struct {
		Context map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.playbookEngine.ExecutePlaybook(c.Request.Context(), playbookID, req.Context)
	if err != nil {
		ExecError(c, "执行失败: "+err.Error())
		return
	}

	Success(c, result)
}

// GetPlaybook 获取剧本
// GET /api/prometheus/playbooks/:id
func (h *PrometheusHandler) GetPlaybook(c *gin.Context) {
	playbookID := c.Param("id")

	playbook, err := h.playbookEngine.GetPlaybook(playbookID)
	if err != nil {
		ExecError(c, "获取失败: "+err.Error())
		return
	}

	Success(c, playbook)
}

// ListPlaybooks 列表剧本
// GET /api/prometheus/playbooks
func (h *PrometheusHandler) ListPlaybooks(c *gin.Context) {
	filter := prometheus.PlaybookFilter{
		Name:    c.Query("name"),
		Trigger: c.Query("trigger"),
	}

	playbooks, err := h.playbookEngine.ListPlaybooks(filter)
	if err != nil {
		ExecError(c, "列表失败: "+err.Error())
		return
	}

	Success(c, playbooks)
}

// GenerateDashboard 生成看板
// POST /api/prometheus/dashboards
func (h *PrometheusHandler) GenerateDashboard(c *gin.Context) {
	var config prometheus.DashboardConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default"
	}

	dashboard, err := h.dashboardGenerator.GenerateDashboard(userID, config)
	if err != nil {
		ExecError(c, "生成失败: "+err.Error())
		return
	}

	Success(c, dashboard)
}

// GetDashboard 获取看板
// GET /api/prometheus/dashboards/:id
func (h *PrometheusHandler) GetDashboard(c *gin.Context) {
	dashboardID := c.Param("id")

	dashboard, err := h.dashboardGenerator.GetDashboard(dashboardID)
	if err != nil {
		ExecError(c, "获取失败: "+err.Error())
		return
	}

	Success(c, dashboard)
}

// ListDashboards 列表看板
// GET /api/prometheus/dashboards
func (h *PrometheusHandler) ListDashboards(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default"
	}

	dashboards, err := h.dashboardGenerator.ListDashboards(userID)
	if err != nil {
		ExecError(c, "列表失败: "+err.Error())
		return
	}

	Success(c, dashboards)
}

// RecommendPanels 推荐面板
// GET /api/prometheus/dashboards/recommend-panels
func (h *PrometheusHandler) RecommendPanels(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default"
	}

	panels, err := h.dashboardGenerator.RecommendPanels(userID)
	if err != nil {
		ExecError(c, "推荐失败: "+err.Error())
		return
	}

	Success(c, panels)
}

// formatScenarios 格式化场景
func (h *PrometheusHandler) formatScenarios(scenarios []prometheus.WhatIfScenario) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, s := range scenarios {
		result = append(result, map[string]interface{}{
			"name":                s.Name,
			"growth_multiplier":   s.GrowthMultiplier,
			"peak_multiplier":     s.PeakMultiplier,
			"projected_metrics":   s.ProjectedMetrics,
			"recommended_action":  s.RecommendedAction,
			"time_to_action_days": s.TimeToAction,
		})
	}
	return result
}

// TranslateAlert 翻译告警
// POST /api/prometheus/translate-alert
func (h *PrometheusHandler) TranslateAlert(c *gin.Context) {
	var req struct {
		AlertName   string            `json:"alert_name" binding:"required"`
		Severity    string            `json:"severity"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	alert := prometheus.AlertMetadata{
		AlertName:   req.AlertName,
		Severity:    req.Severity,
		Description: req.Description,
	}

	context := prometheus.AlertContext{
		Metric: prometheus.MetricSample{
			Metric: req.Labels,
		},
	}

	translator := prometheus.NewAlertTranslator(h.semanticEngine)
	translation := translator.TranslateAlert(alert, context)

	Success(c, gin.H{
		"alert_name":  req.AlertName,
		"translation": translation,
	})
}

// GeneratePromQLFromNL 从自然语言生成 PromQL
// POST /api/prometheus/nl-to-promql
func (h *PrometheusHandler) GeneratePromQLFromNL(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	nlToPromQL := prometheus.NewNLToPromQL(h.promqlGenerator)
	query, err := nlToPromQL.Convert(req.Question)
	if err != nil {
		ExecError(c, "转换失败: "+err.Error())
		return
	}

	Success(c, query)
}

// VisualizeQueryResult 可视化查询结果
// POST /api/prometheus/visualize
func (h *PrometheusHandler) VisualizeQueryResult(c *gin.Context) {
	var req struct {
		Query string    `json:"query" binding:"required"`
		Start time.Time `json:"start" binding:"required"`
		End   time.Time `json:"end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	visualizer := prometheus.NewQueryResultVisualizer(h.promClient)
	visualization, err := visualizer.VisualizeQuery(c.Request.Context(), req.Query, req.Start, req.End)
	if err != nil {
		ExecError(c, "可视化失败: "+err.Error())
		return
	}

	Success(c, visualization)
}

// AnalyzeDashboardContext 分析看板上下文
// POST /api/prometheus/analyze-context
func (h *PrometheusHandler) AnalyzeDashboardContext(c *gin.Context) {
	var req struct {
		Context map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	analyzer := prometheus.NewDashboardContextAnalyzer(h.dashboardGenerator)
	panels, err := analyzer.AnalyzeContext(c.Request.Context(), req.Context)
	if err != nil {
		ExecError(c, "分析失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"recommended_panels": panels,
	})
}

// GetDashboardTemplate 获取看板模板
// GET /api/prometheus/dashboards/template/:name
func (h *PrometheusHandler) GetDashboardTemplate(c *gin.Context) {
	templateName := c.Param("name")
	if templateName == "" {
		ParamError(c, "模板名称不能为空")
		return
	}

	panels, err := h.dashboardGenerator.GetDashboardTemplate(templateName)
	if err != nil {
		ExecError(c, "获取模板失败: "+err.Error())
		return
	}

	Success(c, panels)
}

// ListDashboardTemplates 列表看板模板
// GET /api/prometheus/dashboards/templates
func (h *PrometheusHandler) ListDashboardTemplates(c *gin.Context) {
	templates := h.dashboardGenerator.ListTemplates()
	Success(c, gin.H{
		"templates": templates,
	})
}

// UpdateDashboard 更新看板
// PUT /api/prometheus/dashboards/:id
func (h *PrometheusHandler) UpdateDashboard(c *gin.Context) {
	dashboardID := c.Param("id")
	if dashboardID == "" {
		ParamError(c, "看板 ID 不能为空")
		return
	}

	var config prometheus.DashboardConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if err := h.dashboardGenerator.UpdateDashboard(dashboardID, config); err != nil {
		ExecError(c, "更新失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "更新成功",
		"id":      dashboardID,
	})
}

// DeleteDashboard 删除看板
// DELETE /api/prometheus/dashboards/:id
func (h *PrometheusHandler) DeleteDashboard(c *gin.Context) {
	dashboardID := c.Param("id")
	if dashboardID == "" {
		ParamError(c, "看板 ID 不能为空")
		return
	}

	if err := h.dashboardGenerator.DeleteDashboard(dashboardID); err != nil {
		ExecError(c, "删除失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "删除成功",
		"id":      dashboardID,
	})
}

// LearnBaseline 学习基线
// POST /api/prometheus/learn-baseline
func (h *PrometheusHandler) LearnBaseline(c *gin.Context) {
	var req struct {
		MetricName string    `json:"metric_name" binding:"required"`
		Start      time.Time `json:"start" binding:"required"`
		End        time.Time `json:"end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 查询历史数据
	step := (req.End.Sub(req.Start)) / 100
	if step < time.Minute {
		step = time.Minute
	}

	matrix, err := h.promClient.QueryRange(c.Request.Context(), req.MetricName, req.Start, req.End, step)
	if err != nil {
		ExecError(c, "查询失败: "+err.Error())
		return
	}

	// 转换为 MetricSample
	metrics := make([]prometheus.MetricSample, 0)
	for _, series := range matrix {
		for _, sample := range series.Values {
			metrics = append(metrics, prometheus.MetricSample{
				Metric:    map[string]string{"__name__": req.MetricName},
				Value:     float64(sample.Value),
				Timestamp: int64(sample.Timestamp),
			})
		}
	}

	if len(metrics) == 0 {
		ParamError(c, "未获取到数据")
		return
	}

	// 学习基线
	baseline, err := h.thresholdDetector.LearnBaseline(metrics)
	if err != nil {
		ExecError(c, "学习失败: "+err.Error())
		return
	}

	Success(c, baseline)
}
