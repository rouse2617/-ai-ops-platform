import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface PrometheusQuery {
  id: string
  query: string
  legend?: string
  color?: string
}

export interface TimeRange {
  start: number
  end: number
  step?: string
}

export interface QueryResult {
  query: string
  data: MetricData[]
  timestamp: number
  error?: string
}

export interface MetricData {
  metric: Record<string, string>
  values: [number, string][]
}

export interface ChartConfig {
  id: string
  title: string
  queries: PrometheusQuery[]
  timeRange: TimeRange
  chartType: 'line' | 'area' | 'bar'
  refreshInterval: number
  height: number
}

export const usePrometheusStore = defineStore('prometheus', () => {
  const queries = ref(new Map<string, QueryResult>())
  const timeRange = ref<TimeRange>({
    start: Date.now() - 3600000, // 1小时前
    end: Date.now(),
    step: '15s'
  })
  const refreshInterval = ref(30000) // 30秒
  const connected = ref(false)
  const charts = ref<ChartConfig[]>([])
  const queryHistory = ref<string[]>([])

  const isConnected = computed(() => connected.value)

  const recentQueries = computed(() => {
    return queryHistory.value.slice(-10).reverse()
  })

  /**
   * 执行 PromQL 查询
   */
  async function executeQuery(query: string): Promise<QueryResult> {
    try {
      // 这里应该调用实际的 Prometheus API
      // 示例：使用 fetch 调用 /api/v1/query_range
      const params = new URLSearchParams({
        query,
        start: (timeRange.value.start / 1000).toString(),
        end: (timeRange.value.end / 1000).toString(),
        step: timeRange.value.step || '15s'
      })

      const response = await fetch(`/api/prometheus/query_range?${params}`)
      const data = await response.json()

      const result: QueryResult = {
        query,
        data: data.data?.result || [],
        timestamp: Date.now()
      }

      queries.value.set(query, result)
      addToHistory(query)

      return result
    } catch (error) {
      const errorResult: QueryResult = {
        query,
        data: [],
        timestamp: Date.now(),
        error: error instanceof Error ? error.message : 'Unknown error'
      }

      queries.value.set(query, errorResult)
      return errorResult
    }
  }

  /**
   * 执行即时查询
   */
  async function executeInstantQuery(query: string): Promise<QueryResult> {
    try {
      const params = new URLSearchParams({
        query,
        time: (Date.now() / 1000).toString()
      })

      const response = await fetch(`/api/prometheus/query?${params}`)
      const data = await response.json()

      const result: QueryResult = {
        query,
        data: data.data?.result || [],
        timestamp: Date.now()
      }

      queries.value.set(query, result)
      addToHistory(query)

      return result
    } catch (error) {
      const errorResult: QueryResult = {
        query,
        data: [],
        timestamp: Date.now(),
        error: error instanceof Error ? error.message : 'Unknown error'
      }

      queries.value.set(query, errorResult)
      return errorResult
    }
  }

  /**
   * 更新时间范围
   */
  function updateTimeRange(range: Partial<TimeRange>): void {
    timeRange.value = { ...timeRange.value, ...range }
  }

  /**
   * 设置刷新间隔
   */
  function setRefreshInterval(interval: number): void {
    refreshInterval.value = interval
  }

  /**
   * 添加图表
   */
  function addChart(config: ChartConfig): void {
    charts.value.push(config)
  }

  /**
   * 更新图表
   */
  function updateChart(chartId: string, config: Partial<ChartConfig>): void {
    const chart = charts.value.find(c => c.id === chartId)
    if (chart) {
      Object.assign(chart, config)
    }
  }

  /**
   * 删除图表
   */
  function removeChart(chartId: string): void {
    const index = charts.value.findIndex(c => c.id === chartId)
    if (index !== -1) {
      charts.value.splice(index, 1)
    }
  }

  /**
   * 添加到查询历史
   */
  function addToHistory(query: string): void {
    if (!queryHistory.value.includes(query)) {
      queryHistory.value.push(query)

      // 只保留最近50条
      if (queryHistory.value.length > 50) {
        queryHistory.value = queryHistory.value.slice(-50)
      }

      saveHistory()
    }
  }

  /**
   * 清空查询历史
   */
  function clearHistory(): void {
    queryHistory.value = []
    saveHistory()
  }

  /**
   * 保存历史到 localStorage
   */
  function saveHistory(): void {
    try {
      localStorage.setItem('prometheus-query-history', JSON.stringify(queryHistory.value))
    } catch (error) {
      console.error('Failed to save query history:', error)
    }
  }

  /**
   * 从 localStorage 加载历史
   */
  function loadHistory(): void {
    try {
      const saved = localStorage.getItem('prometheus-query-history')
      if (saved) {
        queryHistory.value = JSON.parse(saved)
      }
    } catch (error) {
      console.error('Failed to load query history:', error)
    }
  }

  /**
   * 检查连接状态
   */
  async function checkConnection(): Promise<boolean> {
    try {
      const response = await fetch('/api/prometheus/health')
      connected.value = response.ok
      return response.ok
    } catch {
      connected.value = false
      return false
    }
  }

  /**
   * 获取标签值
   */
  async function getLabelValues(label: string): Promise<string[]> {
    try {
      const response = await fetch(`/api/prometheus/label/${label}/values`)
      const data = await response.json()
      return data.data || []
    } catch {
      return []
    }
  }

  /**
   * 获取指标列表
   */
  async function getMetrics(): Promise<string[]> {
    try {
      const response = await fetch('/api/prometheus/label/__name__/values')
      const data = await response.json()
      return data.data || []
    } catch {
      return []
    }
  }

  // 初始化
  loadHistory()
  checkConnection()

  return {
    queries,
    timeRange,
    refreshInterval,
    connected,
    charts,
    queryHistory,
    isConnected,
    recentQueries,
    executeQuery,
    executeInstantQuery,
    updateTimeRange,
    setRefreshInterval,
    addChart,
    updateChart,
    removeChart,
    addToHistory,
    clearHistory,
    checkConnection,
    getLabelValues,
    getMetrics
  }
})
