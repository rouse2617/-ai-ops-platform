import request from './request'

// 语义翻译
export interface SemanticTranslation {
  title: string
  description: string
  root_cause: string
  impact: string
  suggestions: string[]
}

export interface TranslateMetricRequest {
  metric_name: string
  value: number
  labels?: Record<string, string>
}

export function translateMetric(data: TranslateMetricRequest): Promise<SemanticTranslation> {
  return request.post('/prometheus/translate', data)
}

export interface TranslateAlertRequest {
  alert_name: string
  severity?: string
  description?: string
  labels?: Record<string, string>
}

export function translateAlert(data: TranslateAlertRequest): Promise<{ alert_name: string; translation: SemanticTranslation }> {
  return request.post('/prometheus/translate-alert', data)
}

// PromQL 生成
export interface PromQLQuery {
  query: string
  range: string
  step: string
  explanation: string
}

export function generatePromQL(question: string, context?: Record<string, unknown>): Promise<PromQLQuery> {
  return request.post('/prometheus/generate-promql', { question, context })
}

export function nlToPromQL(question: string): Promise<PromQLQuery> {
  return request.post('/prometheus/nl-to-promql', { question })
}

// 异常检测
export interface AnomalyResult {
  is_anomaly: boolean
  severity: number
  deviation: number
  recommendation: string
}

export interface BaselineMetrics {
  metric_name: string
  mean: number
  std_dev: number
  p95: number
  p99: number
  updated_at: string
}

export function detectAnomaly(metricName: string, value: number): Promise<AnomalyResult> {
  return request.post('/prometheus/detect-anomaly', { metric_name: metricName, value })
}

export function learnBaseline(metricName: string, start: string, end: string): Promise<BaselineMetrics> {
  return request.post('/prometheus/learn-baseline', { metric_name: metricName, start, end })
}

// 剧本
export interface PlaybookStep {
  name: string
  type: string
  actions?: string[]
  ai_analysis?: boolean
  requires_approval?: boolean
}

export interface Playbook {
  id: string
  name: string
  trigger: string
  steps: PlaybookStep[]
  auto_execute: boolean
  created_at?: string
  updated_at?: string
}

export interface PlaybookExecution {
  playbook_id: string
  status: string
  steps: { name: string; status: string; output?: string; error?: string }[]
  started_at: string
  completed_at?: string
}

export function listPlaybooks(filter?: { name?: string; trigger?: string }): Promise<Playbook[]> {
  return request.get('/prometheus/playbooks', { params: filter })
}

export function getPlaybook(id: string): Promise<Playbook> {
  return request.get(`/prometheus/playbooks/${id}`)
}

export function executePlaybook(id: string, context?: Record<string, unknown>): Promise<PlaybookExecution> {
  return request.post(`/prometheus/playbooks/${id}/execute`, { context })
}

// 查询
export interface QueryResult {
  query: string
  start: string
  end: string
  step: string
  result: unknown
}

export function queryPrometheus(query: string, start: string, end: string, step?: string): Promise<QueryResult> {
  return request.post('/prometheus/query', { query, start, end, step })
}

export function visualizeQuery(query: string, start: string, end: string): Promise<{ data: unknown[]; chart_type: string }> {
  return request.post('/prometheus/visualize', { query, start, end })
}

// 故障复盘
export interface IncidentReplayResult {
  incident_time: string
  host: string
  affected_metrics: string[]
  root_causes: string[]
  recommendations: string[]
  timeline?: unknown
}

export function incidentReplay(incidentTime: string, host: string, lookbackMinutes?: number): Promise<IncidentReplayResult> {
  return request.post('/prometheus/incident-replay', {
    incident_time: incidentTime,
    host,
    lookback_minutes: lookbackMinutes
  })
}

// 容量规划
export interface CapacityPlanningResult {
  trend: { type: string; growth_rate: number; projected_peak: number; days_to_capacity: number }
  scenarios: { name: string; growth_multiplier: number; recommended_action: string; time_to_action_days: number }[]
  recommendation: { current_usage: number; projected_usage: number; priority: string; action_items: string[] }
  data_points: number
}

export function capacityPlanning(
  metricQuery: string,
  days?: number,
  capacityThreshold?: number
): Promise<CapacityPlanningResult> {
  return request.post('/prometheus/capacity-planning', {
    metric_query: metricQuery,
    days,
    capacity_threshold: capacityThreshold
  })
}
