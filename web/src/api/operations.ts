import { request } from './request'

export interface BatchExecuteRequest {
  operation: 'query_log' | 'run_command' | 'check_cpu' | 'check_memory' | 'check_disk' | 'check_process' | 'check_filesystem' | 'check_raid' | 'check_lvm' | 'check_io'
  hosts: string[]
  params?: Record<string, any>
}

export interface BatchExecuteResult {
  host: string
  status: 'success' | 'error'
  result?: any
  error?: string
  elapsed: string
}

export type BatchExecuteResponse = BatchExecuteResult[]

// 批量执行操作
export async function batchExecute(data: BatchExecuteRequest): Promise<BatchExecuteResponse> {
  const response = await request.post<BatchExecuteResponse>('/operations/batch-execute', data)
  // 确保返回数组
  return Array.isArray(response) ? response : []
}


