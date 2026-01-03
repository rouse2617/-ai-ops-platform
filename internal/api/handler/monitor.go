package handler

import (
	"net/http"
	"time"

	"ai-ops/internal/monitor"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// MonitorHandler 监控处理器
type MonitorHandler struct {
	detector  *monitor.AnomalyDetector
	notifier  *monitor.AlertNotifier
	scheduler *monitor.MonitorScheduler
}

// NewMonitorHandler 创建监控处理器
func NewMonitorHandler(detector *monitor.AnomalyDetector, notifier *monitor.AlertNotifier, scheduler *monitor.MonitorScheduler) *MonitorHandler {
	return &MonitorHandler{
		detector:  detector,
		notifier:  notifier,
		scheduler: scheduler,
	}
}

// WebSocketHandler WebSocket 连接处理
func (h *MonitorHandler) WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	h.notifier.RegisterClient(clientID, conn)
	defer h.notifier.UnregisterClient(clientID)

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// AddAlertRule 添加告警规则
func (h *MonitorHandler) AddAlertRule(c *gin.Context) {
	var rule monitor.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	rule.ID = uuid.New().String()
	h.detector.AddRule(&rule)

	Success(c, rule)
}

// GetAlertRules 获取告警规则列表
func (h *MonitorHandler) GetAlertRules(c *gin.Context) {
	// 这里应该从数据库获取
	Success(c, []monitor.AlertRule{})
}

// UpdateMonitorHosts 更新监控主机列表
func (h *MonitorHandler) UpdateMonitorHosts(c *gin.Context) {
	var req struct {
		Hosts []string `json:"hosts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	h.scheduler.SetHosts(req.Hosts)
	Success(c, gin.H{"message": "监控主机列表已更新"})
}

// TriggerCheck 手动触发检查
func (h *MonitorHandler) TriggerCheck(c *gin.Context) {
	var req struct {
		Hosts []string `json:"hosts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	alerts, err := h.detector.DetectAnomalies(c.Request.Context(), req.Hosts)
	if err != nil {
		ExecError(c, "检测失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
		"time":   time.Now(),
	})
}
