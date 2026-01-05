import { defineStore } from 'pinia'
import { ref, computed } from 'vue'


export interface DangerOperation {
  id: string
  command: string
  type: 'delete' | 'restart' | 'stop' | 'modify'
  riskLevel: 'high' | 'critical'
  affectedHosts: string[]
  reason?: string
  timestamp: number
}

export interface ConfirmRecord {
  operationId: string
  command: string
  confirmed: boolean
  timestamp: number
  userId?: string
}

export type DangerLevel = 'safe' | 'low' | 'medium' | 'high' | 'critical'

interface DangerPattern {
  pattern: RegExp
  level: DangerLevel
  reason: string
  type: DangerOperation['type']
}

const DANGER_PATTERNS: DangerPattern[] = [
  {
    pattern: /rm\s+-rf\s+\//,
    level: 'critical',
    reason: '删除根目录，将导致系统完全损坏',
    type: 'delete'
  },
  {
    pattern: /rm\s+-rf\s+\/(?:etc|usr|var|boot|sys)/,
    level: 'critical',
    reason: '删除系统关键目录，将导致系统无法启动',
    type: 'delete'
  },
  {
    pattern: /rm\s+-rf\s+.*\*/,
    level: 'high',
    reason: '使用通配符删除，可能误删重要文件',
    type: 'delete'
  },
  {
    pattern: /shutdown|reboot|init\s+[06]/,
    level: 'high',
    reason: '系统重启或关机，将中断所有服务',
    type: 'restart'
  },
  {
    pattern: /systemctl\s+stop\s+(?:sshd|network|firewalld)/,
    level: 'high',
    reason: '停止关键服务，可能导致无法远程连接',
    type: 'stop'
  },
  {
    pattern: /chmod\s+777\s+/,
    level: 'medium',
    reason: '设置过于宽松的权限，存在安全风险',
    type: 'modify'
  },
  {
    pattern: /dd\s+if=.*of=\/dev\/(?:sda|vda|nvme)/,
    level: 'critical',
    reason: '直接写入磁盘设备，将破坏所有数据',
    type: 'delete'
  },
  {
    pattern: /mkfs\./,
    level: 'critical',
    reason: '格式化磁盘，将清除所有数据',
    type: 'delete'
  },
  {
    pattern: /iptables\s+-F/,
    level: 'high',
    reason: '清空防火墙规则，可能导致安全风险',
    type: 'modify'
  }
]

export const useDangerStore = defineStore('danger', () => {
  const pendingOperations = ref<DangerOperation[]>([])
  const confirmHistory = ref<ConfirmRecord[]>([])
  const autoConfirmDisabled = ref(true)

  const hasPendingOperations = computed(() => pendingOperations.value.length > 0)

  /**
   * 检查命令的危险等级
   */
  function checkDangerLevel(command: string): {
    level: DangerLevel
    reason?: string
    type?: DangerOperation['type']
  } {
    const trimmedCommand = command.trim()

    for (const pattern of DANGER_PATTERNS) {
      if (pattern.pattern.test(trimmedCommand)) {
        return {
          level: pattern.level,
          reason: pattern.reason,
          type: pattern.type
        }
      }
    }

    return { level: 'safe' }
  }

  /**
   * 请求用户确认危险操作
   */
  async function requestConfirm(operation: DangerOperation): Promise<boolean> {
    return new Promise((resolve) => {
      pendingOperations.value.push(operation)

      // 触发确认对话框（由组件监听 pendingOperations 变化）
      const checkInterval = setInterval(() => {
        const index = pendingOperations.value.findIndex(op => op.id === operation.id)
        if (index === -1) {
          clearInterval(checkInterval)
          const record = confirmHistory.value.find(r => r.operationId === operation.id)
          resolve(record?.confirmed || false)
        }
      }, 100)

      // 超时处理（30秒）
      setTimeout(() => {
        clearInterval(checkInterval)
        const index = pendingOperations.value.findIndex(op => op.id === operation.id)
        if (index !== -1) {
          pendingOperations.value.splice(index, 1)
          resolve(false)
        }
      }, 30000)
    })
  }

  /**
   * 记录确认结果
   */
  function recordConfirm(operation: DangerOperation, confirmed: boolean): void {
    const record: ConfirmRecord = {
      operationId: operation.id,
      command: operation.command,
      confirmed,
      timestamp: Date.now()
    }

    confirmHistory.value.push(record)

    // 只保留最近100条记录
    if (confirmHistory.value.length > 100) {
      confirmHistory.value = confirmHistory.value.slice(-100)
    }

    // 从待处理列表中移除
    const index = pendingOperations.value.findIndex(op => op.id === operation.id)
    if (index !== -1) {
      pendingOperations.value.splice(index, 1)
    }

    // 持久化到 localStorage
    saveToStorage()
  }

  /**
   * 取消操作
   */
  function cancelOperation(operationId: string): void {
    const index = pendingOperations.value.findIndex(op => op.id === operationId)
    if (index !== -1) {
      const operation = pendingOperations.value[index]
      recordConfirm(operation, false)
    }
  }

  /**
   * 清空历史记录
   */
  function clearHistory(): void {
    confirmHistory.value = []
    saveToStorage()
  }

  /**
   * 获取统计信息
   */
  const statistics = computed(() => {
    const total = confirmHistory.value.length
    const confirmed = confirmHistory.value.filter(r => r.confirmed).length
    const rejected = total - confirmed

    return {
      total,
      confirmed,
      rejected,
      confirmRate: total > 0 ? (confirmed / total * 100).toFixed(1) : '0'
    }
  })

  /**
   * 保存到 localStorage
   */
  function saveToStorage(): void {
    try {
      localStorage.setItem('danger-confirm-history', JSON.stringify(confirmHistory.value))
    } catch (error) {
      console.error('Failed to save danger confirm history:', error)
    }
  }

  /**
   * 从 localStorage 加载
   */
  function loadFromStorage(): void {
    try {
      const saved = localStorage.getItem('danger-confirm-history')
      if (saved) {
        confirmHistory.value = JSON.parse(saved)
      }
    } catch (error) {
      console.error('Failed to load danger confirm history:', error)
    }
  }

  // 初始化时加载历史记录
  loadFromStorage()

  return {
    pendingOperations,
    confirmHistory,
    autoConfirmDisabled,
    hasPendingOperations,
    statistics,
    checkDangerLevel,
    requestConfirm,
    recordConfirm,
    cancelOperation,
    clearHistory
  }
})
