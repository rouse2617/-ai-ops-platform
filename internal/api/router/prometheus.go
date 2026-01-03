package router

import (
	"ai-ops/internal/api/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPrometheusRoutes 注册 Prometheus 路由
func RegisterPrometheusRoutes(router *gin.Engine, promHandler *handler.PrometheusHandler) {
	group := router.Group("/api/prometheus")

	// 查询相关
	group.POST("/query", promHandler.QueryPrometheus)
	group.GET("/metrics", promHandler.GetMetricsMetadata)
	group.GET("/labels/:label", promHandler.GetLabelValues)

	// 语义翻译相关
	group.POST("/translate", promHandler.TranslateMetric)
	group.POST("/translate-alert", promHandler.TranslateAlert)

	// PromQL 生成相关
	group.POST("/generate-promql", promHandler.GeneratePromQL)
	group.POST("/nl-to-promql", promHandler.GeneratePromQLFromNL)

	// 查询结果可视化
	group.POST("/visualize", promHandler.VisualizeQueryResult)

	// 异常检测相关
	group.POST("/detect-anomaly", promHandler.DetectAnomaly)
	group.POST("/learn-baseline", promHandler.LearnBaseline)

	// 剧本相关
	group.POST("/playbooks", promHandler.CreatePlaybook)
	group.GET("/playbooks", promHandler.ListPlaybooks)
	group.GET("/playbooks/:id", promHandler.GetPlaybook)
	group.POST("/playbooks/:id/execute", promHandler.ExecutePlaybook)

	// 看板相关
	group.POST("/dashboards", promHandler.GenerateDashboard)
	group.GET("/dashboards", promHandler.ListDashboards)
	group.GET("/dashboards/:id", promHandler.GetDashboard)
	group.PUT("/dashboards/:id", promHandler.UpdateDashboard)
	group.DELETE("/dashboards/:id", promHandler.DeleteDashboard)
	group.GET("/dashboards/recommend-panels", promHandler.RecommendPanels)
	group.GET("/dashboards/templates", promHandler.ListDashboardTemplates)
	group.GET("/dashboards/template/:name", promHandler.GetDashboardTemplate)

	// 看板上下文分析
	group.POST("/analyze-context", promHandler.AnalyzeDashboardContext)

	// 故障复盘相关
	group.POST("/incident-replay", promHandler.QueryIncidentMetrics)

	// 容量规划相关
	group.POST("/capacity-planning", promHandler.AnalyzeCapacityPlanning)
}
