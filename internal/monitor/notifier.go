package monitor

import (
	"encoding/json"
	"sync"
	"time"

	"ai-pro/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// AlertNotifier 告警通知器
type AlertNotifier struct {
	clients map[string]*websocket.Conn
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewAlertNotifier 创建告警通知器
func NewAlertNotifier(logger *zap.Logger) *AlertNotifier {
	return &AlertNotifier{
		clients: make(map[string]*websocket.Conn),
		logger:  logger,
	}
}

// RegisterClient 注册客户端
func (n *AlertNotifier) RegisterClient(clientID string, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.clients[clientID] = conn
	n.logger.Info("客户端注册", zap.String("client_id", clientID))
}

// UnregisterClient 注销客户端
func (n *AlertNotifier) UnregisterClient(clientID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.clients, clientID)
	n.logger.Info("客户端注销", zap.String("client_id", clientID))
}

// Notify 推送告警
func (n *AlertNotifier) Notify(alert *Alert) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	message := map[string]interface{}{
		"type": "alert",
		"data": alert,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 广播给所有客户端
	for clientID, conn := range n.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			n.logger.Error("推送告警失败",
				zap.String("client_id", clientID),
				zap.Error(err))
		}
	}

	return nil
}

// BroadcastMetrics 广播指标
func (n *AlertNotifier) BroadcastMetrics(metrics *HostMetrics) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	message := map[string]interface{}{
		"type": "metrics",
		"data": metrics,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for clientID, conn := range n.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			n.logger.Error("推送指标失败",
				zap.String("client_id", clientID),
				zap.Error(err))
		}
	}

	return nil
}

// NotifyCodeCorrelation 推送代码关联分析结果
func (n *AlertNotifier) NotifyCodeCorrelation(correlation interface{}) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	message := map[string]interface{}{
		"type": "code_correlation",
		"data": correlation,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for clientID, conn := range n.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			n.logger.Error("推送代码关联失败",
				zap.String("client_id", clientID),
				zap.Error(err))
		}
	}

	return nil
}

// NotifyProactiveMessage 推送 AI 主动消息到聊天流
func (n *AlertNotifier) NotifyProactiveMessage(msg *model.ProactiveMessage) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}

	message := map[string]interface{}{
		"type": "proactive_message",
		"data": msg,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for clientID, conn := range n.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			n.logger.Error("推送主动消息失败",
				zap.String("client_id", clientID),
				zap.Error(err))
		}
	}

	n.logger.Info("主动消息已推送",
		zap.String("type", msg.Type),
		zap.String("title", msg.Title))

	return nil
}

// SendMorningReport 发送早安报告
func (n *AlertNotifier) SendMorningReport(summary string, hosts []string, metrics map[string]interface{}) error {
	msg := &model.ProactiveMessage{
		Type:     model.ProactiveMorningReport,
		Title:    "早安报告",
		Content:  summary,
		Hosts:    hosts,
		Metrics:  metrics,
		Priority: "low",
	}
	return n.NotifyProactiveMessage(msg)
}

// SendAnomalyAlert 发送异常告警到聊天流
func (n *AlertNotifier) SendAnomalyAlert(title, content string, hosts []string, actions []model.QuickAction) error {
	msg := &model.ProactiveMessage{
		Type:     model.ProactiveAnomalyAlert,
		Title:    title,
		Content:  content,
		Hosts:    hosts,
		Actions:  actions,
		Priority: "high",
	}
	return n.NotifyProactiveMessage(msg)
}

// SendHealthSummary 发送健康摘要
func (n *AlertNotifier) SendHealthSummary(summary string, metrics map[string]interface{}) error {
	msg := &model.ProactiveMessage{
		Type:     model.ProactiveHealthSummary,
		Title:    "系统健康摘要",
		Content:  summary,
		Metrics:  metrics,
		Priority: "medium",
	}
	return n.NotifyProactiveMessage(msg)
}

// SendAutoFixNotification 发送自动修复通知
func (n *AlertNotifier) SendAutoFixNotification(title, content string, host string) error {
	msg := &model.ProactiveMessage{
		Type:     model.ProactiveAutoFix,
		Title:    title,
		Content:  content,
		Hosts:    []string{host},
		Priority: "medium",
	}
	return n.NotifyProactiveMessage(msg)
}
