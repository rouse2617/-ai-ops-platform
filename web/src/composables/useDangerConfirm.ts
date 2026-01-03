import { ref, onMounted, onUnmounted } from 'vue'
import { useDangerStore } from '@/stores/danger'
import type { DangerOperation } from '@/stores/danger'

export function useDangerConfirm() {
  const dangerStore = useDangerStore()
  const pendingOperation = ref<DangerOperation | null>(null)
  const showDialog = ref(false)

  /**
   * 检查命令是否需要确认
   */
  const checkCommand = (command: string): boolean => {
    const { level } = dangerStore.checkDangerLevel(command)
    return level === 'high' || level === 'critical'
  }

  /**
   * 请求用户确认危险操作
   */
  const requestConfirm = async (
    command: string,
    affectedHosts: string[] = []
  ): Promise<boolean> => {
    const { level, reason, type } = dangerStore.checkDangerLevel(command)

    if (level === 'safe' || level === 'low' || level === 'medium') {
      return true
    }

    const operation: DangerOperation = {
      id: `op-${Date.now()}`,
      command,
      type: type || 'modify',
      riskLevel: level as 'high' | 'critical',
      affectedHosts,
      reason,
      timestamp: Date.now()
    }

    pendingOperation.value = operation
    showDialog.value = true

    return await dangerStore.requestConfirm(operation)
  }

  /**
   * 确认操作
   */
  const confirm = () => {
    if (pendingOperation.value) {
      dangerStore.recordConfirm(pendingOperation.value, true)
      showDialog.value = false
      pendingOperation.value = null
    }
  }

  /**
   * 取消操作
   */
  const cancel = () => {
    if (pendingOperation.value) {
      dangerStore.recordConfirm(pendingOperation.value, false)
      showDialog.value = false
      pendingOperation.value = null
    }
  }

  return {
    pendingOperation,
    showDialog,
    checkCommand,
    requestConfirm,
    confirm,
    cancel
  }
}
