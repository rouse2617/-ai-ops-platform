package handler

import (
	"ai-pro/internal/prometheus"

	"github.com/gin-gonic/gin"
)

// PrometheusSemanticHandler Prometheus 语义化处理器
type PrometheusSemanticHandler struct {
	translator      *prometheus.SemanticTranslator
	nlQuery         *prometheus.NLQueryTranslator
	anomalyDetector *prometheus.DynamicAnomalyDetector
	playbook        *prometheus.AIPlaybook
	dashboard       *prometheus.DynamicDashboard
}

// NewPrometheusSemanticHandler 创建处理器
func NewPrometheusSemanticHandler(
	translator *prometheus.SemanticTranslator,
	nlQuery *prometheus.NLQueryTranslator,
	anomalyDetector *prometheus.DynamicAnomalyDetector,
	playbook *prometheus.AIPlaybook,
	dashboard *prometheus.DynamicDashboard,
) *PrometheusSemanticHandler {
	return &PrometheusSemanticHandler{
		translator:      translator,
		nlQuery:         nlQuery,
		anomalyDetector: anomalyDetector,
		playbook:        playbook,
		dashboard:       dashboard,
	}
}

// TranslateAlert 语义化翻译告警
// POST /api/prometheus/semantic/translate
func (h *PrometheusSemanticHandler) TranslateAlert(c *gin.Context) {
	var req prometheus.AlertContext
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.translator.TranslateAlert(c.Request.Context(), req)
	if err != nil {
		InternalError(c, "翻译失败: "+err.Error())
		return
	}

	Success(c, result)
}

// NLQuery 自然语言查询
// POST /api/prometheus/semantic/query
func (h *PrometheusSemanticHandler) NLQuery(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.nlQuery.TranslateAndExecute(c.Request.Context(), req.Query)
	if err != nil {
		InternalError(c, "查询失败: "+err.Error())
		return
	}

	Success(c, result)
}

// GetQuerySuggestions 获取查询建议
// GET /api/prometheus/semantic/suggestions
func (h *PrometheusSemanticHandler) GetQuerySuggestions(c *gin.Context) {
	partial := c.Query("q")
	suggestions := h.nlQuery.GetSuggestions(partial)
	Success(c, gin.H{"suggestions": suggestions})
}

// DetectAnomaly 检测异常
// POST /api/prometheus/semantic/anomaly
func (h *PrometheusSemanticHandler) DetectAnomaly(c *gin.Context) {
	var req struct {
		MetricName   string  `json:"metric_name" binding:"required"`
		Instance     string  `json:"instance" binding:"required"`
		CurrentValue float64 `json:"current_value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.anomalyDetector.DetectAnomaly(c.Request.Context(), req.MetricName, req.Instance, req.CurrentValue)
	if err != nil {
		InternalError(c, "检测失败: "+err.Error())
		return
	}

	Success(c, result)
}

// LearnBaseline 学习基线
// POST /api/prometheus/semantic/baseline
func (h *PrometheusSemanticHandler) LearnBaseline(c *gin.Context) {
	var req struct {
		MetricName string `json:"metric_name" binding:"required"`
		Instance   string `json:"instance" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	baseline, err := h.anomalyDetector.LearnBaseline(c.Request.Context(), req.MetricName, req.Instance)
	if err != nil {
		InternalError(c, "学习基线失败: "+err.Error())
		return
	}

	Success(c, baseline)
}

// GetBaseline 获取基线
// GET /api/prometheus/semantic/baseline
func (h *PrometheusSemanticHandler) GetBaseline(c *gin.Context) {
	metricName := c.Query("metric")
	instance := c.Query("instance")

	if metricName == "" || instance == "" {
		ParamError(c, "metric 和 instance 参数必填")
		return
	}

	baseline := h.anomalyDetector.GetBaseline(metricName, instance)
	if baseline == nil {
		NotFoundError(c, "未找到基线数据")
		return
	}

	Success(c, baseline)
}

// MatchPlaybook 匹配排障剧本
// POST /api/prometheus/semantic/playbook/match
func (h *PrometheusSemanticHandler) MatchPlaybook(c *gin.Context) {
	var req struct {
		AlertName string            `json:"alert_name" binding:"required"`
		Labels    map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	playbook := h.playbook.MatchPlaybook(req.AlertName, req.Labels)
	if playbook == nil {
		Success(c, gin.H{"matched": false, "message": "未找到匹配的剧本"})
		return
	}

	Success(c, gin.H{"matched": true, "playbook": playbook})
}

// ExecutePlaybook 执行排障剧本
// POST /api/prometheus/semantic/playbook/execute
func (h *PrometheusSemanticHandler) ExecutePlaybook(c *gin.Context) {
	var req struct {
		PlaybookID string `json:"playbook_id" binding:"required"`
		Instance   string `json:"instance" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	playbook := h.playbook.GetPlaybook(req.PlaybookID)
	if playbook == nil {
		NotFoundError(c, "剧本不存在")
		return
	}

	execution, err := h.playbook.ExecutePlaybook(c.Request.Context(), playbook, req.Instance)
	if err != nil {
		InternalError(c, "执行失败: "+err.Error())
		return
	}

	Success(c, execution)
}

// ListPlaybooks 列出所有剧本
// GET /api/prometheus/semantic/playbooks
func (h *PrometheusSemanticHandler) ListPlaybooks(c *gin.Context) {
	playbooks := h.playbook.ListPlaybooks()
	Success(c, gin.H{"playbooks": playbooks})
}

// GenerateDashboard 生成动态看板
// POST /api/prometheus/semantic/dashboard
func (h *PrometheusSemanticHandler) GenerateDashboard(c *gin.Context) {
	var req struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 优先使用 topic，否则从 message 中提取
	topic := req.Topic
	if topic == "" && req.Message != "" {
		topic = h.dashboard.ExtractTopicFromMessage(req.Message)
	}

	config := h.dashboard.GenerateDashboard(c.Request.Context(), topic)
	Success(c, config)
}

// GetDashboardTopics 获取可用话题
// GET /api/prometheus/semantic/dashboard/topics
func (h *PrometheusSemanticHandler) GetDashboardTopics(c *gin.Context) {
	topics := h.dashboard.GetAvailableTopics()
	Success(c, gin.H{"topics": topics})
}

// AnalyzeMessage 分析消息中的关键词
// POST /api/prometheus/semantic/analyze
func (h *PrometheusSemanticHandler) AnalyzeMessage(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	matches := h.dashboard.MatchKeywordsInText(req.Message)
	topic := h.dashboard.ExtractTopicFromMessage(req.Message)

	Success(c, gin.H{
		"topic":    topic,
		"keywords": matches,
	})
}

// NotFoundError 返回 404 错误
func NotFoundError(c *gin.Context, message string) {
	c.JSON(404, gin.H{
		"code":    404,
		"message": message,
	})
}
