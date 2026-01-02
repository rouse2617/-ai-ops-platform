import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// MCP 连接状态
export interface MCPConnection {
  id: string
  name: string
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
  latency: number // ms
  lastPing: number
}

// 任务进度
export interface TaskProgress {
  id: string
  type: 'scan' | 'analysis' | 'execution'
  title: string
  progress: number // 0-100
  status: 'running' | 'paused' | 'completed' | 'error'
  startTime: number
  estimatedTime?: number
}

// 会话健康状态
export interface SessionHealth {
  sessionId: string
  score: number // 0-100
  metrics: {
    responseTime: number
    errorRate: number
    toolSuccessRate: number
  }
  lastUpdated: number
}

export const useSystemStore = defineStore('system', () => {
  // MCP 连接状态
  const mcpConnections = ref<MCPConnection[]>([])

  // 任务进度列表
  const taskProgress = ref<TaskProgress[]>([])

  // 会话健康状态
  const sessionHealthMap = ref<Map<string, SessionHealth>>(new Map())

  // 计算属性：整体连接状态
  const overallConnectionStatus = computed(() => {
    if (mcpConnections.value.length === 0) return 'disconnected'
    const allConnected = mcpConnections.value.every(c => c.status === 'connected')
    const anyError = mcpConnections.value.some(c => c.status === 'error')
    if (anyError) return 'error'
    if (allConnected) return 'connected'
    return 'partial'
  })

  // 计算属性：平均延迟
  const averageLatency = computed(() => {
    const connected = mcpConnections.value.filter(c => c.status === 'connected')
    if (connected.length === 0) return 0
    return Math.round(connected.reduce((sum, c) => sum + c.latency, 0) / connected.length)
  })

  // 计算属性：活跃任务数
  const activeTaskCount = computed(() => {
    return taskProgress.value.filter(t => t.status === 'running').length
  })

  // 更新 MCP 连接状态
  function updateMCPConnection(connection: MCPConnection) {
    const index = mcpConnections.value.findIndex(c => c.id === connection.id)
    if (index >= 0) {
      mcpConnections.value[index] = connection
    } else {
      mcpConnections.value.push(connection)
    }
  }

  // 添加任务进度
  function addTaskProgress(task: TaskProgress) {
    taskProgress.value.push(task)
  }

  // 更新任务进度
  function updateTaskProgress(taskId: string, updates: Partial<TaskProgress>) {
    const task = taskProgress.value.find(t => t.id === taskId)
    if (task) {
      Object.assign(task, updates)
      // 自动清理已完成任务（5秒后）
      if (updates.status === 'completed' || updates.status === 'error') {
        setTimeout(() => {
          removeTaskProgress(taskId)
        }, 5000)
      }
    }
  }

  // 移除任务进度
  function removeTaskProgress(taskId: string) {
    const index = taskProgress.value.findIndex(t => t.id === taskId)
    if (index >= 0) {
      taskProgress.value.splice(index, 1)
    }
  }

  // 更新会话健康状态
  function updateSessionHealth(health: SessionHealth) {
    sessionHealthMap.value.set(health.sessionId, health)
  }

  // 获取会话健康状态
  function getSessionHealth(sessionId: string): SessionHealth | undefined {
    return sessionHealthMap.value.get(sessionId)
  }

  // 计算会话健康评分
  function calculateSessionScore(sessionId: string, metrics: {
    avgResponseTime: number
    errorCount: number
    totalMessages: number
    toolSuccessCount: number
    toolTotalCount: number
  }): number {
    let score = 100

    // 响应时间评分 (0-30分)
    if (metrics.avgResponseTime > 5000) score -= 30
    else if (metrics.avgResponseTime > 3000) score -= 20
    else if (metrics.avgResponseTime > 1000) score -= 10

    // 错误率评分 (0-40分)
    const errorRate = metrics.totalMessages > 0 ? metrics.errorCount / metrics.totalMessages : 0
    score -= errorRate * 40

    // 工具成功率评分 (0-30分)
    const toolSuccessRate = metrics.toolTotalCount > 0
      ? metrics.toolSuccessCount / metrics.toolTotalCount
      : 1
    score -= (1 - toolSuccessRate) * 30

    return Math.max(0, Math.min(100, Math.round(score)))
  }

  return {
    mcpConnections,
    taskProgress,
    sessionHealthMap,
    overallConnectionStatus,
    averageLatency,
    activeTaskCount,
    updateMCPConnection,
    addTaskProgress,
    updateTaskProgress,
    removeTaskProgress,
    updateSessionHealth,
    getSessionHealth,
    calculateSessionScore
  }
})
