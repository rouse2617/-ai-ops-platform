package monitor

import (
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelHigh     AlertLevel = "high"
	AlertLevelMedium   AlertLevel = "medium"
	AlertLevelLow      AlertLevel = "low"
)

// AlertType 告警类型
type AlertType string

const (
	AlertTypeCPU        AlertType = "cpu"
	AlertTypeMemory     AlertType = "memory"
	AlertTypeDisk       AlertType = "disk"
	AlertTypeNetwork    AlertType = "network"
	AlertTypeProcess    AlertType = "process"
	AlertTypeService    AlertType = "service"
	AlertTypePredictive AlertType = "predictive"
)

// Alert 告警模型
type Alert struct {
	ID          string                 `json:"id"`
	HostID      string                 `json:"host_id"`
	HostName    string                 `json:"host_name"`
	Type        AlertType              `json:"type"`
	Level       AlertLevel             `json:"level"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Metrics     map[string]interface{} `json:"metrics"`
	Suggestions []string               `json:"suggestions"`
	Actions     []QuickAction          `json:"actions"`
	CreatedAt   time.Time              `json:"created_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Status      string                 `json:"status"` // active, resolved, ignored
	// 用于代码变更关联分析
	Name      string            `json:"name,omitempty"`
	Service   string            `json:"service,omitempty"`
	Severity  string            `json:"severity,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// QuickAction 快捷操作
type QuickAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Dangerous   bool   `json:"dangerous"`
}

// HostMetrics 主机指标
type HostMetrics struct {
	HostID    string    `json:"host_id"`
	HostName  string    `json:"host_name"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Disk      float64   `json:"disk"`
	Load      []float64 `json:"load"`
	IOWait    float64   `json:"io_wait"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      AlertType              `json:"type"`
	Level     AlertLevel             `json:"level"`
	Condition map[string]interface{} `json:"condition"`
	Enabled   bool                   `json:"enabled"`
	Hosts     []string               `json:"hosts"` // 空表示所有主机
}
