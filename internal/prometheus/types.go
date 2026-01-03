package prometheus

import "time"

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelHigh     AlertLevel = "high"
	AlertLevelMedium   AlertLevel = "medium"
	AlertLevelLow      AlertLevel = "low"
)

// MetricSample Prometheus 指标样本
type MetricSample struct {
	Metric    map[string]string
	Value     float64
	Timestamp int64
}

// RangeSeries 范围查询结果序列
type RangeSeries struct {
	Metric map[string]string
	Values [][2]interface{}
}

// RangeQueryResult 范围查询结果
type RangeQueryResult struct {
	ResultType string
	Result     []RangeSeries
}

// InstantQueryResult 即时查询结果
type InstantQueryResult struct {
	ResultType string
	Result     []MetricSample
}

// BaselineMetrics 基线指标
type BaselineMetrics struct {
	MetricName string
	Mean       float64
	StdDev     float64
	P95        float64
	P99        float64
	UpdatedAt  time.Time
}

// AnomalyResult 异常检测结果
type AnomalyResult struct {
	IsAnomaly      bool
	Severity       float64 // 0-1
	Deviation      float64
	Recommendation string
}

// SemanticTranslation 语义翻译结果
type SemanticTranslation struct {
	Title       string
	Description string
	RootCause   string
	Impact      string
	Suggestions []string
}

// NLQuery 自然语言查询
type NLQuery struct {
	Question string
	Context  map[string]interface{}
}

// PromQLQuery PromQL 查询
type PromQLQuery struct {
	Query       string
	Range       string
	Step        string
	Explanation string
}

// AlertMetadata 告警元数据
type AlertMetadata struct {
	AlertName   string
	Severity    string
	Description string
	Runbook     string
	Dashboard   string
}

// AlertContext 告警上下文
type AlertContext struct {
	Metric    MetricSample
	Baseline  BaselineMetrics
	Trend     TrendAnalysis
	Related   []RelatedMetric
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	Direction  string  // up, down, stable
	Rate       float64
	Confidence float64
}

// RelatedMetric 相关指标
type RelatedMetric struct {
	Name  string
	Value float64
}

// Playbook 自愈剧本
type Playbook struct {
	ID          string
	Name        string
	Trigger     string
	Steps       []PlaybookStep
	AutoExecute bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlaybookStep 剧本步骤
type PlaybookStep struct {
	Name            string
	Actions         []string
	AIAnalysis      bool
	RequiresApproval bool
	Timeout         time.Duration
}

// PlaybookExecution 剧本执行记录
type PlaybookExecution struct {
	ID          string
	PlaybookID  string
	Status      string // pending, running, success, failed
	StartTime   time.Time
	EndTime     *time.Time
	Output      string
	Error       string
}

// DashboardConfig 看板配置
type DashboardConfig struct {
	ID        string
	UserID    string
	Name      string
	Panels    []PanelConfig
	Refresh   string
	TimeRange string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PanelConfig 面板配置
type PanelConfig struct {
	ID      string
	Title   string
	Type    string // graph, stat, table, heatmap
	Query   string
	Options map[string]interface{}
}

// Dashboard 看板
type Dashboard struct {
	Config DashboardConfig
	Data   map[string]interface{}
}

// PrometheusConfig Prometheus 客户端配置
type PrometheusConfig struct {
	URL       string
	Timeout   time.Duration
	Username  string
	Password  string
	TLSVerify bool
}

// QueryRange 查询范围
type QueryRange struct {
	Start time.Time
	End   time.Time
	Step  time.Duration
}

// ConversationContext 对话上下文
type ConversationContext struct {
	Keywords []string
	Metrics  []string
	Hosts    []string
	TimeRange TimeRange
	Anomalies []AnomalyResult
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time
	End   time.Time
}
