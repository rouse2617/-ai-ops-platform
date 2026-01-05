package model

import (
	"time"
)

// Analysis 分析历史模型
type Analysis struct {
	ID             string     `json:"id" gorm:"primaryKey"`
	SessionID      string     `json:"session_id" gorm:"index"`              // 会话ID（可选）
	ResultsData    string     `json:"results_data" gorm:"type:text"`        // 执行结果数据（JSON格式）
	AnalysisResult string     `json:"analysis_result" gorm:"type:text"`     // 分析结果
	Question       string     `json:"question" gorm:"type:text"`            // 用户提问（可选）
	ToolCalls      []ToolCall `json:"tool_calls" gorm:"serializer:json"`    // 工具调用记录
	CreatedAt      time.Time  `json:"created_at"`
}

// SemanticInsight 语义化解读结果
type SemanticInsight struct {
	Summary        string   `json:"summary"`                    // 简短摘要
	RiskLevel      string   `json:"risk_level"`                 // normal | warning | critical
	Trend          string   `json:"trend,omitempty"`            // 趋势分析
	Recommendation string   `json:"recommendation,omitempty"`   // 建议操作
	KeyMetrics     []Metric `json:"key_metrics,omitempty"`      // 关键指标
}

// Metric 关键指标
type Metric struct {
	Name      string  `json:"name"`
	Value     string  `json:"value"`
	Status    string  `json:"status"`    // normal | warning | critical
	Threshold string  `json:"threshold,omitempty"`
}

// ContextualAction 上下文相关的推荐操作
type ContextualAction struct {
	ID            string  `json:"id"`
	Label         string  `json:"label"`
	Icon          string  `json:"icon"`
	Command       string  `json:"command,omitempty"`
	Highlight     bool    `json:"highlight"`
	AIRecommended bool    `json:"ai_recommended"`
	Confidence    float64 `json:"confidence,omitempty"`
	RiskLevel     string  `json:"risk_level"`
	Description   string  `json:"description,omitempty"`
}

// HistoricalCorrelation 历史关联分析
type HistoricalCorrelation struct {
	HostID          string `json:"host_id"`
	CurrentIssue    string `json:"current_issue"`
	SimilarIncident string `json:"similar_incident,omitempty"`
	OccurredAt      string `json:"occurred_at,omitempty"`
	Resolution      string `json:"resolution,omitempty"`
	Confidence      float64 `json:"confidence"`
}













