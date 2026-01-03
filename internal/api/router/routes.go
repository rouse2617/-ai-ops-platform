package router

import (
	"ai-ops/internal/api/handler"
	"ai-ops/internal/operation"
	"ai-ops/internal/repository"
	"ai-ops/internal/service"
	"ai-ops/internal/ssh"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, db *gorm.DB, sshPool *ssh.Pool, promAddr string) {
	// 初始化仓库
	hostRepo := repository.NewHostRepository(db)
	operationRepo := repository.NewOperationRepository(db)
	healthRepo := repository.NewHealthCheckRepository(db)
	trendRepo := repository.NewTrendPredictionRepository(db)

	// 初始化核心组件
	confirmMgr := operation.NewConfirmationManager(operationRepo, nil)
	interceptor := operation.NewDangerInterceptor(confirmMgr)

	// 初始化服务
	operationService := service.NewOperationService(interceptor, confirmMgr, sshPool, hostRepo)
	healthService := service.NewHealthService(sshPool, hostRepo, healthRepo)
	trendService := service.NewTrendService(sshPool, hostRepo, trendRepo, healthRepo)

	// 初始化处理器
	operationHandler := handler.NewOperationHandler(operationService, confirmMgr)
	healthHandler := handler.NewHealthHandler(healthService)
	trendHandler := handler.NewTrendHandler(trendService)

	var prometheusHandler *handler.PrometheusHandler
	if promAddr != "" {
		var err error
		prometheusHandler, err = handler.NewPrometheusHandler(promAddr)
		if err != nil {
			// Log error but continue without Prometheus
			_ = err
		}
	}

	// API 路由组
	api := r.Group("/api")
	{
		// 操作确认相关路由
		operations := api.Group("/operations")
		{
			operations.GET("/confirmations/:id", operationHandler.GetConfirmation)
			operations.POST("/confirmations/:id/approve", operationHandler.ApproveConfirmation)
			operations.POST("/confirmations/:id/reject", operationHandler.RejectConfirmation)
			operations.GET("/confirmations", operationHandler.ListConfirmations)
			operations.GET("/confirmations/pending", operationHandler.GetPendingConfirmations)
		}

		// 健康检查相关路由
		health := api.Group("/health")
		{
			health.POST("/check", healthHandler.CheckHost)
			health.POST("/daily-report", healthHandler.DailyHealthCheck)
			health.GET("/reports", healthHandler.GetHealthHistory)
		}

		// 趋势分析相关路由
		trends := api.Group("/trends")
		{
			trends.POST("/analyze", trendHandler.AnalyzeTrends)
			trends.GET("/predictions", trendHandler.GetPredictions)
			trends.GET("/alerts", trendHandler.GetAlerts)
		}

		// Prometheus 相关路由
		if prometheusHandler != nil {
			prom := api.Group("/prometheus")
			{
				prom.POST("/query", prometheusHandler.QueryPrometheus)
				prom.GET("/metrics/metadata", prometheusHandler.GetMetricsMetadata)
				prom.GET("/labels/:label/values", prometheusHandler.GetLabelValues)
				prom.POST("/incident/metrics", prometheusHandler.QueryIncidentMetrics)
				prom.POST("/capacity/analyze", prometheusHandler.AnalyzeCapacityPlanning)
				prom.POST("/translate", prometheusHandler.TranslateMetric)
				prom.POST("/promql/generate", prometheusHandler.GeneratePromQL)
				prom.POST("/anomaly/detect", prometheusHandler.DetectAnomaly)
			}
		}
	}
}
