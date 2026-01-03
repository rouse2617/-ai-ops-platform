import { request } from './request'

export interface SlotSuggestion {
  value: string
  label: string
  description?: string
  icon?: string
  iconColor?: string
  tag?: string
  tagType?: 'success' | 'info' | 'warning' | 'danger'
}

export interface AnalyzeMessageRequest {
  message: string
  hosts?: string[]
}

export interface AnalyzeMessageResponse {
  needsSlotFilling: boolean
  slotType: string
  suggestions: SlotSuggestion[]
  originalMessage: string
}

/**
 * 分析消息是否需要参数补全
 */
export function analyzeMessage(data: AnalyzeMessageRequest) {
  return request.post<AnalyzeMessageResponse>('/slot-filling/analyze', data)
}
