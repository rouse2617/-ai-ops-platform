import { request } from './request'

export interface AnalyzeRequest {
  results: unknown[]
  question?: string
}

export interface ToolCall {
  tool: string
  params: Record<string, unknown>
  result?: unknown
  error?: string
}

export interface AnalyzeResponse {
  id: string
  analysis_result: string
  question?: string
  tool_calls?: ToolCall[]
  created_at: string
}

export interface Analysis {
  id: string
  session_id?: string
  results_data: string
  analysis_result: string
  question?: string
  tool_calls?: ToolCall[]
  created_at: string
}

export interface AnalysisHistoryParams {
  session_id?: string
  limit?: number
}

// 语义化解读相关类型
export interface SemanticInsightRequest {
  tool_name: string
  tool_result: unknown
  host_id?: string
}

export interface Metric {
  name: string
  value: string
  status: 'normal' | 'warning' | 'critical'
  threshold?: string
}

export interface SemanticInsight {
  summary: string
  risk_level: 'normal' | 'warning' | 'critical'
  trend?: string
  recommendation?: string
  key_metrics?: Metric[]
}

// 上下文操作推荐相关类型
export interface ContextualAction {
  id: string
  label: string
  icon: string
  command?: string
  highlight: boolean
  ai_recommended: boolean
  confidence?: number
  risk_level: 'low' | 'medium' | 'high'
  description?: string
}

export interface SuggestActionsRequest {
  tool_results: Array<{
    tool_name: string
    result: unknown
    host_id?: string
  }>
}

// ============================================================================
// Analysis API Functions
// ============================================================================

export function analyzeResults(data: AnalyzeRequest): Promise<AnalyzeResponse> {
  return request.post<AnalyzeResponse>('/analysis/analyze', data)
}

export function getAnalysisHistory(params?: AnalysisHistoryParams): Promise<Analysis[]> {
  return request.get<Analysis[]>('/analysis/history', { params })
}

export function getAnalysis(id: string): Promise<Analysis> {
  return request.get<Analysis>(`/analysis/${id}`)
}

export function deleteAnalysis(id: string): Promise<void> {
  return request.delete(`/analysis/${id}`)
}

// 获取工具执行结果的语义化解读
export function getSemanticInsight(data: SemanticInsightRequest): Promise<SemanticInsight> {
  return request.post<SemanticInsight>('/analysis/semantic', data)
}

// 获取上下文相关的推荐操作
export function getSuggestedActions(data: SuggestActionsRequest): Promise<ContextualAction[]> {
  return request.post<ContextualAction[]>('/analysis/suggest-actions', data)
}

// 历史关联分析相关类型
export interface HistoricalCorrelationRequest {
  host_id: string
  current_issue: string
  issue_type?: string
  keywords?: string[]
}

export interface HistoricalCorrelation {
  host_id: string
  current_issue: string
  similar_incident?: string
  occurred_at?: string
  resolution?: string
  confidence: number
}

// 获取历史关联分析
export function getHistoricalCorrelation(data: HistoricalCorrelationRequest): Promise<HistoricalCorrelation> {
  return request.post<HistoricalCorrelation>('/analysis/correlation', data)
}






