// 监控告警 WebSocket 集成
import { ref, onMounted, onUnmounted } from 'vue'
import { ElNotification } from 'element-plus'
import { useWebSocket } from './useWebSocket'
import type { Message, ProactiveMessageData } from '@/api/chat'

export interface Alert {
  id: string
  host_id: string
  host_name: string
  type: 'cpu' | 'memory' | 'disk' | 'network' | 'process' | 'service' | 'predictive'
  level: 'critical' | 'high' | 'medium' | 'low'
  title: string
  message: string
  metrics: Record<string, any>
  suggestions: string[]
  actions: QuickAction[]
  created_at: string
  resolved_at?: string
  status: 'active' | 'resolved' | 'ignored'
}

export interface QuickAction {
  id: string
  label: string
  command: string
  description?: string
  dangerous?: boolean
}

export interface HostMetrics {
  host_id: string
  host_name: string
  cpu: number
  memory: number
  disk: number
  load: number[]
  io_wait: number
  timestamp: string
}

// 主动消息回调类型
type ProactiveMessageCallback = (message: Message) => void

export function useMonitorAlerts() {
  const alerts = ref<Alert[]>([])
  const latestMetrics = ref<Map<string, HostMetrics>>(new Map())
  const proactiveMessages = ref<ProactiveMessageData[]>([])

  // 主动消息回调列表
  const proactiveCallbacks: ProactiveMessageCallback[] = []

  const wsUrl = `ws://${window.location.host}/api/monitor/ws`
  const { addMessageListener, connect, disconnect, isConnected } = useWebSocket({ url: wsUrl })

  const handleMessage = (message: any) => {
    if (message.type === 'alert') {
      const alert = message.data as Alert
      alerts.value.unshift(alert)
      showAlertNotification(alert)
    } else if (message.type === 'metrics') {
      const metrics = message.data as HostMetrics
      latestMetrics.value.set(metrics.host_name, metrics)
    } else if (message.type === 'proactive_message') {
      // 处理 AI 主动消息
      const proactiveData = message.data as ProactiveMessageData
      proactiveMessages.value.unshift(proactiveData)

      // 转换为聊天消息格式并通知回调
      const chatMessage: Message = {
        role: 'proactive',
        content: proactiveData.content || '',
        timestamp: proactiveData.created_at || Date.now(),
        proactiveType: proactiveData.type,
        proactiveData: proactiveData
      }

      // 通知所有注册的回调
      proactiveCallbacks.forEach(cb => cb(chatMessage))

      // 显示通知
      showProactiveNotification(proactiveData)
    }
  }

  const showAlertNotification = (alert: Alert) => {
    const typeMap = {
      critical: 'error',
      high: 'warning',
      medium: 'warning',
      low: 'info'
    }

    ElNotification({
      title: alert.title,
      message: alert.message,
      type: typeMap[alert.level] as any,
      duration: alert.level === 'critical' ? 0 : 5000,
      position: 'bottom-right',
      customClass: 'monitor-alert-notification',
      onClick: () => {
        window.dispatchEvent(new CustomEvent('show-alert-detail', { detail: alert }))
      }
    })
  }

  const showProactiveNotification = (data: ProactiveMessageData) => {
    const typeMap: Record<string, string> = {
      morning_report: 'success',
      anomaly_alert: 'warning',
      health_summary: 'info',
      auto_fix: 'success'
    }

    const iconMap: Record<string, string> = {
      morning_report: '🌅',
      anomaly_alert: '⚠️',
      health_summary: '📊',
      auto_fix: '🔧'
    }

    ElNotification({
      title: `${iconMap[data.type] || '🤖'} ${data.title}`,
      message: data.summary || data.content?.slice(0, 100) || '',
      type: (typeMap[data.type] || 'info') as any,
      duration: data.priority === 'critical' ? 0 : 8000,
      position: 'bottom-right',
      customClass: 'proactive-notification',
      onClick: () => {
        window.dispatchEvent(new CustomEvent('show-proactive-detail', { detail: data }))
      }
    })
  }

  // 注册主动消息回调（用于注入聊天流）
  const onProactiveMessage = (callback: ProactiveMessageCallback) => {
    proactiveCallbacks.push(callback)
    return () => {
      const index = proactiveCallbacks.indexOf(callback)
      if (index > -1) proactiveCallbacks.splice(index, 1)
    }
  }

  const resolveAlert = (alertId: string) => {
    const alert = alerts.value.find(a => a.id === alertId)
    if (alert) {
      alert.status = 'resolved'
      alert.resolved_at = new Date().toISOString()
    }
  }

  const ignoreAlert = (alertId: string) => {
    const alert = alerts.value.find(a => a.id === alertId)
    if (alert) {
      alert.status = 'ignored'
    }
  }

  const clearAlerts = () => {
    alerts.value = []
  }

  const getActiveAlerts = () => {
    return alerts.value.filter(a => a.status === 'active')
  }

  const markProactiveRead = (id: string) => {
    const msg = proactiveMessages.value.find(m => m.id === id)
    if (msg) msg.read = true
  }

  const getUnreadProactiveCount = () => {
    return proactiveMessages.value.filter(m => !m.read).length
  }

  onMounted(() => {
    const removeListener = addMessageListener(handleMessage)

    onUnmounted(() => {
      removeListener()
      disconnect()
    })
  })

  return {
    alerts,
    latestMetrics,
    proactiveMessages,
    isConnected,
    resolveAlert,
    ignoreAlert,
    clearAlerts,
    getActiveAlerts,
    onProactiveMessage,
    markProactiveRead,
    getUnreadProactiveCount,
    connect,
    disconnect
  }
}
