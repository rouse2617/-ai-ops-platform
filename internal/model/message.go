package model

import (
	"time"
)

// Message 对话消息模型
type Message struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	SessionID string     `json:"session_id" gorm:"index;not null"`
	HostID    string     `json:"host_id" gorm:"index"`
	Role      string     `json:"role"`
	Content   string     `json:"content" gorm:"type:text"`
	ToolCalls []ToolCall `json:"tool_calls" gorm:"serializer:json"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToolCall 工具调用记录
type ToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
	Result interface{}            `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

// MessageRole 消息角色常量
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
	RoleSystem    = "system"
	RoleProactive = "proactive" // AI 主动发起的消息
)

// ProactiveMessageType 主动消息类型
const (
	ProactiveMorningReport  = "morning_report"  // 早安报告
	ProactiveAnomalyAlert   = "anomaly_alert"   // 异常告警
	ProactiveHealthSummary  = "health_summary"  // 健康摘要
	ProactiveAutoFix        = "auto_fix"        // 自动修复通知
)

// ProactiveMessage AI 主动发起的消息
type ProactiveMessage struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id,omitempty"`
	Type          string                 `json:"type"`           // morning_report, anomaly_alert, health_summary, auto_fix
	Title         string                 `json:"title"`
	Content       string                 `json:"content"`
	Summary       string                 `json:"summary,omitempty"` // 简短摘要
	Hosts         []string               `json:"hosts,omitempty"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
	Actions       []QuickAction          `json:"actions,omitempty"` // 快捷操作
	Priority      string                 `json:"priority"`          // low, medium, high, critical
	CreatedAt     int64                  `json:"created_at"`
	Read          bool                   `json:"read"`
}

// QuickAction 快捷操作
type QuickAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	Dangerous   bool   `json:"dangerous"`
}

