import { request } from './request'

export interface MCPServer {
  name: string
  status: 'healthy' | 'unhealthy'
}

export interface MCPTool {
  name: string
  description: string
  inputSchema: Record<string, any>
}

export interface MCPStats {
  clients: number
  adapters: number
  tools_by_server: Record<string, number>
}

export interface AddServerRequest {
  name: string
  url: string
  timeout?: number
}

export function getMCPServers() {
  return request.get<{ servers: MCPServer[]; total: number }>('/mcp/servers')
}

export function getMCPStats() {
  return request.get<MCPStats>('/mcp/stats')
}

export function getMCPHealth() {
  return request.get<Record<string, string>>('/mcp/health')
}

export function addMCPServer(data: AddServerRequest) {
  return request.post<{ message: string; name: string }>('/mcp/servers', data)
}

export function removeMCPServer(name: string) {
  return request.delete<{ message: string; name: string }>(`/mcp/servers/${name}`)
}

export function getMCPServerTools(serverName: string) {
  return request.get<{ server: string; tools: MCPTool[]; total: number }>(`/mcp/servers/${serverName}/tools`)
}

export function getAllMCPTools() {
  return request.get<{ tools_by_server: Record<string, MCPTool[]>; total: number }>('/mcp/tools')
}
