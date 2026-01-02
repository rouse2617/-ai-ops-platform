<template>
  <div class="monitor-chart">
    <div ref="chartRef" class="chart-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { useMetricsStore } from '@/stores/metrics'
import type { EChartsOption } from 'echarts'

interface Props {
  hostId: string
  metric: 'cpu' | 'memory' | 'disk'
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  height: '200px'
})

const emit = defineEmits<{
  anomalyClick: [timestamp: number]
}>()

const metricsStore = useMetricsStore()
const chartRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null

const thresholds = {
  cpu: 80,
  memory: 85,
  disk: 90
}

const metricLabels = {
  cpu: 'CPU',
  memory: 'Memory',
  disk: 'Disk'
}

function initChart() {
  if (!chartRef.value) return

  chartInstance = echarts.init(chartRef.value)
  updateChart()

  window.addEventListener('resize', handleResize)
}

function updateChart() {
  if (!chartInstance) return

  const history = metricsStore.getMetricHistory(props.hostId, props.metric)
  const threshold = thresholds[props.metric]

  const times = history.map(d => new Date(d.timestamp).toLocaleTimeString())
  const values = history.map(d => d.value)
  const anomalies = history.filter(d => d.value > threshold)

  const currentValue = values[values.length - 1] || 0
  const isAbnormal = currentValue > threshold

  const option: EChartsOption = {
    grid: {
      left: 50,
      right: 20,
      top: 40,
      bottom: 30
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLine: { lineStyle: { color: '#ddd' } },
      axisLabel: { color: '#666', fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: '#ddd' } },
      axisLabel: {
        color: '#666',
        formatter: '{value}%'
      },
      splitLine: { lineStyle: { color: '#f0f0f0' } }
    },
    series: [
      {
        name: metricLabels[props.metric],
        type: 'line',
        data: values,
        smooth: true,
        lineStyle: {
          color: isAbnormal ? '#f56c6c' : '#409eff',
          width: 2
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: isAbnormal ? 'rgba(245, 108, 108, 0.3)' : 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: isAbnormal ? 'rgba(245, 108, 108, 0.05)' : 'rgba(64, 158, 255, 0.05)' }
          ])
        },
        markLine: {
          silent: false,
          symbol: 'none',
          data: [
            {
              yAxis: threshold,
              lineStyle: { color: '#e6a23c', type: 'dashed', width: 1 },
              label: {
                formatter: `Threshold: ${threshold}%`,
                color: '#e6a23c',
                fontSize: 11
              }
            }
          ]
        },
        markPoint: {
          symbol: 'pin',
          symbolSize: 40,
          data: anomalies.map(d => ({
            coord: [new Date(d.timestamp).toLocaleTimeString(), d.value],
            value: d.value.toFixed(1),
            itemStyle: { color: '#f56c6c' }
          }))
        }
      }
    ],
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const data = params[0]
        const value = data.value
        const color = value > threshold ? '#f56c6c' : '#409eff'
        return `
          <div style="padding: 5px;">
            <div style="color: ${color}; font-weight: bold;">${metricLabels[props.metric]}: ${value.toFixed(1)}%</div>
            <div style="color: #999; font-size: 11px;">${data.name}</div>
          </div>
        `
      }
    },
    graphic: currentValue ? [
      {
        type: 'text',
        left: 'center',
        top: 10,
        style: {
          text: `${currentValue.toFixed(1)}%`,
          fontSize: 16,
          fontWeight: 'bold',
          fill: isAbnormal ? '#f56c6c' : '#409eff'
        }
      }
    ] : []
  }

  chartInstance.setOption(option)

  // Handle anomaly clicks
  chartInstance.off('click')
  chartInstance.on('click', (params: any) => {
    if (params.componentType === 'markPoint') {
      const timestamp = history[params.dataIndex]?.timestamp
      if (timestamp) {
        emit('anomalyClick', timestamp)
      }
    }
  })
}

function handleResize() {
  chartInstance?.resize()
}

watch(
  () => metricsStore.metricsHistory.get(props.hostId),
  () => {
    updateChart()
  },
  { deep: true }
)

onMounted(() => {
  initChart()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
})
</script>

<style scoped>
.monitor-chart {
  width: 100%;
  height: v-bind(height);
}

.chart-container {
  width: 100%;
  height: 100%;
}
</style>
