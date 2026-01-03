// 监控告警 WebSocket 集成
import { ref, onMounted, onUnmounted } from 'vue'
import { ElNotification } from 'element-plus'
import { useWebSocket } from './useWebSocket'

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

export function useMonitorAlerts() {
  const alerts = ref<Alert[]>([])
  const latestMetrics = ref<Map<string, HostMetrics>>(new Map())

  const wsUrl = `ws://${window.location.host}/api/monitor/ws`
  const { addMessageListener, connect, disconnect, isConnected } = useWebSocket({ url: wsUrl })

  const handleMessage = (message: any) => {
    if (message.type === 'alert') {
      const alert = message.data as Alert
      alerts.value.unshift(alert)

      // 显示通知
      showAlertNotification(alert)
    } else if (message.type === 'metrics') {
      const metrics = message.data as HostMetrics
      latestMetrics.value.set(metrics.host_name, metrics)
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
        // 触发告警详情展示
        window.dispatchEvent(new CustomEvent('show-alert-detail', { detail: alert }))
      }
    })
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
    isConnected,
    resolveAlert,
    ignoreAlert,
    clearAlerts,
    getActiveAlerts,
    connect,
    disconnect
  }
}
