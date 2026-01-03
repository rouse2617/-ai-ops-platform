package handler

import (
	"ai-ops/internal/service"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// CheckHost 检查主机健康状态
// POST /api/health/check
func (h *HealthHandler) CheckHost(c *gin.Context) {
	var req struct {
		HostID string `json:"host_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.healthService.CheckHost(c.Request.Context(), req.HostID)
	if err != nil {
		InternalError(c, "健康检查失败: "+err.Error())
		return
	}

	Success(c, result)
}

// DailyHealthCheck 每日健康体检
// POST /api/health/daily-report
func (h *HealthHandler) DailyHealthCheck(c *gin.Context) {
	results, err := h.healthService.DailyHealthCheck(c.Request.Context())
	if err != nil {
		InternalError(c, "每日体检失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"results": results,
		"summary": h.generateSummary(results),
	})
}

// GetHealthHistory 获取健康检查历史
// GET /api/health/reports
func (h *HealthHandler) GetHealthHistory(c *gin.Context) {
	hostID := c.Query("host_id")
	if hostID == "" {
		ParamError(c, "host_id 不能为空")
		return
	}

	limit := 50
	history, err := h.healthService.GetHealthHistory(hostID, limit)
	if err != nil {
		InternalError(c, "获取历史记录失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"history": history,
	})
}

// generateSummary 生成摘要
func (h *HealthHandler) generateSummary(results map[string]*service.HealthCheckResult) map[string]interface{} {
	summary := map[string]interface{}{
		"total":    len(results),
		"healthy":  0,
		"warning":  0,
		"critical": 0,
	}

	for _, result := range results {
		switch result.Status {
		case "healthy":
			summary["healthy"] = summary["healthy"].(int) + 1
		case "warning":
			summary["warning"] = summary["warning"].(int) + 1
		case "critical":
			summary["critical"] = summary["critical"].(int) + 1
		}
	}

	return summary
}
