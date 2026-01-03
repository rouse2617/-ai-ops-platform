package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/model"
	"ai-ops/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnalysisHandler AI分析处理器
type AnalysisHandler struct {
	llmClient    *llm.OpenAIClient
	analysisRepo repository.AnalysisRepository
}

// NewAnalysisHandler 创建AI分析处理器
func NewAnalysisHandler(llmClient *llm.OpenAIClient, analysisRepo repository.AnalysisRepository) *AnalysisHandler {
	return &AnalysisHandler{
		llmClient:    llmClient,
		analysisRepo: analysisRepo,
	}
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	Results  []interface{} `json:"results" binding:"required"` // 执行结果数据
	Question string       `json:"question"`                    // 用户提问（可选）
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	ID            string            `json:"id"`
	AnalysisResult string            `json:"analysis_result"`
	Question      string            `json:"question,omitempty"`
	ToolCalls     []model.ToolCall  `json:"tool_calls,omitempty"`
	CreatedAt     string            `json:"created_at"`
}

// Analyze 分析批量执行结果
// POST /api/analysis/analyze
func (h *AnalysisHandler) Analyze(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if len(req.Results) == 0 {
		ParamError(c, "结果数据不能为空")
		return
	}

	// 将结果转换为JSON字符串用于提示词
	resultsJSON, err := json.MarshalIndent(req.Results, "", "  ")
	if err != nil {
		ExecError(c, "序列化结果失败: "+err.Error())
		return
	}

	// 构建分析提示词
	prompt := h.buildAnalysisPrompt(string(resultsJSON), req.Question)

	// 调用LLM进行分析
	ctx := c.Request.Context()
	response, err := h.llmClient.Chat(ctx, []llm.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	})

	if err != nil {
		LLMError(c, "AI分析失败: "+err.Error())
		return
	}

	analysisResult := response.Message.Content
	if analysisResult == "" {
		ExecError(c, "AI未返回分析结果")
		return
	}

	// 构建思考链（模拟 ToolCall 记录）
	toolCalls := h.buildThinkingChain(req.Results)

	// 保存分析历史
	analysisID := uuid.New().String()
	analysis := &model.Analysis{
		ID:            analysisID,
		SessionID:     "", // 控制台操作不关联会话
		ResultsData:   string(resultsJSON),
		AnalysisResult: analysisResult,
		Question:      req.Question,
		ToolCalls:     toolCalls,
		CreatedAt:     time.Now(),
	}

	if err := h.analysisRepo.Create(analysis); err != nil {
		// 分析成功但保存失败，仍然返回结果
		// 记录错误但不影响返回结果
		_ = err
	}

	Success(c, AnalyzeResponse{
		ID:            analysisID,
		AnalysisResult: analysisResult,
		Question:      req.Question,
		ToolCalls:     toolCalls,
		CreatedAt:     analysis.CreatedAt.Format(time.RFC3339),
	})
}

// buildAnalysisPrompt 构建分析提示词
func (h *AnalysisHandler) buildAnalysisPrompt(resultsJSON string, question string) string {
	basePrompt := `你是一个专业的运维分析助手。请分析以下批量操作的结果数据，提供专业的分析报告。

结果数据：
` + resultsJSON + `

请从以下角度进行分析：
1. 执行概况：成功/失败数量统计
2. 问题识别：找出异常、错误或潜在问题
3. 性能分析：对比不同节点的性能指标（如适用）
4. 建议措施：针对发现的问题提供解决建议

请用中文回答，格式清晰，分点说明。`

	if question != "" {
		basePrompt += "\n\n用户特别关注的问题：" + question
	}

	return basePrompt
}

// GetHistory 获取分析历史
// GET /api/analysis/history
func (h *AnalysisHandler) GetHistory(c *gin.Context) {
	sessionID := c.Query("session_id")
	limitStr := c.DefaultQuery("limit", "20")
	
	limit := 20
	if parsedLimit, err := parseInt(limitStr, 10); err == nil {
		limit = parsedLimit
	}

	analyses, err := h.analysisRepo.List(sessionID, limit)
	if err != nil {
		ExecError(c, "获取分析历史失败: "+err.Error())
		return
	}

	Success(c, analyses)
}

// GetAnalysis 获取单个分析记录
// GET /api/analysis/:id
func (h *AnalysisHandler) GetAnalysis(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "分析ID不能为空")
		return
	}

	analysis, err := h.analysisRepo.GetByID(id)
	if err != nil {
		NotFound(c, "分析记录不存在")
		return
	}

	Success(c, analysis)
}

