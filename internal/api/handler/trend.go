package handler

import (
	"ai-ops/internal/service"

	"github.com/gin-gonic/gin"
)

// TrendHandler 趋势分析处理器
type TrendHandler struct {
	trendService *service.TrendService
}

// NewTrendHandler 创建趋势分析处理器
func NewTrendHandler(trendService *service.TrendService) *TrendHandler {
	return &TrendHandler{
		trendService: trendService,
	}
}

// AnalyzeTrends 分析趋势
// POST /api/trends/analyze
func (h *TrendHandler) AnalyzeTrends(c *gin.Context) {
	var req struct {
		HostID string `json:"host_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	results, err := h.trendService.AnalyzeTrends(c.Request.Context(), req.HostID)
	if err != nil {
		InternalError(c, "趋势分析失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"predictions": results,
	})
}

// GetPredictions 获取预测历史
// GET /api/trends/predictions
func (h *TrendHandler) GetPredictions(c *gin.Context) {
	hostID := c.Query("host_id")
	if hostID == "" {
		ParamError(c, "host_id 不能为空")
		return
	}

	limit := 50
	predictions, err := h.trendService.GetPredictions(hostID, limit)
	if err != nil {
		InternalError(c, "获取预测历史失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"predictions": predictions,
	})
}

// GetAlerts 获取告警
// GET /api/trends/alerts
func (h *TrendHandler) GetAlerts(c *gin.Context) {
	hostID := c.Query("host_id")
	if hostID == "" {
		ParamError(c, "host_id 不能为空")
		return
	}

	alerts, err := h.trendService.GetAlerts(hostID)
	if err != nil {
		InternalError(c, "获取告警失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"alerts": alerts,
	})
}
