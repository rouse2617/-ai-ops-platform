import request from './request'

export interface FeedbackRequest {
  message_id: string
  session_id: string
  question: string
  answer: string
  rating: number // 1=好, -1=差
  comment?: string
}

export interface FeedbackResponse {
  id: string
  message: string
}

export function submitFeedback(data: FeedbackRequest): Promise<FeedbackResponse> {
  return request.post('/feedback', data)
}

export function getFeedback(messageId: string) {
  return request.get(`/feedback/${messageId}`)
}