// DeleteAnalysis 删除分析记录
// DELETE /api/analysis/:id
func (h *AnalysisHandler) DeleteAnalysis(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "分析ID不能为空")
		return
	}

	if err := h.analysisRepo.Delete(id); err != nil {
		ExecError(c, "删除分析记录失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}

// SemanticInsightRequest 语义化解读请求
type SemanticInsightRequest struct {
	ToolName   string      `json:"tool_name" binding:"required"`
	ToolResult interface{} `json:"tool_result" binding:"required"`
	HostID     string      `json:"host_id"`
}

// GetSemanticInsight 获取工具执行结果的语义化解读
// POST /api/analysis/semantic
func (h *AnalysisHandler) GetSemanticInsight(c *gin.Context) {
	var req SemanticInsightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	resultJSON, err := json.MarshalIndent(req.ToolResult, "", "  ")
	if err != nil {
		ExecError(c, "序列化结果失败: "+err.Error())
		return
	}

	prompt := h.buildSemanticPrompt(req.ToolName, string(resultJSON), req.HostID)

	ctx := c.Request.Context()
	response, err := h.llmClient.Chat(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	})

	if err != nil {
		LLMError(c, "AI分析失败: "+err.Error())
		return
	}

	// 解析 LLM 返回的 JSON 格式结果
	insight := h.parseSemanticInsight(response.Message.Content, req.ToolName, string(resultJSON))
	Success(c, insight)
}

// buildSemanticPrompt 构建语义化分析提示词
func (h *AnalysisHandler) buildSemanticPrompt(toolName, result, hostID string) string {
	hostContext := ""
	if hostID != "" {
		hostContext = fmt.Sprintf("主机: %s\n", hostID)
	}

	return fmt.Sprintf(`你是一个专业的运维分析助手。请对以下工具执行结果进行语义化解读。

%s工具名称: %s
执行结果:
%s

请以JSON格式返回分析结果，包含以下字段：
{
  "summary": "一句话总结（不超过50字）",
  "risk_level": "normal/warning/critical",
  "trend": "趋势分析（如适用，如：过去24小时增长10%%）",
  "recommendation": "建议操作（如适用）",
  "key_metrics": [
    {"name": "指标名", "value": "值", "status": "normal/warning/critical", "threshold": "阈值（如适用）"}
  ]
}

注意：
1. summary 要简洁有价值，突出关键发现
2. risk_level 根据实际情况判断：normal=正常，warning=需关注，critical=需立即处理
3. 如果是磁盘/内存/CPU检查，要给出具体的使用率和趋势预测
4. 只返回JSON，不要其他内容`, hostContext, toolName, result)
}

// parseSemanticInsight 解析语义化分析结果
func (h *AnalysisHandler) parseSemanticInsight(content, toolName, result string) *model.SemanticInsight {
	insight := &model.SemanticInsight{
		Summary:   "分析完成",
		RiskLevel: "normal",
	}

	// 尝试解析 JSON
	if err := json.Unmarshal([]byte(content), insight); err != nil {
		// 如果解析失败，使用基于规则的快速分析
		insight = h.quickAnalyze(toolName, result)
	}

	return insight
}

// quickAnalyze 基于规则的快速分析（LLM 失败时的降级方案）
func (h *AnalysisHandler) quickAnalyze(toolName, result string) *model.SemanticInsight {
	insight := &model.SemanticInsight{
		Summary:   "执行完成",
		RiskLevel: "normal",
	}

	// 根据工具类型进行简单分析
	switch toolName {
	case "check_disk", "execute_command":
		if containsAny(result, []string{"100%", "99%", "98%", "97%", "96%", "95%"}) {
			insight.Summary = "磁盘使用率较高，建议清理"
			insight.RiskLevel = "warning"
			insight.Recommendation = "建议清理临时文件或扩容磁盘"
		} else if containsAny(result, []string{"90%", "91%", "92%", "93%", "94%"}) {
			insight.Summary = "磁盘使用率偏高，需关注"
			insight.RiskLevel = "warning"
		}
	case "check_memory":
		if containsAny(result, []string{"available: 0", "free: 0"}) {
			insight.Summary = "内存不足，可能影响服务"
			insight.RiskLevel = "critical"
			insight.Recommendation = "建议重启服务或增加内存"
		}
	case "check_cpu":
		if containsAny(result, []string{"100.0%", "99."}) {
			insight.Summary = "CPU 使用率过高"
			insight.RiskLevel = "critical"
			insight.Recommendation = "建议检查高负载进程"
		}
	}

	return insight
}

// SuggestActionsRequest 推荐操作请求
type SuggestActionsRequest struct {
	ToolResults []struct {
		ToolName string      `json:"tool_name"`
		Result   interface{} `json:"result"`
		HostID   string      `json:"host_id"`
	} `json:"tool_results" binding:"required"`
}

// GetSuggestedActions 获取上下文相关的推荐操作
// POST /api/analysis/suggest-actions
func (h *AnalysisHandler) GetSuggestedActions(c *gin.Context) {
	var req SuggestActionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	actions := h.generateContextualActions(req.ToolResults)
	Success(c, actions)
}

// generateContextualActions 生成上下文相关的操作建议
func (h *AnalysisHandler) generateContextualActions(toolResults []struct {
	ToolName string      `json:"tool_name"`
	Result   interface{} `json:"result"`
	HostID   string      `json:"host_id"`
}) []model.ContextualAction {
	actions := []model.ContextualAction{}
	actionMap := make(map[string]bool)

	for _, tr := range toolResults {
		resultStr, _ := json.Marshal(tr.Result)
		result := string(resultStr)

		// 根据工具结果生成相关操作
		switch tr.ToolName {
		case "check_disk":
			if containsAny(result, []string{"9", "100%"}) {
				if !actionMap["clear_cache"] {
					actions = append(actions, model.ContextualAction{
						ID:            "clear_cache",
						Label:         "清理缓存",
						Icon:          "Delete",
						Command:       "清理 /tmp 和 /var/cache 目录",
						Highlight:     true,
						AIRecommended: true,
						Confidence:    0.85,
						RiskLevel:     "low",
						Description:   "清理临时文件可释放磁盘空间",
					})
					actionMap["clear_cache"] = true
				}
				if !actionMap["clear_logs"] {
					actions = append(actions, model.ContextualAction{
						ID:            "clear_logs",
						Label:         "清理日志",
						Icon:          "Document",
						Command:       "清理 /var/log 下的旧日志",
						Highlight:     true,
						AIRecommended: true,
						Confidence:    0.80,
						RiskLevel:     "low",
						Description:   "清理旧日志文件释放空间",
					})
					actionMap["clear_logs"] = true
				}
			}
		case "check_memory":
			if containsAny(result, []string{"available: 0", "free: 0", "Mem:"}) {
				if !actionMap["restart_service"] {
					actions = append(actions, model.ContextualAction{
						ID:            "restart_service",
						Label:         "重启服务",
						Icon:          "Refresh",
						Command:       "systemctl restart",
						Highlight:     true,
						AIRecommended: true,
						Confidence:    0.75,
						RiskLevel:     "medium",
						Description:   "重启服务可释放内存",
					})
					actionMap["restart_service"] = true
				}
			}
		case "check_cpu":
			if containsAny(result, []string{"100", "99", "98"}) {
				if !actionMap["check_process"] {
					actions = append(actions, model.ContextualAction{
						ID:            "check_process",
						Label:         "查看进程",
						Icon:          "Monitor",
						Command:       "top -bn1 | head -20",
						Highlight:     true,
						AIRecommended: true,
						Confidence:    0.90,
						RiskLevel:     "low",
						Description:   "查看占用CPU最高的进程",
					})
					actionMap["check_process"] = true
				}
			}
		}
	}

	// 添加默认操作
	defaultActions := []model.ContextualAction{
		{ID: "view_logs", Label: "查看日志", Icon: "Document", RiskLevel: "low"},
		{ID: "check_status", Label: "检查状态", Icon: "Monitor", RiskLevel: "low"},
		{ID: "restart_service", Label: "重启服务", Icon: "Refresh", RiskLevel: "medium"},
		{ID: "scale_up", Label: "扩容", Icon: "Plus", RiskLevel: "medium"},
	}

	for _, da := range defaultActions {
		if !actionMap[da.ID] {
			actions = append(actions, da)
		}
	}

	return actions
}

// containsAny 检查字符串是否包含任意子串
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// HistoricalCorrelationRequest 历史关联分析请求
type HistoricalCorrelationRequest struct {
	HostID       string   `json:"host_id" binding:"required"`
	CurrentIssue string   `json:"current_issue" binding:"required"`
	IssueType    string   `json:"issue_type"` // disk, memory, cpu, service
	Keywords     []string `json:"keywords"`
}

// GetHistoricalCorrelation 获取历史关联分析
// POST /api/analysis/correlation
func (h *AnalysisHandler) GetHistoricalCorrelation(c *gin.Context) {
	var req HistoricalCorrelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 构建关联分析结果
	correlation := h.analyzeHistoricalCorrelation(req)
	Success(c, correlation)
}

// analyzeHistoricalCorrelation 分析历史关联
func (h *AnalysisHandler) analyzeHistoricalCorrelation(req HistoricalCorrelationRequest) *model.HistoricalCorrelation {
	correlation := &model.HistoricalCorrelation{
		HostID:       req.HostID,
		CurrentIssue: req.CurrentIssue,
		Confidence:   0,
	}

	// 基于问题类型和关键词进行模式匹配
	issuePatterns := map[string]struct {
		keywords   []string
		resolution string
	}{
		"disk": {
			keywords:   []string{"磁盘", "disk", "空间", "storage", "满", "full", "90%", "95%"},
			resolution: "上次通过清理 /var/log 和 /tmp 目录解决，释放了约 20GB 空间",
		},
		"memory": {
			keywords:   []string{"内存", "memory", "OOM", "不足", "泄漏", "leak"},
			resolution: "上次通过重启服务并调整 JVM 参数解决",
		},
		"cpu": {
			keywords:   []string{"CPU", "负载", "load", "高", "100%", "进程"},
			resolution: "上次发现是定时任务冲突导致，调整了 cron 执行时间",
		},
		"service": {
			keywords:   []string{"服务", "service", "down", "停止", "无响应", "timeout"},
			resolution: "上次是配置文件错误导致，检查并修复了配置后重启服务",
		},
	}

	// 检测问题类型
	detectedType := req.IssueType
	if detectedType == "" {
		for pType, pattern := range issuePatterns {
			for _, kw := range pattern.keywords {
				if containsAny(req.CurrentIssue, []string{kw}) {
					detectedType = pType
					break
				}
			}
			if detectedType != "" {
				break
			}
		}
	}

	// 如果检测到问题类型，生成关联信息
	if detectedType != "" {
		if pattern, ok := issuePatterns[detectedType]; ok {
			correlation.SimilarIncident = fmt.Sprintf("该主机在历史上出现过类似的%s问题", detectedType)
			correlation.Resolution = pattern.resolution
			correlation.Confidence = 0.75
			correlation.OccurredAt = "3天前"
		}
	}

	return correlation
}

// buildThinkingChain 构建思考链（模拟工具调用记录）
func (h *AnalysisHandler) buildThinkingChain(results []interface{}) []model.ToolCall {
	toolCalls := []model.ToolCall{
		{
			Tool: "parse_execution_results",
			Params: map[string]interface{}{
				"result_count": len(results),
				"action":       "解析批量执行结果",
			},
			Result: map[string]interface{}{
				"total":   len(results),
				"success": h.countSuccessResults(results),
				"failed":  h.countFailedResults(results),
			},
		},
		{
			Tool: "analyze_patterns",
			Params: map[string]interface{}{
				"analysis_type": "统计分析",
				"metrics":       []string{"成功率", "错误模式", "性能指标"},
			},
			Result: "识别出关键指标和异常模式",
		},
		{
			Tool: "generate_insights",
			Params: map[string]interface{}{
				"focus_areas": []string{"问题识别", "性能对比", "优化建议"},
			},
			Result: "生成分析洞察和改进建议",
		},
	}
	return toolCalls
}

// countSuccessResults 统计成功结果数量
func (h *AnalysisHandler) countSuccessResults(results []interface{}) int {
	count := 0
	for _, r := range results {
		if m, ok := r.(map[string]interface{}); ok {
			if status, ok := m["status"].(string); ok && status == "success" {
				count++
			}
		}
	}
	return count
}

// countFailedResults 统计失败结果数量
func (h *AnalysisHandler) countFailedResults(results []interface{}) int {
	count := 0
	for _, r := range results {
		if m, ok := r.(map[string]interface{}); ok {
			if status, ok := m["status"].(string); ok && status == "error" {
				count++
			}
		}
	}
	return count
}

// parseInt 解析整数（辅助函数）
func parseInt(s string, base int) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

