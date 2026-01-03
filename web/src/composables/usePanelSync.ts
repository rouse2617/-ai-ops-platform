import { ref, watch, onMounted, onUnmounted } from 'vue'
import { usePanelStore } from '@/stores/panel'

export function usePanelSync() {
  const panelStore = usePanelStore()
  const syncEnabled = ref(true)
  const refreshTimers = new Map<string, number>()

  /**
   * 同步到指定主机
   */
  const syncToHost = (hostId: string) => {
    if (!syncEnabled.value) return

    panelStore.syncToHost(hostId)
    startAutoRefresh()
  }

  /**
   * 启动自动刷新
   */
  const startAutoRefresh = () => {
    stopAutoRefresh()

    panelStore.activePanels.forEach(panel => {
      if (panel.refreshInterval && panel.refreshInterval > 0) {
        const timer = window.setInterval(() => {
          panelStore.refreshPanel(panel.id)
        }, panel.refreshInterval)

        refreshTimers.set(panel.id, timer)
      }
    })
  }

  /**
   * 停止自动刷新
   */
  const stopAutoRefresh = () => {
    refreshTimers.forEach(timer => clearInterval(timer))
    refreshTimers.clear()
  }

  /**
   * 切换同步状态
   */
  const toggleSync = () => {
    syncEnabled.value = !syncEnabled.value
    panelStore.toggleSync()

    if (syncEnabled.value) {
      startAutoRefresh()
    } else {
      stopAutoRefresh()
    }
  }

  /**
   * 刷新指定面板
   */
  const refreshPanel = async (panelId: string) => {
    await panelStore.refreshPanel(panelId)
  }

  /**
   * 刷新所有面板
   */
  const refreshAllPanels = async () => {
    await panelStore.refreshAllPanels()
  }

  // 监听激活主机变化
  watch(() => panelStore.activeHost, (newHost) => {
    if (newHost && syncEnabled.value) {
      startAutoRefresh()
    }
  })

  // 组件挂载时启动自动刷新
  onMounted(() => {
    if (panelStore.activeHost && syncEnabled.value) {
      startAutoRefresh()
    }
  })

  // 组件卸载时清理定时器
  onUnmounted(() => {
    stopAutoRefresh()
  })

  return {
    syncEnabled,
    activeHost: panelStore.activeHost,
    syncMode: panelStore.syncMode,
    syncToHost,
    toggleSync,
    refreshPanel,
    refreshAllPanels,
    startAutoRefresh,
    stopAutoRefresh
  }
}
