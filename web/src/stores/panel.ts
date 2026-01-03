import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface PanelData {
  type: 'metrics' | 'logs' | 'events' | 'alerts'
  hostId: string
  data: unknown
  timestamp: number
}

export interface PanelConfig {
  id: string
  type: PanelData['type']
  title: string
  refreshInterval?: number
  enabled: boolean
}

export type SyncMode = 'auto' | 'manual'

export const usePanelStore = defineStore('panel', () => {
  const activeHost = ref<string | null>(null)
  const syncMode = ref<SyncMode>('auto')
  const syncEnabled = ref(true)
  const panelData = ref(new Map<string, PanelData>())
  const lastSyncTime = ref(0)
  const panels = ref<PanelConfig[]>([
    {
      id: 'metrics',
      type: 'metrics',
      title: '性能指标',
      refreshInterval: 5000,
      enabled: true
    },
    {
      id: 'logs',
      type: 'logs',
      title: '实时日志',
      refreshInterval: 3000,
      enabled: true
    },
    {
      id: 'events',
      type: 'events',
      title: '系统事件',
      refreshInterval: 10000,
      enabled: true
    },
    {
      id: 'alerts',
      type: 'alerts',
      title: '告警信息',
      refreshInterval: 5000,
      enabled: true
    }
  ])

  const activePanels = computed(() => panels.value.filter(p => p.enabled))

  const currentPanelData = computed(() => {
    if (!activeHost.value) return null

    const data: Record<string, PanelData> = {}
    panelData.value.forEach((value, key) => {
      if (value.hostId === activeHost.value) {
        data[key] = value
      }
    })

    return data
  })

  /**
   * 同步到指定主机
   */
  function syncToHost(hostId: string): void {
    if (!syncEnabled.value && syncMode.value === 'auto') return

    activeHost.value = hostId
    lastSyncTime.value = Date.now()

    // 触发面板数据刷新
    refreshAllPanels()
  }

  /**
   * 更新面板数据
   */
  function updatePanelData(panelId: string, data: PanelData): void {
    const key = `${panelId}-${data.hostId}`
    panelData.value.set(key, data)
  }

  /**
   * 刷新所有面板
   */
  function refreshAllPanels(): void {
    if (!activeHost.value) return

    activePanels.value.forEach(panel => {
      refreshPanel(panel.id)
    })
  }

  /**
   * 刷新单个面板
   */
  async function refreshPanel(panelId: string): Promise<void> {
    if (!activeHost.value) return

    const panel = panels.value.find(p => p.id === panelId)
    if (!panel || !panel.enabled) return

    // 这里应该调用实际的 API 获取数据
    // 示例：模拟数据更新
    const mockData: PanelData = {
      type: panel.type,
      hostId: activeHost.value,
      data: { mock: true },
      timestamp: Date.now()
    }

    updatePanelData(panelId, mockData)
  }

  /**
   * 启用自动同步
   */
  function enableAutoSync(): void {
    syncMode.value = 'auto'
    syncEnabled.value = true
  }

  /**
   * 禁用自动同步
   */
  function disableAutoSync(): void {
    syncMode.value = 'manual'
  }

  /**
   * 切换同步状态
   */
  function toggleSync(): void {
    syncEnabled.value = !syncEnabled.value
  }

  /**
   * 更新面板配置
   */
  function updatePanelConfig(panelId: string, config: Partial<PanelConfig>): void {
    const panel = panels.value.find(p => p.id === panelId)
    if (panel) {
      Object.assign(panel, config)
    }
  }

  /**
   * 启用/禁用面板
   */
  function togglePanel(panelId: string): void {
    const panel = panels.value.find(p => p.id === panelId)
    if (panel) {
      panel.enabled = !panel.enabled
    }
  }

  /**
   * 清空面板数据
   */
  function clearPanelData(hostId?: string): void {
    if (hostId) {
      // 清空指定主机的数据
      const keysToDelete: string[] = []
      panelData.value.forEach((value, key) => {
        if (value.hostId === hostId) {
          keysToDelete.push(key)
        }
      })
      keysToDelete.forEach(key => panelData.value.delete(key))
    } else {
      // 清空所有数据
      panelData.value.clear()
    }
  }

  /**
   * 获取面板数据
   */
  function getPanelData(panelId: string, hostId?: string): PanelData | null {
    const targetHost = hostId || activeHost.value
    if (!targetHost) return null

    const key = `${panelId}-${targetHost}`
    return panelData.value.get(key) || null
  }

  return {
    activeHost,
    syncMode,
    syncEnabled,
    panelData,
    lastSyncTime,
    panels,
    activePanels,
    currentPanelData,
    syncToHost,
    updatePanelData,
    refreshAllPanels,
    refreshPanel,
    enableAutoSync,
    disableAutoSync,
    toggleSync,
    updatePanelConfig,
    togglePanel,
    clearPanelData,
    getPanelData
  }
})
