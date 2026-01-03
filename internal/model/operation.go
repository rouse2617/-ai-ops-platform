package model

import (
	"time"
)

// OperationConfirmation 操作确认模型
type OperationConfirmation struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	SessionID   string     `json:"session_id" gorm:"index;not null"`
	HostID      string     `json:"host_id" gorm:"index;not null"`
	Command     string     `json:"command" gorm:"type:text;not null"`
	RiskLevel   string     `json:"risk_level" gorm:"not null"`
	Status      string     `json:"status" gorm:"index;not null"`
	RequestedAt time.Time  `json:"requested_at" gorm:"not null"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	ConfirmedBy string     `json:"confirmed_by"`
	TimeoutAt   time.Time  `json:"timeout_at" gorm:"index;not null"`
	Reason      string     `json:"reason" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ConfirmationStatus 确认状态常量
const (
	ConfirmationStatusPending  = "pending"
	ConfirmationStatusApproved = "approved"
	ConfirmationStatusRejected = "rejected"
	ConfirmationStatusTimeout  = "timeout"
)

// RiskLevel 风险级别常量
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

// HealthCheck 健康检查模型
type HealthCheck struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	HostID      string                 `json:"host_id" gorm:"index;not null"`
	CheckType   string                 `json:"check_type" gorm:"not null"`
	Status      string                 `json:"status" gorm:"not null"`
	Metrics     map[string]interface{} `json:"metrics" gorm:"serializer:json"`
	Issues      []string               `json:"issues" gorm:"serializer:json"`
	Suggestions []string               `json:"suggestions" gorm:"serializer:json"`
	CheckedAt   time.Time              `json:"checked_at" gorm:"index;not null"`
	CreatedAt   time.Time              `json:"created_at"`
}

// HealthCheckType 健康检查类型常量
const (
	HealthCheckTypeSystem     = "system"
	HealthCheckTypeResource   = "resource"
	HealthCheckTypeService    = "service"
	HealthCheckTypeSecurity   = "security"
	HealthCheckTypeDaily      = "daily"
)

// HealthCheckStatus 健康检查状态常量
const (
	HealthCheckStatusHealthy  = "healthy"
	HealthCheckStatusWarning  = "warning"
	HealthCheckStatusCritical = "critical"
	HealthCheckStatusUnknown  = "unknown"
)

// TrendPrediction 趋势预测模型
type TrendPrediction struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	HostID         string    `json:"host_id" gorm:"index;not null"`
	MetricType     string    `json:"metric_type" gorm:"not null"`
	CurrentValue   float64   `json:"current_value" gorm:"not null"`
	PredictedValue float64   `json:"predicted_value" gorm:"not null"`
	PredictionTime time.Time `json:"prediction_time" gorm:"index;not null"`
	Confidence     float64   `json:"confidence" gorm:"not null"`
	AlertLevel     string    `json:"alert_level"`
	CreatedAt      time.Time `json:"created_at"`
}

// MetricType 指标类型常量
const (
	MetricTypeCPU     = "cpu"
	MetricTypeMemory  = "memory"
	MetricTypeDisk    = "disk"
	MetricTypeNetwork = "network"
	MetricTypeIO      = "io"
)

// PrometheusAlert Prometheus 告警模型
type PrometheusAlert struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	AlertName   string                 `json:"alert_name" gorm:"not null"`
	HostID      string                 `json:"host_id" gorm:"index"`
	Labels      map[string]string      `json:"labels" gorm:"serializer:json"`
	Annotations map[string]string      `json:"annotations" gorm:"serializer:json"`
	Status      string                 `json:"status" gorm:"index;not null"`
	StartsAt    time.Time              `json:"starts_at" gorm:"index;not null"`
	EndsAt      *time.Time             `json:"ends_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// PrometheusAlertStatus Prometheus 告警状态常量
const (
	PrometheusAlertStatusFiring   = "firing"
	PrometheusAlertStatusResolved = "resolved"
)
