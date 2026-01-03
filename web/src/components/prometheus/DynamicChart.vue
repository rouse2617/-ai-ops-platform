<template>
  <div class="prometheus-chart" :style="{ height: `${height}px` }">
    <div class="chart-header">
      <div class="chart-title">
        <el-icon><TrendCharts /></el-icon>
        <span>{{ title }}</span>
      </div>
      <div class="chart-controls">
        <el-button-group size="small">
          <el-button
            v-for="range in timeRanges"
            :key="range.value"
            :type="selectedRange === range.value ? 'primary' : 'default'"
            @click="handleTimeRangeChange(range.value)"
          >
            {{ range.label }}
          </el-button>
        </el-button-group>
        <el-button
          :icon="Refresh"
          circle
          size="small"
          :loading="loading"
          @click="handleRefresh"
        />
      </div>
    </div>

    <div ref="chartRef" class="chart-container" />

    <div v-if="loading" class="chart-loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
    </div>

    <div v-if="error" class="chart-error">
      <el-alert type="error" :closable="false" show-icon>
        <template #title>查询失败</template>
        {{ error }}
      </el-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { TrendCharts, Refresh, Loading } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import { usePrometheusStore } from '@/stores/prometheus'

interface Props {
  query: string
  title?: string
  timeRange?: { start: number; end: number; step?: string }
  refreshInterval?: number
  chartType?: 'line' | 'area' | 'bar'
  height?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: '监控图表',
  refreshInterval: 30000,
  chartType: 'line',
  height: 300
})

const prometheusStore = usePrometheusStore()

const chartRef = ref<HTMLElement>()
const loading = ref(false)
const error = ref<string>()
const selectedRange = ref('1h')
let chartInstance: echarts.ECharts | null = null
let refreshTimer: number | null = null

const timeRanges = [
  { label: '5m', value: '5m', duration: 5 * 60 * 1000 },
  { label: '15m', value: '15m', duration: 15 * 60 * 1000 },
  { label: '1h', value: '1h', duration: 60 * 60 * 1000 },
  { label: '6h', value: '6h', duration: 6 * 60 * 60 * 1000 },
  { label: '24h', value: '24h', duration: 24 * 60 * 60 * 1000 }
]

const initChart = () => {
  if (!chartRef.value) return

  chartInstance = echarts.init(chartRef.value)

  const option: EChartsOption = {
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    xAxis: {
      type: 'time',
      boundaryGap: false
    },
    yAxis: {
      type: 'value'
    },
    series: []
  }

  chartInstance.setOption(option)
}

const updateChart = (data: any[]) => {
  if (!chartInstance) return

  const series = data.map((item, index) => {
    const seriesData = item.values.map(([timestamp, value]: [number, string]) => [
      timestamp * 1000,
      parseFloat(value)
    ])

    const metricLabel = Object.entries(item.metric)
      .map(([key, value]) => `${key}="${value}"`)
      .join(', ')

    return {
      name: metricLabel || `Series ${index + 1}`,
      type: props.chartType,
      data: seriesData,
      smooth: true,
      areaStyle: props.chartType === 'area' ? {} : undefined,
      emphasis: {
        focus: 'series'
      }
    }
  })

  chartInstance.setOption({
    legend: {
      data: series.map(s => s.name),
      type: 'scroll',
      bottom: 0
    },
    series
  })
}

const fetchData = async () => {
  if (!props.query) return

  loading.value = true
  error.value = undefined

  try {
    const result = await prometheusStore.executeQuery(props.query)

    if (result.error) {
      error.value = result.error
    } else {
      updateChart(result.data)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

const handleTimeRangeChange = (range: string) => {
  selectedRange.value = range
  const rangeConfig = timeRanges.find(r => r.value === range)

  if (rangeConfig) {
    const end = Date.now()
    const start = end - rangeConfig.duration

    prometheusStore.updateTimeRange({ start, end })
    fetchData()
  }
}

const handleRefresh = () => {
  fetchData()
}

const startAutoRefresh = () => {
  if (props.refreshInterval > 0) {
    refreshTimer = window.setInterval(() => {
      fetchData()
    }, props.refreshInterval)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(async () => {
  await nextTick()
  initChart()
  fetchData()
  startAutoRefresh()

  window.addEventListener('resize', () => {
    chartInstance?.resize()
  })
})

onUnmounted(() => {
  stopAutoRefresh()
  chartInstance?.dispose()
})

watch(() => props.query, () => {
  fetchData()
})

watch(() => props.timeRange, () => {
  if (props.timeRange) {
    prometheusStore.updateTimeRange(props.timeRange)
    fetchData()
  }
}, { deep: true })
</script>

<style scoped lang="scss">
.prometheus-chart {
  position: relative;
  background: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);

  .chart-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    .chart-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 16px;
      font-weight: 600;
      color: #303133;

      .el-icon {
        color: #409eff;
      }
    }

    .chart-controls {
      display: flex;
      align-items: center;
      gap: 12px;
    }
  }

  .chart-container {
    width: 100%;
    height: calc(100% - 60px);
  }

  .chart-loading {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.9);
    border-radius: 12px;
    z-index: 10;

    .is-loading {
      color: #409eff;
    }
  }

  .chart-error {
    position: absolute;
    inset: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .prometheus-chart {
    padding: 12px;

    .chart-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .chart-controls {
        width: 100%;
        justify-content: space-between;
      }
    }
  }
}
</style>
