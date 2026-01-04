import { request } from './request'

export interface Script {
  id: string
  name: string
  description: string
  language: 'bash' | 'python' | 'powershell'
  content: string
  parameters?: ScriptParameter[]
  enabled: boolean
  createdAt?: string
  updatedAt?: string
}

export interface ScriptParameter {
  name: string
  type: string
  description: string
  required: boolean
  default?: any
}

export interface ScriptListResponse {
  list: Script[]
  total: number
}

// 获取脚本列表 - 使用工具系统的脚本工具 API
export function getScripts(params?: { page?: number; pageSize?: number; keyword?: string }) {
  return request.get<any>('/tools/script').then((res) => {
    // 将工具格式转换为脚本格式
    const tools = Array.isArray(res) ? res : (res?.tools ?? [])
    const scripts: Script[] = tools.map((t: any) => ({
      id: t.name || t.id || '',
      name: t.name || '',
      description: t.description || '',
      language: 'bash' as const,
      content: t.content || '',
      parameters: t.parameters || [],
      enabled: t.enabled ?? true,
      createdAt: t.createdAt,
      updatedAt: t.updatedAt
    }))

    // 应用关键词过滤
    let filtered = scripts
    if (params?.keyword) {
      const kw = params.keyword.toLowerCase()
      filtered = scripts.filter(s =>
        s.name.toLowerCase().includes(kw) ||
        s.description.toLowerCase().includes(kw)
      )
    }

    return {
      list: filtered,
      total: filtered.length
    }
  })
}

// 获取单个脚本
export function getScript(id: string) {
  return request.get<Script>(`/scripts/${id}`)
}

// 创建脚本
export function createScript(data: Omit<Script, 'id' | 'createdAt' | 'updatedAt'>) {
  return request.post<Script>('/scripts', data)
}

// 更新脚本
export function updateScript(id: string, data: Partial<Script>) {
  return request.put<Script>(`/scripts/${id}`, data)
}

// 删除脚本
export function deleteScript(id: string) {
  return request.delete(`/scripts/${id}`)
}

// 启用/禁用脚本
export function toggleScript(id: string, enabled: boolean) {
  return request.put(`/scripts/${id}/toggle`, { enabled })
}

// 上传脚本文件
export function uploadScript(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<Script>('/scripts/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 测试脚本
export function testScript(id: string, params?: Record<string, any>) {
  return request.post<{ output: string; exitCode: number }>(`/scripts/${id}/test`, { params })
}
