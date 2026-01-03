import { ref, computed, watch } from 'vue'
import { usePrometheusStore } from '@/stores/prometheus'
import type { TimeRange, QueryResult } from '@/stores/prometheus'

export function usePrometheus() {
  const prometheusStore = usePrometheusStore()
  const loading = ref(false)
  const error = ref<string>()

  /**
   * 执行查询
   */
  const executeQuery = async (query: string): Promise<QueryResult | null> => {
    if (!query.trim()) return null

    loading.value = true
    error.value = undefined

    try {
      const result = await prometheusStore.executeQuery(query)

      if (result.error) {
        error.value = result.error
        return null
      }

      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error'
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 执行即时查询
   */
  const executeInstantQuery = async (query: string): Promise<QueryResult | null> => {
    if (!query.trim()) return null

    loading.value = true
    error.value = undefined

    try {
      const result = await prometheusStore.executeInstantQuery(query)

      if (result.error) {
        error.value = result.error
        return null
      }

      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error'
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新时间范围
   */
  const updateTimeRange = (range: Partial<TimeRange>) => {
    prometheusStore.updateTimeRange(range)
  }

  /**
   * 设置相对时间范围
   */
  const setRelativeTimeRange = (duration: number) => {
    const end = Date.now()
    const start = end - duration

    updateTimeRange({ start, end })
  }

  /**
   * 获取标签值
   */
  const getLabelValues = async (label: string): Promise<string[]> => {
    try {
      return await prometheusStore.getLabelValues(label)
    } catch {
      return []
    }
  }

  /**
   * 获取指标列表
   */
  const getMetrics = async (): Promise<string[]> => {
    try {
      return await prometheusStore.getMetrics()
    } catch {
      return []
    }
  }

  /**
   * 检查连接状态
   */
  const checkConnection = async (): Promise<boolean> => {
    return await prometheusStore.checkConnection()
  }

  return {
    loading,
    error,
    timeRange: computed(() => prometheusStore.timeRange),
    connected: computed(() => prometheusStore.connected),
    recentQueries: computed(() => prometheusStore.recentQueries),
    executeQuery,
    executeInstantQuery,
    updateTimeRange,
    setRelativeTimeRange,
    getLabelValues,
    getMetrics,
    checkConnection
  }
}
